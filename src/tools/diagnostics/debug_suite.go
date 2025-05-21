package main

import (
	"entitydb/logger"
	"entitydb/storage/binary"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// DebugSuite integrates all debugging tools
type DebugSuite struct {
	dataPath        string
	outputDir       string
	verbose         bool
	
	// Components
	diagnosticLogger  *binary.DiagnosticLogger
	integrityManager  *binary.IntegrityManager
	performanceMetrics *binary.PerformanceMetrics
	dataPathTracer    *binary.DataPathTracer
}

// NewDebugSuite creates a new debug suite
func NewDebugSuite(dataPath, outputDir string, verbose bool) (*DebugSuite, error) {
	// Create output directory if needed
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %v", err)
	}
	
	// Initialize suite
	suite := &DebugSuite{
		dataPath:  dataPath,
		outputDir: outputDir,
		verbose:   verbose,
	}
	
	// Initialize diagnostic logger
	var err error
	suite.diagnosticLogger, err = binary.NewDiagnosticLogger(outputDir, verbose)
	if err != nil {
		return nil, fmt.Errorf("failed to create diagnostic logger: %v", err)
	}
	
	// Initialize integrity manager
	suite.integrityManager, err = binary.NewIntegrityManager(dataPath, suite.diagnosticLogger)
	if err != nil {
		logger.Warn("Failed to create integrity manager: %v", err)
		// Non-fatal, continue without integrity validation
	}
	
	// Initialize performance metrics
	suite.performanceMetrics, err = binary.NewPerformanceMetrics(outputDir, 1, true)
	if err != nil {
		logger.Warn("Failed to create performance metrics: %v", err)
		// Non-fatal, continue without metrics
	}
	
	// Initialize data path tracer
	suite.dataPathTracer, err = binary.NewDataPathTracer(outputDir, 1)
	if err != nil {
		logger.Warn("Failed to create data path tracer: %v", err)
		// Non-fatal, continue without tracing
	}
	
	return suite, nil
}

// RunFullDiagnostics runs a complete diagnostic suite
func (ds *DebugSuite) RunFullDiagnostics() error {
	fmt.Println("Starting EntityDB Full Diagnostics...")
	startTime := time.Now()
	
	// Run binary file analysis first
	if err := ds.runBinaryAnalysis(); err != nil {
		return err
	}
	
	// Run WAL validation
	if err := ds.runWALValidation(); err != nil {
		logger.Error("WAL validation failed: %v", err)
		// Continue with other diagnostics
	}
	
	// Run entity validation
	if err := ds.runEntityValidation(); err != nil {
		logger.Error("Entity validation failed: %v", err)
		// Continue with other diagnostics
	}
	
	// Run corruption scan
	if err := ds.runCorruptionScan(); err != nil {
		logger.Error("Corruption scan failed: %v", err)
		// Continue with other diagnostics
	}
	
	// Generate final report
	if err := ds.generateReport(); err != nil {
		logger.Error("Report generation failed: %v", err)
	}
	
	fmt.Printf("Full diagnostics completed in %.2f seconds\n", time.Since(startTime).Seconds())
	fmt.Printf("Results saved to: %s\n", ds.outputDir)
	
	return nil
}

// runBinaryAnalysis runs binary file analysis
func (ds *DebugSuite) runBinaryAnalysis() error {
	fmt.Println("\n=== Binary File Analysis ===")
	
	// Create analyzer with the right options
	analyzerOptions := map[string]bool{
		"extract":  true,
		"validate": true,
		"repair":   false,
		"verbose":  ds.verbose,
	}
	
	analyzer, err := binary.NewBinaryAnalyzer(
		filepath.Join(ds.dataPath, "entities.ebf"),
		filepath.Join(ds.outputDir, "binary_analysis"),
		analyzerOptions,
	)
	if err != nil {
		return fmt.Errorf("failed to create binary analyzer: %v", err)
	}
	defer analyzer.Close()
	
	// Run analysis
	if err := analyzer.Analyze(); err != nil {
		return fmt.Errorf("binary analysis failed: %v", err)
	}
	
	return nil
}

// runWALValidation runs WAL validation
func (ds *DebugSuite) runWALValidation() error {
	fmt.Println("\n=== WAL Validation ===")
	
	// Create validator with the right options
	validatorOptions := map[string]bool{
		"dryRun":            true,
		"validateChecksums": true,
		"stopOnError":       false,
		"validateIntegrity": true,
	}
	
	validator, err := binary.NewWALValidator(ds.dataPath, ds.diagnosticLogger, validatorOptions)
	if err != nil {
		return fmt.Errorf("failed to create WAL validator: %v", err)
	}
	defer validator.Close()
	
	// Run validation
	status, err := validator.Validate()
	if err != nil {
		return fmt.Errorf("WAL validation failed: %v", err)
	}
	
	// Generate validation report
	report, err := validator.GetValidationReport()
	if err != nil {
		logger.Error("Failed to generate WAL validation report: %v", err)
	} else {
		reportPath := filepath.Join(ds.outputDir, "wal_validation_report.json")
		if err := os.WriteFile(reportPath, report, 0644); err != nil {
			logger.Error("Failed to write WAL validation report: %v", err)
		}
	}
	
	fmt.Printf("WAL Validation Results:\n")
	fmt.Printf("  Entries Read: %d\n", status.EntriesRead)
	fmt.Printf("  Entries Valid: %d\n", status.EntriesReplayed)
	fmt.Printf("  Entries Invalid: %d\n", status.EntriesInvalid)
	
	return nil
}

// runEntityValidation runs entity validation
func (ds *DebugSuite) runEntityValidation() error {
	fmt.Println("\n=== Entity Validation ===")
	
	// Create entity repository
	repo, err := binary.NewEntityRepository(ds.dataPath)
	if err != nil {
		return fmt.Errorf("failed to create repository: %v", err)
	}
	
	// Create validator
	validator := binary.NewEntityValidator(binary.ValidationNormal, map[string]bool{
		"autoRepair":        false,
		"validateChecksums": true,
		"validateTemporal":  true,
	})
	
	// List all entities
	fmt.Println("Listing entities...")
	entities, err := repo.List()
	if err != nil {
		return fmt.Errorf("failed to list entities: %v", err)
	}
	
	fmt.Printf("Found %d entities to validate\n", len(entities))
	
	// Validate a sample of entities
	sampleSize := 10
	if len(entities) < sampleSize {
		sampleSize = len(entities)
	}
	
	fmt.Printf("Validating %d sample entities...\n", sampleSize)
	
	validCount := 0
	for i := 0; i < sampleSize; i++ {
		result := validator.Validate(entities[i])
		
		if result.Valid {
			validCount++
		} else if ds.verbose {
			fmt.Printf("Entity %s validation failed: %s\n", 
				entities[i].ID, 
				strings.Join(result.Errors, "; "))
		}
	}
	
	fmt.Printf("Sample validation results: %d/%d valid (%.1f%%)\n", 
		validCount, sampleSize, float64(validCount) * 100.0 / float64(sampleSize))
	
	return nil
}

// runCorruptionScan runs a corruption scan
func (ds *DebugSuite) runCorruptionScan() error {
	fmt.Println("\n=== Corruption Scan ===")
	
	// Create scanner with the right options
	scannerOptions := map[string]bool{
		"deepScan":    false,
		"validateAll": false,
		"repair":      false,
		"verbose":     ds.verbose,
	}
	
	// Create and run the corruption scanner (without using the tool directly)
	args := []string{
		"-data", ds.dataPath,
		"-output", filepath.Join(ds.outputDir, "corruption_scan"),
		"-concurrency", "4",
	}
	
	// Add options as flags
	if scannerOptions["deepScan"] {
		args = append(args, "-deep")
	}
	if scannerOptions["validateAll"] {
		args = append(args, "-all")
	}
	if scannerOptions["repair"] {
		args = append(args, "-repair")
	}
	if scannerOptions["verbose"] {
		args = append(args, "-verbose")
	}
	
	fmt.Printf("Running corruption scan with: %v\n", args)
	
	// In a real implementation, we would call the scanner directly
	// For this example, we'll just log this step
	logger.Info("Corruption scan would execute with args: %v", args)
	// In a complete implementation, we would:
	// 1. Run the corruption scan
	// 2. Parse the results
	
	// For this example, we simulate success
	logger.Info("Simulated corruption scan completed successfully")
	return nil
}

// generateReport generates a final diagnostic report
func (ds *DebugSuite) generateReport() error {
	fmt.Println("\n=== Generating Final Report ===")
	
	reportPath := filepath.Join(ds.outputDir, "diagnostic_report.txt")
	reportFile, err := os.Create(reportPath)
	if err != nil {
		return fmt.Errorf("failed to create report file: %v", err)
	}
	defer reportFile.Close()
	
	// Write header
	fmt.Fprintf(reportFile, "EntityDB Diagnostic Report\n")
	fmt.Fprintf(reportFile, "Generated: %s\n", time.Now().Format(time.RFC3339))
	fmt.Fprintf(reportFile, "Data Path: %s\n", ds.dataPath)
	fmt.Fprintf(reportFile, "\n")
	
	// Include performance metrics
	if ds.performanceMetrics != nil {
		metrics := ds.performanceMetrics.GetMetrics()
		
		fmt.Fprintf(reportFile, "Performance Metrics:\n")
		fmt.Fprintf(reportFile, "  Timestamp: %s\n", metrics["timestamp"])
		fmt.Fprintf(reportFile, "  Uptime: %.2f seconds\n", metrics["uptime_seconds"])
		
		fmt.Fprintf(reportFile, "  Operation Counts:\n")
		counts := metrics["operation_counts"].(map[string]int64)
		for op, count := range counts {
			fmt.Fprintf(reportFile, "    %s: %d\n", op, count)
		}
		
		fmt.Fprintf(reportFile, "  Throughput (ops/sec):\n")
		throughput := metrics["throughput_ops_per_sec"].(map[string]float64)
		for op, rate := range throughput {
			fmt.Fprintf(reportFile, "    %s: %.2f\n", op, rate)
		}
		
		fmt.Fprintf(reportFile, "  Average Duration (microseconds):\n")
		durations := metrics["avg_duration_us"].(map[string]float64)
		for op, duration := range durations {
			fmt.Fprintf(reportFile, "    %s: %.2f\n", op, duration)
		}
		
		fmt.Fprintf(reportFile, "\n")
	}
	
	// Include diagnostic logger stats
	if ds.diagnosticLogger != nil {
		metrics := ds.diagnosticLogger.GetMetricsSnapshot()
		
		fmt.Fprintf(reportFile, "I/O Metrics:\n")
		fmt.Fprintf(reportFile, "  Total Reads: %d (%.2f ms avg)\n", 
			metrics.TotalReads, metrics.AvgReadTimeMs)
		fmt.Fprintf(reportFile, "  Total Writes: %d (%.2f ms avg)\n", 
			metrics.TotalWrites, metrics.AvgWriteTimeMs)
		fmt.Fprintf(reportFile, "  Total Bytes Read: %d\n", metrics.TotalBytesRead)
		fmt.Fprintf(reportFile, "  Total Bytes Written: %d\n", metrics.TotalBytesWrite)
		fmt.Fprintf(reportFile, "  Cache Hit Ratio: %.2f%%\n", 
			metrics.CacheHitRatio * 100.0)
		
		fmt.Fprintf(reportFile, "\n")
	}
	
	// Include integrity check results
	if ds.integrityManager != nil {
		verified, corrupted, repaired := ds.integrityManager.GetStats()
		
		fmt.Fprintf(reportFile, "Data Integrity:\n")
		fmt.Fprintf(reportFile, "  Entities Verified: %d\n", verified)
		fmt.Fprintf(reportFile, "  Entities Corrupted: %d (%.2f%%)\n", 
			corrupted, float64(corrupted) * 100.0 / float64(verified))
		fmt.Fprintf(reportFile, "  Entities Repaired: %d\n", repaired)
		
		fmt.Fprintf(reportFile, "\n")
	}
	
	// Final advice section
	fmt.Fprintf(reportFile, "Recommendations:\n")
	
	// Add specific recommendations based on findings
	// (In a real implementation, this would analyze the results and provide
	// targeted recommendations)
	fmt.Fprintf(reportFile, "  1. Review corruption scan logs for details on any corrupted entities\n")
	fmt.Fprintf(reportFile, "  2. Check WAL validation report for any transaction log issues\n")
	fmt.Fprintf(reportFile, "  3. Consider running a full integrity check with repairs if corruption is detected\n")
	fmt.Fprintf(reportFile, "  4. Review binary analysis report for file format consistency\n")
	
	fmt.Fprintf(reportFile, "\n")
	fmt.Fprintf(reportFile, "For performance optimization, consider reviewing performance metrics\n")
	fmt.Fprintf(reportFile, "For data recovery, consider using the WAL replay tool\n")
	
	return nil
}

// Close closes the debug suite and any open resources
func (ds *DebugSuite) Close() error {
	var lastErr error
	
	// Close all components
	if ds.diagnosticLogger != nil {
		if err := ds.diagnosticLogger.Close(); err != nil {
			lastErr = err
		}
	}
	
	if ds.performanceMetrics != nil {
		if err := ds.performanceMetrics.Close(); err != nil {
			lastErr = err
		}
	}
	
	if ds.dataPathTracer != nil {
		if err := ds.dataPathTracer.Close(); err != nil {
			lastErr = err
		}
	}
	
	return lastErr
}

func main() {
	// Parse command-line flags
	dataPath := flag.String("data", "/opt/entitydb/var", "Path to the data directory")
	outputDir := flag.String("output", "./diagnostics", "Directory for diagnostic output")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	mode := flag.String("mode", "full", "Diagnostic mode (full, analyze, validate, wal, corruption)")
	
	flag.Parse()
	
	// Configure logging
	if *verbose {
		logger.SetLogLevel("debug")
	} else {
		logger.SetLogLevel("info")
	}
	
	// Create debug suite
	suite, err := NewDebugSuite(*dataPath, *outputDir, *verbose)
	if err != nil {
		fmt.Printf("Error creating debug suite: %v\n", err)
		os.Exit(1)
	}
	defer suite.Close()
	
	// Run diagnostics based on mode
	var runErr error
	
	switch *mode {
	case "full":
		runErr = suite.RunFullDiagnostics()
	case "analyze":
		runErr = suite.runBinaryAnalysis()
	case "validate":
		runErr = suite.runEntityValidation()
	case "wal":
		runErr = suite.runWALValidation()
	case "corruption":
		runErr = suite.runCorruptionScan()
	default:
		fmt.Printf("Unknown mode: %s\n", *mode)
		flag.Usage()
		os.Exit(1)
	}
	
	if runErr != nil {
		fmt.Printf("Diagnostics failed: %v\n", runErr)
		os.Exit(1)
	}
}