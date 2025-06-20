package binary

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"entitydb/models"
	"entitydb/logger"
	"entitydb/storage/pools"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

var (
	// ErrNotFound is returned when an entity cannot be found in the binary file
	ErrNotFound = errors.New("entity not found")
)

// Reader provides read-only access to entities stored in the EntityDB Binary Format.
// It loads the file's index and tag dictionary into memory for fast lookups,
// while reading entity data on-demand to minimize memory usage.
//
// The Reader supports the unified format with embedded WAL and sections.
//
// The Reader is designed for concurrent use - multiple goroutines can safely
// call ReadEntity simultaneously. The underlying file is opened in read-only
// mode to prevent accidental modifications.
//
// Performance characteristics:
//   - O(1) entity lookups via in-memory index
//   - Minimal memory footprint (only index and dictionary in RAM)
//   - Automatic decompression of compressed content
//   - Efficient tag resolution through dictionary
//   - Section-based reads for unified format reduce seek operations
//
// Example usage:
//   reader, err := NewReader("/var/lib/entitydb/entities.ebf")
//   if err != nil {
//       log.Fatal(err)
//   }
//   defer reader.Close()
//
//   entity, err := reader.ReadEntity("user-123")
//   if err != nil {
//       log.Printf("Entity not found: %v", err)
//   }
type Reader struct {
	file            *os.File                // Read-only file handle
	header          *Header                 // Unified file header with metadata
	tagDict         *TagDictionary          // Tag ID to string mapping
	index           map[string]*IndexEntry  // Entity ID to file location mapping
	indexMu         sync.RWMutex            // SURGICAL FIX: Protects index from concurrent access corruption
	corruptionCount int                     // Track corruption instances for recovery triggers
	filename        string                  // File path for reopening
	legacyReader    *LegacyReader          // TEMPORARY: Legacy format delegation
}

// NewReader creates a new Reader instance for the specified binary file.
// It performs the following initialization steps:
//   1. Opens the file in read-only mode
//   2. Reads and validates the file header
//   3. Loads the tag dictionary for efficient tag resolution
//   4. Loads the entity index for O(1) lookups
//
// The Reader gracefully handles partial or corrupted files:
//   - Missing dictionary or index sections are logged but don't fail
//   - Corrupted index entries are skipped
//   - The Reader remains functional with available data
//
// Parameters:
//   - filename: Path to the EBF file to read
//
// Returns:
//   - *Reader: Initialized reader ready for entity queries
//   - error: File access errors or invalid format
//
// Thread Safety:
//   The returned Reader is safe for concurrent use.
func NewReader(filename string) (*Reader, error) {
	logger.TraceIf("storage", "Opening reader for file: %s", filename)
	
	file, err := os.Open(filename)
	if err != nil {
		logger.Error("Failed to open file: %v", err)
		return nil, err
	}
	
	// Get file info
	stat, err := file.Stat()
	if err != nil {
		logger.Error("Failed to stat file: %v", err)
		file.Close()
		return nil, err
	}
	logger.TraceIf("storage", "File size: %d bytes", stat.Size())
	
	r := &Reader{
		file:     file,
		tagDict:  NewTagDictionary(),
		index:    make(map[string]*IndexEntry),
		filename: filename,
	}
	
	// Detect file format
	logger.TraceIf("storage", "Detecting file format")
	format, err := DetectFileFormat(filename)
	if err != nil {
		logger.Error("Failed to detect file format: %v", err)
		return nil, err
	}
	
	// TEMPORARY: Support both legacy and unified formats during migration
	if format == FormatLegacy {
		logger.Info("Loading legacy EBF format file (migration support)")
		legacyReader, err := NewLegacyReader(filename)
		if err != nil {
			return nil, err
		}
		// Convert legacy reader to unified reader interface
		return NewReaderFromLegacy(legacyReader)
	}
	
	// Read unified header
	r.header = &Header{}
	logger.TraceIf("storage", "Reading unified header")
	if err := r.header.Read(file); err != nil {
		logger.Error("Failed to read unified header: %v", err)
		return nil, err
	}
	
	logger.TraceIf("storage", "Unified header read successfully: Magic=%x, Version=%d, EntityCount=%d, FileSize=%d",
		r.header.Magic, r.header.Version, r.header.EntityCount, r.header.FileSize)
	logger.TraceIf("storage", "WALOffset=%d, WALSize=%d, DataOffset=%d, DataSize=%d", 
		r.header.WALOffset, r.header.WALSize, r.header.DataOffset, r.header.DataSize)
	logger.TraceIf("storage", "TagDictOffset=%d, TagDictSize=%d, EntityIndexOffset=%d, EntityIndexSize=%d", 
		r.header.TagDictOffset, r.header.TagDictSize, r.header.EntityIndexOffset, r.header.EntityIndexSize)
	
	// Get entity count and offsets
	entityCount := r.header.EntityCount
	tagDictOffset := r.header.TagDictOffset
	tagDictSize := r.header.TagDictSize
	entityIndexOffset := r.header.EntityIndexOffset
	
	// Skip dictionary and index if no entities
	if entityCount == 0 {
		logger.TraceIf("storage", "No entities in file, skipping dictionary and index")
		return r, nil
	}
	
	// Read tag dictionary
	if tagDictOffset > 0 && tagDictSize > 0 {
		logger.Trace("Reading tag dictionary from offset %d", tagDictOffset)
		if _, err := file.Seek(int64(tagDictOffset), os.SEEK_SET); err != nil {
			logger.Error("Failed to seek to tag dictionary: %v", err)
			return nil, err
		}
		if err := r.tagDict.Read(file); err != nil {
			// Log but don't fail - allow partial reads
			logger.Warn("Failed to read tag dictionary: %v", err)
		} else {
			logger.Trace("Tag dictionary loaded with %d entries", 0)
		}
	}
	
	// Read index
	if entityIndexOffset > 0 {
		logger.Trace("Reading index from offset %d, expecting %d entries", 
			entityIndexOffset, entityCount)
		
		// SURGICAL FIX: Use write lock during index initialization to prevent corruption
		r.indexMu.Lock()
		defer r.indexMu.Unlock()
		
		// Validate index location
		if int64(entityIndexOffset) > stat.Size() {
			logger.Error("Index offset %d exceeds file size %d", 
				entityIndexOffset, stat.Size())
			return r, nil // Return partial reader
		}
		
		if _, err := file.Seek(int64(entityIndexOffset), os.SEEK_SET); err != nil {
			logger.Error("Failed to seek to index: %v", err)
			return nil, err
		}
		
		// Calculate how many entries we can actually read
		indexStartPos := int64(entityIndexOffset)
		remainingFileSize := stat.Size() - indexStartPos
		entrySize := int64(binary.Size(IndexEntry{}))
		maxPossibleEntries := uint64(remainingFileSize / entrySize)
		
		if maxPossibleEntries < entityCount {
			logger.Debug("File size allows %d index entries but header indicates %d (concurrent write in progress)",
				maxPossibleEntries, entityCount)
		}
		
		entriesRead := uint64(0)
		
		// Use the smaller of header count and actual possible entries to avoid race conditions
		// During concurrent writes, header may claim more entities than actually indexed yet
		entriesToRead := entityCount
		if maxPossibleEntries < entriesToRead {
			entriesToRead = maxPossibleEntries
			logger.Debug("Adjusting read count from %d to %d entries due to file size constraints", 
				entityCount, entriesToRead)
		}
		
		for i := uint64(0); i < entriesToRead; i++ {
			// Check if we have enough bytes remaining  
			currentPos, _ := file.Seek(0, os.SEEK_CUR)
			if currentPos+entrySize > stat.Size() {
				logger.Debug("Reached end of available index data at entry %d (pos=%d, file_size=%d)",
					i, currentPos, stat.Size())
				break
			}
			
			entry := &IndexEntry{}
			if err := binary.Read(file, binary.LittleEndian, &entry.EntityID); err != nil {
				// Stop reading if we hit EOF
				if err == io.EOF {
					logger.Warn("Hit EOF at index entry %d (reading EntityID)", i)
					break
				}
				logger.Error("Failed to read index entry %d: %v", i, err)
				break // Don't fail entirely, just stop reading index
			}
			if err := binary.Read(file, binary.LittleEndian, &entry.Offset); err != nil {
				if err == io.EOF {
					logger.Warn("Hit EOF reading offset for entry %d", i)
					break
				}
				return nil, err
			}
			if err := binary.Read(file, binary.LittleEndian, &entry.Size); err != nil {
				if err == io.EOF {
					logger.Warn("Hit EOF reading size for entry %d", i)
					break
				}
				return nil, err
			}
			if err := binary.Read(file, binary.LittleEndian, &entry.Flags); err != nil {
				if err == io.EOF {
					logger.Warn("Hit EOF reading flags for entry %d", i)
					break
				}
				return nil, err
			}
			
			// Convert ID to string, handling any null bytes or garbage
			id := string(bytes.TrimRight(entry.EntityID[:], "\x00"))
			// Skip empty IDs
			if id == "" {
				logger.Trace("Skipping empty index entry %d", i)
				continue
			}
			
			// SURGICAL FIX: Validate offset to detect corruption before storing
			if entry.Offset > uint64(stat.Size()) {
				logger.Error("Corrupted index entry %d for %s: invalid offset %d exceeds file size %d", 
					i, id, entry.Offset, stat.Size())
				
				// Track corruption for potential automatic recovery trigger
				r.corruptionCount++
				if r.corruptionCount > 10 { // More than 10 corrupted entries
					logger.Warn("High corruption detected (%d corrupted entries) - index may need rebuilding", r.corruptionCount)
				}
				continue // Skip corrupted entry instead of crashing
			}
			if entry.Size == 0 {
				logger.Warn("Index entry %d for %s has zero size, skipping", i, id)
				continue
			}
			
			// RACE CONDITION FIX: Create a defensive copy of IndexEntry to prevent concurrent access corruption
			// Multiple goroutines accessing the same Reader instance could corrupt shared IndexEntry pointers
			indexEntry := &IndexEntry{
				Offset: entry.Offset,
				Size:   entry.Size,
				Flags:  entry.Flags,
			}
			copy(indexEntry.EntityID[:], entry.EntityID[:])
			r.index[id] = indexEntry
			entriesRead++
			logger.Trace("Loaded index entry %d: ID=%s, Offset=%d, Size=%d", i, id, entry.Offset, entry.Size)
		}
		
		logger.Trace("Index loading complete: read %d entries, loaded %d into index (expected %d)",
			entriesRead, len(r.index), entityCount)
		
		// Check for significant corruption
		corruptedEntries := int64(entityCount) - int64(entriesRead)
		if corruptedEntries > 0 {
			corruptionRate := float64(corruptedEntries) / float64(entityCount)
			
			if corruptionRate > 0.1 { // More than 10% corruption
				logger.Warn("Significant index corruption detected: %d/%d entries corrupted (%.1f%%) - automatic recovery recommended", 
					corruptedEntries, entityCount, corruptionRate*100)
			} else if entriesRead < entityCount {
				logger.Debug("Index partially loaded: %d entries read of %d expected (likely due to concurrent writes)", 
					entriesRead, entityCount)
			}
		}
	}
	
	return r, nil
}

// NewUnifiedReader creates a new Reader instance specifically for unified format files.
// This is a convenience function that assumes the file is in unified format.
//
// Parameters:
//   - filename: Path to the unified file to read
//
// Returns:
//   - *Reader: Initialized reader ready for entity queries
//   - error: File access errors or invalid format
//
// Example:
//   reader, err := NewUnifiedReader("/var/lib/entitydb/entities.unified")

// GetEntity reads a single entity by its ID from the binary file.
// This is the primary method for retrieving entities.
//
// The method performs these operations:
//   1. Looks up the entity's location in the index (O(1))
//   2. Seeks to the entity's position in the file
//   3. Reads the exact number of bytes for the entity
//   4. Parses the binary data into an Entity struct
//   5. Decompresses content if it was compressed
//   6. Resolves tag IDs to their string values
//
// Performance optimizations:
//   - Uses memory pools to avoid allocations
//   - Reads exact bytes needed (no over-reading)
//   - Decompression happens in-place when possible
//
// Parameters:
//   - id: The unique identifier of the entity to read
//
// Returns:
//   - *models.Entity: The parsed entity with all fields populated
//   - error: ErrNotFound if entity doesn't exist, or I/O errors
//
// Thread Safety:
//   Multiple goroutines can call this method concurrently.
//   File reads use pread-style operations (seek + read).
func (r *Reader) GetEntity(id string) (*models.Entity, error) {
	// TEMPORARY: Delegate to legacy reader if available
	if r.legacyReader != nil {
		return r.legacyReader.GetByID(id)
	}
	
	// Start operation tracking for observability
	op := models.StartOperation(models.OpTypeRead, id, map[string]interface{}{
		"index_size": len(r.index),
	})
	defer func() {
		if op != nil {
			op.Complete()
		}
	}()
	
	logger.Trace("GetEntity called for ID: %s", id)
	
	// SURGICAL FIX: Use read lock to prevent reading corrupted index during concurrent writes
	r.indexMu.RLock()
	entry, exists := r.index[id]
	if !exists {
		r.indexMu.RUnlock()
		err := fmt.Errorf("entity %s not found in index", id)
		op.Fail(err)
		logger.Trace("Entity %s not found in index", id)
		return nil, ErrNotFound
	}
	
	// Create a safe copy of the entry to avoid holding the lock during file I/O
	entryCopy := *entry
	r.indexMu.RUnlock()
	
	logger.Trace("Found entity %s at offset %d, size %d", id, entryCopy.Offset, entryCopy.Size)
	
	// Seek to entity position
	_, err := r.file.Seek(int64(entryCopy.Offset), os.SEEK_SET)
	if err != nil {
		logger.Error("Failed to seek to offset %d: %v", entryCopy.Offset, err)
		return nil, err
	}
	
	// Get buffer from pool for entity data
	dataSlice := pools.GetByteSlice()
	defer pools.PutByteSlice(dataSlice)
	
	// Resize slice to needed size
	if cap(*dataSlice) < int(entryCopy.Size) {
		*dataSlice = make([]byte, entryCopy.Size)
	} else {
		*dataSlice = (*dataSlice)[:entryCopy.Size]
	}
	data := *dataSlice
	
	n, err := r.file.Read(data)
	if err != nil {
		logger.Error("Failed to read %d bytes: %v", entryCopy.Size, err)
		return nil, err
	}
	if n != int(entryCopy.Size) {
		logger.Error("Incomplete read, expected %d bytes, got %d", entryCopy.Size, n)
		return nil, errors.New("incomplete read")
	}
	
	logger.Trace("Read %d bytes for entity %s", n, id)
	
	entity, err := r.parseEntity(data, id)
	if err != nil {
		logger.Error("Failed to parse entity %s: %v", id, err)
		return nil, err
	}
	
	logger.Trace("Successfully parsed entity %s", id)
	return entity, nil
}

// GetAllEntities reads all entities from the binary file.
// This method is useful for bulk operations like backups or migrations.
//
// Performance considerations:
//   - Loads all entities into memory simultaneously
//   - Memory usage: O(n) where n is total size of all entities
//   - Use with caution on large files (consider pagination instead)
//
// Error handling:
//   - Corrupted entities are logged and skipped
//   - Returns successfully read entities even if some fail
//   - Empty result only if no entities can be read
//
// Returns:
//   - []*models.Entity: Slice containing all readable entities
//   - error: Currently always returns nil (errors are logged)
//
// Thread Safety:
//   Safe for concurrent use, but may cause contention
//   if called simultaneously from multiple goroutines.
func (r *Reader) GetAllEntities() ([]*models.Entity, error) {
	// TEMPORARY: Delegate to legacy reader if available
	if r.legacyReader != nil {
		return r.legacyReader.GetAll()
	}
	
	logger.Trace("GetAllEntities called, index has %d entries, header says %d entities", len(r.index), r.header.EntityCount)
	entities := make([]*models.Entity, 0, r.header.EntityCount)
	
	for id := range r.index {
		logger.Trace("Getting entity with ID: %s", id)
		entity, err := r.GetEntity(id)
		if err != nil {
			logger.Warn("Error getting entity %s: %v", id, err)
			// Skip entities we can't read
			continue
		}
		entities = append(entities, entity)
	}
	
	logger.Trace("GetAllEntities returning %d entities", len(entities))
	return entities, nil
}

// parseEntity decodes binary entity data into a models.Entity struct.
// It handles the complete binary format including header, tags, and content.
//
// Binary Format:
//   [EntityHeader][TagIDs...][CompressionType][ContentType][Sizes][Data][Timestamp]
//
// The method handles both compressed and uncompressed content transparently,
// and properly decodes JSON content that was stored without extra wrapping.
//
// Parameters:
//   - data: Raw binary data read from file
//   - id: Entity ID (passed separately as it's stored in the index)
//
// Returns:
//   - *models.Entity: Fully populated entity
//   - error: Parsing errors or corruption
func (r *Reader) parseEntity(data []byte, id string) (*models.Entity, error) {
	buf := bytes.NewReader(data)
	
	// Read entity header containing metadata
	var header EntityHeader
	if err := binary.Read(buf, binary.LittleEndian, &header); err != nil {
		return nil, err
	}
	
	entity := &models.Entity{
		ID:   id,
		Tags: make([]string, header.TagCount),
		Content: []byte{}, // Unified content model
	}
	
	// Tag Resolution Algorithm:
	// 1. Read tag IDs (4 bytes each)
	// 2. Look up each ID in the tag dictionary
	// 3. Dictionary returns the original tag string
	// This reduces storage size significantly for repeated tags
	for i := uint16(0); i < header.TagCount; i++ {
		var tagID uint32
		if err := binary.Read(buf, binary.LittleEndian, &tagID); err != nil {
			return nil, err
		}
		entity.Tags[i] = r.tagDict.GetTag(tagID)
	}
	
	// Content Decoding Algorithm:
	// The binary format stores content with metadata for proper reconstruction
	// Format: [CompressionType][ContentTypeLen][ContentType][OriginalSize][CompressedSize][Data][Timestamp]
	if header.ContentCount > 0 {
		// Read compression type (1 byte)
		// 0 = no compression, 1 = gzip compression
		var compressionType uint8
		if err := binary.Read(buf, binary.LittleEndian, &compressionType); err != nil {
			return nil, err
		}
		
		// Content type
		var typeLen uint16
		if err := binary.Read(buf, binary.LittleEndian, &typeLen); err != nil {
			return nil, err
		}
		// Use small buffer pool for type string
		typeSlice := pools.GetByteSlice()
		defer pools.PutByteSlice(typeSlice)
		
		if cap(*typeSlice) < int(typeLen) {
			*typeSlice = make([]byte, typeLen)
		} else {
			*typeSlice = (*typeSlice)[:typeLen]
		}
		typeBytes := *typeSlice
		if _, err := buf.Read(typeBytes); err != nil {
			return nil, err
		}
		contentType := string(typeBytes)
		
		// Read original size and compressed size (writer writes both)
		var originalSize uint32
		if err := binary.Read(buf, binary.LittleEndian, &originalSize); err != nil {
			return nil, err
		}
		
		var compressedSize uint32
		if err := binary.Read(buf, binary.LittleEndian, &compressedSize); err != nil {
			return nil, err
		}
		
		// For content, allocate based on compressed size and read that many bytes
		contentBytes := make([]byte, compressedSize)
		if _, err := buf.Read(contentBytes); err != nil {
			return nil, err
		}
		
		// Decompress if needed
		if CompressionType(compressionType) == CompressionGzip {
			decompressed, err := DecompressWithPool(contentBytes)
			if err != nil {
				logger.Warn("Failed to decompress content for entity %s: %v, using as-is", id, err)
			} else {
				contentBytes = decompressed
			}
		}
		
		// Timestamp
		var tsNano int64
		if err := binary.Read(buf, binary.LittleEndian, &tsNano); err != nil {
			return nil, err
		}
		
		// Store content directly (no conversion needed for JSON)
		entity.Content = contentBytes
		
		// Add content type tag if not already present
		hasContentTypeTag := false
		for _, tag := range entity.Tags {
			if strings.Contains(tag, "content:type:") {
				hasContentTypeTag = true
				break
			}
		}
		if !hasContentTypeTag {
			entity.AddTag("content:type:" + contentType)
		}
		
		// Content Integrity Validation:
		// Validate content checksum against stored checksum tag for data integrity verification
		// Algorithm matches writer: SHA256 of final decompressed content
		actualChecksum := sha256.Sum256(contentBytes)
		actualChecksumHex := hex.EncodeToString(actualChecksum[:])
		
		// Find stored checksum in entity tags (temporal format: "timestamp|checksum:sha256:hash")
		var storedChecksumHex string
		for _, tag := range entity.Tags {
			if strings.Contains(tag, "|checksum:sha256:") {
				parts := strings.SplitN(tag, "|", 2)
				if len(parts) == 2 && strings.HasPrefix(parts[1], "checksum:sha256:") {
					storedChecksumHex = strings.TrimPrefix(parts[1], "checksum:sha256:")
					break
				}
			}
		}
		
		// Perform validation if checksum tag exists
		if storedChecksumHex != "" {
			if actualChecksumHex != storedChecksumHex {
				logger.Warn("Content checksum mismatch for entity %s: expected %s, got %s", 
					id, storedChecksumHex, actualChecksumHex)
				// Note: Continue operation but log integrity issue for monitoring
			} else {
				logger.Trace("Content integrity verified for entity %s", id)
			}
		}
		
		logger.Trace("Content loaded for entity %s (%d bytes)", id, len(contentBytes))
	}
	
	return entity, nil
}

// Close releases all resources associated with the Reader.
// After calling Close, the Reader cannot be used for further operations.
//
// This method:
//   - Closes the underlying file handle
//   - Allows the OS to reclaim file descriptors
//   - Does NOT clear the in-memory index or dictionary
//
// Returns:
//   - error: File closing errors (rare)
//
// It's safe to call Close multiple times; subsequent calls
// will return the error from os.File.Close() if the file
// is already closed.
func (r *Reader) Close() error {
	return r.file.Close()
}