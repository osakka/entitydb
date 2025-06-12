// Package models provides core data structures and business logic for EntityDB.
// This file implements tag namespace utilities for hierarchical tag management.
package models

import (
	"strings"
)

// TagHierarchy represents a hierarchical tag structure parsed from a namespaced tag.
// EntityDB uses hierarchical tags with colon separators to organize data semantically.
// Tags follow the pattern: namespace:path1:path2:value
//
// Examples:
//   - "type:user" → Namespace: "type", Path: [], Value: "user"
//   - "rbac:perm:entity:create" → Namespace: "rbac", Path: ["perm", "entity"], Value: "create"
//   - "id:user:admin" → Namespace: "id", Path: ["user"], Value: "admin"
//
// This structure enables:
//   - Namespace-based organization
//   - Hierarchical permissions (rbac:perm:*)
//   - Structured entity typing
//   - Efficient tag filtering and searching
type TagHierarchy struct {
	// Namespace is the top-level category for the tag.
	// Common namespaces: "type", "rbac", "id", "status", "conf"
	Namespace string
	
	// Path contains the intermediate hierarchy levels between namespace and value.
	// For "rbac:perm:entity:create", Path would be ["perm", "entity"]
	Path []string
	
	// Value is the final component of the tag hierarchy.
	// For "rbac:perm:entity:create", Value would be "create"
	Value string
}

// ParseTag parses a hierarchical tag string into its component parts.
// Handles both regular tags and temporal tags (TIMESTAMP|tag format).
// Returns nil if the tag doesn't follow the expected hierarchical format.
//
// Parameters:
//   - tag: The tag string to parse (may include temporal prefix)
//
// Returns:
//   - *TagHierarchy: Parsed tag structure, or nil if invalid format
//
// Examples:
//
//	// Regular tag parsing
//	parsed := ParseTag("rbac:perm:entity:create")
//	// parsed.Namespace = "rbac"
//	// parsed.Path = ["perm", "entity"]
//	// parsed.Value = "create"
//	
//	// Temporal tag parsing
//	parsed := ParseTag("1640995200000000000|type:user")
//	// parsed.Namespace = "type"
//	// parsed.Path = []
//	// parsed.Value = "user"
//	
//	// Invalid format
//	parsed := ParseTag("invalidtag")
//	// parsed = nil
func ParseTag(tag string) *TagHierarchy {
	// Handle temporal tags (TIMESTAMP|tag format)
	actualTag := tag
	if parts := strings.SplitN(tag, "|", 2); len(parts) == 2 {
		actualTag = parts[1]
	}
	
	// Split by colon separator
	parts := strings.Split(actualTag, ":")
	if len(parts) < 2 {
		// Invalid format - need at least namespace:value
		return nil
	}
	
	return &TagHierarchy{
		Namespace: parts[0],
		Path:      parts[1 : len(parts)-1],
		Value:     parts[len(parts)-1],
	}
}

// IsNamespace checks if a tag belongs to a specific namespace.
// Handles both regular and temporal tags transparently.
//
// Parameters:
//   - tag: The tag to check (may include temporal prefix)
//   - namespace: The namespace to test membership for
//
// Returns:
//   - bool: true if the tag belongs to the specified namespace
//
// Examples:
//
//	IsNamespace("type:user", "type")                    // true
//	IsNamespace("rbac:perm:entity:create", "rbac")      // true
//	IsNamespace("1640995200000000000|type:user", "type") // true
//	IsNamespace("status:active", "type")                // false
func IsNamespace(tag, namespace string) bool {
	// Handle temporal tags (TIMESTAMP|namespace:value format)
	actualTag := tag
	if parts := strings.SplitN(tag, "|", 2); len(parts) == 2 {
		actualTag = parts[1]
	}
	return strings.HasPrefix(actualTag, namespace+":")
}

// HasPermission checks if a set of tags includes a specific RBAC permission.
// Supports hierarchical wildcard matching for flexible permission inheritance.
// Wildcards can appear at any level in the permission hierarchy.
//
// Permission matching rules:
//   - Exact match: "rbac:perm:entity:create" matches "rbac:perm:entity:create"
//   - Full wildcard: "rbac:perm:*" matches any permission
//   - Partial wildcard: "rbac:perm:entity:*" matches any entity permission
//
// Parameters:
//   - tags: The collection of tags to search through
//   - requiredPerm: The permission to check for (must be in rbac namespace)
//
// Returns:
//   - bool: true if the permission is granted by the tag set
//
// Examples:
//
//	tags := []string{"rbac:perm:*"}
//	HasPermission(tags, "rbac:perm:entity:create")  // true (wildcard match)
//	
//	tags = []string{"rbac:perm:entity:*"}
//	HasPermission(tags, "rbac:perm:entity:create")  // true (partial wildcard)
//	HasPermission(tags, "rbac:perm:user:create")    // false (different category)
//	
//	tags = []string{"rbac:perm:entity:create"}
//	HasPermission(tags, "rbac:perm:entity:create")  // true (exact match)
//	HasPermission(tags, "rbac:perm:entity:delete")  // false (no match)
func HasPermission(tags []string, requiredPerm string) bool {
	// First check for exact permission match
	for _, tag := range tags {
		// Handle temporal tags by extracting actual tag
		actualTag := tag
		if parts := strings.SplitN(tag, "|", 2); len(parts) == 2 {
			actualTag = parts[1]
		}
		if actualTag == requiredPerm {
			return true
		}
	}
	
	// Parse the required permission for wildcard checking
	reqParsed := ParseTag(requiredPerm)
	if reqParsed == nil || reqParsed.Namespace != "rbac" {
		// Only RBAC permissions are supported
		return false
	}
	
	// Check for wildcard permissions that would grant the required permission
	for _, tag := range tags {
		// Only check RBAC tags
		if !IsNamespace(tag, "rbac") {
			continue
		}
		
		parsed := ParseTag(tag)
		if parsed == nil {
			continue
		}
		
		// Check for full wildcard (rbac:perm:*)
		if len(parsed.Path) == 1 && parsed.Path[0] == "perm" && parsed.Value == "*" {
			return true
		}
		
		// Check for partial wildcards (e.g., rbac:perm:entity:*)
		if len(parsed.Path) >= len(reqParsed.Path) {
			// Verify that the path matches up to the wildcard
			match := true
			for i, part := range reqParsed.Path {
				if parsed.Path[i] != part {
					match = false
					break
				}
			}
			
			// If path matches and this tag has a wildcard, permission is granted
			if match && parsed.Value == "*" {
				return true
			}
		}
	}
	
	return false
}

// GetTagsByNamespace filters a collection of tags to only those in a specific namespace.
// Returns the actual tag values (without temporal prefixes) for easier processing.
// This is useful for extracting all tags of a certain type from an entity.
//
// Parameters:
//   - tags: The collection of tags to filter
//   - namespace: The namespace to filter by
//
// Returns:
//   - []string: All tags belonging to the specified namespace
//
// Examples:
//
//	tags := []string{
//	    "type:user",
//	    "rbac:perm:entity:create",
//	    "rbac:role:admin",
//	    "status:active",
//	    "1640995200000000000|rbac:perm:user:view",
//	}
//	
//	rbacTags := GetTagsByNamespace(tags, "rbac")
//	// Result: ["rbac:perm:entity:create", "rbac:role:admin", "rbac:perm:user:view"]
//	
//	typeTags := GetTagsByNamespace(tags, "type")
//	// Result: ["type:user"]
func GetTagsByNamespace(tags []string, namespace string) []string {
	var result []string
	prefix := namespace + ":"
	
	for _, tag := range tags {
		// Handle temporal tags by extracting the actual tag
		actualTag := tag
		if parts := strings.SplitN(tag, "|", 2); len(parts) == 2 {
			actualTag = parts[1]
		}
		
		if strings.HasPrefix(actualTag, prefix) {
			result = append(result, actualTag)
		}
	}
	
	return result
}

// GetTagValue extracts the final value component from a hierarchical tag.
// This is the rightmost part after the last colon separator.
//
// Parameters:
//   - tag: The hierarchical tag to extract the value from
//
// Returns:
//   - string: The value component, or empty string if invalid format
//
// Examples:
//
//	GetTagValue("type:user")                 // "user"
//	GetTagValue("rbac:perm:entity:create")   // "create"
//	GetTagValue("id:username:admin")         // "admin"
//	GetTagValue("status:active")             // "active"
//	GetTagValue("invalidtag")                // ""
func GetTagValue(tag string) string {
	parts := strings.Split(tag, ":")
	if len(parts) < 2 {
		return ""
	}
	return parts[len(parts)-1]
}

// GetTagPath returns the hierarchical path without the final value component.
// This includes the namespace and all intermediate path elements.
//
// Parameters:
//   - tag: The hierarchical tag to extract the path from
//
// Returns:
//   - string: The path without the value, or empty string if invalid format
//
// Examples:
//
//	GetTagPath("type:user")                 // "type"
//	GetTagPath("rbac:perm:entity:create")   // "rbac:perm:entity"
//	GetTagPath("id:username:admin")         // "id:username"
//	GetTagPath("status:active")             // "status"
//	GetTagPath("invalidtag")                // ""
//
// This is useful for:
//   - Grouping tags by their hierarchical path
//   - Building wildcard patterns
//   - Analyzing tag structure and organization
func GetTagPath(tag string) string {
	parts := strings.Split(tag, ":")
	if len(parts) < 2 {
		return ""
	}
	return strings.Join(parts[:len(parts)-1], ":")
}