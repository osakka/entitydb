package models

import (
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

// Benchmark functions moved to memory_optimization_test.go to avoid duplicates