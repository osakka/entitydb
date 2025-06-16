package binary

import (
	"entitydb/logger"
	"entitydb/models"
	"strconv"
	"strings"
	"time"
)

// StorageMetrics tracks storage operation metrics using async collection
type StorageMetrics struct {
	repo models.EntityRepository
	asyncCollector *AsyncMetricsCollector // New async collector
}

// NewStorageMetrics creates a new storage metrics tracker
func NewStorageMetrics(repo models.EntityRepository) *StorageMetrics {
	return &StorageMetrics{repo: repo}
}

// SetAsyncCollector sets the async metrics collector
func (m *StorageMetrics) SetAsyncCollector(collector *AsyncMetricsCollector) {
	m.asyncCollector = collector
	logger.Debug("StorageMetrics: async collector configured")
}

// TrackRead tracks a read operation
func (m *StorageMetrics) TrackRead(operation string, size int64, duration time.Duration, err error) {
	m.storeMetric("storage_read_duration_ms",
		float64(duration.Milliseconds()),
		"milliseconds",
		"Storage read operation duration",
		map[string]string{
			"operation": operation,
			"size_bucket": m.getSizeBucket(size),
			"success": strconv.FormatBool(err == nil),
		})
		
	if err == nil {
		m.storeMetric("storage_read_bytes",
			float64(size),
			"bytes",
			"Bytes read from storage",
			map[string]string{
				"operation": operation,
			})
	}
}

// TrackWrite tracks a write operation
func (m *StorageMetrics) TrackWrite(operation string, size int64, duration time.Duration, err error) {
	m.storeMetric("storage_write_duration_ms",
		float64(duration.Milliseconds()),
		"milliseconds",
		"Storage write operation duration",
		map[string]string{
			"operation": operation,
			"size_bucket": m.getSizeBucket(size),
			"success": strconv.FormatBool(err == nil),
		})
		
	if err == nil {
		m.storeMetric("storage_write_bytes",
			float64(size),
			"bytes",
			"Bytes written to storage",
			map[string]string{
				"operation": operation,
			})
	}
	
	// Track slow operations
	if duration > 100*time.Millisecond {
		logger.Warn("Slow storage write: operation=%s, duration=%v, size=%d", operation, duration, size)
	}
}

// TrackIndexLookup tracks index lookup performance
func (m *StorageMetrics) TrackIndexLookup(indexType string, operation string, duration time.Duration, found bool, resultCount int) {
	m.storeMetric("index_lookup_duration_ms",
		float64(duration.Milliseconds()),
		"milliseconds",
		"Index lookup duration",
		map[string]string{
			"index_type": indexType,
			"operation": operation,
			"found": strconv.FormatBool(found),
		})
		
	if found {
		m.storeMetric("index_lookup_results",
			float64(resultCount),
			"count",
			"Index lookup result count",
			map[string]string{
				"index_type": indexType,
				"operation": operation,
			})
	}
}

// TrackWALOperation tracks WAL operations
func (m *StorageMetrics) TrackWALOperation(operation string, duration time.Duration, size int64, err error) {
	m.storeMetric("wal_operation_duration_ms",
		float64(duration.Milliseconds()),
		"milliseconds",
		"WAL operation duration",
		map[string]string{
			"operation": operation,
			"success": strconv.FormatBool(err == nil),
		})
		
	if operation == "flush" || operation == "checkpoint" {
		m.storeMetric("wal_flush_duration_ms",
			float64(duration.Milliseconds()),
			"milliseconds",
			"WAL flush duration",
			map[string]string{
				"trigger": operation,
			})
			
		if err == nil && size > 0 {
			m.storeMetric("wal_flush_bytes",
				float64(size),
				"bytes",
				"Bytes flushed from WAL",
				map[string]string{
					"trigger": operation,
				})
		}
	}
}

// TrackCompression tracks compression operations
func (m *StorageMetrics) TrackCompression(originalSize, compressedSize int64, duration time.Duration) {
	if originalSize > 0 {
		ratio := float64(compressedSize) / float64(originalSize)
		m.storeMetric("compression_ratio",
			ratio,
			"ratio",
			"Compression ratio (compressed/original)",
			map[string]string{
				"size_bucket": m.getSizeBucket(originalSize),
			})
			
		savedBytes := originalSize - compressedSize
		m.storeMetric("compression_saved_bytes",
			float64(savedBytes),
			"bytes",
			"Bytes saved by compression",
			map[string]string{
				"size_bucket": m.getSizeBucket(originalSize),
			})
	}
	
	m.storeMetric("compression_duration_ms",
		float64(duration.Milliseconds()),
		"milliseconds",
		"Compression operation duration",
		map[string]string{
			"size_bucket": m.getSizeBucket(originalSize),
		})
}

// TrackCacheOperation tracks cache hit/miss
func (m *StorageMetrics) TrackCacheOperation(cacheType string, hit bool) {
	if hit {
		m.storeMetric("storage_cache_hits",
			1,
			"count",
			"Storage cache hits",
			map[string]string{
				"cache_type": cacheType,
			})
	} else {
		m.storeMetric("storage_cache_misses",
			1,
			"count",
			"Storage cache misses",
			map[string]string{
				"cache_type": cacheType,
			})
	}
}

// getSizeBucket returns a bucket label for the size
func (m *StorageMetrics) getSizeBucket(size int64) string {
	switch {
	case size < 1024:
		return "<1KB"
	case size < 10*1024:
		return "1KB-10KB"
	case size < 100*1024:
		return "10KB-100KB"
	case size < 1024*1024:
		return "100KB-1MB"
	case size < 10*1024*1024:
		return "1MB-10MB"
	default:
		return ">10MB"
	}
}

// isMetricEntity checks if an entity is a metrics entity to avoid recursion
func isMetricEntity(entity *models.Entity) bool {
	if entity == nil || entity.Tags == nil {
		return false
	}
	
	// Check if entity has type:metric tag
	for _, tag := range entity.Tags {
		// Handle both timestamped and non-timestamped tags
		if strings.Contains(tag, "type:metric") || strings.Contains(tag, "name:") {
			return true
		}
	}
	return false
}

// storeMetric stores a metric value with labels using async collection
func (m *StorageMetrics) storeMetric(name string, value float64, unit string, description string, labels map[string]string) {
	// Use async collector if available
	if m.asyncCollector != nil {
		m.asyncCollector.CollectMetric(name, value, unit, description, labels)
		logger.Trace("StorageMetrics.storeMetric: queued metric %s = %.2f via async collector", name, value)
		return
	}
	
	// Fallback: skip metrics collection if no async collector (prevents deadlocks)
	logger.Trace("StorageMetrics.storeMetric: no async collector available, skipping metric %s", name)
}

// Global instance
var storageMetrics *StorageMetrics

// Global async metrics collector
var globalAsyncCollector *AsyncMetricsCollector

// DEPRECATED: Global flag to disable storage metrics (replaced by async system)
var storageMetricsDisabled = false // Now enabled by default with async collection

// InitStorageMetrics initializes the global storage metrics
func InitStorageMetrics(repo models.EntityRepository) {
	storageMetrics = NewStorageMetrics(repo)
}

// InitAsyncStorageMetrics initializes storage metrics with async collection
func InitAsyncStorageMetrics(repo models.EntityRepository, asyncCollector *AsyncMetricsCollector) {
	storageMetrics = NewStorageMetrics(repo)
	storageMetrics.SetAsyncCollector(asyncCollector)
	globalAsyncCollector = asyncCollector
	logger.Info("Storage metrics initialized with async collection")
}

// GetStorageMetrics returns the global storage metrics instance
func GetStorageMetrics() *StorageMetrics {
	return storageMetrics
}

// GetGlobalAsyncCollector returns the global async metrics collector
func GetGlobalAsyncCollector() *AsyncMetricsCollector {
	return globalAsyncCollector
}

// SetRepository updates the repository for the global storage metrics
// This is used when metrics need to be initialized before the repository is available
func SetStorageMetricsRepository(repo models.EntityRepository) {
	if storageMetrics != nil {
		storageMetrics.repo = repo
		logger.Info("Updated storage metrics repository")
	}
}