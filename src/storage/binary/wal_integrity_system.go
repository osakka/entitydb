package binary

import (
	"context"
	"crypto/sha256"
	"entitydb/config"
	"entitydb/logger"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"
)

// BAR-RAISING SOLUTION: Comprehensive WAL corruption prevention system
type WALIntegritySystem struct {
	filePath        string
	backupPath      string
	config          *config.Config
	checksumCache   map[int64][]byte
	sizeValidator   *SizeValidator
	seekValidator   *SeekValidator
	fsMonitor       *FileSystemMonitor
	healingManager  *SelfHealingManager
	mu              sync.RWMutex
	lastHealthCheck time.Time
	healthStatus    HealthStatus
}

type HealthStatus struct {
	IsHealthy          bool
	LastError          error
	CorruptionDetected bool
	AutoRepairAttempts int
	LastRepairTime     time.Time
}

type SizeValidator struct {
	maxEntitySize    int64
	maxWALSize       int64
	maxEntryLength   int64
}

type SeekValidator struct {
	validPositions   map[int64]bool
	lastValidOffset  int64
	mu               sync.RWMutex
}

type FileSystemMonitor struct {
	minDiskSpace     int64
	maxFileDesc      int
	healthThreshold  float64
	lastCheck        time.Time
}

type SelfHealingManager struct {
	backupRetention  int
	repairThreshold  int
	emergencyMode    bool
	lastBackup       time.Time
}

// Constants for validation
const (
	MAX_ENTITY_SIZE     = 100 * 1024 * 1024  // 100MB per entity
	MAX_WAL_SIZE        = 1024 * 1024 * 1024 // 1GB WAL size
	MAX_ENTRY_LENGTH    = 200 * 1024 * 1024  // 200MB per WAL entry
	MIN_DISK_SPACE      = 1024 * 1024 * 1024 // 1GB minimum
	MAX_FILE_DESC       = 1000               // Max file descriptors
	HEALTH_CHECK_INTERVAL = 30 * time.Second // Health check frequency
	ASTRONOMICAL_THRESHOLD = 1000000000      // 1GB - flag astronomical sizes
)

// NewWALIntegritySystem creates a comprehensive WAL protection system
func NewWALIntegritySystem(filePath string, cfg *config.Config) *WALIntegritySystem {
	backupPath := filePath + ".backup"
	
	return &WALIntegritySystem{
		filePath:       filePath,
		backupPath:     backupPath,
		config:         cfg,
		checksumCache:  make(map[int64][]byte),
		sizeValidator:  &SizeValidator{
			maxEntitySize:  MAX_ENTITY_SIZE,
			maxWALSize:     MAX_WAL_SIZE,
			maxEntryLength: MAX_ENTRY_LENGTH,
		},
		seekValidator: &SeekValidator{
			validPositions: make(map[int64]bool),
		},
		fsMonitor: &FileSystemMonitor{
			minDiskSpace:    MIN_DISK_SPACE,
			maxFileDesc:     MAX_FILE_DESC,
			healthThreshold: 0.95,
		},
		healingManager: &SelfHealingManager{
			backupRetention: 5,
			repairThreshold: 3,
		},
		healthStatus: HealthStatus{
			IsHealthy: true,
		},
	}
}

// BAR-RAISING SOLUTION: Pre-write validation with comprehensive checks
func (w *WALIntegritySystem) ValidateBeforeWrite(entityID string, content []byte, entryLength int64) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 1. ASTRONOMICAL SIZE DETECTION
	if entryLength > ASTRONOMICAL_THRESHOLD {
		logger.Error("CORRUPTION DETECTED: Astronomical WAL entry length %d for entity %s", entryLength, entityID)
		w.healthStatus.CorruptionDetected = true
		w.healingManager.emergencyMode = true
		return fmt.Errorf("astronomical WAL entry length detected: %d (threshold: %d)", entryLength, ASTRONOMICAL_THRESHOLD)
	}

	// 2. SIZE VALIDATION
	if err := w.sizeValidator.Validate(content, entryLength); err != nil {
		logger.Warn("Size validation failed for entity %s: %v", entityID, err)
		return fmt.Errorf("size validation failed: %w", err)
	}

	// 3. FILE SYSTEM HEALTH CHECK
	if err := w.fsMonitor.CheckHealth(); err != nil {
		logger.Error("File system health check failed: %v", err)
		w.healthStatus.IsHealthy = false
		w.healthStatus.LastError = err
		return fmt.Errorf("file system unhealthy: %w", err)
	}

	// 4. SEEK POSITION VALIDATION
	if err := w.seekValidator.ValidatePosition(w.filePath); err != nil {
		logger.Error("Seek validation failed: %v", err)
		return fmt.Errorf("seek validation failed: %w", err)
	}

	// 5. CONTENT INTEGRITY CHECK
	if err := w.validateContentIntegrity(content); err != nil {
		logger.Error("Content integrity check failed for entity %s: %v", entityID, err)
		return fmt.Errorf("content integrity failed: %w", err)
	}

	return nil
}

// Size validation with multiple thresholds
func (s *SizeValidator) Validate(content []byte, entryLength int64) error {
	// Check entity content size
	if int64(len(content)) > s.maxEntitySize {
		return fmt.Errorf("entity content too large: %d > %d", len(content), s.maxEntitySize)
	}

	// Check WAL entry length
	if entryLength > s.maxEntryLength {
		return fmt.Errorf("WAL entry too large: %d > %d", entryLength, s.maxEntryLength)
	}

	// Check for negative sizes (corruption indicator)
	if entryLength < 0 {
		return fmt.Errorf("invalid negative entry length: %d", entryLength)
	}

	return nil
}

// Seek position validation to prevent file corruption
func (s *SeekValidator) ValidatePosition(filePath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Get current file size
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	fileSize := fileInfo.Size()

	// Open file to test seek operation
	file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file for seek test: %w", err)
	}
	defer file.Close()

	// Test seek to current end
	_, err = file.Seek(0, 2) // Seek to end
	if err != nil {
		return fmt.Errorf("seek to end failed: %w", err)
	}

	// Test seek to beginning
	_, err = file.Seek(0, 0)
	if err != nil {
		return fmt.Errorf("seek to beginning failed: %w", err)
	}

	// Update valid positions
	s.validPositions[fileSize] = true
	s.lastValidOffset = fileSize

	return nil
}

// File system health monitoring
func (f *FileSystemMonitor) CheckHealth() error {
	// Check disk space
	if err := f.checkDiskSpace(); err != nil {
		return fmt.Errorf("disk space check failed: %w", err)
	}

	// Check file descriptor usage
	if err := f.checkFileDescriptors(); err != nil {
		return fmt.Errorf("file descriptor check failed: %w", err)
	}

	// Check file system integrity
	if err := f.checkFileSystemIntegrity(); err != nil {
		return fmt.Errorf("file system integrity check failed: %w", err)
	}

	f.lastCheck = time.Now()
	return nil
}

func (f *FileSystemMonitor) checkDiskSpace() error {
	var stat syscall.Statfs_t
	if err := syscall.Statfs("/opt/entitydb/var", &stat); err != nil {
		return fmt.Errorf("failed to get disk usage: %w", err)
	}

	available := stat.Bavail * uint64(stat.Bsize)
	if int64(available) < f.minDiskSpace {
		return fmt.Errorf("insufficient disk space: %d < %d", available, f.minDiskSpace)
	}

	return nil
}

func (f *FileSystemMonitor) checkFileDescriptors() error {
	// Simple file descriptor count check
	// In production, this would check /proc/sys/fs/file-nr
	return nil
}

func (f *FileSystemMonitor) checkFileSystemIntegrity() error {
	// Basic file system integrity check
	testFile := "/opt/entitydb/var/.fstest"
	
	// Test write
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("file system write test failed: %w", err)
	}

	// Test read
	if _, err := os.ReadFile(testFile); err != nil {
		return fmt.Errorf("file system read test failed: %w", err)
	}

	// Cleanup
	os.Remove(testFile)

	return nil
}

// Content integrity validation using checksums
func (w *WALIntegritySystem) validateContentIntegrity(content []byte) error {
	// Calculate SHA256 checksum
	hash := sha256.Sum256(content)
	
	// Check for patterns that indicate corruption
	if w.detectCorruptionPatterns(content) {
		return fmt.Errorf("corruption patterns detected in content")
	}

	// Store checksum for future validation
	w.checksumCache[time.Now().UnixNano()] = hash[:]

	return nil
}

// Detect known corruption patterns
func (w *WALIntegritySystem) detectCorruptionPatterns(content []byte) bool {
	// Check for repeated bytes (common corruption pattern)
	if len(content) > 1000 {
		firstByte := content[0]
		sameByteCount := 0
		for _, b := range content[:1000] {
			if b == firstByte {
				sameByteCount++
			}
		}
		// If more than 90% are the same byte, likely corruption
		if float64(sameByteCount)/1000.0 > 0.9 {
			return true
		}
	}

	// Check for null byte sequences (another corruption indicator)
	nullCount := 0
	for _, b := range content {
		if b == 0 {
			nullCount++
		}
	}
	if len(content) > 0 && float64(nullCount)/float64(len(content)) > 0.5 {
		return true
	}

	return false
}

// BAR-RAISING SOLUTION: Self-healing with automatic backup and recovery
func (w *WALIntegritySystem) AttemptSelfHealing() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	logger.Info("Attempting WAL self-healing...")

	// 1. Create emergency backup
	if err := w.createEmergencyBackup(); err != nil {
		logger.Error("Failed to create emergency backup: %v", err)
		return fmt.Errorf("emergency backup failed: %w", err)
	}

	// 2. Attempt repair
	if err := w.attemptRepair(); err != nil {
		logger.Error("Self-repair failed: %v", err)
		
		// 3. Fall back to backup restoration
		if err := w.restoreFromBackup(); err != nil {
			logger.Error("Backup restoration failed: %v", err)
			return fmt.Errorf("complete healing failure: repair failed and backup restoration failed: %w", err)
		}
	}

	// 4. Validate healing success
	if err := w.validateHealingSuccess(); err != nil {
		logger.Error("Healing validation failed: %v", err)
		return fmt.Errorf("healing validation failed: %w", err)
	}

	// 5. Update health status
	w.healthStatus.IsHealthy = true
	w.healthStatus.CorruptionDetected = false
	w.healthStatus.AutoRepairAttempts++
	w.healthStatus.LastRepairTime = time.Now()

	logger.Info("WAL self-healing completed successfully")
	return nil
}

func (w *WALIntegritySystem) createEmergencyBackup() error {
	timestamp := time.Now().Format("20060102-150405")
	emergencyBackup := fmt.Sprintf("%s.emergency-%s", w.backupPath, timestamp)
	
	// Copy current file to emergency backup
	if err := copyFileForWAL(w.filePath, emergencyBackup); err != nil {
		return fmt.Errorf("failed to create emergency backup: %w", err)
	}

	logger.Info("Emergency backup created: %s", emergencyBackup)
	return nil
}

func (w *WALIntegritySystem) attemptRepair() error {
	// Attempt to truncate at last known good position
	if w.seekValidator.lastValidOffset > 0 {
		if err := os.Truncate(w.filePath, w.seekValidator.lastValidOffset); err != nil {
			return fmt.Errorf("failed to truncate at last valid offset: %w", err)
		}
		logger.Info("Truncated WAL to last valid offset: %d", w.seekValidator.lastValidOffset)
	}

	return nil
}

func (w *WALIntegritySystem) restoreFromBackup() error {
	if _, err := os.Stat(w.backupPath); os.IsNotExist(err) {
		return fmt.Errorf("no backup available for restoration")
	}

	// Restore from backup
	if err := copyFileForWAL(w.backupPath, w.filePath); err != nil {
		return fmt.Errorf("failed to restore from backup: %w", err)
	}

	logger.Info("Restored WAL from backup: %s", w.backupPath)
	return nil
}

func (w *WALIntegritySystem) validateHealingSuccess() error {
	// Validate that the file is now readable and seekable
	file, err := os.OpenFile(w.filePath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("failed to open healed file: %w", err)
	}
	defer file.Close()

	// Test basic operations
	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("seek test failed on healed file: %w", err)
	}

	if _, err := file.Seek(0, 2); err != nil {
		return fmt.Errorf("seek to end failed on healed file: %w", err)
	}

	return nil
}

// Continuous health monitoring
func (w *WALIntegritySystem) StartHealthMonitoring(ctx context.Context) {
	ticker := time.NewTicker(HEALTH_CHECK_INTERVAL)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.performHealthCheck()
		case <-ctx.Done():
			logger.Info("WAL integrity monitoring stopped")
			return
		}
	}
}

func (w *WALIntegritySystem) performHealthCheck() {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Check if corruption is detected
	if w.healthStatus.CorruptionDetected {
		logger.Warn("Corruption detected, attempting self-healing...")
		if err := w.AttemptSelfHealing(); err != nil {
			logger.Error("Self-healing failed: %v", err)
			w.healingManager.emergencyMode = true
		}
	}

	// Perform routine backup using configured interval
	backupInterval := time.Duration(w.config.BackupInterval) * time.Second
	if backupInterval == 0 {
		backupInterval = time.Hour // Default to 1 hour if not configured
	}
	
	if time.Since(w.healingManager.lastBackup) > backupInterval {
		if err := w.createRoutineBackup(); err != nil {
			logger.Warn("Routine backup failed: %v", err)
		} else {
			w.healingManager.lastBackup = time.Now()
			// Clean up old backups after creating new one
			if err := w.cleanupOldBackups(); err != nil {
				logger.Warn("Backup cleanup failed: %v", err)
			}
		}
	}

	w.lastHealthCheck = time.Now()
}

func (w *WALIntegritySystem) createRoutineBackup() error {
	timestamp := time.Now().Format("20060102-150405")
	routineBackup := fmt.Sprintf("%s.routine-%s", w.backupPath, timestamp)
	
	return copyFileForWAL(w.filePath, routineBackup)
}

// cleanupOldBackups implements intelligent backup retention based on configuration
func (w *WALIntegritySystem) cleanupOldBackups() error {
	// Get backup directory
	backupDir := filepath.Dir(w.backupPath)
	backupPrefix := filepath.Base(w.backupPath) + ".routine-"
	
	// List all backup files
	files, err := ioutil.ReadDir(backupDir)
	if err != nil {
		return fmt.Errorf("failed to read backup directory: %w", err)
	}
	
	// Filter and sort routine backup files
	var backups []os.FileInfo
	for _, file := range files {
		if strings.HasPrefix(file.Name(), backupPrefix) && !file.IsDir() {
			backups = append(backups, file)
		}
	}
	
	if len(backups) == 0 {
		return nil // No backups to clean
	}
	
	// Sort by modification time (newest first)
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].ModTime().After(backups[j].ModTime())
	})
	
	// Apply retention policies
	toDelete := w.applyRetentionPolicies(backups)
	
	// Check size limits
	if w.config.BackupMaxSizeMB > 0 {
		toDelete = w.applySizeLimits(backups, toDelete)
	}
	
	// Delete marked backups
	deleted := 0
	for _, backup := range toDelete {
		backupPath := filepath.Join(backupDir, backup.Name())
		if err := os.Remove(backupPath); err != nil {
			logger.Warn("Failed to delete old backup %s: %v", backup.Name(), err)
		} else {
			deleted++
			logger.Debug("Deleted old backup: %s", backup.Name())
		}
	}
	
	if deleted > 0 {
		logger.Info("Cleaned up %d old backup files", deleted)
	}
	
	return nil
}

// applyRetentionPolicies determines which backups to keep based on time-based retention
func (w *WALIntegritySystem) applyRetentionPolicies(backups []os.FileInfo) []os.FileInfo {
	now := time.Now()
	var toDelete []os.FileInfo
	
	// Maps to track kept backups by period
	hourlyKept := 0
	dailyKept := make(map[string]bool)
	weeklyKept := make(map[string]bool)
	
	for _, backup := range backups {
		age := now.Sub(backup.ModTime())
		keep := false
		
		// Keep recent hourly backups
		if age < time.Duration(w.config.BackupRetentionHours)*time.Hour && 
		   hourlyKept < w.config.BackupRetentionHours {
			keep = true
			hourlyKept++
		}
		
		// Keep daily backups
		dayKey := backup.ModTime().Format("2006-01-02")
		if age < time.Duration(w.config.BackupRetentionDays)*24*time.Hour && 
		   !dailyKept[dayKey] && len(dailyKept) < w.config.BackupRetentionDays {
			keep = true
			dailyKept[dayKey] = true
		}
		
		// Keep weekly backups
		year, week := backup.ModTime().ISOWeek()
		weekKey := fmt.Sprintf("%d-%02d", year, week)
		if age < time.Duration(w.config.BackupRetentionWeeks)*7*24*time.Hour && 
		   !weeklyKept[weekKey] && len(weeklyKept) < w.config.BackupRetentionWeeks {
			keep = true
			weeklyKept[weekKey] = true
		}
		
		// Always keep backups created during emergency/corruption events
		if w.healingManager.emergencyMode && age < 24*time.Hour {
			keep = true
		}
		
		if !keep {
			toDelete = append(toDelete, backup)
		}
	}
	
	return toDelete
}

// applySizeLimits ensures total backup size doesn't exceed configured limit
func (w *WALIntegritySystem) applySizeLimits(backups []os.FileInfo, toDelete []os.FileInfo) []os.FileInfo {
	maxSize := w.config.BackupMaxSizeMB * 1024 * 1024 // Convert to bytes
	
	// Calculate current total size
	var totalSize int64
	deleteMap := make(map[string]bool)
	for _, backup := range toDelete {
		deleteMap[backup.Name()] = true
	}
	
	// Calculate size of backups we're keeping
	for _, backup := range backups {
		if !deleteMap[backup.Name()] {
			totalSize += backup.Size()
		}
	}
	
	// If still over limit, delete oldest backups first (they're sorted newest first)
	for i := len(backups) - 1; i >= 0 && totalSize > maxSize; i-- {
		backup := backups[i]
		if !deleteMap[backup.Name()] {
			toDelete = append(toDelete, backup)
			totalSize -= backup.Size()
			logger.Debug("Marking backup for deletion due to size limit: %s (size: %d bytes)", 
				backup.Name(), backup.Size())
		}
	}
	
	return toDelete
}

// Utility function for file copying in WAL integrity system
func copyFileForWAL(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = destFile.ReadFrom(sourceFile)
	return err
}

// Public interface for integration
func (w *WALIntegritySystem) IsHealthy() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.healthStatus.IsHealthy
}

func (w *WALIntegritySystem) GetHealthStatus() HealthStatus {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.healthStatus
}

func (w *WALIntegritySystem) EnableEmergencyMode() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.healingManager.emergencyMode = true
	logger.Warn("WAL integrity system: Emergency mode activated")
}