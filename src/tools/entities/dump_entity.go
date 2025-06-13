package main

import (
	"encoding/json"
	"entitydb/config"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	
	"entitydb/storage/binary"
)

// Command-line tool to dump entity data
// When compiled, this will be named entitydb_dump

func main() {
	var (
		id           string
		outputFormat string
		outputFile   string
		includeRaw   bool
		timestamps   bool
	)

	flag.StringVar(&id, "id", "", "Entity ID to dump (required)")
	flag.StringVar(&outputFormat, "format", "json", "Output format (json, yaml, or pretty)")
	flag.StringVar(&outputFile, "output", "", "Output file path (default: stdout)")
	flag.BoolVar(&includeRaw, "raw", false, "Include raw binary data")
	flag.BoolVar(&timestamps, "timestamps", true, "Include timestamps in tags")
	flag.Parse()

	// Initialize configuration system
	configManager := config.NewConfigManager(nil)
	configManager.RegisterFlags()
	
	cfg, err := configManager.Initialize()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	if id == "" {
		fmt.Println("Error: Entity ID is required")
		flag.Usage()
		os.Exit(1)
	}

	// Initialize the repository using configured path
	repo, err := binary.NewEntityRepository(cfg.DataPath)
	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}

	// Get the entity
	entity, err := repo.GetByID(id)
	if err != nil {
		log.Fatalf("Failed to get entity: %v", err)
	}

	// Process entity data for output
	var output []byte
	var result map[string]interface{}

	// Convert entity to map
	entityJSON, err := json.Marshal(entity)
	if err != nil {
		log.Fatalf("Failed to marshal entity: %v", err)
	}

	if err := json.Unmarshal(entityJSON, &result); err != nil {
		log.Fatalf("Failed to unmarshal entity to map: %v", err)
	}

	// Process tags
	if !timestamps {
		tags := entity.GetTagsWithoutTimestamp()
		result["tags"] = tags
	}

	// Remove raw data if not requested
	if !includeRaw {
		delete(result, "content")
		
		// Try to decode content based on content type
		for _, tag := range entity.Tags {
			tagValue := tag
			if !timestamps {
				parts := strings.Split(tag, "|")
				if len(parts) > 1 {
					tagValue = parts[len(parts)-1]
				}
			}
			
			if strings.HasPrefix(tagValue, "content:type:") {
				contentType := strings.TrimPrefix(tagValue, "content:type:")
				switch contentType {
				case "json":
					var jsonContent map[string]interface{}
					if err := json.Unmarshal(entity.Content, &jsonContent); err == nil {
						result["decoded_content"] = jsonContent
					}
				case "text/plain":
					result["decoded_content"] = string(entity.Content)
				default:
					result["content_type"] = contentType
					result["content_size"] = len(entity.Content)
				}
				break
			}
		}
	}

	// Generate output according to format
	switch outputFormat {
	case "json":
		output, err = json.Marshal(result)
	case "pretty":
		output, err = json.MarshalIndent(result, "", "  ")
	default:
		log.Fatalf("Unsupported output format: %s", outputFormat)
	}

	if err != nil {
		log.Fatalf("Failed to format output: %v", err)
	}

	// Write output
	if outputFile != "" {
		// Create directory if it doesn't exist
		dir := filepath.Dir(outputFile)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.MkdirAll(dir, 0755); err != nil {
				log.Fatalf("Failed to create directory: %v", err)
			}
		}

		// Write to file
		if err := os.WriteFile(outputFile, output, 0644); err != nil {
			log.Fatalf("Failed to write output: %v", err)
		}
		fmt.Printf("Entity data written to %s\n", outputFile)
	} else {
		// Write to stdout
		fmt.Println(string(output))
	}
}