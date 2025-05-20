package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func main() {
	// Login
	loginReq := `{"username": "admin", "password": "admin"}`
	resp, err := http.Post("http://localhost:8085/api/v1/auth/login", "application/json", strings.NewReader(loginReq))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	
	var loginResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&loginResp)
	token := loginResp["token"].(string)
	
	// Create entity
	createReq := `{
		"tags": ["type:test", "performance:check", "feature:timestamps"],
		"content": []
	}`
	
	req, _ := http.NewRequest("POST", "http://localhost:8085/api/v1/entities/create", strings.NewReader(createReq))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	
	var createResp map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&createResp)
	entityID := createResp["id"].(string)
	
	fmt.Printf("Created entity: %s\n", entityID)
	fmt.Printf("Response tags: %v\n", createResp["tags"])
	
	// Get entity without timestamps
	req, _ = http.NewRequest("GET", fmt.Sprintf("http://localhost:8085/api/v1/entities/get?id=%s", entityID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("\n1. Get without timestamps:\n%s\n", string(body))
	
	// Get entity with timestamps
	req, _ = http.NewRequest("GET", fmt.Sprintf("http://localhost:8085/api/v1/entities/get?id=%s&include_timestamps=true", entityID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	
	body, _ = ioutil.ReadAll(resp.Body)
	fmt.Printf("\n2. Get with timestamps:\n%s\n", string(body))
}