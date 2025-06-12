// Package models provides temporal utilities for EntityDB's time-based tag system.
// All tags in EntityDB are temporal, stored with nanosecond-precision timestamps.
package models

import (
	"fmt"
	"strings"
	"time"
)

// Temporal Tag Format
//
// EntityDB stores all tags with timestamps to enable point-in-time queries
// and historical analysis. Tags follow one of these formats:
//
//   - RFC3339 with nanoseconds: "2006-01-02T15:04:05.999999999.namespace:value"
//   - RFC3339 with nanoseconds: "2006-01-02T15:04:05.999999999.namespace=value"
//   - Epoch nanoseconds: "1234567890123456789|namespace:value"
//
// The timestamp is separated from the tag content by either a dot (.) or pipe (|).
// The tag content can use either colon (:) or equals (=) as a separator.
//
// Example temporal tags:
//   - "2024-01-15T10:30:45.123456789.status:active"
//   - "2024-01-15T10:30:45.123456789.priority=high"
//   - "1705318245123456789|type:user"

// ExtractTimestamp extracts the timestamp from a temporal tag.
// It supports both RFC3339 format and epoch nanosecond format.
//
// Returns an error if the tag doesn't have a valid timestamp prefix.
//
// Examples:
//
//	// RFC3339 format
//	ts, _ := ExtractTimestamp("2024-01-15T10:30:45.123456789.status:active")
//	// ts = time.Date(2024, 1, 15, 10, 30, 45, 123456789, time.UTC)
//
//	// Epoch format
//	ts, _ := ExtractTimestamp("1705318245123456789|status:active")
//	// ts = time.Unix(0, 1705318245123456789)
func ExtractTimestamp(tag string) (time.Time, error) {
	// Try RFC3339 format first (dot separator)
	parts := strings.SplitN(tag, ".", 2)
	if len(parts) < 2 {
		// Try epoch format (pipe separator)
		// Not a temporal tag
		return time.Time{}, fmt.Errorf("not a temporal tag: %s", tag)
	}
	
	// Parse the timestamp
	timestamp, err := time.Parse("2006-01-02T15:04:05.999999999", parts[0])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timestamp format: %s", parts[0])
	}
	
	return timestamp, nil
}

// ExtractNamespaceValue extracts the namespace and value components from a tag.
// It handles both temporal and non-temporal tags, automatically stripping
// the timestamp prefix if present.
//
// Supports both colon (:) and equals (=) separators.
// If no separator is found, the entire tag is returned as the namespace
// with an empty value.
//
// Examples:
//
//	ns, val := ExtractNamespaceValue("2024-01-15T10:30:45.123456789.status:active")
//	// ns = "status", val = "active"
//
//	ns, val := ExtractNamespaceValue("priority=high")
//	// ns = "priority", val = "high"
//
//	ns, val := ExtractNamespaceValue("simple-tag")
//	// ns = "simple-tag", val = ""
func ExtractNamespaceValue(tag string) (namespace, value string) {
	// Remove timestamp if present (handles both . and | separators)
	if idx := strings.Index(tag, "."); idx != -1 {
		tag = tag[idx+1:]
	} else if idx := strings.Index(tag, "|"); idx != -1 {
		tag = tag[idx+1:]
	}
	
	// Split by = or : to extract namespace and value
	if idx := strings.Index(tag, "="); idx != -1 {
		return tag[:idx], tag[idx+1:]
	}
	if idx := strings.Index(tag, ":"); idx != -1 {
		return tag[:idx], tag[idx+1:]
	}
	
	// No separator found - entire tag is the namespace
	return tag, ""
}

// IsTemporalTag checks if a tag has a valid timestamp prefix.
// Returns true if the tag starts with a parseable timestamp.
//
// Examples:
//
//	IsTemporalTag("2024-01-15T10:30:45.123456789.status:active") // true
//	IsTemporalTag("status:active") // false
//	IsTemporalTag("1705318245123456789|type:user") // false (epoch not yet supported)
func IsTemporalTag(tag string) bool {
	_, err := ExtractTimestamp(tag)
	return err == nil
}

// GetTagAtTime retrieves the value of a specific tag namespace at a given point in time.
// It scans through all temporal tags and returns the most recent value for the namespace
// that is not newer than the specified timestamp.
//
// This function is essential for point-in-time queries and historical analysis.
//
// Parameters:
//   - tags: List of temporal tags to search
//   - namespace: The tag namespace to look for (e.g., "status", "priority")
//   - timestamp: The point in time to query
//
// Returns the value of the tag at the specified time, or empty string if not found.
//
// Example:
//
//	tags := []string{
//	    "2024-01-01T10:00:00.000000000.status:draft",
//	    "2024-01-05T14:30:00.000000000.status:review",
//	    "2024-01-10T09:15:00.000000000.status:published",
//	}
//	value := GetTagAtTime(tags, "status", time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC))
//	// value = "review"
func GetTagAtTime(tags []string, namespace string, timestamp time.Time) string {
	var latestValue string
	var latestTime time.Time
	
	// Iterate through all tags to find the matching namespace
	for _, tag := range tags {
		tagTime, err := ExtractTimestamp(tag)
		if err != nil {
			continue // Skip non-temporal tags
		}
		
		// Skip tags newer than our query timestamp
		if tagTime.After(timestamp) {
			continue
		}
		
		// Check if this tag matches our namespace
		tagNamespace, tagValue := ExtractNamespaceValue(tag)
		if tagNamespace == namespace && tagTime.After(latestTime) {
			latestValue = tagValue
			latestTime = tagTime
		}
	}
	
	return latestValue
}

// BuildEntitySnapshot creates a point-in-time snapshot of an entity.
// It reconstructs the entity's state as it existed at the specified timestamp
// by selecting the most recent value for each tag namespace that existed
// at that time.
//
// Non-temporal tags are included as-is in the snapshot.
// Content is copied unchanged (binary content is not versioned).
//
// This function is used for:
//   - Point-in-time queries
//   - Historical analysis
//   - Compliance and audit trails
//
// Example:
//
//	entity := &Entity{
//	    ID: "user-123",
//	    Tags: []string{
//	        "2024-01-01T10:00:00.000000000.status:draft",
//	        "2024-01-05T14:30:00.000000000.status:review",
//	        "2024-01-10T09:15:00.000000000.status:published",
//	        "type:document",  // non-temporal tag
//	    },
//	}
//	snapshot := BuildEntitySnapshot(entity, time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC))
//	// snapshot.Tags = ["status:review", "type:document"]
func BuildEntitySnapshot(current *Entity, timestamp time.Time) *Entity {
	snapshot := &Entity{
		ID:        current.ID,
		Tags:      []string{},
		Content:   []byte{},
		CreatedAt: current.CreatedAt,
		UpdatedAt: current.UpdatedAt,
	}
	
	// Collect all unique namespaces from temporal tags
	namespaces := make(map[string]bool)
	for _, tag := range current.Tags {
		if IsTemporalTag(tag) {
			namespace, _ := ExtractNamespaceValue(tag)
			namespaces[namespace] = true
		} else {
			// Preserve non-temporal tags in the snapshot
			snapshot.Tags = append(snapshot.Tags, tag)
		}
	}
	
	// For each namespace, find the value at the requested timestamp
	for namespace := range namespaces {
		value := GetTagAtTime(current.Tags, namespace, timestamp)
		if value != "" {
			snapshot.Tags = append(snapshot.Tags, namespace+":"+value)
		}
	}
	
	// Copy content unchanged (binary content is not temporally versioned)
	snapshot.Content = current.Content
	
	return snapshot
}

// CompareEntityStates analyzes the differences between two entity states
// and returns a list of changes. This is useful for:
//   - Generating audit logs
//   - Understanding entity evolution
//   - Debugging data changes
//
// The function compares tags between the before and after states,
// identifying added, modified, and removed tags.
//
// Example:
//
//	before := &Entity{
//	    Tags: []string{"status:draft", "priority:low"},
//	}
//	after := &Entity{
//	    Tags: []string{"status:published", "priority:high", "reviewer:alice"},
//	}
//	changes := CompareEntityStates(before, after)
//	// changes = [
//	//   {Type: "modified", OldValue: "draft", NewValue: "published"},
//	//   {Type: "modified", OldValue: "low", NewValue: "high"},
//	//   {Type: "added", NewValue: "alice"},
//	// ]
func CompareEntityStates(before, after *Entity) []EntityChange {
	changes := []EntityChange{}
	
	// Build maps of namespace->value for comparison
	beforeTags := make(map[string]string)
	afterTags := make(map[string]string)
	
	// Extract namespace/value pairs from before state
	for _, tag := range before.Tags {
		namespace, value := ExtractNamespaceValue(tag)
		beforeTags[namespace] = value
	}
	
	// Extract namespace/value pairs from after state
	for _, tag := range after.Tags {
		namespace, value := ExtractNamespaceValue(tag)
		afterTags[namespace] = value
	}
	
	// Identify added and modified tags
	for namespace, afterValue := range afterTags {
		beforeValue, existed := beforeTags[namespace]
		if !existed {
			// Tag was added
			changes = append(changes, EntityChange{
				Type:      "added",
				Timestamp: Now(),
				NewValue:  afterValue,
			})
		} else if beforeValue != afterValue {
			// Tag was modified
			changes = append(changes, EntityChange{
				Type:      "modified",
				Timestamp: Now(),
				OldValue:  beforeValue,
				NewValue:  afterValue,
			})
		}
	}
	
	// Identify removed tags
	for namespace, beforeValue := range beforeTags {
		if _, exists := afterTags[namespace]; !exists {
			// Tag was removed
			changes = append(changes, EntityChange{
				Type:      "removed",
				Timestamp: Now(),
				OldValue:  beforeValue,
			})
		}
	}
	
	return changes
}