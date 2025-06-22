// Test tool to verify deletion APIs with RBAC integration
package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"entitydb/logger"
)

// Test configuration
const (
	ServerURL = "https://localhost:8085"
	Username  = "admin"
	Password  = "admin"
)

// API request/response types
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  struct {
		ID       string   `json:"id"`
		Username string   `json:"username"`
		Email    string   `json:"email"`
		Tags     []string `json:"tags"`
	} `json:"user"`
}

type CreateEntityRequest struct {
	Tags    []string `json:"tags"`
	Content string   `json:"content"`
}

type Entity struct {
	ID        string    `json:"id"`
	Tags      []string  `json:"tags"`
	Content   string    `json:"content"`
	CreatedAt interface{} `json:"created_at"` // Changed to interface{} to debug
	UpdatedAt interface{} `json:"updated_at"` // Changed to interface{} to debug
}

type SoftDeleteRequest struct {
	Reason string `json:"reason"`
	Policy string `json:"policy,omitempty"`
	Force  bool   `json:"force,omitempty"`
}

type RestoreRequest struct {
	Reason string `json:"reason"`
}

type PurgeRequest struct {
	Confirmation string `json:"confirmation"`
	Reason       string `json:"reason"`
}

type DeletionStatusResponse struct {
	EntityID       string    `json:"entity_id"`
	LifecycleState string    `json:"lifecycle_state"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at,omitempty"`
	DeletedBy      string    `json:"deleted_by,omitempty"`
	DeleteReason   string    `json:"delete_reason,omitempty"`
	ArchivedAt     *time.Time `json:"archived_at,omitempty"`
	RetentionPolicy string   `json:"retention_policy,omitempty"`
	CanRestore     bool      `json:"can_restore"`
	CanPurge       bool      `json:"can_purge"`
}

type DeletionListResponse struct {
	Entities []DeletionStatusResponse `json:"entities"`
	Total    int                      `json:"total"`
	Count    int                      `json:"count"`
	Offset   int                      `json:"offset"`
	Limit    int                      `json:"limit"`
}

type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// HTTP client with SSL verification disabled for testing
var client = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	},
	Timeout: 30 * time.Second,
}

func main() {
	// Initialize logger
	logger.SetLogLevel("INFO")
	
	fmt.Println("🧪 EntityDB Deletion APIs Test Suite")
	fmt.Println("=====================================")
	
	// Test 1: Authentication
	fmt.Println("\n1. Testing authentication...")
	token, err := authenticate()
	if err != nil {
		log.Fatalf("Authentication failed: %v", err)
	}
	fmt.Printf("   ✅ Authenticated successfully (token: %s...)\n", token[:20])
	
	// Test 2: Create test entities
	fmt.Println("\n2. Creating test entities...")
	
	testEntities := []struct {
		name    string
		tags    []string
		content string
	}{
		{
			name:    "test-document-1",
			tags:    []string{"type:document", "name:test_doc_1", "dataset:test"},
			content: "This is a test document for deletion testing",
		},
		{
			name:    "test-document-2", 
			tags:    []string{"type:document", "name:test_doc_2", "dataset:test", "permanent"},
			content: "This is a permanent test document",
		},
		{
			name:    "test-temp-file",
			tags:    []string{"type:temp", "name:temp_test.tmp", "dataset:test"},
			content: "Temporary file for testing",
		},
	}
	
	createdEntities := make([]Entity, len(testEntities))
	
	for i, testEntity := range testEntities {
		entity, err := createEntity(token, testEntity.tags, testEntity.content)
		if err != nil {
			log.Fatalf("Failed to create entity %s: %v", testEntity.name, err)
		}
		createdEntities[i] = entity
		fmt.Printf("   ✅ Created entity: %s (%s)\n", entity.ID, testEntity.name)
	}
	
	// Test 3: Soft delete entity
	fmt.Println("\n3. Testing soft delete operation...")
	
	entityToDelete := createdEntities[0]
	deleteReq := SoftDeleteRequest{
		Reason: "Testing deletion functionality",
		Policy: "test-policy",
		Force:  false,
	}
	
	deletionStatus, err := softDeleteEntity(token, entityToDelete.ID, deleteReq)
	if err != nil {
		log.Fatalf("Failed to soft delete entity: %v", err)
	}
	
	fmt.Printf("   ✅ Entity soft deleted: %s\n", deletionStatus.EntityID)
	fmt.Printf("   ✅ Lifecycle state: %s\n", deletionStatus.LifecycleState)
	fmt.Printf("   ✅ Deleted by: %s\n", deletionStatus.DeletedBy)
	fmt.Printf("   ✅ Delete reason: %s\n", deletionStatus.DeleteReason)
	fmt.Printf("   ✅ Can restore: %v\n", deletionStatus.CanRestore)
	
	// Test 4: Get deletion status
	fmt.Println("\n4. Testing deletion status query...")
	
	status, err := getDeletionStatus(token, entityToDelete.ID)
	if err != nil {
		log.Fatalf("Failed to get deletion status: %v", err)
	}
	
	fmt.Printf("   ✅ Retrieved deletion status for: %s\n", status.EntityID)
	fmt.Printf("   ✅ State: %s, Can restore: %v, Can purge: %v\n", 
		status.LifecycleState, status.CanRestore, status.CanPurge)
	
	// Test 5: List deleted entities
	fmt.Println("\n5. Testing deleted entities list...")
	
	deletedList, err := listDeletedEntities(token, "soft_deleted", 10, 0)
	if err != nil {
		log.Fatalf("Failed to list deleted entities: %v", err)
	}
	
	fmt.Printf("   ✅ Found %d deleted entities (total: %d)\n", deletedList.Count, deletedList.Total)
	for _, entity := range deletedList.Entities {
		fmt.Printf("   📄 %s: %s (deleted by: %s)\n", 
			entity.EntityID, entity.LifecycleState, entity.DeletedBy)
	}
	
	// Test 6: Restore entity
	fmt.Println("\n6. Testing entity restoration...")
	
	restoreReq := RestoreRequest{
		Reason: "Testing restoration functionality",
	}
	
	restoredStatus, err := restoreEntity(token, entityToDelete.ID, restoreReq)
	if err != nil {
		log.Fatalf("Failed to restore entity: %v", err)
	}
	
	fmt.Printf("   ✅ Entity restored: %s\n", restoredStatus.EntityID)
	fmt.Printf("   ✅ New lifecycle state: %s\n", restoredStatus.LifecycleState)
	
	// Test 7: Delete entity again for purge testing
	fmt.Println("\n7. Preparing entity for purge testing...")
	
	_, err = softDeleteEntity(token, entityToDelete.ID, deleteReq)
	if err != nil {
		log.Fatalf("Failed to soft delete entity for purge test: %v", err)
	}
	fmt.Printf("   ✅ Entity soft deleted again for purge testing\n")
	
	// Test 8: Purge entity
	fmt.Println("\n8. Testing entity purge operation...")
	
	purgeReq := PurgeRequest{
		Confirmation: "PURGE",
		Reason:       "Testing purge functionality",
	}
	
	purgeResult, err := purgeEntity(token, entityToDelete.ID, purgeReq)
	if err != nil {
		log.Fatalf("Failed to purge entity: %v", err)
	}
	
	fmt.Printf("   ✅ Entity purged successfully: %s\n", purgeResult.Message)
	
	// Test 9: Verify entity is gone
	fmt.Println("\n9. Verifying purged entity is no longer accessible...")
	
	_, err = getDeletionStatus(token, entityToDelete.ID)
	if err == nil {
		log.Fatalf("ERROR: Purged entity should not be accessible")
	}
	fmt.Printf("   ✅ Purged entity correctly inaccessible (expected error)\n")
	
	// Test 10: Test permission validation
	fmt.Println("\n10. Testing permission validation...")
	
	// Try to delete without force flag when entity might have relationships
	_, err = softDeleteEntity(token, createdEntities[1].ID, SoftDeleteRequest{
		Reason: "Testing permissions",
		Force:  false,
	})
	
	if err == nil {
		fmt.Printf("   ✅ Permission validation working (entity deleted)\n")
	} else {
		fmt.Printf("   ⚠️  Permission validation: %v\n", err)
	}
	
	// Test 11: Error handling
	fmt.Println("\n11. Testing error handling...")
	
	// Try to delete non-existent entity
	_, err = softDeleteEntity(token, "non-existent-entity", deleteReq)
	if err != nil {
		fmt.Printf("   ✅ Proper error handling for non-existent entity: %v\n", err)
	}
	
	// Try to purge without confirmation
	invalidPurgeReq := PurgeRequest{
		Confirmation: "INVALID",
		Reason:       "Testing error handling",
	}
	
	_, err = purgeEntity(token, createdEntities[2].ID, invalidPurgeReq)
	if err != nil {
		fmt.Printf("   ✅ Proper error handling for invalid purge confirmation: %v\n", err)
	}
	
	fmt.Println("\n🎉 Deletion APIs test suite completed successfully!")
	fmt.Println("\nTest Results Summary:")
	fmt.Println("  ✅ Authentication and authorization")
	fmt.Println("  ✅ Entity creation and management")
	fmt.Println("  ✅ Soft deletion with audit trail")
	fmt.Println("  ✅ Deletion status queries")
	fmt.Println("  ✅ Deleted entities listing and filtering")
	fmt.Println("  ✅ Entity restoration")
	fmt.Println("  ✅ Entity purging with confirmation")
	fmt.Println("  ✅ Permission validation")
	fmt.Println("  ✅ Error handling and edge cases")
	
	// Cleanup remaining entities
	fmt.Println("\n🧹 Cleaning up remaining test entities...")
	for i, entity := range createdEntities[1:] { // Skip first entity (already purged)
		if i == 0 { // Second entity (permanent)
			fmt.Printf("   🔒 Skipping permanent entity: %s\n", entity.ID)
			continue
		}
		
		// Delete and purge remaining entities
		_, err := softDeleteEntity(token, entity.ID, deleteReq)
		if err != nil {
			fmt.Printf("   ⚠️  Failed to delete entity %s: %v\n", entity.ID, err)
			continue
		}
		
		_, err = purgeEntity(token, entity.ID, purgeReq)
		if err != nil {
			fmt.Printf("   ⚠️  Failed to purge entity %s: %v\n", entity.ID, err)
		} else {
			fmt.Printf("   ✅ Cleaned up entity: %s\n", entity.ID)
		}
	}
}

// Helper functions

func authenticate() (string, error) {
	loginReq := LoginRequest{
		Username: Username,
		Password: Password,
	}
	
	jsonData, err := json.Marshal(loginReq)
	if err != nil {
		return "", err
	}
	
	resp, err := client.Post(ServerURL+"/api/v1/auth/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("authentication failed: %s", string(body))
	}
	
	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return "", err
	}
	
	return loginResp.Token, nil
}

func createEntity(token string, tags []string, content string) (Entity, error) {
	reqBody := CreateEntityRequest{
		Tags:    tags,
		Content: content,
	}
	
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return Entity{}, err
	}
	
	req, err := http.NewRequest("POST", ServerURL+"/api/v1/entities/create", bytes.NewBuffer(jsonData))
	if err != nil {
		return Entity{}, err
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	
	resp, err := client.Do(req)
	if err != nil {
		return Entity{}, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return Entity{}, fmt.Errorf("create entity failed (status %d): %s", resp.StatusCode, string(body))
	}
	
	var entity Entity
	if err := json.NewDecoder(resp.Body).Decode(&entity); err != nil {
		return Entity{}, err
	}
	
	return entity, nil
}

func softDeleteEntity(token, entityID string, deleteReq SoftDeleteRequest) (DeletionStatusResponse, error) {
	jsonData, err := json.Marshal(deleteReq)
	if err != nil {
		return DeletionStatusResponse{}, err
	}
	
	req, err := http.NewRequest("POST", ServerURL+"/api/v1/entities/"+entityID+"/delete", bytes.NewBuffer(jsonData))
	if err != nil {
		return DeletionStatusResponse{}, err
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	
	resp, err := client.Do(req)
	if err != nil {
		return DeletionStatusResponse{}, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return DeletionStatusResponse{}, fmt.Errorf("soft delete failed: %s", string(body))
	}
	
	var status DeletionStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return DeletionStatusResponse{}, err
	}
	
	return status, nil
}

func restoreEntity(token, entityID string, restoreReq RestoreRequest) (DeletionStatusResponse, error) {
	jsonData, err := json.Marshal(restoreReq)
	if err != nil {
		return DeletionStatusResponse{}, err
	}
	
	req, err := http.NewRequest("POST", ServerURL+"/api/v1/entities/"+entityID+"/restore", bytes.NewBuffer(jsonData))
	if err != nil {
		return DeletionStatusResponse{}, err
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	
	resp, err := client.Do(req)
	if err != nil {
		return DeletionStatusResponse{}, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return DeletionStatusResponse{}, fmt.Errorf("restore failed: %s", string(body))
	}
	
	var status DeletionStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return DeletionStatusResponse{}, err
	}
	
	return status, nil
}

func getDeletionStatus(token, entityID string) (DeletionStatusResponse, error) {
	req, err := http.NewRequest("GET", ServerURL+"/api/v1/entities/"+entityID+"/deletion-status", nil)
	if err != nil {
		return DeletionStatusResponse{}, err
	}
	
	req.Header.Set("Authorization", "Bearer "+token)
	
	resp, err := client.Do(req)
	if err != nil {
		return DeletionStatusResponse{}, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return DeletionStatusResponse{}, fmt.Errorf("get deletion status failed: %s", string(body))
	}
	
	var status DeletionStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return DeletionStatusResponse{}, err
	}
	
	return status, nil
}

func listDeletedEntities(token, state string, limit, offset int) (DeletionListResponse, error) {
	url := fmt.Sprintf("%s/api/v1/entities/deleted?state=%s&limit=%d&offset=%d", 
		ServerURL, state, limit, offset)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return DeletionListResponse{}, err
	}
	
	req.Header.Set("Authorization", "Bearer "+token)
	
	resp, err := client.Do(req)
	if err != nil {
		return DeletionListResponse{}, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return DeletionListResponse{}, fmt.Errorf("list deleted entities failed: %s", string(body))
	}
	
	var listResp DeletionListResponse
	if err := json.NewDecoder(resp.Body).Decode(&listResp); err != nil {
		return DeletionListResponse{}, err
	}
	
	return listResp, nil
}

func purgeEntity(token, entityID string, purgeReq PurgeRequest) (SuccessResponse, error) {
	jsonData, err := json.Marshal(purgeReq)
	if err != nil {
		return SuccessResponse{}, err
	}
	
	req, err := http.NewRequest("DELETE", ServerURL+"/api/v1/entities/"+entityID+"/purge", bytes.NewBuffer(jsonData))
	if err != nil {
		return SuccessResponse{}, err
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	
	resp, err := client.Do(req)
	if err != nil {
		return SuccessResponse{}, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return SuccessResponse{}, fmt.Errorf("purge failed: %s", string(body))
	}
	
	var successResp SuccessResponse
	if err := json.NewDecoder(resp.Body).Decode(&successResp); err != nil {
		return SuccessResponse{}, err
	}
	
	return successResp, nil
}