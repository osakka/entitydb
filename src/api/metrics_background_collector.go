package api

import (
	"context"
	"entitydb/models"
	"entitydb/config"
	"entitydb/logger"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"os"
	"path/filepath"
)

// BackgroundMetricsCollector collects system metrics periodically
type BackgroundMetricsCollector struct {
	collector    *MetricsCollector
	repo         models.EntityRepository
	config       *config.Config
	interval     time.Duration
	gentlePause  time.Duration      // Configurable pause between metric collection blocks
	ctx          context.Context
	cancel       context.CancelFunc
	lastValues   map[string]float64 // Track last values for change detection
	mu           sync.RWMutex       // Protect lastValues map
	
	// LEGENDARY FIX: Metric entity cache to prevent lookup storms
	metricCache  map[string]string  // metric name -> entity ID cache
	cacheMu      sync.RWMutex       // Protect metric cache
	
	// BAR-RAISING SOLUTION: Circuit breaker to prevent feedback loops
	failureCount      int         // Count of consecutive failures
	circuitOpen       bool        // True when circuit is open (metrics collection disabled)
	lastFailure       time.Time   // Time of last failure
	metricsInProgress int32       // Atomic flag to prevent recursion
	circuitMu       sync.RWMutex  // Protect circuit breaker state
}

// NewBackgroundMetricsCollector creates a new background metrics collector
func NewBackgroundMetricsCollector(repo models.EntityRepository, cfg *config.Config, interval time.Duration, gentlePause time.Duration) *BackgroundMetricsCollector {
	ctx, cancel := context.WithCancel(context.Background())
	return &BackgroundMetricsCollector{
		collector:    NewMetricsCollector(repo),
		repo:         repo,
		config:       cfg,
		interval:     interval,
		gentlePause:  gentlePause,
		ctx:          ctx,
		cancel:       cancel,
		lastValues:   make(map[string]float64),
		metricCache:  make(map[string]string), // Initialize cache
	}
}

// Start begins the background metrics collection
func (b *BackgroundMetricsCollector) Start() {
	logger.Info("Starting background metrics collector with interval: %v", b.interval)
	
	go func() {
		// Wait a moment for the system to fully initialize before first collection
		logger.Debug("Background metrics collector waiting 5s for system initialization")
		select {
		case <-time.After(5 * time.Second):
			// Collect metrics after initial delay
			b.collectMetrics()
		case <-b.ctx.Done():
			return
		}
		
		// Then collect periodically
		ticker := time.NewTicker(b.interval)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				b.collectMetrics()
			case <-b.ctx.Done():
				logger.Info("Background metrics collector stopped")
				return
			}
		}
	}()
}

// Stop stops the background metrics collection
func (b *BackgroundMetricsCollector) Stop() {
	b.cancel()
}

// collectMetrics collects all system metrics with gentle pacing to reduce CPU spikes
func (b *BackgroundMetricsCollector) collectMetrics() {
	// BAR-RAISING SOLUTION: Check circuit breaker first
	if b.isCircuitOpen() {
		logger.Trace("Metrics collection circuit is open - skipping collection to prevent feedback loops")
		return
	}
	
	logger.Trace("Collecting system metrics with gentle pacing...")
	
	// Memory metrics block
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	b.storeMetric("memory_alloc", float64(m.Alloc), "bytes", "Memory currently allocated")
	b.storeMetric("memory_total_alloc", float64(m.TotalAlloc), "bytes", "Total memory allocated")
	b.storeMetric("memory_sys", float64(m.Sys), "bytes", "Memory obtained from system")
	b.storeMetric("memory_heap_alloc", float64(m.HeapAlloc), "bytes", "Heap memory allocated")
	b.storeMetric("memory_heap_inuse", float64(m.HeapInuse), "bytes", "Heap memory in use")
	
	// Gentle pause between metric collection blocks to smooth CPU usage
	time.Sleep(b.gentlePause)
	
	// GC metrics block
	b.storeMetric("gc_runs", float64(m.NumGC), "count", "Number of GC runs")
	if m.NumGC > 0 {
		b.storeMetric("gc_pause_ns", float64(m.PauseNs[(m.NumGC+255)%256]), "nanoseconds", "Last GC pause duration")
	}
	
	// Goroutine metrics
	b.storeMetric("goroutines", float64(runtime.NumGoroutine()), "count", "Number of goroutines")
	
	// CPU metrics
	b.storeMetric("cpu_count", float64(runtime.NumCPU()), "count", "Number of CPUs")
	
	// Gentle pause before database metrics
	time.Sleep(b.gentlePause)
	
	// Database metrics
	b.collectDatabaseMetrics()
	
	// Gentle pause before entity metrics
	time.Sleep(b.gentlePause)
	
	// Entity metrics
	b.collectEntityMetrics()
	
	logger.Trace("Gentle system metrics collection completed")
}

// collectDatabaseMetrics collects database-specific metrics
func (b *BackgroundMetricsCollector) collectDatabaseMetrics() {
	// Main database file - use configuration (single source of truth)
	if info, err := os.Stat(b.config.DatabaseFilename); err == nil {
		b.storeMetric("database_size", float64(info.Size()), "bytes", "Database file size")
	}
	
	// WAL file - CRITICAL METRIC - use configuration (single source of truth)
	if info, err := os.Stat(b.config.WALFilename); err == nil {
		walSize := float64(info.Size())
		b.storeMetric("wal_size", walSize, "bytes", "WAL file size")
		
		// Also store WAL size in MB for easier monitoring
		b.storeMetric("wal_size_mb", walSize/(1024*1024), "MB", "WAL file size in MB")
		
		// Alert metric if WAL is getting large (>50MB is concerning, >100MB is critical)
		if walSize > 100*1024*1024 {
			b.storeMetric("wal_critical", 1, "boolean", "WAL size critical (>100MB)")
		} else if walSize > 50*1024*1024 {
			b.storeMetric("wal_warning", 1, "boolean", "WAL size warning (>50MB)")
		} else {
			b.storeMetric("wal_critical", 0, "boolean", "WAL size critical (>100MB)")
			b.storeMetric("wal_warning", 0, "boolean", "WAL size warning (>50MB)")
		}
	}
	
	// Index files - use configuration data path
	var indexSize int64
	indexPattern := filepath.Join(b.config.DataPath, "*.idx")
	if matches, err := filepath.Glob(indexPattern); err == nil {
		for _, match := range matches {
			if info, err := os.Stat(match); err == nil {
				indexSize += info.Size()
			}
		}
		b.storeMetric("index_size", float64(indexSize), "bytes", "Total index files size")
	}
}

// collectEntityMetrics collects entity-specific metrics
func (b *BackgroundMetricsCollector) collectEntityMetrics() {
	// Count entities by type
	entityTypes := []string{"user", "metric", "issue", "workspace", "relationship", "dataset"}
	
	for _, entityType := range entityTypes {
		entities, err := b.repo.ListByTag(fmt.Sprintf("type:%s", entityType))
		if err == nil {
			b.storeMetric(fmt.Sprintf("entity_count_%s", entityType), float64(len(entities)), "count", 
				fmt.Sprintf("Number of %s entities", entityType))
		}
	}
	
	// Total entity count
	allEntities, err := b.repo.List()
	if err == nil {
		b.storeMetric("entity_count_total", float64(len(allEntities)), "count", "Total number of entities")
	}
	
	// Count entities created in different time periods
	now := time.Now()
	todayCount := 0
	weekCount := 0
	monthCount := 0
	
	for _, entity := range allEntities {
		created := time.Unix(0, entity.CreatedAt)
		if created.After(now.AddDate(0, 0, -1)) {
			todayCount++
		}
		if created.After(now.AddDate(0, 0, -7)) {
			weekCount++
		}
		if created.After(now.AddDate(0, -1, 0)) {
			monthCount++
		}
	}
	
	b.storeMetric("entities_created_today", float64(todayCount), "count", "Entities created today")
	b.storeMetric("entities_created_week", float64(weekCount), "count", "Entities created this week")
	b.storeMetric("entities_created_month", float64(monthCount), "count", "Entities created this month")
}

// storeMetric stores a metric value using time-series optimized storage pattern
func (b *BackgroundMetricsCollector) storeMetric(name string, value float64, unit string, description string) {
	// Emergency check: prevent metrics feedback loops causing WAL corruption
	// Use atomic flag to detect if we're already storing metrics
	if !atomic.CompareAndSwapInt32(&b.metricsInProgress, 0, 1) {
		logger.Trace("Aborting metric storage for %s - already in metrics operation", name)
		return
	}
	defer atomic.StoreInt32(&b.metricsInProgress, 0)
	// Check if value has changed using change detection
	b.mu.RLock()
	lastValue, exists := b.lastValues[name]
	b.mu.RUnlock()
	
	// Only store if value changed or if it's the first time
	if exists && lastValue == value {
		logger.Trace("Metric %s unchanged (%.2f), skipping storage", name, value)
		return
	}
	
	// Update last value
	b.mu.Lock()
	b.lastValues[name] = value
	b.mu.Unlock()
	
	logger.Debug("Metric %s changed from %.2f to %.2f, storing", name, lastValue, value)
	
	// Create metric entity using UUID architecture with system user ownership
	additionalTags := []string{
		fmt.Sprintf("name:%s", name),
		fmt.Sprintf("unit:%s", unit),
		fmt.Sprintf("description:%s", description),
		"retention:count:100", // Keep last 100 values
		"retention:period:3600", // Keep for 1 hour
	}
	
	// LEGENDARY FIX: Use cached metric lookup to prevent database lookup storms
	// Check cache first to avoid expensive ListByTag operations
	metricID, exists := b.getMetricFromCache(name)
	var metricEntity *models.Entity
	
	if exists {
		// Get cached metric entity
		if cachedEntity, err := b.repo.GetByID(metricID); err == nil {
			metricEntity = cachedEntity
			logger.Trace("Found cached metric entity: %s for metric %s", metricID, name)
		} else {
			// Cache is stale, remove and create new
			b.removeMetricFromCache(name)
			exists = false
		}
	}
	
	if !exists {
		// Create new metric entity using UUID architecture
		newEntity, err := models.NewEntityWithMandatoryTags(
			"metric",                    // entityType
			"system",                    // dataset
			models.SystemUserID,         // createdBy (system user)
			additionalTags,             // additional tags
		)
		if err != nil {
			logger.Error("Failed to create metric entity for %s: %v", name, err)
			return
		}
		
		if err := b.repo.Create(newEntity); err != nil {
			// BAR-RAISING: Detect recursion guard errors and circuit break immediately
			if strings.Contains(err.Error(), "recursion guard") {
				logger.Warn("Metrics collection blocked by recursion guard - opening circuit breaker: %v", err)
				b.forceCircuitOpen()
				return
			}
			logger.Error("Failed to store metric entity %s: %v", newEntity.ID, err)
			b.recordFailure() // Track failures for circuit breaker
			return
		}
		
		metricEntity = newEntity
		metricID = newEntity.ID
		
		// Cache the new metric entity to prevent future lookup storms
		b.cacheMetric(name, metricID)
		logger.Debug("Created and cached metric entity with UUID: %s for metric %s", metricID, name)
	}
	
	// ATOMIC TAG FIX: Add temporal value tag with explicit timestamp to prevent explosion
	valueTag := fmt.Sprintf("value:%.2f", value)
	now := time.Now().UnixNano()
	timestampedValueTag := fmt.Sprintf("%d|%s", now, valueTag)
	
	// Add to existing tags atomically
	metricEntity.Tags = append(metricEntity.Tags, timestampedValueTag)
	if err := b.repo.Update(metricEntity); err != nil {
		// BAR-RAISING: Detect recursion guard errors and circuit break immediately
		if strings.Contains(err.Error(), "recursion guard") {
			logger.Warn("Metrics update blocked by recursion guard - opening circuit breaker: %v", err)
			b.forceCircuitOpen()
			return
		}
		logger.Error("Failed to update metric entity %s: %v", metricID, err)
		b.recordFailure() // Track failures for circuit breaker
		return
	}
	
	// BAR-RAISING: Record successful operation to reset failure count
	b.recordSuccess()
	logger.Trace("Stored metric %s with value: %.2f %s (entity: %s)", name, value, unit, metricID)
}

// LEGENDARY FIX: Metric cache methods to prevent database lookup storms

// getMetricFromCache retrieves a cached metric entity ID by name
func (b *BackgroundMetricsCollector) getMetricFromCache(name string) (string, bool) {
	b.cacheMu.RLock()
	defer b.cacheMu.RUnlock()
	entityID, exists := b.metricCache[name]
	return entityID, exists
}

// cacheMetric stores a metric entity ID in cache by name
func (b *BackgroundMetricsCollector) cacheMetric(name, entityID string) {
	b.cacheMu.Lock()
	defer b.cacheMu.Unlock()
	b.metricCache[name] = entityID
	logger.Trace("Cached metric %s -> entity %s", name, entityID)
}

// removeMetricFromCache removes a stale metric from cache
func (b *BackgroundMetricsCollector) removeMetricFromCache(name string) {
	b.cacheMu.Lock()
	defer b.cacheMu.Unlock()
	delete(b.metricCache, name)
	logger.Trace("Removed stale metric cache for %s", name)
}

// BAR-RAISING SOLUTION: Circuit breaker methods to prevent feedback loops

// isCircuitOpen checks if the circuit breaker is open (metrics collection disabled)
func (b *BackgroundMetricsCollector) isCircuitOpen() bool {
	b.circuitMu.RLock()
	defer b.circuitMu.RUnlock()
	
	// Circuit is open if we have too many failures
	if b.failureCount >= 5 {
		// Auto-recovery after 5 minutes
		if time.Since(b.lastFailure) > 5*time.Minute {
			b.circuitMu.RUnlock()
			b.circuitMu.Lock()
			b.failureCount = 0
			b.circuitOpen = false
			b.circuitMu.Unlock()
			logger.Info("Circuit breaker auto-recovery: reopening metrics collection after 5 minutes")
			b.circuitMu.RLock()
		}
	}
	
	return b.circuitOpen
}

// recordFailure increments failure count and may open the circuit
func (b *BackgroundMetricsCollector) recordFailure() {
	b.circuitMu.Lock()
	defer b.circuitMu.Unlock()
	
	b.failureCount++
	b.lastFailure = time.Now()
	
	if b.failureCount >= 5 && !b.circuitOpen {
		b.circuitOpen = true
		logger.Warn("CIRCUIT BREAKER OPENED: Disabling metrics collection after %d consecutive failures to prevent feedback loops", b.failureCount)
	}
}

// recordSuccess resets failure count and closes circuit if open
func (b *BackgroundMetricsCollector) recordSuccess() {
	b.circuitMu.Lock()
	defer b.circuitMu.Unlock()
	
	if b.failureCount > 0 || b.circuitOpen {
		logger.Info("Circuit breaker: Successful operation - resetting failure count (was %d)", b.failureCount)
		b.failureCount = 0
		b.circuitOpen = false
	}
}

// forceCircuitOpen immediately opens the circuit breaker to prevent further metrics collection
func (b *BackgroundMetricsCollector) forceCircuitOpen() {
	b.circuitMu.Lock()
	defer b.circuitMu.Unlock()
	
	b.circuitOpen = true
	b.failureCount = 10 // Set high failure count to keep circuit open longer
	b.lastFailure = time.Now()
	logger.Error("CIRCUIT BREAKER FORCE OPENED: Metrics collection disabled due to recursion guard protection")
}