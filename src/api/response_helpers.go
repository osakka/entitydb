package api

import (
	"encoding/json"
	"entitydb/storage/pools"
	"net/http"
)

// RespondJSON writes a JSON response with buffer pooling
func RespondJSON(w http.ResponseWriter, code int, payload interface{}) {
	// Get buffer from pool
	buf := pools.GetBuffer()
	defer pools.PutBuffer(buf)
	
	// Create encoder with our buffer
	encoder := json.NewEncoder(buf)
	
	// Encode payload
	if err := encoder.Encode(payload); err != nil {
		// Fallback to non-pooled approach on error
		response, _ := json.Marshal(payload)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		w.Write(response)
		return
	}
	
	// Write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(buf.Bytes())
}

// RespondError writes a JSON error response
func RespondError(w http.ResponseWriter, code int, message string) {
	RespondJSON(w, code, map[string]string{"error": message})
}

// DecodeJSON decodes JSON from request body using pooled decoder
func DecodeJSON(r *http.Request, v interface{}) error {
	// For now, just use regular decoder since Reset is not available
	// TODO: Implement proper decoder pooling with custom wrapper
	return json.NewDecoder(r.Body).Decode(v)
}