package main

// Types shared between security components

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// MockServer represents a simple mock of EntityDBServer for testing
type MockServer struct {
	Entities     map[string]map[string]interface{}
	Relationships map[string]map[string]interface{}
}

// MockUser represents a simple user for testing
type MockUser struct {
	ID       string
	Username string
	Roles    []string
}

// validateToken is a mock implementation for testing
func (s *MockServer) validateToken(token string) *MockUser {
	// Always return a test user for any token
	if token != "" {
		return &MockUser{
			ID:       "usr_test",
			Username: "testuser",
			Roles:    []string{"admin"},
		}
	}
	return nil
}