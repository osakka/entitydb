package api

import (
	"encoding/json"
	"entitydb/logger"
	"net/http"
	"strings"
)

// LogControlHandler manages runtime log configuration
type LogControlHandler struct{}

// NewLogControlHandler creates a new log control handler
func NewLogControlHandler() *LogControlHandler {
	return &LogControlHandler{}
}

// SetLogLevelRequest represents a log level change request
type SetLogLevelRequest struct {
	Level string `json:"level"`
}

// SetTraceSubsystemsRequest represents a trace subsystems change request
type SetTraceSubsystemsRequest struct {
	Enable  []string `json:"enable,omitempty"`
	Disable []string `json:"disable,omitempty"`
	Clear   bool     `json:"clear,omitempty"`
}

// LogStatusResponse represents the current logging configuration
type LogStatusResponse struct {
	Level       string   `json:"level"`
	Subsystems  []string `json:"subsystems"`
}

// SetLogLevel changes the runtime log level
// @Summary Set log level
// @Description Change the runtime log level
// @Tags admin
// @Accept json
// @Produce json
// @Param request body SetLogLevelRequest true "Log level"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/log-level [post]
func (h *LogControlHandler) SetLogLevel(w http.ResponseWriter, r *http.Request) {
	var req SetLogLevelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := logger.SetLogLevel(req.Level); err != nil {
		RespondError(w, http.StatusBadRequest, err.Error())
		return
	}

	RespondJSON(w, http.StatusOK, SuccessResponse{
		Success: true,
		Message: "log level updated to " + strings.ToUpper(req.Level),
	})
}

// GetLogLevel returns the current log level
// @Summary Get log level
// @Description Get the current log level and trace subsystems
// @Tags admin
// @Produce json
// @Success 200 {object} LogStatusResponse
// @Security BearerAuth
// @Router /api/v1/admin/log-level [get]
func (h *LogControlHandler) GetLogLevel(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, http.StatusOK, LogStatusResponse{
		Level:      logger.GetLogLevel(),
		Subsystems: logger.GetTraceSubsystems(),
	})
}

// SetTraceSubsystems manages trace subsystems
// @Summary Configure trace subsystems
// @Description Enable, disable, or clear trace subsystems
// @Tags admin
// @Accept json
// @Produce json
// @Param request body SetTraceSubsystemsRequest true "Trace configuration"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/trace-subsystems [post]
func (h *LogControlHandler) SetTraceSubsystems(w http.ResponseWriter, r *http.Request) {
	var req SetTraceSubsystemsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Clear all if requested
	if req.Clear {
		logger.ClearTrace()
		logger.Info("trace subsystems cleared")
	}

	// Enable subsystems
	if len(req.Enable) > 0 {
		logger.EnableTrace(req.Enable...)
		logger.Info("trace subsystems enabled: %v", req.Enable)
	}

	// Disable subsystems
	if len(req.Disable) > 0 {
		logger.DisableTrace(req.Disable...)
		logger.Info("trace subsystems disabled: %v", req.Disable)
	}

	RespondJSON(w, http.StatusOK, SuccessResponse{
		Success: true,
		Message: "trace subsystems updated",
	})
}

// GetTraceSubsystems returns the current trace subsystems
// @Summary Get trace subsystems
// @Description Get the currently enabled trace subsystems
// @Tags admin
// @Produce json
// @Success 200 {object} LogStatusResponse
// @Security BearerAuth
// @Router /api/v1/admin/trace-subsystems [get]
func (h *LogControlHandler) GetTraceSubsystems(w http.ResponseWriter, r *http.Request) {
	RespondJSON(w, http.StatusOK, LogStatusResponse{
		Level:      logger.GetLogLevel(),
		Subsystems: logger.GetTraceSubsystems(),
	})
}