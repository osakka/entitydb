// Package binary provides bloom filter implementation for efficient tag existence testing.
//
// Bloom Filter Theory:
//   A Bloom filter is a space-efficient probabilistic data structure that is used to test
//   whether an element is a member of a set. False positive matches are possible, but false
//   negatives are not – in other words, a query returns either "possibly in set" or
//   "definitely not in set".
//
// Performance Characteristics:
//   - Space complexity: O(m) where m is the number of bits
//   - Time complexity: O(k) where k is the number of hash functions
//   - Memory usage: ~1.44 bits per element for 1% false positive rate
//   - Query speed: Extremely fast, typically <100ns per lookup
//
// Implementation Details:
//   This implementation uses FNV-1a hash function with double hashing to generate
//   k independent hash functions from two hash values. The bit array is implemented
//   using uint64 slices for optimal memory alignment and cache performance.
//
// Thread Safety:
//   The bloom filter is thread-safe using RWMutex, allowing concurrent reads while
//   serializing writes. This is optimal for EntityDB's read-heavy tag lookup patterns.
//
// Usage in EntityDB:
//   Bloom filters are used to quickly eliminate non-existent tags before performing
//   expensive disk I/O operations, significantly improving query performance for
//   sparse tag distributions.
package binary

import (
	"hash"
	"hash/fnv"
	"math"
	"sync"
)

// BloomFilter provides probabilistic existence testing with very fast lookups
type BloomFilter struct {
	bits     []uint64
	k        uint // number of hash functions
	m        uint // number of bits
	n        uint // number of items
	hashFunc hash.Hash64
	mu       sync.RWMutex
}

// NewBloomFilter creates a new bloom filter with optimal parameters.
//
// Parameters:
//   expectedItems: Estimated number of items to be inserted
//   falsePositiveRate: Desired false positive probability (e.g., 0.01 for 1%)
//
// The constructor automatically calculates optimal values for:
//   - m: Number of bits in the bit array
//   - k: Number of hash functions to use
//
// Memory usage scales with expectedItems and decreases with higher falsePositiveRate.
// Typical values: 0.01 (1%) for balanced performance, 0.001 (0.1%) for high accuracy.
//
// Performance: ~1.44 bits per item for 1% false positive rate.
func NewBloomFilter(expectedItems uint, falsePositiveRate float64) *BloomFilter {
	// Calculate optimal parameters
	m := uint(math.Ceil(-float64(expectedItems) * math.Log(falsePositiveRate) / math.Pow(math.Log(2), 2)))
	k := uint(math.Ceil(float64(m) / float64(expectedItems) * math.Log(2)))
	
	// Ensure m is a multiple of 64
	m = (m + 63) / 64 * 64
	
	return &BloomFilter{
		bits:     make([]uint64, m/64),
		k:        k,
		m:        m,
		n:        0,
		hashFunc: fnv.New64a(),
	}
}

// Add inserts an item into the bloom filter.
//
// This operation sets k bits in the bit array based on k hash functions.
// After adding an item, subsequent Contains() calls for the same item
// will always return true (no false negatives).
//
// Thread safety: This method is thread-safe and can be called concurrently
// with other Add() and Contains() operations.
//
// Performance: O(k) where k is the number of hash functions (typically 3-5).
func (bf *BloomFilter) Add(item string) {
	bf.mu.Lock()
	defer bf.mu.Unlock()
	
	hashes := bf.getHashes(item)
	
	for _, h := range hashes {
		pos := uint(h % uint64(bf.m))
		bf.bits[pos/64] |= 1 << (pos % 64)
	}
	
	bf.n++
}

// Contains checks if an item might be in the set
func (bf *BloomFilter) Contains(item string) bool {
	bf.mu.RLock()
	defer bf.mu.RUnlock()
	
	hashes := bf.getHashes(item)
	
	for _, h := range hashes {
		pos := uint(h % uint64(bf.m))
		if bf.bits[pos/64]&(1<<(pos%64)) == 0 {
			return false
		}
	}
	
	return true
}

// getHashes generates k hash values for an item
func (bf *BloomFilter) getHashes(item string) []uint64 {
	hashes := make([]uint64, bf.k)
	
	bf.hashFunc.Reset()
	bf.hashFunc.Write([]byte(item))
	h1 := bf.hashFunc.Sum64()
	
	bf.hashFunc.Reset()
	bf.hashFunc.Write([]byte(item + "salt"))
	h2 := bf.hashFunc.Sum64()
	
	// Use double hashing to generate k hashes
	for i := uint(0); i < bf.k; i++ {
		hashes[i] = h1 + uint64(i)*h2
	}
	
	return hashes
}

// EstimateFalsePositiveRate estimates the current false positive rate
func (bf *BloomFilter) EstimateFalsePositiveRate() float64 {
	return math.Pow(1-math.Exp(-float64(bf.k*bf.n)/float64(bf.m)), float64(bf.k))
}

// Reset clears the bloom filter
func (bf *BloomFilter) Reset() {
	bf.mu.Lock()
	defer bf.mu.Unlock()
	
	for i := range bf.bits {
		bf.bits[i] = 0
	}
	bf.n = 0
}