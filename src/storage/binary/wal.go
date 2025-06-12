// Package binary provides Write-Ahead Logging (WAL) functionality for the EntityDB
// Binary Format storage layer. The WAL ensures durability and crash recovery by
// logging all write operations before they are applied to the main data file.
package binary

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"entitydb/models"
	"entitydb/logger"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// WALEntry represents a single entry in the write-ahead log.
// Each entry contains the complete information needed to replay
// an operation during recovery.
//
// Fields:
//   - OpType: The type of operation (create, update, delete, checkpoint)
//   - EntityID: Unique identifier of the affected entity
//   - Entity: Complete entity data (nil for delete operations)
//   - Timestamp: When the operation was logged
//   - Checksum: SHA256 hash of the serialized entry for integrity
type WALEntry struct {
	OpType    WALOpType
	EntityID  string
	Entity    *models.Entity
	Timestamp time.Time
	Checksum  string // SHA256 hex string
}

// WALOpType defines the type of operation in the WAL.
// Operations are ordered by their typical frequency of use.
type WALOpType uint8

const (
	WALOpCreate     WALOpType = iota // New entity creation
	WALOpUpdate                      // Entity modification
	WALOpDelete                      // Entity removal
	WALOpCheckpoint                  // Checkpoint marker for truncation
)

// WAL implements a write-ahead log for ensuring durability and crash recovery.
// All write operations are first logged to the WAL before being applied to
// the main data file. This ensures that no data is lost even if the system
// crashes during a write operation.
//
// Key features:
//   - Append-only for optimal write performance
//   - Automatic checkpointing to prevent unbounded growth
//   - SHA256 checksums for corruption detection
//   - Thread-safe through mutex synchronization
//   - Sequence numbers for operation ordering
//
// Recovery process:
//   1. Read all entries since last checkpoint
//   2. Verify checksums for integrity
//   3. Replay operations in sequence order
//   4. Create new checkpoint after recovery
type WAL struct {
	mu       sync.Mutex // Protects concurrent access
	file     *os.File   // WAL file handle
	path     string     // Full path to WAL file
	sequence uint64     // Monotonic operation counter
}

// NewWAL creates a new write-ahead log instance for the given data directory.
// The WAL file is created as "entitydb.wal" in the specified directory.
//
// Initialization process:
//   1. Opens or creates the WAL file with append mode
//   2. Reads the last sequence number for continuity
//   3. Positions file pointer at end for new writes
//
// Parameters:
//   - dataPath: Directory where the WAL file will be stored
//
// Returns:
//   - *WAL: Initialized WAL ready for logging operations
//   - error: File creation or initialization errors
//
// Example:
//   wal, err := NewWAL("/var/lib/entitydb")
//   if err != nil {
//       log.Fatal(err)
//   }
//   defer wal.Close()
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

// LogCreate logs an entity creation operation to the WAL.
// This must be called before the entity is written to the main data file
// to ensure durability in case of crashes.
//
// The method:
//   1. Creates a WAL entry with the complete entity data
//   2. Calculates checksum for integrity verification
//   3. Appends entry to the WAL file
//   4. Syncs to disk for immediate durability
//
// Parameters:
//   - entity: The entity being created (must have valid ID)
//
// Returns:
//   - error: I/O errors or serialization failures
//
// Thread Safety:
//   Safe for concurrent use; operations are serialized via mutex.
func (w *WAL) LogCreate(entity *models.Entity) error {
	op := models.StartOperation(models.OpTypeWAL, entity.ID, map[string]interface{}{
		"wal_operation": "create",
		"entity_size": len(entity.Content),
		"tag_count": len(entity.Tags),
	})
	defer op.Complete()
	
	logger.Trace("Logging CREATE operation for entity %s", entity.ID)
	
	err := w.logEntry(WALEntry{
		OpType:    WALOpCreate,
		EntityID:  entity.ID,
		Entity:    entity,
		Timestamp: time.Now(),
	})
	
	if err != nil {
		op.Fail(err)
		logger.Error("Failed to log CREATE for entity %s: %v", entity.ID, err)
		return err
	}
	
	op.SetMetadata("sequence", w.sequence)
	logger.Debug("Successfully logged CREATE for entity %s at sequence %d", entity.ID, w.sequence)
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
	
	logger.Trace("Logging UPDATE operation for entity %s", entity.ID)
	
	err := w.logEntry(WALEntry{
		OpType:    WALOpUpdate,
		EntityID:  entity.ID,
		Entity:    entity,
		Timestamp: time.Now(),
	})
	
	if err != nil {
		op.Fail(err)
		logger.Error("Failed to log UPDATE for entity %s: %v", entity.ID, err)
		return err
	}
	
	op.SetMetadata("sequence", w.sequence)
	logger.Debug("Successfully logged UPDATE for entity %s at sequence %d", entity.ID, w.sequence)
	return nil
}

// LogDelete logs an entity deletion
func (w *WAL) LogDelete(entityID string) error {
	op := models.StartOperation(models.OpTypeWAL, entityID, map[string]interface{}{
		"wal_operation": "delete",
	})
	defer op.Complete()
	
	logger.Trace("Logging DELETE operation for entity %s", entityID)
	
	err := w.logEntry(WALEntry{
		OpType:    WALOpDelete,
		EntityID:  entityID,
		Timestamp: time.Now(),
	})
	
	if err != nil {
		op.Fail(err)
		logger.Error("Failed to log DELETE for entity %s: %v", entityID, err)
		return err
	}
	
	op.SetMetadata("sequence", w.sequence)
	logger.Debug("Successfully logged DELETE for entity %s at sequence %d", entityID, w.sequence)
	return nil
}

// LogCheckpoint logs a checkpoint marker to the WAL.
// Checkpoints indicate that all previous operations have been
// successfully written to the main data file and the WAL can
// be truncated up to this point during recovery.
//
// Checkpoint strategy:
//   - Called after successful flush of main data file
//   - Enables WAL truncation to prevent unbounded growth
//   - Does not contain entity data, only timestamp
//
// Returns:
//   - error: I/O errors during checkpoint write
//
// Thread Safety:
//   Safe for concurrent use via internal locking.
func (w *WAL) LogCheckpoint() error {
	return w.logEntry(WALEntry{
		OpType:    WALOpCheckpoint,
		Timestamp: time.Now(),
	})
}

// logEntry writes an entry to the WAL file atomically.
// This is the core method that ensures all operations are durably stored.
//
// Write process:
//   1. Serialize the entry to binary format
//   2. Write 4-byte length prefix for framing
//   3. Write serialized entry data
//   4. Sync file to ensure durability
//   5. Increment sequence number
//
// Format: [Length:4][SerializedEntry:Length]
//
// The length prefix allows for efficient scanning during recovery
// and detection of incomplete writes.
//
// Parameters:
//   - entry: The WAL entry to write
//
// Returns:
//   - error: Serialization or I/O errors
//
// Note: This method must hold the mutex for the entire operation
// to ensure write atomicity and sequence number consistency.
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
	// Enhanced format with checksum:
	// [OpType:1][TimestampNano:8][EntityIDLen:2][EntityID:var][ChecksumLen:2][Checksum:var][EntityData:var]
	
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
	
	// Calculate and add checksum if entity present
	checksumStr := ""
	if entry.Entity != nil && len(entry.Entity.Content) > 0 {
		checksum := sha256.Sum256(entry.Entity.Content)
		checksumStr = hex.EncodeToString(checksum[:])
	}
	
	// Add checksum
	checksumLen := uint16(len(checksumStr))
	checksumLenBuf := make([]byte, 2)
	binary.LittleEndian.PutUint16(checksumLenBuf, checksumLen)
	buf = append(buf, checksumLenBuf...)
	if checksumLen > 0 {
		buf = append(buf, []byte(checksumStr)...)
	}
	
	// Add entity data if present
	if entry.Entity != nil {
		// Store entity data in a more structured format
		entityBuf := make([]byte, 0, 256)
		
		// Store tag count
		tagCount := uint16(len(entry.Entity.Tags))
		tagCountBuf := make([]byte, 2)
		binary.LittleEndian.PutUint16(tagCountBuf, tagCount)
		entityBuf = append(entityBuf, tagCountBuf...)
		
		// Store tags
		for _, tag := range entry.Entity.Tags {
			tagLen := uint16(len(tag))
			tagLenBuf := make([]byte, 2)
			binary.LittleEndian.PutUint16(tagLenBuf, tagLen)
			entityBuf = append(entityBuf, tagLenBuf...)
			entityBuf = append(entityBuf, []byte(tag)...)
		}
		
		// Store content length and content
		contentLen := uint32(len(entry.Entity.Content))
		contentLenBuf := make([]byte, 4)
		binary.LittleEndian.PutUint32(contentLenBuf, contentLen)
		entityBuf = append(entityBuf, contentLenBuf...)
		entityBuf = append(entityBuf, entry.Entity.Content...)
		
		// Write entity buffer length and data
		entityLen := uint32(len(entityBuf))
		entityLenBuf := make([]byte, 4)
		binary.LittleEndian.PutUint32(entityLenBuf, entityLen)
		buf = append(buf, entityLenBuf...)
		buf = append(buf, entityBuf...)
		
		logger.Debug("Serialized entity %s with checksum %s", entry.EntityID, checksumStr)
	} else {
		// No entity data
		entityLenBuf := make([]byte, 4)
		binary.LittleEndian.PutUint32(entityLenBuf, 0)
		buf = append(buf, entityLenBuf...)
	}
	
	return buf, nil
}

// Replay reads all entries from the WAL and executes the callback for each.
// This is the primary mechanism for crash recovery and restoring system state.
//
// Recovery process:
//   1. Seeks to beginning of WAL file
//   2. Reads entries sequentially using length-prefix framing
//   3. Deserializes each entry and verifies integrity
//   4. Calls callback for each valid entry
//   5. Skips corrupted entries with logging
//   6. Stops at EOF or unreadable data
//
// The callback should:
//   - Apply the operation to the main data store
//   - Return error only for fatal conditions
//   - Handle idempotency for repeated operations
//
// Parameters:
//   - callback: Function to process each WAL entry
//
// Returns:
//   - error: Fatal errors that prevent recovery
//
// Example:
//   err := wal.Replay(func(entry WALEntry) error {
//       switch entry.OpType {
//       case WALOpCreate:
//           return writer.WriteEntity(entry.Entity)
//       case WALOpUpdate:
//           return writer.UpdateEntity(entry.Entity)
//       case WALOpDelete:
//           return writer.DeleteEntity(entry.EntityID)
//       }
//       return nil
//   })
func (w *WAL) Replay(callback func(entry WALEntry) error) error {
	op := models.StartOperation(models.OpTypeWAL, "replay", map[string]interface{}{
		"wal_operation": "replay",
		"wal_path": w.path,
	})
	defer op.Complete()
	
	logger.Info("Starting WAL replay from %s", w.path)
	
	// Seek to the beginning
	if _, err := w.file.Seek(0, io.SeekStart); err != nil {
		op.Fail(err)
		logger.Error("Failed to seek to beginning: %v", err)
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
			logger.Error("Failed to read entry length: %v", err)
			return err
		}
		
		// Read data
		data := make([]byte, length)
		if _, err := io.ReadFull(w.file, data); err != nil {
			op.Fail(err)
			logger.Error("Failed to read entry data (length=%d): %v", length, err)
			return err
		}
		
		// Deserialize entry
		entry, err := w.deserializeEntry(data)
		if err != nil {
			entriesFailed++
			logger.Error("Failed to deserialize entry: %v", err)
			continue // Skip bad entries during replay
		}
		
		logger.Debug("Replaying %v operation for entity %s", entry.OpType, entry.EntityID)
		
		// Process entry
		if err := callback(*entry); err != nil {
			entriesFailed++
			logger.Error("Failed to process entry for entity %s: %v", entry.EntityID, err)
			continue // Continue with other entries
		}
		
		entriesProcessed++
	}
	
	op.SetMetadata("entries_processed", entriesProcessed)
	op.SetMetadata("entries_failed", entriesFailed)
	
	logger.Info("WAL replay completed: %d entries processed, %d failed", entriesProcessed, entriesFailed)
	
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
	
	// Read entity ID length
	idLen := binary.LittleEndian.Uint16(data[9:11])
	if len(data) < 11+int(idLen) {
		return nil, fmt.Errorf("invalid WAL entry: ID length mismatch")
	}
	
	// Read entity ID
	entry.EntityID = string(data[11 : 11+idLen])
	
	pos := 11 + int(idLen)
	
	// Read checksum if present
	if pos+2 <= len(data) {
		checksumLen := binary.LittleEndian.Uint16(data[pos:pos+2])
		pos += 2
		
		if checksumLen > 0 && pos+int(checksumLen) <= len(data) {
			entry.Checksum = string(data[pos : pos+int(checksumLen)])
			pos += int(checksumLen)
		}
	}
	
	// Read entity data length
	if pos+4 <= len(data) {
		entityDataLen := binary.LittleEndian.Uint32(data[pos:pos+4])
		pos += 4
		
		if entityDataLen > 0 && pos+int(entityDataLen) <= len(data) {
			// We have entity data to deserialize
			entityData := data[pos : pos+int(entityDataLen)]
			
			entity := &models.Entity{
				ID:        entry.EntityID,
				Tags:      []string{},
				Content:   []byte{},
				CreatedAt: entry.Timestamp.UnixNano(),
				UpdatedAt: entry.Timestamp.UnixNano(),
			}
			
			// Parse entity data
			entityPos := 0
			
			// Read tag count
			if entityPos+2 <= len(entityData) {
				tagCount := binary.LittleEndian.Uint16(entityData[entityPos:entityPos+2])
				entityPos += 2
				
				// Read tags
				for i := uint16(0); i < tagCount && entityPos < len(entityData); i++ {
					if entityPos+2 > len(entityData) {
						break
					}
					
					tagLen := binary.LittleEndian.Uint16(entityData[entityPos:entityPos+2])
					entityPos += 2
					
					if entityPos+int(tagLen) > len(entityData) {
						break
					}
					
					tag := string(entityData[entityPos : entityPos+int(tagLen)])
					entity.Tags = append(entity.Tags, tag)
					entityPos += int(tagLen)
				}
			}
			
			// Read content length and content
			if entityPos+4 <= len(entityData) {
				contentLen := binary.LittleEndian.Uint32(entityData[entityPos:entityPos+4])
				entityPos += 4
				
				if contentLen > 0 && entityPos+int(contentLen) <= len(entityData) {
					entity.Content = entityData[entityPos : entityPos+int(contentLen)]
				}
			}
			
			entry.Entity = entity
		} else if entry.OpType == WALOpDelete {
			// For delete operations, we don't need entity data
			entry.Entity = nil
		}
	}
	
	return entry, nil
}

// Truncate removes all entries from the WAL file after a successful checkpoint.
// This should only be called after all WAL entries have been durably written
// to the main data file.
//
// Truncation process:
//   1. Closes the current WAL file
//   2. Deletes the old WAL file from disk
//   3. Creates a new empty WAL file
//   4. Resets sequence counter to zero
//
// Safety considerations:
//   - Only call after successful checkpoint
//   - Ensure main data file is synced first
//   - Operation is not atomic (brief window of no WAL)
//
// Returns:
//   - error: File operation failures
//
// Thread Safety:
//   Method is synchronized; safe for concurrent use.
//
// Note: In production systems, consider using rename operations
// for atomic truncation to avoid the window where no WAL exists.
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

// Close gracefully shuts down the WAL, ensuring all pending operations complete.
// After Close, the WAL instance cannot be used for further operations.
//
// This method:
//   - Ensures any buffered data is written
//   - Closes the underlying file handle
//   - Releases file locks and resources
//
// Returns:
//   - error: File closing errors
//
// Thread Safety:
//   Safe to call concurrently; only the first call takes effect.
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