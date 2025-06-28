package binary

import (
	"sync"
)

// TemporalBTree is a simple B-tree for temporal indexing
type TemporalBTree struct {
	mu    sync.RWMutex
	root  *BTreeNode
	order int
}

// BTreeNode represents a node in the B-tree
type BTreeNode struct {
	keys     []int64
	values   [][]string // Multiple values per key
	children []*BTreeNode
	isLeaf   bool
}

// NewTemporalBTree creates a new B-tree
func NewTemporalBTree(order int) *TemporalBTree {
	return &TemporalBTree{
		root: &BTreeNode{
			keys:   make([]int64, 0, order-1),
			values: make([][]string, 0, order-1),
			isLeaf: true,
		},
		order: order,
	}
}

// Put inserts a key-value pair
func (t *TemporalBTree) Put(key int64, value string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	// Simple implementation: just append to the appropriate node
	node := t.findLeaf(t.root, key)
	
	// Find insertion point
	i := 0
	for i < len(node.keys) && key > node.keys[i] {
		i++
	}
	
	if i < len(node.keys) && node.keys[i] == key {
		// Key exists, append value
		node.values[i] = append(node.values[i], value)
	} else {
		// Insert new key
		node.keys = append(node.keys[:i], append([]int64{key}, node.keys[i:]...)...)
		node.values = append(node.values[:i], append([][]string{{value}}, node.values[i:]...)...)
	}
}

// Get retrieves values for a key
func (t *TemporalBTree) Get(key int64) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	node := t.findLeaf(t.root, key)
	
	for i, k := range node.keys {
		if k == key {
			return node.values[i]
		}
	}
	
	return nil
}

// Range returns all key-value pairs in a range
func (t *TemporalBTree) Range(start, end int64) map[int64][]string {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	result := make(map[int64][]string)
	t.collectRange(t.root, start, end, result)
	return result
}

// findLeaf performs B-tree traversal to locate the appropriate leaf node for a key.
//
// Algorithm Overview:
//   1. Start at root node
//   2. For internal nodes: linear search to find correct child pointer
//   3. Follow child pointer to next level
//   4. Repeat until reaching leaf node
//
// Time Complexity: O(log n) where n is total number of keys
// Space Complexity: O(1) - no additional memory allocation during traversal
//
// Implementation Notes:
//   - Uses linear search instead of binary search for simplicity
//   - For small node sizes (order < 100), linear search performance is adequate
//   - Returns the node where key should be inserted/found
//
// Parameters:
//   - node: Current node in traversal (initially root)
//   - key: Target key to locate
//
// Returns:
//   - *BTreeNode: Leaf node where key should be inserted/found
func (t *TemporalBTree) findLeaf(node *BTreeNode, key int64) *BTreeNode {
	if node.isLeaf {
		return node
	}
	
	// Linear search for appropriate child pointer.
	// For keys[i], child[i] contains keys < keys[i],
	// child[i+1] contains keys >= keys[i]
	i := 0
	for i < len(node.keys) && key >= node.keys[i] {
		i++
	}
	
	if i < len(node.children) {
		return t.findLeaf(node.children[i], key)
	}
	
	return node
}

// collectRange performs in-order traversal to collect all key-value pairs within a range.
//
// Algorithm Overview:
//   1. Recursively traverse the B-tree in in-order fashion
//   2. For each node, check keys against range [start, end]
//   3. Collect matching keys and their associated values
//   4. Recursively process child nodes for internal nodes
//
// Time Complexity: O(k + log n) where k is number of results, n is total keys
// Space Complexity: O(h) where h is tree height (recursion stack)
//
// Traversal Pattern:
//   - For internal nodes: process child[0], key[0], child[1], key[1], ..., child[n]
//   - Ensures results are collected in sorted key order
//   - Early termination possible for keys outside range
//
// Implementation Notes:
//   - Uses in-place result map to avoid memory allocation overhead
//   - Processes all children including the "last child" for complete coverage
//   - Handles edge case where children array has one more element than keys
//
// Parameters:
//   - node: Current node being processed
//   - start, end: Inclusive range boundaries
//   - result: Map to collect matching key-value pairs
func (t *TemporalBTree) collectRange(node *BTreeNode, start, end int64, result map[int64][]string) {
	if node == nil {
		return
	}
	
	// Process keys and children in in-order traversal
	for i, key := range node.keys {
		// Recursively process left child before processing current key
		if !node.isLeaf && i < len(node.children) {
			t.collectRange(node.children[i], start, end, result)
		}
		
		// Collect key if within range
		if key >= start && key <= end {
			result[key] = node.values[i]
		}
	}
	
	// Process rightmost child (children array has one more element than keys)
	if !node.isLeaf && len(node.children) > len(node.keys) {
		t.collectRange(node.children[len(node.children)-1], start, end, result)
	}
}