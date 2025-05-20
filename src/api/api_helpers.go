package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ParamsKey is the context key for URL parameters
type ParamsKey struct{}

// DecodeJSONBody decodes a JSON request body into a struct
func DecodeJSONBody(r *http.Request, dst interface{}) error {
	// Check content type
	contentType := r.Header.Get("Content-Type")
	if contentType != "" {
		if !strings.Contains(contentType, "application/json") {
			return fmt.Errorf("Content-Type header is not application/json")
		}
	}

	// Set max body size to 1MB to prevent memory issues
	r.Body = http.MaxBytesReader(nil, r.Body, 1048576)

	// Create decoder
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	// Decode
	err := dec.Decode(&dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("request body contains malformed JSON (at position %d)", syntaxError.Offset)

		case errors.As(err, &unmarshalTypeError):
			return fmt.Errorf("request body contains incorrect JSON type for field %q", unmarshalTypeError.Field)

		case errors.Is(err, io.EOF):
			return errors.New("request body is empty")

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("request body contains unknown field %s", fieldName)

		default:
			return err
		}
	}

	// Check for additional JSON values
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return errors.New("request body must only contain a single JSON object")
	}

	return nil
}