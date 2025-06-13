//go:build tool
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

// TestResult represents a test result
type TestResult struct {
	Name        string `json:"name"`
	Success     bool   `json:"success"`
	Description string `json:"description"`
	Error       string `json:"error,omitempty"`
}

// Entity represents a simplified entity model
type Entity struct {
	ID      string   `json:"id,omitempty"`
	Tags    []string `json:"tags,omitempty"`
	Content string   `json:"content,omitempty"`
}

// EntityResponse represents a response with entity data
type EntityResponse struct {
	ID      string   `json:"id"`
	Tags    []string `json:"tags"`
	Content []byte   `json:"content"`
}

func main() {
	serverURL := flag.String("server", "http://localhost:8085", "EntityDB server URL")
	sizeKB := flag.Int("size", 500, "Test file size in KB")
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	fmt.Println("=== EntityDB Chunk API Test ===")
	fmt.Printf("Testing with server: %s\n", *serverURL)
	fmt.Printf("Test file size: %d KB\n", *sizeKB)

	// Run the tests
	results := runTests(*serverURL, *sizeKB)

	// Print results
	fmt.Println("\n=== Test Results ===")
	success := true
	for _, result := range results {
		if result.Success {
			fmt.Printf("✅ %s: %s\n", result.Name, result.Description)
		} else {
			fmt.Printf("❌ %s: %s - %s\n", result.Name, result.Description, result.Error)
			success = false
		}
	}

	if success {
		fmt.Println("\n✅ All tests passed successfully!")
		os.Exit(0)
	} else {
		fmt.Println("\n❌ Some tests failed!")
		os.Exit(1)
	}
}

func runTests(serverURL string, sizeKB int) []TestResult {
	results := []TestResult{}

	// Test 1: Generate test data
	testData, result := generateTestData(sizeKB)
	results = append(results, result)
	if !result.Success {
		return results
	}

	// Test 2: Create entity
	entityID, result := createEntity(serverURL, testData)
	results = append(results, result)
	if !result.Success {
		return results
	}

	// Test 3: Retrieve entity with content
	retrievedData, result := retrieveEntity(serverURL, entityID)
	results = append(results, result)
	if !result.Success {
		return results
	}

	// Test 4: Verify content integrity
	result = verifyContent(testData, retrievedData)
	results = append(results, result)

	return results
}

// Generate test data with markers
func generateTestData(sizeKB int) ([]byte, TestResult) {
	result := TestResult{
		Name:        "GenerateTestData",
		Description: fmt.Sprintf("Generate %d KB of test data", sizeKB),
	}

	// Generate random data of specified size
	data := make([]byte, sizeKB*1024)
	n, err := rand.Read(data)
	if err != nil || n != sizeKB*1024 {
		result.Success = false
		result.Error = fmt.Sprintf("Failed to generate random data: %v", err)
		return nil, result
	}

	// Add start and end markers for verification
	startMarker := []byte("START_TEST_DATA")
	endMarker := []byte("END_TEST_DATA")
	testData := bytes.NewBuffer(nil)
	testData.Write(startMarker)
	testData.Write(data)
	testData.Write(endMarker)

	result.Success = true
	return testData.Bytes(), result
}

// Create entity with test data
func createEntity(serverURL string, testData []byte) (string, TestResult) {
	result := TestResult{
		Name:        "CreateEntity",
		Description: "Create entity with test data",
	}

	// Calculate checksum for verification
	hasher := sha256.New()
	hasher.Write(testData)
	checksum := hex.EncodeToString(hasher.Sum(nil))

	// Create entity with test data
	entity := Entity{
		Tags: []string{
			"type:test_chunking",
			"test:api_test",
			fmt.Sprintf("test:size:%d", len(testData)),
			fmt.Sprintf("test:checksum:%s", checksum),
		},
		Content: base64.StdEncoding.EncodeToString(testData),
	}

	entityJSON, err := json.Marshal(entity)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("Failed to marshal entity: %v", err)
		return "", result
	}

	// Create entity via test endpoint
	resp, err := http.Post(
		fmt.Sprintf("%s/api/v1/test/entities/create", serverURL),
		"application/json",
		bytes.NewBuffer(entityJSON),
	)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("Failed to create entity: %v", err)
		return "", result
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		result.Success = false
		result.Error = fmt.Sprintf("Failed to create entity: HTTP %d - %s", resp.StatusCode, string(bodyBytes))
		return "", result
	}

	var response EntityResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("Failed to decode response: %v", err)
		return "", result
	}

	result.Success = true
	log.Printf("Created entity with ID: %s", response.ID)
	return response.ID, result
}

// Retrieve entity with content
func retrieveEntity(serverURL string, entityID string) ([]byte, TestResult) {
	result := TestResult{
		Name:        "RetrieveEntity",
		Description: "Retrieve entity with content",
	}

	// Get entity with content
	resp, err := http.Get(
		fmt.Sprintf("%s/api/v1/entities/get?id=%s&include_content=true", serverURL, entityID),
	)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("Failed to retrieve entity: %v", err)
		return nil, result
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		result.Success = false
		result.Error = fmt.Sprintf("Failed to retrieve entity: HTTP %d - %s", resp.StatusCode, string(bodyBytes))
		return nil, result
	}

	var response map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("Failed to decode response: %v", err)
		return nil, result
	}

	// Extract content from response
	content, ok := response["content"].(string)
	if !ok {
		result.Success = false
		result.Error = "Content field missing or not a string"
		return nil, result
	}

	// Decode base64 content
	decodedContent, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("Failed to decode content: %v", err)
		return nil, result
	}

	result.Success = true
	log.Printf("Retrieved %d bytes of content", len(decodedContent))
	return decodedContent, result
}

// Verify content integrity
func verifyContent(original, retrieved []byte) TestResult {
	result := TestResult{
		Name:        "VerifyContent",
		Description: "Verify content integrity",
	}

	// Check size
	if len(original) != len(retrieved) {
		result.Success = false
		result.Error = fmt.Sprintf("Size mismatch: original=%d, retrieved=%d", len(original), len(retrieved))
		return result
	}

	// Check start marker
	startMarker := []byte("START_TEST_DATA")
	if !bytes.Contains(retrieved, startMarker) {
		result.Success = false
		result.Error = "Start marker not found in retrieved content"
		return result
	}

	// Check end marker
	endMarker := []byte("END_TEST_DATA")
	if !bytes.Contains(retrieved, endMarker) {
		result.Success = false
		result.Error = "End marker not found in retrieved content"
		return result
	}

	// Check content checksum
	originalChecksum := sha256.Sum256(original)
	retrievedChecksum := sha256.Sum256(retrieved)
	if originalChecksum != retrievedChecksum {
		result.Success = false
		result.Error = fmt.Sprintf("Checksum mismatch: original=%x, retrieved=%x",
			originalChecksum, retrievedChecksum)
		return result
	}

	result.Success = true
	return result
}