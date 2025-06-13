//go:build tool
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"
	
	"entitydb/logger"
	"entitydb/models"
	"entitydb/storage/binary"
)

type IntegrityReport struct {
	StartTime        time.Time                `json:"start_time"`
	EndTime          time.Time                `json:"end_time"`
	TotalEntities    int                      `json:"total_entities"`
	ValidEntities    int                      `json:"valid_entities"`
	CorruptedEntities int                     `json:"corrupted_entities"`
	IndexMismatches  int                      `json:"index_mismatches"`
	OrphanedEntries  int                      `json:"orphaned_entries"`
	ChecksumFailures int                      `json:"checksum_failures"`
	Errors           []IntegrityError         `json:"errors"`
	Summary          string                   `json:"summary"`
	HealthScore      float64                  `json:"health_score"`
}

type IntegrityError struct {
	EntityID    string `json:"entity_id"`
	ErrorType   string `json:"error_type"`
	Description string `json:"description"`
	Offset      int64  `json:"offset,omitempty"`
	Expected    string `json:"expected,omitempty"`
	Actual      string `json:"actual,omitempty"`
}

func main() {
	var (
		dataPath    = flag.String("data", "var", "Path to EntityDB data directory")
		outputPath  = flag.String("output", "", "Path to save JSON report (stdout if empty)")
		fix         = flag.Bool("fix", false, "Attempt to fix issues found")
		verbose     = flag.Bool("verbose", false, "Enable verbose logging")
	)
	flag.Parse()
	
	// Verbose mode is always on for this tool
	
	logger.Info("EntityDB Integrity Check Tool v1.0")
	logger.Info("Data path: %s", *dataPath)
	
	report := &IntegrityReport{
		StartTime: time.Now(),
		Errors:    make([]IntegrityError, 0),
	}
	
	// Create repository factory
	factory := &binary.RepositoryFactory{}
	repo, err := factory.CreateRepository(*dataPath)
	if err != nil {
		logger.Error("Failed to create repository: %v", err)
		os.Exit(1)
	}
	
	// Get underlying binary repository for direct access
	var binaryRepo *binary.EntityRepository
	switch r := repo.(type) {
	case *binary.CachedRepository:
		if hr, ok := r.GetUnderlying().(*binary.HighPerformanceRepository); ok {
			binaryRepo = hr.GetBaseRepository()
		} else if br, ok := r.GetUnderlying().(*binary.EntityRepository); ok {
			binaryRepo = br
		}
	case *binary.HighPerformanceRepository:
		binaryRepo = r.GetBaseRepository()
	case *binary.EntityRepository:
		binaryRepo = r
	default:
		logger.Error("Cannot access binary repository from type %T", repo)
		os.Exit(1)
	}
	
	// Phase 1: Check all entities in repository
	logger.Info("Phase 1: Scanning all entities...")
	allEntities, err := repo.ListByTag("")
	if err != nil {
		logger.Error("Failed to list entities: %v", err)
		report.Errors = append(report.Errors, IntegrityError{
			ErrorType:   "REPOSITORY_ERROR",
			Description: fmt.Sprintf("Failed to list entities: %v", err),
		})
	} else {
		report.TotalEntities = len(allEntities)
		logger.Info("Found %d entities in repository", report.TotalEntities)
	}
	
	// Phase 2: Verify each entity
	logger.Info("Phase 2: Verifying entity integrity...")
	for i, entity := range allEntities {
		if i%100 == 0 {
			logger.Debug("Progress: %d/%d entities checked", i, len(allEntities))
		}
		
		// Check 1: Entity can be read
		readEntity, err := repo.GetByID(entity.ID)
		if err != nil {
			report.CorruptedEntities++
			report.Errors = append(report.Errors, IntegrityError{
				EntityID:    entity.ID,
				ErrorType:   "READ_ERROR",
				Description: fmt.Sprintf("Failed to read entity: %v", err),
			})
			continue
		}
		
		// Check 2: Content checksum
		if len(readEntity.Content) > 0 {
			checksum := sha256.Sum256(readEntity.Content)
			checksumStr := hex.EncodeToString(checksum[:])
			
			// In a full implementation, we'd compare against stored checksum
			// For now, we just log it
			logger.Debug("Entity %s content checksum: %s", entity.ID, checksumStr)
		}
		
		// Check 3: Tag integrity
		tags := readEntity.GetTagsWithoutTimestamp()
		if len(tags) == 0 {
			report.Errors = append(report.Errors, IntegrityError{
				EntityID:    entity.ID,
				ErrorType:   "TAG_ERROR",
				Description: "Entity has no tags",
			})
		}
		
		// Check 4: Temporal consistency
		if readEntity.CreatedAt > readEntity.UpdatedAt {
			report.Errors = append(report.Errors, IntegrityError{
				EntityID:    entity.ID,
				ErrorType:   "TEMPORAL_ERROR",
				Description: fmt.Sprintf("CreatedAt (%d) > UpdatedAt (%d)", readEntity.CreatedAt, readEntity.UpdatedAt),
			})
		}
		
		report.ValidEntities++
	}
	
	// Phase 3: Check index integrity
	logger.Info("Phase 3: Checking index integrity...")
	indexErrors := binaryRepo.VerifyIndexIntegrity()
	for _, err := range indexErrors {
		report.IndexMismatches++
		report.Errors = append(report.Errors, IntegrityError{
			ErrorType:   "INDEX_ERROR",
			Description: err.Error(),
		})
	}
	
	// Phase 4: Check for orphaned entries
	logger.Info("Phase 4: Checking for orphaned entries...")
	orphaned := binaryRepo.FindOrphanedEntries()
	report.OrphanedEntries = len(orphaned)
	for _, id := range orphaned {
		report.Errors = append(report.Errors, IntegrityError{
			EntityID:    id,
			ErrorType:   "ORPHANED_ENTRY",
			Description: "Entry in index but not in data file",
		})
	}
	
	// Calculate health score
	report.EndTime = time.Now()
	if report.TotalEntities > 0 {
		report.HealthScore = float64(report.ValidEntities) / float64(report.TotalEntities) * 100
	}
	
	// Generate summary
	report.Summary = fmt.Sprintf(
		"Integrity check completed in %v. Health Score: %.2f%%. "+
			"Total: %d, Valid: %d, Corrupted: %d, Index Mismatches: %d, Orphaned: %d",
		report.EndTime.Sub(report.StartTime),
		report.HealthScore,
		report.TotalEntities,
		report.ValidEntities,
		report.CorruptedEntities,
		report.IndexMismatches,
		report.OrphanedEntries,
	)
	
	logger.Info(report.Summary)
	
	// Fix issues if requested
	if *fix && len(report.Errors) > 0 {
		logger.Info("Phase 5: Attempting to fix issues...")
		fixCount := 0
		
		// Fix index mismatches by rebuilding
		if report.IndexMismatches > 0 {
			logger.Info("Rebuilding index...")
			if err := binaryRepo.RebuildIndex(); err != nil {
				logger.Error("Failed to rebuild index: %v", err)
			} else {
				fixCount += report.IndexMismatches
				logger.Info("Index rebuilt successfully")
			}
		}
		
		// Fix orphaned entries by removing from index
		if report.OrphanedEntries > 0 {
			logger.Info("Removing orphaned entries...")
			for _, id := range orphaned {
				if err := binaryRepo.RemoveFromIndex(id); err != nil {
					logger.Error("Failed to remove orphaned entry %s: %v", id, err)
				} else {
					fixCount++
				}
			}
		}
		
		logger.Info("Fixed %d issues", fixCount)
	}
	
	// Output report
	reportJSON, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		logger.Error("Failed to marshal report: %v", err)
		os.Exit(1)
	}
	
	if *outputPath != "" {
		if err := os.WriteFile(*outputPath, reportJSON, 0644); err != nil {
			logger.Error("Failed to write report: %v", err)
			os.Exit(1)
		}
		logger.Info("Report saved to %s", *outputPath)
	} else {
		fmt.Println(string(reportJSON))
	}
	
	// Exit with error if health is poor
	if report.HealthScore < 90.0 {
		os.Exit(1)
	}
}