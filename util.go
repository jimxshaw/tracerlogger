package tracerlogger

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
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

// RandomHex generates a random hex value.
func RandomHex(n int) (string, error) {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
