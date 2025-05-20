package main

import (
	"entitydb/models"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run add_entity.go <type> <title> [tag1:value1 tag2:value2 ...]")
		fmt.Println("Example: go run add_entity.go workspace \"New Workspace\" status:active owner:admin")
		os.Exit(1)
	}

	entityType := os.Args[1]
	title := os.Args[2]
	
	// Parse optional tags
	tags := make(map[string]string)
	if len(os.Args) > 3 {
		for _, tagPair := range os.Args[3:] {
			parts := strings.SplitN(tagPair, ":", 2)
			if len(parts) == 2 {
				tags[parts[0]] = parts[1]
			} else {
				fmt.Printf("Warning: Ignoring invalid tag format '%s'. Expected format: tag:value\n", tagPair)
			}
		}
	}

	// Connect to database
	db, err := sql.Open("sqlite3", "/opt/entitydb/var/db/entitydb.db")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Generate a unique entity ID
	entityID := models.GenerateID("ent")

	// Insert the entity
	_, err = db.Exec(`
		INSERT INTO entities (id, tags, content) 
		VALUES (?, ?, ?)
	`, entityID, "", "[]")

	if err != nil {
		log.Fatalf("Failed to insert entity: %v", err)
	}

	// Add type tag
	addTagToEntity(db, entityID, "type", entityType)
	
	// Add title content
	addContentToEntity(db, entityID, "title", title)
	
	// Add additional tags
	for tag, value := range tags {
		addTagToEntity(db, entityID, tag, value)
	}

	fmt.Printf("Entity created successfully:\n")
	fmt.Printf("  ID: %s\n", entityID)
	fmt.Printf("  Type: %s\n", entityType)
	fmt.Printf("  Title: %s\n", title)
	if len(tags) > 0 {
		fmt.Printf("  Tags:\n")
		for tag, value := range tags {
			fmt.Printf("    %s: %s\n", tag, value)
		}
	}
}

// addTagToEntity adds a tag to an entity in the database
func addTagToEntity(db *sql.DB, entityID, tag, value string) error {
	// First get current tags
	var tagsJSON string
	err := db.QueryRow("SELECT tags FROM entities WHERE id = ?", entityID).Scan(&tagsJSON)
	if err != nil {
		return fmt.Errorf("error fetching entity tags: %v", err)
	}
	
	// Format the tag
	newTag := fmt.Sprintf("%s:%s", tag, value)
	
	// Append the tag (simple implementation without parsing JSON)
	if tagsJSON == "" {
		tagsJSON = "[]"
	}
	
	// Remove last bracket, add the tag, and close the array
	if tagsJSON == "[]" {
		tagsJSON = fmt.Sprintf("[\"%s\"]", newTag)
	} else {
		tagsJSON = fmt.Sprintf("%s, \"%s\"]", tagsJSON[:len(tagsJSON)-1], newTag)
	}
	
	// Update the entity
	_, err = db.Exec("UPDATE entities SET tags = ? WHERE id = ?", tagsJSON, entityID)
	if err != nil {
		return fmt.Errorf("error updating entity tags: %v", err)
	}
	
	return nil
}

// addContentToEntity adds content to an entity in the database
func addContentToEntity(db *sql.DB, entityID, contentType, contentValue string) error {
	// First get current content
	var contentJSON string
	err := db.QueryRow("SELECT content FROM entities WHERE id = ?", entityID).Scan(&contentJSON)
	if err != nil {
		return fmt.Errorf("error fetching entity content: %v", err)
	}
	
	// Format the content with timestamp
	timestamp := models.GenerateTimestamp()
	newContentItem := fmt.Sprintf(`{"timestamp":"%s","type":"%s","value":"%s"}`, 
		timestamp, contentType, contentValue)
	
	// Append the content (simple implementation without parsing JSON)
	if contentJSON == "" {
		contentJSON = "[]"
	}
	
	// Remove last bracket, add the content, and close the array
	if contentJSON == "[]" {
		contentJSON = fmt.Sprintf("[%s]", newContentItem)
	} else {
		contentJSON = fmt.Sprintf("%s, %s]", contentJSON[:len(contentJSON)-1], newContentItem)
	}
	
	// Update the entity
	_, err = db.Exec("UPDATE entities SET content = ? WHERE id = ?", contentJSON, entityID)
	if err != nil {
		return fmt.Errorf("error updating entity content: %v", err)
	}
	
	return nil
}