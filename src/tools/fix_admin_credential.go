package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	
	"entitydb/storage/binary"
	"entitydb/models"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Get data path
	dataPath := os.Getenv("ENTITYDB_DATA_PATH")
	if dataPath == "" {
		dataPath = "/opt/entitydb/var"
	}
	
	// Create repository
	repo, err := binary.NewEntityRepository(dataPath)
	if err != nil {
		log.Fatalf("Failed to create repository: %v", err)
	}
	defer repo.Close()
	
	// Find the most recent admin user
	adminUsers, err := repo.ListByTag("identity:username:admin")
	if err != nil {
		log.Fatalf("Failed to list admin users: %v", err)
	}
	
	if len(adminUsers) == 0 {
		log.Fatalf("No admin users found")
	}
	
	// Find the most recent one
	var mostRecentAdmin *models.Entity
	var mostRecentTime int64
	
	for _, user := range adminUsers {
		// Get timestamp from ID suffix (it's often based on time)
		if mostRecentAdmin == nil {
			mostRecentAdmin = user
		}
		
		// Check if user has "status:active" tag
		hasActive := false
		for _, tag := range user.Tags {
			actualTag := tag
			if parts := strings.SplitN(tag, "|", 2); len(parts) == 2 {
				actualTag = parts[1]
			}
			if actualTag == "status:active" {
				hasActive = true
				break
			}
		}
		
		if hasActive {
			// Get creation time from first tag
			if len(user.Tags) > 0 {
				parts := strings.SplitN(user.Tags[0], "|", 2)
				if len(parts) == 2 {
					if nanos, err := strconv.ParseInt(parts[0], 10, 64); err == nil && nanos > mostRecentTime {
						mostRecentTime = nanos
						mostRecentAdmin = user
					}
				}
			}
		}
	}
	
	fmt.Printf("Selected admin user: %s\n", mostRecentAdmin.ID)
	
	// Check if it already has a credential
	relationships, err := repo.GetRelationshipsBySource(mostRecentAdmin.ID)
	if err != nil {
		log.Printf("Error getting relationships: %v", err)
	} else if len(relationships) > 0 {
		for _, rel := range relationships {
			if relObj, ok := rel.(*models.EntityRelationship); ok {
				if relObj.RelationshipType == "has_credential" {
					fmt.Printf("Admin user already has credential: %s\n", relObj.TargetID)
					return
				}
			}
		}
	}
	
	// Create a credential for admin/admin
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}
	
	// Create credential entity
	credential := &models.Entity{
		ID: models.GenerateUUID(),
		Tags: []string{
			"type:credential",
			"credential:type:password",
			"status:active",
		},
		Content: hashedPassword,
	}
	
	if err := repo.Create(credential); err != nil {
		log.Fatalf("Failed to create credential: %v", err)
	}
	
	fmt.Printf("Created credential: %s\n", credential.ID)
	
	// Create has_credential relationship
	relationship := &models.EntityRelationship{
		ID:               models.GenerateUUID(),
		SourceID:         mostRecentAdmin.ID,
		TargetID:         credential.ID,
		RelationshipType: "has_credential",
		Type:             "has_credential",
		Properties:       map[string]string{},
		CreatedAt:        models.Now(),
		UpdatedAt:        models.Now(),
	}
	
	if err := repo.CreateRelationship(relationship); err != nil {
		log.Fatalf("Failed to create relationship: %v", err)
	}
	
	fmt.Printf("Created has_credential relationship: %s\n", relationship.ID)
	fmt.Printf("\nAdmin user %s now has credential %s\n", mostRecentAdmin.ID, credential.ID)
	
	// Also ensure the admin user has the admin role
	// Check if already has admin role
	hasAdminRole := false
	for _, rel := range relationships {
		if relObj, ok := rel.(*models.EntityRelationship); ok {
			if relObj.RelationshipType == "has_role" && strings.Contains(relObj.TargetID, "admin") {
				hasAdminRole = true
				fmt.Printf("Admin user already has admin role via: %s\n", relObj.ID)
				break
			}
		}
	}
	
	if !hasAdminRole {
		// Find or create admin role
		adminRoles, err := repo.ListByTag("type:role")
		if err != nil {
			log.Printf("Error finding admin role: %v", err)
		} else {
			var adminRole *models.Entity
			for _, role := range adminRoles {
				for _, tag := range role.Tags {
					actualTag := tag
					if parts := strings.SplitN(tag, "|", 2); len(parts) == 2 {
						actualTag = parts[1]
					}
					if actualTag == "identity:name:admin" {
						adminRole = role
						break
					}
				}
				if adminRole != nil {
					break
				}
			}
			
			if adminRole != nil {
				// Create has_role relationship
				roleRel := &models.EntityRelationship{
					ID:               models.GenerateUUID(),
					SourceID:         mostRecentAdmin.ID,
					TargetID:         adminRole.ID,
					RelationshipType: "has_role",
					Type:             "has_role",
					Properties:       map[string]string{},
					CreatedAt:        models.Now(),
					UpdatedAt:        models.Now(),
				}
				
				if err := repo.CreateRelationship(roleRel); err != nil {
					log.Printf("Failed to create role relationship: %v", err)
				} else {
					fmt.Printf("Created has_role relationship: %s\n", roleRel.ID)
					fmt.Printf("Admin user %s now has admin role %s\n", mostRecentAdmin.ID, adminRole.ID)
				}
			} else {
				fmt.Printf("Warning: No admin role found in the system\n")
			}
		}
	}
}