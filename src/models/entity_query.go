// Package models provides core data structures and business logic for EntityDB.
// This file implements a flexible query builder for entity searches.
package models

import (
	"strconv"
	"strings"
	"time"
)

// EntityQuery provides a fluent builder pattern for constructing complex entity queries.
// It supports filtering by tags, content search, namespaces, custom filters,
// sorting, and pagination. Queries are built up through method chaining and
// executed with the Execute() method.
//
// Example usage:
//
//	// Find all user entities created in the last 24 hours
//	query := NewEntityQuery(repo).
//	    HasTag("type:user").
//	    AddFilter("created_at", "gt", time.Now().Add(-24*time.Hour).UnixNano()).
//	    OrderBy("created_at", "desc").
//	    Limit(10)
//	
//	entities, err := query.Execute()
//
// The query builder supports:
//   - Tag filtering with exact matches and wildcards
//   - Content searching
//   - Namespace filtering
//   - Custom field filters with various operators
//   - Sorting by multiple fields
//   - Pagination with limit and offset
type EntityQuery struct {
	// repo is the underlying repository to query
	repo EntityRepository
	
	// tags contains exact tag matches required (AND logic)
	tags []string
	
	// wildcards contains wildcard patterns for tag matching
	wildcards []string
	
	// content is the search string for content filtering
	content string
	
	// namespace filters entities to a specific tag namespace
	namespace string
	
	// limit sets the maximum number of results to return
	limit int
	
	// offset sets the starting position for pagination
	offset int
	
	// orderBy specifies the field to sort by
	orderBy string
	
	// orderDir specifies the sort direction (asc/desc)
	orderDir string
	
	// filters contains custom field filters to apply
	filters []Filter
	
	// operators contains logical operators (AND/OR) between filters
	operators []string
}

// NewEntityQuery creates a new query builder for the given repository.
// The query starts with no filters and will return all entities if executed immediately.
//
// Parameters:
//   - repo: The entity repository to query against
//
// Returns:
//   - *EntityQuery: A new query builder instance
func NewEntityQuery(repo EntityRepository) *EntityQuery {
	return &EntityQuery{
		repo:      repo,
		tags:      []string{},
		wildcards: []string{},
		filters:   []Filter{},
		operators: []string{},
	}
}

// Filter represents a single filtering condition that can be applied to entities.
// Filters support various operators for different data types (strings, numbers, times).
//
// Example:
//
//	filter := Filter{
//	    Field:    "created_at",
//	    Operator: "gt",
//	    Value:    time.Now().Add(-7*24*time.Hour).UnixNano(),
//	}
type Filter struct {
	// Field is the entity field to filter on.
	// Supported fields: "created_at", "updated_at", "id", "tag_count",
	// "content_type", "content_value", or "tag:namespace" for tag values
	Field string
	
	// Operator is the comparison operator to use.
	// Supported operators:
	//   - String fields: "eq", "ne", "like", "in"
	//   - Numeric/Time fields: "eq", "ne", "gt", "lt", "gte", "lte"
	Operator string
	
	// Value is the value to compare against.
	// Type depends on the field being filtered
	Value interface{}
}

// SortField represents the available fields that entities can be sorted by.
// Use these constants with the OrderBy method for type-safe sorting.
type SortField string

const (
	// SortByCreatedAt sorts by entity creation timestamp
	SortByCreatedAt SortField = "created_at"
	
	// SortByUpdatedAt sorts by entity last update timestamp
	SortByUpdatedAt SortField = "updated_at"
	
	// SortByID sorts by entity ID alphabetically
	SortByID SortField = "id"
	
	// SortByTagCount sorts by the number of tags on each entity
	SortByTagCount SortField = "tag_count"
)

// SortDirection represents the direction for sorting results.
type SortDirection string

const (
	// SortAsc sorts in ascending order (smallest to largest, A to Z, oldest to newest)
	SortAsc SortDirection = "asc"
	
	// SortDesc sorts in descending order (largest to smallest, Z to A, newest to oldest)
	SortDesc SortDirection = "desc"
)


// HasTag adds an exact tag match requirement to the query.
// Multiple tag filters are combined with AND logic - entities must have all specified tags.
// Tags should be in the format "namespace:value" (e.g., "type:user", "status:active").
//
// Parameters:
//   - tag: The exact tag string to match
//
// Returns:
//   - *EntityQuery: The query builder for method chaining
//
// Example:
//
//	// Find entities with both tags
//	query.HasTag("type:user").HasTag("status:active")
func (q *EntityQuery) HasTag(tag string) *EntityQuery {
	q.tags = append(q.tags, tag)
	return q
}

// HasWildcardTag adds a wildcard pattern for tag matching.
// Wildcards use "*" to match any characters after a prefix.
// Multiple wildcard patterns are applied with OR logic.
//
// Parameters:
//   - pattern: The wildcard pattern (e.g., "type:*", "status:act*")
//
// Returns:
//   - *EntityQuery: The query builder for method chaining
//
// Example:
//
//	// Find all entities with any type tag
//	query.HasWildcardTag("type:*")
func (q *EntityQuery) HasWildcardTag(pattern string) *EntityQuery {
	q.wildcards = append(q.wildcards, pattern)
	return q
}

// SearchContent adds a content search filter to find entities containing
// the specified text in their content field. The search is case-insensitive.
// Only one content search can be active at a time - calling this multiple
// times will replace the previous search term.
//
// Parameters:
//   - search: The text to search for in entity content
//
// Returns:
//   - *EntityQuery: The query builder for method chaining
//
// Example:
//
//	// Find entities containing "configuration" in their content
//	query.SearchContent("configuration")
func (q *EntityQuery) SearchContent(search string) *EntityQuery {
	q.content = search
	return q
}

// InNamespace filters entities to only those with tags in the specified namespace.
// This is useful for finding all entities of a certain type or category.
// Only one namespace filter can be active at a time.
//
// Parameters:
//   - namespace: The tag namespace to filter by (e.g., "type", "status")
//
// Returns:
//   - *EntityQuery: The query builder for method chaining
//
// Example:
//
//	// Find all entities with any tag in the "type" namespace
//	query.InNamespace("type")
func (q *EntityQuery) InNamespace(namespace string) *EntityQuery {
	q.namespace = namespace
	return q
}

// Limit sets the maximum number of results to return.
// This is used for pagination in combination with Offset.
// A limit of 0 or negative means no limit.
//
// Parameters:
//   - limit: Maximum number of entities to return
//
// Returns:
//   - *EntityQuery: The query builder for method chaining
//
// Example:
//
//	// Return at most 20 results
//	query.Limit(20)
func (q *EntityQuery) Limit(limit int) *EntityQuery {
	q.limit = limit
	return q
}

// Offset sets the starting position for results.
// This is used for pagination in combination with Limit.
// Results are skipped from the beginning of the filtered set.
//
// Parameters:
//   - offset: Number of entities to skip
//
// Returns:
//   - *EntityQuery: The query builder for method chaining
//
// Example:
//
//	// Skip the first 40 results (page 3 with 20 per page)
//	query.Offset(40).Limit(20)
func (q *EntityQuery) Offset(offset int) *EntityQuery {
	q.offset = offset
	return q
}

// OrderBy sets the field and direction for sorting results.
// Valid fields are defined in the SortField constants.
// Direction should be "asc" or "desc" (use SortDirection constants).
//
// Parameters:
//   - field: The field to sort by (created_at, updated_at, id, tag_count)
//   - direction: Sort direction ("asc" or "desc")
//
// Returns:
//   - *EntityQuery: The query builder for method chaining
//
// Example:
//
//	// Sort by creation date, newest first
//	query.OrderBy("created_at", "desc")
//	// Or using constants
//	query.OrderBy(string(SortByCreatedAt), string(SortDesc))
func (q *EntityQuery) OrderBy(field string, direction string) *EntityQuery {
	q.orderBy = field
	q.orderDir = direction
	return q
}

// AddFilter adds a custom field filter to the query.
// Filters are evaluated in the order they are added.
// Use And() or Or() between filters to control logic.
//
// Parameters:
//   - field: The field to filter on (see Filter type for supported fields)
//   - operator: The comparison operator (eq, ne, gt, lt, gte, lte, like, in)
//   - value: The value to compare against
//
// Returns:
//   - *EntityQuery: The query builder for method chaining
//
// Example:
//
//	// Find entities created in the last hour
//	query.AddFilter("created_at", "gt", time.Now().Add(-1*time.Hour).UnixNano())
//	
//	// Find entities with specific IDs
//	query.AddFilter("id", "in", "id1,id2,id3")
func (q *EntityQuery) AddFilter(field, operator string, value interface{}) *EntityQuery {
	q.filters = append(q.filters, Filter{
		Field:    field,
		Operator: operator,
		Value:    value,
	})
	return q
}

// And adds an AND logical operator between the previous and next filter.
// If no operator is specified between filters, AND is assumed.
// This affects how multiple AddFilter calls are combined.
//
// Returns:
//   - *EntityQuery: The query builder for method chaining
//
// Example:
//
//	// Find users created today
//	query.AddFilter("tag:type", "eq", "user").
//	      And().
//	      AddFilter("created_at", "gt", startOfDay)
func (q *EntityQuery) And() *EntityQuery {
	q.operators = append(q.operators, "AND")
	return q
}

// Or adds an OR logical operator between the previous and next filter.
// This allows for more complex filter logic where entities can match
// either condition.
//
// Returns:
//   - *EntityQuery: The query builder for method chaining
//
// Example:
//
//	// Find entities that are either draft or pending
//	query.AddFilter("tag:status", "eq", "draft").
//	      Or().
//	      AddFilter("tag:status", "eq", "pending")
func (q *EntityQuery) Or() *EntityQuery {
	q.operators = append(q.operators, "OR")
	return q
}

// Execute runs the constructed query and returns matching entities.
// The query execution follows this order:
//   1. Initial filtering based on primary criteria (tags, content, namespace)
//   2. Additional filtering for multiple conditions
//   3. Custom field filters with AND/OR logic
//   4. Sorting by specified field and direction
//   5. Pagination with offset and limit
//
// Returns:
//   - []*Entity: Slice of entities matching all query criteria
//   - error: Any error encountered during query execution
//
// Example:
//
//	entities, err := NewEntityQuery(repo).
//	    HasTag("type:user").
//	    AddFilter("created_at", "gt", yesterday).
//	    OrderBy("created_at", "desc").
//	    Limit(10).
//	    Execute()
//	
//	if err != nil {
//	    return err
//	}
//	for _, entity := range entities {
//	    fmt.Printf("Entity: %s\n", entity.ID)
//	}
func (q *EntityQuery) Execute() ([]*Entity, error) {
	// Start with initial filtering based on primary criteria
	var entities []*Entity
	var err error
	
	// Choose the most specific filter to start with for better performance
	if len(q.tags) > 0 && len(q.tags) == 1 {
		// Single tag filter - most efficient
		entities, err = q.repo.ListByTag(q.tags[0])
	} else if len(q.tags) > 1 {
		// Multiple tags - requires all tags to match
		entities, err = q.repo.ListByTags(q.tags, true) // matchAll = true
	} else if len(q.wildcards) > 0 {
		// Wildcard search - use first wildcard pattern
		entities, err = q.repo.ListByTagWildcard(q.wildcards[0])
	} else if q.content != "" {
		// Content search
		entities, err = q.repo.SearchContent(q.content)
	} else if q.namespace != "" {
		// Namespace filter
		entities, err = q.repo.ListByNamespace(q.namespace)
	} else {
		// No primary filter - get all entities
		entities, err = q.repo.List()
	}
	
	if err != nil {
		return nil, err
	}
	
	// Apply additional filters in sequence
	filtered := entities
	
	// Apply multiple tag filters if needed
	if len(q.tags) > 1 {
		filtered = filterByTags(filtered, q.tags)
	}
	
	// Apply additional wildcard patterns
	if len(q.wildcards) > 1 {
		for _, wildcard := range q.wildcards[1:] {
			filtered = filterByWildcard(filtered, wildcard)
		}
	}
	
	// Apply custom field filters with logical operators
	if len(q.filters) > 0 {
		filtered = q.applyFilters(filtered)
	}
	
	// Apply sorting to the filtered results
	filtered = q.applySorting(filtered)
	
	// Apply pagination - offset first, then limit
	if q.offset > 0 && q.offset < len(filtered) {
		filtered = filtered[q.offset:]
	}
	
	if q.limit > 0 && q.limit < len(filtered) {
		filtered = filtered[:q.limit]
	}
	
	return filtered, nil
}

// Helper functions for query execution

// filterByTags filters entities to only those containing all specified tags.
// This implements AND logic - an entity must have every tag in the list to be included.
func filterByTags(entities []*Entity, tags []string) []*Entity {
	var result []*Entity
	for _, entity := range entities {
		hasAllTags := true
		// Check if entity has all required tags
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

// filterByWildcard filters entities to those with at least one tag matching the pattern.
// This implements OR logic - an entity needs only one matching tag to be included.
func filterByWildcard(entities []*Entity, pattern string) []*Entity {
	var result []*Entity
	for _, entity := range entities {
		// Check if any tag matches the wildcard pattern
		for _, tag := range entity.Tags {
			if matchesWildcard(tag, pattern) {
				result = append(result, entity)
				break // Only need one match per entity
			}
		}
	}
	return result
}

// matchesWildcard performs simple wildcard pattern matching.
// Supports "*" to match anything and "prefix*" to match any tag starting with prefix.
func matchesWildcard(tag, pattern string) bool {
	// "*" matches any tag
	if pattern == "*" {
		return true
	}
	
	// Handle patterns like "type:*" or "status:act*"
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(tag) >= len(prefix) && tag[:len(prefix)] == prefix
	}
	
	// No wildcard - exact match required
	return tag == pattern
}

// sortByCreatedAt sorts entities by creation time in ascending order (oldest first).
// Uses bubble sort for simplicity - suitable for small to medium result sets.
func sortByCreatedAt(entities []*Entity) {
	for i := 0; i < len(entities)-1; i++ {
		for j := 0; j < len(entities)-i-1; j++ {
			if entities[j].CreatedAt > entities[j+1].CreatedAt {
				entities[j], entities[j+1] = entities[j+1], entities[j]
			}
		}
	}
}

// sortByCreatedAtDesc sorts entities by creation time in descending order (newest first).
// Uses bubble sort for simplicity - suitable for small to medium result sets.
func sortByCreatedAtDesc(entities []*Entity) {
	for i := 0; i < len(entities)-1; i++ {
		for j := 0; j < len(entities)-i-1; j++ {
			if entities[j].CreatedAt < entities[j+1].CreatedAt {
				entities[j], entities[j+1] = entities[j+1], entities[j]
			}
		}
	}
}

// applyFilters applies custom field filters to entities with support for AND/OR logic.
// Filters are evaluated in order, with logical operators controlling how results combine.
// If no operator is specified between filters, AND is assumed.
func (q *EntityQuery) applyFilters(entities []*Entity) []*Entity {
	if len(q.filters) == 0 {
		return entities
	}

	result := make([]*Entity, 0)
	
	for _, entity := range entities {
		shouldInclude := true
		
		// Evaluate each filter and combine results based on operators
		for i, filter := range q.filters {
			matches := q.evaluateFilter(entity, filter)
			
			// Apply logical operator between this and previous filter
			if i > 0 && i-1 < len(q.operators) {
				operator := q.operators[i-1]
				if operator == "OR" {
					shouldInclude = shouldInclude || matches
				} else { // Default to AND
					shouldInclude = shouldInclude && matches
				}
			} else {
				// First filter sets initial state
				shouldInclude = matches
			}
		}
		
		if shouldInclude {
			result = append(result, entity)
		}
	}
	
	return result
}

// evaluateFilter checks if an entity matches a single filter condition.
// Supports filtering on standard fields (id, timestamps, tag count) as well as
// content and tag-based filtering with namespace support.
func (q *EntityQuery) evaluateFilter(entity *Entity, filter Filter) bool {
	switch filter.Field {
	case "created_at":
		// Filter by creation timestamp
		return q.evaluateTimeFilter(entity.CreatedAt, filter)
		
	case "updated_at":
		// Filter by last update timestamp
		return q.evaluateTimeFilter(entity.UpdatedAt, filter)
		
	case "id":
		// Filter by entity ID
		return q.evaluateStringFilter(entity.ID, filter)
		
	case "tag_count":
		// Filter by number of tags
		return q.evaluateNumericFilter(float64(len(entity.Tags)), filter)
		
	case "content_type":
		// Filter by content type tag (content:type:*)
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
		// Search within entity content (treated as string)
		if len(entity.Content) > 0 {
			contentStr := string(entity.Content)
			if q.evaluateStringFilter(contentStr, filter) {
				return true
			}
		}
		return false
		
	default:
		// Check if it's a tag namespace filter (e.g., "tag:type", "tag:status")
		if strings.HasPrefix(filter.Field, "tag:") {
			tagNamespace := strings.TrimPrefix(filter.Field, "tag:")
			// Look for tags in this namespace
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

// evaluateStringFilter evaluates string-based filter conditions.
// Supports equality, inequality, pattern matching, and set membership tests.
func (q *EntityQuery) evaluateStringFilter(value string, filter Filter) bool {
	filterValue, ok := filter.Value.(string)
	if !ok {
		return false
	}
	
	switch filter.Operator {
	case "eq":
		// Exact equality match
		return value == filterValue
		
	case "ne":
		// Not equal
		return value != filterValue
		
	case "like":
		// Case-insensitive substring match
		return strings.Contains(strings.ToLower(value), strings.ToLower(filterValue))
		
	case "in":
		// Check if value is in comma-separated list
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

// evaluateNumericFilter evaluates numeric filter conditions with type conversion.
// Supports comparison operators for numbers provided as various types.
func (q *EntityQuery) evaluateNumericFilter(value float64, filter Filter) bool {
	var filterValue float64
	
	// Convert filter value to float64 for comparison
	switch v := filter.Value.(type) {
	case float64:
		filterValue = v
	case int:
		filterValue = float64(v)
	case string:
		// Try to parse string as number
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return false
		}
		filterValue = parsed
	default:
		return false
	}
	
	// Apply numeric comparison operators
	switch filter.Operator {
	case "eq":
		// Equal to
		return value == filterValue
	case "ne":
		// Not equal to
		return value != filterValue
	case "gt":
		// Greater than
		return value > filterValue
	case "lt":
		// Less than
		return value < filterValue
	case "gte":
		// Greater than or equal to
		return value >= filterValue
	case "lte":
		// Less than or equal to
		return value <= filterValue
	default:
		return false
	}
}

// evaluateTimeNanoFilter evaluates a time-based filter for nanosecond epoch timestamps.
// This is a convenience wrapper around evaluateTimeFilter for temporal data.
func (q *EntityQuery) evaluateTimeNanoFilter(value int64, filter Filter) bool {
	return q.evaluateTimeFilter(value, filter)
}

// evaluateTimeFilter evaluates time-based filter conditions with flexible input parsing.
// Supports RFC3339 formatted strings, Unix timestamps, and numeric values.
// All times are converted to nanosecond precision for comparison.
func (q *EntityQuery) evaluateTimeFilter(value int64, filter Filter) bool {
	var filterTime int64
	
	// Convert filter value to nanosecond timestamp
	switch v := filter.Value.(type) {
	case string:
		// Try parsing as RFC3339 first (e.g., "2023-12-25T10:00:00Z")
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			// Fall back to parsing as Unix timestamp string
			timestamp, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return false
			}
			filterTime = timestamp
		} else {
			filterTime = t.UnixNano()
		}
	case int64:
		// Direct nanosecond timestamp
		filterTime = v
	case int:
		// Convert int to int64
		filterTime = int64(v)
	default:
		return false
	}
	
	// Apply temporal comparison operators
	switch filter.Operator {
	case "eq":
		// Exact timestamp match
		return value == filterTime
	case "ne":
		// Not equal timestamp
		return value != filterTime
	case "gt":
		// After the specified time
		return value > filterTime
	case "lt":
		// Before the specified time
		return value < filterTime
	case "gte":
		// At or after the specified time
		return value >= filterTime
	case "lte":
		// At or before the specified time
		return value <= filterTime
	default:
		return false
	}
}

// applySorting applies the specified sorting to entities and returns a sorted copy.
// The original slice is not modified. If no ordering is specified, entities are returned unchanged.
func (q *EntityQuery) applySorting(entities []*Entity) []*Entity {
	if q.orderBy == "" {
		return entities
	}
	
	// Create a copy to avoid modifying the original slice
	sorted := make([]*Entity, len(entities))
	copy(sorted, entities)
	
	// Apply appropriate sort function based on field and direction
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