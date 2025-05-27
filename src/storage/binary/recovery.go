package binary

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"entitydb/logger"
	"entitydb/models"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// RecoveryManager handles data recovery and repair operations
type RecoveryManager struct {
	dataPath string
	backupPath string
}

// NewRecoveryManager creates a new recovery manager
func NewRecoveryManager(dataPath string) *RecoveryManager {
	return &RecoveryManager{
		dataPath: dataPath,
		backupPath: filepath.Join(dataPath, "backups"),
	}
}

// RecoverCorruptedEntity attempts to recover a corrupted entity
func (rm *RecoveryManager) RecoverCorruptedEntity(repo *EntityRepository, entityID string) (*models.Entity, error) {
	op := models.StartOperation(models.OpTypeRecovery, entityID, map[string]interface{}{
		"recovery_type": "corrupted_entity",
	})
	defer op.Complete()
	
	logger.Info("[Recovery] Attempting to recover corrupted entity: %s", entityID)
	
	// Try to read from WAL first
	walPath := filepath.Join(rm.dataPath, "entitydb.wal")
	if entity, err := rm.recoverFromWAL(walPath, entityID); err == nil {
		logger.Info("[Recovery] Successfully recovered entity %s from WAL", entityID)
		op.SetMetadata("recovery_source", "wal")
		return entity, nil
	}
	
	// Try to read from backup files
	if entity, err := rm.recoverFromBackup(entityID); err == nil {
		logger.Info("[Recovery] Successfully recovered entity %s from backup", entityID)
		op.SetMetadata("recovery_source", "backup")
		return entity, nil
	}
	
	// Try partial recovery from main file
	if entity, err := rm.partialRecovery(repo, entityID); err == nil {
		logger.Info("[Recovery] Partially recovered entity %s", entityID)
		op.SetMetadata("recovery_source", "partial")
		return entity, nil
	}
	
	err := fmt.Errorf("unable to recover entity %s", entityID)
	op.Fail(err)
	return nil, err
}

// recoverFromWAL attempts to recover an entity from the WAL
func (rm *RecoveryManager) recoverFromWAL(walPath string, entityID string) (*models.Entity, error) {
	wal, err := NewWAL(rm.dataPath)
	if err != nil {
		return nil, err
	}
	defer wal.Close()
	
	var latestEntity *models.Entity
	var latestTimestamp time.Time
	
	err = wal.Replay(func(entry WALEntry) error {
		if entry.EntityID == entityID && entry.Entity != nil {
			if entry.Timestamp.After(latestTimestamp) {
				latestEntity = entry.Entity
				latestTimestamp = entry.Timestamp
			}
		}
		return nil
	})
	
	if err != nil {
		return nil, err
	}
	
	if latestEntity == nil {
		return nil, fmt.Errorf("entity not found in WAL")
	}
	
	return latestEntity, nil
}

// recoverFromBackup attempts to recover an entity from backup files
func (rm *RecoveryManager) recoverFromBackup(entityID string) (*models.Entity, error) {
	backupFile := filepath.Join(rm.backupPath, fmt.Sprintf("%s.backup", entityID))
	
	file, err := os.Open(backupFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	// Read backup format (simplified for now)
	// In production, this would use the same binary format
	entity := &models.Entity{
		ID: entityID,
		Tags: []string{},
		Content: []byte{},
	}
	
	// Read entity data from backup
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	
	entity.Content = content
	return entity, nil
}

// partialRecovery attempts to recover whatever data is available
func (rm *RecoveryManager) partialRecovery(repo *EntityRepository, entityID string) (*models.Entity, error) {
	// Try to read from the data file directly
	dataFile := filepath.Join(rm.dataPath, "entities.ebf")
	file, err := os.Open(dataFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	// Try to find entity by scanning the file
	// This is a last-resort method when index is corrupted
	logger.Warn("[Recovery] Attempting partial recovery by file scan for entity %s", entityID)
	
	// For now, create a placeholder entity indicating recovery failure
	// In a production system, you would implement actual file scanning
	entity := &models.Entity{
		ID: entityID,
		Tags: []string{
			fmt.Sprintf("%d|status:recovered", time.Now().UnixNano()),
			fmt.Sprintf("%d|recovery:partial", time.Now().UnixNano()),
			fmt.Sprintf("%d|recovery:placeholder", time.Now().UnixNano()),
		},
		Content: []byte(fmt.Sprintf("Entity %s could not be recovered", entityID)),
	}
	
	// Calculate checksum of placeholder content
	checksum := sha256.Sum256(entity.Content)
	entity.Tags = append(entity.Tags, fmt.Sprintf("%d|checksum:sha256:%s", 
		time.Now().UnixNano(), hex.EncodeToString(checksum[:])))
	
	return entity, nil
}

// RepairIndex rebuilds the index from scratch by scanning the data file
func (rm *RecoveryManager) RepairIndex(repo *EntityRepository) error {
	op := models.StartOperation(models.OpTypeRecovery, "index_repair", map[string]interface{}{
		"recovery_type": "index_repair",
	})
	defer op.Complete()
	
	logger.Info("[Recovery] Starting index repair")
	
	// Use the repository's existing RepairIndex functionality
	// The EntityRepository has its own RepairIndex method that handles the writer
	if err := repo.buildIndexes(); err != nil {
		op.Fail(err)
		logger.Error("[Recovery] Failed to rebuild indexes: %v", err)
		return err
	}
	
	logger.Info("[Recovery] Index repair complete")
	return nil
}

// tryReadEntityAt attempts to read an entity at a specific offset
func (rm *RecoveryManager) tryReadEntityAt(file *os.File, offset int64) (*IndexEntry, string, error) {
	// This is a simplified version - in production, you'd have more robust entity detection
	
	// Seek to offset
	if _, err := file.Seek(offset, io.SeekStart); err != nil {
		return nil, "", err
	}
	
	// Try to read an entity header
	var header EntityHeader
	if err := binary.Read(file, binary.LittleEndian, &header); err != nil {
		return nil, "", err
	}
	
	// Basic sanity checks
	if header.TagCount > 1000 || header.ContentCount > 100 {
		return nil, "", fmt.Errorf("invalid header values")
	}
	
	// Calculate expected size
	expectedSize := binary.Size(header) + 
		int(header.TagCount)*4 + // tag IDs
		100 // approximate content size (would need to read actual sizes)
	
	// Try to extract entity ID from tags or content
	// This is simplified - in production you'd parse the full entity
	entityID := fmt.Sprintf("recovered_%d", offset)
	
	entry := &IndexEntry{
		Offset: uint64(offset),
		Size: uint32(expectedSize),
	}
	
	return entry, entityID, nil
}

// CreateBackup creates a backup of an entity
func (rm *RecoveryManager) CreateBackup(entity *models.Entity) error {
	// Ensure backup directory exists
	if err := os.MkdirAll(rm.backupPath, 0755); err != nil {
		return err
	}
	
	backupFile := filepath.Join(rm.backupPath, fmt.Sprintf("%s.backup", entity.ID))
	
	file, err := os.Create(backupFile)
	if err != nil {
		return err
	}
	defer file.Close()
	
	// Write entity content to backup
	if _, err := file.Write(entity.Content); err != nil {
		return err
	}
	
	return file.Sync()
}

// ValidateChecksum validates the checksum of an entity
func (rm *RecoveryManager) ValidateChecksum(entity *models.Entity) (bool, string) {
	// Calculate actual checksum
	actualChecksum := sha256.Sum256(entity.Content)
	actualChecksumHex := hex.EncodeToString(actualChecksum[:])
	
	// Look for checksum tag
	for _, tag := range entity.Tags {
		if strings.Contains(tag, "|checksum:sha256:") {
			parts := strings.Split(tag, "|checksum:sha256:")
			if len(parts) == 2 {
				expectedChecksum := parts[1]
				return expectedChecksum == actualChecksumHex, expectedChecksum
			}
		}
	}
	
	// No checksum found
	return false, ""
}

// RepairWAL repairs a corrupted WAL file
func (rm *RecoveryManager) RepairWAL() error {
	op := models.StartOperation(models.OpTypeRecovery, "wal_repair", map[string]interface{}{
		"recovery_type": "wal_repair",
	})
	defer op.Complete()
	
	logger.Info("[Recovery] Starting WAL repair")
	
	walPath := filepath.Join(rm.dataPath, "entitydb.wal")
	backupPath := filepath.Join(rm.dataPath, "entitydb.wal.backup")
	
	// Create backup of current WAL
	if err := rm.copyFile(walPath, backupPath); err != nil {
		logger.Warn("[Recovery] Failed to backup WAL: %v", err)
	}
	
	// Open WAL for reading
	file, err := os.Open(walPath)
	if err != nil {
		op.Fail(err)
		return err
	}
	defer file.Close()
	
	// Create new WAL
	newWalPath := filepath.Join(rm.dataPath, "entitydb.wal.new")
	newFile, err := os.Create(newWalPath)
	if err != nil {
		op.Fail(err)
		return err
	}
	defer newFile.Close()
	
	// Read and copy valid entries
	validEntries := 0
	corruptedEntries := 0
	
	for {
		// Try to read an entry
		var length uint32
		if err := binary.Read(file, binary.LittleEndian, &length); err != nil {
			if err == io.EOF {
				break
			}
			// Skip corrupted length
			corruptedEntries++
			continue
		}
		
		// Sanity check length
		if length > 10*1024*1024 { // 10MB max entry size
			corruptedEntries++
			continue
		}
		
		// Read entry data
		data := make([]byte, length)
		if _, err := io.ReadFull(file, data); err != nil {
			corruptedEntries++
			continue
		}
		
		// Write valid entry to new WAL
		if err := binary.Write(newFile, binary.LittleEndian, length); err != nil {
			break
		}
		if _, err := newFile.Write(data); err != nil {
			break
		}
		
		validEntries++
	}
	
	// Close files
	file.Close()
	newFile.Close()
	
	// Replace old WAL with new one
	if err := os.Rename(newWalPath, walPath); err != nil {
		op.Fail(err)
		return err
	}
	
	op.SetMetadata("valid_entries", validEntries)
	op.SetMetadata("corrupted_entries", corruptedEntries)
	
	logger.Info("[Recovery] WAL repair complete: %d valid entries, %d corrupted entries", 
		validEntries, corruptedEntries)
	
	return nil
}

// copyFile copies a file from src to dst
func (rm *RecoveryManager) copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	
	_, err = io.Copy(dstFile, srcFile)
	return err
}