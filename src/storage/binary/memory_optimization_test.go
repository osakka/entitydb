package binary

import (
	"entitydb/models"
	"entitydb/config"
	"runtime"
	"sync"
	"testing"
	"time"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
)

// TestBoundedStringInterning tests the bounded string interning with LRU eviction
func TestBoundedStringInterning(t *testing.T) {
	// Set small limits for testing
	models.SetMaxSize(100)
	models.SetMemoryLimit(10 * 1024) // 10KB
	
	// Clear any existing data
	models.Clear()
	
	// Test 1: Basic interning functionality
	str1 := models.Intern("test-string")
	str2 := models.Intern("test-string")
	if str1 != str2 {
		t.Error("String interning failed: same string returned different pointers")
	}
	
	// Test 2: LRU eviction when size limit exceeded
	for i := 0; i < 150; i++ {
		models.Intern(fmt.Sprintf("string-%d", i))
	}
	
	stats := models.Stats()
	if stats.Size > 100 {
		t.Errorf("String cache exceeded max size: %d > 100", stats.Size)
	}
	
	// Test 3: Memory limit enforcement
	bigString := make([]byte, 5*1024) // 5KB string
	for i := range bigString {
		bigString[i] = 'a' + byte(i%26)
	}
	
	// Try to intern several big strings
	for i := 0; i < 5; i++ {
		models.Intern(string(bigString) + fmt.Sprintf("-%d", i))
	}
	
	if models.MemoryUsed() > models.Stats().MemoryLimit {
		t.Errorf("Memory limit exceeded: %d > %d", models.MemoryUsed(), models.Stats().MemoryLimit)
	}
	
	// Test 4: Pressure cleanup
	interner := models.GetDefaultStringInterner()
	interner.TriggerPressureCleanup(0.9) // 90% pressure
	
	newStats := models.Stats()
	if newStats.Size >= stats.Size {
		t.Error("Pressure cleanup did not reduce cache size")
	}
	
	t.Logf("String interning test complete - Final stats: Size=%d, Memory=%d, Evictions=%d", 
		newStats.Size, newStats.MemoryUsed, newStats.Evictions)
}

// TestBoundedEntityCache tests the bounded entity cache with LRU eviction
func TestBoundedEntityCache(t *testing.T) {
	cache := NewBoundedEntityCache(50, 50*1024*1024) // 50 entities, 50MB
	
	// Test 1: Basic cache operations
	entity1 := &models.Entity{
		ID:      "test-entity-1",
		Tags:    []string{"type:test", "name:entity1"},
		Content: []byte("test content"),
	}
	
	cache.Put(entity1.ID, entity1)
	
	retrieved, found := cache.Get(entity1.ID)
	if !found || retrieved == nil {
		t.Error("Failed to retrieve cached entity")
	}
	
	// Test 2: LRU eviction
	for i := 0; i < 60; i++ {
		entity := &models.Entity{
			ID:      fmt.Sprintf("entity-%d", i),
			Tags:    []string{"type:test"},
			Content: []byte(fmt.Sprintf("content-%d", i)),
		}
		cache.Put(entity.ID, entity)
	}
	
	stats := cache.Stats()
	if stats.Size > 50 {
		t.Errorf("Cache exceeded max size: %d > 50", stats.Size)
	}
	
	// Test 3: Memory pressure cleanup
	cache.TriggerPressureCleanup(0.85)
	
	newStats := cache.Stats()
	if newStats.Size >= stats.Size {
		t.Error("Pressure cleanup did not reduce cache size")
	}
	
	t.Logf("Entity cache test complete - Final stats: Size=%d, Memory=%d, HitRate=%.2f%%", 
		newStats.Size, newStats.MemoryUsed, newStats.HitRate*100)
}

// TestMetricsRecursionPrevention tests that metrics collection doesn't cause recursion
func TestMetricsRecursionPrevention(t *testing.T) {
	// Setup
	SetMetricsOperation(false)
	
	// Test 1: Basic recursion prevention
	SetMetricsOperation(true)
	if !IsMetricsOperation() {
		t.Error("Metrics operation flag not set")
	}
	
	// Simulate nested metrics call
	SetMetricsOperation(true) // Should increment depth
	SetMetricsOperation(false) // Should decrement
	
	if !IsMetricsOperation() {
		t.Error("Metrics operation flag cleared too early")
	}
	
	SetMetricsOperation(false) // Should clear completely
	if IsMetricsOperation() {
		t.Error("Metrics operation flag not cleared")
	}
	
	// Test 2: Global disable
	DisableMetricsGlobally()
	if !IsMetricsOperation() {
		t.Error("Global metrics disable not working")
	}
	
	EnableMetricsGlobally()
	if IsMetricsOperation() {
		t.Error("Global metrics enable not working")
	}
}

// TestMemoryMonitor tests the memory monitoring system
func TestMemoryMonitor(t *testing.T) {
	monitor := NewMemoryMonitor()
	
	// Test 1: Basic monitoring
	monitor.Start()
	defer monitor.Stop()
	
	// Wait for initial monitoring
	time.Sleep(100 * time.Millisecond)
	
	stats := monitor.GetStats()
	if stats.CurrentPressure < 0 || stats.CurrentPressure > 1 {
		t.Errorf("Invalid pressure reading: %f", stats.CurrentPressure)
	}
	
	// Test 2: Pressure callbacks
	callbackCount := 0
	
	monitor.AddPressureCallback(func(pressure float64, level PressureLevel) {
		callbackCount++
		t.Logf("Callback fired: pressure=%.2f, level=%s", pressure, level.String())
	})
	
	// Force a check with simulated medium pressure
	monitor.checkMemoryPressure()
	
	// Note: In real scenarios, callbacks only fire for medium+ pressure
	// For testing, we'd need to simulate high memory usage
	t.Logf("Callback count: %d (may be 0 if memory pressure is low)", callbackCount)
}

// TestTemporalRetention tests temporal data retention under memory pressure
func TestTemporalRetention(t *testing.T) {
	// Create a mock repository for testing
	cfg := config.Load()
	tempDir := t.TempDir()
	cfg.DataPath = tempDir
	// Create var directory
	varDir := filepath.Join(tempDir, "var")
	os.MkdirAll(varDir, 0755)
	// Update database paths to use temp directory
	cfg.DatabaseFilename = filepath.Join(varDir, "entities.edb")
	
	factory := &RepositoryFactory{}
	repo, err := factory.CreateRepository(cfg)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	// Type assertion to get the concrete type for Close()
	if binaryRepo, ok := repo.(*EntityRepository); ok {
		defer binaryRepo.Close()
	}
	
	retention := NewTemporalRetentionManager(repo)
	
	// Create entity with many temporal tags
	entity := &models.Entity{
		ID:   "metric-test",
		Tags: []string{"type:metric", "name:test_metric"},
	}
	
	// Add temporal tags
	now := time.Now()
	for i := 0; i < 1500; i++ {
		timestamp := now.Add(-time.Duration(i) * time.Minute).UnixNano()
		tag := fmt.Sprintf("%d|value:%d", timestamp, rand.Intn(100))
		entity.Tags = append(entity.Tags, tag)
	}
	
	originalTagCount := len(entity.Tags)
	
	// Test retention application
	err = retention.ApplyRetention(entity)
	if err != nil {
		t.Errorf("Failed to apply retention: %v", err)
	}
	
	if len(entity.Tags) >= originalTagCount {
		t.Error("Retention did not reduce tag count")
	}
	
	// Test age-based cleanup
	err = retention.CleanupByAge(entity)
	if err != nil {
		t.Errorf("Failed age cleanup: %v", err)
	}
	
	t.Logf("Temporal retention test complete - Tags reduced from %d to %d", 
		originalTagCount, len(entity.Tags))
}

// TestIntegratedMemoryOptimization tests all optimizations working together
func TestIntegratedMemoryOptimization(t *testing.T) {
	// Setup small limits to trigger optimizations
	models.SetMaxSize(100)
	models.SetMemoryLimit(1024 * 1024) // 1MB
	
	cfg := config.Load()
	tempDir := t.TempDir()
	cfg.DataPath = tempDir
	// Create var directory
	varDir := filepath.Join(tempDir, "var")
	os.MkdirAll(varDir, 0755)
	// Update database paths to use temp directory
	cfg.DatabaseFilename = filepath.Join(varDir, "entities.edb")
	cfg.EntityCacheSize = 50
	cfg.EntityCacheMemoryLimit = 10 * 1024 * 1024 // 10MB
	
	factory := &RepositoryFactory{}
	repo, err := factory.CreateRepository(cfg)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	// Type assertion to get the concrete type for Close()
	if binaryRepo, ok := repo.(*EntityRepository); ok {
		defer binaryRepo.Close()
	}
	
	// Initialize memory monitor
	monitor := NewMemoryMonitor()
	monitor.Start()
	defer monitor.Stop()
	
	// Register pressure callbacks
	pressureEvents := 0
	monitor.AddPressureCallback(func(pressure float64, level PressureLevel) {
		pressureEvents++
		t.Logf("Memory pressure event: %.1f%% (%s)", pressure*100, level.String())
		
		// Trigger cleanup
		if level >= PressureHigh {
			models.GetDefaultStringInterner().TriggerPressureCleanup(pressure)
			if binaryRepo, ok := repo.(*EntityRepository); ok {
				binaryRepo.TriggerCachePressureCleanup(pressure)
			}
		}
	})
	
	// Simulate high load scenario
	var wg sync.WaitGroup
	errors := 0
	
	// Producer goroutines - create entities with temporal tags
	for p := 0; p < 5; p++ {
		wg.Add(1)
		go func(producerID int) {
			defer wg.Done()
			
			for i := 0; i < 100; i++ {
				entity := &models.Entity{
					ID:   fmt.Sprintf("entity-%d-%d", producerID, i),
					Tags: []string{"type:test", fmt.Sprintf("producer:%d", producerID)},
				}
				
				// Add temporal tags
				for j := 0; j < 50; j++ {
					tag := fmt.Sprintf("%d|value:%d", time.Now().UnixNano(), rand.Intn(1000))
					entity.Tags = append(entity.Tags, tag)
				}
				
				if err := repo.Create(entity); err != nil {
					errors++
				}
				
				time.Sleep(10 * time.Millisecond)
			}
		}(p)
	}
	
	// Consumer goroutines - read entities
	for c := 0; c < 3; c++ {
		wg.Add(1)
		go func(consumerID int) {
			defer wg.Done()
			
			for i := 0; i < 200; i++ {
				entityID := fmt.Sprintf("entity-%d-%d", rand.Intn(5), rand.Intn(100))
				_, _ = repo.GetByID(entityID)
				
				time.Sleep(5 * time.Millisecond)
			}
		}(c)
	}
	
	// Wait for all operations
	wg.Wait()
	
	// Final statistics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	stringStats := models.Stats()
	// Get repository performance stats if available
	repoStats := map[string]interface{}{"cacheHitRate": 0.0}
	if binaryRepo, ok := repo.(*EntityRepository); ok {
		repoStats = binaryRepo.GetStats()
	}
	monitorStats := monitor.GetStats()
	
	t.Logf("=== Final Memory Optimization Statistics ===")
	t.Logf("Heap In Use: %d MB", memStats.HeapInuse/(1024*1024))
	t.Logf("Heap Allocated: %d MB", memStats.HeapAlloc/(1024*1024))
	t.Logf("String Cache: Size=%d, Memory=%d KB, Evictions=%d", 
		stringStats.Size, stringStats.MemoryUsed/1024, stringStats.Evictions)
	t.Logf("Entity Cache: HitRate=%.1f%%", repoStats["cacheHitRate"])
	t.Logf("Memory Monitor: MaxPressure=%.1f%%, PressureEvents=%d", 
		monitorStats.MaxPressureObserved*100, pressureEvents)
	t.Logf("Errors: %d", errors)
	
	// Verify optimizations worked
	if stringStats.Evictions == 0 && stringStats.Size >= 100 {
		t.Error("String cache eviction not working")
	}
	
	if memStats.HeapInuse > 100*1024*1024 { // 100MB threshold
		t.Errorf("Memory usage too high: %d MB", memStats.HeapInuse/(1024*1024))
	}
}

// BenchmarkMemoryOptimizations benchmarks the performance impact
func BenchmarkMemoryOptimizations(b *testing.B) {
	cfg := config.Load()
	tempDir := b.TempDir()
	cfg.DataPath = tempDir
	// Create var directory
	varDir := filepath.Join(tempDir, "var")
	os.MkdirAll(varDir, 0755)
	// Update database paths to use temp directory
	cfg.DatabaseFilename = filepath.Join(varDir, "entities.edb")
	
	factory := &RepositoryFactory{}
	repo, err := factory.CreateRepository(cfg)
	if err != nil {
		b.Fatalf("Failed to create repository: %v", err)
	}
	// Type assertion to get the concrete type for Close()
	if binaryRepo, ok := repo.(*EntityRepository); ok {
		defer binaryRepo.Close()
	}
	
	// Pre-create test entities
	entities := make([]*models.Entity, 1000)
	for i := range entities {
		entities[i] = &models.Entity{
			ID:      fmt.Sprintf("bench-entity-%d", i),
			Tags:    []string{"type:benchmark", fmt.Sprintf("index:%d", i)},
			Content: []byte(fmt.Sprintf("benchmark content %d", i)),
		}
		
		// Add temporal tags
		for j := 0; j < 10; j++ {
			tag := fmt.Sprintf("%d|value:%d", time.Now().UnixNano(), rand.Intn(1000))
			entities[i].Tags = append(entities[i].Tags, tag)
		}
	}
	
	b.ResetTimer()
	
	b.Run("Create", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			entity := entities[i%len(entities)]
			_ = repo.Create(entity)
		}
	})
	
	b.Run("Read", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			entityID := fmt.Sprintf("bench-entity-%d", i%len(entities))
			_, _ = repo.GetByID(entityID)
		}
	})
	
	b.Run("StringIntern", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = models.Intern(fmt.Sprintf("benchmark-string-%d", i%1000))
		}
	})
}