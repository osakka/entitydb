// Package binary provides atomic file operations for corruption prevention
//
// This system ensures all file operations are atomic, preventing:
// - Torn writes (partial data written during crash)
// - Inconsistent file states
// - Data corruption from interrupted operations
// - Race conditions during concurrent access
package binary

import (
	"crypto/sha256"
	"encoding/hex"
	"entitydb/logger"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

// AtomicFileManager manages atomic file operations
type AtomicFileManager struct {
	mu                sync.RWMutex
	
	// Configuration
	tempSuffix        string        // Suffix for temporary files
	backupSuffix      string        // Suffix for backup files
	syncAfterWrite    bool          // Whether to sync after write
	verifyChecksum    bool          // Whether to verify checksums
	maxRetries        int           // Maximum retry attempts
	retryDelay        time.Duration // Delay between retries
	
	// State tracking
	activeOperations  map[string]*AtomicOperation
	operationCount    int64
	successCount      int64
	failureCount      int64
	
	// Cleanup
	cleanupInterval   time.Duration
	lastCleanup       time.Time
	stopCleanup       chan struct{}
	cleanupRunning    int32
}

// AtomicOperation represents a single atomic file operation
type AtomicOperation struct {
	ID             string
	TargetPath     string
	TempPath       string
	BackupPath     string
	Operation      AtomicOperationType
	StartTime      time.Time
	Data           []byte
	Checksum       string
	Retries        int
	mu             sync.Mutex
}

// AtomicOperationType defines the type of atomic operation
type AtomicOperationType int

const (
	AtomicOpWrite AtomicOperationType = iota
	AtomicOpUpdate
	AtomicOpDelete
	AtomicOpMove
)

// String returns string representation of operation type
func (ot AtomicOperationType) String() string {
	switch ot {
	case AtomicOpWrite:
		return "WRITE"
	case AtomicOpUpdate:
		return "UPDATE"
	case AtomicOpDelete:
		return "DELETE"
	case AtomicOpMove:
		return "MOVE"
	default:
		return "UNKNOWN"
	}
}

// NewAtomicFileManager creates a new atomic file manager
func NewAtomicFileManager() *AtomicFileManager {
	return &AtomicFileManager{
		tempSuffix:        ".tmp",
		backupSuffix:      ".backup",
		syncAfterWrite:    true,
		verifyChecksum:    true,
		maxRetries:        3,
		retryDelay:        100 * time.Millisecond,
		activeOperations:  make(map[string]*AtomicOperation),
		cleanupInterval:   5 * time.Minute,
		lastCleanup:       time.Now(),
		stopCleanup:       make(chan struct{}),
	}
}

// Configure sets atomic file manager parameters
func (afm *AtomicFileManager) Configure(syncAfterWrite, verifyChecksum bool, maxRetries int) {
	afm.mu.Lock()
	defer afm.mu.Unlock()
	
	afm.syncAfterWrite = syncAfterWrite
	afm.verifyChecksum = verifyChecksum
	afm.maxRetries = maxRetries
	
	logger.Info("Atomic file manager configured: sync=%v, verify=%v, retries=%d",
		syncAfterWrite, verifyChecksum, maxRetries)
}

// Start begins cleanup monitoring
func (afm *AtomicFileManager) Start() error {
	if !atomic.CompareAndSwapInt32(&afm.cleanupRunning, 0, 1) {
		return fmt.Errorf("atomic file manager already running")
	}
	
	go afm.cleanupLoop()
	logger.Info("Atomic file manager started with cleanup interval %v", afm.cleanupInterval)
	return nil
}

// Stop gracefully shuts down the manager
func (afm *AtomicFileManager) Stop() error {
	if !atomic.CompareAndSwapInt32(&afm.cleanupRunning, 1, 0) {
		return fmt.Errorf("atomic file manager not running")
	}
	
	close(afm.stopCleanup)
	
	// Wait for active operations to complete
	afm.waitForActiveOperations()
	
	logger.Info("Atomic file manager stopped")
	return nil
}

// cleanupLoop periodically cleans up temporary files
func (afm *AtomicFileManager) cleanupLoop() {
	ticker := time.NewTicker(afm.cleanupInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			afm.cleanupTempFiles()
		case <-afm.stopCleanup:
			return
		}
	}
}

// AtomicWrite atomically writes data to a file
func (afm *AtomicFileManager) AtomicWrite(targetPath string, data []byte) error {
	op := afm.createOperation(targetPath, AtomicOpWrite, data)
	defer afm.removeOperation(op.ID)
	
	for attempt := 0; attempt <= afm.maxRetries; attempt++ {
		if attempt > 0 {
			logger.Debug("Retrying atomic write operation %s (attempt %d/%d)", 
				op.ID, attempt+1, afm.maxRetries+1)
			time.Sleep(afm.retryDelay)
		}
		
		err := afm.performWrite(op)
		if err == nil {
			atomic.AddInt64(&afm.successCount, 1)
			logger.Trace("Atomic write completed successfully: %s", targetPath)
			return nil
		}
		
		logger.Warn("Atomic write attempt %d failed for %s: %v", attempt+1, targetPath, err)
		op.Retries = attempt + 1
		
		// Clean up failed attempt
		afm.cleanupOperation(op)
	}
	
	atomic.AddInt64(&afm.failureCount, 1)
	return fmt.Errorf("atomic write failed after %d attempts", afm.maxRetries+1)
}

// AtomicUpdate atomically updates a file with backup
func (afm *AtomicFileManager) AtomicUpdate(targetPath string, data []byte) error {
	op := afm.createOperation(targetPath, AtomicOpUpdate, data)
	defer afm.removeOperation(op.ID)
	
	for attempt := 0; attempt <= afm.maxRetries; attempt++ {
		if attempt > 0 {
			logger.Debug("Retrying atomic update operation %s (attempt %d/%d)", 
				op.ID, attempt+1, afm.maxRetries+1)
			time.Sleep(afm.retryDelay)
		}
		
		err := afm.performUpdate(op)
		if err == nil {
			atomic.AddInt64(&afm.successCount, 1)
			logger.Trace("Atomic update completed successfully: %s", targetPath)
			return nil
		}
		
		logger.Warn("Atomic update attempt %d failed for %s: %v", attempt+1, targetPath, err)
		op.Retries = attempt + 1
		
		// Clean up failed attempt
		afm.cleanupOperation(op)
	}
	
	atomic.AddInt64(&afm.failureCount, 1)
	return fmt.Errorf("atomic update failed after %d attempts", afm.maxRetries+1)
}

// AtomicMove atomically moves a file
func (afm *AtomicFileManager) AtomicMove(sourcePath, targetPath string) error {
	// Read source file
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}
	
	op := afm.createOperation(targetPath, AtomicOpMove, data)
	op.BackupPath = sourcePath // Source becomes backup
	defer afm.removeOperation(op.ID)
	
	for attempt := 0; attempt <= afm.maxRetries; attempt++ {
		if attempt > 0 {
			logger.Debug("Retrying atomic move operation %s (attempt %d/%d)", 
				op.ID, attempt+1, afm.maxRetries+1)
			time.Sleep(afm.retryDelay)
		}
		
		err := afm.performMove(op, sourcePath)
		if err == nil {
			atomic.AddInt64(&afm.successCount, 1)
			logger.Trace("Atomic move completed successfully: %s -> %s", sourcePath, targetPath)
			return nil
		}
		
		logger.Warn("Atomic move attempt %d failed: %v", attempt+1, err)
		op.Retries = attempt + 1
	}
	
	atomic.AddInt64(&afm.failureCount, 1)
	return fmt.Errorf("atomic move failed after %d attempts", afm.maxRetries+1)
}

// createOperation creates a new atomic operation
func (afm *AtomicFileManager) createOperation(targetPath string, opType AtomicOperationType, data []byte) *AtomicOperation {
	opID := afm.generateOperationID()
	
	op := &AtomicOperation{
		ID:         opID,
		TargetPath: targetPath,
		TempPath:   targetPath + afm.tempSuffix + "." + opID,
		BackupPath: targetPath + afm.backupSuffix + "." + opID,
		Operation:  opType,
		StartTime:  time.Now(),
		Data:       data,
	}
	
	// Calculate checksum if verification enabled
	if afm.verifyChecksum {
		hash := sha256.Sum256(data)
		op.Checksum = hex.EncodeToString(hash[:])
	}
	
	afm.mu.Lock()
	afm.activeOperations[opID] = op
	atomic.AddInt64(&afm.operationCount, 1)
	afm.mu.Unlock()
	
	logger.Trace("Created atomic operation %s for %s (%s)", opID, targetPath, opType)
	return op
}

// performWrite performs the actual write operation
func (afm *AtomicFileManager) performWrite(op *AtomicOperation) error {
	op.mu.Lock()
	defer op.mu.Unlock()
	
	// Create directory if it doesn't exist
	dir := filepath.Dir(op.TargetPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Write to temporary file
	file, err := os.OpenFile(op.TempPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	
	_, err = file.Write(op.Data)
	if err != nil {
		file.Close()
		os.Remove(op.TempPath)
		return fmt.Errorf("failed to write data: %w", err)
	}
	
	// Sync if configured
	if afm.syncAfterWrite {
		if err := file.Sync(); err != nil {
			file.Close()
			os.Remove(op.TempPath)
			return fmt.Errorf("failed to sync: %w", err)
		}
	}
	
	file.Close()
	
	// Verify checksum if configured
	if afm.verifyChecksum {
		if err := afm.verifyFileChecksum(op.TempPath, op.Checksum); err != nil {
			os.Remove(op.TempPath)
			return fmt.Errorf("checksum verification failed: %w", err)
		}
	}
	
	// Atomic rename
	if err := os.Rename(op.TempPath, op.TargetPath); err != nil {
		os.Remove(op.TempPath)
		return fmt.Errorf("failed to rename temp file: %w", err)
	}
	
	return nil
}

// performUpdate performs atomic update with backup
func (afm *AtomicFileManager) performUpdate(op *AtomicOperation) error {
	op.mu.Lock()
	defer op.mu.Unlock()
	
	// Create backup if target exists
	if _, err := os.Stat(op.TargetPath); err == nil {
		if err := afm.createBackup(op); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}
	
	// Perform write
	if err := afm.performWrite(op); err != nil {
		// Restore from backup if write failed
		afm.restoreFromBackup(op)
		return err
	}
	
	// Remove backup on success
	os.Remove(op.BackupPath)
	return nil
}

// performMove performs atomic move operation
func (afm *AtomicFileManager) performMove(op *AtomicOperation, sourcePath string) error {
	op.mu.Lock()
	defer op.mu.Unlock()
	
	// First write to temp location
	if err := afm.performWrite(op); err != nil {
		return fmt.Errorf("failed to write to target: %w", err)
	}
	
	// Remove source file
	if err := os.Remove(sourcePath); err != nil {
		// If we can't remove source, remove the target to maintain consistency
		os.Remove(op.TargetPath)
		return fmt.Errorf("failed to remove source: %w", err)
	}
	
	return nil
}

// createBackup creates a backup of the target file
func (afm *AtomicFileManager) createBackup(op *AtomicOperation) error {
	source, err := os.Open(op.TargetPath)
	if err != nil {
		return err
	}
	defer source.Close()
	
	backup, err := os.Create(op.BackupPath)
	if err != nil {
		return err
	}
	defer backup.Close()
	
	_, err = io.Copy(backup, source)
	if err != nil {
		os.Remove(op.BackupPath)
		return err
	}
	
	if afm.syncAfterWrite {
		return backup.Sync()
	}
	
	return nil
}

// restoreFromBackup restores the original file from backup
func (afm *AtomicFileManager) restoreFromBackup(op *AtomicOperation) {
	if _, err := os.Stat(op.BackupPath); os.IsNotExist(err) {
		return // No backup to restore
	}
	
	if err := os.Rename(op.BackupPath, op.TargetPath); err != nil {
		logger.Error("Failed to restore from backup %s: %v", op.BackupPath, err)
	} else {
		logger.Info("Successfully restored %s from backup", op.TargetPath)
	}
}

// verifyFileChecksum verifies the checksum of a file
func (afm *AtomicFileManager) verifyFileChecksum(filePath, expectedChecksum string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}
	
	actualChecksum := hex.EncodeToString(hash.Sum(nil))
	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}
	
	return nil
}

// generateOperationID generates a unique operation ID
func (afm *AtomicFileManager) generateOperationID() string {
	count := atomic.LoadInt64(&afm.operationCount)
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), count)
}

// removeOperation removes an operation from active tracking
func (afm *AtomicFileManager) removeOperation(opID string) {
	afm.mu.Lock()
	delete(afm.activeOperations, opID)
	afm.mu.Unlock()
}

// cleanupOperation cleans up temporary files for a failed operation
func (afm *AtomicFileManager) cleanupOperation(op *AtomicOperation) {
	os.Remove(op.TempPath)
	os.Remove(op.BackupPath)
}

// cleanupTempFiles removes stale temporary files
func (afm *AtomicFileManager) cleanupTempFiles() {
	afm.mu.Lock()
	defer afm.mu.Unlock()
	
	cutoff := time.Now().Add(-time.Hour) // Remove temp files older than 1 hour
	cleaned := 0
	
	for opID, op := range afm.activeOperations {
		if op.StartTime.Before(cutoff) {
			logger.Debug("Cleaning up stale operation %s (started: %v)", opID, op.StartTime)
			afm.cleanupOperation(op)
			delete(afm.activeOperations, opID)
			cleaned++
		}
	}
	
	afm.lastCleanup = time.Now()
	if cleaned > 0 {
		logger.Info("Cleaned up %d stale atomic operations", cleaned)
	}
}

// waitForActiveOperations waits for all active operations to complete
func (afm *AtomicFileManager) waitForActiveOperations() {
	maxWait := 30 * time.Second
	start := time.Now()
	
	for {
		afm.mu.RLock()
		activeCount := len(afm.activeOperations)
		afm.mu.RUnlock()
		
		if activeCount == 0 {
			break
		}
		
		if time.Since(start) > maxWait {
			logger.Warn("Timeout waiting for %d active operations to complete", activeCount)
			break
		}
		
		time.Sleep(100 * time.Millisecond)
	}
}

// GetStatistics returns operation statistics
func (afm *AtomicFileManager) GetStatistics() map[string]interface{} {
	afm.mu.RLock()
	defer afm.mu.RUnlock()
	
	return map[string]interface{}{
		"cleanup_running":        atomic.LoadInt32(&afm.cleanupRunning) == 1,
		"active_operations":      len(afm.activeOperations),
		"total_operations":       atomic.LoadInt64(&afm.operationCount),
		"successful_operations":  atomic.LoadInt64(&afm.successCount),
		"failed_operations":      atomic.LoadInt64(&afm.failureCount),
		"success_rate":           float64(atomic.LoadInt64(&afm.successCount)) / float64(atomic.LoadInt64(&afm.operationCount)),
		"sync_after_write":       afm.syncAfterWrite,
		"verify_checksum":        afm.verifyChecksum,
		"max_retries":            afm.maxRetries,
		"last_cleanup":           afm.lastCleanup.Format(time.RFC3339),
	}
}