package binary

import (
	"encoding/binary"
	"entitydb/logger"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// IndexCorruptionRecovery handles specific index corruption issues
type IndexCorruptionRecovery struct {
	dataPath string
}

// NewIndexCorruptionRecovery creates a new index corruption recovery manager
func NewIndexCorruptionRecovery(dataPath string) *IndexCorruptionRecovery {
	return &IndexCorruptionRecovery{
		dataPath: dataPath,
	}
}

// DiagnoseAndRecover performs comprehensive index corruption diagnosis and recovery
func (icr *IndexCorruptionRecovery) DiagnoseAndRecover() error {
	logger.Info("Starting index corruption diagnosis and recovery...")
	
	// Step 1: Diagnose the corruption
	corruptionFound, err := icr.diagnoseCorruption()
	if err != nil {
		return fmt.Errorf("corruption diagnosis failed: %v", err)
	}
	
	if !corruptionFound {
		logger.Info("No index corruption detected")
		return nil
	}
	
	// Step 2: Create backup
	if err := icr.createIndexBackup(); err != nil {
		logger.Warn("Failed to create index backup: %v", err)
	}
	
	// Step 3: Recover the index
	if err := icr.recoverIndex(); err != nil {
		return fmt.Errorf("index recovery failed: %v", err)
	}
	
	logger.Info("Index corruption recovery completed successfully")
	return nil
}

// diagnoseCorruption checks for index corruption patterns
func (icr *IndexCorruptionRecovery) diagnoseCorruption() (bool, error) {
	dbPath := filepath.Join(icr.dataPath, "entities.edb")
	
	// Check if unified database file exists
	dbStat, err := os.Stat(dbPath)
	if err != nil {
		return false, fmt.Errorf("unified database file not accessible: %v", err)
	}
	
	// With unified format, we check the file structure rather than separate files
	logger.Debug("Checking unified database file: %s (size: %d)", dbPath, dbStat.Size())
	
	// For unified format, we'll use a simpler corruption check
	// The recovery system in the main codebase handles detailed validation
	logger.Debug("Unified format corruption diagnosis completed - relying on built-in recovery")
	
	return false, nil // Built-in recovery handles corruption
}

// validateIndexFile validates index entries against database file size
func (icr *IndexCorruptionRecovery) validateIndexFile(idxPath string, dbSize int64) (int, error) {
	file, err := os.Open(idxPath)
	if err != nil {
		return 0, err
	}
	defer file.Close()
	
	// Read header
	var entityCount, tagCount uint32
	if err := binary.Read(file, binary.LittleEndian, &entityCount); err != nil {
		return 0, fmt.Errorf("failed to read entity count: %v", err)
	}
	if err := binary.Read(file, binary.LittleEndian, &tagCount); err != nil {
		return 0, fmt.Errorf("failed to read tag count: %v", err)
	}
	
	logger.Debug("Index header: %d entities, %d tags", entityCount, tagCount)
	
	corruptCount := 0
	
	// Validate each entity entry
	for i := uint32(0); i < entityCount; i++ {
		var offset uint64
		var size uint32
		
		if err := binary.Read(file, binary.LittleEndian, &offset); err != nil {
			logger.Debug("Failed to read offset for entry %d: %v", i, err)
			corruptCount++
			continue
		}
		
		if err := binary.Read(file, binary.LittleEndian, &size); err != nil {
			logger.Debug("Failed to read size for entry %d: %v", i, err)
			corruptCount++
			continue
		}
		
		// Check for obviously corrupted offsets
		if offset > uint64(dbSize) {
			logger.Debug("Entry %d has invalid offset %d (file size: %d)", i, offset, dbSize)
			corruptCount++
		}
		
		// Check for suspicious patterns in the corrupted offset from logs
		if offset == 16375026676881176 { // The specific corruption we saw
			logger.Warn("Found the specific corrupted offset pattern at entry %d", i)
			corruptCount++
		}
	}
	
	return corruptCount, nil
}

// createIndexBackup creates a backup of the corrupted unified file
func (icr *IndexCorruptionRecovery) createIndexBackup() error {
	timestamp := time.Now().Format("20060102-150405")
	unifiedPath := filepath.Join(icr.dataPath, "entities.edb")
	backupPath := filepath.Join(icr.dataPath, fmt.Sprintf("entities.edb.corrupt-%s", timestamp))
	
	return icr.copyFile(unifiedPath, backupPath)
}

// recoverIndex rebuilds the index by completely removing it and forcing a rebuild
func (icr *IndexCorruptionRecovery) recoverIndex() error {
	logger.Info("Rebuilding corrupted index...")
	
	idxPath := filepath.Join(icr.dataPath, "entities.db.idx")
	
	// For severe recurring corruption, completely remove the index file
	// This forces a complete rebuild from the database file
	if err := os.Remove(idxPath); err != nil && !os.IsNotExist(err) {
		logger.Warn("Failed to remove corrupted index file: %v", err)
	} else {
		logger.Info("Removed corrupted index file - will be rebuilt from database")
	}
	
	// Also remove any temporary or backup index files that might cause issues
	tempIndexFiles := []string{
		filepath.Join(icr.dataPath, "entities.db.idx.new"),
		filepath.Join(icr.dataPath, "entities.db.idx.tmp"),
		filepath.Join(icr.dataPath, "entities.db.idx.rebuild"),
	}
	
	for _, tempFile := range tempIndexFiles {
		if err := os.Remove(tempFile); err == nil {
			logger.Info("Removed temporary index file: %s", tempFile)
		}
	}
	
	logger.Info("Index recovery completed - clean rebuild will occur during startup")
	return nil
}

// RecoveredIndexEntry represents a valid index entry during recovery
type RecoveredIndexEntry struct {
	Offset uint64
	Size   uint32
}

// extractValidEntries reads the corrupted index and extracts only valid entries
func (icr *IndexCorruptionRecovery) extractValidEntries(idxPath string, dbSize int64) ([]RecoveredIndexEntry, error) {
	file, err := os.Open(idxPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	// Read header
	var entityCount, tagCount uint32
	if err := binary.Read(file, binary.LittleEndian, &entityCount); err != nil {
		return nil, err
	}
	if err := binary.Read(file, binary.LittleEndian, &tagCount); err != nil {
		return nil, err
	}
	
	var validEntries []RecoveredIndexEntry
	
	// Process each entry
	for i := uint32(0); i < entityCount; i++ {
		var offset uint64
		var size uint32
		
		if err := binary.Read(file, binary.LittleEndian, &offset); err != nil {
			logger.Debug("Skipping entry %d due to read error: %v", i, err)
			continue
		}
		
		if err := binary.Read(file, binary.LittleEndian, &size); err != nil {
			logger.Debug("Skipping entry %d due to size read error: %v", i, err)
			continue
		}
		
		// Validate entry
		if offset > uint64(dbSize) {
			logger.Debug("Skipping corrupted entry %d: offset %d exceeds file size %d", 
				i, offset, dbSize)
			continue
		}
		
		if size == 0 {
			logger.Debug("Skipping entry %d with zero size", i)
			continue
		}
		
		// Entry looks valid
		validEntries = append(validEntries, RecoveredIndexEntry{
			Offset: offset,
			Size:   size,
		})
	}
	
	return validEntries, nil
}

// writeCleanIndex writes a new index file with only valid entries
func (icr *IndexCorruptionRecovery) writeCleanIndex(newIdxPath string, entries []RecoveredIndexEntry) error {
	file, err := os.Create(newIdxPath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	// Write header
	entityCount := uint32(len(entries))
	tagCount := uint32(0) // Will be rebuilt by the repository
	
	if err := binary.Write(file, binary.LittleEndian, entityCount); err != nil {
		return err
	}
	if err := binary.Write(file, binary.LittleEndian, tagCount); err != nil {
		return err
	}
	
	// Write valid entries
	for _, entry := range entries {
		if err := binary.Write(file, binary.LittleEndian, entry.Offset); err != nil {
			return err
		}
		if err := binary.Write(file, binary.LittleEndian, entry.Size); err != nil {
			return err
		}
	}
	
	logger.Debug("Wrote clean index with %d entries", len(entries))
	return nil
}

// copyFile copies a file for backup purposes
func (icr *IndexCorruptionRecovery) copyFile(src, dst string) error {
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
	
	buf := make([]byte, 64*1024)
	for {
		n, err := srcFile.Read(buf)
		if n > 0 {
			if _, writeErr := dstFile.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
		}
		if err != nil {
			break
		}
	}
	
	return nil
}

// ForceIndexRebuild removes the index to force a complete rebuild
func (icr *IndexCorruptionRecovery) ForceIndexRebuild() error {
	logger.Info("Forcing complete index rebuild...")
	
	idxPath := filepath.Join(icr.dataPath, "entities.db.idx")
	backupPath := filepath.Join(icr.dataPath, "entities.db.idx.force-rebuild")
	
	// Backup existing index
	if err := icr.copyFile(idxPath, backupPath); err != nil {
		logger.Warn("Failed to backup index before rebuild: %v", err)
	}
	
	// Remove corrupted index
	if err := os.Remove(idxPath); err != nil {
		return fmt.Errorf("failed to remove corrupted index: %v", err)
	}
	
	logger.Info("Index removed - system will rebuild on next startup")
	return nil
}