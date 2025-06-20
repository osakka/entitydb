// Package binary implements backward compatibility for legacy EBF format files.
// TEMPORARY: This file provides migration support during transition to unified format.

package binary

import (
	"encoding/binary"
	"entitydb/models"
	"fmt"
	"io"
	"os"
	"strings"
	"entitydb/logger"
)

// LegacyReader provides read access to legacy EBF format files
// TEMPORARY: For backward compatibility during migration
type LegacyReader struct {
	file     *os.File
	header   *LegacyHeader
	tagDict  *TagDictionary
	index    map[string]*LegacyIndexEntry
	filename string
}

// LegacyIndexEntry represents index entries in legacy format
type LegacyIndexEntry struct {
	EntityID [64]byte  // Legacy used 64-byte entity IDs
	Offset   uint64    // File offset to entity data
	Size     uint32    // Size of entity data block
	Flags    uint32    // Reserved flags
}

// NewLegacyReader creates a new reader for legacy EBF format files
// TEMPORARY: For backward compatibility during migration
func NewLegacyReader(filename string) (*LegacyReader, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	r := &LegacyReader{
		file:     file,
		tagDict:  NewTagDictionary(),
		index:    make(map[string]*LegacyIndexEntry),
		filename: filename,
	}

	// Read legacy header
	r.header = &LegacyHeader{}
	if err := r.readLegacyHeader(file); err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to read legacy header: %w", err)
	}

	logger.Info("Legacy reader: Magic=%x, Version=%d, EntityCount=%d, FileSize=%d",
		r.header.Magic, r.header.Version, r.header.EntityCount, r.header.FileSize)

	// Load tag dictionary
	if r.header.TagDictSize > 0 {
		if _, err := file.Seek(int64(r.header.TagDictOffset), 0); err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to seek to tag dictionary: %w", err)
		}
		if err := r.tagDict.Read(file); err != nil {
			file.Close()
			return nil, fmt.Errorf("failed to read tag dictionary: %w", err)
		}
	}

	// Load entity index
	if err := r.loadLegacyIndex(); err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to load legacy index: %w", err)
	}

	logger.Info("Legacy reader initialized with %d entities", r.header.EntityCount)
	return r, nil
}

// readLegacyHeader reads the legacy format header (64 bytes)
func (r *LegacyReader) readLegacyHeader(file *os.File) error {
	buf := make([]byte, 64)
	if _, err := io.ReadFull(file, buf); err != nil {
		return err
	}

	r.header.Magic = binary.LittleEndian.Uint32(buf[0:4])
	r.header.Version = binary.LittleEndian.Uint32(buf[4:8])
	r.header.FileSize = binary.LittleEndian.Uint64(buf[8:16])
	r.header.TagDictOffset = binary.LittleEndian.Uint64(buf[16:24])
	r.header.TagDictSize = binary.LittleEndian.Uint64(buf[24:32])
	r.header.EntityIndexOffset = binary.LittleEndian.Uint64(buf[32:40])
	r.header.EntityIndexSize = binary.LittleEndian.Uint64(buf[40:48])
	r.header.EntityCount = binary.LittleEndian.Uint64(buf[48:56])
	r.header.LastModified = int64(binary.LittleEndian.Uint64(buf[56:64]))

	if r.header.Magic != LegacyMagicNumber {
		return ErrInvalidFormat
	}

	return nil
}

// loadLegacyIndex loads the entity index from legacy format
func (r *LegacyReader) loadLegacyIndex() error {
	if r.header.EntityIndexSize == 0 {
		return nil
	}

	if _, err := r.file.Seek(int64(r.header.EntityIndexOffset), 0); err != nil {
		return err
	}

	// Legacy index uses 80-byte entries (64 + 8 + 4 + 4)
	entrySize := 80
	expectedSize := int(r.header.EntityCount) * entrySize

	if int(r.header.EntityIndexSize) < expectedSize {
		logger.Warn("Legacy index size mismatch: expected %d, got %d", expectedSize, r.header.EntityIndexSize)
	}

	for i := uint64(0); i < r.header.EntityCount; i++ {
		entry := &LegacyIndexEntry{}
		
		// Read EntityID (64 bytes)
		if _, err := io.ReadFull(r.file, entry.EntityID[:]); err != nil {
			return fmt.Errorf("failed to read entity ID at index %d: %w", i, err)
		}
		
		// Read Offset (8 bytes)
		if err := binary.Read(r.file, binary.LittleEndian, &entry.Offset); err != nil {
			return fmt.Errorf("failed to read offset at index %d: %w", i, err)
		}
		
		// Read Size (4 bytes)
		if err := binary.Read(r.file, binary.LittleEndian, &entry.Size); err != nil {
			return fmt.Errorf("failed to read size at index %d: %w", i, err)
		}
		
		// Read Flags (4 bytes)
		if err := binary.Read(r.file, binary.LittleEndian, &entry.Flags); err != nil {
			return fmt.Errorf("failed to read flags at index %d: %w", i, err)
		}

		// Convert to string and store
		entityID := strings.TrimRight(string(entry.EntityID[:]), "\x00")
		if entityID != "" {
			r.index[entityID] = entry
		}
	}

	logger.Info("Loaded %d legacy index entries", len(r.index))
	return nil
}

// GetByID retrieves an entity by ID from legacy format
func (r *LegacyReader) GetByID(id string) (*models.Entity, error) {
	entry, exists := r.index[id]
	if !exists {
		return nil, fmt.Errorf("entity not found: %s", id)
	}

	// Seek to entity data
	if _, err := r.file.Seek(int64(entry.Offset), 0); err != nil {
		return nil, fmt.Errorf("failed to seek to entity data: %w", err)
	}

	// Read entity data
	entityData := make([]byte, entry.Size)
	if _, err := io.ReadFull(r.file, entityData); err != nil {
		return nil, fmt.Errorf("failed to read entity data: %w", err)
	}

	// Parse legacy entity format
	entity, err := r.parseLegacyEntity(id, entityData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse legacy entity: %w", err)
	}

	return entity, nil
}

// parseLegacyEntity parses entity data from legacy format
func (r *LegacyReader) parseLegacyEntity(id string, data []byte) (*models.Entity, error) {
	if len(data) < 16 {
		return nil, fmt.Errorf("entity data too short")
	}

	// Legacy format: Modified(8) + TagCount(2) + ContentSize(4) + Reserved(2)
	offset := 0
	modified := int64(binary.LittleEndian.Uint64(data[offset : offset+8]))
	offset += 8
	tagCount := binary.LittleEndian.Uint16(data[offset : offset+2])
	offset += 2
	contentSize := binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4
	// Skip reserved 2 bytes
	offset += 2

	entity := &models.Entity{
		ID:        id,
		CreatedAt: modified,
		Tags:      make([]string, 0, tagCount),
	}

	// Read tags
	for i := uint16(0); i < tagCount; i++ {
		if offset+4 > len(data) {
			return nil, fmt.Errorf("truncated tag data")
		}
		
		tagID := binary.LittleEndian.Uint32(data[offset : offset+4])
		offset += 4
		
		tag := r.tagDict.GetTag(tagID)
		if tag != "" {
			entity.Tags = append(entity.Tags, tag)
		}
	}

	// Read content if present
	if contentSize > 0 {
		if offset+int(contentSize) > len(data) {
			return nil, fmt.Errorf("truncated content data")
		}
		entity.Content = make([]byte, contentSize)
		copy(entity.Content, data[offset:offset+int(contentSize)])
	}

	return entity, nil
}

// GetAll retrieves all entities from legacy format
func (r *LegacyReader) GetAll() ([]*models.Entity, error) {
	entities := make([]*models.Entity, 0, len(r.index))
	
	for id := range r.index {
		entity, err := r.GetByID(id)
		if err != nil {
			logger.Warn("Failed to read legacy entity %s: %v", id, err)
			continue
		}
		entities = append(entities, entity)
	}
	
	return entities, nil
}

// Close closes the legacy reader
func (r *LegacyReader) Close() error {
	if r.file != nil {
		return r.file.Close()
	}
	return nil
}

// ListByTag finds entities with matching tags (simplified for migration)
func (r *LegacyReader) ListByTag(tag string) ([]*models.Entity, error) {
	var matching []*models.Entity
	
	for id := range r.index {
		entity, err := r.GetByID(id)
		if err != nil {
			continue
		}
		
		// Simple tag matching
		for _, entityTag := range entity.Tags {
			if strings.Contains(entityTag, tag) {
				matching = append(matching, entity)
				break
			}
		}
	}
	
	return matching, nil
}

// NewReaderFromLegacy creates a unified Reader that wraps a legacy reader
// TEMPORARY: For backward compatibility during migration
func NewReaderFromLegacy(legacyReader *LegacyReader) (*Reader, error) {
	// Create a unified reader structure but populate it from legacy data
	r := &Reader{
		file:     legacyReader.file,
		tagDict:  legacyReader.tagDict,
		index:    make(map[string]*IndexEntry),
		filename: legacyReader.filename,
	}
	
	// Create a compatible header
	r.header = &Header{
		Magic:             MagicNumber, // Use unified magic for compatibility
		Version:           FormatVersion,
		EntityCount:       legacyReader.header.EntityCount,
		FileSize:          legacyReader.header.FileSize,
		LastModified:      legacyReader.header.LastModified,
		// Set minimal unified format fields
		DataOffset:        64, // After legacy header
		DataSize:          legacyReader.header.TagDictOffset - 64,
		TagDictOffset:     legacyReader.header.TagDictOffset,
		TagDictSize:       legacyReader.header.TagDictSize,
		EntityIndexOffset: legacyReader.header.EntityIndexOffset,
		EntityIndexSize:   legacyReader.header.EntityIndexSize,
	}
	
	// Convert legacy index entries to unified format
	for id, legacyEntry := range legacyReader.index {
		entry := &IndexEntry{
			Offset: legacyEntry.Offset,
			Size:   legacyEntry.Size,
			Flags:  legacyEntry.Flags,
		}
		// Copy entity ID (pad to 96 bytes for unified format)
		copy(entry.EntityID[:], []byte(id))
		r.index[id] = entry
	}
	
	// Store reference to legacy reader for delegation
	r.legacyReader = legacyReader
	
	logger.Info("Created unified reader wrapper for legacy format with %d entities", r.header.EntityCount)
	return r, nil
}