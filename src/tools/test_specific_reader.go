package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	
	"entitydb/storage/binary"
	"entitydb/logger"
)

func main() {
	// Initialize logger
	logger.SetLevel(logger.DEBUG)
	
	// Get data path
	dataPath := os.Getenv("ENTITYDB_DATA_PATH")
	if dataPath == "" {
		dataPath = "var"
	}
	
	dataFile := filepath.Join(dataPath, "entities.ebf")
	fmt.Printf("Opening file: %s\n", dataFile)
	
	// Check file
	stat, err := os.Stat(dataFile)
	if err != nil {
		log.Fatalf("Failed to stat file: %v", err)
	}
	fmt.Printf("File size: %d bytes\n", stat.Size())
	
	// Create reader
	reader, err := binary.NewReader(dataFile)
	if err != nil {
		log.Fatalf("Failed to create reader: %v", err)
	}
	defer reader.Close()
	
	fmt.Printf("\nReader created successfully\n")
}