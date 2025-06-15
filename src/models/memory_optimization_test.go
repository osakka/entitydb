// Package models provides comprehensive tests for memory optimization features.
//
// These tests validate the performance and correctness of the exotic memory
// optimization algorithms implemented in EntityDB.
package models

import (
	"fmt"
	"runtime"
	"strconv"
	"testing"
	"time"
)

// BenchmarkTagProcessing compares standard vs optimized tag processing performance.
func BenchmarkTagProcessing(b *testing.B) {
	// Create test entity with temporal tags
	entity := NewEntity()
	
	// Add test tags
	for i := 0; i < 100; i++ {
		tag := fmt.Sprintf("key%d:value%d", i, i)
		entity.AddTag(tag)
	}
	
	b.Run("Standard", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			entity.invalidateTagValueCache()
			entity.buildTagValueCache()
		}
	})
	
	b.Run("Optimized", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			entity.invalidateTagValueCache()
			entity.buildTagValueCacheOptimized()
		}
	})
}

// BenchmarkStringInterning compares standard vs lock-free string interning.
func BenchmarkStringInterning(b *testing.B) {
	testStrings := make([]string, 1000)
	for i := range testStrings {
		testStrings[i] = fmt.Sprintf("test_string_%d", i)
	}
	
	b.Run("Standard", func(b *testing.B) {
		Clear() // Clear standard interner
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			for _, s := range testStrings {
				Intern(s)
			}
		}
	})
	
	b.Run("LockFree", func(b *testing.B) {
		globalLockFreeIntern.Clear()
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			for _, s := range testStrings {
				InternLockFree(s)
			}
		}
	})
}

// BenchmarkZeroCopyTagParsing tests zero-copy tag parsing performance.
func BenchmarkZeroCopyTagParsing(b *testing.B) {
	temporalTag := "1640995200000000000|type:user"
	
	b.Run("StringsSplit", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Simulate standard parsing
			tag := temporalTag
			timestamp, tagPortion := parseTemporalTagStandard(tag)
			_ = timestamp
			_ = tagPortion
		}
	})
	
	b.Run("ZeroCopy", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			tagBytes := stringToBytesZeroCopy(temporalTag)
			if view, ok := NewTemporalTagView(tagBytes); ok {
				timestamp, _ := view.ParseTimestamp()
				tagString := view.TagString()
				_ = timestamp
				_ = tagString
			}
		}
	})
}

// parseTemporalTagStandard simulates the standard parsing approach.
func parseTemporalTagStandard(tag string) (int64, string) {
	parts := make([]string, 0, 2) // Simulate strings.Split allocation
	
	// Find separator
	for i, r := range tag {
		if r == '|' {
			parts = append(parts, tag[:i])
			parts = append(parts, tag[i+1:])
			break
		}
	}
	
	if len(parts) != 2 {
		return 0, tag
	}
	
	timestamp, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, parts[1]
	}
	
	return timestamp, parts[1]
}

// TestZeroCopyCorrectness validates that zero-copy operations produce correct results.
func TestZeroCopyCorrectness(t *testing.T) {
	testCases := []struct {
		input             string
		expectedTimestamp int64
		expectedTag       string
	}{
		{
			input:             "1640995200000000000|type:user",
			expectedTimestamp: 1640995200000000000,
			expectedTag:       "type:user",
		},
		{
			input:             "1640995200000000001|status:active",
			expectedTimestamp: 1640995200000000001,
			expectedTag:       "status:active",
		},
		{
			input:             "1640995200000000002|project:entitydb",
			expectedTimestamp: 1640995200000000002,
			expectedTag:       "project:entitydb",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			// Test zero-copy parsing
			tagBytes := stringToBytesZeroCopy(tc.input)
			view, ok := NewTemporalTagView(tagBytes)
			if !ok {
				t.Fatalf("Failed to parse temporal tag: %s", tc.input)
			}
			
			// Check timestamp
			timestamp, ok := view.ParseTimestamp()
			if !ok {
				t.Fatalf("Failed to parse timestamp from: %s", tc.input)
			}
			if timestamp != tc.expectedTimestamp {
				t.Errorf("Expected timestamp %d, got %d", tc.expectedTimestamp, timestamp)
			}
			
			// Check tag
			tag := view.TagString()
			if tag != tc.expectedTag {
				t.Errorf("Expected tag %s, got %s", tc.expectedTag, tag)
			}
			
			// Test key-value splitting
			if keyView, valueView, hasValue := view.SplitKeyValue(); hasValue {
				// Find the colon separator
				colonPos := -1
				for i, c := range tc.expectedTag {
					if c == ':' {
						colonPos = i
						break
					}
				}
				if colonPos == -1 {
					t.Fatalf("Expected tag to have colon separator: %s", tc.expectedTag)
				}
				expectedKey := tc.expectedTag[:colonPos]
				expectedValue := tc.expectedTag[colonPos+1:]
				
				if keyView.String() != expectedKey {
					t.Errorf("Expected key %s, got %s", expectedKey, keyView.String())
				}
				if valueView.String() != expectedValue {
					t.Errorf("Expected value %s, got %s", expectedValue, valueView.String())
				}
			}
		})
	}
}

// TestLockFreeStringInterning validates lock-free string interning correctness.
func TestLockFreeStringInterning(t *testing.T) {
	interner := NewLockFreeStringIntern()
	
	// Test basic interning
	s1 := "test_string_1"
	s2 := "test_string_1" // Same content
	s3 := "test_string_2" // Different content
	
	interned1 := interner.Intern(s1)
	interned2 := interner.Intern(s2)
	interned3 := interner.Intern(s3)
	
	// Same content should return same interned instance
	if interned1 != interned2 {
		t.Error("Same strings should return same interned instance")
	}
	
	// Different content should return different instances
	if interned1 == interned3 {
		t.Error("Different strings should return different interned instances")
	}
	
	// Check statistics
	stats := interner.GetStats()
	if stats.TotalStrings < 2 {
		t.Errorf("Expected at least 2 interned strings, got %d", stats.TotalStrings)
	}
	
	if stats.HitCount == 0 {
		t.Error("Expected at least one cache hit")
	}
}

// TestOptimizedEntityOperations validates optimized entity operations.
func TestOptimizedEntityOperations(t *testing.T) {
	entity := NewEntity()
	
	// Test optimized tag addition
	entity.AddTagOptimized("type:user")
	entity.AddTagOptimized("status:active")
	entity.AddTagOptimized("project:entitydb")
	
	// Test optimized tag value retrieval
	if value := entity.GetTagValueOptimized("type"); value != "user" {
		t.Errorf("Expected 'user', got '%s'", value)
	}
	
	if value := entity.GetTagValueOptimized("status"); value != "active" {
		t.Errorf("Expected 'active', got '%s'", value)
	}
	
	// Test optimized tag checking
	if !entity.HasTagOptimized("type:user") {
		t.Error("Expected entity to have tag 'type:user'")
	}
	
	if entity.HasTagOptimized("type:admin") {
		t.Error("Expected entity to not have tag 'type:admin'")
	}
	
	// Test optimized tag removal
	removed := entity.RemoveTagOptimized("status:active")
	if !removed {
		t.Error("Expected tag removal to succeed")
	}
	
	if entity.HasTagOptimized("status:active") {
		t.Error("Expected tag to be removed")
	}
	
	// Test optimized tag update
	updated := entity.UpdateTagOptimized("type", "admin")
	if !updated {
		t.Error("Expected tag update to succeed")
	}
	
	if value := entity.GetTagValueOptimized("type"); value != "admin" {
		t.Errorf("Expected 'admin', got '%s'", value)
	}
}

// BenchmarkMemoryUsage compares memory usage between standard and optimized operations.
func BenchmarkMemoryUsage(b *testing.B) {
	// Measure memory usage for standard operations
	var m1, m2 runtime.MemStats
	
	b.Run("StandardOperations", func(b *testing.B) {
		runtime.GC()
		runtime.ReadMemStats(&m1)
		
		for i := 0; i < b.N; i++ {
			entity := NewEntity()
			for j := 0; j < 100; j++ {
				tag := fmt.Sprintf("key%d:value%d", j, j)
				entity.AddTag(tag)
			}
			
			// Trigger cache building
			entity.buildTagValueCache()
			
			// Read all values
			for j := 0; j < 100; j++ {
				key := fmt.Sprintf("key%d", j)
				_ = entity.GetTagValue(key)
			}
		}
		
		runtime.GC()
		runtime.ReadMemStats(&m2)
		standardAllocs := m2.TotalAlloc - m1.TotalAlloc
		b.ReportMetric(float64(standardAllocs), "bytes/op")
	})
	
	b.Run("OptimizedOperations", func(b *testing.B) {
		runtime.GC()
		runtime.ReadMemStats(&m1)
		
		for i := 0; i < b.N; i++ {
			entity := NewEntity()
			for j := 0; j < 100; j++ {
				tag := fmt.Sprintf("key%d:value%d", j, j)
				entity.AddTagOptimized(tag)
			}
			
			// Trigger cache building
			entity.buildTagValueCacheOptimized()
			
			// Read all values
			for j := 0; j < 100; j++ {
				key := fmt.Sprintf("key%d", j)
				_ = entity.GetTagValueOptimized(key)
			}
		}
		
		runtime.GC()
		runtime.ReadMemStats(&m2)
		optimizedAllocs := m2.TotalAlloc - m1.TotalAlloc
		b.ReportMetric(float64(optimizedAllocs), "bytes/op")
	})
}

// TestConcurrentLockFreeOperations validates thread safety of lock-free operations.
func TestConcurrentLockFreeOperations(t *testing.T) {
	interner := NewLockFreeStringIntern()
	
	// Run concurrent interning operations
	const numGoroutines = 10
	const numOperations = 1000
	
	done := make(chan bool, numGoroutines)
	
	for g := 0; g < numGoroutines; g++ {
		go func(goroutineID int) {
			for i := 0; i < numOperations; i++ {
				str := fmt.Sprintf("goroutine_%d_string_%d", goroutineID, i)
				interned := interner.Intern(str)
				
				// Verify correctness
				if interned != str {
					t.Errorf("Interned string doesn't match original: %s != %s", interned, str)
				}
			}
			done <- true
		}(g)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
	
	// Verify final state
	stats := interner.GetStats()
	if stats.TotalStrings == 0 {
		t.Error("Expected some interned strings")
	}
	
	if stats.LookupCount != numGoroutines*numOperations {
		t.Errorf("Expected %d lookups, got %d", numGoroutines*numOperations, stats.LookupCount)
	}
}

// BenchmarkAdaptiveBufferPool tests the performance of the adaptive buffer pool.
func BenchmarkAdaptiveBufferPool(b *testing.B) {
	sizes := []int{1024, 4096, 16384, 65536, 262144} // 1KB to 256KB
	
	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size%d", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				buf := GetAdaptive(size)
				// Simulate some work
				for j := 0; j < len(buf) && j < size; j++ {
					buf = append(buf, byte(j))
				}
				PutAdaptive(buf)
			}
		})
	}
}

// TestMemoryLeakDetection validates that optimizations don't introduce memory leaks.
func TestMemoryLeakDetection(t *testing.T) {
	var m1, m2 runtime.MemStats
	
	// Baseline measurement
	runtime.GC()
	runtime.ReadMemStats(&m1)
	
	// Perform many operations
	for i := 0; i < 10000; i++ {
		entity := NewEntity()
		
		// Add tags using optimized operations
		for j := 0; j < 10; j++ {
			tag := fmt.Sprintf("key%d:value%d", j, i)
			entity.AddTagOptimized(tag)
		}
		
		// Read values
		for j := 0; j < 10; j++ {
			key := fmt.Sprintf("key%d", j)
			_ = entity.GetTagValueOptimized(key)
		}
		
		// Remove tags
		for j := 0; j < 5; j++ {
			tag := fmt.Sprintf("key%d:value%d", j, i)
			entity.RemoveTagOptimized(tag)
		}
		
		// Update tags
		for j := 5; j < 10; j++ {
			key := fmt.Sprintf("key%d", j)
			newValue := fmt.Sprintf("updated%d", i)
			entity.UpdateTagOptimized(key, newValue)
		}
	}
	
	// Force garbage collection
	runtime.GC()
	runtime.GC() // Run twice to ensure cleanup
	time.Sleep(100 * time.Millisecond)
	runtime.ReadMemStats(&m2)
	
	// Check for excessive memory growth
	memoryGrowth := m2.Alloc - m1.Alloc
	if memoryGrowth > 50*1024*1024 { // 50MB threshold
		t.Errorf("Excessive memory growth detected: %d bytes", memoryGrowth)
	}
	
	t.Logf("Memory growth: %d bytes", memoryGrowth)
	t.Logf("Total allocations: %d", m2.TotalAlloc-m1.TotalAlloc)
	t.Logf("GC cycles: %d", m2.NumGC-m1.NumGC)
}