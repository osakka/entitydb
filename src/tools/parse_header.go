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
	
	// Read header (64 bytes)
	buf := make([]byte, 64)
	if _, err := file.Read(buf); err != nil {
		log.Fatal(err)
	}
	
	// Parse header
	magic := binary.LittleEndian.Uint32(buf[0:4])
	version := binary.LittleEndian.Uint32(buf[4:8])
	fileSize := binary.LittleEndian.Uint64(buf[8:16])
	tagDictOffset := binary.LittleEndian.Uint64(buf[16:24])
	tagDictSize := binary.LittleEndian.Uint64(buf[24:32])
	entityIndexOffset := binary.LittleEndian.Uint64(buf[32:40])
	entityIndexSize := binary.LittleEndian.Uint64(buf[40:48])
	entityCount := binary.LittleEndian.Uint64(buf[48:56])
	lastModified := int64(binary.LittleEndian.Uint64(buf[56:64]))
	
	fmt.Printf("Header:\n")
	fmt.Printf("  Magic: 0x%08X (%s)\n", magic, string(buf[0:4]))
	fmt.Printf("  Version: %d\n", version)
	fmt.Printf("  FileSize: %d\n", fileSize)
	fmt.Printf("  TagDictOffset: %d (0x%X)\n", tagDictOffset, tagDictOffset)
	fmt.Printf("  TagDictSize: %d\n", tagDictSize)
	fmt.Printf("  EntityIndexOffset: %d (0x%X)\n", entityIndexOffset, entityIndexOffset)
	fmt.Printf("  EntityIndexSize: %d\n", entityIndexSize)
	fmt.Printf("  EntityCount: %d\n", entityCount)
	fmt.Printf("  LastModified: %d\n", lastModified)
	
	// Get actual file size
	stat, _ := file.Stat()
	fmt.Printf("\nActual file size: %d bytes\n", stat.Size())
}