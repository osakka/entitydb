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

func (t *TemporalBTree) findLeaf(node *BTreeNode, key int64) *BTreeNode {
	if node.isLeaf {
		return node
	}
	
	// Binary search for child
	i := 0
	for i < len(node.keys) && key >= node.keys[i] {
		i++
	}
	
	if i < len(node.children) {
		return t.findLeaf(node.children[i], key)
	}
	
	return node
}

func (t *TemporalBTree) collectRange(node *BTreeNode, start, end int64, result map[int64][]string) {
	if node == nil {
		return
	}
	
	for i, key := range node.keys {
		if key >= start && key <= end {
			result[key] = node.values[i]
		}
		
		if !node.isLeaf && i < len(node.children) {
			t.collectRange(node.children[i], start, end, result)
		}
	}
	
	// Check last child
	if !node.isLeaf && len(node.children) > len(node.keys) {
		t.collectRange(node.children[len(node.children)-1], start, end, result)
	}
}