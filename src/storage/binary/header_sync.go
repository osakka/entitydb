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

// GetTagDictOffset safely returns the tag dictionary offset with validation
func (hs *HeaderSync) GetTagDictOffset() (uint64, error) {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	
	offset := hs.header.TagDictOffset
	
	// Validate offset to prevent EOF errors
	if offset > uint64(1<<31) {
		logger.Error("CORRUPTION DETECTED: Invalid TagDictOffset %d", offset)
		return 0, fmt.Errorf("corrupted header: invalid TagDictOffset %d", offset)
	}
	
	return offset, nil
}

// GetEntityIndexOffset safely returns the entity index offset with validation
func (hs *HeaderSync) GetEntityIndexOffset() (uint64, error) {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	
	offset := hs.header.EntityIndexOffset
	
	// Validate offset to prevent corruption
	if offset > uint64(1<<31) {
		logger.Error("CORRUPTION DETECTED: Invalid EntityIndexOffset %d", offset)
		return 0, fmt.Errorf("corrupted header: invalid EntityIndexOffset %d", offset)
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

// HeaderSnapshot represents a safe snapshot of HeaderSync state
// for preservation during checkpoint operations
type HeaderSnapshot struct {
	Header      Header
	WALSequence uint64
	EntityCount uint64
}

// CreateSnapshot safely captures the current HeaderSync state
// for preservation during checkpoint operations
func (hs *HeaderSync) CreateSnapshot() *HeaderSnapshot {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	
	return &HeaderSnapshot{
		Header:      hs.header,
		WALSequence: hs.walSequence.Load(),
		EntityCount: hs.entityCount.Load(),
	}
}

// RestoreFromSnapshot safely restores HeaderSync state from a snapshot
// Used to recover from checkpoint corruption
func (hs *HeaderSync) RestoreFromSnapshot(snapshot *HeaderSnapshot) {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	
	hs.header = snapshot.Header
	hs.walSequence.Store(snapshot.WALSequence)
	hs.entityCount.Store(snapshot.EntityCount)
}

// ValidateHeader checks if the header has valid values
// Returns true if header is valid, false if corrupted
func (hs *HeaderSync) ValidateHeader() bool {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	
	// Check critical fields for corruption
	if hs.header.WALOffset == 0 || hs.header.WALOffset > uint64(1<<31) {
		logger.Error("HeaderSync validation failed: Invalid WALOffset %d", hs.header.WALOffset)
		return false
	}
	
	// EXTENDED PROTECTION: Validate TagDictOffset for EOF error prevention
	if hs.header.TagDictOffset > uint64(1<<31) {
		logger.Error("HeaderSync validation failed: Invalid TagDictOffset %d", hs.header.TagDictOffset)
		return false
	}
	
	// EXTENDED PROTECTION: Validate EntityIndexOffset for corruption prevention
	if hs.header.EntityIndexOffset > uint64(1<<31) {
		logger.Error("HeaderSync validation failed: Invalid EntityIndexOffset %d", hs.header.EntityIndexOffset)
		return false
	}
	
	if hs.header.Magic != 0x46465545 { // "EUFF"
		logger.Error("HeaderSync validation failed: Invalid magic number %x", hs.header.Magic)
		return false
	}
	
	if hs.header.Version == 0 {
		logger.Error("HeaderSync validation failed: Invalid version %d", hs.header.Version)
		return false
	}
	
	return true
}