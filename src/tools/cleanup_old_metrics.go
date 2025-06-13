//go:build tool
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strings"

	"entitydb/config"
	"entitydb/models"
	"entitydb/storage/binary"
)

// MetricAnalysis holds the results of our analysis
type MetricAnalysis struct {
	OldMetrics      []*models.Entity
	NewMetrics      []*models.Entity
	UnknownMetrics  []*models.Entity
	TotalMetrics    int
}

func main() {
	// Load configuration for default database path
	cfg := config.Load()
	
	var (
		dbPath    = flag.String("db", cfg.DatabasePath(), "Path to entity database")
		dryRun    = flag.Bool("dry-run", true, "Run in dry-run mode (don't delete)")
		verbose   = flag.Bool("v", false, "Verbose output")
		force     = flag.Bool("force", false, "Force deletion without confirmation")
	)
	flag.Parse()

	// Open repository
	repo, err := binary.NewEntityRepository(*dbPath)
	if err != nil {
		log.Fatalf("Failed to open repository: %v", err)
	}
	defer repo.Close()

	fmt.Println("=== EntityDB Metric Cleanup Tool ===")
	fmt.Printf("Database: %s\n", *dbPath)
	fmt.Printf("Mode: %s\n", getDryRunMode(*dryRun))
	fmt.Println()

	// Analyze metrics
	analysis, err := analyzeMetrics(repo, *verbose)
	if err != nil {
		log.Fatalf("Failed to analyze metrics: %v", err)
	}

	// Display results
	displayAnalysis(analysis, *verbose)

	// If not dry-run, ask for confirmation and delete
	if !*dryRun && len(analysis.OldMetrics) > 0 {
		if !*force && !confirmDeletion(len(analysis.OldMetrics)) {
			fmt.Println("Cleanup cancelled.")
			return
		}

		deletedCount, err := deleteOldMetrics(repo, analysis.OldMetrics, *verbose)
		if err != nil {
			log.Fatalf("Failed to delete metrics: %v", err)
		}

		fmt.Printf("\n✅ Successfully deleted %d old metric entities.\n", deletedCount)
		
		// Show final counts
		fmt.Println("\n=== Final State ===")
		finalAnalysis, _ := analyzeMetrics(repo, false)
		fmt.Printf("Total metrics remaining: %d\n", finalAnalysis.TotalMetrics)
		fmt.Printf("New temporal metrics: %d\n", len(finalAnalysis.NewMetrics))
	}
}

func analyzeMetrics(repo *binary.EntityRepository, verbose bool) (*MetricAnalysis, error) {
	analysis := &MetricAnalysis{
		OldMetrics:     []*models.Entity{},
		NewMetrics:     []*models.Entity{},
		UnknownMetrics: []*models.Entity{},
	}

	// Get all entities
	entities, err := repo.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list entities: %w", err)
	}

	// Analyze each entity
	for _, entity := range entities {
		if isMetricEntity(entity) {
			analysis.TotalMetrics++
			
			if isOldMetric(entity) {
				analysis.OldMetrics = append(analysis.OldMetrics, entity)
			} else if isNewTemporalMetric(entity) {
				analysis.NewMetrics = append(analysis.NewMetrics, entity)
			} else {
				analysis.UnknownMetrics = append(analysis.UnknownMetrics, entity)
			}
		}
	}

	return analysis, nil
}

func isMetricEntity(entity *models.Entity) bool {
	for _, tag := range entity.Tags {
		// Strip timestamp if present
		cleanTag := stripTimestamp(tag)
		
		// Old style: hub:metrics
		if strings.HasPrefix(cleanTag, "hub:metrics") {
			return true
		}
		
		// New style: type:metric
		if cleanTag == "type:metric" {
			return true
		}
	}
	return false
}

func isOldMetric(entity *models.Entity) bool {
	hasHubMetrics := false
	hasTemporalValue := false
	
	for _, tag := range entity.Tags {
		cleanTag := stripTimestamp(tag)
		
		if strings.HasPrefix(cleanTag, "hub:metrics") {
			hasHubMetrics = true
		}
		
		// Check for temporal value tags (metric:value:xxx)
		if strings.HasPrefix(cleanTag, "metric:value:") && strings.Contains(tag, "|") {
			hasTemporalValue = true
		}
	}
	
	// Old metrics have hub:metrics tag but no temporal values
	return hasHubMetrics && !hasTemporalValue
}

func isNewTemporalMetric(entity *models.Entity) bool {
	hasTypeMetric := false
	hasTemporalValue := false
	
	for _, tag := range entity.Tags {
		cleanTag := stripTimestamp(tag)
		
		if cleanTag == "type:metric" {
			hasTypeMetric = true
		}
		
		// Check for temporal value tags
		if strings.HasPrefix(cleanTag, "metric:value:") && strings.Contains(tag, "|") {
			hasTemporalValue = true
		}
	}
	
	return hasTypeMetric && hasTemporalValue
}

func stripTimestamp(tag string) string {
	parts := strings.SplitN(tag, "|", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return tag
}

func displayAnalysis(analysis *MetricAnalysis, verbose bool) {
	fmt.Println("=== Metric Analysis ===")
	fmt.Printf("Total metric entities: %d\n", analysis.TotalMetrics)
	fmt.Printf("Old metrics (to be deleted): %d\n", len(analysis.OldMetrics))
	fmt.Printf("New temporal metrics (preserved): %d\n", len(analysis.NewMetrics))
	fmt.Printf("Unknown metrics: %d\n", len(analysis.UnknownMetrics))
	
	if verbose && len(analysis.OldMetrics) > 0 {
		fmt.Println("\n=== Old Metrics (One entity per data point) ===")
		for i, entity := range analysis.OldMetrics {
			if i < 10 { // Show first 10
				displayEntity(entity)
			}
		}
		if len(analysis.OldMetrics) > 10 {
			fmt.Printf("... and %d more\n", len(analysis.OldMetrics)-10)
		}
	}
	
	if verbose && len(analysis.NewMetrics) > 0 {
		fmt.Println("\n=== New Temporal Metrics (One entity per metric type) ===")
		for _, entity := range analysis.NewMetrics {
			displayEntity(entity)
		}
	}
}

func displayEntity(entity *models.Entity) {
	fmt.Printf("\nID: %s\n", entity.ID)
	fmt.Println("Tags:")
	for _, tag := range entity.Tags {
		cleanTag := stripTimestamp(tag)
		if strings.Contains(tag, "|") {
			fmt.Printf("  - %s (temporal)\n", cleanTag)
		} else {
			fmt.Printf("  - %s\n", cleanTag)
		}
	}
	
	// Try to parse content as JSON
	if len(entity.Content) > 0 {
		var content map[string]interface{}
		if err := json.Unmarshal(entity.Content, &content); err == nil {
			fmt.Printf("Content: %v\n", content)
		} else {
			fmt.Printf("Content: %d bytes\n", len(entity.Content))
		}
	}
}

func getDryRunMode(dryRun bool) string {
	if dryRun {
		return "DRY-RUN (no changes will be made)"
	}
	return "LIVE (will delete entities)"
}

func confirmDeletion(count int) bool {
	fmt.Printf("\n⚠️  WARNING: About to delete %d old metric entities!\n", count)
	fmt.Print("Are you sure you want to continue? (yes/no): ")
	
	var response string
	fmt.Scanln(&response)
	
	return strings.ToLower(response) == "yes"
}

func deleteOldMetrics(repo *binary.EntityRepository, oldMetrics []*models.Entity, verbose bool) (int, error) {
	deletedCount := 0
	
	for i, entity := range oldMetrics {
		if verbose {
			fmt.Printf("Deleting %d/%d: %s\n", i+1, len(oldMetrics), entity.ID)
		}
		
		if err := repo.Delete(entity.ID); err != nil {
			return deletedCount, fmt.Errorf("failed to delete entity %s: %w", entity.ID, err)
		}
		
		deletedCount++
		
		// Show progress every 100 entities
		if !verbose && deletedCount%100 == 0 {
			fmt.Printf("Deleted %d/%d entities...\n", deletedCount, len(oldMetrics))
		}
	}
	
	return deletedCount, nil
}