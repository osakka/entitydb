package api

import (
	"encoding/json"
	"entitydb/models"
	"entitydb/storage/binary"
	"net/http"
	"time"
)

// IntegrityMetrics represents data integrity metrics
type IntegrityMetrics struct {
	Timestamp           time.Time                    `json:"timestamp"`
	HealthScore         float64                     `json:"health_score"`
	EntityMetrics       EntityIntegrityMetrics      `json:"entity_metrics"`
	IndexMetrics        IndexIntegrityMetrics       `json:"index_metrics"`
	WALMetrics          WALIntegrityMetrics         `json:"wal_metrics"`
	ChecksumMetrics     ChecksumIntegrityMetrics    `json:"checksum_metrics"`
	OperationMetrics    OperationIntegrityMetrics   `json:"operation_metrics"`
	RecoveryMetrics     RecoveryIntegrityMetrics    `json:"recovery_metrics"`
}

// EntityIntegrityMetrics tracks entity-level integrity
type EntityIntegrityMetrics struct {
	TotalEntities       int     `json:"total_entities"`
	CorruptedEntities   int     `json:"corrupted_entities"`
	RecoveredEntities   int     `json:"recovered_entities"`
	OrphanedEntities    int     `json:"orphaned_entities"`
	SuccessRate         float64 `json:"success_rate"`
}

// IndexIntegrityMetrics tracks index integrity
type IndexIntegrityMetrics struct {
	IndexedEntities     int     `json:"indexed_entities"`
	MissingFromIndex    int     `json:"missing_from_index"`
	IndexMismatches     int     `json:"index_mismatches"`
	LastIndexSave       string  `json:"last_index_save"`
	IndexHealthy        bool    `json:"index_healthy"`
}

// WALIntegrityMetrics tracks WAL integrity
type WALIntegrityMetrics struct {
	WALEntries          int     `json:"wal_entries"`
	CorruptedEntries    int     `json:"corrupted_entries"`
	LastCheckpoint      string  `json:"last_checkpoint"`
	WALSizeMB           float64 `json:"wal_size_mb"`
}

// ChecksumIntegrityMetrics tracks checksum validation
type ChecksumIntegrityMetrics struct {
	EntitiesWithChecksum    int     `json:"entities_with_checksum"`
	ValidChecksums          int     `json:"valid_checksums"`
	InvalidChecksums        int     `json:"invalid_checksums"`
	ChecksumCoverage        float64 `json:"checksum_coverage_percent"`
}

// OperationIntegrityMetrics tracks operation success rates
type OperationIntegrityMetrics struct {
	TotalOperations     int64   `json:"total_operations"`
	SuccessfulOps       int64   `json:"successful_operations"`
	FailedOps           int64   `json:"failed_operations"`
	RecoveredOps        int64   `json:"recovered_operations"`
	SuccessRate         float64 `json:"success_rate"`
	ActiveOperations    int     `json:"active_operations"`
}

// RecoveryIntegrityMetrics tracks recovery operations
type RecoveryIntegrityMetrics struct {
	RecoveryAttempts    int     `json:"recovery_attempts"`
	SuccessfulRecoveries int    `json:"successful_recoveries"`
	FailedRecoveries    int     `json:"failed_recoveries"`
	RecoverySuccessRate float64 `json:"recovery_success_rate"`
	LastRecoveryTime    string  `json:"last_recovery_time"`
}

// IntegrityMetricsHandler handles integrity metrics requests
func IntegrityMetricsHandler(repository models.EntityRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		// Get the repository
		repo, ok := repository.(*binary.EntityRepository)
		if !ok {
			// All repository variants now merged into EntityRepository
			// Check if it's a cached repository wrapping an EntityRepository
			if cachedRepo, ok := repository.(*binary.CachedRepository); ok {
				if entityRepo, ok := cachedRepo.EntityRepository.(*binary.EntityRepository); ok {
					repo = entityRepo
				} else {
					http.Error(w, "Integrity metrics not available for this repository type", http.StatusNotImplemented)
					return
				}
			} else {
				http.Error(w, "Integrity metrics not available for this repository type", http.StatusNotImplemented)
				return
			}
		}
		
		// Collect metrics
		metrics := collectIntegrityMetrics(repo)
		
		// Return JSON response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metrics)
	}
}

func collectIntegrityMetrics(repo *binary.EntityRepository) IntegrityMetrics {
	metrics := IntegrityMetrics{
		Timestamp: time.Now(),
	}
	
	// Entity metrics
	// Use List method instead of ListAll
	allEntities, _ := repo.List()
	metrics.EntityMetrics.TotalEntities = len(allEntities)
	
	// Check for corrupted entities
	corruptedCount := 0
	recoveredCount := 0
	for _, entity := range allEntities {
		for _, tag := range entity.Tags {
			if containsTag(tag, "status:corrupted") {
				corruptedCount++
				break
			}
			if containsTag(tag, "status:recovered") {
				recoveredCount++
				break
			}
		}
	}
	metrics.EntityMetrics.CorruptedEntities = corruptedCount
	metrics.EntityMetrics.RecoveredEntities = recoveredCount
	
	// Find orphaned entries
	orphaned := repo.FindOrphanedEntries()
	metrics.EntityMetrics.OrphanedEntities = len(orphaned)
	
	// Calculate entity success rate
	if metrics.EntityMetrics.TotalEntities > 0 {
		healthyEntities := metrics.EntityMetrics.TotalEntities - 
			metrics.EntityMetrics.CorruptedEntities - 
			metrics.EntityMetrics.OrphanedEntities
		metrics.EntityMetrics.SuccessRate = float64(healthyEntities) / float64(metrics.EntityMetrics.TotalEntities) * 100
	}
	
	// Index metrics
	if err := repo.VerifyIndexHealth(); err != nil {
		metrics.IndexMetrics.IndexHealthy = false
		// Parse error to extract mismatch counts
		// This is simplified - in production you'd have better error parsing
		metrics.IndexMetrics.IndexMismatches = 1
	} else {
		metrics.IndexMetrics.IndexHealthy = true
	}
	metrics.IndexMetrics.IndexedEntities = metrics.EntityMetrics.TotalEntities
	
	// Checksum metrics
	entitiesWithChecksum := 0
	validChecksums := 0
	invalidChecksums := 0
	
	for _, entity := range allEntities {
		isValid, checksum := repo.ValidateEntityChecksum(entity)
		if checksum != "" {
			entitiesWithChecksum++
			if isValid {
				validChecksums++
			} else {
				invalidChecksums++
			}
		}
	}
	
	metrics.ChecksumMetrics.EntitiesWithChecksum = entitiesWithChecksum
	metrics.ChecksumMetrics.ValidChecksums = validChecksums
	metrics.ChecksumMetrics.InvalidChecksums = invalidChecksums
	if metrics.EntityMetrics.TotalEntities > 0 {
		metrics.ChecksumMetrics.ChecksumCoverage = float64(entitiesWithChecksum) / float64(metrics.EntityMetrics.TotalEntities) * 100
	}
	
	// Operation metrics from global tracker
	opStats := models.GetOperationStats()
	metrics.OperationMetrics.TotalOperations = opStats.TotalOperations
	metrics.OperationMetrics.SuccessfulOps = opStats.SuccessfulOperations
	metrics.OperationMetrics.FailedOps = opStats.FailedOperations
	metrics.OperationMetrics.ActiveOperations = opStats.ActiveOperations
	if metrics.OperationMetrics.TotalOperations > 0 {
		metrics.OperationMetrics.SuccessRate = float64(metrics.OperationMetrics.SuccessfulOps) / 
			float64(metrics.OperationMetrics.TotalOperations) * 100
	}
	
	// Recovery metrics
	recoveryStats := models.GetRecoveryStats()
	metrics.RecoveryMetrics.RecoveryAttempts = recoveryStats.TotalAttempts
	metrics.RecoveryMetrics.SuccessfulRecoveries = recoveryStats.Successful
	metrics.RecoveryMetrics.FailedRecoveries = recoveryStats.Failed
	if metrics.RecoveryMetrics.RecoveryAttempts > 0 {
		metrics.RecoveryMetrics.RecoverySuccessRate = float64(metrics.RecoveryMetrics.SuccessfulRecoveries) / 
			float64(metrics.RecoveryMetrics.RecoveryAttempts) * 100
	}
	if !recoveryStats.LastRecoveryTime.IsZero() {
		metrics.RecoveryMetrics.LastRecoveryTime = recoveryStats.LastRecoveryTime.Format(time.RFC3339)
	}
	
	// Calculate overall health score
	metrics.HealthScore = calculateHealthScore(metrics)
	
	return metrics
}

func calculateHealthScore(metrics IntegrityMetrics) float64 {
	score := 100.0
	
	// Deduct points for issues
	if metrics.EntityMetrics.CorruptedEntities > 0 {
		score -= float64(metrics.EntityMetrics.CorruptedEntities) / float64(metrics.EntityMetrics.TotalEntities) * 20
	}
	
	if metrics.EntityMetrics.OrphanedEntities > 0 {
		score -= float64(metrics.EntityMetrics.OrphanedEntities) / float64(metrics.EntityMetrics.TotalEntities) * 10
	}
	
	if !metrics.IndexMetrics.IndexHealthy {
		score -= 15
	}
	
	if metrics.ChecksumMetrics.InvalidChecksums > 0 {
		score -= float64(metrics.ChecksumMetrics.InvalidChecksums) / float64(metrics.ChecksumMetrics.EntitiesWithChecksum) * 25
	}
	
	if metrics.ChecksumMetrics.ChecksumCoverage < 50 {
		score -= (50 - metrics.ChecksumMetrics.ChecksumCoverage) * 0.3
	}
	
	if metrics.OperationMetrics.SuccessRate < 95 {
		score -= (95 - metrics.OperationMetrics.SuccessRate) * 0.5
	}
	
	// Ensure score doesn't go below 0
	if score < 0 {
		score = 0
	}
	
	return score
}

func containsTag(tag, search string) bool {
	// Handle temporal tags
	parts := splitTemporalTag(tag)
	if len(parts) == 2 {
		return parts[1] == search
	}
	return tag == search
}

func splitTemporalTag(tag string) []string {
	// Split by | for temporal tags
	if idx := findFirstPipe(tag); idx != -1 {
		return []string{tag[:idx], tag[idx+1:]}
	}
	return []string{tag}
}

func findFirstPipe(s string) int {
	for i, c := range s {
		if c == '|' {
			return i
		}
	}
	return -1
}