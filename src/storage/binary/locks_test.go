package binary

import (
	"entitydb/models"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestConcurrentAccess(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	
	// Create repository
	repo, err := NewEntityRepository(tempDir)
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}
	
	// Test concurrent writes
	var wg sync.WaitGroup
	numGoroutines := 10
	entitiesPerGoroutine := 5
	
	wg.Add(numGoroutines)
	
	start := time.Now()
	
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()
			
			for j := 0; j < entitiesPerGoroutine; j++ {
				entity := models.NewEntity()
				entity.AddTag("type:test")
				entity.AddTag(fmt.Sprintf("goroutine:%d", goroutineID))
				entity.AddTag(fmt.Sprintf("iteration:%d", j))
				entity.AddTagWithValue("title", fmt.Sprintf("Test entity %d-%d", goroutineID, j))
				
				err := repo.Create(entity)
				if err != nil {
					t.Errorf("Failed to create entity: %v", err)
				}
			}
		}(i)
	}
	
	wg.Wait()
	
	elapsed := time.Since(start)
	t.Logf("Created %d entities in %v", numGoroutines*entitiesPerGoroutine, elapsed)
	
	// Wait a bit to ensure all indexing is complete
	time.Sleep(100 * time.Millisecond)
	
	// Test concurrent reads
	wg.Add(numGoroutines)
	
	start = time.Now()
	
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()
			
			// List by tag
			searchTag := fmt.Sprintf("goroutine:%d", goroutineID)
			entities, err := repo.ListByTag(searchTag)
			if err != nil {
				t.Errorf("Failed to list entities: %v", err)
			}
			
			if len(entities) != entitiesPerGoroutine {
				t.Errorf("Expected %d entities for tag '%s', got %d", entitiesPerGoroutine, searchTag, len(entities))
				// Debug: check what tags are actually in the index
				all, _ := repo.List()
				t.Logf("Total entities in repo: %d", len(all))
				for i, entity := range all {
					if i < 3 { // Just show first few for debugging
						cleanTags := entity.GetTagsWithoutTimestamp()
						t.Logf("Entity %s tags: %v", entity.ID, cleanTags)
					}
				}
			}
		}(i)
	}
	
	wg.Wait()
	
	elapsed = time.Since(start)
	t.Logf("Read %d entities in %v", numGoroutines*entitiesPerGoroutine, elapsed)
	
	// Get locking statistics
	stats := repo.lockManager.GetStats()
	t.Logf("Lock stats: ReadLocks=%d, WriteLocks=%d, WaitTime=%v", 
		stats.ReadLocks, stats.WriteLocks, stats.WaitTime)
}

func BenchmarkConcurrentWrites(b *testing.B) {
	tempDir := b.TempDir()
	
	repo, err := NewEntityRepository(tempDir)
	if err != nil {
		b.Fatalf("Failed to create repository: %v", err)
	}
	
	b.ResetTimer()
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			entity := models.NewEntity()
			entity.AddTag("type:benchmark")
			entity.AddTagWithValue("title", "Benchmark entity")
			
			err := repo.Create(entity)
			if err != nil {
				b.Errorf("Failed to create entity: %v", err)
			}
		}
	})
}

func BenchmarkConcurrentReads(b *testing.B) {
	tempDir := b.TempDir()
	
	repo, err := NewEntityRepository(tempDir)
	if err != nil {
		b.Fatalf("Failed to create repository: %v", err)
	}
	
	// Create some test entities
	for i := 0; i < 100; i++ {
		entity := models.NewEntity()
		entity.AddTag("type:benchmark")
		entity.AddTag(fmt.Sprintf("index:%d", i))
		entity.AddTagWithValue("title", fmt.Sprintf("Benchmark entity %d", i))
		
		err := repo.Create(entity)
		if err != nil {
			b.Fatalf("Failed to create entity: %v", err)
		}
	}
	
	b.ResetTimer()
	
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			entities, err := repo.ListByTag("type:benchmark")
			if err != nil {
				b.Errorf("Failed to list entities: %v", err)
			}
			
			if len(entities) != 100 {
				b.Errorf("Expected 100 entities, got %d", len(entities))
			}
		}
	})
}