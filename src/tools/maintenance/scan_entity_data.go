package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func isValidUUID(id string) bool {
	// Basic UUID format check
	if len(id) != 36 {
		return false
	}
	
	// Check format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	parts := strings.Split(id, "-")
	if len(parts) != 5 {
		return false
	}
	
	if len(parts[0]) != 8 || len(parts[1]) != 4 || len(parts[2]) != 4 || 
	   len(parts[3]) != 4 || len(parts[4]) != 12 {
		return false
	}
	
	// Check if all characters are hex digits or hyphens
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
		fmt.Println("Usage: scan_entity_data <data_directory>")
		os.Exit(1)
	}

	dataDir := os.Args[1]
	ebfFile := dataDir
	
	// Check if directory or file was provided
	info, err := os.Stat(dataDir)
	if err == nil && info.IsDir() {
		ebfFile = filepath.Join(dataDir, "entities.ebf")
	}
	
	// Open the file
	file, err := os.Open(ebfFile)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()
	
	// Read header
	var entityIndexOffset uint64
	var entityCount uint64
	
	// Skip to entity index offset (position 32)
	file.Seek(32, 0)
	binary.Read(file, binary.LittleEndian, &entityIndexOffset)
	file.Seek(48, 0)
	binary.Read(file, binary.LittleEndian, &entityCount)
	
	fmt.Printf("Index at offset: %d\n", entityIndexOffset)
	fmt.Printf("Entity count: %d\n", entityCount)
	
	// Seek to index
	file.Seek(int64(entityIndexOffset), 0)
	
	// Read all index entries first
	type IndexEntry struct {
		ID     string
		Offset uint64
		Size   uint32
		Flags  uint32
	}
	
	entries := make([]IndexEntry, 0, entityCount)
	corrupted := 0
	
	fmt.Println("\nScanning index for corrupted entries...")
	
	for i := uint64(0); i < entityCount; i++ {
		var entityID [36]byte
		var offset uint64
		var size uint32
		var flags uint32
		
		// Read index entry
		err := binary.Read(file, binary.LittleEndian, &entityID)
		if err != nil {
			fmt.Printf("Error reading entry %d: %v\n", i, err)
			break
		}
		
		binary.Read(file, binary.LittleEndian, &offset)
		binary.Read(file, binary.LittleEndian, &size)
		binary.Read(file, binary.LittleEndian, &flags)
		
		id := string(entityID[:])
		
		// Check if ID looks corrupted
		if strings.Contains(id, "{l*h") || !isValidUUID(id) {
			fmt.Printf("\nCorrupted entry %d:\n", i)
			fmt.Printf("  ID (string): %q\n", id)
			fmt.Printf("  ID (hex): %x\n", entityID)
			fmt.Printf("  Offset: %d\n", offset)
			fmt.Printf("  Size: %d\n", size)
			corrupted++
			
			// Try to read the actual entity data
			currentPos, _ := file.Seek(0, 1) // Save current position
			file.Seek(int64(offset), 0)
			
			// Read a sample of the entity data
			sample := make([]byte, 64)
			n, _ := file.Read(sample)
			if n > 0 {
				fmt.Printf("  Data sample (hex): %x\n", sample[:n])
				fmt.Printf("  Data sample (string): %q\n", sample[:n])
			}
			
			file.Seek(currentPos, 0) // Restore position
		}
		
		entries = append(entries, IndexEntry{
			ID:     id,
			Offset: offset,
			Size:   size,
			Flags:  flags,
		})
	}
	
	fmt.Printf("\n\nFound %d corrupted entries out of %d total\n", corrupted, entityCount)
	
	// Now check for data corruption by looking for the error message pattern
	fmt.Println("\nScanning for specific error pattern in entity data...")
	
	for i, entry := range entries {
		if i > 100 { // Limit scan for efficiency
			break
		}
		
		file.Seek(int64(entry.Offset), 0)
		
		// Read entity data
		data := make([]byte, entry.Size)
		n, err := file.Read(data)
		if err != nil || n != int(entry.Size) {
			continue
		}
		
		// Check for specific corruption patterns
		dataStr := string(data)
		if strings.Contains(dataStr, "{l*h") || strings.Contains(dataStr, "M2  N2  O2  P2  Q2") {
			fmt.Printf("\nFound corruption pattern in entity %s:\n", entry.ID)
			fmt.Printf("  Offset: %d\n", entry.Offset)
			fmt.Printf("  Size: %d\n", entry.Size)
			fmt.Printf("  Data sample: %q\n", dataStr[:min(100, len(dataStr))])
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}