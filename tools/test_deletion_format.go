// Test tool to verify enhanced unified file format with deletion sections
package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"time"

	"entitydb/logger"
	"entitydb/models"
	storagebinary "entitydb/storage/binary"
)

func main() {
	// Initialize logger
	logger.SetLogLevel("INFO")
	
	fmt.Println("🗃️  EntityDB Enhanced File Format Test")
	fmt.Println("=====================================")
	
	// Test 1: Create header with deletion sections
	fmt.Println("\n1. Testing enhanced header format...")
	
	header := &storagebinary.Header{
		Magic:               storagebinary.MagicNumber,
		Version:             storagebinary.FormatVersion,
		FileSize:            1024,
		WALOffset:           128,
		WALSize:            256,
		DataOffset:          384,
		DataSize:           256,
		TagDictOffset:      640,
		TagDictSize:        128,
		EntityIndexOffset:  768,
		EntityIndexSize:    128,
		EntityCount:        5,
		LastModified:       time.Now().Unix(),
		WALSequence:        1,
		CheckpointSequence: 1,
		DeletionIndexOffset: 896,
		DeletionIndexSize:   512,
	}
	
	// Test header write/read
	tmpFile := "/tmp/test_header.dat"
	defer os.Remove(tmpFile)
	
	// Write header
	file, err := os.Create(tmpFile)
	if err != nil {
		log.Fatalf("Failed to create test file: %v", err)
	}
	
	if err := header.Write(file); err != nil {
		log.Fatalf("Failed to write header: %v", err)
	}
	file.Close()
	
	// Read header back
	file, err = os.Open(tmpFile)
	if err != nil {
		log.Fatalf("Failed to open test file: %v", err)
	}
	defer file.Close()
	
	readHeader := &storagebinary.Header{}
	if err := readHeader.Read(file); err != nil {
		log.Fatalf("Failed to read header: %v", err)
	}
	
	// Verify header fields
	if readHeader.Magic != storagebinary.MagicNumber {
		log.Fatalf("Magic mismatch: got 0x%x, want 0x%x", readHeader.Magic, storagebinary.MagicNumber)
	}
	if readHeader.Version != storagebinary.FormatVersion {
		log.Fatalf("Version mismatch: got %d, want %d", readHeader.Version, storagebinary.FormatVersion)
	}
	if readHeader.DeletionIndexOffset != 896 {
		log.Fatalf("DeletionIndexOffset mismatch: got %d, want 896", readHeader.DeletionIndexOffset)
	}
	if readHeader.DeletionIndexSize != 512 {
		log.Fatalf("DeletionIndexSize mismatch: got %d, want 512", readHeader.DeletionIndexSize)
	}
	
	fmt.Printf("   ✅ Header format version %d with deletion sections\n", readHeader.Version)
	fmt.Printf("   ✅ Deletion index at offset %d, size %d bytes\n", 
		readHeader.DeletionIndexOffset, readHeader.DeletionIndexSize)
	
	// Test 2: Create and test deletion entries
	fmt.Println("\n2. Testing deletion entry format...")
	
	// Create test deletion entry
	entry := storagebinary.NewDeletionEntry(
		"test-entity-001",
		models.StateSoftDeleted,
		"admin",
		"Testing deletion format",
		"test-policy",
		time.Now().UnixNano(),
	)
	
	// Test entry serialization
	entryFile := "/tmp/test_deletion_entry.dat"
	defer os.Remove(entryFile)
	
	// Write entry
	entryFd, err := os.Create(entryFile)
	if err != nil {
		log.Fatalf("Failed to create entry file: %v", err)
	}
	
	if err := entry.Write(entryFd); err != nil {
		log.Fatalf("Failed to write deletion entry: %v", err)
	}
	entryFd.Close()
	
	// Read entry back
	entryFd, err = os.Open(entryFile)
	if err != nil {
		log.Fatalf("Failed to open entry file: %v", err)
	}
	defer entryFd.Close()
	
	readEntry := &storagebinary.DeletionEntry{}
	if err := readEntry.Read(entryFd); err != nil {
		log.Fatalf("Failed to read deletion entry: %v", err)
	}
	
	// Verify entry data
	if readEntry.GetEntityID() != "test-entity-001" {
		log.Fatalf("EntityID mismatch: got %s, want test-entity-001", readEntry.GetEntityID())
	}
	if readEntry.GetLifecycleState() != models.StateSoftDeleted {
		log.Fatalf("LifecycleState mismatch: got %s, want %s", readEntry.GetLifecycleState(), models.StateSoftDeleted)
	}
	if readEntry.GetDeletedBy() != "admin" {
		log.Fatalf("DeletedBy mismatch: got %s, want admin", readEntry.GetDeletedBy())
	}
	if readEntry.GetReason() != "Testing deletion format" {
		log.Fatalf("Reason mismatch: got %s, want 'Testing deletion format'", readEntry.GetReason())
	}
	if readEntry.GetPolicy() != "test-policy" {
		log.Fatalf("Policy mismatch: got %s, want test-policy", readEntry.GetPolicy())
	}
	
	fmt.Printf("   ✅ Deletion entry: %s (%s)\n", readEntry.GetEntityID(), readEntry.GetLifecycleState())
	fmt.Printf("   ✅ Deleted by: %s, reason: %s\n", readEntry.GetDeletedBy(), readEntry.GetReason())
	fmt.Printf("   ✅ Policy: %s, size: %d bytes\n", readEntry.GetPolicy(), storagebinary.DeletionEntrySize)
	
	// Test 3: Test deletion index
	fmt.Println("\n3. Testing deletion index operations...")
	
	index := storagebinary.NewDeletionIndex()
	
	// Add multiple entries
	entries := []*storagebinary.DeletionEntry{
		storagebinary.NewDeletionEntry("entity-001", models.StateSoftDeleted, "user1", "Reason 1", "policy1", time.Now().UnixNano()),
		storagebinary.NewDeletionEntry("entity-002", models.StateArchived, "user2", "Reason 2", "policy2", time.Now().UnixNano()),
		storagebinary.NewDeletionEntry("entity-003", models.StatePurged, "user3", "Reason 3", "policy3", time.Now().UnixNano()),
	}
	
	for _, entry := range entries {
		index.AddEntry(entry)
	}
	
	fmt.Printf("   ✅ Added %d entries to deletion index\n", index.Count())
	
	// Test retrieval
	if entry, exists := index.GetEntry("entity-002"); exists {
		fmt.Printf("   ✅ Retrieved entry: %s (%s)\n", entry.GetEntityID(), entry.GetLifecycleState())
	} else {
		log.Fatalf("Failed to retrieve entity-002")
	}
	
	// Test filtering by state
	softDeleted := index.GetEntriesByState(models.StateSoftDeleted)
	archived := index.GetEntriesByState(models.StateArchived)
	purged := index.GetEntriesByState(models.StatePurged)
	
	fmt.Printf("   ✅ Entries by state: %d soft deleted, %d archived, %d purged\n",
		len(softDeleted), len(archived), len(purged))
	
	// Test 4: Test index serialization
	fmt.Println("\n4. Testing deletion index serialization...")
	
	indexFile := "/tmp/test_deletion_index.dat"
	defer os.Remove(indexFile)
	
	// Write index
	indexFd, err := os.Create(indexFile)
	if err != nil {
		log.Fatalf("Failed to create index file: %v", err)
	}
	
	if err := index.WriteTo(indexFd); err != nil {
		log.Fatalf("Failed to write deletion index: %v", err)
	}
	indexFd.Close()
	
	// Read index back
	indexFd, err = os.Open(indexFile)
	if err != nil {
		log.Fatalf("Failed to open index file: %v", err)
	}
	defer indexFd.Close()
	
	newIndex := storagebinary.NewDeletionIndex()
	if err := newIndex.ReadFrom(indexFd, 3); err != nil {
		log.Fatalf("Failed to read deletion index: %v", err)
	}
	
	if newIndex.Count() != 3 {
		log.Fatalf("Index count mismatch: got %d, want 3", newIndex.Count())
	}
	
	fmt.Printf("   ✅ Serialized and deserialized %d deletion entries\n", newIndex.Count())
	
	// Verify all entries were preserved
	for _, originalEntry := range entries {
		if readEntry, exists := newIndex.GetEntry(originalEntry.GetEntityID()); exists {
			fmt.Printf("   ✅ Verified entry: %s (%s)\n", 
				readEntry.GetEntityID(), readEntry.GetLifecycleState())
		} else {
			log.Fatalf("Failed to find entry: %s", originalEntry.GetEntityID())
		}
	}
	
	// Test 5: Test backward compatibility
	fmt.Println("\n5. Testing backward compatibility...")
	
	// Create a version 2 header (without deletion sections)
	v2Header := &storagebinary.Header{
		Magic:               storagebinary.MagicNumber,
		Version:             2, // Version 2
		FileSize:            1024,
		WALOffset:           128,
		WALSize:            256,
		DataOffset:          384,
		DataSize:           256,
		TagDictOffset:      640,
		TagDictSize:        128,
		EntityIndexOffset:  768,
		EntityIndexSize:    128,
		EntityCount:        5,
		LastModified:       time.Now().Unix(),
		WALSequence:        1,
		CheckpointSequence: 1,
		// No deletion fields for version 2
	}
	
	v2File := "/tmp/test_v2_header.dat"
	defer os.Remove(v2File)
	
	// Write v2 header (without deletion fields)
	v2Fd, err := os.Create(v2File)
	if err != nil {
		log.Fatalf("Failed to create v2 file: %v", err)
	}
	
	// Manually write v2 header to simulate old format
	buf := make([]byte, 128)
	binary.LittleEndian.PutUint32(buf[0:4], v2Header.Magic)
	binary.LittleEndian.PutUint32(buf[4:8], v2Header.Version)
	binary.LittleEndian.PutUint64(buf[8:16], v2Header.FileSize)
	binary.LittleEndian.PutUint64(buf[16:24], v2Header.WALOffset)
	binary.LittleEndian.PutUint64(buf[24:32], v2Header.WALSize)
	binary.LittleEndian.PutUint64(buf[32:40], v2Header.DataOffset)
	binary.LittleEndian.PutUint64(buf[40:48], v2Header.DataSize)
	binary.LittleEndian.PutUint64(buf[48:56], v2Header.TagDictOffset)
	binary.LittleEndian.PutUint64(buf[56:64], v2Header.TagDictSize)
	binary.LittleEndian.PutUint64(buf[64:72], v2Header.EntityIndexOffset)
	binary.LittleEndian.PutUint64(buf[72:80], v2Header.EntityIndexSize)
	binary.LittleEndian.PutUint64(buf[80:88], v2Header.EntityCount)
	binary.LittleEndian.PutUint64(buf[88:96], uint64(v2Header.LastModified))
	binary.LittleEndian.PutUint64(buf[96:104], v2Header.WALSequence)
	binary.LittleEndian.PutUint64(buf[104:112], v2Header.CheckpointSequence)
	// buf[112:128] remains zero (no deletion fields)
	
	v2Fd.Write(buf)
	v2Fd.Close()
	
	// Read v2 header back
	v2Fd, err = os.Open(v2File)
	if err != nil {
		log.Fatalf("Failed to open v2 file: %v", err)
	}
	defer v2Fd.Close()
	
	readV2Header := &storagebinary.Header{}
	if err := readV2Header.Read(v2Fd); err != nil {
		log.Fatalf("Failed to read v2 header: %v", err)
	}
	
	// Verify v2 compatibility
	if readV2Header.Version != 2 {
		log.Fatalf("V2 version mismatch: got %d, want 2", readV2Header.Version)
	}
	if readV2Header.DeletionIndexOffset != 0 {
		log.Fatalf("V2 should have no deletion index: got offset %d", readV2Header.DeletionIndexOffset)
	}
	if readV2Header.DeletionIndexSize != 0 {
		log.Fatalf("V2 should have no deletion index: got size %d", readV2Header.DeletionIndexSize)
	}
	
	fmt.Printf("   ✅ Version 2 compatibility: deletion fields initialized to 0\n")
	fmt.Printf("   ✅ Supports both version 2 and 3 file formats\n")
	
	fmt.Println("\n🎉 Enhanced file format test completed successfully!")
	fmt.Println("\nFormat enhancements:")
	fmt.Println("  📁 Version 3 unified file format")
	fmt.Println("  🗑️  Dedicated deletion index section")
	fmt.Println("  📋 256-byte deletion entries with full audit trail")
	fmt.Println("  🔄 Backward compatibility with version 2")
	fmt.Println("  ⚡ High-performance binary serialization")
}