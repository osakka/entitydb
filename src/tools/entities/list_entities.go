package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

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
	var tagFilter, typeFilter, idFilter string
	var showContent, showTags bool

	// Default to showing all entities
	showContent = true
	showTags = true

	// Process arguments
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if strings.HasPrefix(arg, "--tag=") {
			tagFilter = strings.TrimPrefix(arg, "--tag=")
		} else if strings.HasPrefix(arg, "--type=") {
			typeFilter = strings.TrimPrefix(arg, "--type=")
		} else if strings.HasPrefix(arg, "--id=") {
			idFilter = strings.TrimPrefix(arg, "--id=")
		} else if arg == "--no-content" {
			showContent = false
		} else if arg == "--no-tags" {
			showTags = false
		} else if arg == "--help" || arg == "-h" {
			printUsage()
			os.Exit(0)
		}
	}

	// Connect to database
	db, err := sql.Open("sqlite3", "/opt/entitydb/var/db/entitydb.db")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Build the query
	query := "SELECT id, tags, content FROM entities"
	var conditions []string
	var params []interface{}

	if idFilter != "" {
		conditions = append(conditions, "id = ?")
		params = append(params, idFilter)
	}

	if typeFilter != "" {
		conditions = append(conditions, "tags LIKE ?")
		params = append(params, fmt.Sprintf("%%\"type:%s\"%%", typeFilter))
	}

	if tagFilter != "" {
		tagParts := strings.SplitN(tagFilter, ":", 2)
		if len(tagParts) == 2 {
			conditions = append(conditions, "tags LIKE ?")
			params = append(params, fmt.Sprintf("%%%s:%s%%", tagParts[0], tagParts[1]))
		} else {
			conditions = append(conditions, "tags LIKE ?")
			params = append(params, fmt.Sprintf("%%%s%%", tagFilter))
		}
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY id"

	// Execute the query
	rows, err := db.Query(query, params...)
	if err != nil {
		log.Fatalf("Failed to query entities: %v", err)
	}
	defer rows.Close()

	// Print the results
	fmt.Println("Entities:")
	fmt.Println("=========")

	count := 0
	for rows.Next() {
		count++
		var id, tagsJSON, contentJSON string
		err := rows.Scan(&id, &tagsJSON, &contentJSON)
		if err != nil {
			log.Fatalf("Error scanning row: %v", err)
		}

		// Parse tags and content
		var tags []string
		var content []ContentItem

		err = json.Unmarshal([]byte(tagsJSON), &tags)
		if err != nil {
			log.Printf("Warning: Failed to parse tags for entity %s: %v", id, err)
			tags = []string{}
		}

		err = json.Unmarshal([]byte(contentJSON), &content)
		if err != nil {
			log.Printf("Warning: Failed to parse content for entity %s: %v", id, err)
			content = []ContentItem{}
		}

		// Print entity details
		fmt.Printf("\nEntity ID: %s\n", id)

		// Extract and print title
		title := getContentValueByType(content, "title")
		if title != "" {
			fmt.Printf("Title: %s\n", title)
		}

		// Print entity type
		entityType := getTagValue(tags, "type")
		if entityType != "" {
			fmt.Printf("Type: %s\n", entityType)
		}

		// Print tags if requested
		if showTags && len(tags) > 0 {
			fmt.Println("Tags:")
			for _, tag := range tags {
				// Skip the tags that follow the timestamp format
				if !strings.Contains(tag, ".") || !strings.Contains(tag, "=") {
					fmt.Printf("  %s\n", tag)
				}
			}
		}

		// Print content if requested
		if showContent && len(content) > 0 {
			fmt.Println("Content:")
			contentMap := groupContentByType(content)
			
			for contentType, values := range contentMap {
				fmt.Printf("  %s:\n", contentType)
				for _, value := range values {
					fmt.Printf("    %s\n", value)
				}
			}
		}
	}

	if count == 0 {
		fmt.Println("No entities found.")
	} else {
		fmt.Printf("\nTotal entities: %d\n", count)
	}
}

// printUsage prints the usage information
func printUsage() {
	fmt.Println("Usage: go run list_entities.go [options]")
	fmt.Println("Options:")
	fmt.Println("  --tag=TAG         Filter entities by tag (e.g. --tag=status:active)")
	fmt.Println("  --type=TYPE       Filter entities by type (e.g. --type=workspace)")
	fmt.Println("  --id=ID           Filter entities by ID")
	fmt.Println("  --no-content      Do not show entity content")
	fmt.Println("  --no-tags         Do not show entity tags")
	fmt.Println("  --help, -h        Show this help message")
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

// groupContentByType groups content values by their type
func groupContentByType(content []ContentItem) map[string][]string {
	result := make(map[string][]string)
	
	for _, item := range content {
		result[item.Type] = append(result[item.Type], item.Value)
	}
	
	return result
}