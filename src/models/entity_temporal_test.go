package models_test

import (
	"strconv"
	"strings"
	"testing"
	"time"
	"entitydb/models"
)

func TestEntityTemporalTags(t *testing.T) {
	// Test AddTag adds timestamp automatically
	e := models.NewEntity()
	e.AddTag("type:test")
	
	if len(e.Tags) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(e.Tags))
	}
	
	// Check tag has timestamp with | delimiter
	tag := e.Tags[0]
	if !strings.Contains(tag, "|") {
		t.Errorf("Tag missing timestamp delimiter: %s", tag)
	}
	
	parts := strings.SplitN(tag, "|", 2)
	if len(parts) != 2 {
		t.Errorf("Tag not formatted correctly: %s", tag)
	}
	
	// Verify timestamp is valid nanosecond epoch format
	timestamp := parts[0]
	_, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		t.Errorf("Invalid timestamp format: %s", timestamp)
	}
	
	// Verify tag content
	if parts[1] != "type:test" {
		t.Errorf("Expected tag 'type:test', got '%s'", parts[1])
	}
}

func TestGetTagsWithoutTimestamp(t *testing.T) {
	e := models.NewEntity()
	e.AddTag("type:test")
	e.AddTag("status:active")
	
	tags := e.GetTagsWithoutTimestamp()
	if len(tags) != 2 {
		t.Errorf("Expected 2 tags, got %d", len(tags))
	}
	
	// Should return tags without timestamps
	if tags[0] != "type:test" {
		t.Errorf("Expected 'type:test', got '%s'", tags[0])
	}
	if tags[1] != "status:active" {
		t.Errorf("Expected 'status:active', got '%s'", tags[1])
	}
}

func TestAddTagWithValue(t *testing.T) {
	e := models.NewEntity()
	e.AddTagWithValue("type", "user")
	
	tags := e.GetTagsWithoutTimestamp()
	if len(tags) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(tags))
	}
	
	if tags[0] != "type:user" {
		t.Errorf("Expected 'type:user', got '%s'", tags[0])
	}
}

func TestGetTagValue(t *testing.T) {
	e := models.NewEntity()
	
	// Add multiple values for same key to test temporal behavior
	e.AddTagWithValue("status", "draft")
	time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	e.AddTagWithValue("status", "published")
	
	// Should return the most recent value
	value := e.GetTagValue("status")
	if value != "published" {
		t.Errorf("Expected 'published', got '%s'", value)
	}
}

func TestHasTag(t *testing.T) {
	e := models.NewEntity()
	e.AddTag("type:test")
	e.AddTag("status:active")
	
	if !e.HasTag("type:test") {
		t.Error("Expected to find 'type:test' tag")
	}
	
	if !e.HasTag("status:active") {
		t.Error("Expected to find 'status:active' tag")
	}
	
	if e.HasTag("nonexistent:tag") {
		t.Error("Should not find nonexistent tag")
	}
}

// Test that temporal-only system enforces timestamps
func TestTemporalOnlyEnforcement(t *testing.T) {
	e := models.NewEntity()
	
	// Directly add a tag without timestamp (should not happen in normal use)
	e.Tags = append(e.Tags, "type:test")
	
	// GetTagsWithoutTimestamp should handle it gracefully
	tags := e.GetTagsWithoutTimestamp()
	if len(tags) != 1 {
		t.Errorf("Expected 1 tag, got %d", len(tags))
	}
	
	// Tag without timestamp should be returned as-is
	if tags[0] != "type:test" {
		t.Errorf("Expected 'type:test', got '%s'", tags[0])
	}
}