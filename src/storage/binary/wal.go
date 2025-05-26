package binary

import (
	"encoding/binary"
	"entitydb/models"
	"entitydb/logger"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// WALEntry represents a single entry in the write-ahead log
type WALEntry struct {
	OpType    WALOpType
	EntityID  string
	Entity    *models.Entity
	Timestamp time.Time
	Checksum  uint32
}

// WALOpType defines the type of operation in the WAL
type WALOpType uint8

const (
	WALOpCreate WALOpType = iota
	WALOpUpdate
	WALOpDelete
	WALOpCheckpoint
)

// WAL represents a write-ahead log for crash recovery
type WAL struct {
	mu       sync.Mutex
	file     *os.File
	path     string
	sequence uint64
}

// NewWAL creates a new write-ahead log
func NewWAL(dataPath string) (*WAL, error) {
	walPath := filepath.Join(dataPath, "entitydb.wal")
	logger.Debug("Creating WAL with dataPath: %s, resulting walPath: %s", dataPath, walPath)
	
	file, err := os.OpenFile(walPath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}
	
	wal := &WAL{
		file: file,
		path: walPath,
	}
	
	// Read the last sequence number
	if err := wal.readSequence(); err != nil {
		return nil, err
	}
	
	return wal, nil
}

// LogCreate logs an entity creation
func (w *WAL) LogCreate(entity *models.Entity) error {
	op := models.StartOperation(models.OpTypeWAL, entity.ID, map[string]interface{}{
		"wal_operation": "create",
		"entity_size": len(entity.Content),
		"tag_count": len(entity.Tags),
	})
	defer op.Complete()
	
	logger.Info("[WAL] Logging CREATE operation for entity %s", entity.ID)
	
	err := w.logEntry(WALEntry{
		OpType:    WALOpCreate,
		EntityID:  entity.ID,
		Entity:    entity,
		Timestamp: time.Now(),
	})
	
	if err != nil {
		op.Fail(err)
		logger.Error("[WAL] Failed to log CREATE for entity %s: %v", entity.ID, err)
		return err
	}
	
	op.SetMetadata("sequence", w.sequence)
	logger.Debug("[WAL] Successfully logged CREATE for entity %s at sequence %d", entity.ID, w.sequence)
	return nil
}

// LogUpdate logs an entity update
func (w *WAL) LogUpdate(entity *models.Entity) error {
	op := models.StartOperation(models.OpTypeWAL, entity.ID, map[string]interface{}{
		"wal_operation": "update",
		"entity_size": len(entity.Content),
		"tag_count": len(entity.Tags),
	})
	defer op.Complete()
	
	logger.Info("[WAL] Logging UPDATE operation for entity %s", entity.ID)
	
	err := w.logEntry(WALEntry{
		OpType:    WALOpUpdate,
		EntityID:  entity.ID,
		Entity:    entity,
		Timestamp: time.Now(),
	})
	
	if err != nil {
		op.Fail(err)
		logger.Error("[WAL] Failed to log UPDATE for entity %s: %v", entity.ID, err)
		return err
	}
	
	op.SetMetadata("sequence", w.sequence)
	logger.Debug("[WAL] Successfully logged UPDATE for entity %s at sequence %d", entity.ID, w.sequence)
	return nil
}

// LogDelete logs an entity deletion
func (w *WAL) LogDelete(entityID string) error {
	op := models.StartOperation(models.OpTypeWAL, entityID, map[string]interface{}{
		"wal_operation": "delete",
	})
	defer op.Complete()
	
	logger.Info("[WAL] Logging DELETE operation for entity %s", entityID)
	
	err := w.logEntry(WALEntry{
		OpType:    WALOpDelete,
		EntityID:  entityID,
		Timestamp: time.Now(),
	})
	
	if err != nil {
		op.Fail(err)
		logger.Error("[WAL] Failed to log DELETE for entity %s: %v", entityID, err)
		return err
	}
	
	op.SetMetadata("sequence", w.sequence)
	logger.Debug("[WAL] Successfully logged DELETE for entity %s at sequence %d", entityID, w.sequence)
	return nil
}

// LogCheckpoint logs a checkpoint
func (w *WAL) LogCheckpoint() error {
	return w.logEntry(WALEntry{
		OpType:    WALOpCheckpoint,
		Timestamp: time.Now(),
	})
}

// logEntry writes an entry to the WAL
func (w *WAL) logEntry(entry WALEntry) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	// Serialize the entry
	data, err := w.serializeEntry(entry)
	if err != nil {
		return err
	}
	
	// Write the length prefix
	if err := binary.Write(w.file, binary.LittleEndian, uint32(len(data))); err != nil {
		return err
	}
	
	// Write the data
	if _, err := w.file.Write(data); err != nil {
		return err
	}
	
	// Sync to ensure durability
	if err := w.file.Sync(); err != nil {
		return err
	}
	
	w.sequence++
	
	return nil
}

// serializeEntry serializes a WAL entry
func (w *WAL) serializeEntry(entry WALEntry) ([]byte, error) {
	// This is a simplified serialization
	// In production, you'd use a proper format like protobuf
	
	// For now, we'll use a simple format:
	// [OpType:1][TimestampNano:8][EntityIDLen:2][EntityID:var][EntityData:var]
	
	buf := make([]byte, 0, 1024)
	buf = append(buf, byte(entry.OpType))
	
	// Add timestamp
	ts := entry.Timestamp.UnixNano()
	tsBuf := make([]byte, 8)
	binary.LittleEndian.PutUint64(tsBuf, uint64(ts))
	buf = append(buf, tsBuf...)
	
	// Add entity ID
	idLen := uint16(len(entry.EntityID))
	idLenBuf := make([]byte, 2)
	binary.LittleEndian.PutUint16(idLenBuf, idLen)
	buf = append(buf, idLenBuf...)
	buf = append(buf, []byte(entry.EntityID)...)
	
	// Add entity data if present
	if entry.Entity != nil {
		// Simplified: just store the entity ID, tags, and content
		// In production, you'd use a proper serialization format
		entityBuf := fmt.Sprintf("%v", entry.Entity)
		entityLen := uint32(len(entityBuf))
		entityLenBuf := make([]byte, 4)
		binary.LittleEndian.PutUint32(entityLenBuf, entityLen)
		buf = append(buf, entityLenBuf...)
		buf = append(buf, []byte(entityBuf)...)
	}
	
	return buf, nil
}

// Replay replays the WAL entries
func (w *WAL) Replay(callback func(entry WALEntry) error) error {
	op := models.StartOperation(models.OpTypeWAL, "replay", map[string]interface{}{
		"wal_operation": "replay",
		"wal_path": w.path,
	})
	defer op.Complete()
	
	logger.Info("[WAL] Starting WAL replay from %s", w.path)
	
	// Seek to the beginning
	if _, err := w.file.Seek(0, io.SeekStart); err != nil {
		op.Fail(err)
		logger.Error("[WAL] Failed to seek to beginning: %v", err)
		return err
	}
	
	entriesProcessed := 0
	entriesFailed := 0
	
	for {
		// Read length prefix
		var length uint32
		if err := binary.Read(w.file, binary.LittleEndian, &length); err != nil {
			if err == io.EOF {
				break
			}
			op.Fail(err)
			logger.Error("[WAL] Failed to read entry length: %v", err)
			return err
		}
		
		// Read data
		data := make([]byte, length)
		if _, err := io.ReadFull(w.file, data); err != nil {
			op.Fail(err)
			logger.Error("[WAL] Failed to read entry data (length=%d): %v", length, err)
			return err
		}
		
		// Deserialize entry
		entry, err := w.deserializeEntry(data)
		if err != nil {
			entriesFailed++
			logger.Error("[WAL] Failed to deserialize entry: %v", err)
			continue // Skip bad entries during replay
		}
		
		logger.Debug("[WAL] Replaying %v operation for entity %s", entry.OpType, entry.EntityID)
		
		// Process entry
		if err := callback(*entry); err != nil {
			entriesFailed++
			logger.Error("[WAL] Failed to process entry for entity %s: %v", entry.EntityID, err)
			continue // Continue with other entries
		}
		
		entriesProcessed++
	}
	
	op.SetMetadata("entries_processed", entriesProcessed)
	op.SetMetadata("entries_failed", entriesFailed)
	
	logger.Info("[WAL] WAL replay completed: %d entries processed, %d failed", entriesProcessed, entriesFailed)
	
	return nil
}

// deserializeEntry deserializes a WAL entry
func (w *WAL) deserializeEntry(data []byte) (*WALEntry, error) {
	if len(data) < 11 { // Minimum size: OpType(1) + Timestamp(8) + IDLen(2)
		return nil, fmt.Errorf("invalid WAL entry: too short")
	}
	
	entry := &WALEntry{}
	
	// Read op type
	entry.OpType = WALOpType(data[0])
	
	// Read timestamp
	ts := binary.LittleEndian.Uint64(data[1:9])
	entry.Timestamp = time.Unix(0, int64(ts))
	
	// Read entity ID
	idLen := binary.LittleEndian.Uint16(data[9:11])
	if len(data) < 11+int(idLen) {
		return nil, fmt.Errorf("invalid WAL entry: ID length mismatch")
	}
	
	entry.EntityID = string(data[11 : 11+idLen])
	
	// Read entity data if present
	pos := 11 + int(idLen)
	if pos < len(data) {
		// Simplified deserialization
		// In production, you'd use proper deserialization
		entry.Entity = &models.Entity{ID: entry.EntityID}
	}
	
	return entry, nil
}

// Truncate truncates the WAL after a successful checkpoint
func (w *WAL) Truncate() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	// Close the current file
	if err := w.file.Close(); err != nil {
		return err
	}
	
	// Remove the old file
	if err := os.Remove(w.path); err != nil {
		return err
	}
	
	// Create a new file
	file, err := os.OpenFile(w.path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	
	w.file = file
	w.sequence = 0
	
	return nil
}

// Close closes the WAL
func (w *WAL) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	return w.file.Close()
}

// readSequence reads the last sequence number from the WAL
func (w *WAL) readSequence() error {
	// Simplified: just count entries
	// In production, you'd store this in a header
	
	count := uint64(0)
	err := w.Replay(func(entry WALEntry) error {
		count++
		return nil
	})
	
	if err != nil && err != io.EOF {
		return err
	}
	
	w.sequence = count
	return nil
}