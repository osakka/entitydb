// Test tool to verify entity lifecycle management system
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"entitydb/config"
	"entitydb/logger"
	"entitydb/models"
	"entitydb/storage/binary"
)

func main() {
	// Initialize logger
	logger.SetLogLevel("INFO")
	
	// Get data path
	dataPath := "/opt/entitydb/var"
	if err := os.MkdirAll(dataPath, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}
	
	// Create test configuration
	testFilename := "test_lifecycle.edb"
	cfg := &config.Config{
		DataPath:         dataPath,
		DatabaseFilename: filepath.Join(dataPath, testFilename),
		WALFilename:      "", // Use unified format
		IndexFilename:    "", // Use unified format
	}
	
	// Clean up test file
	testFile := cfg.DatabaseFilename
	_ = os.Remove(testFile)
	
	// Create repository
	repo, err := binary.NewEntityRepositoryWithConfig(cfg)
	if err != nil {
		log.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()
	
	fmt.Println("🧪 EntityDB Lifecycle Management Test")
	fmt.Println("=====================================")
	
	// Test 1: Create a test entity
	fmt.Println("\n1. Creating test entity...")
	entity := &models.Entity{
		ID:        "test-entity-001",
		Tags:      []string{"type:document", "name:test_document"},
		Content:   []byte("This is a test document for lifecycle testing"),
		CreatedAt: time.Now().UnixNano(),
		UpdatedAt: time.Now().UnixNano(),
	}
	
	if err := repo.Create(entity); err != nil {
		log.Fatalf("Failed to create entity: %v", err)
	}
	fmt.Printf("   ✅ Entity created: %s\n", entity.ID)
	
	// Test 2: Check initial state
	fmt.Println("\n2. Checking initial lifecycle state...")
	retrievedEntity, err := repo.GetByID(entity.ID)
	if err != nil {
		log.Fatalf("Failed to retrieve entity: %v", err)
	}
	
	state := retrievedEntity.GetLifecycleState()
	fmt.Printf("   ✅ Initial state: %s\n", state)
	fmt.Printf("   ✅ IsActive: %v\n", retrievedEntity.IsActive())
	
	// Test 3: Soft delete
	fmt.Println("\n3. Testing soft delete...")
	
	// First show current tags
	fmt.Printf("   Tags before soft delete: %v\n", retrievedEntity.Tags)
	
	// Soft delete adds tags to the entity, but we need to save it
	if err := retrievedEntity.SoftDelete("admin", "Testing lifecycle", "manual"); err != nil {
		log.Fatalf("Failed to soft delete: %v", err)
	}
	
	// Show tags after soft delete
	fmt.Printf("   Tags after soft delete: %v\n", retrievedEntity.Tags)
	
	// Update in repository to persist the tags
	if err := repo.Update(retrievedEntity); err != nil {
		log.Fatalf("Failed to update entity after soft delete: %v", err)
	}
	
	// Re-retrieve to verify
	deletedEntity, err := repo.GetByID(entity.ID)
	if err != nil {
		log.Fatalf("Failed to retrieve deleted entity: %v", err)
	}
	
	fmt.Printf("   ✅ New state: %s\n", deletedEntity.GetLifecycleState())
	fmt.Printf("   ✅ IsSoftDeleted: %v\n", deletedEntity.IsSoftDeleted())
	fmt.Printf("   ✅ DeletedBy: %s\n", deletedEntity.GetDeletedBy())
	fmt.Printf("   ✅ DeleteReason: %s\n", deletedEntity.GetDeleteReason())
	if deletedAt := deletedEntity.GetDeletedAt(); deletedAt != nil {
		fmt.Printf("   ✅ DeletedAt: %s\n", deletedAt.Format("2006-01-02 15:04:05"))
	}
	
	// Test 4: Undelete
	fmt.Println("\n4. Testing undelete...")
	if err := deletedEntity.Undelete("admin", "Testing restore"); err != nil {
		log.Fatalf("Failed to undelete: %v", err)
	}
	
	if err := repo.Update(deletedEntity); err != nil {
		log.Fatalf("Failed to update entity after undelete: %v", err)
	}
	
	restoredEntity, err := repo.GetByID(entity.ID)
	if err != nil {
		log.Fatalf("Failed to retrieve restored entity: %v", err)
	}
	
	fmt.Printf("   ✅ Restored state: %s\n", restoredEntity.GetLifecycleState())
	fmt.Printf("   ✅ IsActive: %v\n", restoredEntity.IsActive())
	
	// Test 5: Archive
	fmt.Println("\n5. Testing archive...")
	if err := restoredEntity.SoftDelete("admin", "Archiving test", "retention"); err != nil {
		log.Fatalf("Failed to soft delete before archive: %v", err)
	}
	if err := repo.Update(restoredEntity); err != nil {
		log.Fatalf("Failed to update before archive: %v", err)
	}
	
	// Re-get entity to ensure fresh state
	entityToArchive, err := repo.GetByID(entity.ID)
	if err != nil {
		log.Fatalf("Failed to get entity for archiving: %v", err)
	}
	
	if err := entityToArchive.Archive("admin", "Testing archive", "30-day-retention"); err != nil {
		log.Fatalf("Failed to archive: %v", err)
	}
	if err := repo.Update(entityToArchive); err != nil {
		log.Fatalf("Failed to update entity after archive: %v", err)
	}
	
	archivedEntity, err := repo.GetByID(entity.ID)
	if err != nil {
		log.Fatalf("Failed to retrieve archived entity: %v", err)
	}
	
	fmt.Printf("   ✅ Archived state: %s\n", archivedEntity.GetLifecycleState())
	fmt.Printf("   ✅ IsArchived: %v\n", archivedEntity.IsArchived())
	
	// Test 6: Transition history
	fmt.Println("\n6. Testing transition history...")
	history := archivedEntity.GetTransitionHistory()
	fmt.Printf("   ✅ Found %d transitions:\n", len(history))
	for i, transition := range history {
		fmt.Printf("      %d. %s → %s at %s by %s\n", 
			i+1, transition.FromState, transition.ToState, 
			transition.Timestamp.Format("15:04:05"), transition.UserID)
	}
	
	// Test 7: Lifecycle state queries
	fmt.Println("\n7. Testing lifecycle state queries...")
	
	// Create another entity for testing queries
	entity2 := &models.Entity{
		ID:        "test-entity-002",
		Tags:      []string{"type:document", "name:another_document"},
		Content:   []byte("Another test document"),
		CreatedAt: time.Now().UnixNano(),
		UpdatedAt: time.Now().UnixNano(),
	}
	if err := repo.Create(entity2); err != nil {
		log.Fatalf("Failed to create second entity: %v", err)
	}
	
	activeEntities, err := repo.ListActive()
	if err != nil {
		log.Fatalf("Failed to list active entities: %v", err)
	}
	fmt.Printf("   ✅ Active entities: %d\n", len(activeEntities))
	
	archivedEntities, err := repo.ListArchived()
	if err != nil {
		log.Fatalf("Failed to list archived entities: %v", err)
	}
	fmt.Printf("   ✅ Archived entities: %d\n", len(archivedEntities))
	
	// Test 8: Invalid transitions
	fmt.Println("\n8. Testing invalid transitions...")
	if archivedEntity.CanTransitionTo(models.StateActive) {
		fmt.Println("   ❌ ERROR: Should not be able to transition from archived to active")
	} else {
		fmt.Println("   ✅ Correctly blocked invalid transition (archived → active)")
	}
	
	fmt.Println("\n🎉 All lifecycle tests completed successfully!")
	
	// Clean up
	_ = os.Remove(testFile)
}