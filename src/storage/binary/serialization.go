package binary

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"entitydb/models"
	"time"
)

// SerializeEntity serializes an entity to binary format
func SerializeEntity(entity *models.Entity) ([]byte, error) {
	buf := new(bytes.Buffer)
	
	// Write tag count
	if err := binary.Write(buf, binary.LittleEndian, uint16(len(entity.Tags))); err != nil {
		return nil, err
	}
	
	// Write tags
	for _, tag := range entity.Tags {
		tagBytes := []byte(tag)
		if err := binary.Write(buf, binary.LittleEndian, uint16(len(tagBytes))); err != nil {
			return nil, err
		}
		if _, err := buf.Write(tagBytes); err != nil {
			return nil, err
		}
	}
	
	// Write content count (1 if we have content, 0 otherwise)
	contentCount := uint16(0)
	if len(entity.Content) > 0 {
		contentCount = 1
	}
	if err := binary.Write(buf, binary.LittleEndian, contentCount); err != nil {
		return nil, err
	}
	
	// Write content as a single binary blob
	// For backward compatibility, we'll write it as one "content" item
	if len(entity.Content) > 0 {
		// Type - use a special "raw" type that tells deserializer not to JSON-wrap
		typeBytes := []byte("raw_content")
		if err := binary.Write(buf, binary.LittleEndian, uint16(len(typeBytes))); err != nil {
			return nil, err
		}
		if _, err := buf.Write(typeBytes); err != nil {
			return nil, err
		}
		
		// Value (the entire content as bytes)
		if err := binary.Write(buf, binary.LittleEndian, uint32(len(entity.Content))); err != nil {
			return nil, err
		}
		if _, err := buf.Write(entity.Content); err != nil {
			return nil, err
		}
		
		// Timestamp (use current time)
		if err := binary.Write(buf, binary.LittleEndian, time.Now().UnixNano()); err != nil {
			return nil, err
		}
	}
	
	return buf.Bytes(), nil
}

// DeserializeEntityLegacy deserializes an entity from binary format (legacy version)
func DeserializeEntityLegacy(data []byte, id string) (*models.Entity, error) {
	buf := bytes.NewReader(data)
	entity := &models.Entity{ID: id}
	
	// Read tag count
	var tagCount uint16
	if err := binary.Read(buf, binary.LittleEndian, &tagCount); err != nil {
		return nil, err
	}
	
	// Read tags
	entity.Tags = make([]string, tagCount)
	for i := uint16(0); i < tagCount; i++ {
		var tagLen uint16
		if err := binary.Read(buf, binary.LittleEndian, &tagLen); err != nil {
			return nil, err
		}
		tagBytes := make([]byte, tagLen)
		if _, err := buf.Read(tagBytes); err != nil {
			return nil, err
		}
		entity.Tags[i] = string(tagBytes)
	}
	
	// Read content count
	var contentCount uint16
	if err := binary.Read(buf, binary.LittleEndian, &contentCount); err != nil {
		return nil, err
	}
	
	// For backward compatibility with old multi-item format,
	// read all items and concatenate as a single JSON array
	if contentCount > 0 {
		var items []interface{}
		for i := uint16(0); i < contentCount; i++ {
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
			
			// Check if this is our special raw_content type
			if string(typeBytes) == "raw_content" {
				// Use the raw bytes directly rather than wrapping in JSON
				entity.Content = valueBytes
				// Skip the items array since we're setting content directly
				return entity, nil
			} else {
				item := map[string]interface{}{
					"type":      string(typeBytes),
					"value":     string(valueBytes),
					"timestamp": time.Unix(0, tsNano).Format(time.RFC3339Nano),
				}
				items = append(items, item)
			}
		}
		
		// Convert to JSON for backward compatibility
		contentJSON, err := json.Marshal(items)
		if err != nil {
			return nil, err
		}
		entity.Content = contentJSON
	} else {
		entity.Content = []byte{}
	}
	
	return entity, nil
}