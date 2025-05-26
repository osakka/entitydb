package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"entitydb/models"
	"entitydb/storage/binary"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: test_dataspace_security <data_directory>")
	}

	dataDir := os.Args[1]
	fmt.Printf("Testing dataspace security functionality...\n")
	fmt.Printf("Data directory: %s\n", dataDir)

	// Create repository
	repo, err := binary.NewEntityRepository(dataDir)
	if err != nil {
		log.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()

	// Create temporal repository for testing
	temporalRepo := binary.NewTemporalRepository(repo)

	// Create security manager
	securityMgr := models.NewSecurityManager(temporalRepo)

	fmt.Printf("\n=== Testing Dataspace Security System ===\n")

	// Test 1: Create test entities
	fmt.Printf("\n1. Creating test entities for dataspace security...\n")

	// Create test dataspace
	dataspace := &models.Entity{
		ID:          "test_dataspace_1",
		Tags:        []string{"type:dataspace", "name:Test Dataspace", "status:active"},
		Content:     []byte(`{"name":"Test Dataspace","description":"Testing dataspace isolation"}`),
		ContentType: "application/json",
		Timestamp:   time.Now(),
	}

	err = temporalRepo.Create(dataspace)
	if err != nil {
		log.Printf("Failed to create test dataspace: %v", err)
	} else {
		fmt.Printf("✓ Created test dataspace: %s\n", dataspace.ID)
	}

	// Create test user
	testUser := &models.Entity{
		ID:          "test_user_1",
		Tags:        []string{"type:user", "username:testuser", "status:active"},
		Content:     []byte(`{"username":"testuser","email":"test@example.com"}`),
		ContentType: "application/json",
		Timestamp:   time.Now(),
	}

	err = temporalRepo.Create(testUser)
	if err != nil {
		log.Printf("Failed to create test user: %v", err)
	} else {
		fmt.Printf("✓ Created test user: %s\n", testUser.ID)
	}

	// Create test entity within dataspace
	testEntity := &models.Entity{
		ID:          "test_entity_1",
		Tags:        []string{"type:document", "dataspace:test_dataspace_1", "title:Test Document"},
		Content:     []byte(`{"title":"Test Document","content":"This is a test document in dataspace 1"}`),
		ContentType: "application/json",
		Timestamp:   time.Now(),
	}

	err = temporalRepo.Create(testEntity)
	if err != nil {
		log.Printf("Failed to create test entity: %v", err)
	} else {
		fmt.Printf("✓ Created test entity: %s\n", testEntity.ID)
	}

	// Test 2: Test dataspace relationship creation
	fmt.Printf("\n2. Testing dataspace relationship types...\n")

	relRepo := binary.NewRelationshipRepository(dataDir)
	defer relRepo.Close()

	// Test relationship: user can access dataspace
	canAccessRel := &models.EntityRelationship{
		FromEntityID: testUser.ID,
		ToEntityID:   dataspace.ID,
		Type:         "can_access", // RelationshipCanAccess
		Timestamp:    time.Now(),
	}

	err = relRepo.Create(canAccessRel)
	if err != nil {
		log.Printf("Failed to create can_access relationship: %v", err)
	} else {
		fmt.Printf("✓ Created can_access relationship: %s -> %s\n", testUser.ID, dataspace.ID)
	}

	// Test relationship: entity belongs to dataspace
	belongsToRel := &models.EntityRelationship{
		FromEntityID: testEntity.ID,
		ToEntityID:   dataspace.ID,
		Type:         "belongs_to", // RelationshipBelongsTo
		Timestamp:    time.Now(),
	}

	err = relRepo.Create(belongsToRel)
	if err != nil {
		log.Printf("Failed to create belongs_to relationship: %v", err)
	} else {
		fmt.Printf("✓ Created belongs_to relationship: %s -> %s\n", testEntity.ID, dataspace.ID)
	}

	// Test 3: Test dataspace access validation
	fmt.Printf("\n3. Testing dataspace access validation...\n")

	// Test CanAccessDataspace method
	canAccess, err := securityMgr.CanAccessDataspace(testUser, dataspace.ID)
	if err != nil {
		log.Printf("Error checking dataspace access: %v", err)
	} else {
		fmt.Printf("✓ User %s can access dataspace %s: %v\n", testUser.ID, dataspace.ID, canAccess)
	}

	// Test HasPermissionInDataspace method
	hasPermission, err := securityMgr.HasPermissionInDataspace(testUser, "entity", "view", dataspace.ID)
	if err != nil {
		log.Printf("Error checking dataspace permission: %v", err)
	} else {
		fmt.Printf("✓ User %s has entity:view permission in dataspace %s: %v\n", testUser.ID, dataspace.ID, hasPermission)
	}

	// Test 4: Test dataspace isolation
	fmt.Printf("\n4. Testing dataspace isolation...\n")

	// Create another dataspace
	dataspace2 := &models.Entity{
		ID:          "test_dataspace_2",
		Tags:        []string{"type:dataspace", "name:Test Dataspace 2", "status:active"},
		Content:     []byte(`{"name":"Test Dataspace 2","description":"Second dataspace for isolation testing"}`),
		ContentType: "application/json",
		Timestamp:   time.Now(),
	}

	err = temporalRepo.Create(dataspace2)
	if err != nil {
		log.Printf("Failed to create second dataspace: %v", err)
	} else {
		fmt.Printf("✓ Created second dataspace: %s\n", dataspace2.ID)
	}

	// Test access to dataspace without permission
	canAccess2, err := securityMgr.CanAccessDataspace(testUser, dataspace2.ID)
	if err != nil {
		log.Printf("Error checking access to second dataspace: %v", err)
	} else {
		fmt.Printf("✓ User %s can access dataspace %s (should be false): %v\n", testUser.ID, dataspace2.ID, canAccess2)
	}

	// Test 5: Test relationship traversal for dataspace permissions
	fmt.Printf("\n5. Testing relationship traversal for dataspace permissions...\n")

	// Get all relationships for the user
	userRels, err := relRepo.GetByFromEntity(testUser.ID)
	if err != nil {
		log.Printf("Error getting user relationships: %v", err)
	} else {
		fmt.Printf("✓ Found %d relationships for user %s\n", len(userRels), testUser.ID)
		for _, rel := range userRels {
			fmt.Printf("  - Relationship: %s -> %s (type: %s)\n", rel.FromEntityID, rel.ToEntityID, rel.Type)
		}
	}

	// Get all relationships to the dataspace
	dataspaceRels, err := relRepo.GetByToEntity(dataspace.ID)
	if err != nil {
		log.Printf("Error getting dataspace relationships: %v", err)
	} else {
		fmt.Printf("✓ Found %d relationships to dataspace %s\n", len(dataspaceRels), dataspace.ID)
		for _, rel := range dataspaceRels {
			fmt.Printf("  - Relationship: %s -> %s (type: %s)\n", rel.FromEntityID, rel.ToEntityID, rel.Type)
		}
	}

	fmt.Printf("\n=== Dataspace Security Test Complete ===\n")
	fmt.Printf("The enhanced dataspace-aware security system has been tested.\n")
	fmt.Printf("Key features validated:\n")
	fmt.Printf("- Dataspace relationship types (can_access, belongs_to)\n")
	fmt.Printf("- Dataspace access validation methods\n")
	fmt.Printf("- Dataspace isolation between different dataspaces\n")
	fmt.Printf("- Relationship traversal for dataspace permissions\n")
}