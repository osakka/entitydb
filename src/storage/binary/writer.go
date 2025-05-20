package binary

import (
	"bytes"
	"encoding/binary"
	"entitydb/models"
	"os"
	"sync"
	"time"
	"entitydb/logger"
)

// Writer handles writing entities to binary format
type Writer struct {
	file     *os.File
	header   *Header
	tagDict  *TagDictionary
	index    map[string]*IndexEntry
	buffer   *bytes.Buffer
	mu       sync.Mutex
}

// NewWriter creates a new writer for the given file
func NewWriter(filename string) (*Writer, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, err
	}
	
	w := &Writer{
		file:    file,
		header:  &Header{Magic: MagicNumber, Version: FormatVersion},
		tagDict: NewTagDictionary(),
		index:   make(map[string]*IndexEntry),
		buffer:  new(bytes.Buffer),
	}
	
	// Try to read existing file
	if stat, err := file.Stat(); err == nil && stat.Size() > 0 {
		if err := w.readExisting(); err != nil {
			logger.Debug("Warning: failed to read existing file, creating new: %v", err)
			// Reset the file
			file.Truncate(0)
			file.Seek(0, 0)
			// Write initial header
			if err := w.writeHeader(); err != nil {
				return nil, err
			}
			if err := w.file.Sync(); err != nil {
				return nil, err
			}
		}
	} else {
		// New file, write initial header
		if err := w.writeHeader(); err != nil {
			return nil, err
		}
		// Ensure data is flushed to disk
		if err := w.file.Sync(); err != nil {
			return nil, err
		}
	}
	
	return w, nil
}

// WriteEntity writes an entity to the file
func (w *Writer) WriteEntity(entity *models.Entity) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	logger.Debug("WriteEntity called for entity %s", entity.ID)
	logger.Debug("Current file position before write: %d", w.getFilePosition())
	logger.Debug("Current entity count: %d", w.header.EntityCount)
	
	// Prepare entity data
	w.buffer.Reset()
	
	// Convert tags to IDs
	tagIDs := make([]uint32, len(entity.Tags))
	for i, tag := range entity.Tags {
		tagIDs[i] = w.tagDict.GetOrCreateID(tag)
		logger.Debug("Tag '%s' assigned ID %d", tag, tagIDs[i])
	}
	
	// Write entity header
	header := EntityHeader{
		Modified:     time.Now().Unix(),
		TagCount:     uint16(len(tagIDs)),
		ContentCount: 1, // Now we store content as a single item
	}
	
	logger.Debug("Writing entity header: Modified=%d, TagCount=%d, ContentCount=%d", 
		header.Modified, header.TagCount, header.ContentCount)
	
	binary.Write(w.buffer, binary.LittleEndian, header)
	
	// Write tag IDs
	for _, id := range tagIDs {
		binary.Write(w.buffer, binary.LittleEndian, id)
	}
	
	// Write content as a single item
	if len(entity.Content) > 0 {
		// Type: application/octet-stream for raw bytes
		contentType := "application/octet-stream"
		binary.Write(w.buffer, binary.LittleEndian, uint16(len(contentType)))
		w.buffer.WriteString(contentType)
		
		// Value: the actual content
		binary.Write(w.buffer, binary.LittleEndian, uint32(len(entity.Content)))
		w.buffer.Write(entity.Content)
		
		// Timestamp: current time
		binary.Write(w.buffer, binary.LittleEndian, time.Now().UnixNano())
	} else {
		// Empty content
		contentType := "application/octet-stream"
		binary.Write(w.buffer, binary.LittleEndian, uint16(len(contentType)))
		w.buffer.WriteString(contentType)
		
		binary.Write(w.buffer, binary.LittleEndian, uint32(0))
		binary.Write(w.buffer, binary.LittleEndian, time.Now().UnixNano())
	}
	
	// Get current file position
	offset, err := w.file.Seek(0, os.SEEK_END)
	if err != nil {
		logger.Error("Failed to seek to end: %v", err)
		return err
	}
	
	logger.Debug("Writing %d bytes at offset %d", w.buffer.Len(), offset)
	
	// Write to file
	n, err := w.file.Write(w.buffer.Bytes())
	if err != nil {
		logger.Error("Failed to write data: %v", err)
		return err
	}
	
	logger.Debug("Actually wrote %d bytes", n)
	
	// Update index
	entry := &IndexEntry{
		Offset: uint64(offset),
		Size:   uint32(n),
	}
	// Clear the EntityID array first to avoid garbage
	for i := range entry.EntityID {
		entry.EntityID[i] = 0
	}
	// Copy the ID, ensuring we don't exceed the array size
	idBytes := []byte(entity.ID)
	if len(idBytes) > len(entry.EntityID) {
		idBytes = idBytes[:len(entry.EntityID)]
	}
	copy(entry.EntityID[:], idBytes)
	w.index[entity.ID] = entry
	
	logger.Debug("Added index entry for %s: offset=%d, size=%d", entity.ID, entry.Offset, entry.Size)
	
	// Update header
	w.header.EntityCount++
	w.header.FileSize = uint64(offset) + uint64(n)
	w.header.LastModified = time.Now().Unix()
	
	logger.Debug("Updated header: EntityCount=%d, FileSize=%d", w.header.EntityCount, w.header.FileSize)
	
	// Write updated header back to file
	currentPos, err := w.file.Seek(0, os.SEEK_CUR)
	if err != nil {
		logger.Error("Failed to get current position: %v", err)
		return err
	}
	logger.Debug("Current position: %d", currentPos)
	
	// Seek to start to write header
	_, err = w.file.Seek(0, os.SEEK_SET)
	if err != nil {
		logger.Error("Failed to seek to start: %v", err)
		return err
	}
	
	logger.Debug("Writing updated header to file")
	if err := w.writeHeader(); err != nil {
		logger.Error("Failed to write header: %v", err)
		return err
	}
	
	// Seek back to original position
	_, err = w.file.Seek(currentPos, os.SEEK_SET)
	if err != nil {
		logger.Error("Failed to seek back to position %d: %v", currentPos, err)
		return err
	}
	
	// Sync data to disk
	if err := w.file.Sync(); err != nil {
		logger.Error("Failed to sync file: %v", err)
		return err
	}
	
	logger.Debug("Entity write completed successfully")
	
	return nil
}

// getFilePosition returns current file position for debugging
func (w *Writer) getFilePosition() int64 {
	pos, _ := w.file.Seek(0, os.SEEK_CUR)
	return pos
}

// Close flushes and closes the writer
func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	logger.Debug("Close called, EntityCount=%d, FileSize=%d", w.header.EntityCount, w.header.FileSize)
	
	// Write tag dictionary
	dictOffset, err := w.file.Seek(0, os.SEEK_END)
	if err != nil {
		logger.Error("Failed to seek for dictionary: %v", err)
		return err
	}
	logger.Debug("Writing tag dictionary at offset %d", dictOffset)
	
	dictBuf := new(bytes.Buffer)
	w.tagDict.Write(dictBuf)
	n, err := w.file.Write(dictBuf.Bytes())
	if err != nil {
		logger.Error("Failed to write tag dictionary: %v", err)
		return err
	}
	logger.Debug("Wrote %d bytes for tag dictionary", n)
	
	// Write index
	indexOffset, err := w.file.Seek(0, os.SEEK_END)
	if err != nil {
		logger.Error("Failed to seek for index: %v", err)
		return err
	}
	logger.Debug("Writing index at offset %d with %d entries", indexOffset, len(w.index))
	
	for id, entry := range w.index {
		logger.Debug("Writing index entry for %s: offset=%d, size=%d", id, entry.Offset, entry.Size)
		binary.Write(w.file, binary.LittleEndian, entry.EntityID)
		binary.Write(w.file, binary.LittleEndian, entry.Offset)
		binary.Write(w.file, binary.LittleEndian, entry.Size)
		binary.Write(w.file, binary.LittleEndian, entry.Flags)
	}
	
	// Update header
	w.header.TagDictOffset = uint64(dictOffset)
	w.header.TagDictSize = uint64(dictBuf.Len())
	w.header.EntityIndexOffset = uint64(indexOffset)
	w.header.EntityIndexSize = uint64(len(w.index) * IndexEntrySize)
	
	logger.Debug("Updated header: TagDictOffset=%d, TagDictSize=%d, EntityIndexOffset=%d, EntityIndexSize=%d",
		w.header.TagDictOffset, w.header.TagDictSize, w.header.EntityIndexOffset, w.header.EntityIndexSize)
	
	// Rewrite header at beginning
	_, err = w.file.Seek(0, os.SEEK_SET)
	if err != nil {
		logger.Error("Failed to seek to start: %v", err)
		return err
	}
	
	logger.Debug("Rewriting header")
	if err := w.writeHeader(); err != nil {
		logger.Error("Failed to write header: %v", err)
		return err
	}
	
	// Ensure all data is written
	if err := w.file.Sync(); err != nil {
		logger.Error("Failed to sync file: %v", err)
		return err
	}
	
	logger.Debug("Close completed successfully")
	
	// Don't close the file if used as singleton
	return nil
}

func (w *Writer) writeHeader() error {
	logger.Debug("writeHeader called - EntityCount=%d, FileSize=%d", w.header.EntityCount, w.header.FileSize)
	err := w.header.Write(w.file)
	if err != nil {
		logger.Error("Failed to write header: %v", err)
		return err
	}
	logger.Debug("Header written successfully")
	return nil
}

func (w *Writer) readExisting() error {
	logger.Debug("readExisting called")
	
	// Read header
	if err := w.header.Read(w.file); err != nil {
		logger.Error("Failed to read header: %v", err)
		return err
	}
	logger.Debug("Read header: EntityCount=%d, FileSize=%d", w.header.EntityCount, w.header.FileSize)
	
	// Read tag dictionary
	logger.Debug("Reading tag dictionary from offset %d", w.header.TagDictOffset)
	w.file.Seek(int64(w.header.TagDictOffset), os.SEEK_SET)
	if err := w.tagDict.Read(w.file); err != nil {
		logger.Error("Failed to read tag dictionary: %v", err)
		return err
	}
	logger.Debug("Loaded tag dictionary")
	
	// Read index
	logger.Debug("Reading index from offset %d, expecting %d entries", w.header.EntityIndexOffset, w.header.EntityCount)
	w.file.Seek(int64(w.header.EntityIndexOffset), os.SEEK_SET)
	w.index = make(map[string]*IndexEntry)
	
	for i := uint64(0); i < w.header.EntityCount; i++ {
		entry := &IndexEntry{}
		if err := binary.Read(w.file, binary.LittleEndian, &entry.EntityID); err != nil {
			logger.Error("Failed to read index entry %d: %v", i, err)
			break
		}
		if err := binary.Read(w.file, binary.LittleEndian, &entry.Offset); err != nil {
			logger.Error("Failed to read offset for entry %d: %v", i, err)
			break
		}
		if err := binary.Read(w.file, binary.LittleEndian, &entry.Size); err != nil {
			logger.Error("Failed to read size for entry %d: %v", i, err)
			break
		}
		if err := binary.Read(w.file, binary.LittleEndian, &entry.Flags); err != nil {
			logger.Error("Failed to read flags for entry %d: %v", i, err)
			break
		}
		
		// Convert ID to string, handling any null bytes or garbage
		id := string(bytes.TrimRight(entry.EntityID[:], "\x00"))
		// Skip empty IDs
		if id == "" {
			logger.Debug("Skipping empty index entry %d", i)
			continue
		}
		w.index[id] = entry
		logger.Debug("Loaded index entry %d: ID=%s, Offset=%d, Size=%d", i, id, entry.Offset, entry.Size)
	}
	
	logger.Debug("Loaded %d index entries", len(w.index))
	
	// Seek to end for writing new data
	w.file.Seek(int64(w.header.FileSize), os.SEEK_SET)
	
	return nil
}