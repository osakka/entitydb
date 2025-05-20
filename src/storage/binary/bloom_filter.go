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

// NewBloomFilter creates a new bloom filter
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

// Add adds an item to the bloom filter
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