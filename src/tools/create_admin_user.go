package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"entitydb/models"
	"entitydb/storage/binary"
	"fmt"
	"log"
	"time"
	
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Open repository
	repo, err := binary.NewEntityRepository("../var")
	if err != nil {
		log.Fatalf("Failed to open repository: %v", err)
	}
	defer repo.Close()
	
	// Create admin user
	username := "admin"
	password := "admin"
	
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}
	
	// Create user content
	userData := map[string]string{
		"username":      username,
		"password_hash": string(hashedPassword),
		"display_name":  "Administrator",
	}
	
	userJSON, err := json.Marshal(userData)
	if err != nil {
		log.Fatalf("Failed to marshal user data: %v", err)
	}
	
	// Create user entity
	userEntity := &models.Entity{
		ID: "user_admin",
		Tags: []string{
			"type:user",
			"id:username:admin",
			"rbac:role:admin",
			"rbac:perm:*",
			"status:active",
		},
		Content: userJSON,
	}
	
	// Create credential
	saltBytes := make([]byte, 16)
	rand.Read(saltBytes)
	salt := hex.EncodeToString(saltBytes)
	
	credID := fmt.Sprintf("cred_%s", hex.EncodeToString([]byte(username)))
	credEntity := &models.Entity{
		ID: credID,
		Tags: []string{
			"type:credential",
			"algorithm:bcrypt",
			"user:user_admin",
			fmt.Sprintf("salt:%s", salt),
			fmt.Sprintf("created:%d", time.Now().UnixNano()),
		},
		Content: hashedPassword,
	}
	
	// Create entities
	if err := repo.Create(userEntity); err != nil {
		log.Fatalf("Failed to create user: %v", err)
	}
	
	if err := repo.Create(credEntity); err != nil {
		log.Fatalf("Failed to create credential: %v", err)
	}
	
	// Create relationship
	relEntity := &models.Entity{
		ID: fmt.Sprintf("rel_user_admin_has_cred_%s", credID),
		Tags: []string{
			"_relationship",
			"_source:user_admin",
			fmt.Sprintf("_target:%s", credID),
			"content:type:relationship",
		},
		Content: []byte(fmt.Sprintf(`{"relationship_type":"has_credential","source_id":"user_admin","target_id":"%s"}`, credID)),
	}
	
	if err := repo.Create(relEntity); err != nil {
		log.Fatalf("Failed to create relationship: %v", err)
	}
	
	fmt.Printf("Admin user created successfully:\n")
	fmt.Printf("Username: %s\n", username)
	fmt.Printf("Password: %s\n", password)
	fmt.Printf("User Entity: %s\n", base64.StdEncoding.EncodeToString(userJSON))
}