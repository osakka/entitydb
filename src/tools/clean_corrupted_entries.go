//go:build tool
package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	
	"entitydb/config"
)

type IndexEntry struct {
	ID     [36]byte
	Offset uint64
	Size   uint32
	Flags  uint32
}

func isValidUUID(id string) bool {
	if len(id) != 36 {
		return false
	}
	
	parts := strings.Split(id, "-")
	if len(parts) != 5 {
		return false
	}
	
	if len(parts[0]) != 8 || len(parts[1]) != 4 || len(parts[2]) != 4 || 
	   len(parts[3]) != 4 || len(parts[4]) != 12 {
		return false
	}
	
	for _, c := range id {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || 
		     (c >= 'A' && c <= 'F') || c == '-') {
			return false
		}
	}
	
	return true
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: clean_corrupted_entries <data_directory>")
		os.Exit(1)
	}

	dataDir := os.Args[1]
	ebfFile := dataDir
	
	info, err := os.Stat(dataDir)
	if err == nil && info.IsDir() {
		// Load configuration to get proper database file path
		cfg := config.Load()
		cfg.DataPath = dataDir
		ebfFile = cfg.DatabaseFilename
	}
	
	// Create backup
	backupFile := ebfFile + ".backup"
	fmt.Printf("Creating backup at %s...\n", backupFile)
	
	// Copy file
	input, err := os.ReadFile(ebfFile)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}
	
	err = os.WriteFile(backupFile, input, 0644)
	if err != nil {
		fmt.Printf("Error creating backup: %v\n", err)
		os.Exit(1)
	}
	
	// Now work on the original file
	file, err := os.OpenFile(ebfFile, os.O_RDWR, 0644)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()
	
	// Read header
	var entityIndexOffset uint64
	var entityCount uint64
	var fileSize uint64
	
	file.Seek(8, 0)
	binary.Read(file, binary.LittleEndian, &fileSize)
	file.Seek(32, 0)
	binary.Read(file, binary.LittleEndian, &entityIndexOffset)
	file.Seek(48, 0)
	binary.Read(file, binary.LittleEndian, &entityCount)
	
	fmt.Printf("File size: %d\n", fileSize)
	fmt.Printf("Index at offset: %d\n", entityIndexOffset)
	fmt.Printf("Entity count: %d\n", entityCount)
	
	// Read all index entries
	file.Seek(int64(entityIndexOffset), 0)
	
	validEntries := make([]IndexEntry, 0)
	removedCount := 0
	
	fmt.Println("\nScanning for corrupted entries to remove...")
	
	for i := uint64(0); i < entityCount; i++ {
		var entry IndexEntry
		
		// Read index entry
		err := binary.Read(file, binary.LittleEndian, &entry.ID)
		if err != nil {
			fmt.Printf("Error reading entry %d: %v\n", i, err)
			break
		}
		
		binary.Read(file, binary.LittleEndian, &entry.Offset)
		binary.Read(file, binary.LittleEndian, &entry.Size)
		binary.Read(file, binary.LittleEndian, &entry.Flags)
		
		id := string(entry.ID[:])
		
		// Check if ID is valid
		isCorrupted := false
		
		// Check for specific corruption patterns
		if strings.Contains(id, "{l*h") || strings.Contains(id, "pl*h") || !isValidUUID(id) {
			isCorrupted = true
		}
		
		// Check for unreasonable size (e.g., > 100MB)
		if entry.Size > 100000000 {
			isCorrupted = true
		}
		
		// Check if offset exceeds file size
		if entry.Offset > fileSize {
			isCorrupted = true
		}
		
		if isCorrupted {
			fmt.Printf("Removing corrupted entry %d: ID=%q, Offset=%d, Size=%d\n", 
				i, id, entry.Offset, entry.Size)
			removedCount++
		} else {
			validEntries = append(validEntries, entry)
		}
	}
	
	fmt.Printf("\nFound %d valid entries, removing %d corrupted entries\n", 
		len(validEntries), removedCount)
	
	if removedCount == 0 {
		fmt.Println("No corrupted entries found, skipping rewrite")
		return
	}
	
	// Write the cleaned index back
	fmt.Println("\nRewriting index with valid entries only...")
	
	// Update entity count in header
	newEntityCount := uint64(len(validEntries))
	file.Seek(48, 0)
	binary.Write(file, binary.LittleEndian, &newEntityCount)
	
	// Rewrite the index
	file.Seek(int64(entityIndexOffset), 0)
	
	for _, entry := range validEntries {
		binary.Write(file, binary.LittleEndian, &entry.ID)
		binary.Write(file, binary.LittleEndian, &entry.Offset)
		binary.Write(file, binary.LittleEndian, &entry.Size)
		binary.Write(file, binary.LittleEndian, &entry.Flags)
	}
	
	// Truncate file if index is now smaller
	newIndexEnd := entityIndexOffset + uint64(len(validEntries)*52)
	if newIndexEnd < fileSize {
		file.Truncate(int64(newIndexEnd))
		
		// Update file size in header
		file.Seek(8, 0)
		binary.Write(file, binary.LittleEndian, &newIndexEnd)
	}
	
	fmt.Printf("\nCleaning complete. Removed %d corrupted entries.\n", removedCount)
	fmt.Println("Original file backed up to:", backupFile)
}