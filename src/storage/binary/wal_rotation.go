// Package binary provides WAL rotation and bounded growth for corruption prevention
//
// This system ensures the WAL doesn't grow unbounded which can lead to:
// - Disk space exhaustion
// - Long recovery times
// - Performance degradation
// - Memory pressure during replay
package binary

import (
	"entitydb/logger"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// WALRotationManager manages WAL size limits and automatic rotation
type WALRotationManager struct {
	wal              *WAL
	mu               sync.RWMutex
	
	// Configuration
	maxSizeBytes     int64         // Maximum WAL size before rotation
	maxAgeMinutes    int64         // Maximum WAL age before rotation
	checkInterval    time.Duration // How often to check for rotation
	
	// State
	running          int32
	stopChan         chan struct{}
	rotationCount    int64
	lastRotationTime time.Time
	
	// Statistics
	totalBytesRotated int64
	totalRotations    int64
	lastCheckTime     time.Time
	
	// Callbacks
	preRotationCallback  func() error
	postRotationCallback func() error
}

// NewWALRotationManager creates a new WAL rotation manager
func NewWALRotationManager(wal *WAL) *WALRotationManager {
	return &WALRotationManager{
		wal:               wal,
		maxSizeBytes:      100 * 1024 * 1024, // 100MB default
		maxAgeMinutes:     60,                 // 60 minutes default
		checkInterval:     5 * time.Minute,    // Check every 5 minutes
		stopChan:          make(chan struct{}),
		lastRotationTime:  time.Now(),
		lastCheckTime:     time.Now(),
	}
}

// Configure sets rotation parameters
func (wrm *WALRotationManager) Configure(maxSizeMB int64, maxAgeMinutes int64, checkIntervalMinutes int64) {
	wrm.mu.Lock()
	defer wrm.mu.Unlock()
	
	if maxSizeMB > 0 {
		wrm.maxSizeBytes = maxSizeMB * 1024 * 1024
	}
	if maxAgeMinutes > 0 {
		wrm.maxAgeMinutes = maxAgeMinutes
	}
	if checkIntervalMinutes > 0 {
		wrm.checkInterval = time.Duration(checkIntervalMinutes) * time.Minute
	}
	
	logger.Info("WAL rotation configured: max size %dMB, max age %dm, check interval %dm",
		wrm.maxSizeBytes/(1024*1024), wrm.maxAgeMinutes, checkIntervalMinutes)
}

// SetCallbacks configures pre and post rotation callbacks
func (wrm *WALRotationManager) SetCallbacks(preRotation, postRotation func() error) {
	wrm.mu.Lock()
	defer wrm.mu.Unlock()
	
	wrm.preRotationCallback = preRotation
	wrm.postRotationCallback = postRotation
}

// Start begins the rotation monitoring
func (wrm *WALRotationManager) Start() error {
	if !atomic.CompareAndSwapInt32(&wrm.running, 0, 1) {
		return fmt.Errorf("WAL rotation manager already running")
	}
	
	go wrm.monitorLoop()
	logger.Info("WAL rotation manager started (max: %dMB, age: %dm)",
		wrm.maxSizeBytes/(1024*1024), wrm.maxAgeMinutes)
	return nil
}

// Stop gracefully shuts down the rotation manager
func (wrm *WALRotationManager) Stop() error {
	if !atomic.CompareAndSwapInt32(&wrm.running, 1, 0) {
		return fmt.Errorf("WAL rotation manager not running")
	}
	
	close(wrm.stopChan)
	logger.Info("WAL rotation manager stopped")
	return nil
}

// monitorLoop is the main monitoring goroutine
func (wrm *WALRotationManager) monitorLoop() {
	ticker := time.NewTicker(wrm.checkInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			wrm.checkAndRotate()
		case <-wrm.stopChan:
			return
		}
	}
}

// checkAndRotate checks if rotation is needed and performs it
func (wrm *WALRotationManager) checkAndRotate() {
	wrm.mu.Lock()
	wrm.lastCheckTime = time.Now()
	wrm.mu.Unlock()
	
	needsRotation, reason := wrm.needsRotation()
	if !needsRotation {
		logger.Trace("WAL rotation check: no rotation needed")
		return
	}
	
	logger.Info("WAL rotation triggered: %s", reason)
	if err := wrm.performRotation(); err != nil {
		logger.Error("WAL rotation failed: %v", err)
	} else {
		atomic.AddInt64(&wrm.totalRotations, 1)
		wrm.mu.Lock()
		wrm.lastRotationTime = time.Now()
		wrm.mu.Unlock()
		logger.Info("WAL rotation completed successfully")
	}
}

// needsRotation determines if rotation is required
func (wrm *WALRotationManager) needsRotation() (bool, string) {
	wrm.mu.RLock()
	defer wrm.mu.RUnlock()
	
	// Check file size
	if wrm.wal != nil {
		currentSize, err := wrm.getWALSize()
		if err != nil {
			logger.Warn("Failed to get WAL size: %v", err)
		} else if currentSize > wrm.maxSizeBytes {
			return true, fmt.Sprintf("size limit exceeded (%d > %d bytes)", currentSize, wrm.maxSizeBytes)
		}
	}
	
	// Check age
	age := time.Since(wrm.lastRotationTime)
	maxAge := time.Duration(wrm.maxAgeMinutes) * time.Minute
	if age > maxAge {
		return true, fmt.Sprintf("age limit exceeded (%v > %v)", age, maxAge)
	}
	
	return false, ""
}

// getWALSize returns the current WAL size in bytes
func (wrm *WALRotationManager) getWALSize() (int64, error) {
	if wrm.wal == nil {
		return 0, fmt.Errorf("WAL not initialized")
	}
	
	// For unified files, we need to calculate WAL section size
	if wrm.wal.isUnified {
		return int64(wrm.wal.walSize), nil
	}
	
	// For standalone WAL files
	if wrm.wal.file == nil {
		return 0, fmt.Errorf("WAL file not open")
	}
	
	stat, err := wrm.wal.file.Stat()
	if err != nil {
		return 0, err
	}
	
	return stat.Size(), nil
}

// performRotation executes the WAL rotation process
func (wrm *WALRotationManager) performRotation() error {
	logger.Debug("Starting WAL rotation process")
	
	// Pre-rotation callback (e.g., checkpoint)
	if wrm.preRotationCallback != nil {
		logger.Debug("Executing pre-rotation callback")
		if err := wrm.preRotationCallback(); err != nil {
			return fmt.Errorf("pre-rotation callback failed: %w", err)
		}
	}
	
	// Get current size for statistics
	oldSize, _ := wrm.getWALSize()
	
	// Perform the rotation
	if err := wrm.rotateWAL(); err != nil {
		return fmt.Errorf("WAL rotation failed: %w", err)
	}
	
	// Update statistics
	atomic.AddInt64(&wrm.totalBytesRotated, oldSize)
	
	// Post-rotation callback
	if wrm.postRotationCallback != nil {
		logger.Debug("Executing post-rotation callback")
		if err := wrm.postRotationCallback(); err != nil {
			logger.Warn("Post-rotation callback failed: %v", err)
			// Don't fail the rotation for post-callback errors
		}
	}
	
	logger.Debug("WAL rotation process completed")
	return nil
}

// rotateWAL performs the actual WAL rotation
func (wrm *WALRotationManager) rotateWAL() error {
	if wrm.wal == nil {
		return fmt.Errorf("WAL not initialized")
	}
	
	// For unified files, we truncate the WAL section
	if wrm.wal.isUnified {
		return wrm.truncateUnifiedWAL()
	}
	
	// For standalone files, we can create a new file
	return wrm.rotateStandaloneWAL()
}

// truncateUnifiedWAL truncates the WAL section in a unified file
func (wrm *WALRotationManager) truncateUnifiedWAL() error {
	logger.Debug("Truncating unified WAL section")
	
	// For unified files, we reset the WAL section by seeking to the beginning
	// and updating the header to indicate zero WAL size
	if wrm.wal.file == nil {
		return fmt.Errorf("unified WAL file not open")
	}
	
	// Seek to WAL section start
	if _, err := wrm.wal.file.Seek(int64(wrm.wal.walOffset), 0); err != nil {
		return fmt.Errorf("failed to seek to WAL section: %w", err)
	}
	
	// Reset WAL size in the WAL struct
	wrm.wal.mu.Lock()
	wrm.wal.walSize = 0
	wrm.wal.sequence = 0
	wrm.wal.mu.Unlock()
	
	// Sync the changes
	if err := wrm.wal.file.Sync(); err != nil {
		return fmt.Errorf("failed to sync after WAL truncation: %w", err)
	}
	
	logger.Debug("Unified WAL section truncated successfully")
	return nil
}

// rotateStandaloneWAL creates a new standalone WAL file
func (wrm *WALRotationManager) rotateStandaloneWAL() error {
	logger.Debug("Rotating standalone WAL file")
	
	// Close current file
	if wrm.wal.file != nil {
		if err := wrm.wal.file.Close(); err != nil {
			logger.Warn("Failed to close old WAL file: %v", err)
		}
	}
	
	// Create backup of old file
	backupPath := wrm.wal.path + ".rotated." + time.Now().Format("20060102-150405")
	if err := os.Rename(wrm.wal.path, backupPath); err != nil {
		logger.Warn("Failed to backup old WAL file: %v", err)
		// Continue with rotation even if backup fails
	} else {
		logger.Debug("Old WAL file backed up to: %s", backupPath)
	}
	
	// Create new WAL file
	file, err := os.OpenFile(wrm.wal.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to create new WAL file: %w", err)
	}
	
	// Update WAL instance
	wrm.wal.mu.Lock()
	wrm.wal.file = file
	wrm.wal.sequence = 0
	wrm.wal.mu.Unlock()
	
	logger.Debug("New standalone WAL file created")
	return nil
}

// GetStatistics returns rotation statistics
func (wrm *WALRotationManager) GetStatistics() map[string]interface{} {
	wrm.mu.RLock()
	defer wrm.mu.RUnlock()
	
	currentSize, _ := wrm.getWALSize()
	
	return map[string]interface{}{
		"running":               atomic.LoadInt32(&wrm.running) == 1,
		"current_size_bytes":    currentSize,
		"current_size_mb":       float64(currentSize) / (1024 * 1024),
		"max_size_bytes":        wrm.maxSizeBytes,
		"max_size_mb":           float64(wrm.maxSizeBytes) / (1024 * 1024),
		"max_age_minutes":       wrm.maxAgeMinutes,
		"total_rotations":       atomic.LoadInt64(&wrm.totalRotations),
		"total_bytes_rotated":   atomic.LoadInt64(&wrm.totalBytesRotated),
		"last_rotation_time":    wrm.lastRotationTime.Format(time.RFC3339),
		"last_check_time":       wrm.lastCheckTime.Format(time.RFC3339),
		"next_check_in":         wrm.checkInterval - time.Since(wrm.lastCheckTime),
	}
}

// ForceRotation manually triggers a WAL rotation
func (wrm *WALRotationManager) ForceRotation() error {
	if atomic.LoadInt32(&wrm.running) == 0 {
		return fmt.Errorf("WAL rotation manager not running")
	}
	
	logger.Info("Manual WAL rotation triggered")
	return wrm.performRotation()
}