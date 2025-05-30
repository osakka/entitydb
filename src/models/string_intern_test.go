package models

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
	"unsafe"
)

func TestStringInterning(t *testing.T) {
	// Clear any existing interned strings
	Clear()
	
	// Test basic interning
	s1 := "test:tag:value"
	s2 := "test:tag:value"
	
	i1 := Intern(s1)
	i2 := Intern(s2)
	
	// They should point to the same memory
	if unsafe.Pointer(&i1) == unsafe.Pointer(&i2) {
		t.Error("Interned strings should share the same underlying memory")
	}
	
	// But their string headers should point to the same data
	if i1 != i2 {
		t.Error("Interned strings should be equal")
	}
	
	// Check size
	if Size() != 1 {
		t.Errorf("Expected 1 interned string, got %d", Size())
	}
}

func TestInternSlice(t *testing.T) {
	Clear()
	
	tags := []string{
		"type:document",
		"status:active",
		"type:document", // Duplicate
		"priority:high",
		"status:active", // Duplicate
	}
	
	InternSlice(tags)
	
	// Should have only 3 unique strings interned
	if Size() != 3 {
		t.Errorf("Expected 3 unique interned strings, got %d", Size())
	}
}

func TestConcurrentInterning(t *testing.T) {
	Clear()
	
	var wg sync.WaitGroup
	concurrency := 100
	iterations := 1000
	
	// Create a set of tags that will be repeated
	baseTags := []string{
		"type:user",
		"type:document", 
		"status:active",
		"status:pending",
		"priority:high",
		"priority:low",
	}
	
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				tag := baseTags[j%len(baseTags)]
				_ = Intern(tag)
			}
		}(i)
	}
	
	wg.Wait()
	
	// Should only have as many interned strings as unique base tags
	if Size() != len(baseTags) {
		t.Errorf("Expected %d interned strings, got %d", len(baseTags), Size())
	}
}

func BenchmarkStringInterning(b *testing.B) {
	Clear()
	
	// Common tags that would be repeated in a real system
	tags := []string{
		"type:document",
		"status:active",
		"priority:high",
		"rbac:role:admin",
		"rbac:perm:entity:view",
	}
	
	b.Run("WithInterning", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tag := tags[i%len(tags)]
			_ = Intern(tag)
		}
	})
	
	b.Run("WithoutInterning", func(b *testing.B) {
		b.ResetTimer()
		result := make([]string, 0, b.N)
		for i := 0; i < b.N; i++ {
			tag := tags[i%len(tags)]
			result = append(result, tag)
		}
		_ = result
	})
}

func BenchmarkMemoryUsage(b *testing.B) {
	b.Run("WithInterning", func(b *testing.B) {
		Clear()
		runtime.GC()
		
		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)
		
		// Create many duplicate strings
		for i := 0; i < 10000; i++ {
			for j := 0; j < 10; j++ {
				tag := fmt.Sprintf("tag:type:%d", j)
				_ = Intern(tag)
			}
		}
		
		runtime.GC()
		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)
		
		alloced := m2.Alloc - m1.Alloc
		b.ReportMetric(float64(alloced), "bytes")
		b.ReportMetric(float64(Size()), "unique_strings")
	})
	
	b.Run("WithoutInterning", func(b *testing.B) {
		runtime.GC()
		
		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)
		
		tags := make([]string, 0, 100000)
		// Create many duplicate strings
		for i := 0; i < 10000; i++ {
			for j := 0; j < 10; j++ {
				tag := fmt.Sprintf("tag:type:%d", j)
				tags = append(tags, tag)
			}
		}
		
		runtime.GC()
		var m2 runtime.MemStats
		runtime.ReadMemStats(&m2)
		
		alloced := m2.Alloc - m1.Alloc
		b.ReportMetric(float64(alloced), "bytes")
		b.ReportMetric(float64(len(tags)), "total_strings")
	})
}