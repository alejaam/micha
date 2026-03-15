package httpadapter

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
)

const maxRequestBytes = 1 << 20 // 1 MB

// decodeJSON decodes the request body into dst, enforcing size limit and unknown-field rejection.
// Returns non-nil error and writes a 400 response when decoding fails.
func decodeJSON(r *http.Request, w http.ResponseWriter, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBytes)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "request body is invalid: "+err.Error())
		return err
	}
	return nil
}

// queryInt reads an integer query param, returning fallback on parse failure.
func queryInt(r *http.Request, key string, fallback int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

// writeJSON serialises data to JSON and writes it with status code.
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("handler: failed to encode response", "error", err)
	}
}

// writeError writes a structured JSON error response.
func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, map[string]any{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
