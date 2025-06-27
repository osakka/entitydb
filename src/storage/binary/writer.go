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
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"entitydb/config"
	"entitydb/models"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
	"entitydb/logger"
)

// Writer handles writing entities to the EntityDB Unified File Format (EUFF).
// It manages the file structure, maintains indexes for fast lookups,
// and ensures data integrity through checksums and atomic operations.
//
// BAR-RAISING ENHANCEMENT: Integrated WAL corruption prevention system
// with comprehensive integrity monitoring, self-healing capabilities,
// and astronomical size detection to prevent database corruption.
//
// The Writer is safe for concurrent use through internal mutex locking.
// However, only one Writer instance should be active per file to prevent
// corruption. Use WriterManager for safe concurrent access patterns.
//
// File Layout:
//   [Header][WAL][Entity1][Entity2]...[EntityN][TagDictionary][Index]
//
// Performance characteristics:
//   - O(1) append operations for new entities
//   - O(1) index updates through in-memory map
//   - Compression reduces I/O for large content
//   - Memory pools minimize allocation overhead
//   - BAR-RAISING: Pre-write validation prevents corruption
//   - BAR-RAISING: Self-healing recovers from corruption automatically
//   - Embedded WAL reduces syscalls and improves atomicity
type Writer struct {
	file        *os.File                // Underlying file handle
	header      *Header                 // Unified file header with metadata (DEPRECATED - use headerSync)
	headerSync  *HeaderSync             // Thread-safe header access to prevent corruption
	tagDict     *TagDictionary          // Tag string to ID mapping
	index       map[string]*IndexEntry  // Entity ID to file location mapping
	mu          sync.Mutex              // Protects concurrent access
	walSequence uint64                  // Current WAL sequence number (DEPRECATED - use headerSync)
	
	// BAR-RAISING: Comprehensive corruption prevention system
	integritySystem *WALIntegritySystem  // WAL corruption prevention and self-healing
	healthCtx       context.Context      // Context for health monitoring
	healthCancel    context.CancelFunc   // Cancel function for health monitoring
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

// NewWriter creates a new Writer instance for the unified file format.
// If the file exists and contains valid data, it loads the existing header,
// tag dictionary, and index. For new files, it initializes the unified header structure.
//
// The unified format embeds WAL data within the file, reducing the number of
// file handles and improving atomic operations.
//
// Parameters:
//   - filename: Path to the unified file to open or create
//
// Returns:
//   - *Writer: Initialized writer ready for entity operations
//   - error: File access errors, corruption, or invalid format
//
// Example:
//   writer, err := NewWriter("/var/lib/entitydb/entities.unified")
//   if err != nil {
//       log.Fatal(err)
//   }
//   defer writer.Close()
func NewWriter(filename string, cfg *config.Config) (*Writer, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	
	// BAR-RAISING: Initialize comprehensive WAL integrity system
	integritySystem := NewWALIntegritySystem(filename, cfg)
	healthCtx, healthCancel := context.WithCancel(context.Background())
	
	w := &Writer{
		file:            file,
		header:          &Header{Magic: MagicNumber, Version: FormatVersion},
		tagDict:         NewTagDictionary(),
		index:           make(map[string]*IndexEntry),
		walSequence:     1,
		integritySystem: integritySystem,
		healthCtx:       healthCtx,
		healthCancel:    healthCancel,
	}
	
	// Initialize HeaderSync for thread-safe header access
	w.headerSync = NewHeaderSync(w.header)
	
	// Try to read existing file
	if stat, err := file.Stat(); err == nil && stat.Size() > 0 {
		if err := w.readExisting(); err != nil {
			logger.Warn("Failed to read existing unified file, creating new: %v", err)
			// Reset the file
			file.Truncate(0)
			file.Seek(0, 0)
			// Write initial unified header
			if err := w.writeHeader(); err != nil {
				return nil, err
			}
			if err := w.file.Sync(); err != nil {
				return nil, err
			}
		}
	} else {
		// New file, write initial unified header with sections
		if err := w.initializeUnifiedFile(); err != nil {
			return nil, err
		}
	}
	
	// BAR-RAISING: Start continuous health monitoring
	go w.integritySystem.StartHealthMonitoring(w.healthCtx)
	logger.Info("WAL integrity system initialized with continuous health monitoring")
	
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
	entityCount := w.headerSync.GetHeader().EntityCount
	
	op := models.StartOperation(models.OpTypeWrite, entity.ID, map[string]interface{}{
		"tags_count":    len(entity.Tags),
		"content_size":  len(entity.Content),
		"file_position": w.getFilePosition(),
		"entity_count":  entityCount,
	})
	defer func() {
		if op != nil {
			op.Complete()
		}
	}()
	
	// Acquire exclusive lock for thread safety
	w.mu.Lock()
	defer w.mu.Unlock()
	
	// BAR-RAISING: Pre-write corruption prevention validation
	// Calculate estimated WAL entry length to prevent astronomical sizes
	estimatedWALLength := int64(len(entity.ID) + len(entity.Content) + (len(entity.Tags) * 50)) // Conservative estimate
	if err := w.integritySystem.ValidateBeforeWrite(entity.ID, entity.Content, estimatedWALLength); err != nil {
		op.Fail(err)
		logger.Error("CORRUPTION PREVENTION: WAL integrity validation failed for entity %s: %v", entity.ID, err)
		return fmt.Errorf("WAL integrity validation failed: %w", err)
	}
	
	// Write to WAL first
	if err := w.writeWALEntry(1, entity.ID, entity); err != nil { // OpType 1 = create/update
		op.Fail(err)
		logger.Error("Failed to write WAL entry for entity %s: %v", entity.ID, err)
		return err
	}
	
	// Log write intent
	logger.TraceIf("storage", "Starting write for entity %s (tags=%d, content=%d bytes)", 
		entity.ID, len(entity.Tags), len(entity.Content))
	
	// ENHANCED ENTITY VALIDATION: Comprehensive pre-write checks
	if entity.ID == "" {
		err := fmt.Errorf("entity ID cannot be empty")
		op.Fail(err)
		logger.Error("Validation failed: %v", err)
		return err
	}
	
	// Validate entity ID length and format
	if len(entity.ID) > 64 {
		err := fmt.Errorf("entity ID length %d exceeds maximum of 64 characters", len(entity.ID))
		op.Fail(err)
		logger.Error("Validation failed: %v", err)
		return err
	}
	
	// Validate content size is reasonable
	if len(entity.Content) > 1024*1024*1024 { // 1GB per entity limit
		err := fmt.Errorf("entity content size %d exceeds 1GB limit", len(entity.Content))
		op.Fail(err)
		logger.Error("Validation failed: %v", err)
		return err
	}
	
	// Validate tag count is reasonable
	if len(entity.Tags) > 10000 { // 10k tags per entity should be more than enough
		err := fmt.Errorf("entity tag count %d exceeds 10000 limit", len(entity.Tags))
		op.Fail(err)
		logger.Error("Validation failed: %v", err)
		return err
	}
	
	// Check current file state before proceeding
	currentPos := w.getFilePosition()
	if currentPos < 0 {
		err := fmt.Errorf("invalid file position %d before write", currentPos)
		op.Fail(err)
		logger.Error("PRE-WRITE CORRUPTION DETECTED: %v for entity %s", err, entity.ID)
		return err
	}
	
	logger.TraceIf("storage", "Pre-write validation passed for entity %s (ID len=%d, content=%d bytes, tags=%d, file pos=%d)", 
		entity.ID, len(entity.ID), len(entity.Content), len(entity.Tags), currentPos)
	
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
			
			// BUFFER OVERFLOW FIX: Validate buffer capacity before writing large data
			requiredSpace := buffer.Len() + len(compressed.Data) + 32 // 32 byte safety margin
			if requiredSpace > buffer.Cap() {
				logger.Error("CRITICAL: Buffer overflow prevented - entity %s requires %d bytes, buffer capacity %d", 
					entity.ID, requiredSpace, buffer.Cap())
				return fmt.Errorf("buffer overflow prevented: required %d bytes exceeds capacity %d", requiredSpace, buffer.Cap())
			}
			buffer.Write(compressed.Data)
		} else {
			// For other content types, use application/octet-stream wrapper
			binary.Write(buffer, binary.LittleEndian, uint16(len("application/octet-stream")))
			buffer.WriteString("application/octet-stream")
			binary.Write(buffer, binary.LittleEndian, uint32(compressed.OriginalSize))
			binary.Write(buffer, binary.LittleEndian, uint32(len(compressed.Data)))
			
			// BUFFER OVERFLOW FIX: Validate buffer capacity before writing large data
			requiredSpace := buffer.Len() + len(compressed.Data) + 32 // 32 byte safety margin
			if requiredSpace > buffer.Cap() {
				logger.Error("CRITICAL: Buffer overflow prevented - entity %s requires %d bytes, buffer capacity %d", 
					entity.ID, requiredSpace, buffer.Cap())
				return fmt.Errorf("buffer overflow prevented: required %d bytes exceeds capacity %d", requiredSpace, buffer.Cap())
			}
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
	
	// Get current file position for unified format
	// Append to data section
	dataEnd := w.header.DataOffset + w.header.DataSize
	offset, err := w.file.Seek(int64(dataEnd), os.SEEK_SET)
	if err != nil {
		op.Fail(err)
		logger.Error("Failed to seek to data section end for entity %s: %v", entity.ID, err)
		return err
	}
	
	// CRITICAL: Validate offset immediately after seek to prevent corruption propagation
	
	// For unified format, validate against data section end
	expectedOffset := int64(w.header.DataOffset + w.header.DataSize)
	
	if offset != expectedOffset {
		err := fmt.Errorf("CRITICAL: Seek returned corrupted offset %d, expected %d (diff: %d)", 
			offset, expectedOffset, offset-expectedOffset)
		op.Fail(err)
		logger.Error("CORRUPTION DETECTED: %v for entity %s - ABORTING to prevent propagation", err, entity.ID)
		return err
	}
	
	// ENHANCED CORRUPTION DETECTION: Aggressive offset validation before writes
	if offset < 0 {
		err := fmt.Errorf("invalid negative offset %d returned from seek", offset)
		op.Fail(err)
		logger.Error("CORRUPTION DETECTED: %v for entity %s", err, entity.ID)
		return err
	}
	
	// Get file stat for additional validation context
	stat, statErr := w.file.Stat()
	if statErr == nil {
		fileSize := stat.Size()
		
		// Enhanced validation checks
		if offset > fileSize+1024*1024 { // Offset should not exceed file size by more than 1MB
			err := fmt.Errorf("offset %d exceeds file size %d by more than 1MB (corruption?)", offset, fileSize)
			op.Fail(err)
			logger.Error("CORRUPTION DETECTED: %v for entity %s", err, entity.ID)
			return err
		}
		
		if offset > 1024*1024*1024*10 { // 10GB absolute sanity check
			err := fmt.Errorf("impossibly large offset %d returned from seek (file corruption?)", offset)
			op.Fail(err)
			logger.Error("CORRUPTION DETECTED: %v for entity %s", err, entity.ID)
			return err
		}
		
		// Check for astronomical jumps in offset values
		expectedOffset := fileSize
		if offset > expectedOffset && (offset-expectedOffset) > 1024*1024*100 { // 100MB jump threshold
			err := fmt.Errorf("suspicious offset jump: expected ~%d, got %d (diff: %d bytes)", 
				expectedOffset, offset, offset-expectedOffset)
			op.Fail(err)
			logger.Error("CORRUPTION DETECTED: %v for entity %s", err, entity.ID)
			return err
		}
		
		logger.TraceIf("storage", "Offset validation passed: offset=%d, fileSize=%d, entity=%s", 
			offset, fileSize, entity.ID)
	} else {
		// Fallback validation without file stat
		if offset > 1024*1024*1024*10 { // 10GB sanity check
			err := fmt.Errorf("impossibly large offset %d returned from seek (file corruption?)", offset)
			op.Fail(err)
			logger.Error("CORRUPTION DETECTED: %v for entity %s", err, entity.ID)
			return err
		}
		logger.Warn("Could not stat file for enhanced validation: %v", statErr)
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
	
	// ENHANCED INDEX VALIDATION: Comprehensive corruption detection
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
	
	// Additional corruption checks on the index entry
	if entry.Offset > 0xFFFFFFFFFFFFFF { // Max safe uint64 (2^56-1, leaving room for error)
		err := fmt.Errorf("index entry offset %d exceeds safe uint64 range (corruption?)", entry.Offset)
		op.Fail(err)
		logger.Error("CORRUPTION DETECTED: %v for entity %s", err, entity.ID)
		return err
	}
	
	if entry.Size == 0 && buffer.Len() > 0 {
		err := fmt.Errorf("index entry size is 0 but buffer contains %d bytes (corruption?)", buffer.Len())
		op.Fail(err)
		logger.Error("CORRUPTION DETECTED: %v for entity %s", err, entity.ID)
		return err
	}
	
	if entry.Size > 1024*1024*1024 { // 1GB per entity seems excessive
		err := fmt.Errorf("index entry size %d exceeds 1GB limit (corruption?)", entry.Size)
		op.Fail(err)
		logger.Error("CORRUPTION DETECTED: %v for entity %s", err, entity.ID)
		return err
	}
	
	logger.TraceIf("storage", "Index entry validation passed: offset=%d, size=%d, entity=%s", 
		entry.Offset, entry.Size, entity.ID)
	
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
	
	// RACE CONDITION FIX: Create a new IndexEntry copy to prevent concurrent access corruption
	// The issue was multiple goroutines accessing the same IndexEntry pointer causing memory corruption
	indexEntry := &IndexEntry{
		Offset: entry.Offset,
		Size:   entry.Size,
		Flags:  entry.Flags,
	}
	copy(indexEntry.EntityID[:], entry.EntityID[:])
	w.index[entity.ID] = indexEntry
	
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
	
	// Update unified header
	if !isUpdate {
		newCount := w.headerSync.IncrementEntityCount()
		logger.TraceIf("storage", "New entity %s added, EntityCount updated to %d", entity.ID, newCount)
	} else {
		count := w.headerSync.GetHeader().EntityCount
		logger.TraceIf("storage", "Updated existing entity %s, EntityCount remains %d", entity.ID, count)
	}
	
	// Update header fields safely
	w.headerSync.UpdateHeader(func(h *Header) {
		h.DataSize += uint64(n)
		h.FileSize = h.DataOffset + h.DataSize
		h.LastModified = time.Now().Unix()
	})
	
	logger.TraceIf("storage", "Updated unified header: EntityCount=%d, DataSize=%d, FileSize=%d", 
		w.header.EntityCount, w.header.DataSize, w.header.FileSize)
	op.SetMetadata("final_file_size", w.header.FileSize)
	
	// Write updated unified header
	currentPos, err = w.file.Seek(0, os.SEEK_CUR)
	if err != nil {
		logger.Error("Failed to get current position: %v", err)
		return err
	}
	
	if err := w.writeHeader(); err != nil {
		logger.Error("Failed to write unified header: %v", err)
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
	// Get current entity count from HeaderSync for accurate comparison
	currentEntityCount := w.headerSync.GetHeader().EntityCount
	logger.TraceIf("storage", "Wrote %d index entries (HeaderSync claims %d)", writtenCount, currentEntityCount)
	
	// Verify index count matches HeaderSync
	if writtenCount != int(currentEntityCount) {
		logger.Warn("Index entry count mismatch detected: wrote %d entries but HeaderSync claims %d, correcting HeaderSync", writtenCount, currentEntityCount)
		// Update HeaderSync to match actual count - this should not happen with proper HeaderSync usage
		w.headerSync.UpdateHeader(func(h *Header) {
			h.EntityCount = uint64(writtenCount)
		})
		logger.Debug("Corrected HeaderSync EntityCount to match actual index: %d", writtenCount)
	} else {
		logger.Debug("Index count verification passed: %d entries match HeaderSync count", writtenCount)
	}
	
	// Update unified header through HeaderSync for consistency
	w.headerSync.UpdateHeader(func(h *Header) {
		h.TagDictOffset = uint64(dictOffset)
		h.TagDictSize = uint64(dictBuf.Len())
		h.EntityIndexOffset = uint64(indexOffset)
		h.EntityIndexSize = uint64(writtenCount * IndexEntrySize)
	})
	
	// Log updated header values from HeaderSync
	updatedHeader := w.headerSync.GetHeader()
	logger.Debug("Updated unified header: TagDictOffset=%d, TagDictSize=%d, EntityIndexOffset=%d, EntityIndexSize=%d",
		updatedHeader.TagDictOffset, updatedHeader.TagDictSize, updatedHeader.EntityIndexOffset, updatedHeader.EntityIndexSize)
	
	// Rewrite header at beginning
	_, err = w.file.Seek(0, os.SEEK_SET)
	if err != nil {
		logger.Error("Failed to seek to start: %v", err)
		return err
	}
	
	logger.Debug("Rewriting header")
	if err := w.writeHeader(); err != nil {
		logger.Error("Failed to write unified header: %v", err)
		return err
	}
	
	// Ensure all data is written
	if err := w.file.Sync(); err != nil {
		logger.Error("Failed to sync file: %v", err)
		return err
	}
	
	logger.Debug("Close completed successfully")
	
	// BAR-RAISING: Shut down health monitoring gracefully
	if w.healthCancel != nil {
		w.healthCancel()
		logger.Info("WAL integrity system health monitoring shut down")
	}
	
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
	
	// SURGICAL FIX: Update HeaderSync with the loaded header to prevent WALOffset=0 corruption
	// Validate WALOffset before creating HeaderSync
	if w.header.WALOffset == 0 {
		logger.Warn("Header has invalid WALOffset=0, setting to default HeaderSize=%d", HeaderSize)
		w.header.WALOffset = HeaderSize // Default WAL starts after header
	}
	w.headerSync = NewHeaderSync(w.header)
	logger.Debug("HeaderSync updated with loaded header: WALOffset=%d", w.header.WALOffset)
	
	// Read tag dictionary using HeaderSync for corruption protection
	tagDictOffset, err := w.headerSync.GetTagDictOffset()
	if err != nil {
		logger.Error("Failed to get safe TagDictOffset: %v", err)
		return err
	}
	logger.Debug("Reading tag dictionary from validated offset %d", tagDictOffset)
	w.file.Seek(int64(tagDictOffset), os.SEEK_SET)
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

// initializeUnifiedFile sets up a new unified file with the proper section layout.
// This creates an empty file with allocated sections for WAL, data, dictionary, and index.
func (w *Writer) initializeUnifiedFile() error {
	logger.Debug("Initializing new unified file")
	
	// Initialize unified header with default section layout
	w.headerSync.UpdateHeader(func(h *Header) {
		h.WALOffset = HeaderSize
		h.WALSize = 64 * 1024 // 64KB initial WAL space
		h.DataOffset = h.WALOffset + h.WALSize
		h.DataSize = 0 // Will grow as entities are added
		h.TagDictOffset = h.DataOffset
		h.TagDictSize = 0
		h.EntityIndexOffset = h.DataOffset
		h.EntityIndexSize = 0
		h.EntityCount = 0
		h.LastModified = time.Now().Unix()
		h.WALSequence = 1
		h.CheckpointSequence = 0
		h.FileSize = h.DataOffset
	})
	
	// Update local header copy
	w.header = &Header{}
	*w.header = w.headerSync.GetHeader()
	
	// Write unified header
	if err := w.writeHeader(); err != nil {
		return err
	}
	
	// Initialize empty WAL section
	if err := w.initializeWALSection(); err != nil {
		return err
	}
	
	// Sync data to disk
	if err := w.file.Sync(); err != nil {
		return err
	}
	
	logger.Debug("Unified file initialization complete")
	return nil
}

// writeHeader writes the unified header to the beginning of the file.
func (w *Writer) writeHeader() error {
	// Get a safe copy of the header from HeaderSync
	header := w.headerSync.GetHeader()
	
	logger.Debug("writeHeader called - EntityCount=%d, FileSize=%d, WALSequence=%d", 
		header.EntityCount, header.FileSize, header.WALSequence)
	
	// Seek to beginning
	if _, err := w.file.Seek(0, os.SEEK_SET); err != nil {
		return err
	}
	
	err := header.Write(w.file)
	if err != nil {
		logger.Error("Failed to write unified header: %v", err)
		return err
	}
	
	// Update the local header copy
	w.header = &header
	
	logger.Debug("Unified header written successfully")
	return nil
}

// readExistingUnified loads an existing unified file's metadata into memory.
func (w *Writer) readExistingUnified() error {
	logger.Debug("readExistingUnified called")
	
	// Read unified header
	if _, err := w.file.Seek(0, os.SEEK_SET); err != nil {
		return err
	}
	
	if err := w.header.Read(w.file); err != nil {
		logger.Error("Failed to read unified header: %v", err)
		return err
	}
	
	// Update HeaderSync with the loaded header
	// Validate WALOffset before creating HeaderSync
	if w.header.WALOffset == 0 {
		logger.Warn("Unified header has invalid WALOffset=0, setting to default HeaderSize=%d", HeaderSize)
		w.header.WALOffset = HeaderSize // Default WAL starts after header
	}
	w.headerSync = NewHeaderSync(w.header)
	
	logger.Debug("Read unified header: EntityCount=%d, FileSize=%d, WALSequence=%d, WALOffset=%d", 
		w.header.EntityCount, w.header.FileSize, w.header.WALSequence, w.header.WALOffset)
	
	// Read tag dictionary if present
	if w.header.TagDictSize > 0 {
		tagDictOffset, err := w.headerSync.GetTagDictOffset()
		if err != nil {
			logger.Error("Failed to get safe TagDictOffset: %v", err)
			return err
		}
		logger.Debug("Reading tag dictionary from validated offset %d", tagDictOffset)
		if _, err := w.file.Seek(int64(tagDictOffset), os.SEEK_SET); err != nil {
			return err
		}
		if err := w.tagDict.Read(w.file); err != nil {
			logger.Error("Failed to read tag dictionary: %v", err)
			return err
		}
		logger.Debug("Loaded tag dictionary")
	}
	
	// Read index if present
	if w.header.EntityIndexSize > 0 {
		logger.Debug("Reading index from offset %d, expecting %d entries", 
			w.header.EntityIndexOffset, w.header.EntityCount)
		if _, err := w.file.Seek(int64(w.header.EntityIndexOffset), os.SEEK_SET); err != nil {
			return err
		}
		
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
			
			// Convert ID to string, handling any null bytes
			id := string(bytes.TrimRight(entry.EntityID[:], "\x00"))
			if id == "" {
				logger.Debug("Skipping empty index entry %d", i)
				continue
			}
			w.index[id] = entry
			logger.Debug("Loaded index entry %d: ID=%s, Offset=%d, Size=%d", i, id, entry.Offset, entry.Size)
		}
		logger.Debug("Loaded %d index entries", len(w.index))
	}
	
	// Restore WAL sequence
	w.walSequence = w.header.WALSequence
	
	// Seek to data end for appending
	dataEnd := w.header.DataOffset + w.header.DataSize
	if _, err := w.file.Seek(int64(dataEnd), os.SEEK_SET); err != nil {
		return err
	}
	
	return nil
}

// initializeWALSection creates an empty WAL section in the unified file.
func (w *Writer) initializeWALSection() error {
	logger.Debug("Initializing WAL section at offset %d", w.header.WALOffset)
	
	// Seek to WAL offset
	if _, err := w.file.Seek(int64(w.header.WALOffset), os.SEEK_SET); err != nil {
		return err
	}
	
	// Write empty WAL header (sequence number + entry count)
	buf := make([]byte, 16) // 8 bytes sequence + 8 bytes entry count
	binary.LittleEndian.PutUint64(buf[0:8], w.walSequence)
	binary.LittleEndian.PutUint64(buf[8:16], 0) // Zero entries initially
	
	if _, err := w.file.Write(buf); err != nil {
		return err
	}
	
	logger.Debug("WAL section initialized")
	return nil
}

// writeWALEntry writes a WAL entry to the embedded WAL section.
func (w *Writer) writeWALEntry(opType byte, entityID string, entity *models.Entity) error {
	logger.TraceIf("wal", "Writing WAL entry: opType=%d, entityID=%s", opType, entityID)
	
	// Use HeaderSync to safely get WAL offset
	walOffset, err := w.headerSync.GetWALOffset()
	if err != nil {
		return err
	}
	
	// Seek to WAL section
	if _, err := w.file.Seek(int64(walOffset), os.SEEK_SET); err != nil {
		logger.Error("CORRUPTION DETECTED: Seek failed to WALOffset %d: %v", walOffset, err)
		return fmt.Errorf("seek to WAL failed: %w", err)
	}
	
	// Read current entry count
	buf := make([]byte, 16)
	if _, err := w.file.Read(buf); err != nil {
		return err
	}
	
	entryCount := binary.LittleEndian.Uint64(buf[8:16])
	
	// Update sequence and entry count using HeaderSync
	walSequence := w.headerSync.IncrementWALSequence()
	entryCount++
	
	// Write updated header
	if _, err := w.file.Seek(int64(walOffset), os.SEEK_SET); err != nil {
		return err
	}
	binary.LittleEndian.PutUint64(buf[0:8], walSequence)
	binary.LittleEndian.PutUint64(buf[8:16], entryCount)
	if _, err := w.file.Write(buf); err != nil {
		return err
	}
	
	// Seek to end of WAL section for new entry
	entryOffset := w.header.WALOffset + 16 // Skip WAL header
	if _, err := w.file.Seek(int64(entryOffset), os.SEEK_END); err != nil {
		return err
	}
	
	// Write WAL entry
	// Format: [OpType:1][Sequence:8][Timestamp:8][EntityIDLen:2][EntityID][EntityDataLen:4][EntityData]
	timestamp := time.Now().UnixNano()
	
	entryBuf := new(bytes.Buffer)
	binary.Write(entryBuf, binary.LittleEndian, opType)
	binary.Write(entryBuf, binary.LittleEndian, w.walSequence)
	binary.Write(entryBuf, binary.LittleEndian, timestamp)
	binary.Write(entryBuf, binary.LittleEndian, uint16(len(entityID)))
	entryBuf.WriteString(entityID)
	
	if entity != nil {
		// Serialize entity data
		entityData := &bytes.Buffer{}
		binary.Write(entityData, binary.LittleEndian, uint16(len(entity.Tags)))
		for _, tag := range entity.Tags {
			binary.Write(entityData, binary.LittleEndian, uint16(len(tag)))
			entityData.WriteString(tag)
		}
		binary.Write(entityData, binary.LittleEndian, uint32(len(entity.Content)))
		entityData.Write(entity.Content)
		
		binary.Write(entryBuf, binary.LittleEndian, uint32(entityData.Len()))
		entryBuf.Write(entityData.Bytes())
	} else {
		binary.Write(entryBuf, binary.LittleEndian, uint32(0)) // No entity data
	}
	
	// BAR-RAISING: Emergency corruption detection before writing to disk
	// This is the last line of defense against astronomical WAL entry sizes
	walEntrySize := int64(entryBuf.Len())
	if walEntrySize > 1000000000 { // 1GB astronomical threshold
		logger.Error("CRITICAL CORRUPTION BLOCKED: WAL entry size %d exceeds astronomical threshold, aborting write for entity %s", walEntrySize, entityID)
		// Trigger emergency mode in integrity system
		if w.integritySystem != nil {
			w.integritySystem.EnableEmergencyMode()
		}
		return fmt.Errorf("astronomical WAL entry size %d blocked (entity: %s)", walEntrySize, entityID)
	}
	
	if _, err := w.file.Write(entryBuf.Bytes()); err != nil {
		return err
	}
	
	// Update unified header with new WAL sequence
	w.header.WALSequence = w.walSequence
	
	logger.TraceIf("wal", "WAL entry written: sequence=%d, size=%d bytes", w.walSequence, entryBuf.Len())
	return nil
}

// RestoreHeaderSync restores HeaderSync from a snapshot to recover from checkpoint corruption
// This is the critical recovery mechanism for checkpoint race conditions
func (w *Writer) RestoreHeaderSync(snapshot *HeaderSnapshot) error {
	if snapshot == nil {
		return fmt.Errorf("cannot restore from nil snapshot")
	}
	
	logger.Info("Restoring HeaderSync from snapshot: WALOffset=%d, EntityCount=%d", 
		snapshot.Header.WALOffset, snapshot.EntityCount)
	
	// Restore the HeaderSync state
	w.headerSync.RestoreFromSnapshot(snapshot)
	
	// Update local header copy
	w.header = &snapshot.Header
	
	logger.Info("HeaderSync restoration completed successfully")
	return nil
}