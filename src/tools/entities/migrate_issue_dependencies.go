package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Tool configuration
var (
	dbPath      string
	dryRun      bool
	batchSize   int
	logFile     string
	force       bool
	interactive bool
)

// EntityRelationship represents a relationship between entities in the migration
type EntityRelationship struct {
	SourceID         string    `json:"source_id"`
	RelationshipType string    `json:"relationship_type"`
	TargetID         string    `json:"target_id"`
	CreatedAt        time.Time `json:"created_at"`
	CreatedBy        string    `json:"created_by,omitempty"`
	Metadata         string    `json:"metadata,omitempty"`
}

// IssueDependency represents a dependency between issues in the migration
type IssueDependency struct {
	ID             string    `json:"id"`
	IssueID        string    `json:"issue_id"`
	DependsOnID    string    `json:"depends_on_id"`
	DependencyType string    `json:"dependency_type"`
	Description    string    `json:"description,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	CreatedBy      string    `json:"created_by,omitempty"`
}

// MigrationStats tracks statistics about the migration
type MigrationStats struct {
	TotalDependencies      int
	ProcessedDependencies  int
	SuccessfulMigrations   int
	FailedMigrations       int
	DuplicateMigrations    int
	MissingEntities        int
	StartTime              time.Time
	EndTime                time.Time
	FailedDependencyIDs    []string
	MissingEntitySourceIDs []string
	MissingEntityTargetIDs []string
}

// MigrationLogger handles logging for the migration tool
type MigrationLogger struct {
	Logger *log.Logger
	File   *os.File
}

func init() {
	// Set up command-line flags
	flag.StringVar(&dbPath, "db", "/opt/entitydb/var/db/entitydb.db", "Path to the SQLite database")
	flag.BoolVar(&dryRun, "dry-run", false, "Perform a dry run without making changes")
	flag.IntVar(&batchSize, "batch-size", 100, "Number of dependencies to process in each batch")
	flag.StringVar(&logFile, "log", "/opt/entitydb/var/log/migration.log", "Path to the log file")
	flag.BoolVar(&force, "force", false, "Force migration even if already migrated")
	flag.BoolVar(&interactive, "interactive", true, "Run in interactive mode with prompts")
}

// newMigrationLogger creates a new logger that writes to both file and stdout
func newMigrationLogger(logFilePath string) (*MigrationLogger, error) {
	// Ensure the directory exists
	logDir := filepath.Dir(logFilePath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open or create log file
	f, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Create a multi-writer logger
	multiWriter := log.New(f, "", log.LstdFlags)
	return &MigrationLogger{Logger: multiWriter, File: f}, nil
}

// logf logs a formatted message to both file and stdout
func (m *MigrationLogger) logf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	m.Logger.Println(msg)
	fmt.Println(msg)
}

// close closes the log file
func (m *MigrationLogger) close() {
	if m.File != nil {
		m.File.Close()
	}
}

// validateDb checks if the database exists and has the required tables
func validateDb(db *sql.DB) error {
	// Check if issue_dependencies table exists
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='issue_dependencies'").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check for issue_dependencies table: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("issue_dependencies table not found in database")
	}

	// Check if entity_relationships table exists
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='entity_relationships'").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check for entity_relationships table: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("entity_relationships table not found in database - run migration first")
	}

	return nil
}

// getMigrationStats gets statistics about dependencies to migrate
func getMigrationStats(db *sql.DB) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM issue_dependencies`
	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get dependency count: %w", err)
	}
	return count, nil
}

// migrateIssueDependencies migrates all issue dependencies to entity relationships
func migrateIssueDependencies(db *sql.DB, logger *MigrationLogger, dryRun bool, batchSize int, force bool) (*MigrationStats, error) {
	stats := &MigrationStats{
		StartTime:              time.Now(),
		FailedDependencyIDs:    make([]string, 0),
		MissingEntitySourceIDs: make([]string, 0),
		MissingEntityTargetIDs: make([]string, 0),
	}

	// Get the total count
	var totalCount int
	err := db.QueryRow("SELECT COUNT(*) FROM issue_dependencies").Scan(&totalCount)
	if err != nil {
		return stats, fmt.Errorf("failed to get total dependencies count: %w", err)
	}
	stats.TotalDependencies = totalCount

	logger.logf("Starting migration of %d issue dependencies to entity relationships", totalCount)
	if dryRun {
		logger.logf("DRY RUN: No changes will be made to the database")
	}

	// Prepare the query for getting dependencies in batches
	query := `
		SELECT 
			id, issue_id, depends_on_id, dependency_type, 
			created_at, created_by
		FROM issue_dependencies
		ORDER BY created_at
		LIMIT ? OFFSET ?
	`

	// Process in batches
	offset := 0
	for {
		// Query dependencies in this batch
		rows, err := db.Query(query, batchSize, offset)
		if err != nil {
			return stats, fmt.Errorf("failed to query dependencies: %w", err)
		}

		// Process each dependency
		batchCount := 0
		for rows.Next() {
			var dependency IssueDependency
			var createdAtStr string
			var createdBy sql.NullString
			var dependencyType sql.NullString

			// Scan the dependency
			err := rows.Scan(
				&dependency.ID,
				&dependency.IssueID,
				&dependency.DependsOnID,
				&dependencyType,
				&createdAtStr,
				&createdBy,
			)
			if err != nil {
				logger.logf("Error scanning dependency: %v", err)
				stats.FailedMigrations++
				continue
			}

			// Parse the created_at timestamp
			dependency.CreatedAt, err = time.Parse("2006-01-02 15:04:05", createdAtStr)
			if err != nil {
				dependency.CreatedAt = time.Now() // Use current time if parsing fails
			}

			// Set nullable fields
			if dependencyType.Valid {
				dependency.DependencyType = dependencyType.String
			} else {
				dependency.DependencyType = "blocker" // Default
			}
			if createdBy.Valid {
				dependency.CreatedBy = createdBy.String
			} else {
				dependency.CreatedBy = "system" // Default
			}

			// Create the entity relationship
			relationship := EntityRelationship{
				SourceID:         dependency.IssueID,
				RelationshipType: "depends_on",
				TargetID:         dependency.DependsOnID,
				CreatedAt:        dependency.CreatedAt,
				CreatedBy:        dependency.CreatedBy,
			}

			// Create metadata
			metadata := map[string]interface{}{
				"dependency_type":  dependency.DependencyType,
				"migration_source": "dependency_migration_tool",
				"original_id":      dependency.ID,
			}
			metadataJSON, err := json.Marshal(metadata)
			if err != nil {
				logger.logf("Error marshaling metadata for dependency %s: %v", dependency.ID, err)
				stats.FailedMigrations++
				stats.FailedDependencyIDs = append(stats.FailedDependencyIDs, dependency.ID)
				continue
			}
			relationship.Metadata = string(metadataJSON)

			// Check if the entity exists for source
			var sourceExists int
			err = db.QueryRow("SELECT COUNT(*) FROM entities WHERE id = ?", relationship.SourceID).Scan(&sourceExists)
			if err != nil {
				logger.logf("Error checking source entity %s: %v", relationship.SourceID, err)
			} else if sourceExists == 0 {
				logger.logf("Warning: Source entity %s does not exist", relationship.SourceID)
				stats.MissingEntitySourceIDs = append(stats.MissingEntitySourceIDs, relationship.SourceID)
			}

			// Check if the entity exists for target
			var targetExists int
			err = db.QueryRow("SELECT COUNT(*) FROM entities WHERE id = ?", relationship.TargetID).Scan(&targetExists)
			if err != nil {
				logger.logf("Error checking target entity %s: %v", relationship.TargetID, err)
			} else if targetExists == 0 {
				logger.logf("Warning: Target entity %s does not exist", relationship.TargetID)
				stats.MissingEntityTargetIDs = append(stats.MissingEntityTargetIDs, relationship.TargetID)
			}

			// Check if the relationship already exists
			var existingCount int
			err = db.QueryRow(
				"SELECT COUNT(*) FROM entity_relationships WHERE source_id = ? AND relationship_type = ? AND target_id = ?",
				relationship.SourceID, relationship.RelationshipType, relationship.TargetID,
			).Scan(&existingCount)
			if err != nil {
				logger.logf("Error checking existing relationship: %v", err)
				stats.FailedMigrations++
				stats.FailedDependencyIDs = append(stats.FailedDependencyIDs, dependency.ID)
				continue
			}

			if existingCount > 0 && !force {
				logger.logf("Skipping dependency %s: relationship already exists", dependency.ID)
				stats.DuplicateMigrations++
				continue
			}

			// In dry run mode, just report what would be done
			if dryRun {
				logger.logf("Would create relationship: %s -[%s]-> %s (from dependency %s)",
					relationship.SourceID, relationship.RelationshipType, relationship.TargetID, dependency.ID)
				stats.SuccessfulMigrations++
			} else {
				// Create the relationship in the database
				var insertQuery string
				var insertArgs []interface{}

				if existingCount > 0 && force {
					// Update existing relationship
					insertQuery = `
						UPDATE entity_relationships 
						SET created_at = ?, created_by = ?, metadata = ?
						WHERE source_id = ? AND relationship_type = ? AND target_id = ?
					`
					insertArgs = []interface{}{
						relationship.CreatedAt.Format(time.RFC3339), relationship.CreatedBy, relationship.Metadata,
						relationship.SourceID, relationship.RelationshipType, relationship.TargetID,
					}
				} else {
					// Insert new relationship
					insertQuery = `
						INSERT INTO entity_relationships 
						(source_id, relationship_type, target_id, created_at, created_by, metadata)
						VALUES (?, ?, ?, ?, ?, ?)
					`
					insertArgs = []interface{}{
						relationship.SourceID, relationship.RelationshipType, relationship.TargetID,
						relationship.CreatedAt.Format(time.RFC3339), relationship.CreatedBy, relationship.Metadata,
					}
				}

				_, err = db.Exec(insertQuery, insertArgs...)
				if err != nil {
					logger.logf("Error creating relationship for dependency %s: %v", dependency.ID, err)
					stats.FailedMigrations++
					stats.FailedDependencyIDs = append(stats.FailedDependencyIDs, dependency.ID)
					continue
				}

				// Insert migration status
				_, err = db.Exec(
					"INSERT OR REPLACE INTO dependency_migration_status (issue_dependency_id, entity_relationship_created, migrated_at) VALUES (?, 1, ?)",
					dependency.ID, time.Now().Format(time.RFC3339),
				)
				if err != nil {
					logger.logf("Warning: Failed to update migration status for dependency %s: %v", dependency.ID, err)
				}

				logger.logf("Created relationship: %s -[%s]-> %s (from dependency %s)",
					relationship.SourceID, relationship.RelationshipType, relationship.TargetID, dependency.ID)
				stats.SuccessfulMigrations++
			}

			batchCount++
			stats.ProcessedDependencies++
		}
		rows.Close()

		// If we processed fewer than the batch size, we're done
		if batchCount < batchSize {
			break
		}

		// Update offset for next batch
		offset += batchSize
	}

	stats.EndTime = time.Now()
	return stats, nil
}

// printStats prints migration statistics
func printStats(stats *MigrationStats, logger *MigrationLogger) {
	duration := stats.EndTime.Sub(stats.StartTime)
	successRate := 0.0
	if stats.ProcessedDependencies > 0 {
		successRate = float64(stats.SuccessfulMigrations) / float64(stats.ProcessedDependencies) * 100
	}

	logger.logf("\n----- Migration Stats -----")
	logger.logf("Total dependencies:     %d", stats.TotalDependencies)
	logger.logf("Processed dependencies: %d", stats.ProcessedDependencies)
	logger.logf("Successful migrations:  %d", stats.SuccessfulMigrations)
	logger.logf("Failed migrations:      %d", stats.FailedMigrations)
	logger.logf("Duplicate migrations:   %d", stats.DuplicateMigrations)
	logger.logf("Missing source entities: %d", len(stats.MissingEntitySourceIDs))
	logger.logf("Missing target entities: %d", len(stats.MissingEntityTargetIDs))
	logger.logf("Duration:               %v", duration)
	logger.logf("Success rate:           %.2f%%", successRate)

	if len(stats.FailedDependencyIDs) > 0 {
		maxFailed := 10
		if len(stats.FailedDependencyIDs) < maxFailed {
			maxFailed = len(stats.FailedDependencyIDs)
		}
		logger.logf("\nFailed dependency IDs (first %d):", maxFailed)
		for i := 0; i < maxFailed; i++ {
			logger.logf("  - %s", stats.FailedDependencyIDs[i])
		}
		if len(stats.FailedDependencyIDs) > maxFailed {
			logger.logf("  ... and %d more", len(stats.FailedDependencyIDs)-maxFailed)
		}
	}

	if len(stats.MissingEntitySourceIDs) > 0 {
		maxMissing := 10
		if len(stats.MissingEntitySourceIDs) < maxMissing {
			maxMissing = len(stats.MissingEntitySourceIDs)
		}
		logger.logf("\nMissing source entity IDs (first %d):", maxMissing)
		for i := 0; i < maxMissing; i++ {
			logger.logf("  - %s", stats.MissingEntitySourceIDs[i])
		}
		if len(stats.MissingEntitySourceIDs) > maxMissing {
			logger.logf("  ... and %d more", len(stats.MissingEntitySourceIDs)-maxMissing)
		}
	}

	if len(stats.MissingEntityTargetIDs) > 0 {
		maxMissing := 10
		if len(stats.MissingEntityTargetIDs) < maxMissing {
			maxMissing = len(stats.MissingEntityTargetIDs)
		}
		logger.logf("\nMissing target entity IDs (first %d):", maxMissing)
		for i := 0; i < maxMissing; i++ {
			logger.logf("  - %s", stats.MissingEntityTargetIDs[i])
		}
		if len(stats.MissingEntityTargetIDs) > maxMissing {
			logger.logf("  ... and %d more", len(stats.MissingEntityTargetIDs)-maxMissing)
		}
	}

	logger.logf("\nMigration completed with %d successes and %d failures", stats.SuccessfulMigrations, stats.FailedMigrations)
}

// promptForConfirmation asks for user confirmation
func promptForConfirmation(message string) bool {
	if !interactive {
		return true // Auto-confirm in non-interactive mode
	}

	var response string
	fmt.Printf("%s (y/n): ", message)
	_, err := fmt.Scanln(&response)
	if err != nil {
		return false
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

func main() {
	// Parse command-line flags
	flag.Parse()

	// Create logger
	logger, err := newMigrationLogger(logFile)
	if err != nil {
		fmt.Printf("Error creating logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.close()

	logger.logf("Starting issue dependency migration tool")
	logger.logf("Database: %s", dbPath)
	logger.logf("Dry run: %v", dryRun)
	logger.logf("Batch size: %d", batchSize)
	logger.logf("Force: %v", force)

	// Open database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		logger.logf("Error opening database: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	// Validate database
	if err := validateDb(db); err != nil {
		logger.logf("Error validating database: %v", err)
		os.Exit(1)
	}

	// Get migration stats
	totalDependencies, err := getMigrationStats(db)
	if err != nil {
		logger.logf("Error getting migration stats: %v", err)
		os.Exit(1)
	}

	// Confirm migration
	if !dryRun && interactive {
		if !promptForConfirmation(fmt.Sprintf("Migrate %d dependencies?", totalDependencies)) {
			logger.logf("Migration canceled by user")
			os.Exit(0)
		}
	}

	// Migrate issue dependencies
	stats, err := migrateIssueDependencies(db, logger, dryRun, batchSize, force)
	if err != nil {
		logger.logf("Error during migration: %v", err)
		os.Exit(1)
	}

	// Print statistics
	printStats(stats, logger)

	// Exit with appropriate status code
	if stats.FailedMigrations > 0 {
		os.Exit(1)
	}
}