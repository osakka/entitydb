package main

import (
	"flag"
	"fmt"
	"log"
	"entitydb/models"
	"entitydb/storage/binary"
)

func main() {
	var storagePath string
	var dryRun bool
	
	flag.StringVar(&storagePath, "storage", "/opt/entitydb/var", "Path to EntityDB storage")
	flag.BoolVar(&dryRun, "dry-run", false, "Show what would be done without making changes")
	flag.Parse()
	
	// Logger is already configured
	
	fmt.Printf("EntityDB RBAC Tag Migration Tool\n")
	fmt.Printf("Storage path: %s\n", storagePath)
	fmt.Printf("Dry run: %v\n\n", dryRun)
	
	// Open repository
	repo, err := binary.NewEntityRepository(storagePath)
	if err != nil {
		log.Fatalf("Failed to open repository: %v", err)
	}
	defer repo.Close()
	
	// Create RBAC tag manager
	rbacManager := models.NewRBACTagManager(repo)
	
	// Find all users
	users, err := repo.ListByTag("type:user")
	if err != nil {
		log.Fatalf("Failed to list users: %v", err)
	}
	
	fmt.Printf("Found %d users\n\n", len(users))
	
	// Process each user
	for _, user := range users {
		fmt.Printf("Processing user %s:\n", user.ID)
		
		// Get user info
		cleanTags := user.GetTagsWithoutTimestamp()
		var username string
		var hasAdminRole bool
		
		for _, tag := range cleanTags {
			if tag == "identity:username:admin" {
				username = "admin"
			} else if tag == "rbac:role:admin" {
				hasAdminRole = true
			}
		}
		
		fmt.Printf("  Tags: %v\n", cleanTags)
		
		// Check if this is the admin user
		if username == "admin" && !hasAdminRole {
			fmt.Printf("  -> Admin user missing rbac:role:admin tag\n")
			
			if !dryRun {
				if err := rbacManager.AssignRoleToUser(user.ID, "admin"); err != nil {
					fmt.Printf("  ERROR: Failed to assign admin role: %v\n", err)
				} else {
					fmt.Printf("  SUCCESS: Added rbac:role:admin tag\n")
				}
			} else {
				fmt.Printf("  Would add rbac:role:admin tag\n")
			}
		} else if username == "admin" && hasAdminRole {
			fmt.Printf("  -> Admin user already has rbac:role:admin tag\n")
		}
		
		// Check relationships and sync tags if needed
		rels, err := repo.GetRelationshipsBySource(user.ID)
		if err == nil {
			for _, rel := range rels {
				if relationship, ok := rel.(*models.EntityRelationship); ok {
					if relationship.Type == "has_role" && relationship.TargetID == "role_admin" && !hasAdminRole {
						fmt.Printf("  -> User has admin role relationship but missing tag\n")
						
						if !dryRun {
							if err := rbacManager.AssignRoleToUser(user.ID, "admin"); err != nil {
								fmt.Printf("  ERROR: Failed to assign admin role: %v\n", err)
							} else {
								fmt.Printf("  SUCCESS: Added rbac:role:admin tag\n")
							}
						} else {
							fmt.Printf("  Would add rbac:role:admin tag\n")
						}
					}
				}
			}
		}
		
		fmt.Println()
	}
	
	fmt.Println("Migration complete!")
	
	if dryRun {
		fmt.Println("\nThis was a dry run. Run without --dry-run to apply changes.")
	}
}