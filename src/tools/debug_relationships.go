package main

import (
	"encoding/json"
	"entitydb/logger"
	"entitydb/models"
	"entitydb/storage/binary"
	"os"
	"strings"
)

func main() {
	logger.Info("=== Debug Relationships Tool ===")
	
	// Check if data directory exists
	dataDir := "var/"
	if len(os.Args) > 1 {
		dataDir = os.Args[1]
	}
	
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		logger.Error("Data directory does not exist: %s", dataDir)
		os.Exit(1)
	}
	
	// Initialize repository
	repo, err := binary.NewEntityRepository(dataDir)
	if err != nil {
		logger.Error("Failed to create repository: %v", err)
		os.Exit(1)
	}
	
	// Step 1: List all entities
	logger.Info("\n=== Step 1: Listing All Entities ===")
	allEntities, err := repo.List()
	if err != nil {
		logger.Error("Failed to list entities: %v", err)
		os.Exit(1)
	}
	logger.Info("Total entities: %d", len(allEntities))
	
	// Step 2: Find all relationship entities
	logger.Info("\n=== Step 2: Finding Relationship Entities ===")
	relationshipEntities := []*models.Entity{}
	hasCredentialEntities := []*models.Entity{}
	adminUserEntities := []*models.Entity{}
	
	for _, entity := range allEntities {
		hasRelType := false
		hasHasCredential := false
		hasAdminUser := false
		
		logger.Debug("\nEntity ID: %s", entity.ID)
		logger.Debug("Tags (%d):", len(entity.Tags))
		for _, tag := range entity.Tags {
			logger.Debug("  - %s", tag)
			
			// Check for relationship type
			if strings.Contains(tag, "type:relationship") {
				hasRelType = true
			}
			
			// Check for has_credential
			if strings.Contains(tag, "_relationship:has_credential") || 
			   strings.Contains(tag, "relationship:has_credential") {
				hasHasCredential = true
			}
			
			// Check for admin user
			if strings.Contains(tag, "username:admin") {
				hasAdminUser = true
			}
		}
		
		if hasRelType {
			relationshipEntities = append(relationshipEntities, entity)
			logger.Info("Found relationship entity: %s", entity.ID)
		}
		
		if hasHasCredential {
			hasCredentialEntities = append(hasCredentialEntities, entity)
			logger.Info("Found has_credential entity: %s", entity.ID)
		}
		
		if hasAdminUser {
			adminUserEntities = append(adminUserEntities, entity)
			logger.Info("Found admin user entity: %s", entity.ID)
		}
	}
	
	logger.Info("\nTotal relationship entities: %d", len(relationshipEntities))
	logger.Info("Total has_credential relationships: %d", len(hasCredentialEntities))
	logger.Info("Total admin users: %d", len(adminUserEntities))
	
	// Step 3: Analyze has_credential relationships in detail
	logger.Info("\n=== Step 3: Analyzing has_credential Relationships ===")
	for i, entity := range hasCredentialEntities {
		logger.Info("\nhas_credential #%d (ID: %s):", i+1, entity.ID)
		
		var sourceID, targetID, relType string
		for _, tag := range entity.Tags {
			logger.Debug("  Tag: %s", tag)
			
			// Extract source ID
			if strings.Contains(tag, "_source:") {
				parts := strings.SplitN(tag, "_source:", 2)
				if len(parts) == 2 {
					sourceID = parts[1]
				}
			} else if strings.Contains(tag, "source:") {
				parts := strings.SplitN(tag, "source:", 2)
				if len(parts) == 2 {
					sourceID = parts[1]
				}
			}
			
			// Extract target ID
			if strings.Contains(tag, "_target:") {
				parts := strings.SplitN(tag, "_target:", 2)
				if len(parts) == 2 {
					targetID = parts[1]
				}
			} else if strings.Contains(tag, "target:") {
				parts := strings.SplitN(tag, "target:", 2)
				if len(parts) == 2 {
					targetID = parts[1]
				}
			}
			
			// Extract relationship type
			if strings.Contains(tag, "_relationship:") {
				parts := strings.SplitN(tag, "_relationship:", 2)
				if len(parts) == 2 {
					relType = parts[1]
				}
			} else if strings.Contains(tag, "relationship:") {
				parts := strings.SplitN(tag, "relationship:", 2)
				if len(parts) == 2 {
					relType = parts[1]
				}
			}
		}
		
		logger.Info("  Source ID: %s", sourceID)
		logger.Info("  Target ID: %s", targetID)
		logger.Info("  Relationship Type: %s", relType)
		
		// Try to parse content
		if len(entity.Content) > 0 {
			var content map[string]interface{}
			if err := json.Unmarshal(entity.Content, &content); err == nil {
				logger.Info("  Content: %v", content)
			} else {
				logger.Info("  Content (raw): %s", string(entity.Content))
			}
		}
	}
	
	// Step 4: Test GetRelationshipsBySource for each admin user
	logger.Info("\n=== Step 4: Testing GetRelationshipsBySource ===")
	for i, adminUser := range adminUserEntities {
		logger.Info("\nTesting for admin user #%d (ID: %s):", i+1, adminUser.ID)
		
		// Test direct tag search
		searchTag := "_source:" + adminUser.ID
		logger.Info("Searching for tag: %s", searchTag)
		
		entities, err := repo.ListByTag(searchTag)
		if err != nil {
			logger.Error("ListByTag failed: %v", err)
		} else {
			logger.Info("ListByTag returned %d entities", len(entities))
			for _, e := range entities {
				logger.Info("  - Entity ID: %s", e.ID)
			}
		}
		
		// Test GetRelationshipsBySource
		logger.Info("Testing GetRelationshipsBySource(%s):", adminUser.ID)
		relationships, err := repo.GetRelationshipsBySource(adminUser.ID)
		if err != nil {
			logger.Error("GetRelationshipsBySource failed: %v", err)
		} else {
			logger.Info("GetRelationshipsBySource returned %d relationships", len(relationships))
			for _, rel := range relationships {
				if r, ok := rel.(*models.EntityRelationship); ok {
					logger.Info("  - Relationship: %s -> %s (%s)", r.SourceID, r.TargetID, r.RelationshipType)
				}
			}
		}
	}
	
	// Step 5: Check tag index directly
	logger.Info("\n=== Step 5: Checking Tag Index ===")
	
	// We need to check the tag index to see what's actually indexed
	// This requires access to the internal tagIndex field
	logger.Info("Checking for _source: tags in index...")
	
	// Try different tag variations
	tagVariations := []string{
		"_source:",
		"source:",
		"_relationship:",
		"relationship:",
		"type:relationship",
	}
	
	for _, prefix := range tagVariations {
		logger.Info("\nChecking tags with prefix '%s':", prefix)
		count := 0
		
		// Check all entities for tags with this prefix
		for _, entity := range allEntities {
			for _, tag := range entity.Tags {
				if strings.Contains(tag, prefix) {
					count++
					logger.Debug("  Entity %s has tag: %s", entity.ID, tag)
				}
			}
		}
		
		logger.Info("Found %d tags with prefix '%s'", count, prefix)
	}
	
	logger.Info("\n=== Debug Complete ===")
}