package binary

import (
	"entitydb/models"
	"entitydb/logger"
	"fmt"
	"strings"
	"encoding/json"
)

// RelationshipRepository provides binary storage for entity relationships
type RelationshipRepository struct {
	entityRepo models.EntityRepository
}

// NewRelationshipRepository creates a new relationship repository
func NewRelationshipRepository(entityRepo models.EntityRepository) *RelationshipRepository {
	return &RelationshipRepository{
		entityRepo: entityRepo,
	}
}

// Create creates a new relationship
func (r *RelationshipRepository) Create(rel *models.EntityRelationship) error {
	// Validate required fields
	if rel.SourceID == "" || rel.TargetID == "" || rel.RelationshipType == "" {
		return fmt.Errorf("invalid relationship: missing required fields")
	}
	
	// Generate ID if not provided
	if rel.ID == "" {
		rel.ID = "rel_" + models.GenerateUUID()
	}
	
	// Set creation time
	if rel.CreatedAt == 0 {
		rel.CreatedAt = models.Now()
	}
	
	// Set default creator
	if rel.CreatedBy == "" {
		rel.CreatedBy = "system"
	}
	
	// Create relationship as entity
	entity := r.relationshipToEntity(rel)
	
	// Store in entity repository
	err := r.entityRepo.Create(entity)
	if err != nil {
		return fmt.Errorf("failed to create relationship entity: %w", err)
	}
	
	return nil
}

// GetByID gets a relationship by ID
func (r *RelationshipRepository) GetByID(id string) (*models.EntityRelationship, error) {
	entity, err := r.entityRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get relationship entity: %w", err)
	}
	
	// Check if this is actually a relationship entity
	isRelationship := false
	for _, tag := range entity.Tags {
		if tag == "type:relationship" {
			isRelationship = true
			break
		}
	}
	
	if !isRelationship {
		return nil, fmt.Errorf("entity %s is not a relationship", id)
	}
	
	return r.entityToRelationship(entity)
}

// GetBySourceID gets all relationships for a source entity
func (r *RelationshipRepository) GetBySource(sourceID string) ([]*models.EntityRelationship, error) {
	// Query entities with _source tag (underscore prefix for consistency)
	searchTag := "_source:" + sourceID
	logger.Debug("RelationshipRepository.GetBySource: searching for tag '%s'", searchTag)
	
	entities, err := r.entityRepo.ListByTag(searchTag)
	if err != nil {
		return nil, fmt.Errorf("failed to query relationships by source: %w", err)
	}
	
	logger.Debug("RelationshipRepository.GetBySource: found %d entities with tag '%s'", len(entities), searchTag)
	
	// Convert to relationships
	relationships := make([]*models.EntityRelationship, 0, len(entities))
	for _, entity := range entities {
		logger.Debug("RelationshipRepository.GetBySource: processing entity %s with tags: %v", entity.ID, entity.Tags)
		
		// Only process relationship entities
		isRelationship := false
		for _, tag := range entity.Tags {
			if tag == "type:relationship" {
				isRelationship = true
				break
			}
		}
		
		if isRelationship {
			if rel, err := r.entityToRelationship(entity); err == nil {
				relationships = append(relationships, rel)
				logger.Debug("RelationshipRepository.GetBySource: successfully converted entity %s to relationship", entity.ID)
			} else {
				logger.Debug("RelationshipRepository.GetBySource: failed to convert entity %s to relationship: %v", entity.ID, err)
			}
		} else {
			logger.Debug("RelationshipRepository.GetBySource: entity %s is not a relationship (missing type:relationship tag)", entity.ID)
		}
	}
	
	logger.Debug("RelationshipRepository.GetBySource: returning %d relationships for source %s", len(relationships), sourceID)
	return relationships, nil
}

// GetBySourceAndType gets all relationships of a given type where entity is the source
func (r *RelationshipRepository) GetBySourceAndType(sourceID, relationshipType string) ([]*models.EntityRelationship, error) {
	// Get all entities with both source tag and relationship type tag
	entities, err := r.entityRepo.ListByTags([]string{
		"type:relationship",
		fmt.Sprintf("_source:%s", sourceID),
		fmt.Sprintf("_relationship:%s", relationshipType),
	}, true)
	if err != nil {
		return nil, fmt.Errorf("failed to query relationships: %w", err)
	}
	
	relationships := make([]*models.EntityRelationship, 0, len(entities))
	for _, entity := range entities {
		rel, err := r.entityToRelationship(entity)
		if err != nil {
			// Skip invalid relationships
			continue
		}
		relationships = append(relationships, rel)
	}
	
	return relationships, nil
}

// GetByTarget gets all relationships for a target entity
func (r *RelationshipRepository) GetByTarget(targetID string) ([]*models.EntityRelationship, error) {
	// Query entities with _target tag (underscore prefix for consistency)
	entities, err := r.entityRepo.ListByTag("_target:" + targetID)
	if err != nil {
		return nil, fmt.Errorf("failed to query relationships by target: %w", err)
	}
	
	// Convert to relationships
	relationships := make([]*models.EntityRelationship, 0, len(entities))
	for _, entity := range entities {
		// Only process relationship entities
		isRelationship := false
		for _, tag := range entity.Tags {
			if tag == "type:relationship" {
				isRelationship = true
				break
			}
		}
		
		if isRelationship {
			if rel, err := r.entityToRelationship(entity); err == nil {
				relationships = append(relationships, rel)
			}
		}
	}
	
	return relationships, nil
}

// GetByTargetAndType gets all relationships of a given type where entity is the target
func (r *RelationshipRepository) GetByTargetAndType(targetID, relationshipType string) ([]*models.EntityRelationship, error) {
	// Get all entities with both target tag and relationship type tag
	entities, err := r.entityRepo.ListByTags([]string{
		"type:relationship",
		fmt.Sprintf("_target:%s", targetID),
		fmt.Sprintf("_relationship:%s", relationshipType),
	}, true)
	if err != nil {
		return nil, fmt.Errorf("failed to query relationships: %w", err)
	}
	
	relationships := make([]*models.EntityRelationship, 0, len(entities))
	for _, entity := range entities {
		rel, err := r.entityToRelationship(entity)
		if err != nil {
			// Skip invalid relationships
			continue
		}
		relationships = append(relationships, rel)
	}
	
	return relationships, nil
}

// GetByType gets all relationships of a specific type
func (r *RelationshipRepository) GetByType(relType string) ([]*models.EntityRelationship, error) {
	// Query entities with _relationship tag (underscore prefix for consistency)
	entities, err := r.entityRepo.ListByTag("_relationship:" + relType)
	if err != nil {
		return nil, fmt.Errorf("failed to query relationships by type: %w", err)
	}
	
	// Convert to relationships
	relationships := make([]*models.EntityRelationship, 0, len(entities))
	for _, entity := range entities {
		// Only process relationship entities
		isRelationship := false
		for _, tag := range entity.Tags {
			if tag == "type:relationship" {
				isRelationship = true
				break
			}
		}
		
		if isRelationship {
			if rel, err := r.entityToRelationship(entity); err == nil {
				relationships = append(relationships, rel)
			}
		}
	}
	
	return relationships, nil
}

// Delete deletes a relationship
func (r *RelationshipRepository) Delete(sourceID, relationshipType, targetID string) error {
	// Construct the ID from the source, type, and target
	id := sourceID + "_" + relationshipType + "_" + targetID
	// Verify it's a relationship before deleting
	rel, err := r.GetByID(id)
	if err != nil {
		return err
	}
	
	// Delete the entity
	err = r.entityRepo.Delete(rel.ID)
	if err != nil {
		return fmt.Errorf("failed to delete relationship entity: %w", err)
	}
	
	return nil
}

// GetRelationship gets a specific relationship
func (r *RelationshipRepository) GetRelationship(sourceID, relationshipType, targetID string) (*models.EntityRelationship, error) {
	// Construct the ID from the source, type, and target
	id := sourceID + "_" + relationshipType + "_" + targetID
	return r.GetByID(id)
}

// Exists checks if a relationship exists
func (r *RelationshipRepository) Exists(sourceID, relationshipType, targetID string) (bool, error) {
	// Construct the ID from the source, type, and target
	id := sourceID + "_" + relationshipType + "_" + targetID
	
	// Try to get the relationship
	_, err := r.GetByID(id)
	if err != nil {
		if err == models.ErrNotFound {
			return false, nil
		}
		return false, err
	}
	
	return true, nil
}

func (r *RelationshipRepository) relationshipToEntity(rel *models.EntityRelationship) *models.Entity {
	entity := &models.Entity{
		ID:        rel.ID,
		Tags:      []string{},
		CreatedAt: models.Now(),
		UpdatedAt: models.Now(),
	}
	
	// Add relationship tags - use underscore prefix for consistency with entity_repository.go
	entity.AddTag("type:relationship")
	entity.AddTag("_relationship:" + rel.RelationshipType)
	entity.AddTag("_source:" + rel.SourceID)
	entity.AddTag("_target:" + rel.TargetID)
	entity.AddTagWithValue("created_by", rel.CreatedBy)
	entity.AddTagWithValue("created_at", fmt.Sprintf("%d", rel.CreatedAt))
	
	if rel.Metadata != "" {
		entity.AddTagWithValue("metadata", rel.Metadata)
	}
	
	// Store metadata as content
	contentData := map[string]interface{}{
		"created_at": rel.CreatedAt,
		"metadata":   rel.Metadata,
	}
	jsonData, _ := json.Marshal(contentData)
	entity.Content = jsonData
	entity.AddTag("content:type:relationship")
	
	return entity
}

func (r *RelationshipRepository) entityToRelationship(entity *models.Entity) (*models.EntityRelationship, error) {
	rel := &models.EntityRelationship{
		ID: entity.ID,
	}
	
	// Extract from tags first
	for _, tag := range entity.Tags {
		parts := strings.Split(tag, ":")
		if len(parts) < 2 {
			continue
		}
		
		key := parts[0]
		value := strings.Join(parts[1:], ":")
		
		switch key {
		case "_source":
			rel.SourceID = value
		case "_target":
			rel.TargetID = value
		case "_relationship":
			rel.RelationshipType = value
		case "created_by":
			rel.CreatedBy = value
		}
	}
	
	// Extract from content (new model stores as JSON)
	if len(entity.Content) > 0 {
		var contentData map[string]interface{}
		if err := json.Unmarshal(entity.Content, &contentData); err == nil {
			if createdAt, ok := contentData["created_at"].(int64); ok {
				rel.CreatedAt = createdAt
			} else if createdAtFloat, ok := contentData["created_at"].(float64); ok {
				rel.CreatedAt = int64(createdAtFloat)
			}
			if metadata, ok := contentData["metadata"].(string); ok {
				rel.Metadata = metadata
			}
		}
	}
	
	// Validate required fields
	if rel.SourceID == "" || rel.TargetID == "" || rel.RelationshipType == "" {
		return nil, fmt.Errorf("invalid relationship entity: missing required fields")
	}
	
	return rel, nil
}