package main

import (
    "encoding/json"
    "fmt"
    "entitydb/models"
    "entitydb/storage/binary"
    "log"
)

func main() {
    // Open repository
    repo, err := binary.NewEntityRepository("/opt/entitydb/var")
    if err != nil {
        log.Fatal("Failed to open repository:", err)
    }
    defer repo.Close()
    
    // Find admin user
    users, err := repo.ListByTag("type:user")
    if err != nil {
        log.Fatal("Failed to list users:", err)
    }
    
    fmt.Printf("Found %d users\n", len(users))
    
    var adminUser *models.Entity
    for _, user := range users {
        // Check tags for admin username
        tags := user.GetTagsWithoutTimestamp()
        fmt.Printf("User %s tags: %v\n", user.ID, tags)
        for _, tag := range tags {
            if tag == "identity:username:admin" {
                adminUser = user
                break
            }
        }
        if adminUser != nil {
            break
        }
    }
    
    if adminUser == nil {
        log.Fatal("Admin user not found")
    }
    
    fmt.Printf("Found admin user: %s\n", adminUser.ID)
    fmt.Printf("Current tags: %v\n", adminUser.Tags)
    
    // Add rbac:role:admin tag if not present
    hasAdminRole := false
    cleanTags := adminUser.GetTagsWithoutTimestamp()
    for _, tag := range cleanTags {
        if tag == "rbac:role:admin" {
            hasAdminRole = true
            break
        }
    }
    
    if !hasAdminRole {
        adminUser.Tags = append(adminUser.Tags, "rbac:role:admin")
        if err := repo.Update(adminUser); err != nil {
            log.Fatal("Failed to update admin user:", err)
        }
        fmt.Println("Added rbac:role:admin tag to admin user")
    } else {
        fmt.Println("Admin user already has rbac:role:admin tag")
    }
}