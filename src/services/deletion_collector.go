// Package services provides the deletion collector service for EntityDB
package services

import (
	"context"
	"entitydb/logger"
	"entitydb/models"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// DeletionCollector manages automatic entity lifecycle transitions based on retention policies
type DeletionCollector struct {
	// Core dependencies
	repository   models.EntityRepository
	policyEngine *models.PolicyEngine
	
	// Configuration
	config DeletionCollectorConfig
	
	// Runtime state
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	running  int32
	
	// Statistics
	stats CollectorStats
	mu    sync.RWMutex
	
	// System user ID for automated operations
	systemUserID string
}

// DeletionCollectorConfig configures the deletion collector behavior
type DeletionCollectorConfig struct {
	// Enabled controls whether the collector is active
	Enabled bool
	
	// Interval determines how often the collector runs
	Interval time.Duration
	
	// BatchSize limits how many entities to process per cycle
	BatchSize int
	
	// MaxRuntime limits how long a single collection cycle can run
	MaxRuntime time.Duration
	
	// DryRun mode logs what would be done without making changes
	DryRun bool
	
	// EnableMetrics controls whether to track collection statistics
	EnableMetrics bool
	
	// Concurrency controls how many entities to process simultaneously
	Concurrency int
}

// CollectorStats tracks deletion collector performance and activity
type CollectorStats struct {
	// Execution statistics
	TotalRuns           int64     `json:"total_runs"`
	LastRunTime         time.Time `json:"last_run_time"`
	LastRunDuration     string    `json:"last_run_duration"`
	AverageRunDuration  string    `json:"average_run_duration"`
	
	// Entity processing statistics
	EntitiesProcessed   int64 `json:"entities_processed"`
	EntitiesTransitioned int64 `json:"entities_transitioned"`
	
	// Lifecycle transition counts
	SoftDeleted         int64 `json:"soft_deleted"`
	Archived            int64 `json:"archived"`
	Purged              int64 `json:"purged"`
	Restored            int64 `json:"restored"`
	
	// Error tracking
	Errors              int64 `json:"errors"`
	LastError           string `json:"last_error,omitempty"`
	LastErrorTime       time.Time `json:"last_error_time,omitempty"`
	
	// Performance tracking
	totalRunDuration    time.Duration
}

// NewDeletionCollector creates a new deletion collector service
func NewDeletionCollector(repository models.EntityRepository, config DeletionCollectorConfig) *DeletionCollector {
	// Default configuration
	if config.Interval == 0 {
		config.Interval = 1 * time.Hour // Default: run every hour
	}
	if config.BatchSize == 0 {
		config.BatchSize = 100 // Default: process 100 entities per batch
	}
	if config.MaxRuntime == 0 {
		config.MaxRuntime = 30 * time.Minute // Default: max 30 minutes per run
	}
	if config.Concurrency == 0 {
		config.Concurrency = 4 // Default: 4 concurrent workers
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	collector := &DeletionCollector{
		repository:   repository,
		policyEngine: models.NewPolicyEngine(),
		config:       config,
		ctx:          ctx,
		cancel:       cancel,
		systemUserID: "system",
		stats:        CollectorStats{},
	}
	
	// Load default policies
	logger.Info("DeletionCollector: Loading default retention policies")
	for _, policy := range models.DefaultPolicies() {
		if err := collector.policyEngine.AddPolicy(policy); err != nil {
			logger.Error("DeletionCollector: Failed to load default policy %s: %v", policy.Name, err)
		} else {
			logger.Debug("DeletionCollector: Loaded policy: %s", policy.Name)
		}
	}
	
	return collector
}

// Start begins the deletion collector background service
func (dc *DeletionCollector) Start() error {
	if !atomic.CompareAndSwapInt32(&dc.running, 0, 1) {
		return fmt.Errorf("deletion collector is already running")
	}
	
	if !dc.config.Enabled {
		logger.Info("DeletionCollector: Service disabled by configuration")
		return nil
	}
	
	logger.Info("DeletionCollector: Starting service (interval: %v, batch size: %d, dry run: %v)", 
		dc.config.Interval, dc.config.BatchSize, dc.config.DryRun)
	
	dc.wg.Add(1)
	go dc.collectionLoop()
	
	return nil
}

// Stop gracefully shuts down the deletion collector
func (dc *DeletionCollector) Stop() error {
	if !atomic.CompareAndSwapInt32(&dc.running, 1, 0) {
		return fmt.Errorf("deletion collector is not running")
	}
	
	logger.Info("DeletionCollector: Stopping service")
	dc.cancel()
	dc.wg.Wait()
	logger.Info("DeletionCollector: Service stopped")
	
	return nil
}

// IsRunning returns true if the collector is currently active
func (dc *DeletionCollector) IsRunning() bool {
	return atomic.LoadInt32(&dc.running) == 1
}

// GetStats returns current collector statistics
func (dc *DeletionCollector) GetStats() CollectorStats {
	dc.mu.RLock()
	defer dc.mu.RUnlock()
	
	stats := dc.stats
	
	// Calculate average run duration
	if stats.TotalRuns > 0 {
		avgNanos := dc.stats.totalRunDuration.Nanoseconds() / stats.TotalRuns
		stats.AverageRunDuration = time.Duration(avgNanos).String()
	}
	
	return stats
}

// AddPolicy adds a new retention policy to the collector
func (dc *DeletionCollector) AddPolicy(policy models.RetentionPolicy) error {
	logger.Info("DeletionCollector: Adding policy: %s", policy.Name)
	return dc.policyEngine.AddPolicy(policy)
}

// RemovePolicy removes a retention policy from the collector
func (dc *DeletionCollector) RemovePolicy(name string) error {
	logger.Info("DeletionCollector: Removing policy: %s", name)
	return dc.policyEngine.RemovePolicy(name)
}

// GetPolicies returns all active retention policies
func (dc *DeletionCollector) GetPolicies() []models.RetentionPolicy {
	return dc.policyEngine.GetPolicies()
}

// RunOnce executes a single collection cycle manually
func (dc *DeletionCollector) RunOnce() error {
	logger.Info("DeletionCollector: Manual collection cycle requested")
	return dc.runCollectionCycle()
}

// collectionLoop is the main background loop
func (dc *DeletionCollector) collectionLoop() {
	defer dc.wg.Done()
	
	ticker := time.NewTicker(dc.config.Interval)
	defer ticker.Stop()
	
	// Run once immediately on startup
	if err := dc.runCollectionCycle(); err != nil {
		logger.Error("DeletionCollector: Initial collection cycle failed: %v", err)
	}
	
	for {
		select {
		case <-dc.ctx.Done():
			logger.Debug("DeletionCollector: Collection loop stopping")
			return
			
		case <-ticker.C:
			if err := dc.runCollectionCycle(); err != nil {
				logger.Error("DeletionCollector: Collection cycle failed: %v", err)
				dc.recordError(err)
			}
		}
	}
}

// runCollectionCycle executes a single collection and cleanup cycle
func (dc *DeletionCollector) runCollectionCycle() error {
	startTime := time.Now()
	
	// Create cycle context with timeout
	cycleCtx, cycleCancel := context.WithTimeout(dc.ctx, dc.config.MaxRuntime)
	defer cycleCancel()
	
	logger.Info("DeletionCollector: Starting collection cycle")
	
	// Update statistics
	dc.mu.Lock()
	dc.stats.TotalRuns++
	dc.stats.LastRunTime = startTime
	dc.mu.Unlock()
	
	// Get all entities for processing
	allEntities, err := dc.repository.List()
	if err != nil {
		return fmt.Errorf("failed to list entities: %w", err)
	}
	
	logger.Debug("DeletionCollector: Evaluating %d entities", len(allEntities))
	
	// Process entities in batches
	processed := 0
	transitioned := 0
	
	for i := 0; i < len(allEntities); i += dc.config.BatchSize {
		// Check for cancellation
		select {
		case <-cycleCtx.Done():
			logger.Warn("DeletionCollector: Collection cycle cancelled due to timeout")
			return cycleCtx.Err()
		default:
		}
		
		// Process batch
		end := i + dc.config.BatchSize
		if end > len(allEntities) {
			end = len(allEntities)
		}
		
		batch := allEntities[i:end]
		batchTransitioned, err := dc.processBatch(cycleCtx, batch)
		if err != nil {
			logger.Error("DeletionCollector: Batch processing failed: %v", err)
			dc.recordError(err)
			// Continue with next batch
		}
		
		processed += len(batch)
		transitioned += batchTransitioned
		
		logger.Debug("DeletionCollector: Processed batch %d-%d (%d transitioned)", i, end-1, batchTransitioned)
	}
	
	duration := time.Since(startTime)
	
	// Update final statistics
	dc.mu.Lock()
	dc.stats.LastRunDuration = duration.String()
	dc.stats.totalRunDuration += duration
	dc.stats.EntitiesProcessed += int64(processed)
	dc.stats.EntitiesTransitioned += int64(transitioned)
	dc.mu.Unlock()
	
	logger.Info("DeletionCollector: Collection cycle completed in %v (processed: %d, transitioned: %d)", 
		duration, processed, transitioned)
	
	return nil
}

// processBatch processes a batch of entities for lifecycle transitions
func (dc *DeletionCollector) processBatch(ctx context.Context, entities []*models.Entity) (int, error) {
	transitioned := 0
	
	// Use a worker pool for concurrent processing
	entityChan := make(chan *models.Entity, len(entities))
	resultChan := make(chan int, dc.config.Concurrency)
	errorChan := make(chan error, dc.config.Concurrency)
	
	// Start workers
	for i := 0; i < dc.config.Concurrency; i++ {
		go dc.entityWorker(ctx, entityChan, resultChan, errorChan)
	}
	
	// Send entities to workers
	for _, entity := range entities {
		select {
		case entityChan <- entity:
		case <-ctx.Done():
			close(entityChan)
			return transitioned, ctx.Err()
		}
	}
	close(entityChan)
	
	// Collect results
	for i := 0; i < dc.config.Concurrency; i++ {
		select {
		case result := <-resultChan:
			transitioned += result
		case err := <-errorChan:
			logger.Error("DeletionCollector: Worker error: %v", err)
		case <-ctx.Done():
			return transitioned, ctx.Err()
		}
	}
	
	return transitioned, nil
}

// entityWorker processes individual entities for lifecycle transitions
func (dc *DeletionCollector) entityWorker(ctx context.Context, entityChan <-chan *models.Entity, 
	resultChan chan<- int, errorChan chan<- error) {
	
	transitioned := 0
	
	for entity := range entityChan {
		select {
		case <-ctx.Done():
			resultChan <- transitioned
			return
		default:
		}
		
		if dc.processEntity(entity) {
			transitioned++
		}
	}
	
	resultChan <- transitioned
}

// processEntity evaluates and applies retention policies to a single entity
func (dc *DeletionCollector) processEntity(entity *models.Entity) bool {
	// Get applicable policies
	policies := dc.policyEngine.GetApplicablePolicies(entity)
	if len(policies) == 0 {
		return false // No policies apply
	}
	
	// Sort policies by priority (lower number = higher priority)
	sort.Slice(policies, func(i, j int) bool {
		return policies[i].Priority < policies[j].Priority
	})
	
	// Process policies in priority order
	for _, policy := range policies {
		for _, rule := range policy.Rules {
			if !rule.Enabled {
				continue
			}
			
			// Check if rule applies
			shouldApply, err := dc.policyEngine.EvaluateRule(entity, rule)
			if err != nil {
				logger.Error("DeletionCollector: Rule evaluation failed for entity %s, policy %s, rule %s: %v", 
					entity.ID, policy.Name, rule.Name, err)
				continue
			}
			
			if shouldApply {
				return dc.applyTransition(entity, policy, rule)
			}
		}
	}
	
	return false
}

// applyTransition executes a lifecycle transition on an entity
func (dc *DeletionCollector) applyTransition(entity *models.Entity, policy models.RetentionPolicy, rule models.RetentionRule) bool {
	logger.Info("DeletionCollector: Applying transition %s->%s to entity %s (policy: %s, rule: %s)", 
		rule.FromState, rule.ToState, entity.ID, policy.Name, rule.Name)
	
	if dc.config.DryRun {
		logger.Info("DeletionCollector: DRY RUN - Would transition entity %s from %s to %s", 
			entity.ID, rule.FromState, rule.ToState)
		return true // Count as transitioned for statistics
	}
	
	// Build reason with policy context
	reason := fmt.Sprintf("%s (policy: %s)", rule.Reason, policy.Name)
	
	// Execute transition
	lifecycle := entity.Lifecycle()
	var err error
	
	switch rule.ToState {
	case models.StateSoftDeleted:
		err = lifecycle.SoftDelete(dc.systemUserID, reason, policy.Name)
		if err == nil {
			dc.incrementStat("soft_deleted")
		}
		
	case models.StateActive:
		err = lifecycle.Undelete(dc.systemUserID, reason)
		if err == nil {
			dc.incrementStat("restored")
		}
		
	case models.StateArchived:
		err = lifecycle.Archive(dc.systemUserID, reason, policy.Name)
		if err == nil {
			dc.incrementStat("archived")
		}
		
	case models.StatePurged:
		err = lifecycle.Purge(dc.systemUserID, reason, policy.Name)
		if err == nil {
			dc.incrementStat("purged")
		}
		
	default:
		err = fmt.Errorf("unknown target state: %s", rule.ToState)
	}
	
	if err != nil {
		logger.Error("DeletionCollector: Failed to transition entity %s: %v", entity.ID, err)
		dc.recordError(err)
		return false
	}
	
	// Save entity changes
	if err := dc.repository.Update(entity); err != nil {
		logger.Error("DeletionCollector: Failed to save entity %s after transition: %v", entity.ID, err)
		dc.recordError(err)
		return false
	}
	
	logger.Debug("DeletionCollector: Successfully transitioned entity %s to %s", entity.ID, rule.ToState)
	return true
}

// incrementStat safely increments a specific statistic counter
func (dc *DeletionCollector) incrementStat(statName string) {
	if !dc.config.EnableMetrics {
		return
	}
	
	dc.mu.Lock()
	defer dc.mu.Unlock()
	
	switch statName {
	case "soft_deleted":
		dc.stats.SoftDeleted++
	case "archived":
		dc.stats.Archived++
	case "purged":
		dc.stats.Purged++
	case "restored":
		dc.stats.Restored++
	}
}

// recordError safely records an error in statistics
func (dc *DeletionCollector) recordError(err error) {
	if !dc.config.EnableMetrics {
		return
	}
	
	dc.mu.Lock()
	defer dc.mu.Unlock()
	
	dc.stats.Errors++
	dc.stats.LastError = err.Error()
	dc.stats.LastErrorTime = time.Now()
}

// DefaultConfig returns a reasonable default configuration
func DefaultConfig() DeletionCollectorConfig {
	return DeletionCollectorConfig{
		Enabled:       true,
		Interval:      1 * time.Hour,
		BatchSize:     100,
		MaxRuntime:    30 * time.Minute,
		DryRun:        false,
		EnableMetrics: true,
		Concurrency:   4,
	}
}