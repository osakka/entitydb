package binary

import (
	"strings"
	"sync"
)

// NamespaceIndex provides efficient namespace queries
type NamespaceIndex struct {
	mu        sync.RWMutex
	index     map[string]map[string][]string // namespace -> tag -> entityIDs
	entityMap map[string]map[string]string   // entityID -> namespace -> latest value
}

// NewNamespaceIndex creates a new namespace index
func NewNamespaceIndex() *NamespaceIndex {
	return &NamespaceIndex{
		index:     make(map[string]map[string][]string),
		entityMap: make(map[string]map[string]string),
	}
}

// AddTag adds a tag to the namespace index
func (ni *NamespaceIndex) AddTag(entityID string, tag string) {
	ni.mu.Lock()
	defer ni.mu.Unlock()
	
	// Extract namespace and value
	namespace, value := "", ""
	
	// Remove timestamp if present
	actualTag := tag
	if idx := strings.Index(tag, "|"); idx != -1 {
		actualTag = tag[idx+1:]
	}
	
	// Extract namespace and value
	if idx := strings.Index(actualTag, ":"); idx != -1 {
		namespace = actualTag[:idx]
		value = actualTag[idx+1:]
	} else if idx := strings.Index(actualTag, "="); idx != -1 {
		namespace = actualTag[:idx]
		value = actualTag[idx+1:]
	} else {
		// Tag without namespace
		namespace = "_unnamespaced"
		value = actualTag
	}
	
	// Initialize namespace map if needed
	if ni.index[namespace] == nil {
		ni.index[namespace] = make(map[string][]string)
	}
	
	// Initialize entity namespace map if needed
	if ni.entityMap[entityID] == nil {
		ni.entityMap[entityID] = make(map[string]string)
	}
	
	// Add to index
	fullTag := namespace + ":" + value
	ni.index[namespace][fullTag] = appendUnique(ni.index[namespace][fullTag], entityID)
	
	// Update entity map with latest value for this namespace
	ni.entityMap[entityID][namespace] = value
}

// RemoveEntity removes all entries for an entity
func (ni *NamespaceIndex) RemoveEntity(entityID string) {
	ni.mu.Lock()
	defer ni.mu.Unlock()
	
	// Get all namespaces for this entity
	namespaces, exists := ni.entityMap[entityID]
	if !exists {
		return
	}
	
	// Remove from all namespace indexes
	for namespace, value := range namespaces {
		fullTag := namespace + ":" + value
		if tags, exists := ni.index[namespace]; exists {
			if entities, exists := tags[fullTag]; exists {
				filtered := make([]string, 0)
				for _, id := range entities {
					if id != entityID {
						filtered = append(filtered, id)
					}
				}
				if len(filtered) == 0 {
					delete(tags, fullTag)
				} else {
					tags[fullTag] = filtered
				}
			}
		}
	}
	
	// Remove from entity map
	delete(ni.entityMap, entityID)
}

// GetByNamespace returns all entities with tags in a namespace
func (ni *NamespaceIndex) GetByNamespace(namespace string) []string {
	ni.mu.RLock()
	defer ni.mu.RUnlock()
	
	entitySet := make(map[string]bool)
	
	// Get all tags in this namespace
	if tags, exists := ni.index[namespace]; exists {
		for _, entities := range tags {
			for _, entityID := range entities {
				entitySet[entityID] = true
			}
		}
	}
	
	// Convert set to slice
	result := make([]string, 0, len(entitySet))
	for entityID := range entitySet {
		result = append(result, entityID)
	}
	
	return result
}

// GetByNamespaceValue returns entities with a specific namespace:value
func (ni *NamespaceIndex) GetByNamespaceValue(namespace, value string) []string {
	ni.mu.RLock()
	defer ni.mu.RUnlock()
	
	fullTag := namespace + ":" + value
	
	if tags, exists := ni.index[namespace]; exists {
		if entities, exists := tags[fullTag]; exists {
			// Return a copy to avoid concurrent modification
			result := make([]string, len(entities))
			copy(result, entities)
			return result
		}
	}
	
	return []string{}
}

// GetNamespaceValues returns all values in a namespace
func (ni *NamespaceIndex) GetNamespaceValues(namespace string) []string {
	ni.mu.RLock()
	defer ni.mu.RUnlock()
	
	values := make([]string, 0)
	
	if tags, exists := ni.index[namespace]; exists {
		for tag := range tags {
			// Extract value from tag
			if idx := strings.Index(tag, ":"); idx != -1 {
				values = append(values, tag[idx+1:])
			}
		}
	}
	
	return values
}

// Helper function to append unique values
func appendUnique(slice []string, value string) []string {
	for _, v := range slice {
		if v == value {
			return slice
		}
	}
	return append(slice, value)
}