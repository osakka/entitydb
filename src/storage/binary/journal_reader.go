package binary

import (
	"encoding/binary"
	"entitydb/models"
	"fmt"
	"io"
	"os"
	"entitydb/logger"
)

// JournalReader implements a reader for the journal format
type JournalReader struct {
	file     *os.File
	entities map[string]int64 // Entity ID -> last offset
}

// NewJournalReader creates a new journal reader
func NewJournalReader(filename string) (*JournalReader, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	
	r := &JournalReader{
		file:     file,
		entities: make(map[string]int64),
	}
	
	// Scan journal to build index
	if err := r.scanJournal(); err != nil {
		file.Close()
		return nil, err
	}
	
	return r, nil
}

// GetEntity retrieves an entity by ID
func (r *JournalReader) GetEntity(id string) (*models.Entity, error) {
	offset, exists := r.entities[id]
	if !exists {
		return nil, ErrNotFound
	}
	
	// Seek to entry
	if _, err := r.file.Seek(offset, os.SEEK_SET); err != nil {
		return nil, err
	}
	
	// Read entry
	var entryType byte
	if err := binary.Read(r.file, binary.LittleEndian, &entryType); err != nil {
		return nil, err
	}
	
	// Read timestamp
	var timestamp int64
	if err := binary.Read(r.file, binary.LittleEndian, &timestamp); err != nil {
		return nil, err
	}
	
	// Read entity ID
	var idLen uint16
	if err := binary.Read(r.file, binary.LittleEndian, &idLen); err != nil {
		return nil, err
	}
	idBytes := make([]byte, idLen)
	if _, err := r.file.Read(idBytes); err != nil {
		return nil, err
	}
	
	// Read data
	var dataLen uint32
	if err := binary.Read(r.file, binary.LittleEndian, &dataLen); err != nil {
		return nil, err
	}
	data := make([]byte, dataLen)
	if _, err := io.ReadFull(r.file, data); err != nil {
		return nil, err
	}
	
	// Deserialize entity
	return DeserializeEntity(data, id)
}

// GetAllEntities retrieves all entities
func (r *JournalReader) GetAllEntities() ([]*models.Entity, error) {
	entities := make([]*models.Entity, 0, len(r.entities))
	
	for id := range r.entities {
		entity, err := r.GetEntity(id)
		if err != nil {
			logger.Debug("Warning: failed to read entity %s: %v", id, err)
			continue
		}
		entities = append(entities, entity)
	}
	
	return entities, nil
}

// Close closes the reader
func (r *JournalReader) Close() error {
	return r.file.Close()
}

// scanJournal scans the journal to build the index
func (r *JournalReader) scanJournal() error {
	// Seek to beginning
	if _, err := r.file.Seek(0, os.SEEK_SET); err != nil {
		return err
	}
	
	for {
		offset, err := r.file.Seek(0, os.SEEK_CUR)
		if err != nil {
			return err
		}
		
		// Read entry type
		var entryType byte
		if err := binary.Read(r.file, binary.LittleEndian, &entryType); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		
		// Read timestamp
		var timestamp int64
		if err := binary.Read(r.file, binary.LittleEndian, &timestamp); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		
		// Read entity ID
		var idLen uint16
		if err := binary.Read(r.file, binary.LittleEndian, &idLen); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		
		if idLen > 1024 { // Sanity check
			return fmt.Errorf("invalid ID length: %d", idLen)
		}
		
		idBytes := make([]byte, idLen)
		if _, err := io.ReadFull(r.file, idBytes); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		entityID := string(idBytes)
		
		// Read data length
		var dataLen uint32
		if err := binary.Read(r.file, binary.LittleEndian, &dataLen); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		
		if dataLen > 10*1024*1024 { // Sanity check: 10MB max
			return fmt.Errorf("invalid data length: %d", dataLen)
		}
		
		// Skip data for now
		if _, err := r.file.Seek(int64(dataLen), os.SEEK_CUR); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		
		// Update index
		if entryType == EntryTypeEntity {
			r.entities[entityID] = offset
			logger.Debug("Found entity %s at offset %d", entityID, offset)
		} else if entryType == EntryTypeDelete {
			delete(r.entities, entityID)
			logger.Debug("Marked entity %s as deleted", entityID)
		}
	}
	
	logger.Debug("Scanned journal, found %d entities", len(r.entities))
	return nil
}