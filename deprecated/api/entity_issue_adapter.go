package api

import (
	"entitydb/models"
	"fmt"
	"strings"
	"time"
)

// EntityIssueAdapter provides bidirectional conversion between Entity and Issue models
// to enable seamless integration of the entity-based architecture.

// ConvertIssueToEntity converts an Issue to an Entity
func ConvertIssueToEntity(issue *models.Issue) *models.Entity {
	// If issue ID is empty, generate a new one
	id := issue.ID
	if id == "" {
		id = models.GenerateID("ent")
	}
	
	entity := models.NewEntity(id)
	
	// Add content items
	entity.AddContent("title", issue.Title)
	entity.AddContent("description", issue.Description)
	
	// Add tags for issue fields
	// Use both formats for tags to ensure compatibility
	// 1. Timestamp format: YYYY-MM-DDTHH:MM:SS.nanos.tag=value
	// 2. Simple format: tag:value
	
	// First, extract type and status tags that might be in the tags array
	issueType := string(issue.Type)
	if issueType == "" {
		issueType = "issue" // Default
	}
	
	status := issue.Status
	if status == "" {
		status = "pending" // Default
	}
	
	// Add core tags
	entity.AddTag("type", issueType)
	entity.AddTag("status", status)
	entity.AddTag("priority", issue.Priority)
	
	// Add relationship tags
	if issue.WorkspaceID != "" {
		entity.AddTag("workspace", issue.WorkspaceID)
	}
	
	if issue.ParentID != "" {
		entity.AddTag("parent", issue.ParentID)
	}
	
	if issue.CreatedBy != "" {
		entity.AddTag("created_by", issue.CreatedBy)
	}
	
	// Add timestamp as tag
	createdAt := issue.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	entity.AddTag("created_at", createdAt.Format(time.RFC3339))
	
	// Add numeric fields as tags
	if issue.EstimatedEffort > 0 {
		entity.AddTag("effort", fmt.Sprintf("%.2f", issue.EstimatedEffort))
	}
	
	if issue.Progress > 0 {
		entity.AddTag("progress", fmt.Sprintf("%d", issue.Progress))
	}
	
	if !issue.DueDate.IsZero() {
		entity.AddTag("due_date", issue.DueDate.Format(time.RFC3339))
	}
	
	// Add child metrics as tags
	if issue.ChildCount > 0 {
		entity.AddTag("child_count", fmt.Sprintf("%d", issue.ChildCount))
	}
	
	if issue.ChildCompleted > 0 {
		entity.AddTag("child_completed", fmt.Sprintf("%d", issue.ChildCompleted))
	}
	
	// Copy additional tags from the issue
	// Skip tags that we've already explicitly added
	for _, tag := range issue.Tags {
		if strings.HasPrefix(tag, "type:") || 
		   strings.HasPrefix(tag, "status:") || 
		   strings.HasPrefix(tag, "priority:") {
			continue // Skip these as we've already handled them
		}
		
		// Add the tag directly for backward compatibility
		entity.Tags = append(entity.Tags, tag)
	}
	
	// Handle assignment if present
	if issue.Assignment != nil {
		// Add assignment details as content
		assignmentData := fmt.Sprintf(`{
			"id": "%s",
			"agent_id": "%s",
			"assigned_at": "%s",
			"assigned_by": "%s",
			"status": "%s",
			"progress": %d
		}`, 
		issue.Assignment.ID,
		issue.Assignment.AgentID,
		issue.Assignment.AssignedAt.Format(time.RFC3339),
		issue.Assignment.AssignedBy,
		issue.Assignment.Status,
		issue.Assignment.Progress)
		
		entity.AddContent("assignment", assignmentData)
		
		// Add assigned_to tag for filtering
		entity.AddTag("assigned_to", issue.Assignment.AgentID)
		
		// Add assignment status
		assignmentStatus := issue.Assignment.Status
		if assignmentStatus != "" {
			entity.AddTag("assignment_status", assignmentStatus)
		}
	}
	
	return entity
}

// ConvertEntityToIssue converts an Entity to an Issue
func ConvertEntityToIssue(entity *models.Entity) *models.Issue {
	// Extract basic issue data from entity
	
	// Get title and description from content
	title := ""
	description := ""
	
	titleValues := entity.GetContentByType("title")
	if len(titleValues) > 0 {
		title = titleValues[0]
	}
	
	descValues := entity.GetContentByType("description")
	if len(descValues) > 0 {
		description = descValues[0]
	}
	
	// Extract tag values (check both formats)
	// 1. Check timestamp format (.tag=value or .tag.)
	// 2. Check simple format (tag:value)
	
	// Extract type
	issueType := extractTagValue(entity, "type")
	if issueType == "" {
		issueType = "issue" // Default
	}
	
	// Extract status
	status := extractTagValue(entity, "status")
	if status == "" {
		status = "pending" // Default
	}
	
	// Extract priority
	priority := extractTagValue(entity, "priority")
	if priority == "" {
		priority = "medium" // Default
	}
	
	// Extract relationships
	workspaceID := extractTagValue(entity, "workspace")
	parentID := extractTagValue(entity, "parent")
	createdBy := extractTagValue(entity, "created_by")
	
	// Extract numeric fields
	estimatedEffort := 0.0
	effortStr := extractTagValue(entity, "effort")
	if effortStr != "" {
		fmt.Sscanf(effortStr, "%f", &estimatedEffort)
	}
	
	progress := 0
	progressStr := extractTagValue(entity, "progress")
	if progressStr != "" {
		fmt.Sscanf(progressStr, "%d", &progress)
	}
	
	childCount := 0
	childCountStr := extractTagValue(entity, "child_count")
	if childCountStr != "" {
		fmt.Sscanf(childCountStr, "%d", &childCount)
	}
	
	childCompleted := 0
	childCompletedStr := extractTagValue(entity, "child_completed")
	if childCompletedStr != "" {
		fmt.Sscanf(childCompletedStr, "%d", &childCompleted)
	}
	
	// Extract timestamps
	createdAt := time.Now()
	createdAtStr := extractTagValue(entity, "created_at")
	if createdAtStr != "" {
		parsedTime, err := time.Parse(time.RFC3339, createdAtStr)
		if err == nil {
			createdAt = parsedTime
		}
	}
	
	dueDate := time.Time{}
	dueDateStr := extractTagValue(entity, "due_date")
	if dueDateStr != "" {
		parsedTime, err := time.Parse(time.RFC3339, dueDateStr)
		if err == nil {
			dueDate = parsedTime
		}
	}
	
	// Create the issue
	issue := &models.Issue{
		ID:              entity.ID,
		Title:           title,
		Description:     description,
		Priority:        priority,
		EstimatedEffort: estimatedEffort,
		DueDate:         dueDate,
		CreatedAt:       createdAt,
		CreatedBy:       createdBy,
		WorkspaceID:     workspaceID,
		ParentID:        parentID,
		Progress:        progress,
		ChildCount:      childCount,
		ChildCompleted:  childCompleted,
	}
	
	// Add type and status tags
	issue.Type = models.IssueType(issueType)
	issue.Status = status
	
	// Extract assignment
	assignmentData := entity.GetContentByType("assignment")
	if len(assignmentData) > 0 {
		// In a real implementation, we would parse the JSON data
		// For this proof of concept, we'll create a minimal assignment
		agentID := extractTagValue(entity, "assigned_to")
		if agentID != "" {
			issue.Assignment = &models.IssueAssignment{
				ID:         models.GenerateID("asgn"),
				IssueID:    entity.ID,
				AgentID:    agentID,
				AssignedAt: time.Now(),
				AssignedBy: createdBy,
				Progress:   0,
			}
			
			// Set assignment status if available
			assignmentStatus := extractTagValue(entity, "assignment_status")
			if assignmentStatus != "" {
				issue.Assignment.Status = assignmentStatus
			} else {
				issue.Assignment.Status = "pending"
			}
		}
	}
	
	// Copy relevant tags (not including the ones we've already added)
	isRelevantTag := func(tag string) bool {
		// Skip tags we've already handled
		return !strings.HasPrefix(tag, "type:") &&
			   !strings.HasPrefix(tag, "status:") &&
			   !strings.HasPrefix(tag, "priority:") &&
			   !strings.HasPrefix(tag, "workspace:") &&
			   !strings.HasPrefix(tag, "parent:") &&
			   !strings.HasPrefix(tag, "created_by:") &&
			   !strings.HasPrefix(tag, "created_at:") &&
			   !strings.HasPrefix(tag, "due_date:") &&
			   !strings.HasPrefix(tag, "effort:") &&
			   !strings.HasPrefix(tag, "progress:") &&
			   !strings.HasPrefix(tag, "child_count:") &&
			   !strings.HasPrefix(tag, "child_completed:") &&
			   !strings.HasPrefix(tag, "assigned_to:") &&
			   !strings.HasPrefix(tag, "assignment_status:") &&
			   !strings.Contains(tag, ".type=") &&
			   !strings.Contains(tag, ".status=") &&
			   !strings.Contains(tag, ".priority=") &&
			   !strings.Contains(tag, ".workspace=") &&
			   !strings.Contains(tag, ".parent=") &&
			   !strings.Contains(tag, ".created_by=") &&
			   !strings.Contains(tag, ".created_at=") &&
			   !strings.Contains(tag, ".due_date=") &&
			   !strings.Contains(tag, ".effort=") &&
			   !strings.Contains(tag, ".progress=") &&
			   !strings.Contains(tag, ".child_count=") &&
			   !strings.Contains(tag, ".child_completed=") &&
			   !strings.Contains(tag, ".assigned_to=") &&
			   !strings.Contains(tag, ".assignment_status=")
	}
	
	for _, tag := range entity.Tags {
		if isRelevantTag(tag) {
			issue.Tags = append(issue.Tags, tag)
		}
	}
	
	return issue
}

// Helper function to extract tag value from an entity
// Checks both timestamp and simple formats
func extractTagValue(entity *models.Entity, tagName string) string {
	// First try simple format (tag:value)
	simplePrefix := tagName + ":"
	for _, tag := range entity.Tags {
		if strings.HasPrefix(tag, simplePrefix) {
			return strings.TrimPrefix(tag, simplePrefix)
		}
	}
	
	// Then try timestamp format (.tag=value)
	timestampSuffix := "." + tagName + "="
	for _, tag := range entity.Tags {
		if strings.Contains(tag, timestampSuffix) {
			parts := strings.SplitN(tag, timestampSuffix, 2)
			if len(parts) == 2 {
				return parts[1]
			}
		}
	}
	
	return ""
}

// ConvertIssueListToEntityList converts a list of issues to a list of entities
func ConvertIssueListToEntityList(issues []*models.Issue) []*models.Entity {
	var entities []*models.Entity
	
	for _, issue := range issues {
		entity := ConvertIssueToEntity(issue)
		entities = append(entities, entity)
	}
	
	return entities
}

// ConvertEntityListToIssueList converts a list of entities to a list of issues
func ConvertEntityListToIssueList(entities []*models.Entity) []*models.Issue {
	var issues []*models.Issue
	
	for _, entity := range entities {
		issue := ConvertEntityToIssue(entity)
		issues = append(issues, issue)
	}
	
	return issues
}

// ConvertIssueFilterToEntityFilter converts an issue filter to an entity filter
func ConvertIssueFilterToEntityFilter(filter map[string]interface{}) map[string]interface{} {
	entityFilter := make(map[string]interface{})

	// Convert common filters
	if val, ok := filter["type"]; ok {
		entityFilter["tag"] = "type:" + val.(string)
	}

	if val, ok := filter["status"]; ok {
		entityFilter["tag"] = "status:" + val.(string)
	}

	if val, ok := filter["priority"]; ok {
		entityFilter["tag"] = "priority:" + val.(string)
	}

	if val, ok := filter["workspace_id"]; ok {
		entityFilter["tag"] = "workspace:" + val.(string)
	}

	if val, ok := filter["parent_id"]; ok {
		entityFilter["tag"] = "parent:" + val.(string)
	}

	if val, ok := filter["agent_id"]; ok {
		entityFilter["tag"] = "assigned_to:" + val.(string)
	}

	return entityFilter
}

// ConvertIssueDependencyToEntityRelationship converts an issue dependency to an entity relationship
func ConvertIssueDependencyToEntityRelationship(dependency *models.IssueDependency) *models.EntityRelationship {
	// Create a new relationship
	relationship := models.NewEntityRelationship(
		dependency.IssueID,
		models.RelationshipTypeDependsOn,
		dependency.DependsOnID,
	)

	// Set creation info
	relationship.CreatedAt = dependency.CreatedAt
	relationship.SetCreatedBy(dependency.CreatedBy)

	// Add metadata if available
	if dependency.DependencyType != "" || dependency.Description != "" {
		metadata := map[string]interface{}{
			"dependency_type": dependency.DependencyType,
			"description":     dependency.Description,
		}
		relationship.AddMetadata(metadata)
	}

	return relationship
}

// ConvertEntityRelationshipToIssueDependency converts an entity relationship to an issue dependency
func ConvertEntityRelationshipToIssueDependency(relationship *models.EntityRelationship) *models.IssueDependency {
	// Create a new dependency
	dependency := &models.IssueDependency{
		ID:          models.GenerateID("dep"),
		IssueID:     relationship.SourceID,
		DependsOnID: relationship.TargetID,
		CreatedAt:   relationship.CreatedAt,
		CreatedBy:   relationship.CreatedBy,
	}

	// Extract metadata if available
	if relationship.Metadata != "" {
		metadata, err := relationship.GetMetadata()
		if err == nil {
			if depType, ok := metadata["dependency_type"].(string); ok {
				dependency.DependencyType = depType
			}
			if desc, ok := metadata["description"].(string); ok {
				dependency.Description = desc
			}
		}
	}

	return dependency
}

// ConvertIssueDependenciesToEntityRelationships converts a list of issue dependencies to entity relationships
func ConvertIssueDependenciesToEntityRelationships(dependencies []*models.IssueDependency) []*models.EntityRelationship {
	var relationships []*models.EntityRelationship

	for _, dependency := range dependencies {
		relationship := ConvertIssueDependencyToEntityRelationship(dependency)
		relationships = append(relationships, relationship)
	}

	return relationships
}

// ConvertEntityRelationshipsToIssueDependencies converts a list of entity relationships to issue dependencies
func ConvertEntityRelationshipsToIssueDependencies(relationships []*models.EntityRelationship) []*models.IssueDependency {
	var dependencies []*models.IssueDependency

	for _, relationship := range relationships {
		// Only convert dependency relationships
		if relationship.RelationshipType == models.RelationshipTypeDependsOn {
			dependency := ConvertEntityRelationshipToIssueDependency(relationship)
			dependencies = append(dependencies, dependency)
		}
	}

	return dependencies
}