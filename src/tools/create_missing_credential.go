package main

import (
	"entitydb/logger"
	"entitydb/models"
	"entitydb/storage/binary"
	"log"
	"time"
)

func main() {
	logger.Info("[main] === Create Missing Credential Entity ===")

	// Open repository
	repo, err := binary.NewEntityRepository("var")
	if err != nil {
		log.Fatalf("Failed to open repository: %v", err)
	}
	defer repo.Close()

	// Create the missing credential entity
	credID := "cred_bc098257eb7fa98e885df720bbaa5f9a"
	logger.Info("[main] Creating credential entity: %s", credID)

	// Create credential entity with the correct password hash for "admin"
	cred := &models.Entity{
		ID:        credID,
		Content:   []byte(`$2a$10$Xe.e4tFUKZ4Y8qGnIkYLVOZaajBTc0IvSXW9n8Fj/mLvKwJVkCXcC`), // bcrypt hash of "admin"
		CreatedAt: time.Now().UnixNano(),
		UpdatedAt: time.Now().UnixNano(),
		Tags:      []string{},
	}

	// Add tags
	cred.AddTag("type:credential")
	cred.AddTag("credential:type:password")
	cred.AddTag("credential:algorithm:bcrypt")
	cred.AddTag("created:" + time.Now().Format(time.RFC3339))

	// Create the entity
	if err := repo.Create(cred); err != nil {
		log.Fatalf("Failed to create credential entity: %v", err)
	}

	logger.Info("[main] Created credential entity: %s", credID)

	// Verify it was created
	verify, err := repo.GetByID(credID)
	if err != nil {
		logger.Error("[main] Failed to verify credential: %v", err)
	} else {
		logger.Info("[main] Verified credential entity: ID=%s, Tags=%d, Content=%d bytes",
			verify.ID, len(verify.Tags), len(verify.Content))
		for i, tag := range verify.Tags {
			logger.Info("[main]   Tag[%d]: %s", i, tag)
		}
	}

	logger.Info("[main] === Complete ===")
}