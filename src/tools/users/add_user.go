//go:build tool
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"golang.org/x/crypto/bcrypt"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if len(os.Args) < 5 {
		fmt.Println("Usage: go run add_user.go <username> <password> <email> <full_name>")
		os.Exit(1)
	}

	username := os.Args[1]
	password := os.Args[2]
	email := os.Args[3]
	fullName := os.Args[4]
	userId := "user_" + username

	// Connect to database
	db, err := sql.Open("sqlite3", "/opt/entitydb/var/db/entitydb.db")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	
	// Check if user exists
	var userExists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)", username).Scan(&userExists)
	if err != nil {
		log.Fatalf("Error checking for user: %v", err)
	}
	
	if userExists {
		fmt.Printf("User %s already exists, deleting it...\n", username)
		_, err = db.Exec("DELETE FROM users WHERE username = ?", username)
		if err != nil {
			log.Fatalf("Failed to delete existing user: %v", err)
		}
	}
	
	// Generate a password hash
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to generate password hash: %v", err)
	}
	
	fmt.Printf("Generated hash for '%s': %s\n", password, string(hash))
	
	// Insert user
	_, err = db.Exec(`
		INSERT INTO users (
			id, username, password_hash, email, display_name, full_name, roles, active, status
		) VALUES (
			?, ?, ?, ?, ?, ?, 'user', 1, 'active'
		)
	`, userId, username, string(hash), email, username, fullName)
	
	if err != nil {
		log.Fatalf("Failed to insert user: %v", err)
	}
	
	fmt.Printf("User %s created successfully with id %s\n", username, userId)
}