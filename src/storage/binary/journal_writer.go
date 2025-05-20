package binary

import (
	"encoding/binary"
	"entitydb/models"
	"fmt"
	"os"
	"sync"
	"time"
	"entitydb/logger"
)

// JournalEntry represents a single entry in the journal
type JournalEntry struct {
	Type      byte   // 1=Entity, 2=Delete, 3=Checkpoint
	Timestamp int64  // Unix timestamp
	EntityID  string // Entity ID
	Data      []byte // Entity data (for Type=1)
}

const (
	EntryTypeEntity     byte = 1
	EntryTypeDelete     byte = 2
	EntryTypeCheckpoint byte = 3
)

// JournalWriter implements a simple journal-based writer
type JournalWriter struct {
	file     *os.File
	mu       sync.Mutex
	entities map[string]int64 // Entity ID -> last offset
}

// NewJournalWriter creates a new journal writer
func NewJournalWriter(filename string) (*JournalWriter, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	
	w := &JournalWriter{
		file:     file,
		entities: make(map[string]int64),
	}
	
	// Scan existing entries to build index
	if err := w.scanJournal(); err != nil {
		logger.Debug("Warning: failed to scan journal: %v", err)
	}
	
	return w, nil
}

// WriteEntity writes an entity to the journal
func (w *JournalWriter) WriteEntity(entity *models.Entity) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	logger.Debug("JournalWriter.WriteEntity called for %s", entity.ID)
	
	// Serialize entity
	data, err := SerializeEntity(entity)
	if err != nil {
		return fmt.Errorf("failed to serialize entity: %w", err)
	}
	
	// Create journal entry
	entry := JournalEntry{
		Type:      EntryTypeEntity,
		Timestamp: time.Now().Unix(),
		EntityID:  entity.ID,
		Data:      data,
	}
	
	// Get current position
	offset, err := w.file.Seek(0, os.SEEK_END)
	if err != nil {
		return fmt.Errorf("failed to seek to end: %w", err)
	}
	
	// Write entry header
	if err := binary.Write(w.file, binary.LittleEndian, entry.Type); err != nil {
		return err
	}
	if err := binary.Write(w.file, binary.LittleEndian, entry.Timestamp); err != nil {
		return err
	}
	
	// Write entity ID length and string
	idBytes := []byte(entry.EntityID)
	if err := binary.Write(w.file, binary.LittleEndian, uint16(len(idBytes))); err != nil {
		return err
	}
	if _, err := w.file.Write(idBytes); err != nil {
		return err
	}
	
	// Write data length and data
	if err := binary.Write(w.file, binary.LittleEndian, uint32(len(entry.Data))); err != nil {
		return err
	}
	if _, err := w.file.Write(entry.Data); err != nil {
		return err
	}
	
	// Sync immediately
	if err := w.file.Sync(); err != nil {
		return fmt.Errorf("failed to sync: %w", err)
	}
	
	// Update index
	w.entities[entity.ID] = offset
	
	logger.Debug("Entity %s written at offset %d", entity.ID, offset)
	return nil
}

// Close closes the writer
func (w *JournalWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	return w.file.Close()
}

// scanJournal scans the journal to build the index
func (w *JournalWriter) scanJournal() error {
	// Seek to beginning
	if _, err := w.file.Seek(0, os.SEEK_SET); err != nil {
		return err
	}
	
	for {
		offset, err := w.file.Seek(0, os.SEEK_CUR)
		if err != nil {
			return err
		}
		
		// Read entry type
		var entryType byte
		if err := binary.Read(w.file, binary.LittleEndian, &entryType); err != nil {
			if err.Error() == "EOF" {
				break
			}
			return err
		}
		
		// Read timestamp
		var timestamp int64
		if err := binary.Read(w.file, binary.LittleEndian, &timestamp); err != nil {
			return err
		}
		
		// Read entity ID
		var idLen uint16
		if err := binary.Read(w.file, binary.LittleEndian, &idLen); err != nil {
			return err
		}
		idBytes := make([]byte, idLen)
		if _, err := w.file.Read(idBytes); err != nil {
			return err
		}
		entityID := string(idBytes)
		
		// Read data length
		var dataLen uint32
		if err := binary.Read(w.file, binary.LittleEndian, &dataLen); err != nil {
			return err
		}
		
		// Skip data for now
		if _, err := w.file.Seek(int64(dataLen), os.SEEK_CUR); err != nil {
			return err
		}
		
		// Update index
		if entryType == EntryTypeEntity {
			w.entities[entityID] = offset
		} else if entryType == EntryTypeDelete {
			delete(w.entities, entityID)
		}
	}
	
	logger.Debug("Scanned journal, found %d entities", len(w.entities))
	return nil
}