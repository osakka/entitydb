// Package binary implements the EntityDB Binary Format (EBF), a custom binary storage
// format optimized for temporal data with high-performance concurrent access.
//
// # Format Overview
//
// The EBF format consists of:
//   - Fixed-size header (64 bytes) containing metadata and offsets
//   - Tag dictionary for string compression and interning
//   - Entity index for O(1) lookups by ID
//   - Entity data blocks containing tags and content
//
// # File Structure
//
//	+----------------+ 0x00
//	|     Header     | 64 bytes
//	+----------------+ 0x40
//	| Tag Dictionary | Variable size
//	+----------------+
//	|  Entity Index  | EntityCount * 112 bytes
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
	// Version 2: Unified file format with embedded WAL and sections
	FormatVersion uint32 = 2
	
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
//	0x70    16    Reserved (for future sections)
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
	Reserved           [16]byte // Reserved for future sections
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
	
	// Serialize all fields to the buffer in little-endian format
	binary.LittleEndian.PutUint32(buf[0:4], h.Magic)
	binary.LittleEndian.PutUint32(buf[4:8], h.Version)
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
	copy(buf[112:128], h.Reserved[:])
	
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
	copy(h.Reserved[:], buf[112:128])
	
	if h.Magic != MagicNumber {
		return ErrInvalidFormat
	}
	if h.Version != FormatVersion {
		return ErrVersionMismatch
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