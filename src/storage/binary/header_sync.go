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
	// entityCount removed - index is single source of truth
}

// NewHeaderSync creates a new synchronized header wrapper
func NewHeaderSync(h *Header) *HeaderSync {
	hs := &HeaderSync{
		header: *h, // Copy the header
	}
	hs.walSequence.Store(h.WALSequence)
	// entityCount initialization removed - index is single source of truth
	return hs
}

// GetHeader returns a copy of the header for safe reading
func (hs *HeaderSync) GetHeader() Header {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	
	// Update atomic counters
	h := hs.header
	h.WALSequence = hs.walSequence.Load()
	// EntityCount comes from header directly - index is single source of truth
	return h
}

// UpdateHeader atomically updates the entire header
func (hs *HeaderSync) UpdateHeader(fn func(*Header)) {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	
	fn(&hs.header)
	
	// Update atomic counters
	hs.walSequence.Store(hs.header.WALSequence)
	// entityCount sync removed - index is single source of truth
}

// IncrementWALSequence atomically increments the WAL sequence
func (hs *HeaderSync) IncrementWALSequence() uint64 {
	return hs.walSequence.Add(1)
}

// IncrementEntityCount removed - index is single source of truth for entity count

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
	fileSize := hs.header.FileSize
	
	// Enhanced validation: Offset must be reasonable relative to file size
	if offset == 0 {
		logger.Error("CORRUPTION DETECTED: TagDictOffset cannot be zero")
		return 0, fmt.Errorf("corrupted header: invalid TagDictOffset %d", offset)
	}
	
	// Check against file size for sanity
	if fileSize > 0 && offset >= fileSize {
		logger.Error("CORRUPTION DETECTED: TagDictOffset %d exceeds file size %d", offset, fileSize)
		return 0, fmt.Errorf("corrupted header: invalid TagDictOffset %d exceeds file size %d", offset, fileSize)
	}
	
	// Prevent astronomical offsets (10GB limit for safety)
	maxReasonableOffset := uint64(10 * 1024 * 1024 * 1024) // 10GB
	if offset > maxReasonableOffset {
		logger.Error("CORRUPTION DETECTED: TagDictOffset %d exceeds reasonable limit %d", offset, maxReasonableOffset)
		return 0, fmt.Errorf("corrupted header: invalid TagDictOffset %d", offset)
	}
	
	return offset, nil
}

// GetEntityIndexOffset safely returns the entity index offset with validation
func (hs *HeaderSync) GetEntityIndexOffset() (uint64, error) {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	
	offset := hs.header.EntityIndexOffset
	fileSize := hs.header.FileSize
	
	// Enhanced validation: Offset must be reasonable relative to file size
	if offset == 0 {
		logger.Error("CORRUPTION DETECTED: EntityIndexOffset cannot be zero")
		return 0, fmt.Errorf("corrupted header: invalid EntityIndexOffset %d", offset)
	}
	
	// Check against file size for sanity
	if fileSize > 0 && offset >= fileSize {
		logger.Error("CORRUPTION DETECTED: EntityIndexOffset %d exceeds file size %d", offset, fileSize)
		return 0, fmt.Errorf("corrupted header: invalid EntityIndexOffset %d exceeds file size %d", offset, fileSize)
	}
	
	// Prevent astronomical offsets (10GB limit for safety)
	maxReasonableOffset := uint64(10 * 1024 * 1024 * 1024) // 10GB
	if offset > maxReasonableOffset {
		logger.Error("CORRUPTION DETECTED: EntityIndexOffset %d exceeds reasonable limit %d", offset, maxReasonableOffset)
		return 0, fmt.Errorf("corrupted header: invalid EntityIndexOffset %d", offset)
	}
	
	return offset, nil
}

// UpdateOffsets safely updates file offsets with corruption prevention
func (hs *HeaderSync) UpdateOffsets(tagDictOffset, entityIndexOffset uint64) error {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	
	fileSize := hs.header.FileSize
	maxReasonableOffset := uint64(10 * 1024 * 1024 * 1024) // 10GB safety limit
	
	// Validate TagDictOffset before updating
	if tagDictOffset == 0 {
		logger.Error("CORRUPTION PREVENTION: Refusing to set TagDictOffset to zero")
		return fmt.Errorf("invalid TagDictOffset: cannot be zero")
	}
	if fileSize > 0 && tagDictOffset >= fileSize {
		logger.Error("CORRUPTION PREVENTION: TagDictOffset %d exceeds file size %d", tagDictOffset, fileSize)
		return fmt.Errorf("invalid TagDictOffset %d: exceeds file size %d", tagDictOffset, fileSize)
	}
	if tagDictOffset > maxReasonableOffset {
		logger.Error("CORRUPTION PREVENTION: TagDictOffset %d exceeds reasonable limit %d", tagDictOffset, maxReasonableOffset)
		return fmt.Errorf("invalid TagDictOffset %d: exceeds reasonable limit", tagDictOffset)
	}
	
	// Validate EntityIndexOffset before updating
	if entityIndexOffset == 0 {
		logger.Error("CORRUPTION PREVENTION: Refusing to set EntityIndexOffset to zero")
		return fmt.Errorf("invalid EntityIndexOffset: cannot be zero")
	}
	if fileSize > 0 && entityIndexOffset >= fileSize {
		logger.Error("CORRUPTION PREVENTION: EntityIndexOffset %d exceeds file size %d", entityIndexOffset, fileSize)
		return fmt.Errorf("invalid EntityIndexOffset %d: exceeds file size %d", entityIndexOffset, fileSize)
	}
	if entityIndexOffset > maxReasonableOffset {
		logger.Error("CORRUPTION PREVENTION: EntityIndexOffset %d exceeds reasonable limit %d", entityIndexOffset, maxReasonableOffset)
		return fmt.Errorf("invalid EntityIndexOffset %d: exceeds reasonable limit", entityIndexOffset)
	}
	
	// Only update if validation passes
	hs.header.TagDictOffset = tagDictOffset
	hs.header.EntityIndexOffset = entityIndexOffset
	
	logger.Trace("HeaderSync: Updated offsets - TagDict: %d, EntityIndex: %d", tagDictOffset, entityIndexOffset)
	return nil
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
		EntityCount: hs.header.EntityCount, // Use header directly - index is source of truth
	}
}

// RestoreFromSnapshot safely restores HeaderSync state from a snapshot
// Used to recover from checkpoint corruption
func (hs *HeaderSync) RestoreFromSnapshot(snapshot *HeaderSnapshot) {
	hs.mu.Lock()
	defer hs.mu.Unlock()
	
	hs.header = snapshot.Header
	hs.walSequence.Store(snapshot.WALSequence)
	// EntityCount restored via header - index is source of truth
}

// ValidateHeader checks if the header has valid values
// Returns true if header is valid, false if corrupted
func (hs *HeaderSync) ValidateHeader() bool {
	hs.mu.RLock()
	defer hs.mu.RUnlock()
	
	// Check magic number and version first
	if hs.header.Magic != 0x46465545 { // "EUFF"
		logger.Error("HeaderSync validation failed: Invalid magic number %x", hs.header.Magic)
		return false
	}
	
	if hs.header.Version == 0 {
		logger.Error("HeaderSync validation failed: Invalid version %d", hs.header.Version)
		return false
	}
	
	fileSize := hs.header.FileSize
	maxReasonableOffset := uint64(10 * 1024 * 1024 * 1024) // 10GB safety limit
	
	// Enhanced WALOffset validation
	if hs.header.WALOffset == 0 {
		logger.Error("HeaderSync validation failed: WALOffset cannot be zero")
		return false
	}
	if fileSize > 0 && hs.header.WALOffset >= fileSize {
		logger.Error("HeaderSync validation failed: WALOffset %d exceeds file size %d", hs.header.WALOffset, fileSize)
		return false
	}
	if hs.header.WALOffset > maxReasonableOffset {
		logger.Error("HeaderSync validation failed: WALOffset %d exceeds reasonable limit", hs.header.WALOffset)
		return false
	}
	
	// Enhanced TagDictOffset validation
	if hs.header.TagDictOffset == 0 {
		logger.Error("HeaderSync validation failed: TagDictOffset cannot be zero")
		return false
	}
	if fileSize > 0 && hs.header.TagDictOffset >= fileSize {
		logger.Error("HeaderSync validation failed: TagDictOffset %d exceeds file size %d", hs.header.TagDictOffset, fileSize)
		return false
	}
	if hs.header.TagDictOffset > maxReasonableOffset {
		logger.Error("HeaderSync validation failed: TagDictOffset %d exceeds reasonable limit", hs.header.TagDictOffset)
		return false
	}
	
	// Enhanced EntityIndexOffset validation
	if hs.header.EntityIndexOffset == 0 {
		logger.Error("HeaderSync validation failed: EntityIndexOffset cannot be zero")
		return false
	}
	if fileSize > 0 && hs.header.EntityIndexOffset >= fileSize {
		logger.Error("HeaderSync validation failed: EntityIndexOffset %d exceeds file size %d", hs.header.EntityIndexOffset, fileSize)
		return false
	}
	if hs.header.EntityIndexOffset > maxReasonableOffset {
		logger.Error("HeaderSync validation failed: EntityIndexOffset %d exceeds reasonable limit", hs.header.EntityIndexOffset)
		return false
	}
	
	return true
}