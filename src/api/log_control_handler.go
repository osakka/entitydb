package api

import (
	"encoding/json"
	"net/http"
	"strings"
	
	"entitydb/logger"
)

// LogLevelRequest represents a log level change request
type LogLevelRequest struct {
	Level string   `json:"level"`
	Trace []string `json:"trace,omitempty"`
}

// LogLevelResponse represents the current log configuration
type LogLevelResponse struct {
	Level string   `json:"level"`
	Trace []string `json:"trace"`
}

// GetLogLevel returns the current log configuration
func (h *EntityHandler) GetLogLevel(w http.ResponseWriter, r *http.Request) {
	response := LogLevelResponse{
		Level: logger.GetLogLevel(),
		Trace: logger.GetTraceSubsystems(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("failed to encode log level response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// SetLogLevel updates the log configuration dynamically
func (h *EntityHandler) SetLogLevel(w http.ResponseWriter, r *http.Request) {
	var req LogLevelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("invalid log level request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Update log level if provided
	if req.Level != "" {
		if err := logger.SetLogLevel(req.Level); err != nil {
			logger.Warn("invalid log level specified: %s", req.Level)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		logger.Info("log level changed to %s", req.Level)
	}
	
	// Update trace subsystems if provided
	if len(req.Trace) > 0 {
		// Clear existing trace subsystems
		logger.ClearTrace()
		
		// Enable new trace subsystems
		logger.EnableTrace(req.Trace...)
		logger.Info("trace subsystems updated: %s", strings.Join(req.Trace, ", "))
	}
	
	// Return current configuration
	response := LogLevelResponse{
		Level: logger.GetLogLevel(),
		Trace: logger.GetTraceSubsystems(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error("failed to encode log level response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}