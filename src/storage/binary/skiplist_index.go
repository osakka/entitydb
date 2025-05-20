package binary

import (
	"math/rand"
	"sync"
	"time"
)

const (
	maxLevel = 16
	p        = 0.5
)

// SkipListNode represents a node in the skip list
type SkipListNode struct {
	key     string
	value   []string // entity IDs
	forward []*SkipListNode
}

// SkipList provides O(log n) lookups with cache-friendly access patterns
type SkipList struct {
	header *SkipListNode
	level  int
	mu     sync.RWMutex
	rng    *rand.Rand
}

// NewSkipList creates a new skip list index
func NewSkipList() *SkipList {
	header := &SkipListNode{
		forward: make([]*SkipListNode, maxLevel),
	}
	
	return &SkipList{
		header: header,
		level:  0,
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// randomLevel generates a random level for new nodes
func (sl *SkipList) randomLevel() int {
	level := 0
	for level < maxLevel-1 && sl.rng.Float64() < p {
		level++
	}
	return level
}

// Insert adds a key-value pair to the skip list
func (sl *SkipList) Insert(key string, entityID string) {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	
	update := make([]*SkipListNode, maxLevel)
	current := sl.header
	
	// Find insert position
	for i := sl.level; i >= 0; i-- {
		for current.forward[i] != nil && current.forward[i].key < key {
			current = current.forward[i]
		}
		update[i] = current
	}
	
	current = current.forward[0]
	
	// Update existing node or create new one
	if current != nil && current.key == key {
		// Append to existing values
		current.value = append(current.value, entityID)
	} else {
		// Create new node
		newLevel := sl.randomLevel()
		newNode := &SkipListNode{
			key:     key,
			value:   []string{entityID},
			forward: make([]*SkipListNode, newLevel+1),
		}
		
		// Update level if necessary
		if newLevel > sl.level {
			for i := sl.level + 1; i <= newLevel; i++ {
				update[i] = sl.header
			}
			sl.level = newLevel
		}
		
		// Insert node
		for i := 0; i <= newLevel; i++ {
			newNode.forward[i] = update[i].forward[i]
			update[i].forward[i] = newNode
		}
	}
}

// Search finds all entity IDs for a given key
func (sl *SkipList) Search(key string) []string {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	
	current := sl.header
	
	// Find node
	for i := sl.level; i >= 0; i-- {
		for current.forward[i] != nil && current.forward[i].key < key {
			current = current.forward[i]
		}
	}
	
	current = current.forward[0]
	
	if current != nil && current.key == key {
		// Return copy to avoid concurrent modification
		result := make([]string, len(current.value))
		copy(result, current.value)
		return result
	}
	
	return nil
}

// RangeSearch finds all values in a key range
func (sl *SkipList) RangeSearch(startKey, endKey string) map[string][]string {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	
	result := make(map[string][]string)
	current := sl.header
	
	// Find start position
	for i := sl.level; i >= 0; i-- {
		for current.forward[i] != nil && current.forward[i].key < startKey {
			current = current.forward[i]
		}
	}
	
	current = current.forward[0]
	
	// Collect all nodes in range
	for current != nil && current.key <= endKey {
		if current.key >= startKey {
			result[current.key] = make([]string, len(current.value))
			copy(result[current.key], current.value)
		}
		current = current.forward[0]
	}
	
	return result
}

// Delete removes an entity ID from a key
func (sl *SkipList) Delete(key string, entityID string) {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	
	update := make([]*SkipListNode, maxLevel)
	current := sl.header
	
	// Find node
	for i := sl.level; i >= 0; i-- {
		for current.forward[i] != nil && current.forward[i].key < key {
			current = current.forward[i]
		}
		update[i] = current
	}
	
	current = current.forward[0]
	
	if current != nil && current.key == key {
		// Remove entity ID from values
		newValues := make([]string, 0, len(current.value))
		for _, v := range current.value {
			if v != entityID {
				newValues = append(newValues, v)
			}
		}
		
		if len(newValues) == 0 {
			// Remove node if no values left
			for i := 0; i <= sl.level; i++ {
				if update[i].forward[i] != current {
					break
				}
				update[i].forward[i] = current.forward[i]
			}
			
			// Update level if necessary
			for sl.level > 0 && sl.header.forward[sl.level] == nil {
				sl.level--
			}
		} else {
			current.value = newValues
		}
	}
}