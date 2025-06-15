package binary

import (
	"bytes"
	"encoding/binary"
	"entitydb/models"
	"time"
)

// SerializeEntity serializes an entity to EntityDB Binary Format (EBF).
//
// Binary Format Structure:
//   - Tag count (uint16)
//   - For each tag: length (uint16) + tag bytes
//   - Content count (uint16, 0 or 1)
//   - Content data (if present)
//   - Created timestamp (int64)
//   - Updated timestamp (int64)
//
// Parameters:
//   - entity: Entity instance to serialize (must not be nil)
//
// Returns:
//   - []byte: Serialized binary data in EBF format
//   - error: Encoding errors (buffer write failures, timestamp issues)
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

