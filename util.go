package tracerlogger

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"

	log "github.com/jimxshaw/tracerlogger/logger"
	"go.uber.org/zap"
)

// Header for HTTP Strict Transport Security (HSTS) policy mechanism on web servers.
// It's a time duration of one year that includes sub-domains.
const strictTransportSecurity = "max-age=31536000; includeSubDomains"

// RespondWithJSON send a JSON-formatted response, including the HSTS policy header.
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Strict-Transport-Security", strictTransportSecurity)
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

// RespondWithError send a JSON-formatted error.
func RespondWithError(w http.ResponseWriter, code int, err error) {
	log.Error("request has error", zap.Error(err))
	if err == nil {
		RespondWithJSON(
			w,
			code,
			map[string]string{
				"error": "Something went wrong. Please report the issue to Administrators.",
			},
		)
		return
	}
	RespondWithJSON(w, code, map[string]string{"error": err.Error()})
}

// RandomHex generates a random hex value.
func RandomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
