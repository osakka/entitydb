package api

import (
	"encoding/json"
	"net/http"
	"time"

	"entitydb/models"
)

// AgentPoolHandler handles agent pool-related API endpoints
type AgentPoolHandler struct {
	repo models.AgentPoolRepository
}

// NewAgentPoolHandler creates a new agent pool handler
func NewAgentPoolHandler(repo models.AgentPoolRepository) *AgentPoolHandler {
	return &AgentPoolHandler{
		repo: repo,
	}
}

// CreatePoolRequest represents the request to create a pool
type CreatePoolRequest struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	Specialization string `json:"specialization"`
	MinCapability  int    `json:"min_capability"`
}

// CreatePool handles pool creation
// POST /api/v1/pools/create
func (h *AgentPoolHandler) CreatePool(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req CreatePoolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.Name == "" {
		RespondError(w, http.StatusBadRequest, "Pool name is required")
		return
	}

	// Create new pool
	pool := models.NewAgentPool(req.Name, req.Description, req.Specialization)
	
	// Set min capability if provided
	if req.MinCapability > 0 {
		pool.MinCapability = req.MinCapability
	}

	// Save to repository
	if err := h.repo.Create(pool); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to create agent pool")
		return
	}

	// Return the created pool
	RespondJSON(w, http.StatusCreated, pool)
}

// ListPools handles listing pools
// GET /api/v1/pools/list
func (h *AgentPoolHandler) ListPools(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()
	
	// Build filter
	filter := make(map[string]interface{})
	
	if status := query.Get("status"); status != "" {
		filter["status"] = status
	}
	
	if specialization := query.Get("specialization"); specialization != "" {
		filter["specialization"] = specialization
	}
	
	if name := query.Get("name"); name != "" {
		filter["name"] = name
	}

	// Get pools from repository
	pools, err := h.repo.List(filter)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to list agent pools")
		return
	}

	// Return pools
	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"pools": pools,
		"total": len(pools),
	})
}

// GetPool handles retrieving a single pool
// GET /api/v1/pools/get
func (h *AgentPoolHandler) GetPool(w http.ResponseWriter, r *http.Request) {
	// Extract pool ID from query parameters
	poolID := r.URL.Query().Get("pool_id")
	if poolID == "" {
		RespondError(w, http.StatusBadRequest, "Pool ID is required")
		return
	}

	// Get pool from repository
	pool, err := h.repo.GetByID(poolID)
	if err != nil {
		RespondError(w, http.StatusNotFound, "Pool not found")
		return
	}

	// Return pool
	RespondJSON(w, http.StatusOK, pool)
}

// UpdatePoolRequest represents the request to update a pool
type UpdatePoolRequest struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	Specialization string `json:"specialization"`
	MinCapability  int    `json:"min_capability"`
	Status         string `json:"status"`
}

// UpdatePool handles updating a pool
// PUT /api/v1/pools/update
func (h *AgentPoolHandler) UpdatePool(w http.ResponseWriter, r *http.Request) {
	// Extract pool ID from query parameters
	poolID := r.URL.Query().Get("pool_id")
	if poolID == "" {
		RespondError(w, http.StatusBadRequest, "Pool ID is required")
		return
	}

	// Get pool from repository
	pool, err := h.repo.GetByID(poolID)
	if err != nil {
		RespondError(w, http.StatusNotFound, "Pool not found")
		return
	}

	// Parse request body
	var req UpdatePoolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Update pool fields
	if req.Name != "" {
		pool.Name = req.Name
	}
	
	if req.Description != "" {
		pool.Description = req.Description
	}
	
	if req.Specialization != "" {
		pool.Specialization = req.Specialization
	}
	
	if req.MinCapability > 0 {
		pool.MinCapability = req.MinCapability
	}
	
	if req.Status != "" {
		// Validate status
		validStatuses := []string{"active", "inactive", "archived"}
		isValidStatus := false
		for _, status := range validStatuses {
			if req.Status == status {
				isValidStatus = true
				break
			}
		}
		
		if !isValidStatus {
			RespondError(w, http.StatusBadRequest, "Invalid status. Must be one of: active, inactive, archived")
			return
		}
		
		pool.Status = req.Status
	}
	
	// Update timestamp
	pool.UpdatedAt = time.Now()

	// Save to repository
	if err := h.repo.Update(pool); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to update agent pool")
		return
	}

	// Return updated pool
	RespondJSON(w, http.StatusOK, pool)
}

// DeletePool handles deleting a pool
// DELETE /api/v1/pools/delete
func (h *AgentPoolHandler) DeletePool(w http.ResponseWriter, r *http.Request) {
	// Extract pool ID from query parameters
	poolID := r.URL.Query().Get("pool_id")
	if poolID == "" {
		RespondError(w, http.StatusBadRequest, "Pool ID is required")
		return
	}

	// Delete pool from repository
	if err := h.repo.Delete(poolID); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to delete agent pool")
		return
	}

	// Return success
	w.WriteHeader(http.StatusNoContent)
}

// AddAgentRequest represents the request to add an agent to a pool
type AddAgentRequest struct {
	PoolID  string `json:"pool_id"`
	AgentID string `json:"agent_id"`
	Agent   string `json:"agent"` // For backward compatibility with test scripts
}

// AddAgentToPool handles adding an agent to a pool
// POST /api/v1/pools/agents/add
func (h *AgentPoolHandler) AddAgentToPool(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req AddAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.PoolID == "" {
		RespondError(w, http.StatusBadRequest, "Pool ID is required")
		return
	}

	// Get agent ID from either agent_id or agent field for backward compatibility
	agentID := req.AgentID
	
	// If agent_id is empty, try to use agent field instead
	if agentID == "" {
		agentID = req.Agent
	}
	
	if agentID == "" {
		RespondError(w, http.StatusBadRequest, "Agent ID is required (use either agent_id or agent field)")
		return
	}

	// Add agent to pool
	if err := h.repo.AddAgentToPool(req.PoolID, agentID); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to add agent to pool")
		return
	}

	// Get updated pool
	pool, err := h.repo.GetByID(req.PoolID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to retrieve updated pool")
		return
	}

	// Return updated pool
	RespondJSON(w, http.StatusOK, pool)
}

// RemoveAgentRequest represents the request to remove an agent from a pool
type RemoveAgentRequest struct {
	PoolID  string `json:"pool_id"`
	AgentID string `json:"agent_id"`
	Agent   string `json:"agent"` // For backward compatibility with test scripts
}

// RemoveAgentFromPool handles removing an agent from a pool
// POST /api/v1/pools/agents/remove
func (h *AgentPoolHandler) RemoveAgentFromPool(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req RemoveAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.PoolID == "" {
		RespondError(w, http.StatusBadRequest, "Pool ID is required")
		return
	}

	// Get agent ID from either agent_id or agent field for backward compatibility
	agentID := req.AgentID
	
	// If agent_id is empty, try to use agent field instead
	if agentID == "" {
		agentID = req.Agent
	}
	
	if agentID == "" {
		RespondError(w, http.StatusBadRequest, "Agent ID is required (use either agent_id or agent field)")
		return
	}

	// Remove agent from pool
	if err := h.repo.RemoveAgentFromPool(req.PoolID, agentID); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to remove agent from pool")
		return
	}

	// Get updated pool
	pool, err := h.repo.GetByID(req.PoolID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to retrieve updated pool")
		return
	}

	// Return updated pool
	RespondJSON(w, http.StatusOK, pool)
}

// AddWorkspaceRequest represents the request to add a workspace to a pool
type AddWorkspaceRequest struct {
	PoolID      string `json:"pool_id"`
	WorkspaceID string `json:"workspace_id"`
}

// AddWorkspaceToPool handles adding a workspace to a pool
// POST /api/v1/pools/workspaces/add
func (h *AgentPoolHandler) AddWorkspaceToPool(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req AddWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.PoolID == "" {
		RespondError(w, http.StatusBadRequest, "Pool ID is required")
		return
	}

	if req.WorkspaceID == "" {
		RespondError(w, http.StatusBadRequest, "Workspace ID is required")
		return
	}

	// Add workspace to pool
	if err := h.repo.AddWorkspaceToPool(req.PoolID, req.WorkspaceID); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to add workspace to pool")
		return
	}

	// Get updated pool
	pool, err := h.repo.GetByID(req.PoolID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to retrieve updated pool")
		return
	}

	// Return updated pool
	RespondJSON(w, http.StatusOK, pool)
}

// RemoveWorkspaceRequest represents the request to remove a workspace from a pool
type RemoveWorkspaceRequest struct {
	PoolID      string `json:"pool_id"`
	WorkspaceID string `json:"workspace_id"`
}

// RemoveWorkspaceFromPool handles removing a workspace from a pool
// POST /api/v1/pools/workspaces/remove
func (h *AgentPoolHandler) RemoveWorkspaceFromPool(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req RemoveWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.PoolID == "" {
		RespondError(w, http.StatusBadRequest, "Pool ID is required")
		return
	}

	if req.WorkspaceID == "" {
		RespondError(w, http.StatusBadRequest, "Workspace ID is required")
		return
	}

	// Remove workspace from pool
	if err := h.repo.RemoveWorkspaceFromPool(req.PoolID, req.WorkspaceID); err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to remove workspace from pool")
		return
	}

	// Get updated pool
	pool, err := h.repo.GetByID(req.PoolID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to retrieve updated pool")
		return
	}

	// Return updated pool
	RespondJSON(w, http.StatusOK, pool)
}

// GetPoolsByAgent handles retrieving all pools that an agent belongs to
// GET /api/v1/pools/by-agent
func (h *AgentPoolHandler) GetPoolsByAgent(w http.ResponseWriter, r *http.Request) {
	// Extract agent ID from query parameters (support both agent_id and agent for backward compatibility)
	agentID := r.URL.Query().Get("agent_id")
	
	// If agent_id parameter is empty, try to use agent parameter
	if agentID == "" {
		agentID = r.URL.Query().Get("agent")
	}
	
	if agentID == "" {
		RespondError(w, http.StatusBadRequest, "Agent ID is required (use either agent_id or agent parameter)")
		return
	}

	// Get pools from repository
	pools, err := h.repo.GetPoolsByAgent(agentID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to retrieve agent pools")
		return
	}

	// Return pools
	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"pools": pools,
		"total": len(pools),
	})
}

// GetPoolsByWorkspace handles retrieving all pools assigned to a workspace
// GET /api/v1/pools/by-workspace
func (h *AgentPoolHandler) GetPoolsByWorkspace(w http.ResponseWriter, r *http.Request) {
	// Extract workspace ID from query parameters
	workspaceID := r.URL.Query().Get("workspace_id")
	if workspaceID == "" {
		RespondError(w, http.StatusBadRequest, "Workspace ID is required")
		return
	}

	// Get pools from repository
	pools, err := h.repo.GetPoolsByWorkspace(workspaceID)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "Failed to retrieve workspace pools")
		return
	}

	// Return pools
	RespondJSON(w, http.StatusOK, map[string]interface{}{
		"pools": pools,
		"total": len(pools),
	})
}