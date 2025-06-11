package example

import (
	"fmt"
	"strings"
	"entitydb/models"
)

// RelationshipTag provides helper methods for tag-based relationships
type RelationshipTag struct{}

// CreateRelationship adds a relationship tag to an entity
func (rt *RelationshipTag) CreateRelationship(entity *models.Entity, relType, targetID string, metadata map[string]string) {
	// Basic relationship tag
	tag := fmt.Sprintf("relationship:%s:%s", relType, targetID)
	entity.AddTag(tag)
	
	// Add metadata as separate tags if needed
	for key, value := range metadata {
		metaTag := fmt.Sprintf("relationship_meta:%s:%s:%s:%s", relType, targetID, key, value)
		entity.AddTag(metaTag)
	}
}

// GetRelationships extracts all relationships from entity tags
func (rt *RelationshipTag) GetRelationships(entity *models.Entity) []Relationship {
	var relationships []Relationship
	
	// Get clean tags without timestamps
	tags := entity.GetTagsWithoutTimestamp()
	
	for _, tag := range tags {
		if strings.HasPrefix(tag, "relationship:") {
			parts := strings.Split(tag, ":")
			if len(parts) >= 3 {
				rel := Relationship{
					Type:     parts[1],
					TargetID: strings.Join(parts[2:], ":"),
					Metadata: make(map[string]string),
				}
				
				// Look for metadata tags
				metaPrefix := fmt.Sprintf("relationship_meta:%s:%s:", rel.Type, rel.TargetID)
				for _, metaTag := range tags {
					if strings.HasPrefix(metaTag, metaPrefix) {
						metaParts := strings.Split(strings.TrimPrefix(metaTag, metaPrefix), ":")
						if len(metaParts) >= 2 {
							rel.Metadata[metaParts[0]] = strings.Join(metaParts[1:], ":")
						}
					}
				}
				
				relationships = append(relationships, rel)
			}
		}
	}
	
	return relationships
}

// GetRelationshipsByType returns relationships of a specific type
func (rt *RelationshipTag) GetRelationshipsByType(entity *models.Entity, relType string) []string {
	var targetIDs []string
	prefix := fmt.Sprintf("relationship:%s:", relType)
	
	for _, tag := range entity.GetTagsWithoutTimestamp() {
		if strings.HasPrefix(tag, prefix) {
			targetID := strings.TrimPrefix(tag, prefix)
			targetIDs = append(targetIDs, targetID)
		}
	}
	
	return targetIDs
}

// HasRelationship checks if entity has a specific relationship
func (rt *RelationshipTag) HasRelationship(entity *models.Entity, relType, targetID string) bool {
	tag := fmt.Sprintf("relationship:%s:%s", relType, targetID)
	
	for _, entityTag := range entity.GetTagsWithoutTimestamp() {
		if entityTag == tag {
			return true
		}
	}
	
	return false
}

// RemoveRelationship removes a relationship tag
func (rt *RelationshipTag) RemoveRelationship(entity *models.Entity, relType, targetID string) {
	tag := fmt.Sprintf("relationship:%s:%s", relType, targetID)
	metaPrefix := fmt.Sprintf("relationship_meta:%s:%s:", relType, targetID)
	
	// Remove the main relationship tag and any metadata tags
	var newTags []string
	for _, entityTag := range entity.Tags {
		// Extract actual tag from temporal tag
		actualTag := entityTag
		if idx := strings.Index(entityTag, "|"); idx > 0 {
			actualTag = entityTag[idx+1:]
		}
		
		// Keep tag if it's not the relationship or its metadata
		if actualTag != tag && !strings.HasPrefix(actualTag, metaPrefix) {
			newTags = append(newTags, entityTag)
		}
	}
	
	entity.Tags = newTags
}

// Relationship represents a parsed relationship
type Relationship struct {
	Type     string
	TargetID string
	Metadata map[string]string
}

// Example: Security implementation using tag-based relationships
type TagBasedSecurityManager struct {
	entityRepo models.EntityRepository
	rt         *RelationshipTag
}

func (sm *TagBasedSecurityManager) CreateUser(username, password, email string) (*models.Entity, error) {
	// Create user entity
	userID := "user_" + models.GenerateUUID()
	userEntity := &models.Entity{
		ID: userID,
		Tags: []string{
			"type:user",
			"identity:username:" + username,
			"profile:email:" + email,
			"status:active",
		},
	}
	
	// Create user
	if err := sm.entityRepo.Create(userEntity); err != nil {
		return nil, err
	}
	
	// Create credential entity
	credID := "cred_" + models.GenerateUUID()
	// ... create credential entity with hashed password ...
	
	// Add relationship as tag
	sm.rt.CreateRelationship(userEntity, "has_credential", credID, map[string]string{
		"primary": "true",
		"created": models.NowString(),
	})
	
	// Update user entity with relationship
	if err := sm.entityRepo.Update(userEntity); err != nil {
		return nil, err
	}
	
	return userEntity, nil
}

func (sm *TagBasedSecurityManager) AuthenticateUser(username, password string) (*models.Entity, error) {
	// Find user
	users, err := sm.entityRepo.ListByTag("identity:username:" + username)
	if err != nil || len(users) == 0 {
		return nil, fmt.Errorf("user not found")
	}
	
	user := users[0]
	
	// Get credential IDs from relationship tags
	credentialIDs := sm.rt.GetRelationshipsByType(user, "has_credential")
	if len(credentialIDs) == 0 {
		return nil, fmt.Errorf("no credentials found")
	}
	
	// Fetch credential entity and verify password
	for _, credID := range credentialIDs {
		credEntity, err := sm.entityRepo.GetByID(credID)
		if err != nil {
			continue
		}
		
		// Verify password against credential...
		// If successful, return user
	}
	
	return user, nil
}

// Query examples for tag-based relationships
func ExampleQueries(repo models.EntityRepository) {
	// Find all users with admin role
	users, _ := repo.ListByTag("relationship:has_role:role_admin")
	
	// Find all entities that belong to a dataset
	entities, _ := repo.ListByTag("relationship:belongs_to:dataset_123")
	
	// Find all credentials for a user (inverse relationship)
	creds, _ := repo.ListByTag("relationship_inverse:has_credential:user_456")
	
	// Find all relationships of any type to a target
	related, _ := repo.ListByTagWildcard("relationship:*:target_789")
}

// Migration helper to convert existing relationships
func MigrateRelationshipToTags(repo models.EntityRepository, relRepo models.EntityRelationshipRepository) error {
	// Get all relationship entities
	relationships, err := repo.ListByTag("type:relationship")
	if err != nil {
		return err
	}
	
	rt := &RelationshipTag{}
	
	for _, relEntity := range relationships {
		// Parse relationship data
		var sourceID, targetID, relType string
		for _, tag := range relEntity.GetTagsWithoutTimestamp() {
			if strings.HasPrefix(tag, "_source:") {
				sourceID = strings.TrimPrefix(tag, "_source:")
			} else if strings.HasPrefix(tag, "_target:") {
				targetID = strings.TrimPrefix(tag, "_target:")
			} else if strings.HasPrefix(tag, "_relationship:") {
				relType = strings.TrimPrefix(tag, "_relationship:")
			}
		}
		
		if sourceID == "" || targetID == "" || relType == "" {
			continue
		}
		
		// Add relationship tag to source entity
		sourceEntity, err := repo.GetByID(sourceID)
		if err != nil {
			continue
		}
		
		rt.CreateRelationship(sourceEntity, relType, targetID, nil)
		
		// Update source entity
		if err := repo.Update(sourceEntity); err != nil {
			return fmt.Errorf("failed to update source entity %s: %v", sourceID, err)
		}
		
		// Optionally add inverse relationship to target
		if inverseType, ok := getInverseRelationType(relType); ok {
			targetEntity, err := repo.GetByID(targetID)
			if err == nil {
				rt.CreateRelationship(targetEntity, inverseType, sourceID, nil)
				repo.Update(targetEntity)
			}
		}
	}
	
	return nil
}

func getInverseRelationType(relType string) (string, bool) {
	inverseMap := map[string]string{
		"has_credential":   "credential_of",
		"member_of":        "has_member",
		"has_role":         "role_of",
		"parent_of":        "child_of",
		"owns":             "owned_by",
		"created_by":       "created",
		"assigned_to":      "assigned",
		"belongs_to":       "contains",
	}
	
	inverse, ok := inverseMap[relType]
	return inverse, ok
}