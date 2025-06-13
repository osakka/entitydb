package api

import (
	"encoding/json"
	"entitydb/storage/pools"
	"io"
	"net/http"
)

// RespondJSON writes a JSON response with optimized encoder pooling
func RespondJSON(w http.ResponseWriter, code int, payload interface{}) {
	// Get encoder wrapper from pool (includes buffer)
	encoderWrapper := pools.GetJSONEncoder()
	defer pools.PutJSONEncoder(encoderWrapper)
	
	// Encode payload using the wrapper's encoder
	if err := encoderWrapper.Encoder.Encode(payload); err != nil {
		// Fallback to non-pooled approach on error
		response, _ := json.Marshal(payload)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		w.Write(response)
		return
	}
	
	// Write response from the wrapper's buffer
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(encoderWrapper.Buffer.Bytes())
}

// RespondError writes a JSON error response
func RespondError(w http.ResponseWriter, code int, message string) {
	RespondJSON(w, code, map[string]string{"error": message})
}

// DecodeJSON decodes JSON from request body (simple, no pooling for decoders)
func DecodeJSON(r *http.Request, v interface{}) error {
	// JSON decoders are harder to pool efficiently since they bind to specific readers
	// For now, use standard approach but with pooled buffers when possible
	decoder := json.NewDecoder(r.Body)
	return decoder.Decode(v)
}

// DecodeJSONWithOptions decodes JSON with additional configuration options
func DecodeJSONWithOptions(r io.Reader, v interface{}, disallowUnknownFields bool) error {
	// Create decoder for the specific reader
	decoder := json.NewDecoder(r)
	
	// Configure decoder options
	if disallowUnknownFields {
		decoder.DisallowUnknownFields()
	}
	
	// Decode the JSON
	return decoder.Decode(v)
}