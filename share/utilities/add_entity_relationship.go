package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run add_entity_relationship.go <source_id> <relationship_type> <target_id> [metadata_key1=value1 metadata_key2=value2 ...]")
		fmt.Println("Example: go run add_entity_relationship.go entity_123 depends_on entity_456 dependency_type=blocker description=\"Entity 123 depends on Entity 456\"")
		fmt.Println("\nAvailable relationship types:")
		fmt.Println("  depends_on       - The source entity depends on the target entity")
		fmt.Println("  blocks           - The source entity blocks the target entity")
		fmt.Println("  parent_of        - The source entity is a parent of the target entity")
		fmt.Println("  child_of         - The source entity is a child of the target entity")
		fmt.Println("  related_to       - The source entity is related to the target entity")
		fmt.Println("  duplicate_of     - The source entity is a duplicate of the target entity")
		fmt.Println("  assigned_to      - The source entity is assigned to the target entity")
		fmt.Println("  belongs_to       - The source entity belongs to the target entity")
		fmt.Println("  linked_to        - The source entity is linked to the target entity")
		os.Exit(1)
	}

	sourceID := os.Args[1]
	relationshipType := os.Args[2]
	targetID := os.Args[3]
	
	// Parse optional metadata
	metadata := make(map[string]string)
	if len(os.Args) > 4 {
		for _, metadataPair := range os.Args[4:] {
			key, value := parseKeyValue(metadataPair)
			if key != "" {
				metadata[key] = value
			} else {
				fmt.Printf("Warning: Ignoring invalid metadata format '%s'. Expected format: key=value\n", metadataPair)
			}
		}
	}

	// Connect to database
	db, err := sql.Open("sqlite3", "/opt/entitydb/var/db/entitydb.db")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Verify that source and target entities exist
	if !entityExists(db, sourceID) {
		log.Fatalf("Source entity %s does not exist", sourceID)
	}
	
	if !entityExists(db, targetID) {
		log.Fatalf("Target entity %s does not exist", targetID)
	}
	
	// Check if relationship already exists
	var relationshipExists bool
	err = db.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM entity_relationships 
			WHERE source_id = ? AND relationship_type = ? AND target_id = ?
		)
	`, sourceID, relationshipType, targetID).Scan(&relationshipExists)
	
	if err != nil {
		log.Fatalf("Error checking for existing relationship: %v", err)
	}
	
	if relationshipExists {
		fmt.Printf("Relationship already exists. Updating metadata...\n")
		updateRelationshipMetadata(db, sourceID, relationshipType, targetID, metadata)
		fmt.Println("Relationship metadata updated successfully")
		return
	}

	// Create the relationship
	metadataJSON := "{}"
	if len(metadata) > 0 {
		metadataBytes, err := json.Marshal(metadata)
		if err != nil {
			log.Fatalf("Failed to marshal metadata: %v", err)
		}
		metadataJSON = string(metadataBytes)
	}
	
	_, err = db.Exec(`
		INSERT INTO entity_relationships 
		(source_id, relationship_type, target_id, created_at, created_by, metadata) 
		VALUES (?, ?, ?, ?, ?, ?)
	`, sourceID, relationshipType, targetID, time.Now().Format(time.RFC3339), "system", metadataJSON)
	
	if err != nil {
		log.Fatalf("Failed to create relationship: %v", err)
	}

	fmt.Printf("Relationship created successfully:\n")
	fmt.Printf("  Source: %s\n", sourceID)
	fmt.Printf("  Type: %s\n", relationshipType)
	fmt.Printf("  Target: %s\n", targetID)
	if len(metadata) > 0 {
		fmt.Printf("  Metadata:\n")
		for key, value := range metadata {
			fmt.Printf("    %s: %s\n", key, value)
		}
	}
}

// entityExists checks if an entity exists in the database
func entityExists(db *sql.DB, entityID string) bool {
	var exists bool
	db.QueryRow("SELECT EXISTS(SELECT 1 FROM entities WHERE id = ?)", entityID).Scan(&exists)
	return exists
}

// updateRelationshipMetadata updates the metadata of an existing relationship
func updateRelationshipMetadata(db *sql.DB, sourceID, relationshipType, targetID string, newMetadata map[string]string) error {
	// Get current metadata
	var metadataJSON string
	err := db.QueryRow(`
		SELECT metadata FROM entity_relationships 
		WHERE source_id = ? AND relationship_type = ? AND target_id = ?
	`, sourceID, relationshipType, targetID).Scan(&metadataJSON)
	
	if err != nil {
		return fmt.Errorf("error fetching relationship metadata: %v", err)
	}
	
	// Parse current metadata
	currentMetadata := make(map[string]interface{})
	if metadataJSON != "" && metadataJSON != "{}" {
		err = json.Unmarshal([]byte(metadataJSON), &currentMetadata)
		if err != nil {
			return fmt.Errorf("error parsing relationship metadata: %v", err)
		}
	}
	
	// Update with new metadata
	for key, value := range newMetadata {
		currentMetadata[key] = value
	}
	
	// Marshal the updated metadata
	updatedMetadataBytes, err := json.Marshal(currentMetadata)
	if err != nil {
		return fmt.Errorf("error marshaling updated metadata: %v", err)
	}
	
	// Update the relationship
	_, err = db.Exec(`
		UPDATE entity_relationships SET metadata = ? 
		WHERE source_id = ? AND relationship_type = ? AND target_id = ?
	`, string(updatedMetadataBytes), sourceID, relationshipType, targetID)
	
	if err != nil {
		return fmt.Errorf("error updating relationship metadata: %v", err)
	}
	
	return nil
}

// parseKeyValue parses a key=value string
// Handles quoted values for strings with spaces
func parseKeyValue(input string) (string, string) {
	for i := 0; i < len(input); i++ {
		if input[i] == '=' {
			key := input[:i]
			value := input[i+1:]
			
			// Handle quoted values
			if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
				value = value[1 : len(value)-1]
			}
			
			return key, value
		}
	}
	
	return "", ""
}