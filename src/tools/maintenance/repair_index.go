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
		fmt.Println("Usage: fix_index <database_file>")
		fmt.Println("Example: fix_index /opt/entitydb/var/entities.edb")
		os.Exit(1)
	}

	ebfFile := os.Args[1]
	
	// Open the file
	file, err := os.Open(ebfFile)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()
	
	// Get file size
	stat, err := file.Stat()
	if err != nil {
		fmt.Printf("Error getting file stats: %v\n", err)
		os.Exit(1)
	}
	
	fileSize := stat.Size()
	fmt.Printf("File size: %d bytes\n", fileSize)
	
	// Check the problematic offsets
	invalidOffsets := []int64{
		55370718392923,
		55409373098596,
		55448027804269,
		55332063687250,
		55263344210498,
	}
	
	fmt.Println("\nChecking invalid offsets:")
	for _, offset := range invalidOffsets {
		fmt.Printf("Offset %d: ", offset)
		if offset > fileSize {
			fmt.Printf("INVALID (exceeds file size by %d bytes)\n", offset-fileSize)
		} else {
			fmt.Printf("Valid\n")
		}
	}
	
	// Read header to understand the structure
	var magic uint32
	var version uint32
	var fileSizeHeader uint64
	var entityCount uint64
	var entityIndexOffset uint64
	var entityIndexSize uint64
	
	// Read header fields
	binary.Read(file, binary.LittleEndian, &magic)
	binary.Read(file, binary.LittleEndian, &version)
	binary.Read(file, binary.LittleEndian, &fileSizeHeader)
	file.Seek(32, 0) // Skip to entity index offset
	binary.Read(file, binary.LittleEndian, &entityIndexOffset)
	binary.Read(file, binary.LittleEndian, &entityIndexSize)
	file.Seek(48, 0) // Skip to entity count
	binary.Read(file, binary.LittleEndian, &entityCount)
	
	fmt.Printf("\nHeader info:\n")
	fmt.Printf("Magic: 0x%X\n", magic)
	fmt.Printf("Version: %d\n", version)
	fmt.Printf("File size (header): %d\n", fileSizeHeader)
	fmt.Printf("Entity count: %d\n", entityCount)
	fmt.Printf("Index offset: %d\n", entityIndexOffset)
	fmt.Printf("Index size: %d\n", entityIndexSize)
	
	// Check for corruption patterns
	fmt.Printf("\nCorruption analysis:\n")
	
	// These offsets look like they might be corrupted pointers
	// Let's see if they're multiples or have a pattern
	for _, offset := range invalidOffsets {
		// Check if high bits are set incorrectly
		highBits := offset >> 32
		lowBits := offset & 0xFFFFFFFF
		fmt.Printf("Offset %d: high=%d, low=%d\n", offset, highBits, lowBits)
		
		// This might be a 32-bit value incorrectly interpreted as 64-bit
		possibleReal := int64(lowBits)
		if possibleReal < fileSize {
			fmt.Printf("  Possible real offset: %d\n", possibleReal)
		}
	}
	
	// Scan the index for corruption
	fmt.Printf("\nScanning index at offset %d...\n", entityIndexOffset)
	
	if entityIndexOffset > 0 && int64(entityIndexOffset) < fileSize {
		file.Seek(int64(entityIndexOffset), 0)
		
		corruptedEntries := 0
		for i := uint64(0); i < entityCount && i < 10; i++ { // Check first 10 entries
			var entityID [36]byte
			var offset uint64
			var size uint32
			var flags uint32
			
			// Read index entry
			err := binary.Read(file, binary.LittleEndian, &entityID)
			if err != nil {
				break
			}
			err = binary.Read(file, binary.LittleEndian, &offset)
			if err != nil {
				break
			}
			err = binary.Read(file, binary.LittleEndian, &size)
			if err != nil {
				break
			}
			err = binary.Read(file, binary.LittleEndian, &flags)
			if err != nil {
				break
			}
			
			// Check if offset is valid
			if offset > uint64(fileSize) {
				fmt.Printf("Entry %d: ID=%s, Offset=%d (INVALID)\n", i, string(entityID[:]), offset)
				corruptedEntries++
			}
		}
		
		fmt.Printf("Found %d corrupted entries in first 10\n", corruptedEntries)
	}
}