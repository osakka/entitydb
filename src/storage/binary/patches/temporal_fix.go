package binary

import (
	"entitydb/models"
	"entitydb/logger"
	"fmt"
	"strings"
	"time"
	"strconv"
	"errors"
)

// ParseTemporalTagImproved parses a tag with timestamp prefix (enhanced version)
// Format can be either:
// 1. "2025-05-20T20:02:48.098692124+01:00|type:test" (pipe format)
// 2. "2025-05-20T20:02:48.098692124.type:test" (dot format, deprecated)
func ParseTemporalTagImproved(tag string) (int64, string, error) {
	// Try pipe format first (current standard)
	parts := strings.SplitN(tag, "|", 2)
	if len(parts) == 2 {
		// Parse timestamp
		t, err := time.Parse(time.RFC3339Nano, parts[0])
		if err != nil {
			return 0, "", fmt.Errorf("invalid timestamp in tag: %v", err)
		}
		return t.UnixNano(), parts[1], nil
	}
	
	// Try dot format (legacy)
	parts = strings.SplitN(tag, ".", 2)
	if len(parts) == 2 {
		// Parse timestamp
		t, err := time.Parse("2006-01-02T15:04:05.999999999", parts[0])
		if err != nil {
			return 0, "", fmt.Errorf("invalid timestamp in tag: %v", err)
		}
		return t.UnixNano(), parts[1], nil
	}
	
	// Special case: Unix timestamp format (numeric only)
	if parts := strings.SplitN(tag, "|", 2); len(parts) == 2 {
		if ts, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
			return ts, parts[1], nil
		}
	}
	
	return 0, "", errors.New("not a temporal tag")
}

// FormatTagWithTimestampImproved formats a tag with its timestamp (enhanced version)
func FormatTagWithTimestampImproved(tag string, timestamp int64) string {
	// Convert nanosecond timestamp to RFC3339Nano format
	t := time.Unix(0, timestamp)
	timeStr := t.Format(time.RFC3339Nano)
	return timeStr + "|" + tag
}

// Fixed implementation of GetEntityAsOf
func (r *TemporalRepository) GetEntityAsOfFixed(entityID string, asOf time.Time) (*models.Entity, error) {
	logger.Debug("Getting entity %s as of %v", entityID, asOf)
	
	// Get the current entity first
	current, err := r.GetByID(entityID)
	if err != nil {
		return nil, fmt.Errorf("entity not found: %s", err)
	}
	
	// Timestamp we're looking for
	asOfNanos := asOf.UnixNano()
	
	// Create a snapshot entity
	snapshot := &models.Entity{
		ID:      current.ID,
		Tags:    []string{},
		Content: current.Content, // Copy content as-is for now
	}
	
	// Process each tag from the current entity
	validTags := make([]string, 0)
	for _, tag := range current.Tags {
		// Parse the temporal tag
		tagTime, _, err := ParseTemporalTagImproved(tag)
		if err != nil {
			// If not a temporal tag, add as-is
			validTags = append(validTags, tag)
			continue
		}
		
		// Only include tags from before or at our target time
		if tagTime <= asOfNanos {
			// Keep the formatted temporal tag
			validTags = append(validTags, tag)
		}
	}
	
	// Set the filtered tags
	snapshot.Tags = validTags
	
	// Set timestamps to match
	if t, err := time.Parse(time.RFC3339, current.CreatedAt); err == nil && t.UnixNano() <= asOfNanos {
		snapshot.CreatedAt = current.CreatedAt
	} else {
		snapshot.CreatedAt = asOf.Format(time.RFC3339)
	}
	
	if t, err := time.Parse(time.RFC3339, current.UpdatedAt); err == nil && t.UnixNano() <= asOfNanos {
		snapshot.UpdatedAt = current.UpdatedAt
	} else {
		snapshot.UpdatedAt = asOf.Format(time.RFC3339)
	}
	
	return snapshot, nil
}

// Fixed implementation of GetEntityHistory
func (r *TemporalRepository) GetEntityHistoryFixed(entityID string, limit int) ([]*models.EntityChange, error) {
	logger.Debug("Getting history for entity %s (limit: %d)", entityID, limit)
	
	// Get the current entity
	current, err := r.GetByID(entityID)
	if err != nil {
		return nil, fmt.Errorf("entity not found: %s", err)
	}
	
	// Extract all timestamps from tags
	type TimestampedTag struct {
		Timestamp int64
		Tag       string
		Original  string
	}
	
	// Extract and collect all temporal tags
	tagHistory := make([]TimestampedTag, 0)
	for _, tag := range current.Tags {
		timestamp, cleanTag, err := ParseTemporalTagImproved(tag)
		if err == nil {
			tagHistory = append(tagHistory, TimestampedTag{
				Timestamp: timestamp,
				Tag:       cleanTag,
				Original:  tag,
			})
		}
	}
	
	// Sort by timestamp (newest first)
	for i := 0; i < len(tagHistory); i++ {
		for j := i + 1; j < len(tagHistory); j++ {
			if tagHistory[i].Timestamp < tagHistory[j].Timestamp {
				tagHistory[i], tagHistory[j] = tagHistory[j], tagHistory[i]
			}
		}
	}
	
	// Create change records
	changes := make([]*models.EntityChange, 0)
	for i := 0; i < len(tagHistory) && (limit <= 0 || i < limit); i++ {
		entry := tagHistory[i]
		change := &models.EntityChange{
			Type:      "tag_added",
			Timestamp: time.Unix(0, entry.Timestamp),
			NewValue:  entry.Tag,
		}
		changes = append(changes, change)
	}
	
	return changes, nil
}

// Fixed implementation of GetRecentChanges
func (r *TemporalRepository) GetRecentChangesFixed(limit int) ([]*models.EntityChange, error) {
	logger.Debug("Getting recent changes (limit: %d)", limit)
	
	// Get all entities
	entities, err := r.List()
	if err != nil {
		return nil, err
	}
	
	// Collect all changes across entities
	allChanges := make([]*models.EntityChange, 0)
	
	for _, entity := range entities {
		// Get history for this entity
		entityChanges, err := r.GetEntityHistoryFixed(entity.ID, limit)
		if err == nil && len(entityChanges) > 0 {
			// Add entity ID to each change
			for _, change := range entityChanges {
				change.NewValue = fmt.Sprintf("%s: %s", entity.ID, change.NewValue)
			}
			allChanges = append(allChanges, entityChanges...)
		}
	}
	
	// Sort all changes by timestamp (newest first)
	for i := 0; i < len(allChanges); i++ {
		for j := i + 1; j < len(allChanges); j++ {
			if allChanges[i].Timestamp.Before(allChanges[j].Timestamp) {
				allChanges[i], allChanges[j] = allChanges[j], allChanges[i]
			}
		}
	}
	
	// Apply limit
	if limit > 0 && len(allChanges) > limit {
		allChanges = allChanges[:limit]
	}
	
	return allChanges, nil
}

// Fixed implementation of GetEntityDiff
func (r *TemporalRepository) GetEntityDiffFixed(entityID string, startTime, endTime time.Time) (*models.Entity, *models.Entity, error) {
	logger.Debug("Getting diff for entity %s between %v and %v", entityID, startTime, endTime)
	
	// Get entity at both times
	entity1, err1 := r.GetEntityAsOfFixed(entityID, startTime)
	entity2, err2 := r.GetEntityAsOfFixed(entityID, endTime)
	
	// Handle cases where entity doesn't exist at one time
	if err1 != nil && err2 == nil {
		// Entity created between t1 and t2
		return nil, entity2, nil
	} else if err1 == nil && err2 != nil {
		// Entity deleted between t1 and t2
		return entity1, nil, nil
	} else if err1 != nil && err2 != nil {
		return nil, nil, fmt.Errorf("entity not found at either time")
	}
	
	return entity1, entity2, nil
}