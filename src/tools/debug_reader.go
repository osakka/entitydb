package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
)

func main() {
	file, err := os.Open("/opt/entitydb/var/entities.ebf")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	
	stat, _ := file.Stat()
	fmt.Printf("File size: %d bytes\n", stat.Size())
	
	// Read header manually
	buf := make([]byte, 64)
	n, err := file.Read(buf)
	if err != nil {
		log.Fatal("Failed to read header:", err)
	}
	fmt.Printf("Read %d bytes for header\n", n)
	
	// Parse header fields
	magic := binary.LittleEndian.Uint32(buf[0:4])
	version := binary.LittleEndian.Uint32(buf[4:8])
	entityCount := binary.LittleEndian.Uint64(buf[48:56])
	entityIndexOffset := binary.LittleEndian.Uint64(buf[32:40])
	
	fmt.Printf("\nHeader:\n")
	fmt.Printf("  Magic: 0x%08X (%s)\n", magic, string(buf[0:4]))
	fmt.Printf("  Version: %d\n", version)
	fmt.Printf("  EntityCount: %d\n", entityCount)
	fmt.Printf("  EntityIndexOffset: %d (0x%X)\n", entityIndexOffset, entityIndexOffset)
	
	// Try to read index
	if entityIndexOffset > 0 && entityCount > 0 {
		fmt.Printf("\nSeeking to index at offset %d...\n", entityIndexOffset)
		pos, err := file.Seek(int64(entityIndexOffset), 0)
		if err != nil {
			log.Fatal("Failed to seek to index:", err)
		}
		fmt.Printf("Seeked to position: %d\n", pos)
		
		// Read first index entry (52 bytes: 36 for ID + 8 offset + 4 size + 4 flags)
		indexBuf := make([]byte, 52)
		n, err = file.Read(indexBuf)
		if err != nil {
			log.Printf("Failed to read index entry: %v", err)
		} else {
			fmt.Printf("Read %d bytes for first index entry\n", n)
			
			// Parse the entity ID (36 bytes)
			entityID := string(indexBuf[0:36])
			offset := binary.LittleEndian.Uint64(indexBuf[36:44])
			size := binary.LittleEndian.Uint32(indexBuf[44:48])
			flags := binary.LittleEndian.Uint32(indexBuf[48:52])
			
			fmt.Printf("\nFirst index entry:\n")
			fmt.Printf("  EntityID: %s\n", entityID)
			fmt.Printf("  Offset: %d (0x%X)\n", offset, offset)
			fmt.Printf("  Size: %d\n", size)
			fmt.Printf("  Flags: %d\n", flags)
			
			// Try to read the entity data
			if offset > 0 && size > 0 {
				fmt.Printf("\nTrying to read entity data at offset %d...\n", offset)
				pos, err = file.Seek(int64(offset), 0)
				if err != nil {
					log.Printf("Failed to seek to entity data: %v", err)
				} else {
					fmt.Printf("Seeked to position: %d\n", pos)
					
					// Read first 100 bytes of entity data
					entityBuf := make([]byte, 100)
					n, err = file.Read(entityBuf)
					if err != nil {
						log.Printf("Failed to read entity data: %v", err)
					} else {
						fmt.Printf("Read %d bytes of entity data\n", n)
						fmt.Printf("First 100 bytes (hex):\n")
						for i := 0; i < n && i < 100; i += 16 {
							fmt.Printf("%04X: ", i)
							for j := 0; j < 16 && i+j < n; j++ {
								fmt.Printf("%02X ", entityBuf[i+j])
							}
							fmt.Printf("\n")
						}
					}
				}
			}
		}
	}
}