package api

import (
	"encoding/json"
	"net/http"
	"time"

	"entitydb/models"
)

// SessionHandler manages session-related API endpoints
type SessionHandler struct {
	repo models.SessionRepository
}

// NewSessionHandler creates a new session handler
func NewSessionHandler(repo models.SessionRepository) *SessionHandler {
	return &SessionHandler{
		repo: repo,
	}
}


// CreateSessionRequest represents the request to create a session
type CreateSessionRequest struct {
	AgentID     string   `json:"agent_id"`
	ProjectID   string   `json:"project_id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	ContextFile string   `json:"context_file,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// CreateSession handles session creation
// POST /api/v1/sessions
func (h *SessionHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.AgentID == "" {
		RespondError(w, http.StatusBadRequest, "Agent ID is required")
		return
	}

	if req.ProjectID == "" {
		RespondError(w, http.StatusBadRequest, "Project ID is required")
		return
	}

	if req.Name == "" {
		RespondError(w, http.StatusBadRequest, "Name is required")
		return
	}

	// Create new session
	session := models.NewSession(
		req.AgentID,
		req.ProjectID,
		req.Name,
		req.Description,
	)

	// Set optional fields
	if req.ContextFile != "" {
		session.ContextFile = req.ContextFile
	}

	if len(req.Tags) > 0 {
		session.Tags = req.Tags
	}

	// Save to repository
	if err := h.repo.Create(session); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	// Return the created session
	RespondJSON(w, http.StatusCreated, session)
}

// ListSessions handles listing sessions
// GET /api/v1/sessions
func (h *SessionHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()
	
	// Build filter
	filter := make(map[string]interface{})
	
	if agentID := query.Get("agent_id"); agentID != "" {
		filter["agent_id"] = agentID
	}
	
	if projectID := query.Get("project_id"); projectID != "" {
		filter["project_id"] = projectID
	}
	
	if status := query.Get("status"); status != "" {
		filter["status"] = status
	}

	// Get sessions from repository
	sessions, err := h.repo.List(filter)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to list sessions")
		return
	}

	// Return sessions
	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"sessions": sessions,
		"total":    len(sessions),
	})
}

// GetSession handles retrieving a single session
// GET /api/v1/sessions/get
func (h *SessionHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	// Extract session ID from query parameters
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		RespondError(w, http.StatusBadRequest, "Session ID is required")
		return
	}

	// Get session from repository
	session, err := h.repo.GetByID(sessionID)
	if err != nil {
		RespondError(w, http.StatusNotFound, "Session not found")
		return
	}

	// Return session
	RespondJSON(w, http.StatusOK, session)
}

// UpdateSessionRequest represents the request to update a session
type UpdateSessionRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	ContextFile string   `json:"context_file"`
	Tags        []string `json:"tags"`
}

// UpdateSession handles updating a session
// PUT /api/v1/sessions/update
func (h *SessionHandler) UpdateSession(w http.ResponseWriter, r *http.Request) {
	// Extract session ID from query parameters
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		RespondError(w, http.StatusBadRequest, "Session ID is required")
		return
	}

	// Get session from repository
	session, err := h.repo.GetByID(sessionID)
	if err != nil {
		RespondError(w, http.StatusNotFound, "Session not found")
		return
	}

	// Parse request body
	var req UpdateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Update session fields
	if req.Name != "" {
		session.Name = req.Name
	}
	
	if req.Description != "" {
		session.Description = req.Description
	}
	
	if req.ContextFile != "" {
		session.ContextFile = req.ContextFile
	}
	
	if len(req.Tags) > 0 {
		session.Tags = req.Tags
	}

	// Update timestamp
	session.UpdatedAt = time.Now()

	// Save to repository
	if err := h.repo.Update(session); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to update session")
		return
	}

	// Return updated session
	RespondJSON(w, http.StatusOK, session)
}

// UpdateSessionStatusRequest represents the request to update a session's status
type UpdateSessionStatusRequest struct {
	Status  string `json:"status"`
	Summary string `json:"summary,omitempty"`
}

// UpdateSessionStatus handles updating a session's status
// PUT /api/v1/sessions/update-status
func (h *SessionHandler) UpdateSessionStatus(w http.ResponseWriter, r *http.Request) {
	// Extract session ID from query parameters
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		RespondError(w, http.StatusBadRequest, "Session ID is required")
		return
	}

	// Parse request body
	var req UpdateSessionStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate status
	if req.Status == "" {
		RespondError(w, http.StatusBadRequest, "Status is required")
		return
	}

	// Check if status is valid
	validStatuses := []string{models.SessionStatusActive, models.SessionStatusPaused, models.SessionStatusEnded, models.SessionStatusStale}
	isValidStatus := false
	for _, status := range validStatuses {
		if req.Status == status {
			isValidStatus = true
			break
		}
	}

	if !isValidStatus {
		RespondError(w, http.StatusBadRequest, "Invalid status. Must be one of: active, paused, ended, stale")
		return
	}

	// Update session status
	if err := h.repo.UpdateStatus(sessionID, req.Status); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to update session status")
		return
	}

	// Get updated session
	session, err := h.repo.GetByID(sessionID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to retrieve updated session")
		return
	}

	// Prepare response
	response := map[string]interface{}{
		"id":        session.ID,
		"status":    session.Status,
		"updatedAt": session.UpdatedAt,
	}

	// Add ended_at if status is "ended"
	if req.Status == models.SessionStatusEnded {
		response["endedAt"] = session.EndedAt
	}

	// Return response
	RespondJSON(w, http.StatusOK, response)
}

// PingSession handles updating a session's last activity timestamp
// POST /api/v1/sessions/ping
func (h *SessionHandler) PingSession(w http.ResponseWriter, r *http.Request) {
	// Extract session ID from query parameters
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		RespondError(w, http.StatusBadRequest, "Session ID is required")
		return
	}

	// Update session's last active timestamp
	if err := h.repo.UpdateLastActive(sessionID); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to update session last active timestamp")
		return
	}

	// Get updated session
	session, err := h.repo.GetByID(sessionID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to retrieve updated session")
		return
	}

	// Return response
	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"id":        session.ID,
		"updatedAt": session.UpdatedAt,
	})
}

// SetContextValueRequest represents the request to set a context value
type SetContextValueRequest struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// SetContextValue handles setting a context value for a session
// POST /api/v1/sessions/set-context
func (h *SessionHandler) SetContextValue(w http.ResponseWriter, r *http.Request) {
	// Extract session ID from query parameters
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		RespondError(w, http.StatusBadRequest, "Session ID is required")
		return
	}

	// Parse request body
	var req SetContextValueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.Key == "" {
		RespondError(w, http.StatusBadRequest, "Key is required")
		return
	}

	// Create context entry
	context := models.NewSessionContext(sessionID, req.Key, req.Value)

	// Save to repository
	if err := h.repo.SetContext(context); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to set context value")
		return
	}

	// Return the created context
	RespondJSON(w, http.StatusCreated, context)
}

// GetAllContextValues handles retrieving all context values for a session
// GET /api/v1/sessions/get-context
func (h *SessionHandler) GetAllContextValues(w http.ResponseWriter, r *http.Request) {
	// Extract session ID from query parameters
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		RespondError(w, http.StatusBadRequest, "Session ID is required")
		return
	}

	// Get context values from repository
	contextValues, err := h.repo.GetAllContext(sessionID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to get context values")
		return
	}

	// Return context values
	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"context": contextValues,
	})
}


// DeleteContextValue handles deleting a context value for a session
// DELETE /api/v1/sessions/delete-context
func (h *SessionHandler) DeleteContextValue(w http.ResponseWriter, r *http.Request) {
	// Extract session ID and key from query parameters
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		RespondError(w, http.StatusBadRequest, "Session ID is required")
		return
	}
	
	key := r.URL.Query().Get("key")
	if key == "" {
		RespondError(w, http.StatusBadRequest, "Key is required")
		return
	}

	// Delete context value from repository
	if err := h.repo.DeleteContext(sessionID, key); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to delete context value")
		return
	}

	// Return 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

// GetSessionStats handles retrieving statistics for a session
// GET /api/v1/sessions/stats
func (h *SessionHandler) GetSessionStats(w http.ResponseWriter, r *http.Request) {
	// Extract session ID from query parameters
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		RespondError(w, http.StatusBadRequest, "Session ID is required")
		return
	}

	// Get session from repository
	session, err := h.repo.GetByID(sessionID)
	if err != nil {
		RespondError(w, http.StatusNotFound, "Session not found")
		return
	}

	// Calculate duration
	var durationSeconds int64 = 0
	if session.Status == models.SessionStatusEnded {
		durationSeconds = int64(session.EndedAt.Sub(session.CreatedAt).Seconds())
	} else {
		durationSeconds = int64(time.Since(session.CreatedAt).Seconds())
	}

	// TODO: Get task count and completed tasks count from task repository
	taskCount := 0
	completedTasks := 0

	// Return statistics
	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"id":              session.ID,
		"agentId":         session.AgentID,
		"projectId":       session.ProjectID,
		"name":            session.Name,
		"status":          session.Status,
		"createdAt":       session.CreatedAt,
		"updatedAt":       session.UpdatedAt,
		"endedAt":         session.EndedAt,
		"durationSeconds": durationSeconds,
		"summary":         "", // TODO: Get summary from a summary field
		"taskCount":       taskCount,
		"completedTasks":  completedTasks,
	})
}