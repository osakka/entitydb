package binary

import (
	"entitydb/logger"
	"entitydb/models"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// StorageMetrics tracks storage operation metrics
type StorageMetrics struct {
	repo models.EntityRepository
}

// NewStorageMetrics creates a new storage metrics tracker
func NewStorageMetrics(repo models.EntityRepository) *StorageMetrics {
	return &StorageMetrics{repo: repo}
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

// storeMetric stores a metric value with labels
func (m *StorageMetrics) storeMetric(name string, value float64, unit string, description string, labels map[string]string) {
	// Check if repository is available
	if m.repo == nil {
		logger.Trace("StorageMetrics.storeMetric: repository is nil, skipping %s", name)
		return
	}
	
	// Use single entity per metric name (not per label combination)
	metricID := "metric_" + name
	
	logger.Trace("StorageMetrics.storeMetric: storing metric %s = %.2f", metricID, value)
	
	// Check if metric exists
	entity, err := m.repo.GetByID(metricID)
	if err != nil {
		// Create new metric entity
		tags := []string{
			"type:metric",
			"dataset:system",
			"name:" + name,
			"unit:" + unit,
			"description:" + description,
			"retention:count:1000", 
			"retention:period:43200",
		}
		
		newEntity := &models.Entity{
			ID:      metricID,
			Tags:    tags,
			Content: []byte{},
		}
		
		if err := m.repo.Create(newEntity); err != nil {
			logger.Warn("Failed to create metric entity %s: %v", metricID, err)
			return
		}
		logger.Debug("Successfully created metric entity %s", metricID)
	}
	
	// Build value tag with labels embedded
	valueTag := fmt.Sprintf("value:%.2f", value)
	
	// Add sorted labels to value tag for dimensional data
	if len(labels) > 0 {
		var sortedKeys []string
		for k := range labels {
			sortedKeys = append(sortedKeys, k)
		}
		sort.Strings(sortedKeys)
		
		var labelParts []string
		for _, k := range sortedKeys {
			labelParts = append(labelParts, fmt.Sprintf("%s=%s", k, labels[k]))
		}
		valueTag += ":" + strings.Join(labelParts, ":")
	}
	
	// For counters, we need special handling
	if unit == "count" {
		// For counters with labels, we need to track the current value for this label combination
		currentValue := 0.0
		targetLabelString := ""
		var labelParts []string
		
		if len(labels) > 0 {
			var sortedKeys []string
			for k := range labels {
				sortedKeys = append(sortedKeys, k)
			}
			sort.Strings(sortedKeys)
			
			for _, k := range sortedKeys {
				labelParts = append(labelParts, fmt.Sprintf("%s=%s", k, labels[k]))
			}
			targetLabelString = ":" + strings.Join(labelParts, ":")
		}
		
		// Look for existing value with same labels
		if entity != nil {
			for _, tag := range entity.GetTagsWithoutTimestamp() {
				if strings.HasPrefix(tag, "value:") && strings.Contains(tag, targetLabelString) {
					valueStr := strings.TrimPrefix(tag, "value:")
					if colonIdx := strings.Index(valueStr, ":"); colonIdx > 0 {
						valueStr = valueStr[:colonIdx]
					}
					if val, err := strconv.ParseFloat(valueStr, 64); err == nil {
						currentValue = val
						break
					}
				}
			}
		}
		
		// Update value tag with incremented value
		valueTag = fmt.Sprintf("value:%.2f", currentValue + value)
		if len(labels) > 0 {
			valueTag += ":" + strings.Join(labelParts, ":")
		}
	}
	
	// Add temporal value tag to single metric entity
	if err := m.repo.AddTag(metricID, valueTag); err != nil {
		// Don't log error to avoid recursion
	}
}

// Global instance
var storageMetrics *StorageMetrics

// InitStorageMetrics initializes the global storage metrics
func InitStorageMetrics(repo models.EntityRepository) {
	storageMetrics = NewStorageMetrics(repo)
}

// GetStorageMetrics returns the global storage metrics instance
func GetStorageMetrics() *StorageMetrics {
	return storageMetrics
}

// SetRepository updates the repository for the global storage metrics
// This is used when metrics need to be initialized before the repository is available
func SetStorageMetricsRepository(repo models.EntityRepository) {
	if storageMetrics != nil {
		storageMetrics.repo = repo
		logger.Info("Updated storage metrics repository")
	}
}