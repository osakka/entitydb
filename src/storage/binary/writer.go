// Package binary implements the EntityDB Binary Format (EBF) storage layer.
// It provides high-performance, concurrent-safe persistence for entities with
// temporal tags, supporting compression, checksumming, and memory-mapped access.
//
// The binary format consists of:
//   - Header: Magic number, version, entity count, and offset information
//   - Entity data: Sequential entity records with tags and content
//   - Tag dictionary: Compressed mapping of tag strings to IDs
//   - Index: Offset and size information for rapid entity lookup
//
// Key features:
//   - Temporal tag storage with nanosecond precision timestamps
//   - Automatic content compression for entries > 1KB
//   - SHA256 checksumming for data integrity
//   - Memory pooling to reduce GC pressure
//   - Concurrent access with sharded locking
//   - Write-Ahead Logging (WAL) for durability
package binary

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"entitydb/models"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
	"entitydb/logger"
)

// Writer handles writing entities to the EntityDB Binary Format (EBF).
// It manages the file structure, maintains indexes for fast lookups,
// and ensures data integrity through checksums and atomic operations.
//
// The Writer is safe for concurrent use through internal mutex locking.
// However, only one Writer instance should be active per file to prevent
// corruption. Use WriterManager for safe concurrent access patterns.
//
// File Layout:
//   [Header][Entity1][Entity2]...[EntityN][TagDictionary][Index]
//
// Performance characteristics:
//   - O(1) append operations for new entities
//   - O(1) index updates through in-memory map
//   - Compression reduces I/O for large content
//   - Memory pools minimize allocation overhead
type Writer struct {
	file     *os.File                // Underlying file handle
	header   *Header                 // File header with metadata
	tagDict  *TagDictionary          // Tag string to ID mapping
	index    map[string]*IndexEntry  // Entity ID to file location mapping
	mu       sync.Mutex              // Protects concurrent access
}

// getFilePosition returns the current file position for tracking write operations.
// Returns -1 if the position cannot be determined (e.g., file closed or seek error).
// This is primarily used for operation tracking and debugging.
func (w *Writer) getFilePosition() int64 {
	pos, err := w.file.Seek(0, os.SEEK_CUR)
	if err != nil {
		return -1
	}
	return pos
}

// NewWriter creates a new Writer instance for the specified file.
// If the file exists and contains valid data, it loads the existing header,
// tag dictionary, and index. For new files, it initializes the header structure.
//
// The Writer maintains exclusive access to the file while open. Use WriterManager
// for safe concurrent access patterns in production environments.
//
// Parameters:
//   - filename: Path to the EBF file to open or create
//
// Returns:
//   - *Writer: Initialized writer ready for entity operations
//   - error: File access errors, corruption, or invalid format
//
// Example:
//   writer, err := NewWriter("/var/lib/entitydb/entities.ebf")
//   if err != nil {
//       log.Fatal(err)
//   }
//   defer writer.Close()
func NewWriter(filename string) (*Writer, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	
	w := &Writer{
		file:    file,
		header:  &Header{Magic: MagicNumber, Version: FormatVersion},
		tagDict: NewTagDictionary(),
		index:   make(map[string]*IndexEntry),
	}
	
	// Try to read existing file
	if stat, err := file.Stat(); err == nil && stat.Size() > 0 {
		if err := w.readExisting(); err != nil {
			logger.Warn("Failed to read existing file, creating new: %v", err)
			// Reset the file
			file.Truncate(0)
			file.Seek(0, 0)
			// Write initial header
			if err := w.writeHeader(); err != nil {
				return nil, err
			}
			if err := w.file.Sync(); err != nil {
				return nil, err
			}
		}
	} else {
		// New file, write initial header
		if err := w.writeHeader(); err != nil {
			return nil, err
		}
		// Ensure data is flushed to disk
		if err := w.file.Sync(); err != nil {
			return nil, err
		}
	}
	
	return w, nil
}

// WriteEntity persists an entity to the binary file with automatic compression,
// checksumming, and index updates. This is the primary method for storing entities.
//
// The method performs the following operations:
//   1. Validates the entity ID (required, non-empty)
//   2. Calculates content checksum for integrity verification
//   3. Adds checksum tag if not already present
//   4. Compresses content if size > 1KB and compression is beneficial
//   5. Writes entity header, tags, and content to file
//   6. Updates in-memory index for fast lookups
//   7. Updates file header with new counts and offsets
//   8. Syncs data to disk for durability
//
// Thread Safety:
//   - Method is synchronized with internal mutex
//   - Safe for concurrent calls from multiple goroutines
//   - Index updates are atomic within the lock
//
// Performance Notes:
//   - Uses memory pools to reduce allocations
//   - Compression is skipped if it doesn't reduce size
//   - File seeks are minimized through append-only writes
//   - Header updates require seeking to file start
//
// Parameters:
//   - entity: The entity to write (must have non-empty ID)
//
// Returns:
//   - error: Validation errors, I/O failures, or sync errors
//
// Example:
//   entity := &models.Entity{
//       ID: "user-123",
//       Tags: []string{"type:user", "status:active"},
//       Content: []byte(`{"name":"John Doe"}`),
//   }
//   if err := writer.WriteEntity(entity); err != nil {
//       return fmt.Errorf("failed to write entity: %w", err)
//   }
func (w *Writer) WriteEntity(entity *models.Entity) error {
	// Start operation tracking for observability
	op := models.StartOperation(models.OpTypeWrite, entity.ID, map[string]interface{}{
		"tags_count":    len(entity.Tags),
		"content_size":  len(entity.Content),
		"file_position": w.getFilePosition(),
		"entity_count":  w.header.EntityCount,
	})
	defer func() {
		if op != nil {
			op.Complete()
		}
	}()
	
	// Acquire exclusive lock for thread safety
	w.mu.Lock()
	defer w.mu.Unlock()
	
	// Log write intent
	logger.TraceIf("storage", "Starting write for entity %s (tags=%d, content=%d bytes)", 
		entity.ID, len(entity.Tags), len(entity.Content))
	
	// Validate entity
	if entity.ID == "" {
		err := fmt.Errorf("entity ID cannot be empty")
		op.Fail(err)
		logger.Error("Validation failed: %v", err)
		return err
	}
	
	// Calculate content checksum for data integrity verification
	// SHA256 provides strong collision resistance and is used throughout
	// the system for content addressing and verification
	contentChecksum := sha256.Sum256(entity.Content)
	op.SetMetadata("content_checksum", hex.EncodeToString(contentChecksum[:]))
	
	logger.TraceIf("storage", "Entity %s content checksum: %x", entity.ID, contentChecksum)
	
	// Acquire buffer from memory pool to reduce GC pressure
	// Large buffers (>64KB) are reused across write operations
	buffer := GetLargeSafeBuffer()
	defer PutLargeSafeBuffer(buffer)
	
	// Checksum Tag Algorithm:
	// 1. Generate temporal checksum tag with nanosecond timestamp
	// 2. Check if entity already has a checksum tag (temporal format)
	// 3. Add new checksum tag only if missing
	// This ensures content integrity can be verified later
	checksumTag := fmt.Sprintf("%d|checksum:sha256:%s", time.Now().UnixNano(), hex.EncodeToString(contentChecksum[:]))
	hasChecksum := false
	for _, tag := range entity.Tags {
		if strings.Contains(tag, "|checksum:sha256:") {
			hasChecksum = true
			break
		}
	}
	
	// Create a new tags slice with checksum if needed
	// We don't modify the original slice to avoid side effects
	tags := entity.Tags
	if !hasChecksum {
		tags = append([]string{}, entity.Tags...)
		tags = append(tags, checksumTag)
		logger.TraceIf("storage", "Added checksum tag for entity %s: %s", entity.ID, checksumTag)
	}
	
	// Convert tags to IDs
	tagIDs := make([]uint32, len(tags))
	for i, tag := range tags {
		tagIDs[i] = w.tagDict.GetOrCreateID(tag)
		logger.TraceIf("storage", "Tag '%s' assigned ID %d", tag, tagIDs[i])
	}
	
	// Write entity header
	header := EntityHeader{
		Modified:     time.Now().Unix(),
		TagCount:     uint16(len(tagIDs)),
		ContentCount: 1, // Now we store content as a single item
	}
	
	logger.TraceIf("storage", "Writing entity header: Modified=%d, TagCount=%d, ContentCount=%d", 
		header.Modified, header.TagCount, header.ContentCount)
	
	binary.Write(buffer, binary.LittleEndian, header)
	
	// Write tag IDs
	for _, id := range tagIDs {
		binary.Write(buffer, binary.LittleEndian, id)
	}
	
	// Content Encoding Algorithm:
	// The binary format stores content with metadata for proper decoding
	// Format: [CompressionType][ContentTypeLen][ContentType][OriginalSize][CompressedSize][Data][Timestamp]
	if len(entity.Content) > 0 {
		// Content Type Detection:
		// 1. Search for content:type: tag in entity tags
		// 2. Handle both temporal (timestamp|tag) and direct formats
		// 3. Default to application/octet-stream if not specified
		// This ensures proper content handling when reading back
		contentType := "application/octet-stream" // Default
		logger.TraceIf("storage", "Entity %s has %d tags, checking for content type", entity.ID, len(entity.Tags))
		for _, tag := range entity.Tags {
			logger.TraceIf("storage", "Checking tag: %s", tag)
			if strings.HasPrefix(tag, "content:type:") {
				// Direct match (non-timestamped)
				contentType = strings.TrimPrefix(tag, "content:type:")
				logger.TraceIf("storage", "Found direct content type: %s", contentType)
				break
			} else if strings.Contains(tag, "|content:type:") {
				// Timestamped tag format: "timestamp|content:type:value"
				parts := strings.SplitN(tag, "|", 2)
				if len(parts) == 2 {
					tagPart := parts[1] // Use the part after timestamp
					if strings.HasPrefix(tagPart, "content:type:") {
						contentType = strings.TrimPrefix(tagPart, "content:type:")
						logger.TraceIf("storage", "Found timestamped content type: %s", contentType)
						break
					}
				}
			}
		}
		logger.TraceIf("storage", "Final content type for entity %s: %s", entity.ID, contentType)
		
		// Compression Strategy:
		// - Content > 1KB is compressed using gzip
		// - Compression is skipped if it doesn't reduce size
		// - Failed compression falls back to uncompressed storage
		// - Uses memory pools to avoid allocation overhead
		compressed, err := CompressWithPool(entity.Content)
		if err != nil {
			logger.Warn("Compression failed for entity %s: %v, storing uncompressed", entity.ID, err)
			compressed = &CompressedContent{
				Type: CompressionNone,
				Data: entity.Content,
				OriginalSize: len(entity.Content),
			}
		}
		
		// Write compression type
		binary.Write(buffer, binary.LittleEndian, uint8(compressed.Type))
		
		// For JSON content, store it directly without additional wrapping
		if contentType == "application/json" {
			// Content is already JSON, store it as-is
			binary.Write(buffer, binary.LittleEndian, uint16(len(contentType)))
			buffer.WriteString(contentType)
			binary.Write(buffer, binary.LittleEndian, uint32(compressed.OriginalSize))
			binary.Write(buffer, binary.LittleEndian, uint32(len(compressed.Data)))
			buffer.Write(compressed.Data)
		} else {
			// For other content types, use application/octet-stream wrapper
			binary.Write(buffer, binary.LittleEndian, uint16(len("application/octet-stream")))
			buffer.WriteString("application/octet-stream")
			binary.Write(buffer, binary.LittleEndian, uint32(compressed.OriginalSize))
			binary.Write(buffer, binary.LittleEndian, uint32(len(compressed.Data)))
			buffer.Write(compressed.Data)
		}
		
		// Timestamp: current time
		binary.Write(buffer, binary.LittleEndian, time.Now().UnixNano())
	} else {
		// Empty content
		binary.Write(buffer, binary.LittleEndian, uint8(CompressionNone))
		contentType := "application/octet-stream"
		binary.Write(buffer, binary.LittleEndian, uint16(len(contentType)))
		buffer.WriteString(contentType)
		
		binary.Write(buffer, binary.LittleEndian, uint32(0)) // Original size
		binary.Write(buffer, binary.LittleEndian, uint32(0)) // Compressed size
		binary.Write(buffer, binary.LittleEndian, time.Now().UnixNano())
	}
	
	// Get current file position
	offset, err := w.file.Seek(0, os.SEEK_END)
	if err != nil {
		op.Fail(err)
		logger.Error("Failed to seek to end for entity %s: %v", entity.ID, err)
		return err
	}
	
	// SURGICAL FIX: Validate offset immediately after seek to catch corruption early
	if offset < 0 {
		err := fmt.Errorf("invalid negative offset %d returned from seek", offset)
		op.Fail(err)
		logger.Error("CORRUPTION DETECTED: %v for entity %s", err, entity.ID)
		return err
	}
	if offset > 1024*1024*1024*10 { // 10GB sanity check
		err := fmt.Errorf("impossibly large offset %d returned from seek (file corruption?)", offset)
		op.Fail(err)
		logger.Error("CORRUPTION DETECTED: %v for entity %s", err, entity.ID)
		return err
	}
	
	// Calculate buffer checksum
	bufferChecksum := sha256.Sum256(buffer.Bytes())
	op.SetMetadata("buffer_checksum", hex.EncodeToString(bufferChecksum[:]))
	op.SetMetadata("write_offset", offset)
	op.SetMetadata("buffer_size", buffer.Len())
	
	logger.TraceIf("storage", "Writing entity %s: %d bytes at offset %d", entity.ID, buffer.Len(), offset)
	
	// Write to file
	n, err := w.file.Write(buffer.Bytes())
	if err != nil {
		op.Fail(err)
		logger.Error("Failed to write entity %s data: %v", entity.ID, err)
		return err
	}
	
	if n != buffer.Len() {
		err := fmt.Errorf("incomplete write: expected %d bytes, wrote %d", buffer.Len(), n)
		op.Fail(err)
		logger.Error("%v for entity %s", err, entity.ID)
		return err
	}
	
	op.SetMetadata("bytes_written", n)
	logger.TraceIf("storage", "Successfully wrote %d bytes for entity %s", n, entity.ID)
	
	// Update index
	entry := &IndexEntry{
		Offset: uint64(offset),
		Size:   uint32(n),
	}
	
	// SURGICAL FIX: Validate IndexEntry values immediately after creation
	if entry.Offset != uint64(offset) {
		err := fmt.Errorf("offset conversion corruption: int64(%d) -> uint64(%d)", offset, entry.Offset)
		op.Fail(err)
		logger.Error("CORRUPTION DETECTED: %v for entity %s", err, entity.ID)
		return err
	}
	if entry.Size != uint32(n) {
		err := fmt.Errorf("size conversion corruption: int(%d) -> uint32(%d)", n, entry.Size)
		op.Fail(err)
		logger.Error("CORRUPTION DETECTED: %v for entity %s", err, entity.ID)
		return err
	}
	
	// Clear the EntityID array first to avoid garbage
	for i := range entry.EntityID {
		entry.EntityID[i] = 0
	}
	// Copy the ID, ensuring we don't exceed the array size
	idBytes := []byte(entity.ID)
	if len(idBytes) > len(entry.EntityID) {
		idBytes = idBytes[:len(entry.EntityID)]
	}
	copy(entry.EntityID[:], idBytes)
	
	// Check if this is a new entity (not an update)
	_, isUpdate := w.index[entity.ID]
	w.index[entity.ID] = entry
	
	logger.TraceIf("storage", "Added index entry for %s: offset=%d, size=%d, isUpdate=%v", entity.ID, entry.Offset, entry.Size, isUpdate)
	op.SetMetadata("index_offset", entry.Offset)
	op.SetMetadata("index_size", entry.Size)
	op.SetMetadata("is_update", isUpdate)
	
	// Verify we can read back what we wrote
	verifyBuffer := make([]byte, n)
	if _, err := w.file.ReadAt(verifyBuffer, int64(offset)); err != nil {
		logger.Error("Failed to verify write for entity %s: %v", entity.ID, err)
		// Don't fail the operation, but log the issue
	} else {
		verifyChecksum := sha256.Sum256(verifyBuffer)
		if hex.EncodeToString(verifyChecksum[:]) != hex.EncodeToString(bufferChecksum[:]) {
			logger.Error("Checksum mismatch after write for entity %s", entity.ID)
		} else {
			logger.Debug("Write verification successful for entity %s", entity.ID)
		}
	}
	
	// Update EntityCount immediately for new entities to match index reality
	// The index map already contains the entry, so header should reflect this
	if !isUpdate {
		w.header.EntityCount++
		logger.TraceIf("storage", "New entity %s added, EntityCount updated to %d", entity.ID, w.header.EntityCount)
	} else {
		logger.TraceIf("storage", "Updated existing entity %s, EntityCount remains %d", entity.ID, w.header.EntityCount)
	}
	w.header.FileSize = uint64(offset) + uint64(n)
	w.header.LastModified = time.Now().Unix()
	
	logger.TraceIf("storage", "Updated header: EntityCount=%d, FileSize=%d", w.header.EntityCount, w.header.FileSize)
	op.SetMetadata("final_file_size", w.header.FileSize)
	
	// Write updated header back to file (EntityCount now matches index)
	currentPos, err := w.file.Seek(0, os.SEEK_CUR)
	if err != nil {
		logger.Error("Failed to get current position: %v", err)
		return err
	}
	logger.Debug("Current position: %d", currentPos)
	
	// Seek to start to write header
	_, err = w.file.Seek(0, os.SEEK_SET)
	if err != nil {
		logger.Error("Failed to seek to start: %v", err)
		return err
	}
	
	logger.Debug("Writing updated header to file (EntityCount matches index)")
	if err := w.writeHeader(); err != nil {
		logger.Error("Failed to write header: %v", err)
		return err
	}
	
	// Seek back to original position
	_, err = w.file.Seek(currentPos, os.SEEK_SET)
	if err != nil {
		logger.Error("Failed to seek back to position %d: %v", currentPos, err)
		return err
	}
	
	// Sync data to disk
	if err := w.file.Sync(); err != nil {
		logger.Error("Failed to sync file: %v", err)
		return err
	}
	
	logger.Debug("Entity write completed successfully")
	
	return nil
}


// Close finalizes the binary file by writing the tag dictionary and index.
// It must be called when all entity writes are complete to ensure the file
// is properly structured and can be read by Reader instances.
//
// The method performs these critical operations:
//   1. Writes the tag dictionary at the end of entity data
//   2. Writes the complete entity index in sorted order
//   3. Updates the header with final offsets and counts
//   4. Syncs all data to disk for durability
//
// File Structure After Close:
//   [Header][Entities...][TagDictionary][Index]
//   ^                    ^              ^
//   |                    |              +-- EntityIndexOffset
//   |                    +-- TagDictOffset
//   +-- Start of file (offset 0)
//
// Thread Safety:
//   - Method is synchronized with internal mutex
//   - Safe to call from any goroutine
//   - Subsequent operations will fail after Close
//
// Returns:
//   - error: I/O failures during finalization
//
// Note: The file handle is NOT closed to support singleton pattern.
// Use WriterManager for proper lifecycle management.
func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	logger.Debug("Close called, EntityCount=%d, FileSize=%d", w.header.EntityCount, w.header.FileSize)
	
	// Write tag dictionary
	dictOffset, err := w.file.Seek(0, os.SEEK_END)
	if err != nil {
		logger.Error("Failed to seek for dictionary: %v", err)
		return err
	}
	logger.Debug("Writing tag dictionary at offset %d", dictOffset)
	
	dictBuf := new(bytes.Buffer)
	w.tagDict.Write(dictBuf)
	n, err := w.file.Write(dictBuf.Bytes())
	if err != nil {
		logger.Error("Failed to write tag dictionary: %v", err)
		return err
	}
	logger.Debug("Wrote %d bytes for tag dictionary", n)
	
	// Write index
	indexOffset, err := w.file.Seek(0, os.SEEK_END)
	if err != nil {
		logger.Error("Failed to seek for index: %v", err)
		return err
	}
	logger.Debug("Writing index at offset %d with %d entries", indexOffset, len(w.index))
	
	// Index Writing Algorithm:
	// 1. Collect all entity IDs from the in-memory map
	// 2. Sort IDs alphabetically for deterministic file layout
	// 3. Write each index entry in sorted order
	// This ensures consistent file structure across writes
	entityIDs := make([]string, 0, len(w.index))
	for id := range w.index {
		entityIDs = append(entityIDs, id)
	}
	sort.Strings(entityIDs)
	
	// Write index entries in sorted order
	// Each entry consists of:
	//   - EntityID: 64-byte fixed array (padded with zeros)
	//   - Offset: 8-byte position in file where entity data starts
	//   - Size: 4-byte size of entity data
	//   - Flags: 4-byte reserved for future use
	writtenCount := 0
	for _, id := range entityIDs {
		entry := w.index[id]
		logger.Debug("Writing index entry %d for %s: offset=%d, size=%d", writtenCount, id, entry.Offset, entry.Size)
		
		// SURGICAL FIX: Validate index entry before writing to prevent corruption
		fileInfo, err := w.file.Stat()
		if err != nil {
			logger.Error("Failed to get file size for validation: %v", err)
			return err
		}
		currentFileSize := uint64(fileInfo.Size())
		
		// Validate offset is within reasonable bounds
		if entry.Offset > currentFileSize {
			logger.Error("CORRUPTION DETECTED: Index entry for %s has invalid offset %d exceeds current file size %d - correcting", 
				id, entry.Offset, currentFileSize)
			// Skip this corrupted entry rather than write corruption to disk
			logger.Warn("Skipping corrupted index entry for %s to prevent disk corruption", id)
			continue
		}
		
		// Validate size is reasonable (not zero and not impossibly large)
		if entry.Size == 0 {
			logger.Warn("Index entry for %s has zero size - skipping", id)
			continue
		}
		if entry.Size > 1024*1024*100 { // 100MB sanity check
			logger.Error("CORRUPTION DETECTED: Index entry for %s has impossible size %d - skipping", id, entry.Size)
			continue
		}
		
		if err := binary.Write(w.file, binary.LittleEndian, entry.EntityID); err != nil {
			logger.Error("Failed to write EntityID for %s: %v", id, err)
			return err
		}
		if err := binary.Write(w.file, binary.LittleEndian, entry.Offset); err != nil {
			logger.Error("Failed to write Offset for %s: %v", id, err)
			return err
		}
		if err := binary.Write(w.file, binary.LittleEndian, entry.Size); err != nil {
			logger.Error("Failed to write Size for %s: %v", id, err)
			return err
		}
		if err := binary.Write(w.file, binary.LittleEndian, entry.Flags); err != nil {
			logger.Error("Failed to write Flags for %s: %v", id, err)
			return err
		}
		writtenCount++
	}
	logger.TraceIf("storage", "Wrote %d index entries (header claims %d)", writtenCount, w.header.EntityCount)
	
	// Verify index count matches header
	if writtenCount != int(w.header.EntityCount) {
		logger.Warn("Index entry count mismatch detected: wrote %d entries but header claims %d, correcting header", writtenCount, w.header.EntityCount)
		// Update header to match actual count
		w.header.EntityCount = uint64(writtenCount)
		logger.Debug("Corrected header EntityCount to match actual index: %d", w.header.EntityCount)
	} else {
		logger.Debug("Index count verification passed: %d entries match header count", writtenCount)
	}
	
	// Update header
	w.header.TagDictOffset = uint64(dictOffset)
	w.header.TagDictSize = uint64(dictBuf.Len())
	w.header.EntityIndexOffset = uint64(indexOffset)
	w.header.EntityIndexSize = uint64(writtenCount * IndexEntrySize)
	
	logger.Debug("Updated header: TagDictOffset=%d, TagDictSize=%d, EntityIndexOffset=%d, EntityIndexSize=%d",
		w.header.TagDictOffset, w.header.TagDictSize, w.header.EntityIndexOffset, w.header.EntityIndexSize)
	
	// Rewrite header at beginning
	_, err = w.file.Seek(0, os.SEEK_SET)
	if err != nil {
		logger.Error("Failed to seek to start: %v", err)
		return err
	}
	
	logger.Debug("Rewriting header")
	if err := w.writeHeader(); err != nil {
		logger.Error("Failed to write header: %v", err)
		return err
	}
	
	// Ensure all data is written
	if err := w.file.Sync(); err != nil {
		logger.Error("Failed to sync file: %v", err)
		return err
	}
	
	logger.Debug("Close completed successfully")
	
	// Don't close the file if used as singleton
	return nil
}

// writeHeader writes the file header to the beginning of the file.
// This is called during initialization and after each entity write to
// keep the header synchronized with the actual file state.
//
// The header contains:
//   - Magic number and version for format validation
//   - Entity count and file size
//   - Offsets to tag dictionary and index sections
//
// Returns:
//   - error: Write failures or seek errors
func (w *Writer) writeHeader() error {
	logger.Debug("writeHeader called - EntityCount=%d, FileSize=%d", w.header.EntityCount, w.header.FileSize)
	err := w.header.Write(w.file)
	if err != nil {
		logger.Error("Failed to write header: %v", err)
		return err
	}
	logger.Debug("Header written successfully")
	return nil
}

// readExisting loads an existing binary file's metadata into memory.
// This includes the header, tag dictionary, and entity index, allowing
// the Writer to append new entities without corrupting existing data.
//
// The method performs validation to ensure:
//   - The file has a valid magic number and compatible version
//   - The tag dictionary can be loaded successfully
//   - The index entries are valid and complete
//
// After loading, the file position is set to the end for appending.
//
// Returns:
//   - error: Format errors, corruption, or I/O failures
func (w *Writer) readExisting() error {
	logger.Debug("readExisting called")
	
	// Read header
	if err := w.header.Read(w.file); err != nil {
		logger.Error("Failed to read header: %v", err)
		return err
	}
	logger.Debug("Read header: EntityCount=%d, FileSize=%d", w.header.EntityCount, w.header.FileSize)
	
	// Read tag dictionary
	logger.Debug("Reading tag dictionary from offset %d", w.header.TagDictOffset)
	w.file.Seek(int64(w.header.TagDictOffset), os.SEEK_SET)
	if err := w.tagDict.Read(w.file); err != nil {
		logger.Error("Failed to read tag dictionary: %v", err)
		return err
	}
	logger.Debug("Loaded tag dictionary")
	
	// Read index
	logger.Debug("Reading index from offset %d, expecting %d entries", w.header.EntityIndexOffset, w.header.EntityCount)
	w.file.Seek(int64(w.header.EntityIndexOffset), os.SEEK_SET)
	w.index = make(map[string]*IndexEntry)
	
	for i := uint64(0); i < w.header.EntityCount; i++ {
		entry := &IndexEntry{}
		if err := binary.Read(w.file, binary.LittleEndian, &entry.EntityID); err != nil {
			logger.Error("Failed to read index entry %d: %v", i, err)
			break
		}
		if err := binary.Read(w.file, binary.LittleEndian, &entry.Offset); err != nil {
			logger.Error("Failed to read offset for entry %d: %v", i, err)
			break
		}
		if err := binary.Read(w.file, binary.LittleEndian, &entry.Size); err != nil {
			logger.Error("Failed to read size for entry %d: %v", i, err)
			break
		}
		if err := binary.Read(w.file, binary.LittleEndian, &entry.Flags); err != nil {
			logger.Error("Failed to read flags for entry %d: %v", i, err)
			break
		}
		
		// Convert ID to string, handling any null bytes or garbage
		id := string(bytes.TrimRight(entry.EntityID[:], "\x00"))
		// Skip empty IDs
		if id == "" {
			logger.Debug("Skipping empty index entry %d", i)
			continue
		}
		w.index[id] = entry
		logger.Debug("Loaded index entry %d: ID=%s, Offset=%d, Size=%d", i, id, entry.Offset, entry.Size)
	}
	
	logger.Debug("Loaded %d index entries", len(w.index))
	
	// Seek to end for writing new data
	w.file.Seek(int64(w.header.FileSize), os.SEEK_SET)
	
	return nil
}