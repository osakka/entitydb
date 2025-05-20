package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"entitydb/models"
)

// IssueHandler manages issue-related API endpoints
type IssueHandler struct {
	repo models.IssueRepository
	auth *Auth
}

// NewIssueHandler creates a new issue handler
func NewIssueHandler(repo models.IssueRepository, auth *Auth) *IssueHandler {
	return &IssueHandler{
		repo: repo,
		auth: auth,
	}
}


// CreateIssueRequest represents the request to create an issue
type CreateIssueRequest struct {
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	Priority        string    `json:"priority"`
	Type            string    `json:"type"`
	ParentID        string    `json:"parent_id"`
	EstimatedEffort float64   `json:"estimated_effort"`
	DueDate         time.Time `json:"due_date"`
	WorkspaceID     string    `json:"workspace_id"`
	Tags            []string  `json:"tags"`
}

// CreateIssue handles issue creation
// POST /api/v1/issues
func (h *IssueHandler) CreateIssue(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req CreateIssueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if req.Title == "" {
		RespondError(w, http.StatusBadRequest, "Title is required")
		return
	}

	// Make workspace ID optional with a default for backwards compatibility
	if req.WorkspaceID == "" {
		// Try to find the default workspace by ID directly
		defaultWorkspace, err := h.repo.GetByID("workspace_entitydb")
		if err == nil && defaultWorkspace != nil {
			log.Printf("Using default 'workspace_entitydb' for issue creation")
			req.WorkspaceID = defaultWorkspace.ID
		} else {
			// Try to find by title as a fallback
			defaultWorkspace, err := h.repo.GetByTitle("EntityDB Default Workspace")
			if err == nil && defaultWorkspace != nil {
				log.Printf("Found default workspace by title for issue creation")
				req.WorkspaceID = defaultWorkspace.ID
			} else {
				// Log the error and return a more helpful error message
				log.Printf("No default workspace found: %v", err)
				RespondError(w, http.StatusBadRequest, "No default workspace found. Please specify a workspace_id.")
				return
			}
		}
	}

	// Set default priority if not provided
	priority := req.Priority
	if priority == "" {
		priority = models.IssuePriorityMedium
	} else {
		// Validate priority
		validPriorities := []string{models.IssuePriorityHigh, models.IssuePriorityMedium, models.IssuePriorityLow}
		isValidPriority := false
		for _, p := range validPriorities {
			if priority == p {
				isValidPriority = true
				break
			}
		}

		if !isValidPriority {
			RespondError(w, http.StatusBadRequest, "Invalid priority. Must be one of: high, medium, low")
			return
		}
	}

	// Get creator ID from authenticated user
	var creatorID string
	var userID string
	var userName string

	// First try the standard authentication flow
	user, err := h.auth.GetAuthenticatedUser(r)
	if err != nil {
		// Handle unauthenticated requests
		log.Printf("Warning: Unauthenticated request to create issue: %v", err)

		// Check if we have a test or development environment bypass header
		devAgent := r.Header.Get("X-EntityDB-Agent-ID")
		if devAgent != "" {
			log.Printf("Using development agent ID from header: %s", devAgent)
			creatorID = devAgent

			// For audit purposes, log that this was a development bypass
			userID = "development"
			userName = "development_user"
		} else {
			// Return unauthorized error - require authentication
			RespondError(w, http.StatusUnauthorized, "Authentication required to create issues")
			return
		}
	} else {
		// Store authenticated user information for audit and tracking
		userID = user.ID
		userName = user.Username

		// Handle authenticated users
		// Try to get the agent ID from authentication context first
		agentID, err := h.auth.GetAuthenticatedAgentID(r)
		if err == nil && agentID != "" {
			// Got agent ID from auth context - use it
			log.Printf("Creating issue with authenticated agent ID: %s (User: %s)", agentID, user.Username)
			creatorID = agentID
		} else if user.AgentID != "" {
			// Use user's linked agent ID if available
			log.Printf("Creating issue with user's linked agent ID: %s (User: %s)", user.AgentID, user.Username)
			creatorID = user.AgentID
		} else {
			// User doesn't have an agent ID - return error
			log.Printf("No agent ID available for user %s", user.Username)
			RespondError(w, http.StatusBadRequest, "User account must be linked to an agent to create issues")
			return
		}
	}

	// Validate issue type if provided
	issueType := models.IssueTypeIssue // Default to standard issue
	if req.Type != "" {
		// Validate issue type
		validTypes := map[string]models.IssueType{
			"epic":      models.IssueTypeEpic,
			"story":     models.IssueTypeStory,
			"issue":     models.IssueTypeIssue,
			"subissue":  models.IssueTypeSubissue,
			"workspace": "workspace", // New workspace type
		}
		
		if val, ok := validTypes[req.Type]; ok {
			issueType = val
		} else {
			RespondError(w, http.StatusBadRequest, "Invalid issue type. Must be one of: epic, story, issue, subissue, workspace")
			return
		}
	}
	
	// Handle different issue types appropriately
	var issue *models.Issue
	var issueErr error
	
	switch issueType {
	case models.IssueTypeEpic:
		// Create epic directly using repository method
		issue, issueErr = h.repo.CreateEpic(
			req.Title,
			req.Description,
			priority,
			creatorID,
			req.WorkspaceID,
		)
		if issueErr != nil {
			RespondError(w, http.StatusInternalServerError, "Failed to create epic: "+issueErr.Error())
			return
		}

		// Add user context for better tracking and auditing
		if userID != "" {
			issue.CreatedByUserID = userID
			issue.CreatedByUsername = userName

			// We need to update the issue to save the user context
			if updateErr := h.repo.Update(issue); updateErr != nil {
				log.Printf("Warning: Failed to update epic with user context: %v", updateErr)
			}
		}
	case models.IssueTypeStory:
		// Stories must have a parent epic
		if req.ParentID == "" {
			RespondError(w, http.StatusBadRequest, "Parent epic ID is required for stories")
			return
		}
		
		// Create story using repository method
		issue, issueErr = h.repo.CreateStory(
			req.Title,
			req.Description,
			priority,
			creatorID,
			req.WorkspaceID,
			req.ParentID,
		)
		if issueErr != nil {
			RespondError(w, http.StatusInternalServerError, "Failed to create story: "+issueErr.Error())
			return
		}

		// Add user context for better tracking and auditing
		if userID != "" {
			issue.CreatedByUserID = userID
			issue.CreatedByUsername = userName

			// We need to update the issue to save the user context
			if updateErr := h.repo.Update(issue); updateErr != nil {
				log.Printf("Warning: Failed to update story with user context: %v", updateErr)
			}
		}
	default:
		// Create standard issue or subissue
		issue = models.NewIssue(
			req.Title,
			req.Description,
			priority,
			issueType,
			creatorID, // Agent ID
			req.WorkspaceID,
			req.ParentID) // Use provided parent ID

		// Add user context for better tracking and auditing
		if userID != "" {
			issue.CreatedByUserID = userID
			issue.CreatedByUsername = userName
		}

		// For standard issues, we need to save it using the repo
		log.Printf("About to save issue: Title=%s, Type=%s, CreatedBy=%s, WorkspaceID=%s, UserID=%s",
			issue.Title, issue.Type, issue.CreatedBy, issue.WorkspaceID, userID)
		if createErr := h.repo.Create(issue); createErr != nil {
			log.Printf("Error creating issue: %v", createErr)
			RespondError(w, http.StatusInternalServerError, "Failed to create issue: "+createErr.Error())
			return
		}
		log.Printf("Issue successfully persisted with ID: %s by user %s", issue.ID, userName)
	}

	// Set optional fields
	if req.EstimatedEffort > 0 {
		issue.EstimatedEffort = req.EstimatedEffort
	}

	if !req.DueDate.IsZero() {
		issue.DueDate = req.DueDate
	}

	if len(req.Tags) > 0 {
		log.Printf("DEBUG: Setting tags for issue ID=%s: %v", issue.ID, req.Tags)
		issue.Tags = req.Tags
	} else {
		log.Printf("DEBUG: No tags provided in request for issue ID=%s", issue.ID)
	}

	// Always update the issue to save tags and other optional fields
	// regardless of issue type (epic, story, issue, subissue)
	log.Printf("DEBUG: About to update issue ID=%s with Tags=%v", issue.ID, issue.Tags)
	
	// Additional validation before update
	if issue.ID == "" {
		log.Printf("ERROR: Cannot update issue with empty ID")
		RespondError(w, http.StatusInternalServerError, "Cannot update issue with empty ID")
		return
	}
	
	// Check for any invalid tag data
	if len(issue.Tags) > 0 {
		hasEmptyTags := false
		for i, tag := range issue.Tags {
			if tag == "" {
				hasEmptyTags = true
				log.Printf("WARNING: Issue has empty tag at index %d, will be skipped", i)
			}
		}
		
		// Filter out empty tags to avoid issues during insertion
		if hasEmptyTags {
			log.Printf("DEBUG: Filtering out empty tags before updating")
			filteredTags := make([]string, 0, len(issue.Tags))
			for _, tag := range issue.Tags {
				if tag != "" {
					filteredTags = append(filteredTags, tag)
				}
			}
			issue.Tags = filteredTags
			log.Printf("DEBUG: After filtering, issue has %d tags: %v", len(issue.Tags), issue.Tags)
		}
	}
	
	// Attempt to update the issue with enhanced error handling
	if updateErr := h.repo.Update(issue); updateErr != nil {
		log.Printf("ERROR: Failed to update issue with optional fields: %v", updateErr)
		
		// Try to provide more specific error messages based on the error
		if updateErr.Error() == "issue_tags table does not exist in the database" {
			RespondError(w, http.StatusInternalServerError, "Tag storage system not available - contact administrator")
			return
		} else if updateErr.Error() == fmt.Sprintf("cannot handle tags for non-existent issue %s", issue.ID) {
			RespondError(w, http.StatusInternalServerError, "Cannot save tags: Issue does not exist in database")
			return
		} else {
			RespondError(w, http.StatusInternalServerError, "Failed to update issue with optional fields: "+updateErr.Error())
			return
		}
	}
	
	log.Printf("DEBUG: Successfully updated issue ID=%s with optional fields", issue.ID)

	// Return the created issue
	RespondJSON(w, http.StatusCreated, issue)
}

// The rest of the file remains the same...