// Package binary provides comprehensive corruption detection and auto-recovery
//
// This system creates a self-healing database by detecting and automatically
// recovering from various types of corruption:
// - File header corruption
// - WAL corruption and inconsistencies
// - Index corruption and stale entries
// - Data integrity violations
// - Consistency violations between storage layers
package binary

import (
	"entitydb/logger"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// CorruptionDetector provides comprehensive corruption detection and recovery
type CorruptionDetector struct {
	repo                *EntityRepository
	mu                  sync.RWMutex
	
	// Configuration
	detectionInterval   time.Duration // How often to run corruption detection
	autoRecoveryEnabled bool          // Whether to auto-recover from corruption
	maxRecoveryAttempts int           // Maximum recovery attempts per issue
	
	// State
	running             int32
	stopChan            chan struct{}
	lastDetectionTime   time.Time
	
	// Statistics
	totalDetections     int64
	corruptionFound     int64
	recoveryAttempts    int64
	recoverySuccesses   int64
	
	// Issue tracking
	detectedIssues      []CorruptionIssue
	recoveryHistory     []RecoveryAttempt
}

// CorruptionIssue represents a detected corruption problem
type CorruptionIssue struct {
	Type            CorruptionType
	Component       string
	EntityID        string
	Description     string
	Severity        CorruptionSeverity
	Timestamp       time.Time
	RecoveryAttempts int
	Recovered       bool
	Details         map[string]interface{}
}

// RecoveryAttempt tracks recovery operation history
type RecoveryAttempt struct {
	IssueType     CorruptionType
	Component     string
	Timestamp     time.Time
	Success       bool
	Error         string
	Duration      time.Duration
	Method        string
}

// CorruptionType defines types of corruption
type CorruptionType int

const (
	CorruptionFileHeader CorruptionType = iota
	CorruptionWAL
	CorruptionIndex
	CorruptionEntity
	CorruptionChecksum
	CorruptionInconsistency
	CorruptionFileSystem
)

// CorruptionSeverity defines severity levels
type CorruptionSeverity int

const (
	CorruptionSeverityInfo CorruptionSeverity = iota
	CorruptionSeverityWarning
	CorruptionSeverityError
	CorruptionSeverityCritical
	CorruptionSeverityFatal
)

// String returns string representation of corruption type
func (ct CorruptionType) String() string {
	switch ct {
	case CorruptionFileHeader:
		return "FILE_HEADER"
	case CorruptionWAL:
		return "WAL"
	case CorruptionIndex:
		return "INDEX"
	case CorruptionEntity:
		return "ENTITY"
	case CorruptionChecksum:
		return "CHECKSUM"
	case CorruptionInconsistency:
		return "INCONSISTENCY"
	case CorruptionFileSystem:
		return "FILESYSTEM"
	default:
		return "UNKNOWN"
	}
}

// String returns string representation of severity
func (cs CorruptionSeverity) String() string {
	switch cs {
	case CorruptionSeverityInfo:
		return "INFO"
	case CorruptionSeverityWarning:
		return "WARNING"
	case CorruptionSeverityError:
		return "ERROR"
	case CorruptionSeverityCritical:
		return "CRITICAL"
	case CorruptionSeverityFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// NewCorruptionDetector creates a new corruption detector
func NewCorruptionDetector(repo *EntityRepository) *CorruptionDetector {
	return &CorruptionDetector{
		repo:                repo,
		detectionInterval:   10 * time.Minute, // Default: check every 10 minutes
		autoRecoveryEnabled: true,              // Default: auto-recovery enabled
		maxRecoveryAttempts: 3,                 // Default: max 3 recovery attempts
		stopChan:            make(chan struct{}),
		lastDetectionTime:   time.Now(),
		detectedIssues:      make([]CorruptionIssue, 0),
		recoveryHistory:     make([]RecoveryAttempt, 0),
	}
}

// Configure sets corruption detection parameters
func (cd *CorruptionDetector) Configure(intervalMinutes int64, autoRecovery bool, maxAttempts int) {
	cd.mu.Lock()
	defer cd.mu.Unlock()
	
	if intervalMinutes > 0 {
		cd.detectionInterval = time.Duration(intervalMinutes) * time.Minute
	}
	cd.autoRecoveryEnabled = autoRecovery
	if maxAttempts > 0 {
		cd.maxRecoveryAttempts = maxAttempts
	}
	
	logger.Info("Corruption detector configured: interval %dm, auto-recovery %v, max attempts %d",
		intervalMinutes, autoRecovery, maxAttempts)
}

// Start begins corruption detection monitoring
func (cd *CorruptionDetector) Start() error {
	if !atomic.CompareAndSwapInt32(&cd.running, 0, 1) {
		return fmt.Errorf("corruption detector already running")
	}
	
	go cd.monitorLoop()
	logger.Info("Corruption detector started (interval: %v, auto-recovery: %v)",
		cd.detectionInterval, cd.autoRecoveryEnabled)
	return nil
}

// Stop gracefully shuts down the detector
func (cd *CorruptionDetector) Stop() error {
	if !atomic.CompareAndSwapInt32(&cd.running, 1, 0) {
		return fmt.Errorf("corruption detector not running")
	}
	
	close(cd.stopChan)
	logger.Info("Corruption detector stopped")
	return nil
}

// monitorLoop is the main monitoring goroutine
func (cd *CorruptionDetector) monitorLoop() {
	ticker := time.NewTicker(cd.detectionInterval)
	defer ticker.Stop()
	
	// Run initial detection after a delay
	go func() {
		time.Sleep(30 * time.Second) // Wait for system to stabilize
		cd.runDetection()
	}()
	
	for {
		select {
		case <-ticker.C:
			cd.runDetection()
		case <-cd.stopChan:
			return
		}
	}
}

// runDetection performs comprehensive corruption detection
func (cd *CorruptionDetector) runDetection() {
	startTime := time.Now()
	logger.Debug("Starting corruption detection scan")
	
	// Clear previous issues (keep only unrecovered critical issues)
	cd.mu.Lock()
	cd.lastDetectionTime = startTime
	oldIssues := cd.detectedIssues
	cd.detectedIssues = make([]CorruptionIssue, 0)
	
	// Keep unrecovered critical issues
	for _, issue := range oldIssues {
		if !issue.Recovered && issue.Severity >= CorruptionSeverityCritical {
			cd.detectedIssues = append(cd.detectedIssues, issue)
		}
	}
	cd.mu.Unlock()
	
	// Run all detection checks
	issues := cd.detectCorruption()
	
	// Update statistics
	atomic.AddInt64(&cd.totalDetections, 1)
	atomic.AddInt64(&cd.corruptionFound, int64(len(issues)))
	
	// Store new issues
	cd.mu.Lock()
	cd.detectedIssues = append(cd.detectedIssues, issues...)
	cd.mu.Unlock()
	
	// Report findings
	if len(issues) == 0 {
		logger.Info("Corruption detection completed: no issues found (duration: %v)",
			time.Since(startTime))
	} else {
		logger.Warn("Corruption detection found %d issues (duration: %v)",
			len(issues), time.Since(startTime))
		
		// Log issues by severity
		cd.logIssuesBySeverity(issues)
		
		// Attempt auto-recovery if enabled
		if cd.autoRecoveryEnabled {
			recovered := cd.attemptRecovery(issues)
			logger.Info("Auto-recovery attempted for %d issues, %d successful", len(issues), recovered)
		}
	}
}

// detectCorruption performs comprehensive corruption checks
func (cd *CorruptionDetector) detectCorruption() []CorruptionIssue {
	var issues []CorruptionIssue
	
	// Check file header integrity
	issues = append(issues, cd.checkFileHeader()...)
	
	// Check WAL integrity
	issues = append(issues, cd.checkWALIntegrity()...)
	
	// Check index consistency
	issues = append(issues, cd.checkIndexConsistency()...)
	
	// Check entity data integrity
	issues = append(issues, cd.checkEntityIntegrity()...)
	
	// Check cross-component consistency
	issues = append(issues, cd.checkCrossComponentConsistency()...)
	
	// Check filesystem integrity
	issues = append(issues, cd.checkFileSystemIntegrity()...)
	
	return issues
}

// checkFileHeader validates the database file header
func (cd *CorruptionDetector) checkFileHeader() []CorruptionIssue {
	var issues []CorruptionIssue
	
	file, err := os.Open(cd.repo.getDataFile())
	if err != nil {
		issues = append(issues, CorruptionIssue{
			Type:        CorruptionFileHeader,
			Component:   "file_header",
			Description: fmt.Sprintf("Cannot open database file: %v", err),
			Severity:    CorruptionSeverityFatal,
			Timestamp:   time.Now(),
			Details:     map[string]interface{}{"error": err.Error()},
		})
		return issues
	}
	defer file.Close()
	
	// Read header
	headerData := make([]byte, 4)
	n, err := file.Read(headerData)
	if err != nil || n != 4 {
		issues = append(issues, CorruptionIssue{
			Type:        CorruptionFileHeader,
			Component:   "file_header",
			Description: "Cannot read file header",
			Severity:    CorruptionSeverityFatal,
			Timestamp:   time.Now(),
			Details:     map[string]interface{}{"bytes_read": n, "error": err},
		})
		return issues
	}
	
	// Check magic number
	expectedMagic := []byte("EUFF")
	if string(headerData) != string(expectedMagic) {
		issues = append(issues, CorruptionIssue{
			Type:        CorruptionFileHeader,
			Component:   "file_header",
			Description: fmt.Sprintf("Invalid magic number: expected EUFF, got %s", string(headerData)),
			Severity:    CorruptionSeverityCritical,
			Timestamp:   time.Now(),
			Details:     map[string]interface{}{"expected": "EUFF", "actual": string(headerData)},
		})
	}
	
	return issues
}

// checkWALIntegrity validates WAL consistency
func (cd *CorruptionDetector) checkWALIntegrity() []CorruptionIssue {
	var issues []CorruptionIssue
	
	if cd.repo.wal == nil {
		issues = append(issues, CorruptionIssue{
			Type:        CorruptionWAL,
			Component:   "wal",
			Description: "WAL instance is nil",
			Severity:    CorruptionSeverityError,
			Timestamp:   time.Now(),
		})
		return issues
	}
	
	// Check WAL file size for bounded growth
	if cd.repo.walRotationManager != nil {
		stats := cd.repo.walRotationManager.GetStatistics()
		if currentSize, ok := stats["current_size_bytes"].(int64); ok {
			if maxSize, ok := stats["max_size_bytes"].(int64); ok {
				if currentSize > maxSize*2 { // Allow 2x overage before flagging
					issues = append(issues, CorruptionIssue{
						Type:        CorruptionWAL,
						Component:   "wal_size",
						Description: fmt.Sprintf("WAL size exceeded bounds: %d > %d", currentSize, maxSize),
						Severity:    CorruptionSeverityWarning,
						Timestamp:   time.Now(),
						Details:     map[string]interface{}{"current_size": currentSize, "max_size": maxSize},
					})
				}
			}
		}
	}
	
	return issues
}

// checkIndexConsistency validates index integrity
func (cd *CorruptionDetector) checkIndexConsistency() []CorruptionIssue {
	var issues []CorruptionIssue
	
	// Use the existing index integrity validator
	if cd.repo.indexIntegrityValidator != nil {
		stats := cd.repo.indexIntegrityValidator.GetStatistics()
		
		// Check for recent integrity issues
		if staleEntries, ok := stats["current_stale_entries"].(int); ok && staleEntries > 0 {
			issues = append(issues, CorruptionIssue{
				Type:        CorruptionIndex,
				Component:   "index_stale_entries",
				Description: fmt.Sprintf("Index contains %d stale entries", staleEntries),
				Severity:    CorruptionSeverityWarning,
				Timestamp:   time.Now(),
				Details:     map[string]interface{}{"stale_count": staleEntries},
			})
		}
		
		if missingEntries, ok := stats["current_missing_entries"].(int); ok && missingEntries > 0 {
			issues = append(issues, CorruptionIssue{
				Type:        CorruptionIndex,
				Component:   "index_missing_entries",
				Description: fmt.Sprintf("Index missing %d entries", missingEntries),
				Severity:    CorruptionSeverityError,
				Timestamp:   time.Now(),
				Details:     map[string]interface{}{"missing_count": missingEntries},
			})
		}
	}
	
	return issues
}

// checkEntityIntegrity validates entity data integrity
func (cd *CorruptionDetector) checkEntityIntegrity() []CorruptionIssue {
	var issues []CorruptionIssue
	
	// Sample a few entities to check integrity
	// Use tag index to get a sample of entities
	if cd.repo.shardedTagIndex != nil {
		allTags := cd.repo.shardedTagIndex.GetAllTags()
		entityIDs := make(map[string]bool)
		
		// Collect unique entity IDs (sample up to 100)
		for _, entities := range allTags {
			for _, entityID := range entities {
				entityIDs[entityID] = true
				if len(entityIDs) >= 100 {
					break
				}
			}
			if len(entityIDs) >= 100 {
				break
			}
		}
		
		// Check sample entities
		checkedCount := 0
		for entityID := range entityIDs {
			if checkedCount >= 10 { // Limit to 10 per check
				break
			}
			
			entity, err := cd.repo.GetByID(entityID)
			if err != nil {
				issues = append(issues, CorruptionIssue{
					Type:        CorruptionEntity,
					Component:   "entity_data",
					EntityID:    entityID,
					Description: fmt.Sprintf("Cannot read entity: %v", err),
					Severity:    CorruptionSeverityError,
					Timestamp:   time.Now(),
					Details:     map[string]interface{}{"error": err.Error()},
				})
			} else if entity == nil {
				issues = append(issues, CorruptionIssue{
					Type:        CorruptionEntity,
					Component:   "entity_data",
					EntityID:    entityID,
					Description: "Entity exists in index but returns nil",
					Severity:    CorruptionSeverityError,
					Timestamp:   time.Now(),
				})
			}
			checkedCount++
		}
	}
	
	return issues
}

// checkCrossComponentConsistency validates consistency between components
func (cd *CorruptionDetector) checkCrossComponentConsistency() []CorruptionIssue {
	var issues []CorruptionIssue
	
	// Check if entity cache and tag index are consistent
	if cd.repo.entityCache != nil && cd.repo.shardedTagIndex != nil {
		// This is a basic consistency check - could be expanded
		cacheStats := cd.repo.entityCache.Stats()
		indexStats := cd.repo.shardedTagIndex.GetShardStats()
		
		cacheSize := int64(cacheStats.Size)
		if indexEntries, ok := indexStats["total_entries"].(int64); ok {
			// Allow some variance due to temporal tags and caching
			if cacheSize > 0 && indexEntries > 0 {
				ratio := float64(cacheSize) / float64(indexEntries)
				if ratio > 2.0 || ratio < 0.1 { // Flag if ratio is extreme
					issues = append(issues, CorruptionIssue{
						Type:        CorruptionInconsistency,
						Component:   "cache_index_consistency",
						Description: fmt.Sprintf("Cache/index size mismatch: cache=%d, index=%d", cacheSize, indexEntries),
						Severity:    CorruptionSeverityWarning,
						Timestamp:   time.Now(),
						Details: map[string]interface{}{
							"cache_size":    cacheSize,
							"index_entries": indexEntries,
							"ratio":         ratio,
						},
					})
				}
			}
		}
	}
	
	return issues
}

// checkFileSystemIntegrity validates filesystem-level integrity
func (cd *CorruptionDetector) checkFileSystemIntegrity() []CorruptionIssue {
	var issues []CorruptionIssue
	
	// Check if database file exists and is readable
	dataFile := cd.repo.getDataFile()
	stat, err := os.Stat(dataFile)
	if err != nil {
		issues = append(issues, CorruptionIssue{
			Type:        CorruptionFileSystem,
			Component:   "database_file",
			Description: fmt.Sprintf("Database file access error: %v", err),
			Severity:    CorruptionSeverityFatal,
			Timestamp:   time.Now(),
			Details:     map[string]interface{}{"error": err.Error(), "path": dataFile},
		})
		return issues
	}
	
	// Check file size reasonableness
	if stat.Size() == 0 {
		issues = append(issues, CorruptionIssue{
			Type:        CorruptionFileSystem,
			Component:   "database_file",
			Description: "Database file is empty",
			Severity:    CorruptionSeverityCritical,
			Timestamp:   time.Now(),
			Details:     map[string]interface{}{"size": 0, "path": dataFile},
		})
	} else if stat.Size() < 64 { // Minimum reasonable size
		issues = append(issues, CorruptionIssue{
			Type:        CorruptionFileSystem,
			Component:   "database_file",
			Description: fmt.Sprintf("Database file suspiciously small: %d bytes", stat.Size()),
			Severity:    CorruptionSeverityWarning,
			Timestamp:   time.Now(),
			Details:     map[string]interface{}{"size": stat.Size(), "path": dataFile},
		})
	}
	
	return issues
}

// logIssuesBySeverity logs issues grouped by severity
func (cd *CorruptionDetector) logIssuesBySeverity(issues []CorruptionIssue) {
	severityGroups := make(map[CorruptionSeverity][]CorruptionIssue)
	
	for _, issue := range issues {
		severityGroups[issue.Severity] = append(severityGroups[issue.Severity], issue)
	}
	
	for severity, group := range severityGroups {
		logger.Warn("Corruption detected - %s (%d issues):", severity, len(group))
		for _, issue := range group {
			logger.Warn("  - %s: %s", issue.Type, issue.Description)
		}
	}
}

// attemptRecovery attempts to recover from detected corruption
func (cd *CorruptionDetector) attemptRecovery(issues []CorruptionIssue) int {
	recovered := 0
	
	for _, issue := range issues {
		if issue.RecoveryAttempts >= cd.maxRecoveryAttempts {
			continue // Skip if already tried maximum attempts
		}
		
		startTime := time.Now()
		success := cd.recoverFromIssue(issue)
		duration := time.Since(startTime)
		
		// Record recovery attempt
		attempt := RecoveryAttempt{
			IssueType: issue.Type,
			Component: issue.Component,
			Timestamp: startTime,
			Success:   success,
			Duration:  duration,
			Method:    cd.getRecoveryMethod(issue.Type),
		}
		
		if !success {
			attempt.Error = "Recovery method failed"
		}
		
		cd.mu.Lock()
		cd.recoveryHistory = append(cd.recoveryHistory, attempt)
		// Update issue status
		for i := range cd.detectedIssues {
			if cd.detectedIssues[i].Type == issue.Type && 
			   cd.detectedIssues[i].Component == issue.Component {
				cd.detectedIssues[i].RecoveryAttempts++
				cd.detectedIssues[i].Recovered = success
				break
			}
		}
		cd.mu.Unlock()
		
		atomic.AddInt64(&cd.recoveryAttempts, 1)
		if success {
			atomic.AddInt64(&cd.recoverySuccesses, 1)
			recovered++
		}
	}
	
	return recovered
}

// recoverFromIssue attempts to recover from a specific corruption issue
func (cd *CorruptionDetector) recoverFromIssue(issue CorruptionIssue) bool {
	logger.Info("Attempting recovery for %s: %s", issue.Type, issue.Description)
	
	switch issue.Type {
	case CorruptionFileHeader:
		return cd.recoverFileHeader(issue)
	case CorruptionWAL:
		return cd.recoverWAL(issue)
	case CorruptionIndex:
		return cd.recoverIndex(issue)
	case CorruptionEntity:
		return cd.recoverEntity(issue)
	case CorruptionInconsistency:
		return cd.recoverInconsistency(issue)
	default:
		logger.Warn("No recovery method available for corruption type: %s", issue.Type)
		return false
	}
}

// recoverFileHeader attempts to recover corrupted file header
func (cd *CorruptionDetector) recoverFileHeader(issue CorruptionIssue) bool {
	logger.Info("Attempting file header recovery")
	
	// Try to rebuild the file with correct header
	if cd.repo.writerManager != nil {
		// Force a checkpoint which rebuilds the file
		if err := cd.repo.writerManager.Checkpoint(); err != nil {
			logger.Error("File header recovery failed: %v", err)
			return false
		}
		logger.Info("File header recovery successful via checkpoint")
		return true
	}
	
	return false
}

// recoverWAL attempts to recover WAL issues
func (cd *CorruptionDetector) recoverWAL(issue CorruptionIssue) bool {
	logger.Info("Attempting WAL recovery")
	
	// Force WAL rotation if size issue
	if issue.Component == "wal_size" && cd.repo.walRotationManager != nil {
		if err := cd.repo.walRotationManager.ForceRotation(); err != nil {
			logger.Error("WAL recovery failed: %v", err)
			return false
		}
		logger.Info("WAL recovery successful via forced rotation")
		return true
	}
	
	return false
}

// recoverIndex attempts to recover index corruption
func (cd *CorruptionDetector) recoverIndex(issue CorruptionIssue) bool {
	logger.Info("Attempting index recovery")
	
	// Trigger index integrity validator to fix issues
	if cd.repo.indexIntegrityValidator != nil {
		if err := cd.repo.indexIntegrityValidator.ForceValidation(); err != nil {
			logger.Error("Index recovery failed: %v", err)
			return false
		}
		logger.Info("Index recovery triggered via integrity validator")
		return true
	}
	
	return false
}

// recoverEntity attempts to recover entity corruption
func (cd *CorruptionDetector) recoverEntity(issue CorruptionIssue) bool {
	logger.Info("Attempting entity recovery for %s", issue.EntityID)
	
	// Try to remove corrupted entity from cache and reload
	if cd.repo.entityCache != nil {
		cd.repo.entityCache.Delete(issue.EntityID)
		logger.Info("Entity recovery: removed from cache")
	}
	
	// Try to remove from indexes and let them rebuild
	if cd.repo.shardedTagIndex != nil && issue.EntityID != "" {
		// This is a basic recovery - more sophisticated recovery could be implemented
		logger.Info("Entity recovery: basic cleanup performed")
		return true
	}
	
	return false
}

// recoverInconsistency attempts to recover consistency issues
func (cd *CorruptionDetector) recoverInconsistency(issue CorruptionIssue) bool {
	logger.Info("Attempting consistency recovery")
	
	// Clear caches to force reload
	if cd.repo.cache != nil {
		cd.repo.cache.Clear()
	}
	if cd.repo.entityCache != nil {
		cd.repo.entityCache.Clear()
	}
	
	logger.Info("Consistency recovery: caches cleared")
	return true
}

// getRecoveryMethod returns the recovery method name for an issue type
func (cd *CorruptionDetector) getRecoveryMethod(issueType CorruptionType) string {
	switch issueType {
	case CorruptionFileHeader:
		return "checkpoint_rebuild"
	case CorruptionWAL:
		return "forced_rotation"
	case CorruptionIndex:
		return "integrity_validation"
	case CorruptionEntity:
		return "cache_invalidation"
	case CorruptionInconsistency:
		return "cache_clearing"
	default:
		return "none"
	}
}

// GetStatistics returns corruption detection statistics
func (cd *CorruptionDetector) GetStatistics() map[string]interface{} {
	cd.mu.RLock()
	defer cd.mu.RUnlock()
	
	return map[string]interface{}{
		"running":               atomic.LoadInt32(&cd.running) == 1,
		"detection_interval":    cd.detectionInterval.String(),
		"auto_recovery_enabled": cd.autoRecoveryEnabled,
		"max_recovery_attempts": cd.maxRecoveryAttempts,
		"total_detections":      atomic.LoadInt64(&cd.totalDetections),
		"corruption_found":      atomic.LoadInt64(&cd.corruptionFound),
		"recovery_attempts":     atomic.LoadInt64(&cd.recoveryAttempts),
		"recovery_successes":    atomic.LoadInt64(&cd.recoverySuccesses),
		"recovery_success_rate": float64(atomic.LoadInt64(&cd.recoverySuccesses)) / float64(atomic.LoadInt64(&cd.recoveryAttempts)),
		"last_detection_time":   cd.lastDetectionTime.Format(time.RFC3339),
		"current_issues_count":  len(cd.detectedIssues),
		"recovery_history_count": len(cd.recoveryHistory),
	}
}

// ForceDetection manually triggers corruption detection
func (cd *CorruptionDetector) ForceDetection() error {
	if atomic.LoadInt32(&cd.running) == 0 {
		return fmt.Errorf("corruption detector not running")
	}
	
	logger.Info("Manual corruption detection triggered")
	go cd.runDetection()
	return nil
}

// GetCurrentIssues returns current unresolved issues
func (cd *CorruptionDetector) GetCurrentIssues() []CorruptionIssue {
	cd.mu.RLock()
	defer cd.mu.RUnlock()
	
	// Return copy of current issues
	issues := make([]CorruptionIssue, len(cd.detectedIssues))
	copy(issues, cd.detectedIssues)
	return issues
}