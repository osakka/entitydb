package models

import (
	"strings"
)

// TagHierarchy represents a hierarchical tag structure
type TagHierarchy struct {
	Namespace string   // First level (e.g., "rbac", "type", "id")
	Path      []string // Subsequent levels
	Value     string   // Final value
}

// ParseTag parses a hierarchical tag into its components
func ParseTag(tag string) *TagHierarchy {
	// Handle temporal tags (TIMESTAMP|tag)
	actualTag := tag
	if parts := strings.SplitN(tag, "|", 2); len(parts) == 2 {
		actualTag = parts[1]
	}
	
	parts := strings.Split(actualTag, ":")
	if len(parts) < 2 {
		return nil
	}
	
	return &TagHierarchy{
		Namespace: parts[0],
		Path:      parts[1 : len(parts)-1],
		Value:     parts[len(parts)-1],
	}
}

// IsNamespace checks if a tag belongs to a specific namespace
func IsNamespace(tag, namespace string) bool {
	// Handle temporal tags (TIMESTAMP|namespace:value)
	actualTag := tag
	if parts := strings.SplitN(tag, "|", 2); len(parts) == 2 {
		actualTag = parts[1]
	}
	return strings.HasPrefix(actualTag, namespace+":")
}

// HasPermission checks if a set of tags includes a specific permission
// Supports wildcards at any level (e.g., rbac:perm:* or rbac:perm:entity:*)
func HasPermission(tags []string, requiredPerm string) bool {
	// Check for exact match
	for _, tag := range tags {
		// Handle temporal tags
		actualTag := tag
		if parts := strings.SplitN(tag, "|", 2); len(parts) == 2 {
			actualTag = parts[1]
		}
		if actualTag == requiredPerm {
			return true
		}
	}
	
	// Parse the required permission
	reqParsed := ParseTag(requiredPerm)
	if reqParsed == nil || reqParsed.Namespace != "rbac" {
		return false
	}
	
	// Check for wildcard permissions
	for _, tag := range tags {
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
			match := true
			for i, part := range reqParsed.Path {
				if parsed.Path[i] != part {
					match = false
					break
				}
			}
			
			if match && parsed.Value == "*" {
				return true
			}
		}
	}
	
	return false
}

// GetTagsByNamespace returns all tags in a specific namespace
func GetTagsByNamespace(tags []string, namespace string) []string {
	var result []string
	prefix := namespace + ":"
	
	for _, tag := range tags {
		// Handle temporal tags (TIMESTAMP|namespace:value)
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

// GetTagValue extracts the value from a namespaced tag
// For example: "id:username:admin" returns "admin"
func GetTagValue(tag string) string {
	parts := strings.Split(tag, ":")
	if len(parts) < 2 {
		return ""
	}
	return parts[len(parts)-1]
}

// GetTagPath returns the full path without the value
// For example: "rbac:perm:entity:create" returns "rbac:perm:entity"
func GetTagPath(tag string) string {
	parts := strings.Split(tag, ":")
	if len(parts) < 2 {
		return ""
	}
	return strings.Join(parts[:len(parts)-1], ":")
}