// Package binary provides thread-safe header synchronization
//
// This critical fix prevents header corruption during concurrent writes
// by ensuring all header modifications are properly synchronized.
package binary

import (
	"entitydb/logger"
	"fmt"
	"sync"
	"sync/atomic"
)

// HeaderSync provides thread-safe access to the file header
// preventing corruption during concurrent operations
type HeaderSync struct {
	mu     sync.RWMutex
	header Header
	
	// Atomic counters for fast path
	walSequence atomic.Uint64
	entityCount atomic.Uint64
}

// NewHeaderSync creates a new synchronized header wrapper
func NewHeaderSync(h *Header) *HeaderSync {
	hs := &HeaderSync{
		header: *h, // Copy the header
	}
	hs.walSequence.Store(h.WALSequence)
	hs.entityCount.Store(h.EntityCount)
	return hs
}

// GetHeader returns a copy of the header for safe reading
func (hs *HeaderSync) GetHeader() Header {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	
	// Update atomic counters
	h := hs.header
	h.WALSequence = hs.walSequence.Load()
	h.EntityCount = hs.entityCount.Load()
	return h
}

// UpdateHeader atomically updates the entire header
func (hs *HeaderSync) UpdateHeader(fn func(*Header)) {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	
	fn(&hs.header)
	
	// Update atomic counters
	hs.walSequence.Store(hs.header.WALSequence)
	hs.entityCount.Store(hs.header.EntityCount)
}

// IncrementWALSequence atomically increments the WAL sequence
func (hs *HeaderSync) IncrementWALSequence() uint64 {
	return hs.walSequence.Add(1)
}

// IncrementEntityCount atomically increments the entity count
func (hs *HeaderSync) IncrementEntityCount() uint64 {
	return hs.entityCount.Add(1)
}

// GetWALOffset safely returns the WAL offset with validation
func (hs *HeaderSync) GetWALOffset() (uint64, error) {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	
	offset := hs.header.WALOffset
	
	// Validate offset to prevent corruption
	if offset == 0 || offset > uint64(1<<31) {
		logger.Error("CORRUPTION DETECTED: Invalid WALOffset %d", offset)
		return 0, fmt.Errorf("corrupted header: invalid WALOffset %d", offset)
	}
	
	return offset, nil
}

// UpdateOffsets safely updates file offsets
func (hs *HeaderSync) UpdateOffsets(tagDictOffset, entityIndexOffset uint64) {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	
	hs.header.TagDictOffset = tagDictOffset
	hs.header.EntityIndexOffset = entityIndexOffset
}