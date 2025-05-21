package main

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	serverURL := "http://localhost:8085"
	sizeKB := 100
	
	if len(os.Args) > 1 {
		fmt.Sscanf(os.Args[1], "%d", &sizeKB)
	}
	
	fmt.Printf("Testing with %dKB file\n", sizeKB)
	
	// Create test data
	testData := createTestData(sizeKB)
	fmt.Printf("Created test data: %d bytes\n", len(testData))
	
	// Create entity
	entityID, err := createEntity(serverURL, testData)
	if err != nil {
		log.Fatalf("Failed to create entity: %v", err)
	}
	fmt.Printf("Created entity: %s\n", entityID)
	
	// Retrieve entity
	retrievedData, err := retrieveEntity(serverURL, entityID)
	if err != nil {
		log.Fatalf("Failed to retrieve entity: %v", err)
	}
	fmt.Printf("Retrieved data: %d bytes\n", len(retrievedData))
	
	// Verify data
	if !verifyData(testData, retrievedData) {
		log.Fatalf("Data verification failed")
	}
	fmt.Println("Data verification successful!")
}

func createTestData(sizeKB int) []byte {
	// Create buffer with markers
	buffer := bytes.NewBuffer(nil)
	buffer.WriteString("START_TEST_DATA")
	
	// Generate random data
	randomData := make([]byte, sizeKB*1024)
	rand.Read(randomData)
	buffer.Write(randomData)
	
	buffer.WriteString("END_TEST_DATA")
	return buffer.Bytes()
}

func createEntity(serverURL string, data []byte) (string, error) {
	fmt.Println("Creating entity...")
	
	// Prepare request
	requestData := map[string]interface{}{
		"tags": []string{
			"type:test",
			"test:single_chunk",
		},
		"content": base64.StdEncoding.EncodeToString(data),
	}
	
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}
	
	// Send request
	resp, err := http.Post(
		fmt.Sprintf("%s/api/v1/test/entities/create", serverURL),
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return "", fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status code: %d - %s", resp.StatusCode, string(body))
	}
	
	// Parse response
	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %v", err)
	}
	
	entityID, ok := response["id"].(string)
	if !ok {
		return "", fmt.Errorf("no entity ID in response")
	}
	
	return entityID, nil
}

func retrieveEntity(serverURL string, entityID string) ([]byte, error) {
	start := time.Now()
	fmt.Println("Retrieving entity...")
	
	// Send request for raw content
	resp, err := http.Get(
		fmt.Sprintf("%s/api/v1/entities/get?id=%s&include_content=true&raw=true", serverURL, entityID),
	)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code: %d - %s", resp.StatusCode, string(body))
	}
	
	// Read content
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}
	
	elapsed := time.Since(start)
	fmt.Printf("Retrieved in %v (%0.2f KB/s)\n", elapsed, float64(len(data))/1024/elapsed.Seconds())
	
	return data, nil
}

func verifyData(original, retrieved []byte) bool {
	if len(original) != len(retrieved) {
		fmt.Printf("Size mismatch: original=%d, retrieved=%d\n", len(original), len(retrieved))
		return false
	}
	
	// Check start marker
	if !bytes.Contains(retrieved, []byte("START_TEST_DATA")) {
		fmt.Println("Start marker not found")
		return false
	}
	
	// Check end marker
	if !bytes.Contains(retrieved, []byte("END_TEST_DATA")) {
		fmt.Println("End marker not found")
		return false
	}
	
	// Full comparison (might be slow for large data)
	if !bytes.Equal(original, retrieved) {
		fmt.Println("Content mismatch")
		return false
	}
	
	return true
}