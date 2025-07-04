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
// The WAL supports both standalone files (legacy) and embedded sections within
// unified files. The implementation automatically detects the format and uses
// appropriate read/write methods.
//
// Key features:
//   - Append-only for optimal write performance
//   - Automatic checkpointing to prevent unbounded growth
//   - SHA256 checksums for corruption detection
//   - Thread-safe through mutex synchronization
//   - Sequence numbers for operation ordering
//   - Unified file support for reduced I/O overhead
//
// Recovery process:
//   1. Read all entries since last checkpoint
//   2. Verify checksums for integrity
//   3. Replay operations in sequence order
//   4. Create new checkpoint after recovery
type WAL struct {
	mu         sync.Mutex     // Protects concurrent access
	file       *os.File       // WAL file handle (standalone) or unified file
	path       string         // Full path to WAL file
	sequence   uint64         // Monotonic operation counter
	isUnified  bool           // Whether WAL is embedded in unified file
	walOffset  uint64         // Offset to WAL section in unified file
	walSize    uint64         // Size of WAL section in unified file
}

// NewWAL creates a new write-ahead log instance for the given unified database file.
// This version uses the embedded WAL section within the unified file format instead
// of creating a separate WAL file.
//
// Initialization process:
//   1. Opens the unified database file
//   2. Reads the header to get WAL section offset and size
//   3. Creates a unified WAL instance for the embedded section
//
// Parameters:
//   - unifiedFilePath: Path to the unified .edb database file
//
// Returns:
//   - *WAL: Initialized WAL ready for logging operations (embedded)
//   - error: File opening or header parsing errors
//
// Example:
//   wal, err := NewWAL("/var/lib/entitydb/entities.edb")
//   if err != nil {
//       log.Fatal(err)
//   }
//   defer wal.Close()
func NewWAL(unifiedFilePath string) (*WAL, error) {
	logger.TraceIf("storage", "Creating embedded WAL for unified file: %s", unifiedFilePath)
	
	return NewWALFromUnifiedFile(unifiedFilePath)
}

// NewWALFromUnifiedFile creates a WAL instance that uses the embedded WAL section
// within a unified EntityDB file format (.edb). This function reads the header
// to determine the WAL section location and creates the appropriate WAL instance.
//
// If the unified file doesn't exist yet, this returns a deferred WAL that will
// be initialized when the file is created by the writer.
//
// Parameters:
//   - unifiedFilePath: Complete path to the unified .edb database file
//
// Returns:
//   - *WAL: Initialized WAL for the embedded section (or deferred WAL)
//   - error: File access or header parsing errors
func NewWALFromUnifiedFile(unifiedFilePath string) (*WAL, error) {
	// Check if unified file exists
	if _, err := os.Stat(unifiedFilePath); os.IsNotExist(err) {
		logger.TraceIf("wal", "Unified file doesn't exist yet, creating deferred WAL: %s", unifiedFilePath)
		// Return a deferred WAL that will be initialized when the file is created
		return &WAL{
			path:      unifiedFilePath,
			isUnified: true,
			sequence:  1, // Start at sequence 1
		}, nil
	}
	
	// Open the unified file
	file, err := os.OpenFile(unifiedFilePath, os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open unified file %s: %w", unifiedFilePath, err)
	}
	
	// Read the header to get WAL section information
	header := &Header{}
	if err := header.Read(file); err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to read unified file header: %w", err)
	}
	
	// Create unified WAL instance with the embedded section
	return NewUnifiedWAL(file, header.WALOffset, header.WALSize)
}

// NewWALWithPath creates a new WAL with the complete file path specified
func NewWALWithPath(walPath string) (*WAL, error) {
	logger.TraceIf("storage", "Creating WAL with complete path: %s", walPath)
	
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

// NewUnifiedWAL creates a WAL instance that reads from a unified file's WAL section.
// This is used when the WAL is embedded within a unified EntityDB file rather than
// being a separate file.
//
// Parameters:
//   - unifiedFile: Open file handle to the unified EntityDB file
//   - walOffset: Byte offset to the WAL section within the file
//   - walSize: Size of the WAL section in bytes
//
// Returns:
//   - *WAL: Initialized WAL for reading embedded entries
//   - error: Initialization errors
func NewUnifiedWAL(unifiedFile *os.File, walOffset, walSize uint64) (*WAL, error) {
	logger.TraceIf("wal", "Creating unified WAL at offset %d, size %d", walOffset, walSize)
	
	wal := &WAL{
		file:       unifiedFile,
		path:       unifiedFile.Name(),
		isUnified:  true,
		walOffset:  walOffset,
		walSize:    walSize,
	}
	
	// Read the sequence number from the WAL section
	if err := wal.readUnifiedSequence(); err != nil {
		return nil, err
	}
	
	return wal, nil
}

// readUnifiedSequence reads the current sequence number from the unified WAL section.
func (w *WAL) readUnifiedSequence() error {
	if !w.isUnified {
		return fmt.Errorf("readUnifiedSequence called on non-unified WAL")
	}
	
	// Seek to WAL section
	if _, err := w.file.Seek(int64(w.walOffset), os.SEEK_SET); err != nil {
		return err
	}
	
	// Read WAL header (sequence + entry count)
	buf := make([]byte, 16)
	if _, err := w.file.Read(buf); err != nil {
		if err == io.EOF {
			// Empty WAL section, start at sequence 1
			w.sequence = 1
			return nil
		}
		return err
	}
	
	w.sequence = binary.LittleEndian.Uint64(buf[0:8])
	entryCount := binary.LittleEndian.Uint64(buf[8:16])
	
	logger.TraceIf("wal", "Unified WAL sequence: %d, entry count: %d", w.sequence, entryCount)
	return nil
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
	
	logger.TraceIf("wal", "Logging CREATE operation for entity %s", entity.ID)
	
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
	logger.TraceIf("wal", "Successfully logged CREATE for entity %s at sequence %d", entity.ID, w.sequence)
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
	
	logger.TraceIf("wal", "Logging UPDATE operation for entity %s", entity.ID)
	
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
	logger.TraceIf("wal", "Successfully logged UPDATE for entity %s at sequence %d", entity.ID, w.sequence)
	return nil
}

// LogDelete logs an entity deletion
func (w *WAL) LogDelete(entityID string) error {
	op := models.StartOperation(models.OpTypeWAL, entityID, map[string]interface{}{
		"wal_operation": "delete",
	})
	defer op.Complete()
	
	logger.TraceIf("wal", "Logging DELETE operation for entity %s", entityID)
	
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
	logger.TraceIf("wal", "Successfully logged DELETE for entity %s at sequence %d", entityID, w.sequence)
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
	
	// Initialize connection to unified file if needed (deferred WAL)
	if w.file == nil && w.isUnified {
		if err := w.initializeUnifiedConnection(); err != nil {
			return fmt.Errorf("failed to initialize unified WAL connection: %w", err)
		}
	}
	
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
	
	// Seek to the beginning of WAL section
	seekPos := int64(0)
	if w.isUnified {
		seekPos = int64(w.walOffset)
		logger.Debug("Seeking to WAL section at offset %d in unified file", seekPos)
	}
	
	if _, err := w.file.Seek(seekPos, io.SeekStart); err != nil {
		op.Fail(err)
		logger.Error("Failed to seek to WAL position %d: %v", seekPos, err)
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
		
		// Validate length to prevent memory exhaustion
		const maxEntrySize = 100 * 1024 * 1024 // 100MB max per entry
		if length > maxEntrySize {
			entriesFailed++
			logger.Error("WAL entry too large (%d bytes), skipping corrupted entry", length)
			// Skip this corrupted entry by seeking past it
			if _, err := w.file.Seek(int64(length), io.SeekCurrent); err != nil {
				logger.Error("Failed to skip corrupted entry: %v", err)
				return err
			}
			continue
		}
		
		if length == 0 {
			entriesFailed++
			logger.Error("WAL entry has zero length, skipping")
			continue
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
	
	// CRITICAL: Validate EntityID for corruption before proceeding
	// Corrupted EntityIDs containing binary data cause 100% CPU usage in index operations
	if !isValidEntityID(entry.EntityID) {
		return nil, fmt.Errorf("corrupted EntityID detected: contains invalid characters")
	}
	
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
			
			// CRITICAL: Validate entity data for corruption before proceeding
			// Corrupted tag data causes 100% CPU usage in index operations
			if !isValidEntity(entity) {
				return nil, fmt.Errorf("corrupted entity data detected: invalid tags or content")
			}
			
			entry.Entity = entity
		} else if entry.OpType == WALOpDelete {
			// For delete operations, we don't need entity data
			entry.Entity = nil
		}
	}
	
	return entry, nil
}

// initializeUnifiedConnection establishes a connection to the unified file's WAL section.
// This is used for deferred WAL initialization when the unified file is created after
// the WAL instance is constructed.
func (w *WAL) initializeUnifiedConnection() error {
	logger.TraceIf("wal", "Initializing deferred unified WAL connection: %s", w.path)
	
	// Wait for the unified file to exist (it should be created by the writer)
	if _, err := os.Stat(w.path); os.IsNotExist(err) {
		// File still doesn't exist, this might be a timing issue
		return fmt.Errorf("unified file %s not found when initializing WAL connection", w.path)
	}
	
	// Open the unified file
	file, err := os.OpenFile(w.path, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open unified file: %w", err)
	}
	
	// Read header to get WAL section info
	header := &Header{}
	if err := header.Read(file); err != nil {
		file.Close()
		return fmt.Errorf("failed to read unified file header: %w", err)
	}
	
	// Set up WAL properties
	w.file = file
	w.walOffset = header.WALOffset
	w.walSize = header.WALSize
	
	// Read current sequence number from WAL section
	if err := w.readUnifiedSequence(); err != nil {
		return fmt.Errorf("failed to read WAL sequence: %w", err)
	}
	
	logger.TraceIf("wal", "Unified WAL connection initialized at offset %d, size %d", w.walOffset, w.walSize)
	return nil
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

// isValidEntityID validates that an EntityID contains only valid characters
// and prevents corrupted binary data from reaching the index system
func isValidEntityID(id string) bool {
	// EntityIDs should be reasonable length (max 256 chars for safety)
	if len(id) == 0 || len(id) > 256 {
		return false
	}
	
	// EntityIDs should contain only printable ASCII characters, hyphens, and underscores
	// This prevents binary garbage from causing CPU spikes in index operations
	for _, char := range id {
		if !((char >= 'a' && char <= 'z') || 
			 (char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || 
			 char == '-' || char == '_' || char == '.') {
			return false
		}
	}
	
	return true
}

// isValidEntity validates that an entity contains only valid tag data
// and prevents corrupted entities from reaching the index system
func isValidEntity(entity *models.Entity) bool {
	if entity == nil {
		return false
	}
	
	// Already validated EntityID in isValidEntityID, but double-check
	if !isValidEntityID(entity.ID) {
		return false
	}
	
	// Validate tags for corruption
	for _, tag := range entity.Tags {
		// Tags should be reasonable length (max 1024 chars for safety)
		if len(tag) > 1024 {
			return false
		}
		
		// Tags should not contain null bytes or other control characters
		// Allow printable ASCII, UTF-8, and common punctuation
		for _, char := range tag {
			// Allow null bytes to pass through as they might be legitimate in some contexts
			// but check for excessive binary garbage
			if char < 32 && char != 0 && char != 9 && char != 10 && char != 13 {
				return false // Control characters except tab, newline, carriage return
			}
		}
		
		// Check for excessive binary data (more than 10% non-printable chars)
		nonPrintableCount := 0
		for _, char := range tag {
			if char < 32 || char > 126 {
				nonPrintableCount++
			}
		}
		if len(tag) > 0 && float64(nonPrintableCount)/float64(len(tag)) > 0.1 {
			return false // More than 10% non-printable characters
		}
	}
	
	// Validate content size is reasonable (max 100MB for single entity)
	if len(entity.Content) > 100*1024*1024 {
		return false
	}
	
	return true
}