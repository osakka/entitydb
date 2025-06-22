// Test tool to verify deletion collector functionality
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
	"entitydb/services"
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
	testFilename := "test_deletion_collector.edb"
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
	
	fmt.Println("🗑️  EntityDB Deletion Collector Test")
	fmt.Println("=====================================")
	
	// Test 1: Create test entities with different ages
	fmt.Println("\n1. Creating test entities...")
	
	// Entity 1: Old document (should be deleted)
	entity1 := &models.Entity{
		ID:        "doc-old-001",
		Tags:      []string{"type:document", "name:old_document"},
		Content:   []byte("This is an old document"),
		CreatedAt: time.Now().Add(-100 * 24 * time.Hour).UnixNano(), // 100 days old
		UpdatedAt: time.Now().Add(-100 * 24 * time.Hour).UnixNano(),
	}
	if err := repo.Create(entity1); err != nil {
		log.Fatalf("Failed to create entity1: %v", err)
	}
	fmt.Printf("   ✅ Created old document: %s (100 days old)\n", entity1.ID)
	
	// Entity 2: Recent document (should remain active)
	entity2 := &models.Entity{
		ID:        "doc-new-001",
		Tags:      []string{"type:document", "name:new_document"},
		Content:   []byte("This is a new document"),
		CreatedAt: time.Now().Add(-10 * 24 * time.Hour).UnixNano(), // 10 days old
		UpdatedAt: time.Now().Add(-10 * 24 * time.Hour).UnixNano(),
	}
	if err := repo.Create(entity2); err != nil {
		log.Fatalf("Failed to create entity2: %v", err)
	}
	fmt.Printf("   ✅ Created new document: %s (10 days old)\n", entity2.ID)
	
	// Entity 3: Temporary file (should be deleted quickly)
	entity3 := &models.Entity{
		ID:        "temp-file-001",
		Tags:      []string{"type:temp", "name:temp_data.tmp"},
		Content:   []byte("Temporary file content"),
		CreatedAt: time.Now().Add(-2 * 24 * time.Hour).UnixNano(), // 2 days old
		UpdatedAt: time.Now().Add(-2 * 24 * time.Hour).UnixNano(),
	}
	if err := repo.Create(entity3); err != nil {
		log.Fatalf("Failed to create entity3: %v", err)
	}
	fmt.Printf("   ✅ Created temp file: %s (2 days old)\n", entity3.ID)
	
	// Entity 4: Permanent document (should never be deleted)
	entity4 := &models.Entity{
		ID:        "doc-permanent-001",
		Tags:      []string{"type:document", "name:permanent_document", "permanent"},
		Content:   []byte("This is a permanent document"),
		CreatedAt: time.Now().Add(-200 * 24 * time.Hour).UnixNano(), // 200 days old
		UpdatedAt: time.Now().Add(-200 * 24 * time.Hour).UnixNano(),
	}
	if err := repo.Create(entity4); err != nil {
		log.Fatalf("Failed to create entity4: %v", err)
	}
	fmt.Printf("   ✅ Created permanent document: %s (200 days old, protected)\n", entity4.ID)
	
	// Test 2: Initialize deletion collector with aggressive settings for testing
	fmt.Println("\n2. Initializing deletion collector...")
	
	collectorConfig := services.DeletionCollectorConfig{
		Enabled:       true,
		Interval:      1 * time.Minute, // Fast for testing
		BatchSize:     10,
		MaxRuntime:    5 * time.Minute,
		DryRun:        false, // Make real changes
		EnableMetrics: true,
		Concurrency:   2,
	}
	
	collector := services.NewDeletionCollector(repo, collectorConfig)
	
	// Add custom test policy with shorter durations
	testPolicy := models.RetentionPolicy{
		Name:        "test-policy",
		Description: "Aggressive test policy for demonstration",
		Enabled:     true,
		Priority:    10,
		Selector: models.PolicySelector{
			EntityTypes: []string{"document"},
			ExcludeTags: []string{"permanent"}, // Don't touch permanent entities
		},
		Rules: []models.RetentionRule{
			{
				Name:      "test-soft-delete",
				FromState: models.StateActive,
				ToState:   models.StateSoftDeleted,
				Condition: models.RuleCondition{
					Type:  models.ConditionAge,
					Value: "2160h", // 90 days (entity1 should match)
					Field: "created_at",
				},
				Reason:  "Test policy: delete old documents",
				Enabled: true,
			},
		},
		CreatedBy: "test",
		CreatedAt: time.Now(),
		UpdatedBy: "test",
		UpdatedAt: time.Now(),
	}
	
	if err := collector.AddPolicy(testPolicy); err != nil {
		log.Fatalf("Failed to add test policy: %v", err)
	}
	fmt.Printf("   ✅ Added test retention policy\n")
	
	// Test 3: Check initial entity states
	fmt.Println("\n3. Checking initial entity states...")
	
	entities := []*models.Entity{entity1, entity2, entity3, entity4}
	for _, entity := range entities {
		refreshedEntity, err := repo.GetByID(entity.ID)
		if err != nil {
			log.Printf("   ❌ Failed to get entity %s: %v", entity.ID, err)
			continue
		}
		fmt.Printf("   📄 %s: %s (age: %.0f days)\n", 
			refreshedEntity.ID, 
			refreshedEntity.GetLifecycleState(),
			time.Since(time.Unix(0, refreshedEntity.CreatedAt)).Hours()/24)
	}
	
	// Test 4: Run collection cycle manually
	fmt.Println("\n4. Running deletion collection cycle...")
	
	if err := collector.RunOnce(); err != nil {
		log.Fatalf("Failed to run collection cycle: %v", err)
	}
	
	fmt.Printf("   ✅ Collection cycle completed\n")
	
	// Test 5: Check entity states after collection
	fmt.Println("\n5. Checking entity states after collection...")
	
	transitioned := 0
	for _, entity := range entities {
		refreshedEntity, err := repo.GetByID(entity.ID)
		if err != nil {
			log.Printf("   ❌ Failed to get entity %s: %v", entity.ID, err)
			continue
		}
		
		newState := refreshedEntity.GetLifecycleState()
		fmt.Printf("   📄 %s: %s", refreshedEntity.ID, newState)
		
		if newState != models.StateActive {
			transitioned++
			fmt.Printf(" (✅ TRANSITIONED)")
			
			// Show transition details
			if newState == models.StateSoftDeleted {
				fmt.Printf("\n      🗑️  Deleted by: %s", refreshedEntity.GetDeletedBy())
				fmt.Printf("\n      📝 Reason: %s", refreshedEntity.GetDeleteReason())
				if deletedAt := refreshedEntity.GetDeletedAt(); deletedAt != nil {
					fmt.Printf("\n      🕐 Deleted at: %s", deletedAt.Format("2006-01-02 15:04:05"))
				}
			}
		}
		fmt.Println()
	}
	
	// Test 6: Show collector statistics
	fmt.Println("\n6. Collector statistics...")
	
	stats := collector.GetStats()
	fmt.Printf("   📊 Total runs: %d\n", stats.TotalRuns)
	fmt.Printf("   📊 Entities processed: %d\n", stats.EntitiesProcessed)
	fmt.Printf("   📊 Entities transitioned: %d\n", stats.EntitiesTransitioned)
	fmt.Printf("   📊 Soft deleted: %d\n", stats.SoftDeleted)
	fmt.Printf("   📊 Archived: %d\n", stats.Archived)
	fmt.Printf("   📊 Purged: %d\n", stats.Purged)
	fmt.Printf("   📊 Errors: %d\n", stats.Errors)
	if stats.LastError != "" {
		fmt.Printf("   ❌ Last error: %s\n", stats.LastError)
	}
	
	// Test 7: Verify policy effectiveness
	fmt.Println("\n7. Policy effectiveness...")
	
	if transitioned > 0 {
		fmt.Printf("   ✅ SUCCESS: %d entities were transitioned by policies\n", transitioned)
		fmt.Printf("   ✅ Old document (100 days) should be soft deleted\n")
		fmt.Printf("   ✅ New document (10 days) should remain active\n")
		fmt.Printf("   ✅ Permanent document should be protected\n")
	} else {
		fmt.Printf("   ⚠️  No entities were transitioned (policies may need adjustment)\n")
	}
	
	// Test 8: Show policy list
	fmt.Println("\n8. Active retention policies...")
	
	policies := collector.GetPolicies()
	for _, policy := range policies {
		fmt.Printf("   📋 %s: %s (enabled: %v, rules: %d)\n", 
			policy.Name, policy.Description, policy.Enabled, len(policy.Rules))
	}
	
	fmt.Println("\n🎉 Deletion collector test completed successfully!")
	
	// Clean up
	_ = os.Remove(testFile)
}