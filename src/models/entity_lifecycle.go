// Package models provides entity lifecycle management for EntityDB temporal deletion system
package models

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// EntityLifecycleState represents the current state of an entity in its lifecycle
type EntityLifecycleState string

const (
	StateActive      EntityLifecycleState = "active"
	StateSoftDeleted EntityLifecycleState = "soft_deleted"
	StateArchived    EntityLifecycleState = "archived"
	StatePurged      EntityLifecycleState = "purged"
)

// IsValidState checks if the provided state is a valid lifecycle state
func IsValidState(state string) bool {
	switch EntityLifecycleState(state) {
	case StateActive, StateSoftDeleted, StateArchived, StatePurged:
		return true
	default:
		return false
	}
}

// LifecycleTransition represents a state change in entity lifecycle
type LifecycleTransition struct {
	FromState EntityLifecycleState `json:"from_state"`
	ToState   EntityLifecycleState `json:"to_state"`
	Timestamp time.Time            `json:"timestamp"`
	UserID    string               `json:"user_id"`
	Reason    string               `json:"reason"`
	Policy    string               `json:"policy"`
}

// EntityLifecycle manages the lifecycle state and transitions of entities
type EntityLifecycle struct {
	entity *Entity
}

// NewEntityLifecycle creates a new lifecycle manager for an entity
func NewEntityLifecycle(entity *Entity) *EntityLifecycle {
	return &EntityLifecycle{
		entity: entity,
	}
}

// GetCurrentState returns the current lifecycle state of the entity
func (el *EntityLifecycle) GetCurrentState() EntityLifecycleState {
	// Check for explicit status tags (most recent wins due to temporal ordering)
	statusTags := el.getTagsWithPrefix("status:")
	
	if len(statusTags) == 0 {
		// No explicit status, default to active for existing entities
		return StateActive
	}
	
	// Find the most recent status tag
	var latestTimestamp int64
	var latestState EntityLifecycleState = StateActive
	
	for _, tag := range statusTags {
		parts := strings.Split(tag, "|")
		if len(parts) >= 1 {
			statusPart := strings.TrimPrefix(parts[0], "status:")
			
			if IsValidState(statusPart) {
				state := EntityLifecycleState(statusPart)
				
				// If there's a timestamp, use it for ordering
				if len(parts) >= 2 {
					if timestamp, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
						if timestamp > latestTimestamp {
							latestTimestamp = timestamp
							latestState = state
						}
					}
				} else {
					// No timestamp, this is the current state
					latestState = state
				}
			}
		}
	}
	
	return latestState
}

// IsActive returns true if the entity is in active state
func (el *EntityLifecycle) IsActive() bool {
	return el.GetCurrentState() == StateActive
}

// IsSoftDeleted returns true if the entity is in soft deleted state
func (el *EntityLifecycle) IsSoftDeleted() bool {
	return el.GetCurrentState() == StateSoftDeleted
}

// IsArchived returns true if the entity is in archived state
func (el *EntityLifecycle) IsArchived() bool {
	return el.GetCurrentState() == StateArchived
}

// IsPurged returns true if the entity is in purged state
func (el *EntityLifecycle) IsPurged() bool {
	return el.GetCurrentState() == StatePurged
}

// CanTransitionTo checks if a transition to the target state is valid
func (el *EntityLifecycle) CanTransitionTo(targetState EntityLifecycleState) bool {
	currentState := el.GetCurrentState()
	
	// Define valid transitions
	validTransitions := map[EntityLifecycleState][]EntityLifecycleState{
		StateActive: {StateSoftDeleted},                           // Active can only be soft deleted
		StateSoftDeleted: {StateActive, StateArchived},           // Soft deleted can be undeleted or archived
		StateArchived: {StatePurged},                             // Archived can only be purged
		StatePurged: {},                                          // Purged is final state
	}
	
	allowedStates, exists := validTransitions[currentState]
	if !exists {
		return false
	}
	
	for _, allowedState := range allowedStates {
		if allowedState == targetState {
			return true
		}
	}
	
	return false
}

// TransitionTo changes the entity state with full audit trail
func (el *EntityLifecycle) TransitionTo(targetState EntityLifecycleState, userID, reason, policy string) error {
	currentState := el.GetCurrentState()
	
	// Validate transition
	if !el.CanTransitionTo(targetState) {
		return fmt.Errorf("invalid state transition from %s to %s", currentState, targetState)
	}
	
	timestamp := time.Now()
	timestampNano := timestamp.UnixNano()
	
	// Add new status tag with timestamp
	statusTag := fmt.Sprintf("status:%s|%d", targetState, timestampNano)
	el.entity.AddTag(statusTag)
	
	// Add transition metadata
	switch targetState {
	case StateSoftDeleted:
		el.entity.AddTag(fmt.Sprintf("deleted_by:%s|%d", userID, timestampNano))
		el.entity.AddTag(fmt.Sprintf("delete_reason:%s|%d", reason, timestampNano))
		if policy != "" {
			el.entity.AddTag(fmt.Sprintf("deletion_policy:%s|%d", policy, timestampNano))
		}
		
	case StateActive:
		// This is an undelete operation
		el.entity.AddTag(fmt.Sprintf("restored_by:%s|%d", userID, timestampNano))
		el.entity.AddTag(fmt.Sprintf("restore_reason:%s|%d", reason, timestampNano))
		
	case StateArchived:
		el.entity.AddTag(fmt.Sprintf("archived_by:%s|%d", userID, timestampNano))
		el.entity.AddTag(fmt.Sprintf("archive_reason:%s|%d", reason, timestampNano))
		if policy != "" {
			el.entity.AddTag(fmt.Sprintf("archive_policy:%s|%d", policy, timestampNano))
		}
		
	case StatePurged:
		el.entity.AddTag(fmt.Sprintf("purged_by:%s|%d", userID, timestampNano))
		el.entity.AddTag(fmt.Sprintf("purge_reason:%s|%d", reason, timestampNano))
		if policy != "" {
			el.entity.AddTag(fmt.Sprintf("purge_policy:%s|%d", policy, timestampNano))
		}
	}
	
	// Add transition audit tag
	transitionTag := fmt.Sprintf("transition:%s->%s|%d", currentState, targetState, timestampNano)
	el.entity.AddTag(transitionTag)
	
	// Update entity timestamps
	el.entity.UpdatedAt = timestampNano
	
	return nil
}

// SoftDelete transitions entity to soft deleted state
func (el *EntityLifecycle) SoftDelete(userID, reason, policy string) error {
	return el.TransitionTo(StateSoftDeleted, userID, reason, policy)
}

// Undelete transitions entity back to active state
func (el *EntityLifecycle) Undelete(userID, reason string) error {
	return el.TransitionTo(StateActive, userID, reason, "")
}

// Archive transitions entity to archived state
func (el *EntityLifecycle) Archive(userID, reason, policy string) error {
	return el.TransitionTo(StateArchived, userID, reason, policy)
}

// Purge transitions entity to purged state
func (el *EntityLifecycle) Purge(userID, reason, policy string) error {
	return el.TransitionTo(StatePurged, userID, reason, policy)
}

// GetTransitionHistory returns all lifecycle transitions for this entity
func (el *EntityLifecycle) GetTransitionHistory() []LifecycleTransition {
	var transitions []LifecycleTransition
	
	transitionTags := el.getTagsWithPrefix("transition:")
	
	for _, tag := range transitionTags {
		// Parse transition tag: "transition:active->soft_deleted|1234567890"
		parts := strings.Split(tag, "|")
		if len(parts) >= 2 {
			transitionPart := strings.TrimPrefix(parts[0], "transition:")
			stateParts := strings.Split(transitionPart, "->")
			
			if len(stateParts) == 2 {
				if timestamp, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
					transition := LifecycleTransition{
						FromState: EntityLifecycleState(stateParts[0]),
						ToState:   EntityLifecycleState(stateParts[1]),
						Timestamp: time.Unix(0, timestamp),
					}
					
					// Extract additional metadata from corresponding tags
					transition.UserID = el.getMetadataForTimestamp("_by:", timestamp)
					transition.Reason = el.getMetadataForTimestamp("_reason:", timestamp)
					transition.Policy = el.getMetadataForTimestamp("_policy:", timestamp)
					
					transitions = append(transitions, transition)
				}
			}
		}
	}
	
	return transitions
}

// getMetadataForTimestamp finds metadata tags with specific timestamp
func (el *EntityLifecycle) getMetadataForTimestamp(suffix string, timestamp int64) string {
	timestampStr := fmt.Sprintf("|%d", timestamp)
	
	for _, tag := range el.entity.Tags {
		if strings.Contains(tag, suffix) && strings.HasSuffix(tag, timestampStr) {
			parts := strings.Split(tag, ":")
			if len(parts) >= 2 {
				valuePart := strings.Join(parts[1:], ":")
				valuePart = strings.TrimSuffix(valuePart, timestampStr)
				return valuePart
			}
		}
	}
	
	return ""
}

// GetDeletedAt returns when the entity was deleted (if it has been deleted)
func (el *EntityLifecycle) GetDeletedAt() *time.Time {
	for _, tag := range el.entity.Tags {
		if strings.HasPrefix(tag, "status:soft_deleted|") {
			parts := strings.Split(tag, "|")
			if len(parts) >= 2 {
				if timestamp, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
					deletedAt := time.Unix(0, timestamp)
					return &deletedAt
				}
			}
		}
	}
	return nil
}

// GetArchivedAt returns when the entity was archived (if it has been archived)
func (el *EntityLifecycle) GetArchivedAt() *time.Time {
	for _, tag := range el.entity.Tags {
		if strings.HasPrefix(tag, "status:archived|") {
			parts := strings.Split(tag, "|")
			if len(parts) >= 2 {
				if timestamp, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
					archivedAt := time.Unix(0, timestamp)
					return &archivedAt
				}
			}
		}
	}
	return nil
}

// GetDeletionPolicy returns the deletion policy applied to this entity
func (el *EntityLifecycle) GetDeletionPolicy() string {
	policyTags := el.getTagsWithPrefix("deletion_policy:")
	
	if len(policyTags) == 0 {
		return ""
	}
	
	// Find the most recent policy tag
	var latestTimestamp int64
	var latestPolicy string
	
	for _, tag := range policyTags {
		parts := strings.Split(tag, "|")
		policyPart := strings.TrimPrefix(parts[0], "deletion_policy:")
		
		if len(parts) >= 2 {
			if timestamp, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
				if timestamp > latestTimestamp {
					latestTimestamp = timestamp
					latestPolicy = policyPart
				}
			}
		} else {
			latestPolicy = policyPart
		}
	}
	
	return latestPolicy
}

// GetDeletedBy returns who deleted the entity
func (el *EntityLifecycle) GetDeletedBy() string {
	return el.getLatestMetadata("deleted_by:")
}

// GetDeleteReason returns why the entity was deleted
func (el *EntityLifecycle) GetDeleteReason() string {
	return el.getLatestMetadata("delete_reason:")
}

// getLatestMetadata gets the most recent value for a metadata tag prefix
func (el *EntityLifecycle) getLatestMetadata(prefix string) string {
	metadataTags := el.getTagsWithPrefix(prefix)
	
	if len(metadataTags) == 0 {
		return ""
	}
	
	var latestTimestamp int64
	var latestValue string
	
	for _, tag := range metadataTags {
		parts := strings.Split(tag, "|")
		valuePart := strings.TrimPrefix(parts[0], prefix)
		
		if len(parts) >= 2 {
			if timestamp, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
				if timestamp > latestTimestamp {
					latestTimestamp = timestamp
					latestValue = valuePart
				}
			}
		} else {
			latestValue = valuePart
		}
	}
	
	return latestValue
}

// getTagsWithPrefix returns all tags that start with the given prefix
func (el *EntityLifecycle) getTagsWithPrefix(prefix string) []string {
	var matchingTags []string
	
	for _, tag := range el.entity.Tags {
		// Handle temporal tags by extracting the actual tag content
		actualTag := tag
		if strings.Contains(tag, "|") {
			parts := strings.SplitN(tag, "|", 2)
			if len(parts) >= 2 {
				actualTag = parts[1]
			}
		}
		
		if strings.HasPrefix(actualTag, prefix) {
			matchingTags = append(matchingTags, tag) // Return the full temporal tag
		}
	}
	
	return matchingTags
}