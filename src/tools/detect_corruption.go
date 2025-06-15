//go:build tool
package main

import (
	"bufio"
	"entitydb/logger"
	"entitydb/storage/binary"
	"entitydb/config"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// CorruptionScanner scans for corrupted entities
type CorruptionScanner struct {
	config             *config.Config
	outputDir          string
	concurrency        int
	deepScan           bool
	validateAll        bool
	repairMode         bool
	verbose            bool
	
	// Statistics
	entitiesScanned    int
	entitiesCorrupted  int
	entitiesRepaired   int
	errorsByType       map[string]int
	
	// Components
	validator          *binary.EntityValidator
	diagnosticLogger   *binary.DiagnosticLogger
}

// NewCorruptionScanner creates a new corruption scanner
func NewCorruptionScanner(cfg *config.Config, outputDir string, options map[string]bool, concurrency int) (*CorruptionScanner, error) {
	// Create output directory if needed
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %v", err)
	}
	
	// Initialize diagnostic logger
	diagnosticLogger, err := binary.NewDiagnosticLogger(outputDir, options["verbose"])
	if err != nil {
		return nil, fmt.Errorf("failed to create diagnostic logger: %v", err)
	}
	
	// Initialize validator
	validationLevel := binary.ValidationNormal
	if options["deepScan"] {
		validationLevel = binary.ValidationStrict
	}
	
	validator := binary.NewEntityValidator(validationLevel, map[string]bool{
		"autoRepair":        options["repair"],
		"validateChecksums": true,
		"validateTemporal":  options["deepScan"],
	})
	
	scanner := &CorruptionScanner{
		config:            cfg,
		outputDir:         outputDir,
		concurrency:       concurrency,
		deepScan:          options["deepScan"],
		validateAll:       options["validateAll"],
		repairMode:        options["repair"],
		verbose:           options["verbose"],
		entitiesScanned:   0,
		entitiesCorrupted: 0,
		entitiesRepaired:  0,
		errorsByType:      make(map[string]int),
		validator:         validator,
		diagnosticLogger:  diagnosticLogger,
	}
	
	return scanner, nil
}

// ScanForCorruption scans the repository for corrupted entities
func (cs *CorruptionScanner) ScanForCorruption() error {
	startTime := time.Now()
	
	// Create entity repository
	repo, err := binary.NewEntityRepositoryWithConfig(cs.config)
	if err != nil {
		return fmt.Errorf("failed to create repository: %v", err)
	}
	
	// Get all entities
	fmt.Println("Listing all entities...")
	entities, err := repo.List()
	if err != nil {
		return fmt.Errorf("failed to list entities: %v", err)
	}
	
	fmt.Printf("Found %d entities to scan\n", len(entities))
	
	// Create corruption report file
	reportPath := filepath.Join(cs.outputDir, fmt.Sprintf("corruption_scan_%s.log", time.Now().Format("20060102_150405")))
	reportFile, err := os.Create(reportPath)
	if err != nil {
		return fmt.Errorf("failed to create report file: %v", err)
	}
	defer reportFile.Close()
	
	// Create repair log file if in repair mode
	var repairFile *os.File
	if cs.repairMode {
		repairPath := filepath.Join(cs.outputDir, fmt.Sprintf("repair_log_%s.log", time.Now().Format("20060102_150405")))
		repairFile, err = os.Create(repairPath)
		if err != nil {
			return fmt.Errorf("failed to create repair log file: %v", err)
		}
		defer repairFile.Close()
	}
	
	// Write header to report file
	fmt.Fprintf(reportFile, "EntityDB Corruption Scan Report\n")
	fmt.Fprintf(reportFile, "Scan Time: %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(reportFile, "Data Path: %s\n", cs.config.DataPath)
	fmt.Fprintf(reportFile, "Options: deepScan=%v, validateAll=%v, repairMode=%v\n\n", 
		cs.deepScan, cs.validateAll, cs.repairMode)
	
	// Scan entities in parallel
	if cs.concurrency > 1 {
		fmt.Printf("Scanning with %d workers...\n", cs.concurrency)
		cs.scanEntitiesParallel(entities, repo, reportFile, repairFile)
	} else {
		fmt.Println("Scanning sequentially...")
		cs.scanEntitiesSerial(entities, repo, reportFile, repairFile)
	}
	
	// Write summary to report file
	fmt.Fprintf(reportFile, "\nScan Summary\n")
	fmt.Fprintf(reportFile, "Entities Scanned: %d\n", cs.entitiesScanned)
	fmt.Fprintf(reportFile, "Entities Corrupted: %d (%.2f%%)\n", 
		cs.entitiesCorrupted, float64(cs.entitiesCorrupted) * 100.0 / float64(cs.entitiesScanned))
	
	if cs.repairMode {
		fmt.Fprintf(reportFile, "Entities Repaired: %d (%.2f%%)\n", 
			cs.entitiesRepaired, float64(cs.entitiesRepaired) * 100.0 / float64(cs.entitiesCorrupted))
	}
	
	fmt.Fprintf(reportFile, "\nErrors by Type:\n")
	for errorType, count := range cs.errorsByType {
		fmt.Fprintf(reportFile, "  %s: %d\n", errorType, count)
	}
	
	fmt.Fprintf(reportFile, "\nScan Duration: %.2f seconds\n", time.Since(startTime).Seconds())
	
	// Print summary to console
	fmt.Printf("\nScan complete in %.2f seconds\n", time.Since(startTime).Seconds())
	fmt.Printf("Entities Scanned: %d\n", cs.entitiesScanned)
	fmt.Printf("Entities Corrupted: %d (%.2f%%)\n", 
		cs.entitiesCorrupted, float64(cs.entitiesCorrupted) * 100.0 / float64(cs.entitiesScanned))
	
	if cs.repairMode {
		fmt.Printf("Entities Repaired: %d (%.2f%%)\n", 
			cs.entitiesRepaired, float64(cs.entitiesRepaired) * 100.0 / float64(cs.entitiesCorrupted))
	}
	
	fmt.Printf("Report saved to: %s\n", reportPath)
	
	return nil
}

// scanEntitiesSerial scans entities serially
func (cs *CorruptionScanner) scanEntitiesSerial(entities []*models.Entity, repo models.EntityRepository, reportFile, repairFile *os.File) {
	progressInterval := len(entities) / 20 // 5% increments
	if progressInterval == 0 {
		progressInterval = 1
	}
	
	for i, entity := range entities {
		if i % progressInterval == 0 {
			fmt.Printf("Progress: %.1f%% (%d/%d)\n", float64(i) * 100.0 / float64(len(entities)), i, len(entities))
		}
		
		cs.scanEntity(entity, repo, reportFile, repairFile)
	}
}

// scanEntitiesParallel scans entities in parallel
func (cs *CorruptionScanner) scanEntitiesParallel(entities []*models.Entity, repo models.EntityRepository, reportFile, repairFile *os.File) {
	wg := sync.WaitGroup{}
	entityChan := make(chan *models.Entity, cs.concurrency)
	resultChan := make(chan string, cs.concurrency)
	repairChan := make(chan string, cs.concurrency)
	
	// Start worker goroutines
	for i := 0; i < cs.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			
			for entity := range entityChan {
				cs.scanEntity(entity, repo, nil, nil)
			}
		}()
	}
	
	// Start result writer goroutine
	var reportWg sync.WaitGroup
	reportWg.Add(1)
	go func() {
		defer reportWg.Done()
		
		for result := range resultChan {
			fmt.Fprintf(reportFile, "%s\n", result)
		}
	}()
	
	// Start repair writer goroutine if in repair mode
	if cs.repairMode {
		reportWg.Add(1)
		go func() {
			defer reportWg.Done()
			
			for result := range repairChan {
				fmt.Fprintf(repairFile, "%s\n", result)
			}
		}()
	}
	
	// Feed entities to workers
	progressInterval := len(entities) / 20 // 5% increments
	if progressInterval == 0 {
		progressInterval = 1
	}
	
	for i, entity := range entities {
		if i % progressInterval == 0 {
			fmt.Printf("Progress: %.1f%% (%d/%d)\n", float64(i) * 100.0 / float64(len(entities)), i, len(entities))
		}
		
		entityChan <- entity
	}
	
	// Close channels and wait for completion
	close(entityChan)
	wg.Wait()
	
	close(resultChan)
	if cs.repairMode {
		close(repairChan)
	}
	reportWg.Wait()
}

// scanEntity scans a single entity for corruption
func (cs *CorruptionScanner) scanEntity(entity *models.Entity, repo models.EntityRepository, reportFile, repairFile *os.File) {
	// If we're writing to files directly, lock for thread safety
	var mu sync.Mutex
	if reportFile != nil || repairFile != nil {
		mu.Lock()
		defer mu.Unlock()
	}
	
	cs.entitiesScanned++
	
	// For validateAll mode, validate all entities
	// For normal mode, only validate entities without type:user tag (skip user entities)
	if !cs.validateAll {
		isUserEntity := false
		for _, tag := range entity.Tags {
			tagValue := tag
			if strings.Contains(tag, "|") {
				parts := strings.SplitN(tag, "|", 2)
				tagValue = parts[1]
			}
			
			if tagValue == "type:user" {
				isUserEntity = true
				break
			}
		}
		
		if isUserEntity {
			return
		}
	}
	
	// Validate entity
	result := cs.validator.Validate(entity)
	
	if !result.Valid {
		cs.entitiesCorrupted++
		
		// Log corruption to diagnostic logger
		if cs.diagnosticLogger != nil {
			errorDetails := strings.Join(result.Errors, "; ")
			cs.diagnosticLogger.LogCorruption(entity.ID, "", "", errorDetails)
		}
		
		// Categorize errors
		for _, err := range result.Errors {
			errorType := "other"
			
			// Extract error type from message
			if strings.Contains(err, "checksum") {
				errorType = "checksum"
			} else if strings.Contains(err, "timestamp") {
				errorType = "timestamp"
			} else if strings.Contains(err, "tag") {
				errorType = "tag"
			} else if strings.Contains(err, "content") {
				errorType = "content"
			} else if strings.Contains(err, "ID") {
				errorType = "entity_id"
			}
			
			cs.errorsByType[errorType]++
		}
		
		// Report corruption
		report := fmt.Sprintf("[CORRUPT] Entity %s: %s", entity.ID, strings.Join(result.Errors, "; "))
		
		if reportFile != nil {
			fmt.Fprintf(reportFile, "%s\n", report)
		}
		
		// Try repair if enabled
		if cs.repairMode && result.RepairedEntity != nil {
			// Update entity in repository
			if err := repo.Update(result.RepairedEntity); err != nil {
				repairReport := fmt.Sprintf("[REPAIR_FAILED] Entity %s: %v", entity.ID, err)
				
				if repairFile != nil {
					fmt.Fprintf(repairFile, "%s\n", repairReport)
				}
				
				logger.Error("Failed to repair entity %s: %v", entity.ID, err)
			} else {
				cs.entitiesRepaired++
				
				repairReport := fmt.Sprintf("[REPAIRED] Entity %s: %s", 
					entity.ID, strings.Join(result.FixesApplied, "; "))
				
				if repairFile != nil {
					fmt.Fprintf(repairFile, "%s\n", repairReport)
				}
				
				logger.Info("Repaired entity %s", entity.ID)
			}
		}
	}
}

// Close closes the scanner and any open resources
func (cs *CorruptionScanner) Close() error {
	if cs.diagnosticLogger != nil {
		return cs.diagnosticLogger.Close()
	}
	return nil
}

func main() {
	// Initialize configuration system
	configManager := config.NewConfigManager(nil)
	configManager.RegisterFlags()
	
	// Additional tool-specific flags
	outputDir := flag.String("output", "./corruption_scan", "Directory for scan reports and logs")
	concurrency := flag.Int("concurrency", 4, "Number of concurrent scan workers")
	deepScan := flag.Bool("deep", false, "Perform deep scan (more thorough but slower)")
	validateAll := flag.Bool("all", false, "Validate all entities (including user entities)")
	repairMode := flag.Bool("repair", false, "Attempt to repair corrupted entities")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	
	flag.Parse()
	
	cfg, err := configManager.Initialize()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}
	
	// Configure logging
	if *verbose {
		logger.SetLogLevel("debug")
	} else {
		logger.SetLogLevel("info")
	}
	
	options := map[string]bool{
		"deepScan":    *deepScan,
		"validateAll": *validateAll,
		"repair":      *repairMode,
		"verbose":     *verbose,
	}
	
	// Create scanner
	scanner, err := NewCorruptionScanner(cfg, *outputDir, options, *concurrency)
	if err != nil {
		fmt.Printf("Error creating scanner: %v\n", err)
		os.Exit(1)
	}
	defer scanner.Close()
	
	// Perform scan
	fmt.Printf("Starting corruption scan in %s...\n", cfg.DataPath)
	if err := scanner.ScanForCorruption(); err != nil {
		fmt.Printf("Scan failed: %v\n", err)
		os.Exit(1)
	}
}