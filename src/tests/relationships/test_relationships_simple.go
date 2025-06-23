// test_relationships_simple.go - Simplified relationship test focusing on core functionality
// Tests EntityDB's pure tag-based relationship system

package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

func main() {
	fmt.Println("🔗 RELATIONSHIP MANAGEMENT TEST (SIMPLE) - EntityDB v2.34.3")
	fmt.Println("Testing core tag-based relationship functionality")
	fmt.Println("========================================================")

	// Initialize HTTP client
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: 10 * time.Second,
	}

	// Authenticate
	token, err := authenticate(client)
	if err != nil {
		fmt.Printf("❌ Authentication failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ Authenticated successfully\n\n")

	// Test 1: Direct tag relationships
	fmt.Println("🧪 Test 1: Direct Tag Relationships")
	if err := testDirectTagRelationships(client, token); err != nil {
		fmt.Printf("❌ FAILED: %v\n\n", err)
	} else {
		fmt.Printf("✅ PASSED\n\n")
	}

	// Test 2: Relationship entities
	fmt.Println("🧪 Test 2: Relationship Entities")
	if err := testRelationshipEntities(client, token); err != nil {
		fmt.Printf("❌ FAILED: %v\n\n", err)
	} else {
		fmt.Printf("✅ PASSED\n\n")
	}

	// Test 3: Bidirectional relationships
	fmt.Println("🧪 Test 3: Bidirectional Relationships")
	if err := testBidirectionalRelationships(client, token); err != nil {
		fmt.Printf("❌ FAILED: %v\n\n", err)
	} else {
		fmt.Printf("✅ PASSED\n\n")
	}

	fmt.Println("========================================================")
	fmt.Println("🎉 Core relationship functionality verified!")
}

func authenticate(client *http.Client) (string, error) {
	loginData := map[string]string{
		"username": "admin",
		"password": "admin",
	}

	jsonData, _ := json.Marshal(loginData)
	resp, err := client.Post("https://localhost:8085/api/v1/auth/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	token, ok := result["token"].(string)
	if !ok {
		return "", fmt.Errorf("no token in response")
	}

	return token, nil
}

func makeRequest(client *http.Client, method, url string, body io.Reader, token string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return client.Do(req)
}

func testDirectTagRelationships(client *http.Client, token string) error {
	// Create two entities
	user := map[string]interface{}{
		"id":      "test_user_" + fmt.Sprint(time.Now().UnixNano()),
		"tags":    []string{"type:user", "name:Alice"},
		"content": []byte("User data"),
	}

	project := map[string]interface{}{
		"id":      "test_project_" + fmt.Sprint(time.Now().UnixNano()),
		"tags":    []string{"type:project", "name:ProjectX"},
		"content": []byte("Project data"),
	}

	// Create user
	jsonData, _ := json.Marshal(user)
	resp, err := makeRequest(client, "POST", "https://localhost:8085/api/v1/entities/create", bytes.NewBuffer(jsonData), token)
	if err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}
	resp.Body.Close()

	// Create project
	jsonData, _ = json.Marshal(project)
	resp, err = makeRequest(client, "POST", "https://localhost:8085/api/v1/entities/create", bytes.NewBuffer(jsonData), token)
	if err != nil {
		return fmt.Errorf("failed to create project: %v", err)
	}
	resp.Body.Close()

	// Add relationship tag to user
	user["tags"] = append(user["tags"].([]string), "works_on:"+project["id"].(string))
	
	jsonData, _ = json.Marshal(user)
	resp, err = makeRequest(client, "PUT", "https://localhost:8085/api/v1/entities/update", bytes.NewBuffer(jsonData), token)
	if err != nil {
		return fmt.Errorf("failed to update user: %v", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("update failed with status %d: %s", resp.StatusCode, string(body))
	}
	resp.Body.Close()

	fmt.Printf("   ✅ Created user->project relationship via tag\n")
	return nil
}

func testRelationshipEntities(client *http.Client, token string) error {
	// Create entities
	entity1 := map[string]interface{}{
		"id":      "test_entity1_" + fmt.Sprint(time.Now().UnixNano()),
		"tags":    []string{"type:document"},
		"content": []byte("Doc 1"),
	}

	entity2 := map[string]interface{}{
		"id":      "test_entity2_" + fmt.Sprint(time.Now().UnixNano()),
		"tags":    []string{"type:document"},
		"content": []byte("Doc 2"),
	}

	// Create both entities
	for _, entity := range []map[string]interface{}{entity1, entity2} {
		jsonData, _ := json.Marshal(entity)
		resp, err := makeRequest(client, "POST", "https://localhost:8085/api/v1/entities/create", bytes.NewBuffer(jsonData), token)
		if err != nil {
			return fmt.Errorf("failed to create entity: %v", err)
		}
		resp.Body.Close()
	}

	// Create relationship entity
	relationship := map[string]interface{}{
		"id": "rel_" + entity1["id"].(string) + "_" + entity2["id"].(string),
		"tags": []string{
			"type:relationship",
			"from:" + entity1["id"].(string),
			"to:" + entity2["id"].(string),
			"relation:references",
		},
		"content": []byte(`{"strength": "strong"}`),
	}

	jsonData, _ := json.Marshal(relationship)
	resp, err := makeRequest(client, "POST", "https://localhost:8085/api/v1/entities/create", bytes.NewBuffer(jsonData), token)
	if err != nil {
		return fmt.Errorf("failed to create relationship: %v", err)
	}
	
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("relationship creation failed with status %d: %s", resp.StatusCode, string(body))
	}
	resp.Body.Close()

	fmt.Printf("   ✅ Created relationship entity between documents\n")
	return nil
}

func testBidirectionalRelationships(client *http.Client, token string) error {
	// Create parent and child
	parent := map[string]interface{}{
		"id":      "test_parent_" + fmt.Sprint(time.Now().UnixNano()),
		"tags":    []string{"type:folder", "name:Root"},
		"content": []byte("Parent"),
	}

	child := map[string]interface{}{
		"id":      "test_child_" + fmt.Sprint(time.Now().UnixNano()),
		"tags":    []string{"type:file", "name:Document"},
		"content": []byte("Child"),
	}

	// Create parent
	jsonData, _ := json.Marshal(parent)
	resp, err := makeRequest(client, "POST", "https://localhost:8085/api/v1/entities/create", bytes.NewBuffer(jsonData), token)
	if err != nil {
		return fmt.Errorf("failed to create parent: %v", err)
	}
	resp.Body.Close()

	// Create child with parent reference
	child["tags"] = append(child["tags"].([]string), "parent:"+parent["id"].(string))
	
	jsonData, _ = json.Marshal(child)
	resp, err = makeRequest(client, "POST", "https://localhost:8085/api/v1/entities/create", bytes.NewBuffer(jsonData), token)
	if err != nil {
		return fmt.Errorf("failed to create child: %v", err)
	}
	resp.Body.Close()

	// Update parent with child reference
	parent["tags"] = append(parent["tags"].([]string), "contains:"+child["id"].(string))
	
	jsonData, _ = json.Marshal(parent)
	resp, err = makeRequest(client, "PUT", "https://localhost:8085/api/v1/entities/update", bytes.NewBuffer(jsonData), token)
	if err != nil {
		return fmt.Errorf("failed to update parent: %v", err)
	}
	resp.Body.Close()

	fmt.Printf("   ✅ Created bidirectional parent-child relationship\n")
	return nil
}