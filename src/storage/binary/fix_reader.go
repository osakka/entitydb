package binary

import (
	"encoding/binary"
	"encoding/json"
	"entitydb/models"
	"fmt"
	"io"
	"os"
)

// FixedReader is an improved reader with better error handling
type FixedReader struct {
	file     *os.File
	header   *Header
	tagDict  *TagDictionary
	index    map[string]*IndexEntry
	entities []EntityData
}

// EntityData holds raw entity data
type EntityData struct {
	ID     string
	Offset uint64
	Size   uint32
}

// NewFixedReader creates a new improved reader
func NewFixedReader(filename string) (*FixedReader, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	
	// Get file size
	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}
	
	if stat.Size() < int64(HeaderSize) {
		file.Close()
		return nil, fmt.Errorf("file too small to be valid binary format")
	}
	
	r := &FixedReader{
		file:     file,
		header:   &Header{},
		tagDict:  NewTagDictionary(),
		index:    make(map[string]*IndexEntry),
		entities: []EntityData{},
	}
	
	// Read header
	if err := r.header.Read(file); err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to read header: %w", err)
	}
	
	// Skip reading tag dictionary and index for now
	// Just try to find entities in the file
	return r, nil
}

// ScanForEntities scans the binary file for entity data
func (r *FixedReader) ScanForEntities() ([]*models.Entity, error) {
	var entities []*models.Entity
	
	// Start after header
	offset := int64(HeaderSize)
	r.file.Seek(offset, os.SEEK_SET)
	
	for {
		// Try to read entity header
		var header EntityHeader
		if err := binary.Read(r.file, binary.LittleEndian, &header); err != nil {
			if err == io.EOF {
				break
			}
			// Skip this position and try next
			offset++
			r.file.Seek(offset, os.SEEK_SET)
			continue
		}
		
		// Sanity check
		if header.TagCount > 1000 || header.ContentCount > 1000 {
			// Invalid header, skip
			offset++
			r.file.Seek(offset, os.SEEK_SET)
			continue
		}
		
		// Try to read entity data
		entity := &models.Entity{
			ID:      fmt.Sprintf("entity_%d", offset),
			Tags:    make([]string, 0, header.TagCount),
			Content: []byte{}, // New model uses byte slice for content
		}
		
		// Read tags
		for i := uint16(0); i < header.TagCount; i++ {
			var tagID uint32
			if err := binary.Read(r.file, binary.LittleEndian, &tagID); err != nil {
				goto nextEntity
			}
			// For now, just use tag ID as string
			entity.Tags = append(entity.Tags, fmt.Sprintf("tag:%d", tagID))
		}
		
		// Read content
		for i := uint16(0); i < header.ContentCount; i++ {
			// Type length
			var typeLen uint16
			if err := binary.Read(r.file, binary.LittleEndian, &typeLen); err != nil {
				goto nextEntity
			}
			
			if typeLen > 1000 {
				goto nextEntity
			}
			
			typeBytes := make([]byte, typeLen)
			if _, err := r.file.Read(typeBytes); err != nil {
				goto nextEntity
			}
			
			// Value length
			var valueLen uint32
			if err := binary.Read(r.file, binary.LittleEndian, &valueLen); err != nil {
				goto nextEntity
			}
			
			if valueLen > 100000 {
				goto nextEntity
			}
			
			valueBytes := make([]byte, valueLen)
			if _, err := r.file.Read(valueBytes); err != nil {
				goto nextEntity
			}
			
			// Timestamp
			var tsNano int64
			if err := binary.Read(r.file, binary.LittleEndian, &tsNano); err != nil {
				goto nextEntity
			}
			
			// In the new model, we'll store the content as JSON
			contentData := map[string]string{
				string(typeBytes): string(valueBytes),
			}
			jsonData, _ := json.Marshal(contentData)
			entity.Content = jsonData
			entity.AddTag("content:type:" + string(typeBytes))
		}
		
		// Found valid entity
		entities = append(entities, entity)
		
	nextEntity:
		// Move to next position
		current, _ := r.file.Seek(0, os.SEEK_CUR)
		offset = current
	}
	
	return entities, nil
}

// Close closes the reader
func (r *FixedReader) Close() error {
	return r.file.Close()
}