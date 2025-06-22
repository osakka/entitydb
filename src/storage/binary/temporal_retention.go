package binary

import (
	"entitydb/logger"
	"entitydb/models"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// TemporalRetentionManager provides automatic, efficient retention for temporal data
// without requiring separate retention entities or complex aggregation
type TemporalRetentionManager struct {
	repo              models.EntityRepository
	retentionPolicies map[string]RetentionPolicy
	mu                sync.RWMutex
}

// RetentionPolicy defines how long to keep temporal data
type RetentionPolicy struct {
	MaxAge       time.Duration
	MaxTags      int
	CleanupBatch int
}

// NewTemporalRetentionManager creates a self-cleaning retention system
func NewTemporalRetentionManager(repo models.EntityRepository) *TemporalRetentionManager {
	return &TemporalRetentionManager{
		repo: repo,
		retentionPolicies: map[string]RetentionPolicy{
			"type:metric": {
				MaxAge:       24 * time.Hour,        // Keep 24 hours of raw metrics
				MaxTags:      1000,                  // Max 1000 temporal tags per metric
				CleanupBatch: 100,                   // Clean 100 old tags at a time
			},
			"type:session": {
				MaxAge:       7 * 24 * time.Hour,    // Keep sessions for 7 days
				MaxTags:      50,                    // Max 50 temporal tags per session
				CleanupBatch: 10,                    // Clean 10 old tags at a time
			},
			"default": {
				MaxAge:       30 * 24 * time.Hour,   // Keep 30 days by default
				MaxTags:      500,                   // Max 500 temporal tags
				CleanupBatch: 50,                    // Clean 50 old tags at a time
			},
		},
	}
}

// ApplyRetention applies retention policies during normal operations
// This is called efficiently during entity updates, not as a separate process
func (trm *TemporalRetentionManager) ApplyRetention(entity *models.Entity) error {
	if entity == nil || len(entity.Tags) == 0 {
		return nil
	}
	
	// Mark as metrics operation to prevent recursion
	SetMetricsOperation(true)
	defer SetMetricsOperation(false)
	
	// Get retention policy for this entity type
	policy := trm.getRetentionPolicy(entity)
	
	// Extract temporal tags (those with timestamps)
	temporalTags := extractTemporalTags(entity.Tags)
	
	// If we're under limits, no cleanup needed
	if len(temporalTags) <= policy.MaxTags {
		return nil
	}
	
	// Sort by timestamp (oldest first)
	sortedTags := sortTemporalTagsByAge(temporalTags)
	
	// Calculate how many to remove
	toRemove := len(temporalTags) - policy.MaxTags + policy.CleanupBatch
	if toRemove > len(sortedTags) {
		toRemove = len(sortedTags)
	}
	
	// Remove oldest tags efficiently
	tagsToRemove := sortedTags[:toRemove]
	newTags := make([]string, 0, len(entity.Tags))
	
	removeMap := make(map[string]bool)
	for _, tag := range tagsToRemove {
		removeMap[tag] = true
	}
	
	// Filter out old tags
	for _, tag := range entity.Tags {
		if !removeMap[tag] {
			newTags = append(newTags, tag)
		}
	}
	
	// Update entity with cleaned tags
	entity.Tags = newTags
	
	logger.Debug("Applied temporal retention to %s: removed %d old tags", entity.ID, toRemove)
	return nil
}

// getRetentionPolicy returns the appropriate retention policy for an entity
func (trm *TemporalRetentionManager) getRetentionPolicy(entity *models.Entity) RetentionPolicy {
	trm.mu.RLock()
	defer trm.mu.RUnlock()
	
	// Check for specific type policies
	for _, tag := range entity.Tags {
		cleanTag := strings.TrimSpace(tag)
		if strings.Contains(cleanTag, "|") {
			parts := strings.SplitN(cleanTag, "|", 2)
			if len(parts) == 2 {
				cleanTag = parts[1]
			}
		}
		
		if policy, exists := trm.retentionPolicies[cleanTag]; exists {
			return policy
		}
	}
	
	// Return default policy
	return trm.retentionPolicies["default"]
}

// extractTemporalTags extracts tags that have timestamps
func extractTemporalTags(tags []string) []string {
	var temporal []string
	for _, tag := range tags {
		if strings.Contains(tag, "|") && isTemporalTag(tag) {
			temporal = append(temporal, tag)
		}
	}
	return temporal
}

// isTemporalTag checks if a tag is temporal (has valid timestamp)
func isTemporalTag(tag string) bool {
	if !strings.Contains(tag, "|") {
		return false
	}
	
	parts := strings.SplitN(tag, "|", 2)
	if len(parts) != 2 {
		return false
	}
	
	// Try to parse as nanosecond timestamp
	if _, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
		return true
	}
	
	return false
}

// sortTemporalTagsByAge sorts temporal tags by timestamp (oldest first)
func sortTemporalTagsByAge(tags []string) []string {
	type tagWithTime struct {
		tag       string
		timestamp int64
	}
	
	var tagTimes []tagWithTime
	for _, tag := range tags {
		if !strings.Contains(tag, "|") {
			continue
		}
		
		parts := strings.SplitN(tag, "|", 2)
		if len(parts) != 2 {
			continue
		}
		
		if timestamp, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
			tagTimes = append(tagTimes, tagWithTime{
				tag:       tag,
				timestamp: timestamp,
			})
		}
	}
	
	// Sort by timestamp (oldest first)
	for i := 0; i < len(tagTimes)-1; i++ {
		for j := i + 1; j < len(tagTimes); j++ {
			if tagTimes[i].timestamp > tagTimes[j].timestamp {
				tagTimes[i], tagTimes[j] = tagTimes[j], tagTimes[i]
			}
		}
	}
	
	// Extract sorted tags
	sorted := make([]string, len(tagTimes))
	for i, tt := range tagTimes {
		sorted[i] = tt.tag
	}
	
	return sorted
}

// UpdateRetentionPolicy allows dynamic policy updates
func (trm *TemporalRetentionManager) UpdateRetentionPolicy(entityType string, policy RetentionPolicy) {
	trm.mu.Lock()
	defer trm.mu.Unlock()
	trm.retentionPolicies[entityType] = policy
	logger.Info("Updated retention policy for %s: MaxAge=%v, MaxTags=%d", 
		entityType, policy.MaxAge, policy.MaxTags)
}

// CleanupByAge removes temporal tags older than the policy MaxAge
// This is efficient and runs during normal operations
func (trm *TemporalRetentionManager) CleanupByAge(entity *models.Entity) error {
	if entity == nil || len(entity.Tags) == 0 {
		return nil
	}
	
	// Mark as metrics operation to prevent recursion
	SetMetricsOperation(true)
	defer SetMetricsOperation(false)
	
	// Get retention policy for this entity type
	policy := trm.getRetentionPolicy(entity)
	
	// Calculate cutoff time - be more aggressive under memory pressure
	memPressure := trm.getMemoryPressure()
	maxAge := policy.MaxAge
	
	// Under memory pressure, reduce retention time
	if memPressure > 0.8 {
		maxAge = time.Duration(float64(maxAge) * 0.5) // Keep half the normal time
		logger.Debug("High memory pressure (%.1f%%), reducing retention from %v to %v", 
			memPressure*100, policy.MaxAge, maxAge)
	} else if memPressure > 0.6 {
		maxAge = time.Duration(float64(maxAge) * 0.75) // Keep 75% of normal time
	}
	
	cutoffTime := time.Now().Add(-maxAge)
	cutoffNanos := cutoffTime.UnixNano()
	
	// Filter out old temporal tags
	newTags := make([]string, 0, len(entity.Tags))
	removedCount := 0
	
	for _, tag := range entity.Tags {
		keepTag := true
		
		// Check if it's a temporal tag with timestamp
		if strings.Contains(tag, "|") && isTemporalTag(tag) {
			parts := strings.SplitN(tag, "|", 2)
			if len(parts) == 2 {
				if timestamp, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
					if timestamp < cutoffNanos {
						keepTag = false
						removedCount++
					}
				}
			}
		}
		
		if keepTag {
			newTags = append(newTags, tag)
		}
	}
	
	// Update entity if we removed any tags
	if removedCount > 0 {
		entity.Tags = newTags
		logger.Debug("Temporal retention removed %d old tags from entity %s (policy: %v)", 
			removedCount, entity.ID, policy.MaxAge)
	}
	
	return nil
}

// ShouldApplyRetention determines if retention should be applied to this entity
func (trm *TemporalRetentionManager) ShouldApplyRetention(entity *models.Entity) bool {
	if entity == nil || len(entity.Tags) == 0 {
		return false
	}
	
	// Don't apply retention during metrics operations to prevent recursion
	if isMetricsOperation() {
		return false
	}
	
	// Check if entity has temporal tags
	temporalTags := extractTemporalTags(entity.Tags)
	if len(temporalTags) == 0 {
		return false
	}
	
	// Under memory pressure, be more aggressive about retention
	memPressure := trm.getMemoryPressure()
	policy := trm.getRetentionPolicy(entity)
	
	// Apply retention if:
	// 1. Too many temporal tags (always)
	// 2. Under memory pressure (more frequent cleanup)
	// 3. Periodic cleanup based on memory pressure
	return len(temporalTags) > policy.MaxTags ||
		   memPressure > 0.7 || // High memory pressure
		   (memPressure > 0.5 && len(temporalTags) > policy.MaxTags/2) // Medium pressure
}

// getMemoryPressure returns a value between 0.0 and 1.0 indicating memory pressure
func (trm *TemporalRetentionManager) getMemoryPressure() float64 {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	
	// Calculate pressure based on heap usage vs available memory
	// This is a heuristic - in production you might want more sophisticated metrics
	heapMB := float64(mem.HeapInuse) / (1024 * 1024)
	sysMB := float64(mem.Sys) / (1024 * 1024)
	
	// Pressure increases as heap approaches system memory
	if sysMB == 0 {
		return 0.0
	}
	
	pressure := heapMB / sysMB
	if pressure > 1.0 {
		pressure = 1.0
	}
	
	return pressure
}