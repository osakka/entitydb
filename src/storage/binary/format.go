package binary

import (
	"encoding/binary"
	"entitydb/models"
	"errors"
	"io"
	"sync"
	"entitydb/logger"
)

const (
	// Magic number "EBDF" (EntityDB Format)
	MagicNumber uint32 = 0x45424446
	
	// Current format version
	FormatVersion uint32 = 1
	
	// Header size in bytes
	HeaderSize = 64
	
	// Index entry size (96-byte ID + 8-byte offset + 4-byte size + 4-byte flags)
	IndexEntrySize = 112
)

var (
	ErrInvalidFormat = errors.New("invalid file format")
	ErrVersionMismatch = errors.New("unsupported format version")
)

// Header represents the file header
type Header struct {
	Magic            uint32
	Version          uint32
	FileSize         uint64
	TagDictOffset    uint64
	TagDictSize      uint64
	EntityIndexOffset uint64
	EntityIndexSize  uint64
	EntityCount      uint64
	LastModified     int64
}

// Write writes the header to writer
func (h *Header) Write(w io.Writer) error {
	buf := make([]byte, HeaderSize)
	
	binary.LittleEndian.PutUint32(buf[0:4], h.Magic)
	binary.LittleEndian.PutUint32(buf[4:8], h.Version)
	binary.LittleEndian.PutUint64(buf[8:16], h.FileSize)
	binary.LittleEndian.PutUint64(buf[16:24], h.TagDictOffset)
	binary.LittleEndian.PutUint64(buf[24:32], h.TagDictSize)
	binary.LittleEndian.PutUint64(buf[32:40], h.EntityIndexOffset)
	binary.LittleEndian.PutUint64(buf[40:48], h.EntityIndexSize)
	binary.LittleEndian.PutUint64(buf[48:56], h.EntityCount)
	binary.LittleEndian.PutUint64(buf[56:64], uint64(h.LastModified))
	
	logger.Debug("Header.Write - EntityCount=%d at offset 48, FileSize=%d at offset 8", h.EntityCount, h.FileSize)
	n, err := w.Write(buf)
	if err != nil {
		logger.Error("Header.Write failed: %v", err)
		return err
	}
	logger.Debug("Header.Write wrote %d bytes", n)
	return nil
}

// Read reads the header from reader
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
	h.TagDictOffset = binary.LittleEndian.Uint64(buf[16:24])
	h.TagDictSize = binary.LittleEndian.Uint64(buf[24:32])
	h.EntityIndexOffset = binary.LittleEndian.Uint64(buf[32:40])
	h.EntityIndexSize = binary.LittleEndian.Uint64(buf[40:48])
	h.EntityCount = binary.LittleEndian.Uint64(buf[48:56])
	h.LastModified = int64(binary.LittleEndian.Uint64(buf[56:64]))
	
	if h.Magic != MagicNumber {
		return ErrInvalidFormat
	}
	if h.Version != FormatVersion {
		return ErrVersionMismatch
	}
	
	return nil
}

// IndexEntry represents an entry in the entity index
type IndexEntry struct {
	EntityID [96]byte  // UUID with prefix (up to 96 bytes)
	Offset   uint64
	Size     uint32
	Flags    uint32
}

// EntityHeader represents the header of an entity data block
type EntityHeader struct {
	Modified     int64
	TagCount     uint16
	ContentCount uint16
	Reserved     uint32
}

// TagDictionary manages tag string compression with interning
type TagDictionary struct {
	idToTag map[uint32]string
	tagToID map[string]uint32
	nextID  uint32
	mu      sync.RWMutex // Add mutex for thread safety
}

// NewTagDictionary creates a new tag dictionary
func NewTagDictionary() *TagDictionary {
	return &TagDictionary{
		idToTag: make(map[uint32]string),
		tagToID: make(map[string]uint32),
		nextID:  1,
	}
}

// GetOrCreateID returns the ID for a tag, creating if necessary
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

// GetTag returns the tag string for an ID
func (d *TagDictionary) GetTag(id uint32) string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.idToTag[id]
}

// Write writes the dictionary to writer
func (d *TagDictionary) Write(w io.Writer) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	
	// Write count
	if err := binary.Write(w, binary.LittleEndian, uint32(len(d.idToTag))); err != nil {
		return err
	}
	
	// Write entries
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

// Read reads the dictionary from reader
func (d *TagDictionary) Read(r io.Reader) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	var count uint32
	if err := binary.Read(r, binary.LittleEndian, &count); err != nil {
		return err
	}
	
	for i := uint32(0); i < count; i++ {
		var id uint32
		var length uint16
		
		if err := binary.Read(r, binary.LittleEndian, &id); err != nil {
			return err
		}
		if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
			return err
		}
		
		tag := make([]byte, length)
		if _, err := io.ReadFull(r, tag); err != nil {
			return err
		}
		
		// Intern the tag string
		tagStr := models.Intern(string(tag))
		d.idToTag[id] = tagStr
		d.tagToID[tagStr] = id
		if id >= d.nextID {
			d.nextID = id + 1
		}
	}
	
	return nil
}