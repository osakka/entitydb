package api

import (
	"entitydb/models"
	"entitydb/logger"
	"net/http"
	"time"
)

// DashboardHandler handles dashboard-related API requests
type DashboardHandler struct {
	entityRepo *models.RepositoryQueryWrapper
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(entityRepo models.EntityRepository) *DashboardHandler {
	return &DashboardHandler{
		entityRepo: models.NewRepositoryQueryWrapper(entityRepo),
	}
}

// DashboardStatsResponse represents dashboard statistics
type DashboardStatsResponse struct {
	AgentStats     AgentStats     `json:"agent_stats"`
	IssueStats     IssueStats     `json:"issue_stats"`
	WorkspaceCount int            `json:"workspace_count"`
	UserCount      int            `json:"user_count"`
	RecentActivity []ActivityItem `json:"recent_activity"`
}

type AgentStats struct {
	Total  int `json:"total"`
	Active int `json:"active"`
}

type IssueStats struct {
	Total     int            `json:"total"`
	ByStatus  map[string]int `json:"by_status"`
	ByPriority map[string]int `json:"by_priority"`
}

type ActivityItem struct {
	Timestamp   time.Time `json:"timestamp"`
	Type        string    `json:"type"`
	Description string    `json:"description"`
}

// DashboardStats returns statistics for the dashboard
// @Summary Dashboard statistics
// @Description Get dashboard statistics and metrics
// @Tags dashboard
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} DashboardStatsResponse
// @Router /api/v1/dashboard/stats [get]
func (h *DashboardHandler) DashboardStats(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Handling dashboard stats request")
	
	// Count agents
	agents, err := h.entityRepo.Query().
		HasTag("type:agent").
		Execute()
	if err != nil {
		logger.Debug("Error getting agents: %v", err)
		http.Error(w, "Error retrieving agent data", http.StatusInternalServerError)
		return
	}
	
	activeAgents := 0
	for _, agent := range agents {
		for _, tag := range agent.Tags {
			if tag == "status:active" {
				activeAgents++
				break
			}
		}
	}
	
	// Count issues
	issues, err := h.entityRepo.Query().
		HasTag("type:issue").
		Execute()
	if err != nil {
		logger.Debug("Error getting issues: %v", err)
		http.Error(w, "Error retrieving issue data", http.StatusInternalServerError)
		return
	}
	
	// Analyze issue statistics
	issueByStatus := make(map[string]int)
	issueByPriority := make(map[string]int)
	
	for _, issue := range issues {
		for _, tag := range issue.Tags {
			if len(tag) > 7 && tag[:7] == "status:" {
				status := tag[7:]
				issueByStatus[status]++
			}
			if len(tag) > 9 && tag[:9] == "priority:" {
				priority := tag[9:]
				issueByPriority[priority]++
			}
		}
	}
	
	// Count workspaces
	workspaces, err := h.entityRepo.Query().
		HasTag("type:workspace").
		Execute()
	if err != nil {
		logger.Debug("Error getting workspaces: %v", err)
		workspaces = []*models.Entity{}
	}
	
	// Count users
	users, err := h.entityRepo.Query().
		HasTag("type:user").
		Execute()
	if err != nil {
		logger.Debug("Error getting users: %v", err)
		users = []*models.Entity{}
	}
	
	// Get recent activity (last 10 entities)
	recentEntities, err := h.entityRepo.Query().
		Limit(10).
		OrderBy("created_at", "desc").
		Execute()
	
	recentActivity := []ActivityItem{}
	if err == nil {
		for _, entity := range recentEntities {
			entityType := "unknown"
			for _, tag := range entity.Tags {
				if len(tag) > 5 && tag[:5] == "type:" {
					entityType = tag[5:]
					break
				}
			}
			
			createdAt, _ := time.Parse(time.RFC3339, entity.CreatedAt)
			activity := ActivityItem{
				Timestamp:   createdAt,
				Type:        entityType,
				Description: "Created " + entityType + " " + entity.ID,
			}
			recentActivity = append(recentActivity, activity)
		}
	}
	
	// Build response
	response := DashboardStatsResponse{
		AgentStats: AgentStats{
			Total:  len(agents),
			Active: activeAgents,
		},
		IssueStats: IssueStats{
			Total:      len(issues),
			ByStatus:   issueByStatus,
			ByPriority: issueByPriority,
		},
		WorkspaceCount: len(workspaces),
		UserCount:      len(users),
		RecentActivity: recentActivity,
	}
	
	RespondJSON(w, http.StatusOK, response)
}

