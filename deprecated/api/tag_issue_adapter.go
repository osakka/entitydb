package api

import (
	"entitydb/models"
	"fmt"
	"strings"
	"time"
)

// TagIssueAdapter provides adapter functions to convert between tag-based and explicit field-based issue models
// This helps maintain backward compatibility during the transition to tag-based architecture

// ConvertFilterToTagExpression converts traditional filter parameters to a tag expression
func ConvertFilterToTagExpression(filter map[string]interface{}) string {
	var expressions []string
	
	// Convert type filter
	if typeVal, ok := filter["type"].(string); ok && typeVal != "" {
		expressions = append(expressions, fmt.Sprintf("type:%s", typeVal))
	}
	
	// Convert status filter
	if statusVal, ok := filter["status"].(string); ok && statusVal != "" {
		expressions = append(expressions, fmt.Sprintf("status:%s", statusVal))
	}
	
	// Convert priority filter
	if priorityVal, ok := filter["priority"].(string); ok && priorityVal != "" {
		expressions = append(expressions, fmt.Sprintf("priority:%s", priorityVal))
	}
	
	// Convert workspaceID filter
	if wsVal, ok := filter["workspace_id"].(string); ok && wsVal != "" {
		expressions = append(expressions, fmt.Sprintf("workspace:%s", wsVal))
	}
	
	// Convert parentID filter
	if parentVal, ok := filter["parent_id"].(string); ok && parentVal != "" {
		expressions = append(expressions, fmt.Sprintf("parent:%s", parentVal))
	}
	
	// Convert agent filter (handle differently since it's not a direct tag)
	if agentVal, ok := filter["agent_id"].(string); ok && agentVal != "" {
		expressions = append(expressions, fmt.Sprintf("assigned:%s", agentVal))
	}
	
	// Join expressions with AND operator
	if len(expressions) == 0 {
		return ""
	}
	return strings.Join(expressions, " AND ")
}

// ExtractIssueData converts an issue with tags to a response with explicit fields
func ExtractIssueData(issue *models.Issue) map[string]interface{} {
	// Base fields that don't depend on tags
	data := map[string]interface{}{
		"id":              issue.ID,
		"title":           issue.Title,
		"description":     issue.Description,
		"priority":        issue.Priority,
		"estimated_effort": issue.EstimatedEffort,
		"created_at":      issue.CreatedAt,
		"created_by":      issue.CreatedBy,
		"workspace_id":    issue.WorkspaceID,
		"parent_id":       issue.ParentID,
		"tags":            issue.Tags,
		"child_count":     issue.ChildCount,
		"child_completed": issue.ChildCompleted,
		"progress":        issue.Progress,
	}
	
	// Get issue type
	issueType := issue.Type
	if issueType == "" {
		// Default to standard issue if no type tag exists
		issueType = "issue"
	}
	data["type"] = issueType
	
	// Get issue status
	status := issue.Status
	if status == "" {
		// Default to pending if no status tag exists
		status = "pending"
	}
	data["status"] = status
	
	// Add assignment data if present
	if issue.Assignment != nil {
		// Use Assignment.Status directly
		assignmentStatus := issue.Assignment.Status
		if assignmentStatus == "" {
			assignmentStatus = "pending"
		}

		data["assignment"] = map[string]interface{}{
			"id":           issue.Assignment.ID,
			"agent_id":     issue.Assignment.AgentID,
			"assigned_at":  issue.Assignment.AssignedAt,
			"assigned_by":  issue.Assignment.AssignedBy,
			"status":       assignmentStatus,
			"progress":     issue.Assignment.Progress,
			"started_at":   issue.Assignment.StartedAt,
			"completed_at": issue.Assignment.CompletedAt,
		}
	}
	
	return data
}

// CreateIssueFromRequest creates a new issue with tags from a traditional request
func CreateIssueFromRequest(req CreateIssueRequest, createdBy string) *models.Issue {
	// Create basic issue with the proper fields
	issue := &models.Issue{
		ID:          models.GenerateID("issue"),
		Title:       req.Title,
		Description: req.Description,
		Type:        models.IssueType(req.Type),
		Status:      "pending", // Default status
		Priority:    req.Priority,
		CreatedBy:   createdBy,
		CreatedAt:   time.Now(),
		WorkspaceID: req.WorkspaceID, // Correctly use WorkspaceID from the request
		ParentID:    req.ParentID,    // Correctly use ParentID from the request
		Tags:        req.Tags,
	}

	return issue
}

// UpdateIssueFromRequest updates an issue with tag-based architecture from a traditional request
func UpdateIssueFromRequest(issue *models.Issue, req map[string]interface{}) {
	// Update basic fields
	if title, ok := req["title"].(string); ok && title != "" {
		issue.Title = title
	}
	
	if description, ok := req["description"].(string); ok {
		issue.Description = description
	}
	
	if priority, ok := req["priority"].(string); ok && priority != "" {
		issue.Priority = priority

		// Add priority tag
		addOrReplaceTag(issue, "priority:"+priority, "priority:")
	}

	if workspaceID, ok := req["workspace_id"].(string); ok && workspaceID != "" {
		issue.WorkspaceID = workspaceID
	}

	if parentID, ok := req["parent_id"].(string); ok {
		issue.ParentID = parentID
	}

	// Update type and status directly
	if issueType, ok := req["type"].(string); ok && issueType != "" {
		issue.Type = models.IssueType(issueType)

		// Add type tag
		addOrReplaceTag(issue, "type:"+issueType, "type:")
	}

	if status, ok := req["status"].(string); ok && status != "" {
		issue.Status = status

		// Add status tag
		addOrReplaceTag(issue, "status:"+status, "status:")
	}
	
	// Update tags if provided
	if tags, ok := req["tags"].([]string); ok {
		issue.Tags = tags
	}
}

// Helper function to add or replace a tag with a specific prefix
func addOrReplaceTag(issue *models.Issue, newTag, prefix string) {
	// First check if we already have a tag with this prefix
	replaced := false
	for i, tag := range issue.Tags {
		if strings.HasPrefix(tag, prefix) {
			// Replace this tag
			issue.Tags[i] = newTag
			replaced = true
			break
		}
	}

	// If we didn't replace anything, add the new tag
	if !replaced {
		issue.Tags = append(issue.Tags, newTag)
	}
}