package binary

import (
	"context"
	"entitydb/logger"
	"entitydb/models"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// MetricPoint represents a single metric data point with labels
type MetricPoint struct {
	Name        string
	Value       float64
	Unit        string
	Description string
	Labels      map[string]string
	Timestamp   time.Time
	Priority    MetricPriority
}

// MetricPriority determines handling when channels are full
type MetricPriority int

const (
	PriorityLow MetricPriority = iota
	PriorityMedium
	PriorityHigh
	PriorityCritical
)

// AsyncMetricsCollector provides non-blocking metrics collection
type AsyncMetricsCollector struct {
	// Core components
	systemContext *SystemContext
	directRepo    DirectRepository
	
	// Channel-based communication
	metricChan    chan MetricPoint
	shutdownChan  chan struct{}
	wg            sync.WaitGroup
	
	// Configuration
	bufferSize    int
	workerCount   int
	batchSize     int
	flushInterval time.Duration
	
	// State management
	running       int32
	droppedCount  int64
	processedCount int64
	
	// In-memory aggregation
	aggregator    *MetricAggregator
	aggregateMu   sync.RWMutex
	
	// Metrics about metrics (meta-metrics)
	metaMetrics   *MetaMetrics
}

// MetricAggregator handles in-memory metric aggregation
type MetricAggregator struct {
	data      map[string]*AggregatedMetric
	mutex     sync.RWMutex
	maxSize   int
	lastFlush time.Time
}

// AggregatedMetric represents aggregated metric data
type AggregatedMetric struct {
	Name        string
	Unit        string
	Description string
	Values      []float64
	Labels      map[string]string
	LastUpdate  time.Time
	Count       int64
	Sum         float64
	Min         float64
	Max         float64
}

// MetaMetrics tracks metrics collection performance
type MetaMetrics struct {
	CollectionLatency   time.Duration
	PersistenceLatency  time.Duration
	ChannelUtilization  float64
	DropRate           float64
	ThroughputPerSecond float64
}

// SystemContext provides pre-authenticated context for metrics operations
type SystemContext struct {
	UserID    string
	SessionID string
	IsSystem  bool
	mutex     sync.RWMutex
}

// DirectRepository interface for non-instrumented repository operations
type DirectRepository interface {
	CreateDirect(entity *models.Entity) error
	GetByIDDirect(id string) (*models.Entity, error)
	ListByTagDirect(tag string) ([]*models.Entity, error)
	AddTagDirect(entityID, tag string) error
}

// NewAsyncMetricsCollector creates a new async metrics collector
func NewAsyncMetricsCollector(repo models.EntityRepository, config AsyncMetricsConfig) (*AsyncMetricsCollector, error) {
	// Create system context
	systemContext, err := NewSystemContext(repo)
	if err != nil {
		return nil, fmt.Errorf("failed to create system context: %v", err)
	}
	
	// Create direct repository wrapper
	directRepo := NewDirectRepositoryWrapper(repo)
	
	collector := &AsyncMetricsCollector{
		systemContext: systemContext,
		directRepo:    directRepo,
		metricChan:    make(chan MetricPoint, config.BufferSize),
		shutdownChan:  make(chan struct{}),
		bufferSize:    config.BufferSize,
		workerCount:   config.WorkerCount,
		batchSize:     config.BatchSize,
		flushInterval: config.FlushInterval,
		aggregator:    NewMetricAggregator(config.MaxAggregationSize),
		metaMetrics:   &MetaMetrics{},
	}
	
	return collector, nil
}

// AsyncMetricsConfig holds configuration for the async metrics collector
type AsyncMetricsConfig struct {
	BufferSize         int
	WorkerCount        int
	BatchSize          int
	FlushInterval      time.Duration
	MaxAggregationSize int
	EnableMetaMetrics  bool
}

// DefaultAsyncMetricsConfig returns default configuration
func DefaultAsyncMetricsConfig() AsyncMetricsConfig {
	return AsyncMetricsConfig{
		BufferSize:         10000,
		WorkerCount:        4,
		BatchSize:          100,
		FlushInterval:      30 * time.Second,
		MaxAggregationSize: 5000,
		EnableMetaMetrics:  true,
	}
}

// Start begins the async metrics collection
func (amc *AsyncMetricsCollector) Start(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&amc.running, 0, 1) {
		return fmt.Errorf("metrics collector already running")
	}
	
	logger.Info("Starting async metrics collector with %d workers", amc.workerCount)
	
	// Start worker goroutines
	for i := 0; i < amc.workerCount; i++ {
		amc.wg.Add(1)
		go amc.worker(ctx, i)
	}
	
	// Start aggregation flusher
	amc.wg.Add(1)
	go amc.aggregationFlusher(ctx)
	
	// Start meta-metrics collector
	amc.wg.Add(1)
	go amc.metaMetricsCollector(ctx)
	
	return nil
}

// Stop stops the async metrics collection
func (amc *AsyncMetricsCollector) Stop() error {
	if !atomic.CompareAndSwapInt32(&amc.running, 1, 0) {
		return fmt.Errorf("metrics collector not running")
	}
	
	logger.Info("Stopping async metrics collector")
	
	close(amc.shutdownChan)
	amc.wg.Wait()
	
	// Flush remaining metrics
	amc.flushAggregatedMetrics()
	
	logger.Info("Async metrics collector stopped. Processed: %d, Dropped: %d", 
		atomic.LoadInt64(&amc.processedCount), 
		atomic.LoadInt64(&amc.droppedCount))
	
	return nil
}

// CollectMetric collects a metric point asynchronously
func (amc *AsyncMetricsCollector) CollectMetric(name string, value float64, unit string, description string, labels map[string]string) {
	if atomic.LoadInt32(&amc.running) == 0 {
		return
	}
	
	metric := MetricPoint{
		Name:        name,
		Value:       value,
		Unit:        unit,
		Description: description,
		Labels:      labels,
		Timestamp:   time.Now(),
		Priority:    amc.determinePriority(name),
	}
	
	select {
	case amc.metricChan <- metric:
		// Successfully queued
	default:
		// Channel full, handle based on priority
		if metric.Priority >= PriorityHigh {
			// For high priority metrics, try to drop a low priority metric
			amc.tryDropLowPriorityMetric()
			select {
			case amc.metricChan <- metric:
				// Successfully queued after making space
			default:
				// Still full, drop this metric
				atomic.AddInt64(&amc.droppedCount, 1)
				logger.Trace("Dropped high priority metric due to full channel: %s", name)
			}
		} else {
			// Drop low/medium priority metrics immediately
			atomic.AddInt64(&amc.droppedCount, 1)
			logger.Trace("Dropped metric due to full channel: %s", name)
		}
	}
}

// determinePriority determines the priority of a metric based on its name
func (amc *AsyncMetricsCollector) determinePriority(name string) MetricPriority {
	switch {
	case strings.Contains(name, "error") || strings.Contains(name, "failure"):
		return PriorityCritical
	case strings.Contains(name, "authentication") || strings.Contains(name, "security"):
		return PriorityHigh
	case strings.Contains(name, "cache") || strings.Contains(name, "performance"):
		return PriorityMedium
	default:
		return PriorityLow
	}
}

// tryDropLowPriorityMetric attempts to drop a low priority metric from the channel
func (amc *AsyncMetricsCollector) tryDropLowPriorityMetric() {
	// This is a best-effort attempt - we can't easily peek into the channel
	// In practice, the priority system works by controlling what gets dropped
	// when the channel is full
}

// worker processes metrics from the channel
func (amc *AsyncMetricsCollector) worker(ctx context.Context, workerID int) {
	defer amc.wg.Done()
	
	logger.Debug("Metrics worker %d started", workerID)
	
	batch := make([]MetricPoint, 0, amc.batchSize)
	ticker := time.NewTicker(amc.flushInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			logger.Debug("Metrics worker %d stopping due to context cancellation", workerID)
			amc.processBatch(batch)
			return
			
		case <-amc.shutdownChan:
			logger.Debug("Metrics worker %d stopping due to shutdown signal", workerID)
			amc.processBatch(batch)
			return
			
		case metric := <-amc.metricChan:
			batch = append(batch, metric)
			
			if len(batch) >= amc.batchSize {
				amc.processBatch(batch)
				batch = batch[:0] // Reset slice but keep capacity
			}
			
		case <-ticker.C:
			if len(batch) > 0 {
				amc.processBatch(batch)
				batch = batch[:0] // Reset slice but keep capacity
			}
		}
	}
}

// processBatch processes a batch of metrics
func (amc *AsyncMetricsCollector) processBatch(batch []MetricPoint) {
	if len(batch) == 0 {
		return
	}
	
	startTime := time.Now()
	
	// Add to aggregator
	amc.aggregateMu.Lock()
	for _, metric := range batch {
		amc.aggregator.AddMetric(metric)
	}
	amc.aggregateMu.Unlock()
	
	atomic.AddInt64(&amc.processedCount, int64(len(batch)))
	
	processingTime := time.Since(startTime)
	logger.Trace("Processed batch of %d metrics in %v", len(batch), processingTime)
	
	// Update meta-metrics
	amc.metaMetrics.CollectionLatency = processingTime
}

// aggregationFlusher periodically flushes aggregated metrics to storage
func (amc *AsyncMetricsCollector) aggregationFlusher(ctx context.Context) {
	defer amc.wg.Done()
	
	ticker := time.NewTicker(amc.flushInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-amc.shutdownChan:
			return
		case <-ticker.C:
			amc.flushAggregatedMetrics()
		}
	}
}

// flushAggregatedMetrics persists aggregated metrics to storage
func (amc *AsyncMetricsCollector) flushAggregatedMetrics() {
	startTime := time.Now()
	
	amc.aggregateMu.Lock()
	metricsToFlush := amc.aggregator.GetAndClearMetrics()
	amc.aggregateMu.Unlock()
	
	if len(metricsToFlush) == 0 {
		return
	}
	
	successCount := 0
	for _, metric := range metricsToFlush {
		if err := amc.persistAggregatedMetric(metric); err != nil {
			logger.Warn("Failed to persist aggregated metric %s: %v", metric.Name, err)
		} else {
			successCount++
		}
	}
	
	flushDuration := time.Since(startTime)
	amc.metaMetrics.PersistenceLatency = flushDuration
	
	logger.Debug("Flushed %d/%d aggregated metrics in %v", successCount, len(metricsToFlush), flushDuration)
}

// persistAggregatedMetric persists a single aggregated metric
func (amc *AsyncMetricsCollector) persistAggregatedMetric(metric *AggregatedMetric) error {
	// Find or create metric entity using direct repository operations
	existingEntities, err := amc.directRepo.ListByTagDirect(fmt.Sprintf("name:%s", metric.Name))
	var metricEntity *models.Entity
	var metricID string
	
	if err == nil && len(existingEntities) > 0 {
		metricEntity = existingEntities[0]
		metricID = metricEntity.ID
	} else {
		// Create new metric entity
		additionalTags := []string{
			"name:" + metric.Name,
			"unit:" + metric.Unit,
			"description:" + metric.Description,
			"retention:count:1000",
			"retention:period:43200", // 12 hours
		}
		
		// Add label tags
		for k, v := range metric.Labels {
			additionalTags = append(additionalTags, "label:"+k+":"+v)
		}
		
		newEntity, err := models.NewEntityWithMandatoryTags(
			"metric",
			"system",
			amc.systemContext.UserID,
			additionalTags,
		)
		if err != nil {
			return fmt.Errorf("failed to create metric entity: %v", err)
		}
		
		if err := amc.directRepo.CreateDirect(newEntity); err != nil {
			return fmt.Errorf("failed to store metric entity: %v", err)
		}
		
		metricEntity = newEntity
		metricID = newEntity.ID
	}
	
	// Create value tags for aggregated data
	valueTags := []string{
		fmt.Sprintf("value:%.2f", metric.Sum),
		fmt.Sprintf("count:%d", metric.Count),
		fmt.Sprintf("avg:%.2f", metric.Sum/float64(metric.Count)),
		fmt.Sprintf("min:%.2f", metric.Min),
		fmt.Sprintf("max:%.2f", metric.Max),
	}
	
	// Add sorted labels to value tags for dimensional data
	if len(metric.Labels) > 0 {
		var sortedKeys []string
		for k := range metric.Labels {
			sortedKeys = append(sortedKeys, k)
		}
		sort.Strings(sortedKeys)
		
		var labelParts []string
		for _, k := range sortedKeys {
			labelParts = append(labelParts, fmt.Sprintf("%s=%s", k, metric.Labels[k]))
		}
		labelString := ":" + strings.Join(labelParts, ":")
		
		for i, valueTag := range valueTags {
			valueTags[i] = valueTag + labelString
		}
	}
	
	// Add temporal value tags to metric entity
	for _, valueTag := range valueTags {
		if err := amc.directRepo.AddTagDirect(metricID, valueTag); err != nil {
			logger.Warn("Failed to add value tag %s to metric %s: %v", valueTag, metricID, err)
		}
	}
	
	return nil
}

// metaMetricsCollector collects metrics about the metrics system itself
func (amc *AsyncMetricsCollector) metaMetricsCollector(ctx context.Context) {
	defer amc.wg.Done()
	
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-amc.shutdownChan:
			return
		case <-ticker.C:
			amc.collectMetaMetrics()
		}
	}
}

// collectMetaMetrics collects meta-metrics about the collection system
func (amc *AsyncMetricsCollector) collectMetaMetrics() {
	// Channel utilization
	channelLen := float64(len(amc.metricChan))
	channelCap := float64(cap(amc.metricChan))
	utilization := (channelLen / channelCap) * 100
	
	// Drop rate
	processed := atomic.LoadInt64(&amc.processedCount)
	dropped := atomic.LoadInt64(&amc.droppedCount)
	total := processed + dropped
	dropRate := float64(0)
	if total > 0 {
		dropRate = (float64(dropped) / float64(total)) * 100
	}
	
	// Update meta-metrics
	amc.metaMetrics.ChannelUtilization = utilization
	amc.metaMetrics.DropRate = dropRate
	
	logger.Debug("Metrics meta-metrics - Channel utilization: %.1f%%, Drop rate: %.2f%%, Processed: %d, Dropped: %d",
		utilization, dropRate, processed, dropped)
}

// NewMetricAggregator creates a new metric aggregator
func NewMetricAggregator(maxSize int) *MetricAggregator {
	return &MetricAggregator{
		data:    make(map[string]*AggregatedMetric),
		maxSize: maxSize,
	}
}

// AddMetric adds a metric to the aggregator
func (ma *MetricAggregator) AddMetric(metric MetricPoint) {
	ma.mutex.Lock()
	defer ma.mutex.Unlock()
	
	key := ma.generateKey(metric)
	
	if existing, ok := ma.data[key]; ok {
		// Update existing metric
		existing.Values = append(existing.Values, metric.Value)
		existing.Count++
		existing.Sum += metric.Value
		if metric.Value < existing.Min {
			existing.Min = metric.Value
		}
		if metric.Value > existing.Max {
			existing.Max = metric.Value
		}
		existing.LastUpdate = metric.Timestamp
	} else {
		// Create new aggregated metric
		if len(ma.data) >= ma.maxSize {
			// Remove oldest metric (simple LRU)
			ma.evictOldest()
		}
		
		ma.data[key] = &AggregatedMetric{
			Name:        metric.Name,
			Unit:        metric.Unit,
			Description: metric.Description,
			Values:      []float64{metric.Value},
			Labels:      metric.Labels,
			LastUpdate:  metric.Timestamp,
			Count:       1,
			Sum:         metric.Value,
			Min:         metric.Value,
			Max:         metric.Value,
		}
	}
}

// generateKey generates a unique key for the metric
func (ma *MetricAggregator) generateKey(metric MetricPoint) string {
	key := metric.Name
	
	if len(metric.Labels) > 0 {
		var sortedKeys []string
		for k := range metric.Labels {
			sortedKeys = append(sortedKeys, k)
		}
		sort.Strings(sortedKeys)
		
		var labelParts []string
		for _, k := range sortedKeys {
			labelParts = append(labelParts, fmt.Sprintf("%s=%s", k, metric.Labels[k]))
		}
		key += ":" + strings.Join(labelParts, ":")
	}
	
	return key
}

// evictOldest removes the oldest metric from the aggregator
func (ma *MetricAggregator) evictOldest() {
	var oldestKey string
	var oldestTime time.Time
	
	for key, metric := range ma.data {
		if oldestKey == "" || metric.LastUpdate.Before(oldestTime) {
			oldestKey = key
			oldestTime = metric.LastUpdate
		}
	}
	
	if oldestKey != "" {
		delete(ma.data, oldestKey)
	}
}

// GetAndClearMetrics returns all aggregated metrics and clears the aggregator
func (ma *MetricAggregator) GetAndClearMetrics() []*AggregatedMetric {
	ma.mutex.Lock()
	defer ma.mutex.Unlock()
	
	metrics := make([]*AggregatedMetric, 0, len(ma.data))
	for _, metric := range ma.data {
		metrics = append(metrics, metric)
	}
	
	// Clear the data
	ma.data = make(map[string]*AggregatedMetric)
	ma.lastFlush = time.Now()
	
	return metrics
}

// NewSystemContext creates a system context for metrics operations
func NewSystemContext(repo models.EntityRepository) (*SystemContext, error) {
	// Get system user - this is resolved once at startup, preventing deadlocks
	systemUser, err := repo.GetByID(models.SystemUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get system user: %v", err)
	}
	
	return &SystemContext{
		UserID:   systemUser.ID,
		IsSystem: true,
	}, nil
}

// GetUserID returns the system user ID
func (sc *SystemContext) GetUserID() string {
	sc.mutex.RLock()
	defer sc.mutex.RUnlock()
	return sc.UserID
}

// IsSystemUser returns true if this is a system context
func (sc *SystemContext) IsSystemUser() bool {
	sc.mutex.RLock()
	defer sc.mutex.RUnlock()
	return sc.IsSystem
}