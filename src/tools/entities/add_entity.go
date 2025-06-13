//go:build tool
package main

import (
	"crypto/rand"
	"database/sql"
	"entitydb/config"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Helper function to generate a unique ID with the given prefix
func generateID(prefix string) string {
	// Generate a random 16-byte value
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	
	// Format as a hexadecimal string
	id := fmt.Sprintf("%x", b)
	
	// Add prefix if provided
	if prefix != "" {
		return prefix + "_" + id
	}
	
	return id
}

func main() {
	// Initialize configuration system
	configManager := config.NewConfigManager(nil)
	configManager.RegisterFlags()
	
	var (
		entityType = flag.String("type", "", "Entity type (required)")
		title = flag.String("title", "", "Entity title (required)")
		tags = flag.String("tags", "", "Comma-separated tags in format tag1:value1,tag2:value2")
	)
	flag.Parse()
	
	cfg, err := configManager.Initialize()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}
	
	if *entityType == "" || *title == "" {
		fmt.Println("Usage: add_entity --type=<type> --title=<title> [--tags=tag1:value1,tag2:value2]")
		fmt.Println("Example: add_entity --type=workspace --title=\"New Workspace\" --tags=status:active,owner:admin")
		os.Exit(1)
	}
	
	// Parse optional tags
	tagMap := make(map[string]string)
	if *tags != "" {
		for _, tagPair := range strings.Split(*tags, ",") {
			tagPair = strings.TrimSpace(tagPair)
			parts := strings.SplitN(tagPair, ":", 2)
			if len(parts) == 2 {
				tagMap[parts[0]] = parts[1]
			} else {
				fmt.Printf("Warning: Ignoring invalid tag format '%s'. Expected format: tag:value\n", tagPair)
			}
		}
	}

	// Connect to database using configured path
	dbPath := cfg.DatabasePath()
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Generate a unique entity ID
	entityID := generateID("ent")

	// Insert the entity
	_, err = db.Exec(`
		INSERT INTO entities (id, tags, content) 
		VALUES (?, ?, ?)
	`, entityID, "", "[]")

	if err != nil {
		log.Fatalf("Failed to insert entity: %v", err)
	}

	// Add type tag
	addTagToEntity(db, entityID, "type", *entityType)
	
	// Add title content
	addContentToEntity(db, entityID, "title", *title)
	
	// Add additional tags
	for tag, value := range tagMap {
		addTagToEntity(db, entityID, tag, value)
	}

	fmt.Printf("Entity created successfully:\n")
	fmt.Printf("  ID: %s\n", entityID)
	fmt.Printf("  Type: %s\n", *entityType)
	fmt.Printf("  Title: %s\n", *title)
	if len(tagMap) > 0 {
		fmt.Printf("  Tags:\n")
		for tag, value := range tagMap {
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
	timestamp := time.Now().UTC().Format(time.RFC3339Nano)
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