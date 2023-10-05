// This package is a reduced version of the go trace implementations
package tracer

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	util "github.com/jimxshaw/tracerlogger"
)

const (
	traceCtxKey       string = "trace"
	traceVersion      string = "00"
	traceparentHeader string = "traceparent"
	tracestateHeader  string = "tracestate"
	traceStateFormat  string = "node=%02x"

	sampleFlag       = traceFlag(0x01)
	traceIDHexLength = 32
	spanIDHexLength  = 16
	flagLength       = 2
	maxVersion       = 0
	versionLength    = 2
	traceElements    = 4
)

var (
	errInvalidHex         = errors.New("invalid hex value")
	errInvalidTraceLength = errors.New("length of traceid must be 32")
	errInvalidTraceID     = errors.New("invalid trace ID")
	errInvalidSpanID      = errors.New("invalid span ID")
	errInvalidSpanLength  = errors.New("length of traceid must be 16")
	traceRegex            = regexp.MustCompile(`^(?P<version>[0-9a-f]{2})-(?P<traceID>[a-f0-9]{32})-(?P<spanID>[a-f0-9]{16})-(?P<traceFlags>[a-f0-9]{2})(?:-.*)?$`)
	stateRegex            = regexp.MustCompile(`^node=(?P<node>[0-9a-f]{2})(?:-.*)?$`)
)

type traceID [16]byte

var nilTraceID traceID

func (ti traceID) IsValid() bool {
	return !bytes.Equal(ti[:], nilTraceID[:])
}

func (ti traceID) MarshalJSON() ([]byte, error) {
	return json.Marshal(ti.String())
}

func (ti traceID) String() string {
	return hex.EncodeToString(ti[:])
}

type spanID [8]byte

var nilSpanID spanID

func (si spanID) IsValid() bool {
	return !bytes.Equal(si[:], nilSpanID[:])
}

func (si spanID) MarshalJSON() ([]byte, error) {
	return json.Marshal(si.String())
}

func (si spanID) String() string {
	return hex.EncodeToString(si[:])
}

type traceFlag byte

// HexToTraceID convert string to traceID.
func HexToTraceID(hexString string) (traceID, error) {
	trace := traceID{}
	if len(hexString) != traceIDHexLength {
		return trace, errInvalidTraceLength
	}

	err := decodeHex(hexString, trace[:])
	if err != nil {
		return trace, err
	}

	if !trace.IsValid() {
		return trace, errInvalidTraceID
	}

	return trace, nil
}

// HexToSpanID convert string to spanID.
func HexToSpanID(hexString string) (spanID, error) {
	span := spanID{}
	if len(hexString) != spanIDHexLength {
		return span, errInvalidSpanLength
	}

	err := decodeHex(hexString, span[:])
	if err != nil {
		return span, err
	}

	if !span.IsValid() {
		return span, errInvalidSpanID
	}

	return span, nil
}

// decodeHex decode string to bytes.
func decodeHex(hexString string, b []byte) error {
	decoded, err := hex.DecodeString(hexString)
	if err != nil {
		return errInvalidHex
	}

	copy(b, decoded)
	return nil
}

// ipToHex convert the ip to hex value.
func ipToHex(ip net.IP) string {
	ip = ip.To16()
	i := int(ip[12]) * 16777216
	i += int(ip[13]) * 65536
	i += int(ip[14]) * 256
	i += int(ip[15])
	return fmt.Sprintf("%08x", i)
}

// makeTimestamp get the current time in miliseconds.
func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

type traceState struct {
	Node string
}

// TracerContext store the trace context to be propagated.
// This wraps the http.ResponseWriter, injecting the trace headers in
// the various end point responses.
type TracerContext struct {
	http.ResponseWriter
	Span       spanID
	Trace      traceID
	Flag       traceFlag
	childs     int
	traceState traceState
	nodeID     string
	internal   bool
}

// TraceField is the sanitized version of the TracerContext in case that
// the TracerContext will be returned as data structure.
type TraceField struct {
	TraceID string `json:"trace_id"`
	SpandID string `json:"span_id"`
}

func (tc *TracerContext) String() string {
	return fmt.Sprintf(
		"%02x-%s-%s-%s",
		maxVersion,
		tc.Trace.String(),
		tc.nodeID,
		"00",
	)
}

// mapSpan maps the request position as node in the request tree.
func (tc *TracerContext) mapSpan() {
	if !tc.Span.IsValid() {
		tc.nodeID = "0100000000000000"
		return
	}
	var nodeID spanID

	var key [1]byte
	decodeHex("00", key[:])

	var newKey [1]byte
	decodeHex(tc.traceState.Node, newKey[:])

	index := bytes.Index(tc.Span[:], key[:])
	if index < 0 {
		tc.nodeID = "0100000000000000"
		return
	}

	copy(nodeID[:], tc.Span[:])
	copy(nodeID[index:], newKey[:])

	tc.nodeID = hex.EncodeToString(nodeID[:])
}

// next generates or gets the traceID and generates the state for the next node.
func (tc *TracerContext) next() (trace string, state string) {
	tc.childs++
	newNodeID := fmt.Sprintf(traceStateFormat, tc.childs)
	currentTrace := tc.String()
	return currentTrace, newNodeID
}

// IsValid returns true if the Trace and Span are valid.
func (tc *TracerContext) IsValid() bool {
	isValidTrace := tc.Trace.IsValid()
	isValidSpan := tc.Span.IsValid()
	return isValidTrace && isValidSpan
}

// propagateHeaders add the trace headers in the responses.
func (tc *TracerContext) propagateHeaders() {

	tc.ResponseWriter.Header().Add(traceparentHeader, tc.String())
	state := fmt.Sprintf("node=%s", tc.traceState.Node)
	if tc.internal {
		tc.ResponseWriter.Header().Add(tracestateHeader, state)
	}

}

func (tc *TracerContext) WriteHeader(statusCode int) {
	tc.propagateHeaders()
	tc.ResponseWriter.WriteHeader(statusCode)
}

// Propagate injects the TracerContext into http.Request and http.ResponseWriter.
func (tc *TracerContext) Propagate(r *http.Request, w http.ResponseWriter) (*http.Request, http.ResponseWriter) {
	tc.ResponseWriter = w
	newTrace, internal := generateTrace(r)
	tc.internal = internal
	if !internal || !tc.IsValid() {
		tc.Trace = newTrace
		tc.traceState.Node = "01"
	}
	tc.mapSpan()
	return InjectInRequest(r, tc), tc
}

func (tc *TracerContext) Sanitize() TraceField {
	return TraceField{
		SpandID: tc.nodeID,
		TraceID: tc.Trace.String(),
	}
}

// Propagator implements the logic to be propagated in the request.
type Propagator interface {
	// Inject the trace context in the http.Request context and the http.ResponseWriter
	Propagate(r *http.Request, w http.ResponseWriter) (*http.Request, http.ResponseWriter)
	// Return a sanitized version of the TracerContext
	Sanitize() TraceField
	// Get the next node values
	next() (trace string, state string)
}

// ExtractFromRequest extracts the trace information from the request headers.
func ExtractFromRequest(r *http.Request) Propagator {
	tracer := &TracerContext{}
	rawTraceParent := r.Header.Get(traceparentHeader)
	rawTraceState := r.Header.Get(tracestateHeader)
	state := extractState(rawTraceState)
	elements := traceRegex.FindStringSubmatch(rawTraceParent)

	if len(elements) == 0 {
		return tracer
	}
	// the FindStringSubmatch return string match and  the sub matches
	// thats why traceElements + 1
	if len(elements) < traceElements+1 {
		return tracer
	}

	if len(elements[1]) != versionLength {
		return tracer
	}

	verByte, err := hex.DecodeString(elements[1])
	if err != nil {
		return tracer
	}

	version := int(verByte[0])
	if version > maxVersion {
		return tracer
	}

	if len(elements[2]) != traceIDHexLength {
		return tracer
	}

	if len(elements[3]) != spanIDHexLength {
		return tracer
	}

	if len(elements[4]) != flagLength {
		return tracer
	}
	flag, err := hex.DecodeString(elements[4])
	if err != nil {
		return tracer
	}

	traceID, err := HexToTraceID(elements[2])
	if err != nil {
		return tracer
	}

	spanID, err := HexToSpanID(elements[3])
	if err != nil {
		return tracer
	}

	tracer.Flag = traceFlag(flag[0])
	tracer.Trace = traceID
	tracer.Span = spanID
	tracer.traceState = state

	return tracer
}

// ExtractFromCtx extracts the trace information from the context.
// In case the middleware is not loaded the the function returns an empty propagator.
func ExtractFromCtx(ctx context.Context) Propagator {
	ctxValue, ok := ctx.Value(traceCtxKey).(*TracerContext)
	if ok {
		return ctxValue
	}

	return NewTracerContext()
}

// InjectInRequest the trace information in the request context.
func InjectInRequest(r *http.Request, trace *TracerContext) *http.Request {
	ctx := InjectInCtx(r.Context(), trace)
	return r.WithContext(ctx)
}

// InjectHeaders injects the trace information in the request headers.
// The intention is that it will be used in API clients for internal calls.
func InjectHeaders(ctx context.Context, req *http.Request) *http.Request {
	propagator := ExtractFromCtx(ctx)
	trace, state := propagator.next()
	req.Header.Set(traceparentHeader, trace)
	req.Header.Set(tracestateHeader, state)
	return req
}

// InjectInCtx injects the trace information in the context.
func InjectInCtx(ctx context.Context, trace *TracerContext) context.Context {
	ctxNew := context.WithValue(ctx, traceCtxKey, trace)
	return ctxNew
}

// getServerIp gets the server ip.
func getServerIp() (net.IP, *net.IPNet, error) {
	containerHostname, err := os.Hostname()
	if err != nil {
		return net.IP{}, &net.IPNet{}, err
	}

	ipAddr, err := net.ResolveIPAddr("ip", containerHostname)
	if err != nil {
		return net.IP{}, &net.IPNet{}, err
	}

	containerIp := net.ParseIP(ipAddr.String())
	ipMask := containerIp.DefaultMask()
	network := containerIp.Mask(ipMask)
	net := &net.IPNet{
		IP:   network,
		Mask: ipMask,
	}
	return containerIp, net, nil
}

// generateTrace gets the trace or generates a new one for externals request.
func generateTrace(r *http.Request) (traceID, bool) {
	var ipHex string
	var isInternalRequest bool
	var requestIp net.IP
	ip, network, err := getServerIp()
	remoteAddr := strings.Split(r.RemoteAddr, ":")[0]
	if err == nil {
		requestIp = net.ParseIP(remoteAddr)
		isInternalRequest = network.Contains(requestIp) ||
			isIpPrivateOrLocal(requestIp)
		ipHex = ipToHex(ip)
	} else {
		isInternalRequest = false
		ipHex, _ = util.RandomHex(4)

	}
	currentTime := makeTimestamp()
	uniqueID, _ := util.RandomHex(5)
	traceIDHex := fmt.Sprintf("%s%d0%s", ipHex, currentTime, uniqueID)
	trace, _ := HexToTraceID(traceIDHex)

	return trace, isInternalRequest
}

// NewTracerContext creates a new TracerContext.
func NewTracerContext() *TracerContext {
	ipHex, _ := util.RandomHex(4)
	currentTime := makeTimestamp()
	uniqueID, _ := util.RandomHex(5)
	traceIDHex := fmt.Sprintf("%s%d0%s", ipHex, currentTime, uniqueID)
	trace, _ := HexToTraceID(traceIDHex)

	tc := &TracerContext{
		Trace: trace,
		traceState: traceState{
			Node: "01",
		},
	}
	tc.mapSpan()

	return tc
}

// extractState extracts the tracestate in a string value.
func extractState(value string) traceState {
	defaultState := traceState{
		Node: "01",
	}
	matches := stateRegex.FindStringSubmatch(value)

	if len(matches) != 2 {
		return defaultState
	}

	if !isHexString(matches[1]) {
		return defaultState
	}

	return traceState{
		Node: matches[1],
	}
}

// isHexString validate if string is a hex value.
func isHexString(s string) bool {
	_, err := hex.DecodeString(s)
	return err == nil
}

// privateIPBlocks is a map of private and loopbacks ips.
var privateIPBlocks []*net.IPNet

// isIpPrivateOrLocal returns true if the ip is local or private.
func isIpPrivateOrLocal(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}

	for _, block := range privateIPBlocks {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

func init() {
	for _, cidr := range []string{
		"127.0.0.0/8",    // IPv4 loopback
		"10.0.0.0/8",     // RFC1918
		"172.16.0.0/12",  // RFC1918
		"192.168.0.0/16", // RFC1918
		"169.254.0.0/16", // RFC3927 link-local
		"::1/128",        // IPv6 loopback
		"fe80::/10",      // IPv6 link-local
		"fc00::/7",       // IPv6 unique local addr
	} {
		_, block, err := net.ParseCIDR(cidr)
		if err != nil {
			panic(fmt.Errorf("parse error on %q: %v", cidr, err))
		}
		privateIPBlocks = append(privateIPBlocks, block)
	}
}
