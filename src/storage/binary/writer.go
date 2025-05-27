package binary

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"entitydb/models"
	"fmt"
	"os"
	"sort"
	"strings"
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

// getFilePosition returns the current file position
func (w *Writer) getFilePosition() int64 {
	pos, err := w.file.Seek(0, os.SEEK_CUR)
	if err != nil {
		return -1
	}
	return pos
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
	// Start operation tracking
	op := models.StartOperation(models.OpTypeWrite, entity.ID, map[string]interface{}{
		"tags_count":    len(entity.Tags),
		"content_size":  len(entity.Content),
		"file_position": w.getFilePosition(),
		"entity_count":  w.header.EntityCount,
	})
	defer func() {
		if op != nil {
			op.Complete()
		}
	}()
	
	w.mu.Lock()
	defer w.mu.Unlock()
	
	// Log write intent
	logger.Info("[Writer] Starting write for entity %s (tags=%d, content=%d bytes)", 
		entity.ID, len(entity.Tags), len(entity.Content))
	
	// Validate entity
	if entity.ID == "" {
		err := fmt.Errorf("entity ID cannot be empty")
		op.Fail(err)
		logger.Error("[Writer] Validation failed: %v", err)
		return err
	}
	
	// Calculate content checksum before write
	contentChecksum := sha256.Sum256(entity.Content)
	op.SetMetadata("content_checksum", hex.EncodeToString(contentChecksum[:]))
	
	logger.Debug("[Writer] Entity %s content checksum: %x", entity.ID, contentChecksum)
	
	// Prepare entity data
	w.buffer.Reset()
	
	// Add checksum tag if not already present
	checksumTag := fmt.Sprintf("%d|checksum:sha256:%s", time.Now().UnixNano(), hex.EncodeToString(contentChecksum[:]))
	hasChecksum := false
	for _, tag := range entity.Tags {
		if strings.Contains(tag, "|checksum:sha256:") {
			hasChecksum = true
			break
		}
	}
	
	// Create a new tags slice with checksum if needed
	tags := entity.Tags
	if !hasChecksum {
		tags = append([]string{}, entity.Tags...)
		tags = append(tags, checksumTag)
		logger.Info("[Writer] Added checksum tag for entity %s: %s", entity.ID, checksumTag)
	}
	
	// Convert tags to IDs
	tagIDs := make([]uint32, len(tags))
	for i, tag := range tags {
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
		// Determine content type from entity tags or default to application/octet-stream
		contentType := "application/octet-stream" // Default
		logger.Debug("Entity %s has %d tags, checking for content type", entity.ID, len(entity.Tags))
		for _, tag := range entity.Tags {
			logger.Debug("Checking tag: %s", tag)
			if strings.HasPrefix(tag, "content:type:") {
				// Direct match (non-timestamped)
				contentType = strings.TrimPrefix(tag, "content:type:")
				logger.Debug("Found direct content type: %s", contentType)
				break
			} else if strings.Contains(tag, "|content:type:") {
				// Timestamped tag
				parts := strings.SplitN(tag, "|", 2)
				if len(parts) == 2 {
					tagPart := parts[1] // Use the part after timestamp
					if strings.HasPrefix(tagPart, "content:type:") {
						contentType = strings.TrimPrefix(tagPart, "content:type:")
						logger.Debug("Found timestamped content type: %s", contentType)
						break
					}
				}
			}
		}
		logger.Debug("Final content type for entity %s: %s", entity.ID, contentType)
		
		// For JSON content, store it directly without additional wrapping
		if contentType == "application/json" {
			// Content is already JSON, store it as-is
			binary.Write(w.buffer, binary.LittleEndian, uint16(len(contentType)))
			w.buffer.WriteString(contentType)
			binary.Write(w.buffer, binary.LittleEndian, uint32(len(entity.Content)))
			w.buffer.Write(entity.Content)
		} else {
			// For other content types, use application/octet-stream wrapper
			binary.Write(w.buffer, binary.LittleEndian, uint16(len("application/octet-stream")))
			w.buffer.WriteString("application/octet-stream")
			binary.Write(w.buffer, binary.LittleEndian, uint32(len(entity.Content)))
			w.buffer.Write(entity.Content)
		}
		
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
		op.Fail(err)
		logger.Error("[Writer] Failed to seek to end for entity %s: %v", entity.ID, err)
		return err
	}
	
	// Calculate buffer checksum
	bufferChecksum := sha256.Sum256(w.buffer.Bytes())
	op.SetMetadata("buffer_checksum", hex.EncodeToString(bufferChecksum[:]))
	op.SetMetadata("write_offset", offset)
	op.SetMetadata("buffer_size", w.buffer.Len())
	
	logger.Info("[Writer] Writing entity %s: %d bytes at offset %d", entity.ID, w.buffer.Len(), offset)
	
	// Write to file
	n, err := w.file.Write(w.buffer.Bytes())
	if err != nil {
		op.Fail(err)
		logger.Error("[Writer] Failed to write entity %s data: %v", entity.ID, err)
		return err
	}
	
	if n != w.buffer.Len() {
		err := fmt.Errorf("incomplete write: expected %d bytes, wrote %d", w.buffer.Len(), n)
		op.Fail(err)
		logger.Error("[Writer] %v for entity %s", err, entity.ID)
		return err
	}
	
	op.SetMetadata("bytes_written", n)
	logger.Info("[Writer] Successfully wrote %d bytes for entity %s", n, entity.ID)
	
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
	
	logger.Info("[Writer] Added index entry for %s: offset=%d, size=%d", entity.ID, entry.Offset, entry.Size)
	op.SetMetadata("index_offset", entry.Offset)
	op.SetMetadata("index_size", entry.Size)
	
	// Verify we can read back what we wrote
	verifyBuffer := make([]byte, n)
	if _, err := w.file.ReadAt(verifyBuffer, int64(offset)); err != nil {
		logger.Error("[Writer] Failed to verify write for entity %s: %v", entity.ID, err)
		// Don't fail the operation, but log the issue
	} else {
		verifyChecksum := sha256.Sum256(verifyBuffer)
		if hex.EncodeToString(verifyChecksum[:]) != hex.EncodeToString(bufferChecksum[:]) {
			logger.Error("[Writer] Checksum mismatch after write for entity %s", entity.ID)
		} else {
			logger.Debug("[Writer] Write verification successful for entity %s", entity.ID)
		}
	}
	
	// Update header
	w.header.EntityCount++
	w.header.FileSize = uint64(offset) + uint64(n)
	w.header.LastModified = time.Now().Unix()
	
	logger.Info("[Writer] Updated header: EntityCount=%d, FileSize=%d", w.header.EntityCount, w.header.FileSize)
	op.SetMetadata("final_entity_count", w.header.EntityCount)
	op.SetMetadata("final_file_size", w.header.FileSize)
	
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
	
	// Collect all entity IDs and sort them to ensure deterministic order
	entityIDs := make([]string, 0, len(w.index))
	for id := range w.index {
		entityIDs = append(entityIDs, id)
	}
	sort.Strings(entityIDs)
	
	// Write index entries in sorted order
	writtenCount := 0
	for _, id := range entityIDs {
		entry := w.index[id]
		logger.Debug("Writing index entry %d for %s: offset=%d, size=%d", writtenCount, id, entry.Offset, entry.Size)
		if err := binary.Write(w.file, binary.LittleEndian, entry.EntityID); err != nil {
			logger.Error("Failed to write EntityID for %s: %v", id, err)
			return err
		}
		if err := binary.Write(w.file, binary.LittleEndian, entry.Offset); err != nil {
			logger.Error("Failed to write Offset for %s: %v", id, err)
			return err
		}
		if err := binary.Write(w.file, binary.LittleEndian, entry.Size); err != nil {
			logger.Error("Failed to write Size for %s: %v", id, err)
			return err
		}
		if err := binary.Write(w.file, binary.LittleEndian, entry.Flags); err != nil {
			logger.Error("Failed to write Flags for %s: %v", id, err)
			return err
		}
		writtenCount++
	}
	logger.Info("Wrote %d index entries (header claims %d)", writtenCount, w.header.EntityCount)
	
	// Verify index count matches header
	if writtenCount != int(w.header.EntityCount) {
		logger.Error("Index entry count mismatch: wrote %d entries but header claims %d", writtenCount, w.header.EntityCount)
		// Update header to match actual count
		w.header.EntityCount = uint64(writtenCount)
		logger.Info("Updated header EntityCount to match actual index: %d", w.header.EntityCount)
	}
	
	// Update header
	w.header.TagDictOffset = uint64(dictOffset)
	w.header.TagDictSize = uint64(dictBuf.Len())
	w.header.EntityIndexOffset = uint64(indexOffset)
	w.header.EntityIndexSize = uint64(writtenCount * IndexEntrySize)
	
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