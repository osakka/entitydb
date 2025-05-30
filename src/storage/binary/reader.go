package binary

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"entitydb/models"
	"entitydb/logger"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	ErrNotFound = errors.New("entity not found")
)

// Reader handles reading entities from binary format
type Reader struct {
	file    *os.File
	header  *Header
	tagDict *TagDictionary
	index   map[string]*IndexEntry
}

// NewReader creates a new reader for the given file
func NewReader(filename string) (*Reader, error) {
	logger.Trace("Opening reader for file: %s", filename)
	
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
	logger.Trace("File size: %d bytes", stat.Size())
	
	r := &Reader{
		file:    file,
		header:  &Header{},
		tagDict: NewTagDictionary(),
		index:   make(map[string]*IndexEntry),
	}
	
	// Read header
	logger.Trace("Reading header")
	if err := r.header.Read(file); err != nil {
		logger.Error("Failed to read header: %v", err)
		return nil, err
	}
	
	logger.Trace("Header read successfully: Magic=%x, Version=%d, EntityCount=%d, FileSize=%d",
		r.header.Magic, r.header.Version, r.header.EntityCount, r.header.FileSize)
	logger.Trace("TagDictOffset=%d, TagDictSize=%d", r.header.TagDictOffset, r.header.TagDictSize)
	logger.Trace("EntityIndexOffset=%d, EntityIndexSize=%d", r.header.EntityIndexOffset, r.header.EntityIndexSize)
	
	// Skip dictionary and index if no entities
	if r.header.EntityCount == 0 {
		logger.Trace("No entities in file, skipping dictionary and index")
		return r, nil
	}
	
	// Read tag dictionary
	if r.header.TagDictOffset > 0 && r.header.TagDictSize > 0 {
		logger.Trace("Reading tag dictionary from offset %d", r.header.TagDictOffset)
		if _, err := file.Seek(int64(r.header.TagDictOffset), os.SEEK_SET); err != nil {
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
	if r.header.EntityIndexOffset > 0 {
		logger.Trace("Reading index from offset %d, expecting %d entries", 
			r.header.EntityIndexOffset, r.header.EntityCount)
		
		// Validate index location
		if int64(r.header.EntityIndexOffset) > stat.Size() {
			logger.Error("Index offset %d exceeds file size %d", 
				r.header.EntityIndexOffset, stat.Size())
			return r, nil // Return partial reader
		}
		
		if _, err := file.Seek(int64(r.header.EntityIndexOffset), os.SEEK_SET); err != nil {
			logger.Error("Failed to seek to index: %v", err)
			return nil, err
		}
		
		// Calculate how many entries we can actually read
		indexStartPos := int64(r.header.EntityIndexOffset)
		remainingFileSize := stat.Size() - indexStartPos
		entrySize := int64(binary.Size(IndexEntry{}))
		maxPossibleEntries := uint64(remainingFileSize / entrySize)
		
		if maxPossibleEntries < r.header.EntityCount {
			logger.Warn("File can only hold %d index entries but header claims %d",
				maxPossibleEntries, r.header.EntityCount)
		}
		
		entriesRead := uint64(0)
		for i := uint64(0); i < r.header.EntityCount; i++ {
			// Check if we have enough bytes remaining
			currentPos, _ := file.Seek(0, os.SEEK_CUR)
			if currentPos+entrySize > stat.Size() {
				logger.Warn("Not enough data for index entry %d (pos=%d, need %d bytes, file_size=%d)",
					i, currentPos, entrySize, stat.Size())
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
			r.index[id] = entry
			entriesRead++
			logger.Trace("Loaded index entry %d: ID=%s, Offset=%d, Size=%d", i, id, entry.Offset, entry.Size)
		}
		
		logger.Trace("Index loading complete: read %d entries, loaded %d into index (expected %d)",
			entriesRead, len(r.index), r.header.EntityCount)
		
		if entriesRead < r.header.EntityCount {
			logger.Warn("Index is incomplete: missing %d entries", r.header.EntityCount - entriesRead)
		}
	}
	
	return r, nil
}

// GetEntity reads an entity by ID
func (r *Reader) GetEntity(id string) (*models.Entity, error) {
	// Start operation tracking
	op := models.StartOperation(models.OpTypeRead, id, map[string]interface{}{
		"index_size": len(r.index),
	})
	defer func() {
		if op != nil {
			op.Complete()
		}
	}()
	
	logger.Trace("GetEntity called for ID: %s", id)
	
	entry, exists := r.index[id]
	if !exists {
		err := fmt.Errorf("entity %s not found in index", id)
		op.Fail(err)
		logger.Trace("Entity %s not found in index", id)
		return nil, ErrNotFound
	}
	
	logger.Trace("Found entity %s at offset %d, size %d", id, entry.Offset, entry.Size)
	
	// Seek to entity position
	_, err := r.file.Seek(int64(entry.Offset), os.SEEK_SET)
	if err != nil {
		logger.Error("Failed to seek to offset %d: %v", entry.Offset, err)
		return nil, err
	}
	
	// Read entity data
	data := make([]byte, entry.Size)
	n, err := r.file.Read(data)
	if err != nil {
		logger.Error("Failed to read %d bytes: %v", entry.Size, err)
		return nil, err
	}
	if n != int(entry.Size) {
		logger.Error("Incomplete read, expected %d bytes, got %d", entry.Size, n)
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

// GetAllEntities reads all entities
func (r *Reader) GetAllEntities() ([]*models.Entity, error) {
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

// parseEntity parses entity data from bytes
func (r *Reader) parseEntity(data []byte, id string) (*models.Entity, error) {
	buf := bytes.NewReader(data)
	
	// Read entity header
	var header EntityHeader
	if err := binary.Read(buf, binary.LittleEndian, &header); err != nil {
		return nil, err
	}
	
	entity := &models.Entity{
		ID:   id,
		Tags: make([]string, header.TagCount),
		Content: []byte{}, // New model uses byte slice
	}
	
	// Read tag IDs and convert to strings
	for i := uint16(0); i < header.TagCount; i++ {
		var tagID uint32
		if err := binary.Read(buf, binary.LittleEndian, &tagID); err != nil {
			return nil, err
		}
		entity.Tags[i] = r.tagDict.GetTag(tagID)
	}
	
	// Read content (new unified format)
	if header.ContentCount > 0 {
		// Content type
		var typeLen uint16
		if err := binary.Read(buf, binary.LittleEndian, &typeLen); err != nil {
			return nil, err
		}
		typeBytes := make([]byte, typeLen)
		if _, err := buf.Read(typeBytes); err != nil {
			return nil, err
		}
		contentType := string(typeBytes)
		
		// Content data
		var contentLen uint32
		if err := binary.Read(buf, binary.LittleEndian, &contentLen); err != nil {
			return nil, err
		}
		contentBytes := make([]byte, contentLen)
		if _, err := buf.Read(contentBytes); err != nil {
			return nil, err
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
		
		// Verify checksum if present
		actualChecksum := sha256.Sum256(contentBytes)
		actualChecksumHex := hex.EncodeToString(actualChecksum[:])
		
		checksumValid := false
		for _, tag := range entity.Tags {
			if strings.Contains(tag, "|checksum:sha256:") {
				// Extract checksum from tag
				parts := strings.Split(tag, "|checksum:sha256:")
				if len(parts) == 2 {
					expectedChecksum := parts[1]
					if expectedChecksum == actualChecksumHex {
						checksumValid = true
						logger.Trace("Checksum verification passed for entity %s", id)
					} else {
						logger.Error("Checksum mismatch for entity %s: expected %s, got %s", 
							id, expectedChecksum, actualChecksumHex)
					}
					break
				}
			}
		}
		
		if !checksumValid {
			logger.Warn("No checksum found for entity %s, content integrity not verified", id)
		}
	}
	
	return entity, nil
}

// Close closes the reader
func (r *Reader) Close() error {
	return r.file.Close()
}