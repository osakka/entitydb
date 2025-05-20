package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

// Simple test utility to verify chunked content retrieval

type Entity struct {
	ID      string   `json:"id"`
	Tags    []string `json:"tags"`
	Content []byte   `json:"content,omitempty"`
}

func main() {
	// Disable SSL verification for testing
	http.DefaultTransport.(*http.Transport).TLSClientConfig.InsecureSkipVerify = true

	// Create test data (5MB to force chunking)
	fmt.Println("Creating test data (5MB)...")
	testData := make([]byte, 5*1024*1024)
	_, err := rand.Read(testData)
	if err != nil {
		fmt.Printf("Error creating test data: %v\n", err)
		os.Exit(1)
	}

	// Create a marker pattern at the beginning, middle, and end to verify integrity
	copy(testData[0:16], []byte("START_TEST_DATA_"))
	copy(testData[len(testData)/2:len(testData)/2+16], []byte("MIDDLE_TEST_DATA"))
	copy(testData[len(testData)-16:], []byte("_END_TEST_DATA_"))

	// Step 1: Create an entity without content first
	fmt.Println("Creating entity...")
	entity := createEntity()
	if entity == nil {
		fmt.Println("Failed to create entity")
		os.Exit(1)
	}

	fmt.Printf("Created entity with ID: %s\n", entity.ID)
	
	// Step 2: Update the entity with the test data
	fmt.Println("Updating entity with large content...")
	if !updateEntityContent(entity.ID, testData) {
		fmt.Println("Failed to update entity with content")
		os.Exit(1)
	}

	// Step 3: Verify the entity got chunked
	fmt.Println("Verifying entity was chunked...")
	entity = getEntity(entity.ID, false)
	if entity == nil {
		fmt.Println("Failed to retrieve entity")
		os.Exit(1)
	}

	isChunked := false
	numChunks := 0
	for _, tag := range entity.Tags {
		if len(tag) > 14 && tag[len(tag)-14:] == "content:chunks:" {
			isChunked = true
			fmt.Sscanf(tag[len(tag)-1:], "%d", &numChunks)
			break
		}
	}

	if !isChunked {
		fmt.Println("Entity was not chunked as expected")
		os.Exit(1)
	}

	fmt.Printf("Entity is chunked with %d chunks\n", numChunks)

	// Step 4: Retrieve the entity with content
	fmt.Println("Retrieving entity with content...")
	entityWithContent := getEntity(entity.ID, true)
	if entityWithContent == nil {
		fmt.Println("Failed to retrieve entity with content")
		os.Exit(1)
	}

	// Step 5: Verify the content was correctly reassembled
	fmt.Printf("Retrieved content size: %d bytes\n", len(entityWithContent.Content))
	
	if len(entityWithContent.Content) != len(testData) {
		fmt.Printf("Content size mismatch: expected %d, got %d\n", 
			len(testData), len(entityWithContent.Content))
		os.Exit(1)
	}

	// Check start, middle and end markers
	startOK := bytes.Equal(entityWithContent.Content[0:16], []byte("START_TEST_DATA_"))
	middleOK := bytes.Equal(
		entityWithContent.Content[len(entityWithContent.Content)/2:len(entityWithContent.Content)/2+16], 
		[]byte("MIDDLE_TEST_DATA"))
	endOK := bytes.Equal(
		entityWithContent.Content[len(entityWithContent.Content)-16:], 
		[]byte("_END_TEST_DATA_"))

	if !startOK || !middleOK || !endOK {
		fmt.Println("Content integrity check failed:")
		fmt.Printf("  Start marker: %v\n", startOK)
		fmt.Printf("  Middle marker: %v\n", middleOK)
		fmt.Printf("  End marker: %v\n", endOK)
		os.Exit(1)
	}

	fmt.Println("SUCCESS: Chunked content was correctly stored and retrieved!")
}

func createEntity() *Entity {
	// Create a simple entity
	reqBody := map[string]interface{}{
		"tags": []string{"type:test", "test:chunking"},
	}

	jsonData, _ := json.Marshal(reqBody)
	resp, err := http.Post(
		"https://localhost:8085/api/v1/test/entity/create",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	
	if err != nil {
		fmt.Printf("Error creating entity: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		fmt.Printf("Error creating entity: HTTP %d\n", resp.StatusCode)
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("Response: %s\n", string(body))
		return nil
	}

	var entity Entity
	if err := json.NewDecoder(resp.Body).Decode(&entity); err != nil {
		fmt.Printf("Error decoding response: %v\n", err)
		return nil
	}

	return &entity
}

func updateEntityContent(id string, content []byte) bool {
	// Update the entity with content
	reqBody := map[string]interface{}{
		"id": id,
		"content": base64.StdEncoding.EncodeToString(content),
	}

	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(
		"PUT", 
		fmt.Sprintf("https://localhost:8085/api/v1/entities/update?id=%s", id),
		bytes.NewBuffer(jsonData),
	)
	
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	
	if err != nil {
		fmt.Printf("Error updating entity: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error updating entity: HTTP %d\n", resp.StatusCode)
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("Response: %s\n", string(body))
		return false
	}

	return true
}

func getEntity(id string, includeContent bool) *Entity {
	// Get entity with or without content
	url := fmt.Sprintf("https://localhost:8085/api/v1/entities/get?id=%s", id)
	if includeContent {
		url += "&include_content=true"
	}

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	
	if err != nil {
		fmt.Printf("Error getting entity: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Error getting entity: HTTP %d\n", resp.StatusCode)
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Printf("Response: %s\n", string(body))
		return nil
	}

	var entity Entity
	if err := json.NewDecoder(resp.Body).Decode(&entity); err != nil {
		fmt.Printf("Error decoding response: %v\n", err)
		return nil
	}

	return &entity
}