package models

import (
	"fmt"
	"strings"
	"time"
)

// ExtractTimestamp extracts the timestamp from a temporal tag
func ExtractTimestamp(tag string) (time.Time, error) {
	// Format: YYYY-MM-DDTHH:MM:SS.nnnnnnnnn.namespace=value
	// or: YYYY-MM-DDTHH:MM:SS.nnnnnnnnn.namespace:value
	
	parts := strings.SplitN(tag, ".", 2)
	if len(parts) < 2 {
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

// ExtractNamespaceValue extracts namespace and value from a tag
func ExtractNamespaceValue(tag string) (namespace, value string) {
	// Remove timestamp if present
	if idx := strings.Index(tag, "."); idx != -1 {
		tag = tag[idx+1:]
	}
	
	// Split by = or :
	if idx := strings.Index(tag, "="); idx != -1 {
		return tag[:idx], tag[idx+1:]
	}
	if idx := strings.Index(tag, ":"); idx != -1 {
		return tag[:idx], tag[idx+1:]
	}
	
	return tag, ""
}

// IsTemporalTag checks if a tag has a timestamp prefix
func IsTemporalTag(tag string) bool {
	_, err := ExtractTimestamp(tag)
	return err == nil
}

// GetTagAtTime gets the latest value of a tag namespace at a given time
func GetTagAtTime(tags []string, namespace string, timestamp time.Time) string {
	var latestValue string
	var latestTime time.Time
	
	for _, tag := range tags {
		tagTime, err := ExtractTimestamp(tag)
		if err != nil {
			continue // Not a temporal tag
		}
		
		if tagTime.After(timestamp) {
			continue // Tag is from the future
		}
		
		tagNamespace, tagValue := ExtractNamespaceValue(tag)
		if tagNamespace == namespace && tagTime.After(latestTime) {
			latestValue = tagValue
			latestTime = tagTime
		}
	}
	
	return latestValue
}

// BuildEntitySnapshot builds an entity snapshot at a given point in time
func BuildEntitySnapshot(current *Entity, timestamp time.Time) *Entity {
	snapshot := &Entity{
		ID:        current.ID,
		Tags:      []string{},
		Content:   []byte{}, // Changed from []ContentItem{} to []byte{}
		CreatedAt: current.CreatedAt,
		UpdatedAt: current.UpdatedAt,
	}
	
	// Collect all unique namespaces
	namespaces := make(map[string]bool)
	for _, tag := range current.Tags {
		if IsTemporalTag(tag) {
			namespace, _ := ExtractNamespaceValue(tag)
			namespaces[namespace] = true
		} else {
			// Add non-temporal tags as-is
			snapshot.Tags = append(snapshot.Tags, tag)
		}
	}
	
	// For each namespace, get the value at the timestamp
	for namespace := range namespaces {
		value := GetTagAtTime(current.Tags, namespace, timestamp)
		if value != "" {
			snapshot.Tags = append(snapshot.Tags, namespace+":"+value)
		}
	}
	
	// Copy content (binary data doesn't have timestamps in new model)
	// Just copy the whole content as-is
	snapshot.Content = current.Content
	
	return snapshot
}

// CompareEntityStates compares two entity states and returns changes
func CompareEntityStates(before, after *Entity) []EntityChange {
	changes := []EntityChange{}
	
	// Compare tags
	beforeTags := make(map[string]string)
	afterTags := make(map[string]string)
	
	for _, tag := range before.Tags {
		namespace, value := ExtractNamespaceValue(tag)
		beforeTags[namespace] = value
	}
	
	for _, tag := range after.Tags {
		namespace, value := ExtractNamespaceValue(tag)
		afterTags[namespace] = value
	}
	
	// Find added and modified tags
	for namespace, afterValue := range afterTags {
		beforeValue, existed := beforeTags[namespace]
		if !existed {
			changes = append(changes, EntityChange{
				Type:      "added",
				Timestamp: time.Now(),
				NewValue:  afterValue,
			})
		} else if beforeValue != afterValue {
			changes = append(changes, EntityChange{
				Type:      "modified",
				Timestamp: time.Now(),
				OldValue:  beforeValue,
				NewValue:  afterValue,
			})
		}
	}
	
	// Find removed tags
	for namespace, beforeValue := range beforeTags {
		if _, exists := afterTags[namespace]; !exists {
			changes = append(changes, EntityChange{
				Type:      "removed",
				Timestamp: time.Now(),
				OldValue:  beforeValue,
			})
		}
	}
	
	return changes
}