package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	
	"entitydb/storage/binary"
)

func main() {
	// Get data path
	dataPath := os.Getenv("ENTITYDB_DATA_PATH")
	if dataPath == "" {
		dataPath = "var"
	}
	
	walPath := filepath.Join(dataPath, "entitydb.wal")
	
	// Check WAL file
	stat, err := os.Stat(walPath)
	if err != nil {
		log.Fatalf("Failed to stat WAL file: %v", err)
	}
	
	fmt.Printf("WAL file: %s\n", walPath)
	fmt.Printf("WAL size: %d bytes\n", stat.Size())
	
	// Create WAL reader
	wal, err := binary.NewWAL(dataPath)
	if err != nil {
		log.Fatalf("Failed to create WAL: %v", err)
	}
	
	// Count entries by type
	createCount := 0
	updateCount := 0
	deleteCount := 0
	entityIDs := make(map[string]bool)
	
	fmt.Printf("\nReading WAL entries...\n")
	err = wal.Replay(func(entry binary.WALEntry) error {
		entityIDs[entry.EntityID] = true
		
		switch entry.OpType {
		case binary.WALOpCreate:
			createCount++
			if createCount <= 5 {
				fmt.Printf("CREATE: %s\n", entry.EntityID)
			}
		case binary.WALOpUpdate:
			updateCount++
			if updateCount <= 5 {
				fmt.Printf("UPDATE: %s\n", entry.EntityID)
			}
		case binary.WALOpDelete:
			deleteCount++
			fmt.Printf("DELETE: %s\n", entry.EntityID)
		}
		
		return nil
	})
	
	if err != nil {
		log.Printf("Failed to replay WAL: %v", err)
	}
	
	fmt.Printf("\nWAL Summary:\n")
	fmt.Printf("  CREATE operations: %d\n", createCount)
	fmt.Printf("  UPDATE operations: %d\n", updateCount)
	fmt.Printf("  DELETE operations: %d\n", deleteCount)
	fmt.Printf("  Unique entities: %d\n", len(entityIDs))
}