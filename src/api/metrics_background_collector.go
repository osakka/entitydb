package api

import (
	"context"
	"entitydb/models"
	"entitydb/logger"
	"fmt"
	"runtime"
	"sync"
	"time"
	"os"
	"path/filepath"
)

// BackgroundMetricsCollector collects system metrics periodically
type BackgroundMetricsCollector struct {
	collector  *MetricsCollector
	repo       models.EntityRepository
	interval   time.Duration
	ctx        context.Context
	cancel     context.CancelFunc
	lastValues map[string]float64 // Track last values for change detection
	mu         sync.RWMutex       // Protect lastValues map
}

// NewBackgroundMetricsCollector creates a new background metrics collector
func NewBackgroundMetricsCollector(repo models.EntityRepository, interval time.Duration) *BackgroundMetricsCollector {
	ctx, cancel := context.WithCancel(context.Background())
	return &BackgroundMetricsCollector{
		collector:  NewMetricsCollector(repo),
		repo:       repo,
		interval:   interval,
		ctx:        ctx,
		cancel:     cancel,
		lastValues: make(map[string]float64),
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

// collectMetrics collects all system metrics
func (b *BackgroundMetricsCollector) collectMetrics() {
	logger.Trace("Collecting system metrics...")
	
	// Memory metrics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	b.storeMetric("memory_alloc", float64(m.Alloc), "bytes", "Memory currently allocated")
	b.storeMetric("memory_total_alloc", float64(m.TotalAlloc), "bytes", "Total memory allocated")
	b.storeMetric("memory_sys", float64(m.Sys), "bytes", "Memory obtained from system")
	b.storeMetric("memory_heap_alloc", float64(m.HeapAlloc), "bytes", "Heap memory allocated")
	b.storeMetric("memory_heap_inuse", float64(m.HeapInuse), "bytes", "Heap memory in use")
	
	// GC metrics
	b.storeMetric("gc_runs", float64(m.NumGC), "count", "Number of GC runs")
	if m.NumGC > 0 {
		b.storeMetric("gc_pause_ns", float64(m.PauseNs[(m.NumGC+255)%256]), "nanoseconds", "Last GC pause duration")
	}
	
	// Goroutine metrics
	b.storeMetric("goroutines", float64(runtime.NumGoroutine()), "count", "Number of goroutines")
	
	// CPU metrics
	b.storeMetric("cpu_count", float64(runtime.NumCPU()), "count", "Number of CPUs")
	
	// Database metrics
	b.collectDatabaseMetrics()
	
	// Entity metrics
	b.collectEntityMetrics()
	
	logger.Trace("System metrics collection completed")
}

// collectDatabaseMetrics collects database-specific metrics
func (b *BackgroundMetricsCollector) collectDatabaseMetrics() {
	// Get database file stats
	storagePath := os.Getenv("ENTITYDB_STORAGE_PATH")
	if storagePath == "" {
		storagePath = "/opt/entitydb/var"
	}
	
	// Main database file
	dbPath := filepath.Join(storagePath, "entities.ebf")
	if info, err := os.Stat(dbPath); err == nil {
		b.storeMetric("database_size", float64(info.Size()), "bytes", "Database file size")
	}
	
	// WAL file - CRITICAL METRIC
	walPath := filepath.Join(storagePath, "entities.wal")
	if info, err := os.Stat(walPath); err == nil {
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
	
	// Index files
	var indexSize int64
	indexPattern := filepath.Join(storagePath, "*.idx")
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
	
	// Try to find existing metric entity first by searching for name tag
	existingEntities, err := b.repo.ListByTag(fmt.Sprintf("name:%s", name))
	var metricEntity *models.Entity
	var metricID string
	
	if err == nil && len(existingEntities) > 0 {
		// Use existing metric entity
		metricEntity = existingEntities[0]
		metricID = metricEntity.ID
		logger.Trace("Found existing metric entity: %s for metric %s", metricID, name)
	} else {
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
			logger.Error("Failed to store metric entity %s: %v", newEntity.ID, err)
			return
		}
		
		metricEntity = newEntity
		metricID = newEntity.ID
		logger.Debug("Created metric entity with UUID: %s for metric %s", metricID, name)
	}
	
	// Add temporal value tag (similar to request metrics middleware)
	valueTag := fmt.Sprintf("value:%.2f", value)
	if err := b.repo.AddTag(metricID, valueTag); err != nil {
		logger.Error("Failed to add value tag to metric %s: %v", metricID, err)
		return
	}
	
	logger.Trace("Stored metric %s with value: %.2f %s (entity: %s)", name, value, unit, metricID)
}