//go:build tool
package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type EntityRelationship struct {
	SourceID         string    `json:"source_id"`
	RelationshipType string    `json:"relationship_type"`
	TargetID         string    `json:"target_id"`
	CreatedAt        time.Time `json:"created_at"`
	CreatedBy        string    `json:"created_by,omitempty"`
	Metadata         string    `json:"metadata,omitempty"`
}

type Entity struct {
	ID      string        `json:"id"`
	Tags    []string      `json:"tags"`
	Content []ContentItem `json:"content"`
}

type ContentItem struct {
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
	Value     string `json:"value"`
}

func main() {
	// Parse command line arguments
	var sourceID, targetID, relationshipType string
	var showMetadata bool = true

	// Process arguments
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if strings.HasPrefix(arg, "--source=") {
			sourceID = strings.TrimPrefix(arg, "--source=")
		} else if strings.HasPrefix(arg, "--target=") {
			targetID = strings.TrimPrefix(arg, "--target=")
		} else if strings.HasPrefix(arg, "--type=") {
			relationshipType = strings.TrimPrefix(arg, "--type=")
		} else if arg == "--no-metadata" {
			showMetadata = false
		} else if arg == "--help" || arg == "-h" {
			printUsage()
			os.Exit(0)
		}
	}

	// Require at least one filter
	if sourceID == "" && targetID == "" && relationshipType == "" {
		fmt.Println("Error: At least one filter is required (--source, --target, or --type)")
		printUsage()
		os.Exit(1)
	}

	// Connect to database
	db, err := sql.Open("sqlite3", "/opt/entitydb/var/db/entitydb.db")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Build the query
	query := `
		SELECT 
			er.source_id, 
			er.relationship_type, 
			er.target_id, 
			er.created_at, 
			er.created_by, 
			er.metadata,
			s.tags AS source_tags,
			s.content AS source_content,
			t.tags AS target_tags,
			t.content AS target_content
		FROM entity_relationships er
		JOIN entities s ON er.source_id = s.id
		JOIN entities t ON er.target_id = t.id
		WHERE 1=1
	`
	var conditions []string
	var params []interface{}

	if sourceID != "" {
		conditions = append(conditions, "er.source_id = ?")
		params = append(params, sourceID)
	}

	if targetID != "" {
		conditions = append(conditions, "er.target_id = ?")
		params = append(params, targetID)
	}

	if relationshipType != "" {
		conditions = append(conditions, "er.relationship_type = ?")
		params = append(params, relationshipType)
	}

	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY er.created_at DESC"

	// Execute the query
	rows, err := db.Query(query, params...)
	if err != nil {
		log.Fatalf("Failed to query relationships: %v", err)
	}
	defer rows.Close()

	// Print the results
	fmt.Println("Entity Relationships:")
	fmt.Println("=====================")

	count := 0
	for rows.Next() {
		count++
		var rel EntityRelationship
		var createdAtStr, createdBy, metadata sql.NullString
		var sourceTagsJSON, sourceContentJSON, targetTagsJSON, targetContentJSON string
		
		err := rows.Scan(
			&rel.SourceID,
			&rel.RelationshipType,
			&rel.TargetID,
			&createdAtStr,
			&createdBy,
			&metadata,
			&sourceTagsJSON,
			&sourceContentJSON,
			&targetTagsJSON,
			&targetContentJSON,
		)
		if err != nil {
			log.Fatalf("Error scanning row: %v", err)
		}

		// Parse created at
		if createdAtStr.Valid {
			createdAt, err := time.Parse(time.RFC3339, createdAtStr.String)
			if err != nil {
				log.Printf("Warning: Failed to parse created_at for relationship: %v", err)
			} else {
				rel.CreatedAt = createdAt
			}
		}

		// Parse created by
		if createdBy.Valid {
			rel.CreatedBy = createdBy.String
		} else {
			rel.CreatedBy = "system"
		}

		// Parse metadata
		if metadata.Valid {
			rel.Metadata = metadata.String
		}

		// Parse source and target entity details
		sourceEntity := parseEntity(rel.SourceID, sourceTagsJSON, sourceContentJSON)
		targetEntity := parseEntity(rel.TargetID, targetTagsJSON, targetContentJSON)

		// Print relationship details
		fmt.Printf("\nRelationship:\n")
		fmt.Printf("  Source: %s (%s) - %s\n", 
			sourceEntity.ID, 
			getTagValue(sourceEntity.Tags, "type"),
			getContentValueByType(sourceEntity.Content, "title"),
		)
		fmt.Printf("  Type: %s\n", rel.RelationshipType)
		fmt.Printf("  Target: %s (%s) - %s\n", 
			targetEntity.ID, 
			getTagValue(targetEntity.Tags, "type"),
			getContentValueByType(targetEntity.Content, "title"),
		)
		fmt.Printf("  Created: %s by %s\n", rel.CreatedAt.Format(time.RFC3339), rel.CreatedBy)

		// Print metadata if requested
		if showMetadata && rel.Metadata != "" && rel.Metadata != "{}" {
			fmt.Println("  Metadata:")
			var metadataMap map[string]interface{}
			err := json.Unmarshal([]byte(rel.Metadata), &metadataMap)
			if err != nil {
				log.Printf("Warning: Failed to parse metadata: %v", err)
			} else {
				for key, value := range metadataMap {
					fmt.Printf("    %s: %v\n", key, value)
				}
			}
		}
	}

	if count == 0 {
		fmt.Println("No relationships found.")
	} else {
		fmt.Printf("\nTotal relationships: %d\n", count)
	}
}

// printUsage prints the usage information
func printUsage() {
	fmt.Println("Usage: go run list_entity_relationships.go [options]")
	fmt.Println("Options:")
	fmt.Println("  --source=ID       Filter relationships by source entity ID")
	fmt.Println("  --target=ID       Filter relationships by target entity ID")
	fmt.Println("  --type=TYPE       Filter relationships by type")
	fmt.Println("  --no-metadata     Do not show relationship metadata")
	fmt.Println("  --help, -h        Show this help message")
	fmt.Println("\nNote: At least one filter is required (--source, --target, or --type)")
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
}

// parseEntity parses an entity from JSON
func parseEntity(id, tagsJSON, contentJSON string) Entity {
	entity := Entity{ID: id}

	// Parse tags
	err := json.Unmarshal([]byte(tagsJSON), &entity.Tags)
	if err != nil {
		log.Printf("Warning: Failed to parse tags for entity %s: %v", id, err)
		entity.Tags = []string{}
	}

	// Parse content
	err = json.Unmarshal([]byte(contentJSON), &entity.Content)
	if err != nil {
		log.Printf("Warning: Failed to parse content for entity %s: %v", id, err)
		entity.Content = []ContentItem{}
	}

	return entity
}

// getTagValue extracts the value of a tag from a list of tags
func getTagValue(tags []string, tagName string) string {
	prefix := tagName + ":"
	for _, tag := range tags {
		if strings.HasPrefix(tag, prefix) {
			return strings.TrimPrefix(tag, prefix)
		}
	}
	return ""
}

// getContentValueByType extracts the value of a content item by type
func getContentValueByType(content []ContentItem, contentType string) string {
	var latestTimestamp, latestValue string
	
	for _, item := range content {
		if item.Type == contentType && item.Timestamp > latestTimestamp {
			latestTimestamp = item.Timestamp
			latestValue = item.Value
		}
	}
	
	return latestValue
}