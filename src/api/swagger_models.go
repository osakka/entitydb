package api

import (
	"entitydb/models"
	"time"
)

// LoginRequest represents authentication credentials
// @Description Login credentials
type LoginRequest struct {
	Username string `json:"username" example:"admin"`
	Password string `json:"password" example:"password123"`
}

// LoginResponse represents successful authentication
// @Description Authentication response with session token
type LoginResponse struct {
	Token     string    `json:"token" example:"token_abc123..."`
	ExpiresAt time.Time `json:"expires_at" example:"2025-01-01T12:00:00Z"`
	User      UserInfo  `json:"user"`
}

// UserInfo represents user information
// @Description User details
type UserInfo struct {
	ID       string   `json:"id" example:"entity_user_admin"`
	Username string   `json:"username" example:"admin"`
	Roles    []string `json:"roles" example:"admin,user"`
}

// AuthStatusResponse represents authentication status
// @Description Current authentication status
type AuthStatusResponse struct {
	Authenticated bool      `json:"authenticated" example:"true"`
	ExpiresAt     time.Time `json:"expires_at" example:"2025-01-01T12:00:00Z"`
	User          UserInfo  `json:"user"`
}

// RefreshResponse represents token refresh response
// @Description Refreshed session information
type RefreshResponse struct {
	Token     string    `json:"token" example:"token_xyz789..."`
	ExpiresAt time.Time `json:"expires_at" example:"2025-01-01T14:00:00Z"`
}

// StatusResponse represents a simple status response
// @Description Status message
type StatusResponse struct {
	Status string `json:"status" example:"ok"`
}

// ErrorResponse represents an error response
// @Description Error information
type ErrorResponse struct {
	Error string `json:"error" example:"Invalid request"`
}

// EntityRequest represents entity creation/update request
// @Description Entity data for creation or update
type EntityRequest struct {
	ID      string   `json:"id,omitempty" example:"entity_123"`
	Type    string   `json:"type" example:"user"`
	Title   string   `json:"title,omitempty" example:"Test Entity"`
	Tags    []string `json:"tags" example:"status:active,priority:high"`
	Content []ContentItem `json:"content,omitempty"`
}

// ContentItem represents entity content
// @Description Content item within an entity
type ContentItem struct {
	Type      string    `json:"type" example:"description"`
	Value     string    `json:"value" example:"Entity description"`
	Timestamp time.Time `json:"timestamp,omitempty" example:"2025-01-01T10:00:00Z"`
}

// EntityResponse represents a complete entity
// @Description Complete entity with all data
type EntityResponse struct {
	ID        string        `json:"id" example:"entity_123"`
	Tags      []string      `json:"tags" example:"type:user,status:active"`
	Content   []ContentItem `json:"content"`
	CreatedAt time.Time     `json:"created_at" example:"2025-01-01T10:00:00Z"`
	UpdatedAt time.Time     `json:"updated_at" example:"2025-01-01T11:00:00Z"`
}

// EntityListResponse represents a list of entities
// @Description List of entities
type EntityListResponse struct {
	Entities []EntityResponse `json:"entities"`
	Count    int              `json:"count" example:"10"`
}

// RelationshipRequest represents relationship creation request
// @Description Relationship creation data
type RelationshipRequest struct {
	SourceID         string `json:"source_id" example:"entity_123"`
	RelationshipType string `json:"relationship_type" example:"contains"`
	TargetID         string `json:"target_id" example:"entity_456"`
}

// RelationshipResponse represents a relationship
// @Description Entity relationship
type RelationshipResponse struct {
	ID               string    `json:"id" example:"rel_789"`
	SourceID         string    `json:"source_id" example:"entity_123"`
	RelationshipType string    `json:"relationship_type" example:"contains"`
	TargetID         string    `json:"target_id" example:"entity_456"`
	CreatedAt        time.Time `json:"created_at" example:"2025-01-01T10:00:00Z"`
}

// RelationshipListResponse represents a list of relationships
// @Description List of entity relationships
type RelationshipListResponse struct {
	Relationships []RelationshipResponse `json:"relationships"`
	Count         int                    `json:"count" example:"5"`
}

// TimeRangeQuery represents a time range query
// @Description Time range for temporal queries
type TimeRangeQuery struct {
	From time.Time `json:"from" example:"2025-01-01T00:00:00Z"`
	To   time.Time `json:"to" example:"2025-01-31T23:59:59Z"`
}

// EntityChange represents a change to an entity
// @Description Change event for an entity
type EntityChange struct {
	EntityID  string    `json:"entity_id" example:"entity_123"`
	Timestamp time.Time `json:"timestamp" example:"2025-01-01T10:00:00Z"`
	Type      string    `json:"type" example:"modified"`
	Tag       string    `json:"tag" example:"status:active"`
	OldValue  string    `json:"old_value,omitempty" example:"draft"`
	NewValue  string    `json:"new_value,omitempty" example:"active"`
}

// DashboardStats represents dashboard statistics
// @Description System statistics for dashboard
type DashboardStats struct {
	TotalEntities int            `json:"total_entities" example:"1000"`
	EntityTypes   map[string]int `json:"entity_types" example:"user:10,task:50"`
	ActiveUsers   int            `json:"active_users" example:"5"`
	LastActivity  time.Time      `json:"last_activity" example:"2025-01-01T12:00:00Z"`
}

// ConfigResponse represents configuration data
// @Description System configuration
type ConfigResponse struct {
	Config map[string]interface{} `json:"config"`
}

// FeatureFlagRequest represents feature flag update
// @Description Feature flag update request
type FeatureFlagRequest struct {
	Flag    string `json:"flag" example:"new_ui"`
	Enabled bool   `json:"enabled" example:"true"`
}

// APIStatusResponse represents API status
// @Description API status information
type APIStatusResponse struct {
	Status  string    `json:"status" example:"ok"`
	Version string    `json:"version" example:"2.6.0"`
	Time    time.Time `json:"time" example:"2025-01-01T12:00:00Z"`
}

// SuccessResponse represents a generic success response
// @Description Generic success response
type SuccessResponse struct {
	Success bool   `json:"success" example:"true"`
	Message string `json:"message,omitempty" example:"Operation completed successfully"`
}

// ConfigSetRequest represents a configuration update request
type ConfigSetRequest struct {
	Key       string      `json:"key"`
	Value     interface{} `json:"value"`
	Namespace string      `json:"namespace,omitempty"`
}

// QueryEntityResponse represents a response from the advanced query endpoint
type QueryEntityResponse struct {
	Entities []*models.Entity `json:"entities"`
	Total    int              `json:"total"`
	Offset   int              `json:"offset"`  
	Limit    int              `json:"limit"`
}