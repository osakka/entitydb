package binary

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"entitydb/models"
	"entitydb/logger"
	"errors"
	"io"
	"os"
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
	logger.Debug("Opening reader for file: %s", filename)
	
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
	logger.Debug("File size: %d bytes", stat.Size())
	
	r := &Reader{
		file:    file,
		header:  &Header{},
		tagDict: NewTagDictionary(),
		index:   make(map[string]*IndexEntry),
	}
	
	// Read header
	logger.Debug("Reading header")
	if err := r.header.Read(file); err != nil {
		logger.Error("Failed to read header: %v", err)
		return nil, err
	}
	
	logger.Debug("Header read successfully: Magic=%x, Version=%d, EntityCount=%d, FileSize=%d",
		r.header.Magic, r.header.Version, r.header.EntityCount, r.header.FileSize)
	logger.Debug("TagDictOffset=%d, TagDictSize=%d", r.header.TagDictOffset, r.header.TagDictSize)
	logger.Debug("EntityIndexOffset=%d, EntityIndexSize=%d", r.header.EntityIndexOffset, r.header.EntityIndexSize)
	
	// Skip dictionary and index if no entities
	if r.header.EntityCount == 0 {
		logger.Debug("No entities in file, skipping dictionary and index")
		return r, nil
	}
	
	// Read tag dictionary
	if r.header.TagDictOffset > 0 && r.header.TagDictSize > 0 {
		logger.Debug("Reading tag dictionary from offset %d", r.header.TagDictOffset)
		if _, err := file.Seek(int64(r.header.TagDictOffset), os.SEEK_SET); err != nil {
			logger.Error("Failed to seek to tag dictionary: %v", err)
			return nil, err
		}
		if err := r.tagDict.Read(file); err != nil {
			// Log but don't fail - allow partial reads
			logger.Warn("Failed to read tag dictionary: %v", err)
		} else {
			logger.Debug("Tag dictionary loaded with %d entries", 0)
		}
	}
	
	// Read index
	if r.header.EntityIndexOffset > 0 {
		logger.Debug("Reading index from offset %d, expecting %d entries", 
			r.header.EntityIndexOffset, r.header.EntityCount)
		
		if _, err := file.Seek(int64(r.header.EntityIndexOffset), os.SEEK_SET); err != nil {
			logger.Error("Failed to seek to index: %v", err)
			return nil, err
		}
		
		for i := uint64(0); i < r.header.EntityCount; i++ {
			entry := &IndexEntry{}
			if err := binary.Read(file, binary.LittleEndian, &entry.EntityID); err != nil {
				// Stop reading if we hit EOF
				if err == io.EOF {
					logger.Warn("Hit EOF at index entry %d", i)
					break
				}
				logger.Error("Failed to read index entry %d: %v", i, err)
				return nil, err
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
				logger.Debug("Skipping empty index entry %d", i)
				continue
			}
			r.index[id] = entry
			logger.Debug("Loaded index entry %d: ID=%s, Offset=%d, Size=%d", i, id, entry.Offset, entry.Size)
		}
		
		logger.Debug("Loaded %d index entries", len(r.index))
	}
	
	return r, nil
}

// GetEntity reads an entity by ID
func (r *Reader) GetEntity(id string) (*models.Entity, error) {
	logger.Debug("GetEntity called for ID: %s", id)
	
	entry, exists := r.index[id]
	if !exists {
		logger.Debug("Entity %s not found in index", id)
		return nil, ErrNotFound
	}
	
	logger.Debug("Found entity %s at offset %d, size %d", id, entry.Offset, entry.Size)
	
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
	
	logger.Debug("Read %d bytes for entity %s", n, id)
	
	entity, err := r.parseEntity(data, id)
	if err != nil {
		logger.Error("Failed to parse entity %s: %v", id, err)
		return nil, err
	}
	
	logger.Debug("Successfully parsed entity %s", id)
	return entity, nil
}

// GetAllEntities reads all entities
func (r *Reader) GetAllEntities() ([]*models.Entity, error) {
	logger.Debug("GetAllEntities called, index has %d entries, header says %d entities", len(r.index), r.header.EntityCount)
	entities := make([]*models.Entity, 0, r.header.EntityCount)
	
	for id := range r.index {
		logger.Debug("Getting entity with ID: %s", id)
		entity, err := r.GetEntity(id)
		if err != nil {
			logger.Debug("Error getting entity %s: %v", id, err)
			// Skip entities we can't read
			continue
		}
		entities = append(entities, entity)
	}
	
	logger.Debug("GetAllEntities returning %d entities", len(entities))
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
	
	// Read content (old format - convert to new format)
	contentData := make(map[string]string)
	for i := uint16(0); i < header.ContentCount; i++ {
		// Type
		var typeLen uint16
		if err := binary.Read(buf, binary.LittleEndian, &typeLen); err != nil {
			return nil, err
		}
		typeBytes := make([]byte, typeLen)
		if _, err := buf.Read(typeBytes); err != nil {
			return nil, err
		}
		
		// Value
		var valueLen uint32
		if err := binary.Read(buf, binary.LittleEndian, &valueLen); err != nil {
			return nil, err
		}
		valueBytes := make([]byte, valueLen)
		if _, err := buf.Read(valueBytes); err != nil {
			return nil, err
		}
		
		// Timestamp
		var tsNano int64
		if err := binary.Read(buf, binary.LittleEndian, &tsNano); err != nil {
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

// Close closes the reader
func (r *Reader) Close() error {
	return r.file.Close()
}