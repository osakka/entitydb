// Package models provides retention policy management for EntityDB temporal deletion system
package models

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// RetentionPolicy defines rules for automatic entity lifecycle management
type RetentionPolicy struct {
	// Name is the unique identifier for this policy
	Name string `json:"name"`
	
	// Description explains what this policy does
	Description string `json:"description"`
	
	// Enabled controls whether this policy is active
	Enabled bool `json:"enabled"`
	
	// Priority determines execution order (lower numbers = higher priority)
	Priority int `json:"priority"`
	
	// Selector determines which entities this policy applies to
	Selector PolicySelector `json:"selector"`
	
	// Rules define the lifecycle transitions and timing
	Rules []RetentionRule `json:"rules"`
	
	// CreatedBy tracks who created this policy
	CreatedBy string `json:"created_by"`
	
	// CreatedAt tracks when this policy was created
	CreatedAt time.Time `json:"created_at"`
	
	// UpdatedBy tracks who last modified this policy
	UpdatedBy string `json:"updated_by"`
	
	// UpdatedAt tracks when this policy was last modified
	UpdatedAt time.Time `json:"updated_at"`
}

// PolicySelector defines criteria for selecting entities
type PolicySelector struct {
	// TagFilters apply to entities that have ALL of these tags
	TagFilters []string `json:"tag_filters,omitempty"`
	
	// TagPatterns apply to entities with tags matching these regex patterns
	TagPatterns []string `json:"tag_patterns,omitempty"`
	
	// EntityTypes apply to specific entity types (type:value tags)
	EntityTypes []string `json:"entity_types,omitempty"`
	
	// Datasets apply to specific datasets (dataset:value tags)
	Datasets []string `json:"datasets,omitempty"`
	
	// ExcludeTags exclude entities that have ANY of these tags
	ExcludeTags []string `json:"exclude_tags,omitempty"`
	
	// MinAge requires entities to be at least this old before policy applies
	MinAge string `json:"min_age,omitempty"`
	
	// MaxAge applies policy only to entities younger than this
	MaxAge string `json:"max_age,omitempty"`
}

// RetentionRule defines a specific lifecycle transition
type RetentionRule struct {
	// Name identifies this rule within the policy
	Name string `json:"name"`
	
	// FromState is the current lifecycle state to transition from
	FromState EntityLifecycleState `json:"from_state"`
	
	// ToState is the target lifecycle state to transition to
	ToState EntityLifecycleState `json:"to_state"`
	
	// Condition defines when this transition should occur
	Condition RuleCondition `json:"condition"`
	
	// Reason will be recorded in the entity's audit trail
	Reason string `json:"reason"`
	
	// Enabled controls whether this rule is active
	Enabled bool `json:"enabled"`
}

// RuleCondition defines when a retention rule should be executed
type RuleCondition struct {
	// Type determines the condition evaluation method
	Type ConditionType `json:"type"`
	
	// Value is the condition parameter (duration, count, etc.)
	Value string `json:"value"`
	
	// Field specifies which timestamp to use for age-based conditions
	Field string `json:"field,omitempty"`
}

// ConditionType defines the types of conditions supported
type ConditionType string

const (
	// ConditionAge triggers based on time elapsed since a specific timestamp
	ConditionAge ConditionType = "age"
	
	// ConditionStateAge triggers based on time elapsed since entering current state
	ConditionStateAge ConditionType = "state_age"
	
	// ConditionSize triggers based on entity content size
	ConditionSize ConditionType = "size"
	
	// ConditionTagExists triggers when a specific tag is present
	ConditionTagExists ConditionType = "tag_exists"
	
	// ConditionTagMissing triggers when a specific tag is missing
	ConditionTagMissing ConditionType = "tag_missing"
	
	// ConditionAlways triggers immediately (for testing)
	ConditionAlways ConditionType = "always"
)

// PolicyEngine manages retention policies and their execution
type PolicyEngine struct {
	policies []RetentionPolicy
	compiled map[string]*CompiledPolicy
}

// CompiledPolicy contains pre-compiled regex patterns for efficient matching
type CompiledPolicy struct {
	Policy      RetentionPolicy
	TagPatterns []*regexp.Regexp
}

// NewPolicyEngine creates a new retention policy engine
func NewPolicyEngine() *PolicyEngine {
	return &PolicyEngine{
		policies: make([]RetentionPolicy, 0),
		compiled: make(map[string]*CompiledPolicy),
	}
}

// AddPolicy adds a new retention policy to the engine
func (pe *PolicyEngine) AddPolicy(policy RetentionPolicy) error {
	// Validate policy
	if err := pe.validatePolicy(policy); err != nil {
		return fmt.Errorf("invalid policy: %w", err)
	}
	
	// Compile regex patterns
	compiled, err := pe.compilePolicy(policy)
	if err != nil {
		return fmt.Errorf("failed to compile policy patterns: %w", err)
	}
	
	// Add to engine
	pe.policies = append(pe.policies, policy)
	pe.compiled[policy.Name] = compiled
	
	return nil
}

// RemovePolicy removes a retention policy from the engine
func (pe *PolicyEngine) RemovePolicy(name string) error {
	for i, policy := range pe.policies {
		if policy.Name == name {
			// Remove from slice
			pe.policies = append(pe.policies[:i], pe.policies[i+1:]...)
			// Remove compiled version
			delete(pe.compiled, name)
			return nil
		}
	}
	return fmt.Errorf("policy not found: %s", name)
}

// GetPolicies returns all retention policies
func (pe *PolicyEngine) GetPolicies() []RetentionPolicy {
	return append([]RetentionPolicy(nil), pe.policies...)
}

// GetApplicablePolicies returns policies that apply to the given entity
func (pe *PolicyEngine) GetApplicablePolicies(entity *Entity) []RetentionPolicy {
	var applicable []RetentionPolicy
	
	for _, policy := range pe.policies {
		if !policy.Enabled {
			continue
		}
		
		if pe.matchesSelector(entity, policy.Selector) {
			applicable = append(applicable, policy)
		}
	}
	
	return applicable
}

// validatePolicy checks if a policy is valid
func (pe *PolicyEngine) validatePolicy(policy RetentionPolicy) error {
	if policy.Name == "" {
		return fmt.Errorf("policy name is required")
	}
	
	if len(policy.Rules) == 0 {
		return fmt.Errorf("policy must have at least one rule")
	}
	
	for i, rule := range policy.Rules {
		if rule.Name == "" {
			return fmt.Errorf("rule %d: name is required", i)
		}
		
		if !IsValidState(string(rule.FromState)) {
			return fmt.Errorf("rule %d: invalid from_state: %s", i, rule.FromState)
		}
		
		if !IsValidState(string(rule.ToState)) {
			return fmt.Errorf("rule %d: invalid to_state: %s", i, rule.ToState)
		}
		
		if rule.Reason == "" {
			return fmt.Errorf("rule %d: reason is required", i)
		}
		
		// Validate condition
		if err := pe.validateCondition(rule.Condition); err != nil {
			return fmt.Errorf("rule %d: %w", i, err)
		}
	}
	
	return nil
}

// validateCondition checks if a condition is valid
func (pe *PolicyEngine) validateCondition(condition RuleCondition) error {
	switch condition.Type {
	case ConditionAge, ConditionStateAge:
		if condition.Value == "" {
			return fmt.Errorf("age condition requires a duration value")
		}
		if _, err := time.ParseDuration(condition.Value); err != nil {
			return fmt.Errorf("invalid duration: %s", condition.Value)
		}
		
	case ConditionSize:
		if condition.Value == "" {
			return fmt.Errorf("size condition requires a size value")
		}
		
	case ConditionTagExists, ConditionTagMissing:
		if condition.Value == "" {
			return fmt.Errorf("tag condition requires a tag name")
		}
		
	case ConditionAlways:
		// Always valid
		
	default:
		return fmt.Errorf("unknown condition type: %s", condition.Type)
	}
	
	return nil
}

// compilePolicy pre-compiles regex patterns for efficient matching
func (pe *PolicyEngine) compilePolicy(policy RetentionPolicy) (*CompiledPolicy, error) {
	compiled := &CompiledPolicy{
		Policy:      policy,
		TagPatterns: make([]*regexp.Regexp, 0),
	}
	
	// Compile tag patterns
	for _, pattern := range policy.Selector.TagPatterns {
		regex, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid tag pattern %s: %w", pattern, err)
		}
		compiled.TagPatterns = append(compiled.TagPatterns, regex)
	}
	
	return compiled, nil
}

// matchesSelector checks if an entity matches the policy selector
func (pe *PolicyEngine) matchesSelector(entity *Entity, selector PolicySelector) bool {
	// Check tag filters (must have ALL)
	for _, requiredTag := range selector.TagFilters {
		if !pe.entityHasTag(entity, requiredTag) {
			return false
		}
	}
	
	// Check excluded tags (must have NONE)
	for _, excludedTag := range selector.ExcludeTags {
		if pe.entityHasTag(entity, excludedTag) {
			return false
		}
	}
	
	// Check entity types
	if len(selector.EntityTypes) > 0 {
		entityType := entity.GetTagValue("type")
		found := false
		for _, allowedType := range selector.EntityTypes {
			if entityType == allowedType {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	
	// Check datasets
	if len(selector.Datasets) > 0 {
		dataset := entity.GetTagValue("dataset")
		found := false
		for _, allowedDataset := range selector.Datasets {
			if dataset == allowedDataset {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	
	// Check tag patterns
	if len(selector.TagPatterns) > 0 {
		compiled := pe.compiled[selector.TagFilters[0]] // This needs fixing - should use policy name
		if compiled != nil {
			matched := false
			for _, tag := range entity.GetTagsWithoutTimestamp() {
				for _, pattern := range compiled.TagPatterns {
					if pattern.MatchString(tag) {
						matched = true
						break
					}
				}
				if matched {
					break
				}
			}
			if !matched {
				return false
			}
		}
	}
	
	// Check age constraints
	now := time.Now()
	entityAge := time.Unix(0, entity.CreatedAt)
	
	if selector.MinAge != "" {
		minDuration, err := time.ParseDuration(selector.MinAge)
		if err == nil {
			if now.Sub(entityAge) < minDuration {
				return false
			}
		}
	}
	
	if selector.MaxAge != "" {
		maxDuration, err := time.ParseDuration(selector.MaxAge)
		if err == nil {
			if now.Sub(entityAge) > maxDuration {
				return false
			}
		}
	}
	
	return true
}

// entityHasTag checks if an entity has a specific tag (ignoring timestamps)
func (pe *PolicyEngine) entityHasTag(entity *Entity, targetTag string) bool {
	for _, tag := range entity.GetTagsWithoutTimestamp() {
		if tag == targetTag {
			return true
		}
	}
	return false
}

// EvaluateRule checks if a retention rule should be executed for an entity
func (pe *PolicyEngine) EvaluateRule(entity *Entity, rule RetentionRule) (bool, error) {
	// Check if entity is in the correct state
	if entity.GetLifecycleState() != rule.FromState {
		return false, nil
	}
	
	// Evaluate condition
	return pe.evaluateCondition(entity, rule.Condition)
}

// evaluateCondition checks if a specific condition is met
func (pe *PolicyEngine) evaluateCondition(entity *Entity, condition RuleCondition) (bool, error) {
	now := time.Now()
	
	switch condition.Type {
	case ConditionAge:
		duration, err := time.ParseDuration(condition.Value)
		if err != nil {
			return false, fmt.Errorf("invalid duration: %s", condition.Value)
		}
		
		var referenceTime time.Time
		switch condition.Field {
		case "updated_at":
			referenceTime = time.Unix(0, entity.UpdatedAt)
		case "created_at", "":
			referenceTime = time.Unix(0, entity.CreatedAt)
		default:
			return false, fmt.Errorf("unknown time field: %s", condition.Field)
		}
		
		return now.Sub(referenceTime) >= duration, nil
		
	case ConditionStateAge:
		duration, err := time.ParseDuration(condition.Value)
		if err != nil {
			return false, fmt.Errorf("invalid duration: %s", condition.Value)
		}
		
		// Get when entity entered current state
		var stateTimestamp *time.Time
		switch entity.GetLifecycleState() {
		case StateSoftDeleted:
			stateTimestamp = entity.GetDeletedAt()
		case StateArchived:
			stateTimestamp = entity.GetArchivedAt()
		default:
			// For active state, use created_at as fallback
			created := time.Unix(0, entity.CreatedAt)
			stateTimestamp = &created
		}
		
		if stateTimestamp == nil {
			return false, nil
		}
		
		return now.Sub(*stateTimestamp) >= duration, nil
		
	case ConditionSize:
		sizeStr := strings.ToLower(condition.Value)
		var targetSize int64
		
		// Parse size with units (1kb, 1mb, 1gb)
		if strings.HasSuffix(sizeStr, "kb") {
			size, err := strconv.ParseInt(strings.TrimSuffix(sizeStr, "kb"), 10, 64)
			if err != nil {
				return false, fmt.Errorf("invalid size: %s", condition.Value)
			}
			targetSize = size * 1024
		} else if strings.HasSuffix(sizeStr, "mb") {
			size, err := strconv.ParseInt(strings.TrimSuffix(sizeStr, "mb"), 10, 64)
			if err != nil {
				return false, fmt.Errorf("invalid size: %s", condition.Value)
			}
			targetSize = size * 1024 * 1024
		} else if strings.HasSuffix(sizeStr, "gb") {
			size, err := strconv.ParseInt(strings.TrimSuffix(sizeStr, "gb"), 10, 64)
			if err != nil {
				return false, fmt.Errorf("invalid size: %s", condition.Value)
			}
			targetSize = size * 1024 * 1024 * 1024
		} else {
			// Assume bytes
			size, err := strconv.ParseInt(sizeStr, 10, 64)
			if err != nil {
				return false, fmt.Errorf("invalid size: %s", condition.Value)
			}
			targetSize = size
		}
		
		return int64(len(entity.Content)) >= targetSize, nil
		
	case ConditionTagExists:
		return pe.entityHasTag(entity, condition.Value), nil
		
	case ConditionTagMissing:
		return !pe.entityHasTag(entity, condition.Value), nil
		
	case ConditionAlways:
		return true, nil
		
	default:
		return false, fmt.Errorf("unknown condition type: %s", condition.Type)
	}
}

// DefaultPolicies returns a set of reasonable default retention policies
func DefaultPolicies() []RetentionPolicy {
	return []RetentionPolicy{
		{
			Name:        "standard-cleanup",
			Description: "Standard entity cleanup: soft delete after 90 days, archive after 30 days deleted, purge after 1 year archived",
			Enabled:     true,
			Priority:    100,
			Selector: PolicySelector{
				EntityTypes: []string{"document", "file", "content"},
				ExcludeTags: []string{"permanent", "no-delete"},
			},
			Rules: []RetentionRule{
				{
					Name:      "soft-delete-old",
					FromState: StateActive,
					ToState:   StateSoftDeleted,
					Condition: RuleCondition{
						Type:  ConditionAge,
						Value: "2160h", // 90 days
						Field: "updated_at",
					},
					Reason:  "Automatic cleanup: inactive for 90 days",
					Enabled: true,
				},
				{
					Name:      "archive-deleted",
					FromState: StateSoftDeleted,
					ToState:   StateArchived,
					Condition: RuleCondition{
						Type:  ConditionStateAge,
						Value: "720h", // 30 days
					},
					Reason:  "Automatic archive: deleted for 30 days",
					Enabled: true,
				},
				{
					Name:      "purge-archived",
					FromState: StateArchived,
					ToState:   StatePurged,
					Condition: RuleCondition{
						Type:  ConditionStateAge,
						Value: "8760h", // 1 year
					},
					Reason:  "Automatic purge: archived for 1 year",
					Enabled: true,
				},
			},
			CreatedBy: "system",
			CreatedAt: time.Now(),
			UpdatedBy: "system",
			UpdatedAt: time.Now(),
		},
		{
			Name:        "temp-file-cleanup",
			Description: "Aggressive cleanup for temporary files",
			Enabled:     true,
			Priority:    50,
			Selector: PolicySelector{
				TagFilters:  []string{"type:temp"},
				TagPatterns: []string{"name:temp_.*", "name:.*\\.tmp"},
			},
			Rules: []RetentionRule{
				{
					Name:      "delete-temp-files",
					FromState: StateActive,
					ToState:   StateSoftDeleted,
					Condition: RuleCondition{
						Type:  ConditionAge,
						Value: "24h", // 1 day
					},
					Reason:  "Automatic cleanup: temporary file older than 1 day",
					Enabled: true,
				},
				{
					Name:      "purge-temp-files",
					FromState: StateSoftDeleted,
					ToState:   StatePurged,
					Condition: RuleCondition{
						Type:  ConditionStateAge,
						Value: "168h", // 7 days
					},
					Reason:  "Automatic purge: deleted temporary file",
					Enabled: true,
				},
			},
			CreatedBy: "system",
			CreatedAt: time.Now(),
			UpdatedBy: "system",
			UpdatedAt: time.Now(),
		},
	}
}