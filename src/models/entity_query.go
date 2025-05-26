package models

import (
	"strconv"
	"strings"
	"time"
)

// EntityQuery provides a builder pattern for querying entities
type EntityQuery struct {
	repo       EntityRepository
	tags       []string
	wildcards  []string
	content    string
	namespace  string
	limit      int
	offset     int
	orderBy    string
	orderDir   string
	filters    []Filter
	operators  []string  // AND/OR operators for filters
}

// NewEntityQuery creates a new query builder for the given repository
func NewEntityQuery(repo EntityRepository) *EntityQuery {
	return &EntityQuery{
		repo:      repo,
		tags:      []string{},
		wildcards: []string{},
		filters:   []Filter{},
		operators: []string{},
	}
}

// Filter represents a filtering condition
type Filter struct {
	Field    string
	Operator string // eq, ne, gt, lt, gte, lte, like, in
	Value    interface{}
}

// SortField represents available fields for sorting
type SortField string

const (
	SortByCreatedAt  SortField = "created_at"
	SortByUpdatedAt  SortField = "updated_at"
	SortByID         SortField = "id"
	SortByTagCount   SortField = "tag_count"
)

// SortDirection represents sorting direction
type SortDirection string

const (
	SortAsc  SortDirection = "asc"
	SortDesc SortDirection = "desc"
)


// HasTag adds a tag filter
func (q *EntityQuery) HasTag(tag string) *EntityQuery {
	q.tags = append(q.tags, tag)
	return q
}

// HasWildcardTag adds a wildcard tag filter
func (q *EntityQuery) HasWildcardTag(pattern string) *EntityQuery {
	q.wildcards = append(q.wildcards, pattern)
	return q
}

// SearchContent adds a content search filter
func (q *EntityQuery) SearchContent(search string) *EntityQuery {
	q.content = search
	return q
}

// InNamespace filters by namespace
func (q *EntityQuery) InNamespace(namespace string) *EntityQuery {
	q.namespace = namespace
	return q
}

// Limit sets the maximum number of results
func (q *EntityQuery) Limit(limit int) *EntityQuery {
	q.limit = limit
	return q
}

// Offset sets the starting offset
func (q *EntityQuery) Offset(offset int) *EntityQuery {
	q.offset = offset
	return q
}

// OrderBy sets the ordering
func (q *EntityQuery) OrderBy(field string, direction string) *EntityQuery {
	q.orderBy = field
	q.orderDir = direction
	return q
}

// AddFilter adds a filter condition
func (q *EntityQuery) AddFilter(field, operator string, value interface{}) *EntityQuery {
	q.filters = append(q.filters, Filter{
		Field:    field,
		Operator: operator,
		Value:    value,
	})
	return q
}

// And adds an AND operator for the next filter
func (q *EntityQuery) And() *EntityQuery {
	q.operators = append(q.operators, "AND")
	return q
}

// Or adds an OR operator for the next filter
func (q *EntityQuery) Or() *EntityQuery {
	q.operators = append(q.operators, "OR")
	return q
}

// Execute runs the query
func (q *EntityQuery) Execute() ([]*Entity, error) {
	// Start with all entities
	var entities []*Entity
	var err error
	
	// Apply filters based on query parameters
	if len(q.tags) > 0 && len(q.tags) == 1 {
		entities, err = q.repo.ListByTag(q.tags[0])
	} else if len(q.tags) > 1 {
		entities, err = q.repo.ListByTags(q.tags, true) // matchAll = true
	} else if len(q.wildcards) > 0 {
		// Handle wildcards - use first one
		entities, err = q.repo.ListByTagWildcard(q.wildcards[0])
	} else if q.content != "" {
		entities, err = q.repo.SearchContent(q.content)
	} else if q.namespace != "" {
		entities, err = q.repo.ListByNamespace(q.namespace)
	} else {
		entities, err = q.repo.List()
	}
	
	if err != nil {
		return nil, err
	}
	
	// Apply additional filters
	filtered := entities
	
	// Apply multiple tag filters
	if len(q.tags) > 1 {
		filtered = filterByTags(filtered, q.tags)
	}
	
	// Apply wildcards if there are multiple
	if len(q.wildcards) > 1 {
		for _, wildcard := range q.wildcards[1:] {
			filtered = filterByWildcard(filtered, wildcard)
		}
	}
	
	// Apply custom filters
	if len(q.filters) > 0 {
		filtered = q.applyFilters(filtered)
	}
	
	// Apply sorting
	filtered = q.applySorting(filtered)
	
	// Apply pagination (limit and offset)
	if q.offset > 0 && q.offset < len(filtered) {
		filtered = filtered[q.offset:]
	}
	
	if q.limit > 0 && q.limit < len(filtered) {
		filtered = filtered[:q.limit]
	}
	
	return filtered, nil
}

// Helper functions

func filterByTags(entities []*Entity, tags []string) []*Entity {
	var result []*Entity
	for _, entity := range entities {
		hasAllTags := true
		for _, tag := range tags {
			found := false
			for _, entityTag := range entity.Tags {
				if entityTag == tag {
					found = true
					break
				}
			}
			if !found {
				hasAllTags = false
				break
			}
		}
		if hasAllTags {
			result = append(result, entity)
		}
	}
	return result
}

func filterByWildcard(entities []*Entity, pattern string) []*Entity {
	var result []*Entity
	for _, entity := range entities {
		for _, tag := range entity.Tags {
			if matchesWildcard(tag, pattern) {
				result = append(result, entity)
				break
			}
		}
	}
	return result
}

func matchesWildcard(tag, pattern string) bool {
	// Simple wildcard matching
	if pattern == "*" {
		return true
	}
	
	// Handle patterns like "type:*"
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(tag) >= len(prefix) && tag[:len(prefix)] == prefix
	}
	
	return tag == pattern
}

func sortByCreatedAt(entities []*Entity) {
	// Simple bubble sort for simplicity
	for i := 0; i < len(entities)-1; i++ {
		for j := 0; j < len(entities)-i-1; j++ {
			if entities[j].CreatedAt > entities[j+1].CreatedAt {
				entities[j], entities[j+1] = entities[j+1], entities[j]
			}
		}
	}
}

func sortByCreatedAtDesc(entities []*Entity) {
	// Simple bubble sort for simplicity
	for i := 0; i < len(entities)-1; i++ {
		for j := 0; j < len(entities)-i-1; j++ {
			if entities[j].CreatedAt < entities[j+1].CreatedAt {
				entities[j], entities[j+1] = entities[j+1], entities[j]
			}
		}
	}
}

// applyFilters applies custom filters to entities
func (q *EntityQuery) applyFilters(entities []*Entity) []*Entity {
	if len(q.filters) == 0 {
		return entities
	}

	result := make([]*Entity, 0)
	
	for _, entity := range entities {
		shouldInclude := true
		
		for i, filter := range q.filters {
			matches := q.evaluateFilter(entity, filter)
			
			// Handle logical operators
			if i > 0 && i-1 < len(q.operators) {
				operator := q.operators[i-1]
				if operator == "OR" {
					shouldInclude = shouldInclude || matches
				} else { // Default to AND
					shouldInclude = shouldInclude && matches
				}
			} else {
				shouldInclude = matches
			}
		}
		
		if shouldInclude {
			result = append(result, entity)
		}
	}
	
	return result
}

// evaluateFilter checks if an entity matches a filter
func (q *EntityQuery) evaluateFilter(entity *Entity, filter Filter) bool {
	switch filter.Field {
	case "created_at":
		return q.evaluateTimeFilter(entity.CreatedAt, filter)
	case "updated_at":
		return q.evaluateTimeFilter(entity.UpdatedAt, filter)
	case "id":
		return q.evaluateStringFilter(entity.ID, filter)
	case "tag_count":
		return q.evaluateNumericFilter(float64(len(entity.Tags)), filter)
	case "content_type":
		// For new model, content type is stored in tags
		for _, tag := range entity.Tags {
			if strings.HasPrefix(tag, "content:type:") {
				contentType := strings.TrimPrefix(tag, "content:type:")
				if q.evaluateStringFilter(contentType, filter) {
					return true
				}
			}
		}
		return false
	case "content_value":
		// Search in content binary data (convert to string for filtering)
		if len(entity.Content) > 0 {
			contentStr := string(entity.Content)
			if q.evaluateStringFilter(contentStr, filter) {
				return true
			}
		}
		return false
	default:
		// Check if it's a tag filter (e.g., "tag:type")
		if strings.HasPrefix(filter.Field, "tag:") {
			tagNamespace := strings.TrimPrefix(filter.Field, "tag:")
			for _, tag := range entity.Tags {
				if strings.HasPrefix(tag, tagNamespace+":") {
					tagValue := strings.TrimPrefix(tag, tagNamespace+":")
					if q.evaluateStringFilter(tagValue, filter) {
						return true
					}
				}
			}
		}
		return false
	}
}

// evaluateStringFilter evaluates a string filter
func (q *EntityQuery) evaluateStringFilter(value string, filter Filter) bool {
	filterValue, ok := filter.Value.(string)
	if !ok {
		return false
	}
	
	switch filter.Operator {
	case "eq":
		return value == filterValue
	case "ne":
		return value != filterValue
	case "like":
		return strings.Contains(strings.ToLower(value), strings.ToLower(filterValue))
	case "in":
		// Handle comma-separated values
		values := strings.Split(filterValue, ",")
		for _, v := range values {
			if value == strings.TrimSpace(v) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

// evaluateNumericFilter evaluates a numeric filter
func (q *EntityQuery) evaluateNumericFilter(value float64, filter Filter) bool {
	var filterValue float64
	
	switch v := filter.Value.(type) {
	case float64:
		filterValue = v
	case int:
		filterValue = float64(v)
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return false
		}
		filterValue = parsed
	default:
		return false
	}
	
	switch filter.Operator {
	case "eq":
		return value == filterValue
	case "ne":
		return value != filterValue
	case "gt":
		return value > filterValue
	case "lt":
		return value < filterValue
	case "gte":
		return value >= filterValue
	case "lte":
		return value <= filterValue
	default:
		return false
	}
}

// evaluateTimeNanoFilter evaluates a time-based filter for nanosecond epoch timestamps
func (q *EntityQuery) evaluateTimeNanoFilter(value int64, filter Filter) bool {
	return q.evaluateTimeFilter(value, filter)
}

// evaluateTimeFilter evaluates a time-based filter
func (q *EntityQuery) evaluateTimeFilter(value int64, filter Filter) bool {
	var filterTime int64
	
	switch v := filter.Value.(type) {
	case string:
		// Parse time string (supports RFC3339)
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			// Try parsing as Unix timestamp
			timestamp, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return false
			}
			filterTime = timestamp
		} else {
			filterTime = t.UnixNano()
		}
	case int64:
		filterTime = v
	case int:
		filterTime = int64(v)
	default:
		return false
	}
	
	switch filter.Operator {
	case "eq":
		return value == filterTime
	case "ne":
		return value != filterTime
	case "gt":
		return value > filterTime
	case "lt":
		return value < filterTime
	case "gte":
		return value >= filterTime
	case "lte":
		return value <= filterTime
	default:
		return false
	}
}

// applySorting applies sorting to entities
func (q *EntityQuery) applySorting(entities []*Entity) []*Entity {
	if q.orderBy == "" {
		return entities
	}
	
	// Create a copy to avoid modifying the original slice
	sorted := make([]*Entity, len(entities))
	copy(sorted, entities)
	
	// Define sort function based on field and direction
	switch q.orderBy {
	case "created_at":
		if q.orderDir == "desc" {
			sortByCreatedAtDesc(sorted)
		} else {
			sortByCreatedAt(sorted)
		}
	case "updated_at":
		if q.orderDir == "desc" {
			sortByUpdatedAtDesc(sorted)
		} else {
			sortByUpdatedAt(sorted)
		}
	case "id":
		if q.orderDir == "desc" {
			sortByIDDesc(sorted)
		} else {
			sortByID(sorted)
		}
	case "tag_count":
		if q.orderDir == "desc" {
			sortByTagCountDesc(sorted)
		} else {
			sortByTagCount(sorted)
		}
	}
	
	return sorted
}

// Additional sort functions
func sortByUpdatedAt(entities []*Entity) {
	for i := 0; i < len(entities)-1; i++ {
		for j := 0; j < len(entities)-i-1; j++ {
			if entities[j].UpdatedAt > entities[j+1].UpdatedAt {
				entities[j], entities[j+1] = entities[j+1], entities[j]
			}
		}
	}
}

func sortByUpdatedAtDesc(entities []*Entity) {
	for i := 0; i < len(entities)-1; i++ {
		for j := 0; j < len(entities)-i-1; j++ {
			if entities[j].UpdatedAt < entities[j+1].UpdatedAt {
				entities[j], entities[j+1] = entities[j+1], entities[j]
			}
		}
	}
}

func sortByID(entities []*Entity) {
	for i := 0; i < len(entities)-1; i++ {
		for j := 0; j < len(entities)-i-1; j++ {
			if entities[j].ID > entities[j+1].ID {
				entities[j], entities[j+1] = entities[j+1], entities[j]
			}
		}
	}
}

func sortByIDDesc(entities []*Entity) {
	for i := 0; i < len(entities)-1; i++ {
		for j := 0; j < len(entities)-i-1; j++ {
			if entities[j].ID < entities[j+1].ID {
				entities[j], entities[j+1] = entities[j+1], entities[j]
			}
		}
	}
}

func sortByTagCount(entities []*Entity) {
	for i := 0; i < len(entities)-1; i++ {
		for j := 0; j < len(entities)-i-1; j++ {
			if len(entities[j].Tags) > len(entities[j+1].Tags) {
				entities[j], entities[j+1] = entities[j+1], entities[j]
			}
		}
	}
}

func sortByTagCountDesc(entities []*Entity) {
	for i := 0; i < len(entities)-1; i++ {
		for j := 0; j < len(entities)-i-1; j++ {
			if len(entities[j].Tags) < len(entities[j+1].Tags) {
				entities[j], entities[j+1] = entities[j+1], entities[j]
			}
		}
	}
}

// compareNanoTimestamps compares two nanosecond epoch timestamps
func compareNanoTimestamps(t1, t2 int64) int {
	if t1 < t2 {
		return -1
	} else if t1 > t2 {
		return 1
	}
	return 0
}