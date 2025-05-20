package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// User represents a user in the system
type User struct {
	ID       string
	Username string
	Password string
	Roles    []string
	Token    string
}

// EntityDBServer represents the main server that manages entities
type EntityDBServer struct {
	entities      map[string]map[string]interface{}
	relationships []map[string]interface{}
	users         map[string]*User
	tokens        map[string]*User
	port          int
}

// NewEntityDBServer creates a new server instance
func NewEntityDBServer(port int) *EntityDBServer {
	log.Printf("EntityDB Server: Initializing server on port %d", port)
	
	server := &EntityDBServer{
		entities:      make(map[string]map[string]interface{}),
		relationships: []map[string]interface{}{},
		users:         make(map[string]*User),
		tokens:        make(map[string]*User),
		port:          port,
	}
	
	// Initialize users
	adminUser := &User{
		ID:       "usr_admin",
		Username: "admin",
		Password: "password",
		Roles:    []string{"admin"},
	}
	server.users[adminUser.Username] = adminUser
	server.tokens["tk_admin_1234567890"] = adminUser
	
	osakkaUser := &User{
		ID:       "usr_osakka",
		Username: "osakka",
		Password: "mypassword",
		Roles:    []string{"admin"},
	}
	server.users[osakkaUser.Username] = osakkaUser
	
	regularUser := &User{
		ID:       "usr_regular",
		Username: "regular_user",
		Password: "password123",
		Roles:    []string{"user"},
	}
	server.users[regularUser.Username] = regularUser
	
	readonlyUser := &User{
		ID:       "usr_readonly",
		Username: "readonly_user",
		Password: "password123",
		Roles:    []string{"readonly"},
	}
	server.users[readonlyUser.Username] = readonlyUser
	
	// Create user entities with pure tags (no duplicate fields)
	log.Printf("EntityDB Server: Creating user entities with pure tags")
	
	// Admin user entity
	adminEntity := map[string]interface{}{
		"id": "entity_user_admin",
		"tags": []string{
			"type:user",
			"id:username:admin",
			"rbac:role:admin",
			"rbac:perm:*",
			"status:active",
		},
		"content": []map[string]interface{}{
			{"type": "username", "value": adminUser.Username},
			{"type": "password_hash", "value": adminUser.Password},
			{"type": "display_name", "value": "Administrator"},
		},
		"created_at": time.Now().Format(time.RFC3339),
		"created_by": "system",
	}
	server.entities["entity_user_admin"] = adminEntity
	
	// Osakka user entity
	osakkaEntity := map[string]interface{}{
		"id": "entity_user_osakka",
		"tags": []string{
			"type:user",
			"id:username:osakka",
			"rbac:role:admin",
			"rbac:perm:*",
			"status:active",
		},
		"content": []map[string]interface{}{
			{"type": "username", "value": osakkaUser.Username},
			{"type": "password_hash", "value": osakkaUser.Password},
			{"type": "display_name", "value": "Osakka"},
		},
		"created_at": time.Now().Format(time.RFC3339),
		"created_by": "system",
	}
	server.entities["entity_user_osakka"] = osakkaEntity
	
	// Regular user entity
	regularEntity := map[string]interface{}{
		"id": "entity_user_regular_user",
		"tags": []string{
			"type:user",
			"id:username:regular_user",
			"rbac:role:user",
			"rbac:perm:entity:read",
			"rbac:perm:entity:create",
			"rbac:perm:entity:update",
			"status:active",
		},
		"content": []map[string]interface{}{
			{"type": "username", "value": regularUser.Username},
			{"type": "password_hash", "value": regularUser.Password},
			{"type": "display_name", "value": "Regular User"},
		},
		"created_at": time.Now().Format(time.RFC3339),
		"created_by": "system",
	}
	server.entities["entity_user_regular_user"] = regularEntity
	
	// Read-only user entity  
	readonlyEntity := map[string]interface{}{
		"id": "entity_user_readonly_user",
		"tags": []string{
			"type:user",
			"id:username:readonly_user",
			"rbac:role:readonly",
			"rbac:perm:entity:read",
			"rbac:perm:issue:read",
			"status:active",
		},
		"content": []map[string]interface{}{
			{"type": "username", "value": readonlyUser.Username},
			{"type": "password_hash", "value": readonlyUser.Password},
			{"type": "display_name", "value": "Read-only User"},
		},
		"created_at": time.Now().Format(time.RFC3339),
		"created_by": "system",
	}
	server.entities["entity_user_readonly_user"] = readonlyEntity
	
	log.Printf("EntityDB Server: User entity creation completed")
	
	// Load entities from SQLite database
	log.Printf("EntityDB Server: Loading entities from database")
	server.loadEntitiesFromDatabase()
	log.Printf("EntityDB Server: Loaded %d total entities", len(server.entities))
	
	return server
}

// extractTagValue extracts value from tags for a given key
func extractTagValue(tags []string, key string) string {
	prefix := key + ":"
	for _, tag := range tags {
		if strings.HasPrefix(tag, prefix) {
			return strings.TrimPrefix(tag, prefix)
		}
	}
	return ""
}

// extractContentValue extracts value from content items for a given type
func extractContentValue(content []map[string]interface{}, contentType string) string {
	for _, item := range content {
		if itemType, ok := item["type"].(string); ok && itemType == contentType {
			if value, ok := item["value"].(string); ok {
				return value
			}
		}
	}
	return ""
}

// handleEntityCreate handles entity creation with pure tags
func (s *EntityDBServer) handleEntityCreate(w http.ResponseWriter, r *http.Request) {
	var entity map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&entity); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"message": "Invalid request body",
		})
		return
	}

	// Generate entity ID if not provided
	if _, hasID := entity["id"]; !hasID {
		entity["id"] = fmt.Sprintf("entity_%d", time.Now().UnixNano())
	}

	// Set timestamps
	now := time.Now().Format(time.RFC3339)
	entity["created_at"] = now
	entity["updated_at"] = now

	// Ensure we have tags array
	if _, hasTags := entity["tags"]; !hasTags {
		entity["tags"] = []string{}
	}

	// Ensure we have content array
	if _, hasContent := entity["content"]; !hasContent {
		entity["content"] = []map[string]interface{}{}
	}

	// Store the entity
	entityID := entity["id"].(string)
	s.entities[entityID] = entity

	log.Printf("EntityDB Server: Created entity with ID: %s", entityID)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"message": "Entity created successfully",
		"data": entity,
	})
}

// handleEntityUpdate handles entity updates with pure tags
func (s *EntityDBServer) handleEntityUpdate(w http.ResponseWriter, r *http.Request) {
	var updateRequest map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updateRequest); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"message": "Invalid request body",
		})
		return
	}

	// Get entity ID from request
	entityID, ok := updateRequest["id"].(string)
	if !ok || entityID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"message": "Entity ID is required",
		})
		return
	}

	// Get existing entity
	existingEntity, exists := s.entities[entityID]
	if !exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"message": "Entity not found",
		})
		return
	}

	// Update tags if provided
	if tags, ok := updateRequest["tags"]; ok {
		existingEntity["tags"] = tags
	}

	// Update content if provided
	if content, ok := updateRequest["content"]; ok {
		existingEntity["content"] = content
	}

	// Update timestamp
	existingEntity["updated_at"] = time.Now().Format(time.RFC3339)

	// Store updated entity
	s.entities[entityID] = existingEntity

	log.Printf("EntityDB Server: Updated entity with ID: %s", entityID)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"message": "Entity updated successfully",
		"data": existingEntity,
	})
}

// Start starts the EntityDB Server
func (s *EntityDBServer) Start() error {
	log.Printf("EntityDB Server: Starting with pure tag-based architecture")

	// Handle everything through our main handler
	http.HandleFunc("/", s.HandleRequest)

	// Start server
	addr := fmt.Sprintf("localhost:%d", s.port)
	log.Printf("EntityDB Server: Starting server on %s", addr)

	return http.ListenAndServe(addr, nil)
}

// HandleRequest is the main request router
func (s *EntityDBServer) HandleRequest(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	method := r.Method

	log.Printf("EntityDB Server: %s %s", method, path)

	// Static file serving
	if strings.HasPrefix(path, "/assets/") || strings.HasPrefix(path, "/css/") || 
	   strings.HasPrefix(path, "/js/") || strings.HasSuffix(path, ".html") ||
	   path == "/" {
		s.ServeStaticFile(w, r)
		return
	}

	// API routing
	switch {
	case path == "/api/v1/entities" && method == "POST":
		s.handleEntityCreate(w, r)
	case path == "/api/v1/entities/update" && method == "PUT":
		s.handleEntityUpdate(w, r)
	case path == "/api/v1/entities/list" && method == "GET":
		s.HandleEntityList(w, r)
	case path == "/api/v1/entities/get" && method == "GET":
		s.HandleEntityGet(w, r)
	case path == "/api/v1/auth/login" && method == "POST":
		s.HandleLogin(w, r)
	case path == "/api/v1/auth/status" && method == "GET":
		s.HandleAuthStatus(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 page not found"))
	}
}

// ServeStaticFile serves static files from the htdocs directory
func (s *EntityDBServer) ServeStaticFile(w http.ResponseWriter, r *http.Request) {
	rootDir := "/opt/entitydb/share/htdocs"
	filePath := r.URL.Path

	if filePath == "/" {
		filePath = "/index.html"
	}

	fullPath := rootDir + filePath

	// Security check - prevent directory traversal
	cleanPath := filepath.Clean(fullPath)
	if !strings.HasPrefix(cleanPath, rootDir) {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, fullPath)
}

// HandleEntityList handles entity listing with filters
func (s *EntityDBServer) HandleEntityList(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	entityType := r.URL.Query().Get("type")
	tags := r.URL.Query().Get("tags")
	status := r.URL.Query().Get("status")

	// Filter entities based on parameters
	var filteredEntities []map[string]interface{}
	
	for _, entity := range s.entities {
		// Type filter
		if entityType != "" {
			entityTags := entity["tags"].([]string)
			typeFromTags := extractTagValue(entityTags, "type")
			if typeFromTags != entityType {
				continue
			}
		}

		// Status filter
		if status != "" {
			entityTags := entity["tags"].([]string)
			statusFromTags := extractTagValue(entityTags, "status")
			if statusFromTags != status {
				continue
			}
		}

		// Tags filter
		if tags != "" {
			// TODO: Implement tag filtering
		}

		filteredEntities = append(filteredEntities, entity)
	}

	// Return filtered entities
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"data": filteredEntities,
		"count": len(filteredEntities),
		"message": "Entity API is accessible",
	})
}

// HandleEntityGet retrieves a single entity
func (s *EntityDBServer) HandleEntityGet(w http.ResponseWriter, r *http.Request) {
	entityID := r.URL.Query().Get("id")
	if entityID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"message": "Entity ID is required",
		})
		return
	}

	entity, exists := s.entities[entityID]
	if !exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"message": "Entity not found",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"data": entity,
	})
}

// HandleLogin handles user login
func (s *EntityDBServer) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var loginReq struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"message": "Invalid request body",
		})
		return
	}

	// Check user credentials
	user, exists := s.users[loginReq.Username]
	if !exists || user.Password != loginReq.Password {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"message": "Invalid username or password",
		})
		return
	}

	// Generate token
	token := fmt.Sprintf("tk_%s_%d", user.Username, time.Now().Unix())
	s.tokens[token] = user

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"token": token,
		"user": map[string]interface{}{
			"id": user.ID,
			"username": user.Username,
			"roles": user.Roles,
		},
	})
}

// HandleAuthStatus handles authentication status check
func (s *EntityDBServer) HandleAuthStatus(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"message": "No token provided",
		})
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	user, exists := s.tokens[token]
	if !exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "error",
			"message": "Invalid token",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"authenticated": true,
		"user": map[string]interface{}{
			"id": user.ID,
			"username": user.Username,
			"roles": user.Roles,
		},
	})
}

// loadEntitiesFromDatabase loads entities from the SQLite database
func (s *EntityDBServer) loadEntitiesFromDatabase() {
	// Open the database
	dbPath := "/opt/entitydb/var/db/entitydb.db"
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Printf("EntityDB Server: Failed to open database: %v", err)
		return
	}
	defer db.Close()

	// Check if the entities table exists
	var tableExists bool
	err = db.QueryRow("SELECT EXISTS (SELECT 1 FROM sqlite_master WHERE type='table' AND name='entities')").Scan(&tableExists)
	if err != nil {
		log.Printf("EntityDB Server: Failed to check if entities table exists: %v", err)
		return
	}

	if !tableExists {
		log.Printf("EntityDB Server: Entities table does not exist in the database")
		return
	}

	// Query all entities from the database
	rows, err := db.Query("SELECT id, tags, content FROM entities")
	if err != nil {
		log.Printf("EntityDB Server: Failed to query entities: %v", err)
		return
	}
	defer rows.Close()

	// Parse each entity and add it to the server's entities map
	count := 0
	for rows.Next() {
		var id, tagsJSON, contentJSON string
		if err := rows.Scan(&id, &tagsJSON, &contentJSON); err != nil {
			log.Printf("EntityDB Server: Failed to scan entity row: %v", err)
			continue
		}

		// Parse tags from JSON
		var tags []string
		if err := json.Unmarshal([]byte(tagsJSON), &tags); err != nil {
			log.Printf("EntityDB Server: Failed to parse entity tags: %v", err)
			continue
		}

		// Parse content items from JSON
		var contentItems []map[string]interface{}
		if err := json.Unmarshal([]byte(contentJSON), &contentItems); err != nil {
			log.Printf("EntityDB Server: Failed to parse entity content: %v", err)
			continue
		}

		// Create entity map with pure tags
		entity := map[string]interface{}{
			"id": id,
			"tags": tags,
			"content": contentItems,
			"created_at": time.Now().Format(time.RFC3339), // Default timestamp
			"updated_at": time.Now().Format(time.RFC3339), // Default timestamp
		}

		// Store the entity
		s.entities[id] = entity
		count++
	}

	log.Printf("EntityDB Server: Loaded %d entities from database", count)
}

func main() {
	// Parse command line arguments
	portFlag := flag.Int("port", 8085, "Port to listen on")
	flag.Parse()

	// If port is provided as a positional argument, use that instead
	port := *portFlag
	args := flag.Args()
	if len(args) > 0 {
		if p, err := strconv.Atoi(args[0]); err == nil {
			port = p
		}
	}

	// Create and start server
	server := NewEntityDBServer(port)

	// Set up signal handler for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("EntityDB Server: Server started successfully on port %d", port)
	log.Printf("EntityDB Server: Press Ctrl+C to stop")

	// Wait for signal
	sig := <-sigChan
	log.Printf("EntityDB Server: Received signal: %v", sig)
	
	log.Printf("EntityDB Server: Shutting down gracefully...")
	os.Exit(0)
}