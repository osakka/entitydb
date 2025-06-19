package binary

import (
	"entitydb/models"
	"entitydb/logger"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"syscall"
	"unsafe"
	"encoding/binary"
	"bytes"
)

// MMapReader provides zero-copy reads using memory-mapped files
type MMapReader struct {
	file     *os.File
	data     []byte
	size     int64
	header   *Header
	index    map[string]*IndexEntry
	indexMu  sync.RWMutex
}

// NewMMapReader creates a new memory-mapped reader
func NewMMapReader(filename string) (*MMapReader, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	
	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("error getting file stats: %w", err)
	}
	
	size := stat.Size()
	if size == 0 {
		file.Close()
		return nil, fmt.Errorf("file is empty")
	}
	
	// Memory-map the file
	data, err := syscall.Mmap(int(file.Fd()), 0, int(size), syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("error mapping file: %w", err)
	}
	
	reader := &MMapReader{
		file: file,
		data: data,
		size: size,
		index: make(map[string]*IndexEntry),
	}
	
	// Read header
	if err := reader.readHeader(); err != nil {
		reader.Close()
		return nil, err
	}
	
	// Build index
	if err := reader.buildIndex(); err != nil {
		reader.Close()
		return nil, err
	}
	
	return reader, nil
}

// readHeader reads the file header from memory-mapped data
func (r *MMapReader) readHeader() error {
	if len(r.data) < 64 {
		return fmt.Errorf("file too small for header")
	}
	
	// Direct cast to header struct (zero-copy)
	r.header = (*Header)(unsafe.Pointer(&r.data[0]))
	
	// Validate magic number
	if r.header.Magic != MagicNumber {
		return fmt.Errorf("invalid magic number")
	}
	
	return nil
}

// buildIndex builds the entity index from memory-mapped data
func (r *MMapReader) buildIndex() error {
	if r.header.EntityIndexOffset == 0 || r.header.EntityCount == 0 {
		return nil // No entities
	}
	
	offset := r.header.EntityIndexOffset
	
	for i := uint64(0); i < r.header.EntityCount; i++ {
		if offset+52 > uint64(r.size) {
			return fmt.Errorf("index entry %d exceeds file size", i)
		}
		
		// Direct memory access
		tempEntry := &IndexEntry{
			Offset: *(*uint64)(unsafe.Pointer(&r.data[offset+36])),
			Size:   *(*uint32)(unsafe.Pointer(&r.data[offset+44])),
			Flags:  *(*uint32)(unsafe.Pointer(&r.data[offset+48])),
		}
		
		// Extract entity ID
		entityID := string(r.data[offset:offset+36])
		
		// RACE CONDITION FIX: Create a defensive copy of IndexEntry to prevent concurrent access corruption
		// Unsafe pointer operations can create shared memory access patterns causing corruption
		indexEntry := &IndexEntry{
			Offset: tempEntry.Offset,
			Size:   tempEntry.Size,
			Flags:  tempEntry.Flags,
		}
		// Copy EntityID from the extracted string
		copy(indexEntry.EntityID[:], []byte(entityID))
		r.index[entityID] = indexEntry
		
		offset += 52
	}
	
	return nil
}

// GetEntity retrieves an entity by ID with zero-copy
func (r *MMapReader) GetEntity(id string) (*models.Entity, error) {
	r.indexMu.RLock()
	entry, exists := r.index[id]
	if !exists {
		r.indexMu.RUnlock()
		return nil, fmt.Errorf("entity not found: %s", id)
	}
	
	// RACE CONDITION FIX: Create defensive copy of IndexEntry to prevent shared pointer access
	entryCopy := &IndexEntry{
		Offset: entry.Offset,
		Size:   entry.Size,
		Flags:  entry.Flags,
	}
	copy(entryCopy.EntityID[:], entry.EntityID[:])
	r.indexMu.RUnlock()
	
	// Use entryCopy instead of entry for safe access
	entry = entryCopy
	
	if entry.Offset+uint64(entry.Size) > uint64(r.size) {
		return nil, fmt.Errorf("entity data exceeds file size")
	}
	
	// Direct deserialize from memory-mapped data
	// The data includes an entity header before the actual entity data
	data := r.data[entry.Offset:entry.Offset+uint64(entry.Size)]
	
	// Parse entity header (8 bytes)
	if len(data) < 8 {
		return nil, fmt.Errorf("entity data too small for header")
	}
	
	reader := bytes.NewReader(data)
	
	// Read entity header
	var entityHeader EntityHeader
	if err := binary.Read(reader, binary.LittleEndian, &entityHeader.Modified); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &entityHeader.TagCount); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &entityHeader.ContentCount); err != nil {
		return nil, err
	}
	if err := binary.Read(reader, binary.LittleEndian, &entityHeader.Reserved); err != nil {
		return nil, err
	}
	
	// Now deserialize the rest of the entity data
	entity := &models.Entity{ID: id}
	
	// Read tags
	entity.Tags = make([]string, entityHeader.TagCount)
	for i := uint16(0); i < entityHeader.TagCount; i++ {
		var tagID uint32
		if err := binary.Read(reader, binary.LittleEndian, &tagID); err != nil {
			return nil, err
		}
		// For now, we'll just use the tag ID as the tag string
		// In a real implementation, we'd look up the tag in the dictionary
		entity.Tags[i] = fmt.Sprintf("tag:%d", tagID)
	}
	
	// Read content items (old format - convert to new format)
	contentData := make(map[string]string)
	for i := uint16(0); i < entityHeader.ContentCount; i++ {
		// Type
		var typeLen uint16
		if err := binary.Read(reader, binary.LittleEndian, &typeLen); err != nil {
			return nil, err
		}
		typeBytes := make([]byte, typeLen)
		if _, err := reader.Read(typeBytes); err != nil {
			return nil, err
		}
		
		// Value
		var valueLen uint32
		if err := binary.Read(reader, binary.LittleEndian, &valueLen); err != nil {
			return nil, err
		}
		valueBytes := make([]byte, valueLen)
		if _, err := reader.Read(valueBytes); err != nil {
			return nil, err
		}
		
		// Timestamp
		var tsNano int64
		if err := binary.Read(reader, binary.LittleEndian, &tsNano); err != nil {
			return nil, err
		}
		
		// Add to content data map
		contentData[string(typeBytes)] = string(valueBytes)
		entity.AddTag("content:type:" + string(typeBytes))
	}
	
	// Convert content data to JSON
	if len(contentData) > 0 {
		jsonData, _ := json.Marshal(contentData)
		entity.Content = jsonData
	}
	
	return entity, nil
}

// GetAllEntities reads all entities with zero-copy
func (r *MMapReader) GetAllEntities() ([]*models.Entity, error) {
	r.indexMu.RLock()
	defer r.indexMu.RUnlock()
	
	entities := make([]*models.Entity, 0, len(r.index))
	
	for id, entry := range r.index {
		if entry.Offset+uint64(entry.Size) > uint64(r.size) {
			logger.Error("Entity %s data exceeds file size", id)
			continue
		}
		
		entity, err := r.GetEntity(id)
		if err != nil {
			logger.Error("Failed to deserialize entity %s: %v", id, err)
			continue
		}
		
		entities = append(entities, entity)
	}
	
	return entities, nil
}

// Close unmaps the file and closes it
func (r *MMapReader) Close() error {
	if r.data != nil {
		if err := syscall.Munmap(r.data); err != nil {
			return fmt.Errorf("error unmapping file: %w", err)
		}
	}
	
	if r.file != nil {
		return r.file.Close()
	}
	
	return nil
}