package main

import "time"

// Common constants for all e2e tests
const (
	BaseURL      = "https://localhost:8085"
	TestUsername = "admin"
	TestPassword = "admin"
	TestTimeout  = 30 * time.Second
)

// Common response structures shared across e2e tests
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	UserID    string    `json:"user_id"`
	User      struct {
		ID       string   `json:"id"`
		Username string   `json:"username"`
		Email    string   `json:"email"`
		Roles    []string `json:"roles"`
	} `json:"user"`
}

type Entity struct {
	ID        string   `json:"id"`
	Tags      []string `json:"tags"`
	Content   []byte   `json:"content"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

type TestResult struct {
	TestName string
	Success  bool
	Duration time.Duration
	Error    string
	Details  map[string]interface{}
}