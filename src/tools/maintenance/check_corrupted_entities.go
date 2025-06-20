//go:build tool
package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: check_corrupted_entities <database_file>")
		fmt.Println("Example: check_corrupted_entities /opt/entitydb/var/entities.edb")
		os.Exit(1)
	}

	// Accept database file path directly
	ebfFile := os.Args[1]
	
	// Open the file
	file, err := os.Open(ebfFile)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()
	
	// Read header to get index location
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
	
	// Read index entries and look for problematic entities
	fmt.Println("\nChecking for corrupted entity IDs in index:")
	
	for i := uint64(0); i < entityCount && i < 20; i++ {
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
		
		// Check for corrupted IDs (non-ASCII characters)
		isCorrupted := false
		for _, b := range entityID {
			if b != 0 && (b < 32 || b > 126) {
				isCorrupted = true
				break
			}
		}
		
		if isCorrupted {
			fmt.Printf("Entry %d: CORRUPTED ID (hex: %x), Offset=%d, Size=%d\n", 
				i, entityID, offset, size)
			
			// Check if this might be the fixed offset
			if offset < 100000 {
				fmt.Printf("  -> This looks like a fixed offset (low value)\n")
			}
			
			// Also show as string for debugging
			fmt.Printf("  -> As string: %q\n", string(entityID[:]))
		} else {
			id := string(entityID[:])
			if len(id) > 0 && id[0] != 0 {
				fmt.Printf("Entry %d: ID=%s, Offset=%d, Size=%d\n", i, id, offset, size)
			}
		}
	}
}