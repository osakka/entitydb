// Package binary implements the EntityDB Binary Format (EBF), a custom binary storage
// format optimized for temporal data with high-performance concurrent access.
//
// # Format Overview
//
// The EBF format consists of:
//   - Fixed-size header (128 bytes) containing metadata and offsets
//   - Write-Ahead Log for durability and concurrent access
//   - Tag dictionary for string compression and interning
//   - Entity index for O(1) lookups by ID
//   - Deletion index for tracking deleted/purged entities
//   - Entity data blocks containing tags and content
//
// # File Structure
//
//	+----------------+ 0x00
//	|     Header     | 128 bytes
//	+----------------+ 0x80
//	|      WAL       | Variable size
//	+----------------+
//	| Tag Dictionary | Variable size
//	+----------------+
//	|  Entity Index  | EntityCount * 112 bytes
//	+----------------+
//	| Deletion Index | DeletionCount * 256 bytes
//	+----------------+
//	|  Entity Data   | Variable size blocks
//	+----------------+
//
// # Design Principles
//
//   - Memory-mapped file support for zero-copy reads
//   - String interning to reduce memory usage
//   - Fixed-size index entries for predictable performance
//   - Write-Ahead Logging (WAL) for durability
//   - Lock-free reads with sharded write locks
//
// # Example
//
//	// Reading an EBF file
//	reader, err := NewReader("entities.ebf")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer reader.Close()
//
//	entity, err := reader.GetByID("user-123")
//	if err != nil {
//	    log.Fatal(err)
//	}
package binary

import (
	"encoding/binary"
	"entitydb/models"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"entitydb/logger"
)

const (
	// MagicNumber identifies EntityDB Unified File Format (EUFF).
	// Value: 0x45555446 ("EUFF" - EntityDB Unified File Format)
	MagicNumber uint32 = 0x45555446
	
	// FormatVersion indicates the unified format version.
	// Version 3: Added deletion sections for temporal deletion architecture
	FormatVersion uint32 = 3
	
	// HeaderSize is the fixed size of the unified file header in bytes.
	// The unified header contains metadata and section offsets for all components.
	HeaderSize = 128
	
	// IndexEntrySize is the size of each entity index entry.
	// Structure: EntityID (96 bytes) + Offset (8 bytes) + Size (4 bytes) + Flags (4 bytes)
	IndexEntrySize = 112
)

var (
	// ErrInvalidFormat is returned when the file magic number doesn't match
	ErrInvalidFormat = errors.New("invalid file format")
	
	// ErrVersionMismatch is returned when the format version is unsupported
	ErrVersionMismatch = errors.New("unsupported format version")
	
	// ErrUnknownFormat is returned when file format cannot be determined
	ErrUnknownFormat = errors.New("unknown file format")
)

// Header represents the unified file header containing metadata and section offsets
// for all components in a single file. The header is always 128 bytes and located at the
// beginning of the file.
//
// # Binary Layout (Little Endian)
//
//	Offset  Size  Field
//	0x00    4     Magic (0x45555446) "EUFF"
//	0x04    4     Version (2)
//	0x08    8     FileSize
//	0x10    8     WALOffset
//	0x18    8     WALSize
//	0x20    8     DataOffset
//	0x28    8     DataSize
//	0x30    8     TagDictOffset
//	0x38    8     TagDictSize
//	0x40    8     EntityIndexOffset
//	0x48    8     EntityIndexSize
//	0x50    8     EntityCount
//	0x58    8     LastModified (Unix timestamp)
//	0x60    8     WALSequence
//	0x68    8     CheckpointSequence
//	0x70    8     DeletionIndexOffset
//	0x78    8     DeletionIndexSize
type Header struct {
	Magic              uint32  // File format identifier (must be MagicNumber)
	Version            uint32  // Format version (must be FormatVersion)
	FileSize           uint64  // Total file size in bytes
	WALOffset          uint64  // Offset to WAL section
	WALSize            uint64  // Size of WAL section in bytes
	DataOffset         uint64  // Offset to entity data section
	DataSize           uint64  // Size of entity data section in bytes
	TagDictOffset      uint64  // Offset to tag dictionary section
	TagDictSize        uint64  // Size of tag dictionary in bytes
	EntityIndexOffset  uint64  // Offset to entity index section
	EntityIndexSize    uint64  // Size of entity index in bytes
	EntityCount        uint64  // Number of entities in the file
	LastModified       int64   // Unix timestamp of last modification
	WALSequence        uint64  // Current WAL sequence number
	CheckpointSequence uint64  // Last checkpoint sequence number
	DeletionIndexOffset uint64  // Offset to deletion index section
	DeletionIndexSize   uint64  // Size of deletion index in bytes
}

// DeletionEntry represents a single entry in the deletion index.
// Each entry tracks when an entity was deleted and its lifecycle state.
//
// # Binary Layout (Little Endian)
//
//	Offset  Size  Field
//	0x00    96    EntityID (null-terminated string)
//	0x60    8     DeletionTimestamp (Unix nanoseconds)
//	0x68    4     LifecycleState (0=active, 1=soft_deleted, 2=archived, 3=purged)
//	0x6C    4     Flags (reserved for future use)
//	0x70    32    DeletedBy (user ID who performed deletion)
//	0x90    64    Reason (deletion reason, null-terminated)
//	0xD0    32    Policy (retention policy name, null-terminated)
//	0xF0    16    Reserved
//
// Total size: 256 bytes per entry
const DeletionEntrySize = 256

type DeletionEntry struct {
	EntityID          [96]byte  // Entity identifier (null-terminated)
	DeletionTimestamp int64     // When the deletion occurred (Unix nanoseconds)
	LifecycleState    uint32    // Current lifecycle state (0-3)
	Flags             uint32    // Reserved flags for future use
	DeletedBy         [32]byte  // User ID who performed the deletion
	Reason            [64]byte  // Reason for deletion (null-terminated)
	Policy            [32]byte  // Retention policy name (null-terminated)
	Reserved          [16]byte  // Reserved for future extensions
}

// DeletionState represents the lifecycle states in the deletion index
type DeletionState uint32

const (
	DeletionStateActive      DeletionState = 0
	DeletionStateSoftDeleted DeletionState = 1
	DeletionStateArchived    DeletionState = 2
	DeletionStatePurged      DeletionState = 3
)

// String returns the string representation of a deletion state
func (ds DeletionState) String() string {
	switch ds {
	case DeletionStateActive:
		return "active"
	case DeletionStateSoftDeleted:
		return "soft_deleted"
	case DeletionStateArchived:
		return "archived"
	case DeletionStatePurged:
		return "purged"
	default:
		return "unknown"
	}
}

// NewDeletionEntry creates a new deletion entry from entity lifecycle data
func NewDeletionEntry(entityID string, state models.EntityLifecycleState, deletedBy, reason, policy string, timestamp int64) *DeletionEntry {
	entry := &DeletionEntry{
		DeletionTimestamp: timestamp,
		Flags:            0,
	}
	
	// Copy strings with bounds checking
	copy(entry.EntityID[:], entityID)
	copy(entry.DeletedBy[:], deletedBy)
	copy(entry.Reason[:], reason)
	copy(entry.Policy[:], policy)
	
	// Convert lifecycle state to deletion state
	switch state {
	case models.StateActive:
		entry.LifecycleState = uint32(DeletionStateActive)
	case models.StateSoftDeleted:
		entry.LifecycleState = uint32(DeletionStateSoftDeleted)
	case models.StateArchived:
		entry.LifecycleState = uint32(DeletionStateArchived)
	case models.StatePurged:
		entry.LifecycleState = uint32(DeletionStatePurged)
	}
	
	return entry
}

// GetEntityID returns the entity ID as a string (null-terminated)
func (de *DeletionEntry) GetEntityID() string {
	// Find null terminator
	for i, b := range de.EntityID {
		if b == 0 {
			return string(de.EntityID[:i])
		}
	}
	return string(de.EntityID[:])
}

// GetDeletedBy returns the user ID who performed the deletion
func (de *DeletionEntry) GetDeletedBy() string {
	for i, b := range de.DeletedBy {
		if b == 0 {
			return string(de.DeletedBy[:i])
		}
	}
	return string(de.DeletedBy[:])
}

// GetReason returns the deletion reason
func (de *DeletionEntry) GetReason() string {
	for i, b := range de.Reason {
		if b == 0 {
			return string(de.Reason[:i])
		}
	}
	return string(de.Reason[:])
}

// GetPolicy returns the retention policy name
func (de *DeletionEntry) GetPolicy() string {
	for i, b := range de.Policy {
		if b == 0 {
			return string(de.Policy[:i])
		}
	}
	return string(de.Policy[:])
}

// GetLifecycleState returns the lifecycle state as a models type
func (de *DeletionEntry) GetLifecycleState() models.EntityLifecycleState {
	switch DeletionState(de.LifecycleState) {
	case DeletionStateActive:
		return models.StateActive
	case DeletionStateSoftDeleted:
		return models.StateSoftDeleted
	case DeletionStateArchived:
		return models.StateArchived
	case DeletionStatePurged:
		return models.StatePurged
	default:
		return models.StateActive
	}
}

// LegacyHeader removed - single source of truth: unified format only

// FileFormat represents the detected file format type
type FileFormat int

const (
	FormatUnknown FileFormat = iota
	FormatUnified             // Unified EUFF format
)

// DetectFileFormat determines the file format by reading the magic number.
// Single source of truth: only unified format supported.
func DetectFileFormat(filename string) (FileFormat, error) {
	file, err := os.Open(filename)
	if err != nil {
		return FormatUnknown, err
	}
	defer file.Close()
	
	var magic uint32
	err = binary.Read(file, binary.LittleEndian, &magic)
	if err != nil {
		return FormatUnknown, err
	}
	
	if magic == MagicNumber {
		return FormatUnified, nil
	}
	
	return FormatUnknown, ErrUnknownFormat
}

// Write serializes the unified header to the provided writer.
// The header is written as a fixed 128-byte block in little-endian format.
//
// Returns an error if the write operation fails.
func (h *Header) Write(w io.Writer) error {
	buf := make([]byte, HeaderSize)
	
	// CRITICAL FIX: Validate and correct header fields before serialization
	// This prevents the version corruption issue causing "unsupported format version" errors
	magic := h.Magic
	version := h.Version
	
	// Ensure magic number is correct
	if magic != MagicNumber {
		logger.Warn("Header magic number corrupted (0x%x), correcting to 0x%x", magic, MagicNumber)
		magic = MagicNumber
	}
	
	// Ensure version is correct - this fixes the version 678 corruption issue
	if version != FormatVersion {
		logger.Warn("Header version corrupted (%d), correcting to %d", version, FormatVersion)
		version = FormatVersion
	}
	
	// Update in-memory header with corrected values to prevent future corruption
	h.Magic = magic
	h.Version = version
	
	// Serialize all fields to the buffer in little-endian format
	binary.LittleEndian.PutUint32(buf[0:4], magic)
	binary.LittleEndian.PutUint32(buf[4:8], version)
	binary.LittleEndian.PutUint64(buf[8:16], h.FileSize)
	binary.LittleEndian.PutUint64(buf[16:24], h.WALOffset)
	binary.LittleEndian.PutUint64(buf[24:32], h.WALSize)
	binary.LittleEndian.PutUint64(buf[32:40], h.DataOffset)
	binary.LittleEndian.PutUint64(buf[40:48], h.DataSize)
	binary.LittleEndian.PutUint64(buf[48:56], h.TagDictOffset)
	binary.LittleEndian.PutUint64(buf[56:64], h.TagDictSize)
	binary.LittleEndian.PutUint64(buf[64:72], h.EntityIndexOffset)
	binary.LittleEndian.PutUint64(buf[72:80], h.EntityIndexSize)
	binary.LittleEndian.PutUint64(buf[80:88], h.EntityCount)
	binary.LittleEndian.PutUint64(buf[88:96], uint64(h.LastModified))
	binary.LittleEndian.PutUint64(buf[96:104], h.WALSequence)
	binary.LittleEndian.PutUint64(buf[104:112], h.CheckpointSequence)
	binary.LittleEndian.PutUint64(buf[112:120], h.DeletionIndexOffset)
	binary.LittleEndian.PutUint64(buf[120:128], h.DeletionIndexSize)
	
	logger.Debug("Header.Write - EntityCount=%d, WALSequence=%d, FileSize=%d", 
		h.EntityCount, h.WALSequence, h.FileSize)
	n, err := w.Write(buf)
	if err != nil {
		logger.Error("Header.Write failed: %v", err)
		return err
	}
	logger.Debug("Header.Write wrote %d bytes", n)
	return nil
}

// Read deserializes the unified header from the provided reader.
// The method reads exactly 128 bytes and validates the magic number and version.
//
// Returns:
//   - ErrInvalidFormat if the magic number doesn't match
//   - ErrVersionMismatch if the version is unsupported
//   - io.ErrUnexpectedEOF if the header is incomplete
func (h *Header) Read(r io.Reader) error {
	buf := make([]byte, HeaderSize)
	n, err := io.ReadFull(r, buf)
	if err != nil {
		// Check if we have a partial header
		if n > 0 && err == io.ErrUnexpectedEOF {
			// Try to parse what we have
			if n >= 8 {
				h.Magic = binary.LittleEndian.Uint32(buf[0:4])
				h.Version = binary.LittleEndian.Uint32(buf[4:8])
				if h.Magic == MagicNumber && h.Version == FormatVersion {
					// Valid header but incomplete - assume empty file
					h.EntityCount = 0
					return nil
				}
			}
		}
		return err
	}
	
	h.Magic = binary.LittleEndian.Uint32(buf[0:4])
	h.Version = binary.LittleEndian.Uint32(buf[4:8])
	h.FileSize = binary.LittleEndian.Uint64(buf[8:16])
	h.WALOffset = binary.LittleEndian.Uint64(buf[16:24])
	h.WALSize = binary.LittleEndian.Uint64(buf[24:32])
	h.DataOffset = binary.LittleEndian.Uint64(buf[32:40])
	h.DataSize = binary.LittleEndian.Uint64(buf[40:48])
	h.TagDictOffset = binary.LittleEndian.Uint64(buf[48:56])
	h.TagDictSize = binary.LittleEndian.Uint64(buf[56:64])
	h.EntityIndexOffset = binary.LittleEndian.Uint64(buf[64:72])
	h.EntityIndexSize = binary.LittleEndian.Uint64(buf[72:80])
	h.EntityCount = binary.LittleEndian.Uint64(buf[80:88])
	h.LastModified = int64(binary.LittleEndian.Uint64(buf[88:96]))
	h.WALSequence = binary.LittleEndian.Uint64(buf[96:104])
	h.CheckpointSequence = binary.LittleEndian.Uint64(buf[104:112])
	h.DeletionIndexOffset = binary.LittleEndian.Uint64(buf[112:120])
	h.DeletionIndexSize = binary.LittleEndian.Uint64(buf[120:128])
	
	if h.Magic != MagicNumber {
		return ErrInvalidFormat
	}
	// Support both version 2 (without deletion index) and version 3 (with deletion index)
	if h.Version != 2 && h.Version != FormatVersion {
		return ErrVersionMismatch
	}
	
	// For version 2 files, initialize deletion index fields to zero
	if h.Version == 2 {
		h.DeletionIndexOffset = 0
		h.DeletionIndexSize = 0
	}
	
	return nil
}





// IndexEntry represents an entry in the entity index section.
// Each entry maps an entity ID to its location in the file.
//
// # Binary Layout (112 bytes)
//
//	Offset  Size  Field
//	0x00    96    EntityID (UUID with optional prefix, null-terminated)
//	0x60    8     Offset (file position of entity data)
//	0x68    4     Size (size of entity data in bytes)
//	0x6C    4     Flags (reserved for future use)
type IndexEntry struct {
	EntityID [96]byte  // UUID with optional prefix (e.g., "dataset:uuid")
	Offset   uint64    // File offset to entity data
	Size     uint32    // Size of entity data block
	Flags    uint32    // Reserved flags (0 = normal, 1 = compressed)
}

// EntityHeader represents the header of an entity data block.
// This header precedes the tag and content data for each entity.
//
// # Binary Layout (16 bytes)
//
//	Offset  Size  Field
//	0x00    8     Modified (Unix timestamp in nanoseconds)
//	0x08    2     TagCount (number of tags)
//	0x0A    2     ContentCount (number of content chunks)
//	0x0C    4     Reserved (must be 0)
type EntityHeader struct {
	Modified     int64   // Last modification timestamp (Unix nanoseconds)
	TagCount     uint16  // Number of tags in this entity
	ContentCount uint16  // Number of content chunks (for autochunking)
	Reserved     uint32  // Reserved for future use (must be 0)
}

// TagDictionary manages tag string compression using dictionary encoding.
// It maps tag strings to numeric IDs to reduce storage space and improve performance.
//
// The dictionary uses string interning to ensure each unique tag string
// is stored only once in memory, significantly reducing memory usage for
// systems with many repeated tags.
//
// Thread-safe for concurrent access.
type TagDictionary struct {
	idToTag map[uint32]string   // Maps numeric IDs to tag strings
	tagToID map[string]uint32   // Maps tag strings to numeric IDs
	nextID  uint32              // Next available ID
	mu      sync.RWMutex        // Protects concurrent access
}

// NewTagDictionary creates a new empty tag dictionary.
// IDs start at 1 (0 is reserved for "no tag").
func NewTagDictionary() *TagDictionary {
	return &TagDictionary{
		idToTag: make(map[uint32]string),
		tagToID: make(map[string]uint32),
		nextID:  1,
	}
}

// GetOrCreateID returns the numeric ID for a tag string.
// If the tag doesn't exist in the dictionary, it's added with a new ID.
// The tag string is interned to reduce memory usage.
//
// Thread-safe for concurrent access.
func (d *TagDictionary) GetOrCreateID(tag string) uint32 {
	// Intern the string first
	tag = models.Intern(tag)
	
	// Fast path - read lock
	d.mu.RLock()
	if id, exists := d.tagToID[tag]; exists {
		d.mu.RUnlock()
		return id
	}
	d.mu.RUnlock()
	
	// Slow path - write lock
	d.mu.Lock()
	defer d.mu.Unlock()
	
	// Double check
	if id, exists := d.tagToID[tag]; exists {
		return id
	}
	
	// Intern the tag string to save memory
	internedTag := models.Intern(tag)
	
	id := d.nextID
	d.nextID++
	d.idToTag[id] = internedTag
	d.tagToID[internedTag] = id
	return id
}

// GetTag returns the tag string for a given numeric ID.
// Returns empty string if the ID doesn't exist.
//
// Thread-safe for concurrent access.
func (d *TagDictionary) GetTag(id uint32) string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.idToTag[id]
}

// Write serializes the tag dictionary to the provided writer.
//
// # Binary Format
//
//	4 bytes: Count (number of entries)
//	For each entry:
//	  4 bytes: ID
//	  2 bytes: Tag length
//	  N bytes: Tag string
//
// Returns an error if the write operation fails.
func (d *TagDictionary) Write(w io.Writer) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	
	// Write entry count
	if err := binary.Write(w, binary.LittleEndian, uint32(len(d.idToTag))); err != nil {
		return err
	}
	
	// Write each dictionary entry
	for id, tag := range d.idToTag {
		if err := binary.Write(w, binary.LittleEndian, id); err != nil {
			return err
		}
		if err := binary.Write(w, binary.LittleEndian, uint16(len(tag))); err != nil {
			return err
		}
		if _, err := w.Write([]byte(tag)); err != nil {
			return err
		}
	}
	
	return nil
}

// Read deserializes the tag dictionary from the provided reader.
// Existing dictionary contents are preserved - new entries are merged.
//
// The method updates nextID to ensure new tags get unique IDs.
//
// Returns an error if the read operation fails or data is corrupted.
func (d *TagDictionary) Read(r io.Reader) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	// Read entry count
	var count uint32
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return err
	}
	
	// Read each dictionary entry
	for i := uint32(0); i < count; i++ {
		var id uint32
		var length uint16
		
		// Read ID
		if err := binary.Read(r, binary.LittleEndian, &id); err != nil {
			return err
		}
		// Read tag length
		if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
			return err
		}
		
		// Read tag string
		tag := make([]byte, length)
		if _, err := io.ReadFull(r, tag); err != nil {
			return err
		}
		
		// Intern the tag string to save memory
		tagStr := models.Intern(string(tag))
		d.idToTag[id] = tagStr
		d.tagToID[tagStr] = id
		
		// Update nextID to avoid collisions
		if id >= d.nextID {
			d.nextID = id + 1
		}
	}
	
	return nil
}

// =============================================================================
// Deletion Index Functions
// =============================================================================

// WriteDeletionEntry writes a deletion entry to the provided writer
func (de *DeletionEntry) Write(w io.Writer) error {
	buf := make([]byte, DeletionEntrySize)
	
	// Copy fixed-size fields
	copy(buf[0:96], de.EntityID[:])
	binary.LittleEndian.PutUint64(buf[96:104], uint64(de.DeletionTimestamp))
	binary.LittleEndian.PutUint32(buf[104:108], de.LifecycleState)
	binary.LittleEndian.PutUint32(buf[108:112], de.Flags)
	copy(buf[112:144], de.DeletedBy[:])
	copy(buf[144:208], de.Reason[:])
	copy(buf[208:240], de.Policy[:])
	copy(buf[240:256], de.Reserved[:])
	
	_, err := w.Write(buf)
	return err
}

// ReadDeletionEntry reads a deletion entry from the provided reader
func (de *DeletionEntry) Read(r io.Reader) error {
	buf := make([]byte, DeletionEntrySize)
	if _, err := io.ReadFull(r, buf); err != nil {
		return err
	}
	
	// Extract fixed-size fields
	copy(de.EntityID[:], buf[0:96])
	de.DeletionTimestamp = int64(binary.LittleEndian.Uint64(buf[96:104]))
	de.LifecycleState = binary.LittleEndian.Uint32(buf[104:108])
	de.Flags = binary.LittleEndian.Uint32(buf[108:112])
	copy(de.DeletedBy[:], buf[112:144])
	copy(de.Reason[:], buf[144:208])
	copy(de.Policy[:], buf[208:240])
	copy(de.Reserved[:], buf[240:256])
	
	return nil
}

// DeletionIndex manages the deletion index for tracking deleted/purged entities
type DeletionIndex struct {
	entries map[string]*DeletionEntry // EntityID -> DeletionEntry
	mu      sync.RWMutex              // Protect concurrent access
}

// NewDeletionIndex creates a new deletion index
func NewDeletionIndex() *DeletionIndex {
	return &DeletionIndex{
		entries: make(map[string]*DeletionEntry),
	}
}

// AddEntry adds or updates a deletion entry
func (di *DeletionIndex) AddEntry(entry *DeletionEntry) {
	di.mu.Lock()
	defer di.mu.Unlock()
	
	entityID := entry.GetEntityID()
	di.entries[entityID] = entry
}

// GetEntry retrieves a deletion entry by entity ID
func (di *DeletionIndex) GetEntry(entityID string) (*DeletionEntry, bool) {
	di.mu.RLock()
	defer di.mu.RUnlock()
	
	entry, exists := di.entries[entityID]
	return entry, exists
}

// RemoveEntry removes a deletion entry (e.g., when entity is restored)
func (di *DeletionIndex) RemoveEntry(entityID string) {
	di.mu.Lock()
	defer di.mu.Unlock()
	
	delete(di.entries, entityID)
}

// GetAllEntries returns all deletion entries
func (di *DeletionIndex) GetAllEntries() []*DeletionEntry {
	di.mu.RLock()
	defer di.mu.RUnlock()
	
	entries := make([]*DeletionEntry, 0, len(di.entries))
	for _, entry := range di.entries {
		entries = append(entries, entry)
	}
	
	return entries
}

// GetEntriesByState returns all entries in a specific lifecycle state
func (di *DeletionIndex) GetEntriesByState(state models.EntityLifecycleState) []*DeletionEntry {
	di.mu.RLock()
	defer di.mu.RUnlock()
	
	var result []*DeletionEntry
	for _, entry := range di.entries {
		if entry.GetLifecycleState() == state {
			result = append(result, entry)
		}
	}
	
	return result
}

// Count returns the number of deletion entries
func (di *DeletionIndex) Count() int {
	di.mu.RLock()
	defer di.mu.RUnlock()
	
	return len(di.entries)
}

// WriteTo writes the entire deletion index to a writer
func (di *DeletionIndex) WriteTo(w io.Writer) error {
	di.mu.RLock()
	defer di.mu.RUnlock()
	
	for _, entry := range di.entries {
		if err := entry.Write(w); err != nil {
			return fmt.Errorf("failed to write deletion entry: %w", err)
		}
	}
	
	return nil
}

// ReadFrom reads deletion entries from a reader
func (di *DeletionIndex) ReadFrom(r io.Reader, entryCount int) error {
	di.mu.Lock()
	defer di.mu.Unlock()
	
	for i := 0; i < entryCount; i++ {
		entry := &DeletionEntry{}
		if err := entry.Read(r); err != nil {
			return fmt.Errorf("failed to read deletion entry %d: %w", i, err)
		}
		
		entityID := entry.GetEntityID()
		di.entries[entityID] = entry
	}
	
	return nil
}