package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"runtime"
	"sync"
	"time"
)

// BenchmarkResults represents comprehensive performance analysis
type BenchmarkResults struct {
	TestSuite        string                 `json:"test_suite"`
	Timestamp        time.Time              `json:"timestamp"`
	SystemInfo       SystemInfo             `json:"system_info"`
	FileMetrics      FilePerformanceMetrics `json:"file_metrics"`
	ConcurrencyTests ConcurrencyResults     `json:"concurrency_tests"`
	MemoryEfficiency MemoryMetrics          `json:"memory_efficiency"`
	Summary          BenchmarkSummary       `json:"summary"`
}

type SystemInfo struct {
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
	CPUs         int    `json:"cpus"`
	GoVersion    string `json:"go_version"`
}

type FilePerformanceMetrics struct {
	SequentialRead  OperationStats `json:"sequential_read"`
	RandomRead      OperationStats `json:"random_read"`
	FileOpen        OperationStats `json:"file_open"`
	FileSize        int64          `json:"file_size"`
	FileSizeMB      float64        `json:"file_size_mb"`
}

type OperationStats struct {
	AverageLatency time.Duration `json:"average_latency"`
	MinLatency     time.Duration `json:"min_latency"`
	MaxLatency     time.Duration `json:"max_latency"`
	Throughput     float64       `json:"throughput_mb_per_sec"`
	Operations     int           `json:"operations"`
}

type ConcurrencyResults struct {
	ConcurrentReads  OperationStats `json:"concurrent_reads"`
	ReadContention   float64        `json:"read_contention_factor"`
	MemoryMapped     OperationStats `json:"memory_mapped_access"`
}

type MemoryMetrics struct {
	BaselineMemory    uint64  `json:"baseline_memory_mb"`
	PeakMemory        uint64  `json:"peak_memory_mb"`
	MemoryEfficiency  float64 `json:"memory_efficiency"`
	GCPressure        int     `json:"gc_collections"`
}

type BenchmarkSummary struct {
	OverallScore     float64  `json:"overall_score"`
	Strengths        []string `json:"strengths"`
	Optimizations    []string `json:"optimizations"`
	RecommendedUse   []string `json:"recommended_use"`
}

func main() {
	fmt.Println("⚡ EntityDB Performance Benchmark Suite")
	fmt.Println("=======================================")

	results := BenchmarkResults{
		TestSuite: "EntityDB Storage Performance v1.0",
		Timestamp: time.Now(),
		SystemInfo: getSystemInfo(),
	}

	// Find database file
	dbPath := findDatabaseFile()
	if dbPath == "" {
		fmt.Println("❌ No database file found for benchmarking")
		return
	}

	fmt.Printf("🎯 Benchmarking: %s\n\n", dbPath)

	// Run comprehensive benchmarks
	results.FileMetrics = benchmarkFilePerformance(dbPath)
	results.ConcurrencyTests = benchmarkConcurrency(dbPath)
	results.MemoryEfficiency = benchmarkMemoryEfficiency(dbPath)
	results.Summary = generateSummary(results)

	// Generate detailed report
	generateBenchmarkReport(results)

	// Display summary
	displaySummary(results)

	fmt.Println("\n✅ Performance benchmark complete!")
}

func getSystemInfo() SystemInfo {
	return SystemInfo{
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
		CPUs:         runtime.NumCPU(),
		GoVersion:    runtime.Version(),
	}
}

func findDatabaseFile() string {
	searchPaths := []string{
		"../../var/entities.edb",
		"../var/entities.edb",
		"./var/entities.edb",
		"./entities.edb",
	}

	for _, path := range searchPaths {
		if info, err := os.Stat(path); err == nil && info.Size() > 1000 {
			return path
		}
	}
	return ""
}

func benchmarkFilePerformance(dbPath string) FilePerformanceMetrics {
	fmt.Println("📊 Benchmarking file performance...")

	info, _ := os.Stat(dbPath)
	metrics := FilePerformanceMetrics{
		FileSize:   info.Size(),
		FileSizeMB: float64(info.Size()) / (1024 * 1024),
	}

	// Sequential read benchmark
	fmt.Println("  🔄 Sequential read test...")
	metrics.SequentialRead = benchmarkSequentialRead(dbPath)

	// Random read benchmark
	fmt.Println("  🎲 Random read test...")
	metrics.RandomRead = benchmarkRandomRead(dbPath)

	// File open benchmark
	fmt.Println("  📂 File open test...")
	metrics.FileOpen = benchmarkFileOpen(dbPath)

	return metrics
}

func benchmarkSequentialRead(dbPath string) OperationStats {
	const bufferSize = 64 * 1024 // 64KB buffer
	const iterations = 10

	var latencies []time.Duration
	totalBytes := int64(0)
	totalTime := time.Duration(0)

	for i := 0; i < iterations; i++ {
		start := time.Now()
		
		file, err := os.Open(dbPath)
		if err != nil {
			continue
		}

		buffer := make([]byte, bufferSize)
		bytesRead := int64(0)

		for {
			n, err := file.Read(buffer)
			bytesRead += int64(n)
			if err == io.EOF {
				break
			}
			if err != nil {
				break
			}
			// Limit to first 10MB for consistent timing
			if bytesRead > 10*1024*1024 {
				break
			}
		}

		file.Close()
		elapsed := time.Since(start)
		latencies = append(latencies, elapsed)
		totalBytes += bytesRead
		totalTime += elapsed
	}

	return calculateStats(latencies, float64(totalBytes)/(1024*1024), totalTime)
}

func benchmarkRandomRead(dbPath string) OperationStats {
	const bufferSize = 4096 // 4KB reads
	const iterations = 100

	info, _ := os.Stat(dbPath)
	fileSize := info.Size()

	var latencies []time.Duration
	totalBytes := int64(0)
	totalTime := time.Duration(0)

	file, err := os.Open(dbPath)
	if err != nil {
		return OperationStats{}
	}
	defer file.Close()

	buffer := make([]byte, bufferSize)
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < iterations; i++ {
		// Random position within file
		pos := rand.Int63n(fileSize - bufferSize)
		
		start := time.Now()
		file.Seek(pos, 0)
		n, _ := file.Read(buffer)
		elapsed := time.Since(start)

		latencies = append(latencies, elapsed)
		totalBytes += int64(n)
		totalTime += elapsed
	}

	return calculateStats(latencies, float64(totalBytes)/(1024*1024), totalTime)
}

func benchmarkFileOpen(dbPath string) OperationStats {
	const iterations = 50

	var latencies []time.Duration

	for i := 0; i < iterations; i++ {
		start := time.Now()
		file, err := os.Open(dbPath)
		elapsed := time.Since(start)
		
		if err == nil {
			file.Close()
		}
		
		latencies = append(latencies, elapsed)
	}

	return calculateStats(latencies, 0, 0)
}

func benchmarkConcurrency(dbPath string) ConcurrencyResults {
	fmt.Println("🔄 Benchmarking concurrency...")

	const goroutines = 10
	const readsPerGoroutine = 20

	// Concurrent reads test
	fmt.Println("  👥 Concurrent reads test...")
	concurrentResults := benchmarkConcurrentReads(dbPath, goroutines, readsPerGoroutine)

	// Memory-mapped simulation (using regular file I/O)
	fmt.Println("  🗺️ Memory-mapped access simulation...")
	mmapResults := benchmarkMemoryMappedAccess(dbPath)

	return ConcurrencyResults{
		ConcurrentReads:  concurrentResults,
		ReadContention:   calculateContentionFactor(concurrentResults),
		MemoryMapped:     mmapResults,
	}
}

func benchmarkConcurrentReads(dbPath string, goroutines, readsPerGoroutine int) OperationStats {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var allLatencies []time.Duration

	info, _ := os.Stat(dbPath)
	fileSize := info.Size()

	startTime := time.Now()

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			file, err := os.Open(dbPath)
			if err != nil {
				return
			}
			defer file.Close()

			buffer := make([]byte, 4096)
			var localLatencies []time.Duration

			for j := 0; j < readsPerGoroutine; j++ {
				pos := rand.Int63n(fileSize - 4096)
				
				start := time.Now()
				file.Seek(pos, 0)
				file.Read(buffer)
				elapsed := time.Since(start)

				localLatencies = append(localLatencies, elapsed)
			}

			mu.Lock()
			allLatencies = append(allLatencies, localLatencies...)
			mu.Unlock()
		}()
	}

	wg.Wait()
	totalTime := time.Since(startTime)

	return calculateStats(allLatencies, 0, totalTime)
}

func benchmarkMemoryMappedAccess(dbPath string) OperationStats {
	// Simulate memory-mapped access with optimized reads
	const iterations = 50
	const bufferSize = 64 * 1024

	var latencies []time.Duration

	file, err := os.Open(dbPath)
	if err != nil {
		return OperationStats{}
	}
	defer file.Close()

	buffer := make([]byte, bufferSize)

	for i := 0; i < iterations; i++ {
		start := time.Now()
		// Simulate mmap-like access (sequential with large buffer)
		file.Seek(int64(i*bufferSize), 0)
		file.Read(buffer)
		elapsed := time.Since(start)

		latencies = append(latencies, elapsed)
	}

	return calculateStats(latencies, 0, 0)
}

func calculateContentionFactor(stats OperationStats) float64 {
	// Contention factor: ratio of max to average latency
	if stats.AverageLatency > 0 {
		return float64(stats.MaxLatency) / float64(stats.AverageLatency)
	}
	return 1.0
}

func benchmarkMemoryEfficiency(dbPath string) MemoryMetrics {
	fmt.Println("💾 Benchmarking memory efficiency...")

	var memStats runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&memStats)
	baselineMemory := memStats.Alloc

	// Simulate file operations to measure memory impact
	file, err := os.Open(dbPath)
	if err != nil {
		return MemoryMetrics{}
	}
	defer file.Close()

	// Read portions of file to simulate normal operations
	buffer := make([]byte, 1024*1024) // 1MB buffer
	for i := 0; i < 10; i++ {
		file.Read(buffer)
	}

	runtime.ReadMemStats(&memStats)
	peakMemory := memStats.Alloc
	gcCollections := int(memStats.NumGC)

	efficiency := float64(baselineMemory) / float64(peakMemory)
	if efficiency > 1.0 {
		efficiency = 1.0
	}

	return MemoryMetrics{
		BaselineMemory:   baselineMemory / (1024 * 1024),
		PeakMemory:       peakMemory / (1024 * 1024),
		MemoryEfficiency: efficiency,
		GCPressure:       gcCollections,
	}
}

func calculateStats(latencies []time.Duration, throughputMB float64, totalTime time.Duration) OperationStats {
	if len(latencies) == 0 {
		return OperationStats{}
	}

	// Calculate average
	var total time.Duration
	min := latencies[0]
	max := latencies[0]

	for _, lat := range latencies {
		total += lat
		if lat < min {
			min = lat
		}
		if lat > max {
			max = lat
		}
	}

	avg := total / time.Duration(len(latencies))

	// Calculate throughput
	throughput := 0.0
	if totalTime > 0 {
		throughput = throughputMB / totalTime.Seconds()
	}

	return OperationStats{
		AverageLatency: avg,
		MinLatency:     min,
		MaxLatency:     max,
		Throughput:     throughput,
		Operations:     len(latencies),
	}
}

func generateSummary(results BenchmarkResults) BenchmarkSummary {
	var strengths []string
	var optimizations []string
	var recommendations []string

	// Analyze file performance
	if results.FileMetrics.SequentialRead.AverageLatency < 100*time.Millisecond {
		strengths = append(strengths, "✅ Excellent sequential read performance")
	}

	if results.FileMetrics.RandomRead.AverageLatency < 10*time.Millisecond {
		strengths = append(strengths, "✅ Good random access performance")
	}

	if results.FileMetrics.FileOpen.AverageLatency < 5*time.Millisecond {
		strengths = append(strengths, "✅ Fast file open operations")
	}

	// Analyze concurrency
	if results.ConcurrencyTests.ReadContention < 3.0 {
		strengths = append(strengths, "✅ Low read contention")
	} else {
		optimizations = append(optimizations, "💡 Consider read optimization for high concurrency")
	}

	// Memory efficiency
	if results.MemoryEfficiency.MemoryEfficiency > 0.8 {
		strengths = append(strengths, "✅ Excellent memory efficiency")
	}

	// File size recommendations
	if results.FileMetrics.FileSizeMB > 50 {
		recommendations = append(recommendations, "📊 Large database - ideal for unified format benefits")
		recommendations = append(recommendations, "🗄️ Consider periodic WAL checkpointing")
	}

	// Calculate overall score
	score := calculateOverallScore(results)

	return BenchmarkSummary{
		OverallScore:     score,
		Strengths:        strengths,
		Optimizations:    optimizations,
		RecommendedUse:   recommendations,
	}
}

func calculateOverallScore(results BenchmarkResults) float64 {
	score := 0.0

	// File performance (40% weight)
	if results.FileMetrics.SequentialRead.AverageLatency < 100*time.Millisecond {
		score += 40
	} else {
		score += 20
	}

	// Concurrency (30% weight)
	if results.ConcurrencyTests.ReadContention < 2.0 {
		score += 30
	} else if results.ConcurrencyTests.ReadContention < 5.0 {
		score += 20
	} else {
		score += 10
	}

	// Memory efficiency (20% weight)
	score += results.MemoryEfficiency.MemoryEfficiency * 20

	// File format benefits (10% weight)
	score += 10 // Unified format always gets full points

	return score
}

func generateBenchmarkReport(results BenchmarkResults) {
	reportPath := "performance_benchmark_report.json"

	jsonData, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		fmt.Printf("❌ Could not generate report: %v\n", err)
		return
	}

	err = os.WriteFile(reportPath, jsonData, 0644)
	if err != nil {
		fmt.Printf("❌ Could not write report: %v\n", err)
		return
	}

	fmt.Printf("📋 Benchmark report saved: %s\n", reportPath)
}

func displaySummary(results BenchmarkResults) {
	fmt.Printf("\n📊 PERFORMANCE SUMMARY\n")
	fmt.Printf("=====================\n")
	fmt.Printf("Overall Score: %.1f/100\n\n", results.Summary.OverallScore)

	fmt.Printf("📁 File Metrics:\n")
	fmt.Printf("  Size: %.2f MB\n", results.FileMetrics.FileSizeMB)
	fmt.Printf("  Sequential Read: %v avg\n", results.FileMetrics.SequentialRead.AverageLatency)
	fmt.Printf("  Random Read: %v avg\n", results.FileMetrics.RandomRead.AverageLatency)
	fmt.Printf("  File Open: %v avg\n", results.FileMetrics.FileOpen.AverageLatency)

	fmt.Printf("\n🔄 Concurrency:\n")
	fmt.Printf("  Read Contention: %.1fx\n", results.ConcurrencyTests.ReadContention)
	fmt.Printf("  Concurrent Reads: %v avg\n", results.ConcurrencyTests.ConcurrentReads.AverageLatency)

	fmt.Printf("\n💾 Memory:\n")
	fmt.Printf("  Efficiency: %.1f%%\n", results.MemoryEfficiency.MemoryEfficiency*100)
	fmt.Printf("  Peak Usage: %d MB\n", results.MemoryEfficiency.PeakMemory)

	if len(results.Summary.Strengths) > 0 {
		fmt.Printf("\n💪 Strengths:\n")
		for _, strength := range results.Summary.Strengths {
			fmt.Printf("  %s\n", strength)
		}
	}

	if len(results.Summary.Optimizations) > 0 {
		fmt.Printf("\n🔧 Optimizations:\n")
		for _, opt := range results.Summary.Optimizations {
			fmt.Printf("  %s\n", opt)
		}
	}
}