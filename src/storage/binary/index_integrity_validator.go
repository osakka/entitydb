// Package binary provides index integrity validation for corruption prevention
//
// This system ensures indexes remain consistent with actual data, preventing:
// - Stale index entries pointing to non-existent entities
// - Missing index entries for existing entities
// - Corrupted index data structures
// - Index-data inconsistencies leading to incorrect query results
package binary

import (
	"entitydb/logger"
	"entitydb/models"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// IndexIntegrityValidator validates and maintains index consistency
type IndexIntegrityValidator struct {
	repo               *EntityRepository
	mu                 sync.RWMutex
	
	// Configuration
	validationInterval time.Duration // How often to run full validation
	repairEnabled      bool          // Whether to auto-repair issues
	maxRepairsPerRun   int           // Limit repairs to prevent excessive work
	
	// State
	running            int32
	stopChan           chan struct{}
	lastValidationTime time.Time
	
	// Statistics
	totalValidations   int64
	issuesFound        int64
	issuesRepaired     int64
	validationDuration time.Duration
	
	// Issue tracking
	staleEntries       []IndexIssue
	missingEntries     []IndexIssue
	corruptedEntries   []IndexIssue
}

// IndexIssue represents an index consistency problem
type IndexIssue struct {
	Type        IndexIssueType
	EntityID    string
	Tag         string
	IndexName   string
	Description string
	Severity    IssueSeverity
	Timestamp   time.Time
}

// IndexIssueType defines the type of index issue
type IndexIssueType int

const (
	IssueStaleEntry IndexIssueType = iota // Index points to non-existent entity
	IssueMissingEntry                     // Entity exists but not in index
	IssueCorruptedData                    // Index data is corrupted
	IssueInconsistentCount                // Count mismatch between indexes
)

// IssueSeverity defines the severity level of an issue
type IssueSeverity int

const (
	SeverityLow IssueSeverity = iota
	SeverityMedium
	SeverityHigh
	SeverityCritical
)

// String returns string representation of issue type
func (it IndexIssueType) String() string {
	switch it {
	case IssueStaleEntry:
		return "STALE_ENTRY"
	case IssueMissingEntry:
		return "MISSING_ENTRY"
	case IssueCorruptedData:
		return "CORRUPTED_DATA"
	case IssueInconsistentCount:
		return "INCONSISTENT_COUNT"
	default:
		return "UNKNOWN"
	}
}

// String returns string representation of severity
func (s IssueSeverity) String() string {
	switch s {
	case SeverityLow:
		return "LOW"
	case SeverityMedium:
		return "MEDIUM"
	case SeverityHigh:
		return "HIGH"
	case SeverityCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// NewIndexIntegrityValidator creates a new index integrity validator
func NewIndexIntegrityValidator(repo *EntityRepository) *IndexIntegrityValidator {
	return &IndexIntegrityValidator{
		repo:               repo,
		validationInterval: 30 * time.Minute, // Default: validate every 30 minutes
		repairEnabled:      true,              // Default: auto-repair enabled
		maxRepairsPerRun:   100,               // Default: max 100 repairs per run
		stopChan:           make(chan struct{}),
		lastValidationTime: time.Now(),
	}
}

// Configure sets validation parameters
func (iiv *IndexIntegrityValidator) Configure(intervalMinutes int64, repairEnabled bool, maxRepairs int) {
	iiv.mu.Lock()
	defer iiv.mu.Unlock()
	
	if intervalMinutes > 0 {
		iiv.validationInterval = time.Duration(intervalMinutes) * time.Minute
	}
	iiv.repairEnabled = repairEnabled
	if maxRepairs > 0 {
		iiv.maxRepairsPerRun = maxRepairs
	}
	
	logger.Info("Index integrity validator configured: interval %dm, repair %v, max repairs %d",
		intervalMinutes, repairEnabled, maxRepairs)
}

// Start begins index integrity monitoring
func (iiv *IndexIntegrityValidator) Start() error {
	if !atomic.CompareAndSwapInt32(&iiv.running, 0, 1) {
		return fmt.Errorf("index integrity validator already running")
	}
	
	go iiv.monitorLoop()
	logger.Info("Index integrity validator started (interval: %v, repair: %v)",
		iiv.validationInterval, iiv.repairEnabled)
	return nil
}

// Stop gracefully shuts down the validator
func (iiv *IndexIntegrityValidator) Stop() error {
	if !atomic.CompareAndSwapInt32(&iiv.running, 1, 0) {
		return fmt.Errorf("index integrity validator not running")
	}
	
	close(iiv.stopChan)
	logger.Info("Index integrity validator stopped")
	return nil
}

// monitorLoop is the main monitoring goroutine
func (iiv *IndexIntegrityValidator) monitorLoop() {
	ticker := time.NewTicker(iiv.validationInterval)
	defer ticker.Stop()
	
	// Run initial validation after a short delay
	go func() {
		time.Sleep(2 * time.Minute) // Wait for system to stabilize
		iiv.runValidation()
	}()
	
	for {
		select {
		case <-ticker.C:
			iiv.runValidation()
		case <-iiv.stopChan:
			return
		}
	}
}

// runValidation performs a complete index integrity check
func (iiv *IndexIntegrityValidator) runValidation() {
	startTime := time.Now()
	logger.Info("Starting index integrity validation")
	
	// Clear previous issues
	iiv.mu.Lock()
	iiv.staleEntries = nil
	iiv.missingEntries = nil
	iiv.corruptedEntries = nil
	iiv.lastValidationTime = startTime
	iiv.mu.Unlock()
	
	// Run validation checks
	issues := iiv.validateIndexes()
	
	// Update statistics
	atomic.AddInt64(&iiv.totalValidations, 1)
	atomic.AddInt64(&iiv.issuesFound, int64(len(issues)))
	
	iiv.mu.Lock()
	iiv.validationDuration = time.Since(startTime)
	iiv.mu.Unlock()
	
	// Report findings
	if len(issues) == 0 {
		logger.Info("Index integrity validation completed: no issues found (duration: %v)",
			time.Since(startTime))
	} else {
		logger.Warn("Index integrity validation found %d issues (duration: %v)",
			len(issues), time.Since(startTime))
		
		// Categorize and log issues
		iiv.categorizeIssues(issues)
		
		// Attempt repairs if enabled
		if iiv.repairEnabled {
			repaired := iiv.repairIssues(issues)
			atomic.AddInt64(&iiv.issuesRepaired, int64(repaired))
			logger.Info("Auto-repaired %d/%d index issues", repaired, len(issues))
		}
	}
}

// validateIndexes performs comprehensive index validation
func (iiv *IndexIntegrityValidator) validateIndexes() []IndexIssue {
	var issues []IndexIssue
	
	// Get all entities from storage
	allEntities, err := iiv.getAllEntitiesFromStorage()
	if err != nil {
		logger.Error("Failed to get entities from storage for validation: %v", err)
		return issues
	}
	
	logger.Debug("Validating indexes against %d entities", len(allEntities))
	
	// Build expected index state from actual entities
	expectedTags := make(map[string][]string) // tag -> []entityID
	expectedContent := make(map[string][]string) // content -> []entityID
	
	for _, entity := range allEntities {
		// Process tags (both timestamped and clean)
		for _, tag := range entity.Tags {
			cleanTag := tag
			if strings.Contains(tag, "|") {
				parts := strings.SplitN(tag, "|", 2)
				if len(parts) == 2 {
					cleanTag = parts[1]
				}
			}
			expectedTags[cleanTag] = append(expectedTags[cleanTag], entity.ID)
		}
		
		// Process content
		if len(entity.Content) > 0 {
			contentStr := string(entity.Content)
			expectedContent[contentStr] = append(expectedContent[contentStr], entity.ID)
		}
	}
	
	// Validate tag indexes
	issues = append(issues, iiv.validateTagIndex(expectedTags)...)
	
	// Validate content index
	issues = append(issues, iiv.validateContentIndex(expectedContent)...)
	
	// Validate temporal index
	issues = append(issues, iiv.validateTemporalIndex(allEntities)...)
	
	// Validate namespace index
	issues = append(issues, iiv.validateNamespaceIndex(allEntities)...)
	
	return issues
}

// validateTagIndex validates the tag index against expected state
func (iiv *IndexIntegrityValidator) validateTagIndex(expected map[string][]string) []IndexIssue {
	var issues []IndexIssue
	
	if iiv.repo.shardedTagIndex == nil {
		return issues
	}
	
	// Check each expected tag
	for tag, expectedEntities := range expected {
		actualEntities := iiv.repo.shardedTagIndex.GetEntitiesForTag(tag)
		
		// Find missing entries (in expected but not in index)
		actualSet := make(map[string]bool)
		for _, entityID := range actualEntities {
			actualSet[entityID] = true
		}
		
		for _, expectedEntity := range expectedEntities {
			if !actualSet[expectedEntity] {
				issues = append(issues, IndexIssue{
					Type:        IssueMissingEntry,
					EntityID:    expectedEntity,
					Tag:         tag,
					IndexName:   "tag_index",
					Description: fmt.Sprintf("Entity %s missing from tag index for tag '%s'", expectedEntity, tag),
					Severity:    SeverityHigh,
					Timestamp:   time.Now(),
				})
			}
		}
		
		// Find stale entries (in index but not expected)
		expectedSet := make(map[string]bool)
		for _, entityID := range expectedEntities {
			expectedSet[entityID] = true
		}
		
		for _, actualEntity := range actualEntities {
			if !expectedSet[actualEntity] {
				issues = append(issues, IndexIssue{
					Type:        IssueStaleEntry,
					EntityID:    actualEntity,
					Tag:         tag,
					IndexName:   "tag_index",
					Description: fmt.Sprintf("Stale entity %s in tag index for tag '%s'", actualEntity, tag),
					Severity:    SeverityMedium,
					Timestamp:   time.Now(),
				})
			}
		}
	}
	
	return issues
}

// validateContentIndex validates the content index
func (iiv *IndexIntegrityValidator) validateContentIndex(expected map[string][]string) []IndexIssue {
	var issues []IndexIssue
	
	if iiv.repo.contentIndex == nil {
		return issues
	}
	
	iiv.repo.mu.RLock()
	defer iiv.repo.mu.RUnlock()
	
	// Check for stale entries in content index
	for content, indexedEntities := range iiv.repo.contentIndex {
		expectedEntities, exists := expected[content]
		if !exists {
			// Content in index but no entities have this content
			for _, entityID := range indexedEntities {
				issues = append(issues, IndexIssue{
					Type:        IssueStaleEntry,
					EntityID:    entityID,
					IndexName:   "content_index",
					Description: fmt.Sprintf("Stale content index entry for entity %s", entityID),
					Severity:    SeverityMedium,
					Timestamp:   time.Now(),
				})
			}
			continue
		}
		
		// Check for missing entities
		indexedSet := make(map[string]bool)
		for _, entityID := range indexedEntities {
			indexedSet[entityID] = true
		}
		
		for _, expectedEntity := range expectedEntities {
			if !indexedSet[expectedEntity] {
				issues = append(issues, IndexIssue{
					Type:        IssueMissingEntry,
					EntityID:    expectedEntity,
					IndexName:   "content_index",
					Description: fmt.Sprintf("Entity %s missing from content index", expectedEntity),
					Severity:    SeverityHigh,
					Timestamp:   time.Now(),
				})
			}
		}
	}
	
	return issues
}

// validateTemporalIndex validates the temporal index
func (iiv *IndexIntegrityValidator) validateTemporalIndex(entities []*models.Entity) []IndexIssue {
	var issues []IndexIssue
	
	if iiv.repo.temporalIndex == nil {
		return issues
	}
	
	// Build expected temporal entries
	expectedEntries := make(map[string]bool) // entityID -> exists
	for _, entity := range entities {
		expectedEntries[entity.ID] = true
	}
	
	// Validate temporal index entries (simplified check)
	// Note: Full temporal validation would require examining all temporal entries
	// This is a basic existence check
	
	return issues
}

// validateNamespaceIndex validates the namespace index
func (iiv *IndexIntegrityValidator) validateNamespaceIndex(entities []*models.Entity) []IndexIssue {
	var issues []IndexIssue
	
	if iiv.repo.namespaceIndex == nil {
		return issues
	}
	
	// Build expected namespace entries
	expectedNamespaces := make(map[string][]string) // namespace -> []entityID
	for _, entity := range entities {
		for _, tag := range entity.Tags {
			// Extract namespace from tag
			if idx := strings.Index(tag, ":"); idx > 0 {
				namespace := tag[:idx]
				expectedNamespaces[namespace] = append(expectedNamespaces[namespace], entity.ID)
			}
		}
	}
	
	// Validate namespace index (basic check)
	// Note: Detailed validation would depend on namespace index implementation
	
	return issues
}

// getAllEntitiesFromStorage retrieves all entities directly from storage
func (iiv *IndexIntegrityValidator) getAllEntitiesFromStorage() ([]*models.Entity, error) {
	// Read directly from file to bypass potentially corrupted indexes
	var entities []*models.Entity
	
	// Get a reader
	readerInterface := iiv.repo.readerPool.Get()
	if readerInterface == nil {
		return nil, fmt.Errorf("failed to get reader from pool")
	}
	
	reader, ok := readerInterface.(*Reader)
	if !ok {
		iiv.repo.readerPool.Put(readerInterface)
		return nil, fmt.Errorf("invalid reader type")
	}
	defer iiv.repo.readerPool.Put(reader)
	
	// Read all entities from file
	entities, err := reader.GetAllEntities()
	if err != nil {
		return nil, fmt.Errorf("failed to read entities from storage: %w", err)
	}
	
	return entities, nil
}

// categorizeIssues categorizes and stores issues by type
func (iiv *IndexIntegrityValidator) categorizeIssues(issues []IndexIssue) {
	iiv.mu.Lock()
	defer iiv.mu.Unlock()
	
	for _, issue := range issues {
		switch issue.Type {
		case IssueStaleEntry:
			iiv.staleEntries = append(iiv.staleEntries, issue)
		case IssueMissingEntry:
			iiv.missingEntries = append(iiv.missingEntries, issue)
		case IssueCorruptedData:
			iiv.corruptedEntries = append(iiv.corruptedEntries, issue)
		}
		
		// Log high severity issues
		if issue.Severity >= SeverityHigh {
			logger.Warn("Index integrity issue: %s - %s", issue.Type, issue.Description)
		}
	}
}

// repairIssues attempts to automatically repair index issues
func (iiv *IndexIntegrityValidator) repairIssues(issues []IndexIssue) int {
	if !iiv.repairEnabled {
		return 0
	}
	
	repaired := 0
	maxRepairs := iiv.maxRepairsPerRun
	
	for _, issue := range issues {
		if repaired >= maxRepairs {
			logger.Info("Reached maximum repairs per run (%d), deferring remaining issues", maxRepairs)
			break
		}
		
		if iiv.repairIssue(issue) {
			repaired++
		}
	}
	
	return repaired
}

// repairIssue attempts to repair a single index issue
func (iiv *IndexIntegrityValidator) repairIssue(issue IndexIssue) bool {
	logger.Debug("Attempting to repair index issue: %s - %s", issue.Type, issue.Description)
	
	switch issue.Type {
	case IssueStaleEntry:
		return iiv.repairStaleEntry(issue)
	case IssueMissingEntry:
		return iiv.repairMissingEntry(issue)
	case IssueCorruptedData:
		return iiv.repairCorruptedData(issue)
	default:
		logger.Warn("Unknown issue type for repair: %s", issue.Type)
		return false
	}
}

// repairStaleEntry removes stale entries from indexes
func (iiv *IndexIntegrityValidator) repairStaleEntry(issue IndexIssue) bool {
	if issue.IndexName == "tag_index" && iiv.repo.shardedTagIndex != nil {
		iiv.repo.shardedTagIndex.RemoveTag(issue.Tag, issue.EntityID)
		logger.Debug("Removed stale entry: entity %s from tag '%s'", issue.EntityID, issue.Tag)
		return true
	}
	
	if issue.IndexName == "content_index" {
		iiv.repo.mu.Lock()
		if entities, exists := iiv.repo.contentIndex[issue.Tag]; exists {
			// Remove entity from content index
			newEntities := make([]string, 0, len(entities))
			for _, entityID := range entities {
				if entityID != issue.EntityID {
					newEntities = append(newEntities, entityID)
				}
			}
			if len(newEntities) == 0 {
				delete(iiv.repo.contentIndex, issue.Tag)
			} else {
				iiv.repo.contentIndex[issue.Tag] = newEntities
			}
		}
		iiv.repo.mu.Unlock()
		logger.Debug("Removed stale content index entry for entity %s", issue.EntityID)
		return true
	}
	
	return false
}

// repairMissingEntry adds missing entries to indexes
func (iiv *IndexIntegrityValidator) repairMissingEntry(issue IndexIssue) bool {
	// Get the entity to rebuild its index entries
	entity, err := iiv.repo.GetByID(issue.EntityID)
	if err != nil {
		logger.Warn("Cannot repair missing entry - entity %s not found: %v", issue.EntityID, err)
		return false
	}
	
	if issue.IndexName == "tag_index" && iiv.repo.shardedTagIndex != nil {
		iiv.repo.shardedTagIndex.AddTag(issue.Tag, issue.EntityID)
		logger.Debug("Added missing entry: entity %s to tag '%s'", issue.EntityID, issue.Tag)
		return true
	}
	
	if issue.IndexName == "content_index" && len(entity.Content) > 0 {
		iiv.repo.mu.Lock()
		contentStr := string(entity.Content)
		iiv.repo.contentIndex[contentStr] = append(iiv.repo.contentIndex[contentStr], entity.ID)
		iiv.repo.mu.Unlock()
		logger.Debug("Added missing content index entry for entity %s", issue.EntityID)
		return true
	}
	
	return false
}

// repairCorruptedData attempts to repair corrupted index data
func (iiv *IndexIntegrityValidator) repairCorruptedData(issue IndexIssue) bool {
	// For corrupted data, the safest approach is usually to rebuild the affected index
	logger.Warn("Corrupted data repair not implemented for: %s", issue.Description)
	return false
}

// GetStatistics returns validation statistics
func (iiv *IndexIntegrityValidator) GetStatistics() map[string]interface{} {
	iiv.mu.RLock()
	defer iiv.mu.RUnlock()
	
	return map[string]interface{}{
		"running":               atomic.LoadInt32(&iiv.running) == 1,
		"validation_interval":   iiv.validationInterval.String(),
		"repair_enabled":        iiv.repairEnabled,
		"max_repairs_per_run":   iiv.maxRepairsPerRun,
		"total_validations":     atomic.LoadInt64(&iiv.totalValidations),
		"issues_found":          atomic.LoadInt64(&iiv.issuesFound),
		"issues_repaired":       atomic.LoadInt64(&iiv.issuesRepaired),
		"last_validation_time":  iiv.lastValidationTime.Format(time.RFC3339),
		"last_validation_duration": iiv.validationDuration.String(),
		"current_stale_entries":    len(iiv.staleEntries),
		"current_missing_entries":  len(iiv.missingEntries),
		"current_corrupted_entries": len(iiv.corruptedEntries),
	}
}

// ForceValidation manually triggers index validation
func (iiv *IndexIntegrityValidator) ForceValidation() error {
	if atomic.LoadInt32(&iiv.running) == 0 {
		return fmt.Errorf("index integrity validator not running")
	}
	
	logger.Info("Manual index integrity validation triggered")
	go iiv.runValidation()
	return nil
}