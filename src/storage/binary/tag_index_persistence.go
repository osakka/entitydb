package binary

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	
	"entitydb/logger"
)

const (
	TagIndexMagic   = "TIDX"
	TagIndexVersion = 1
	TagIndexFooter  = "ENDT"
)

// TagIndexHeader represents the header of a tag index file
type TagIndexHeader struct {
	Magic      [4]byte
	Version    uint16
	EntryCount uint64
	Checksum   [32]byte
}

// TagIndexWriter writes tag index to persistent storage
type TagIndexWriter struct {
	file     *os.File
	hasher   hash.Hash
	tempPath string
	finalPath string
}

// NewTagIndexWriter creates a new tag index writer
func NewTagIndexWriter(dataFile string) (*TagIndexWriter, error) {
	// Create index filename based on data file
	dir := filepath.Dir(dataFile)
	base := filepath.Base(dataFile)
	idxFile := filepath.Join(dir, strings.TrimSuffix(base, ".ebf") + ".idx")
	tempFile := idxFile + ".tmp"
	
	// Create temp file
	file, err := os.Create(tempFile)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp index file: %w", err)
	}
	
	return &TagIndexWriter{
		file:      file,
		hasher:    sha256.New(),
		tempPath:  tempFile,
		finalPath: idxFile,
	}, nil
}

// WriteHeader writes the index file header
func (w *TagIndexWriter) WriteHeader(entryCount uint64) error {
	header := TagIndexHeader{
		Version:    TagIndexVersion,
		EntryCount: entryCount,
	}
	copy(header.Magic[:], TagIndexMagic)
	
	// Write header (checksum will be updated at close)
	if err := binary.Write(w.file, binary.LittleEndian, header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}
	
	return nil
}

// WriteEntry writes a single tag entry with its entity IDs
func (w *TagIndexWriter) WriteEntry(tag string, entityIDs []string) error {
	// Write tag length
	tagBytes := []byte(tag)
	if err := binary.Write(w.file, binary.LittleEndian, uint32(len(tagBytes))); err != nil {
		return err
	}
	w.hasher.Write([]byte{byte(len(tagBytes) >> 24), byte(len(tagBytes) >> 16), byte(len(tagBytes) >> 8), byte(len(tagBytes))})
	
	// Write tag
	if _, err := w.file.Write(tagBytes); err != nil {
		return err
	}
	w.hasher.Write(tagBytes)
	
	// Write entity count
	if err := binary.Write(w.file, binary.LittleEndian, uint32(len(entityIDs))); err != nil {
		return err
	}
	w.hasher.Write([]byte{byte(len(entityIDs) >> 24), byte(len(entityIDs) >> 16), byte(len(entityIDs) >> 8), byte(len(entityIDs))})
	
	// Write entity IDs
	for _, id := range entityIDs {
		if _, err := w.file.WriteString(id); err != nil {
			return err
		}
		w.hasher.Write([]byte(id))
	}
	
	return nil
}

// Close finalizes the index file
func (w *TagIndexWriter) Close() error {
	// Write footer
	if _, err := w.file.Write([]byte(TagIndexFooter)); err != nil {
		w.file.Close()
		os.Remove(w.tempPath)
		return fmt.Errorf("failed to write footer: %w", err)
	}
	
	// Get checksum
	checksum := w.hasher.Sum(nil)
	
	// Seek back to header to update checksum
	if _, err := w.file.Seek(6, 0); err != nil { // Skip magic + version
		w.file.Close()
		os.Remove(w.tempPath)
		return fmt.Errorf("failed to seek to checksum: %w", err)
	}
	
	// Skip entry count
	if _, err := w.file.Seek(8, 1); err != nil {
		w.file.Close()
		os.Remove(w.tempPath)
		return fmt.Errorf("failed to seek past entry count: %w", err)
	}
	
	// Write checksum
	if _, err := w.file.Write(checksum); err != nil {
		w.file.Close()
		os.Remove(w.tempPath)
		return fmt.Errorf("failed to write checksum: %w", err)
	}
	
	// Close file
	if err := w.file.Close(); err != nil {
		os.Remove(w.tempPath)
		return fmt.Errorf("failed to close file: %w", err)
	}
	
	// Atomically rename temp file to final
	if err := os.Rename(w.tempPath, w.finalPath); err != nil {
		os.Remove(w.tempPath)
		return fmt.Errorf("failed to rename index file: %w", err)
	}
	
	logger.Info("Tag index persisted to %s", w.finalPath)
	return nil
}

// TagIndexReader reads tag index from persistent storage
type TagIndexReader struct {
	file   *os.File
	header TagIndexHeader
}

// NewTagIndexReader creates a new tag index reader
func NewTagIndexReader(dataFile string) (*TagIndexReader, error) {
	// Create index filename based on data file
	dir := filepath.Dir(dataFile)
	base := filepath.Base(dataFile)
	idxFile := filepath.Join(dir, strings.TrimSuffix(base, ".ebf") + ".idx")
	
	// Check if index file exists
	if _, err := os.Stat(idxFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("index file does not exist: %s", idxFile)
	}
	
	// Open file
	file, err := os.Open(idxFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open index file: %w", err)
	}
	
	reader := &TagIndexReader{file: file}
	
	// Read header
	if err := reader.readHeader(); err != nil {
		file.Close()
		return nil, err
	}
	
	return reader, nil
}

// readHeader reads and validates the index header
func (r *TagIndexReader) readHeader() error {
	if err := binary.Read(r.file, binary.LittleEndian, &r.header); err != nil {
		return fmt.Errorf("failed to read header: %w", err)
	}
	
	// Validate magic
	if string(r.header.Magic[:]) != TagIndexMagic {
		return fmt.Errorf("invalid magic number: %s", r.header.Magic)
	}
	
	// Validate version
	if r.header.Version != TagIndexVersion {
		return fmt.Errorf("unsupported version: %d", r.header.Version)
	}
	
	return nil
}

// ReadAllEntries reads all tag entries from the index
func (r *TagIndexReader) ReadAllEntries() (map[string][]string, error) {
	tagIndex := make(map[string][]string)
	hasher := sha256.New()
	
	startTime := time.Now()
	entriesRead := uint64(0)
	
	for entriesRead < r.header.EntryCount {
		// Read tag length
		var tagLen uint32
		if err := binary.Read(r.file, binary.LittleEndian, &tagLen); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("failed to read tag length: %w", err)
		}
		hasher.Write([]byte{byte(tagLen >> 24), byte(tagLen >> 16), byte(tagLen >> 8), byte(tagLen)})
		
		// Read tag
		tagBytes := make([]byte, tagLen)
		if _, err := io.ReadFull(r.file, tagBytes); err != nil {
			return nil, fmt.Errorf("failed to read tag: %w", err)
		}
		hasher.Write(tagBytes)
		tag := string(tagBytes)
		
		// Read entity count
		var entityCount uint32
		if err := binary.Read(r.file, binary.LittleEndian, &entityCount); err != nil {
			return nil, fmt.Errorf("failed to read entity count: %w", err)
		}
		hasher.Write([]byte{byte(entityCount >> 24), byte(entityCount >> 16), byte(entityCount >> 8), byte(entityCount)})
		
		// Read entity IDs
		entityIDs := make([]string, entityCount)
		for i := uint32(0); i < entityCount; i++ {
			idBytes := make([]byte, 36) // UUID length
			if _, err := io.ReadFull(r.file, idBytes); err != nil {
				return nil, fmt.Errorf("failed to read entity ID: %w", err)
			}
			hasher.Write(idBytes)
			entityIDs[i] = string(idBytes)
		}
		
		tagIndex[tag] = entityIDs
		entriesRead++
		
		// Log progress periodically
		if entriesRead % 10000 == 0 {
			logger.Debug("Loaded %d/%d tag entries", entriesRead, r.header.EntryCount)
		}
	}
	
	// Verify checksum
	computedChecksum := hasher.Sum(nil)
	if string(computedChecksum) != string(r.header.Checksum[:]) {
		return nil, fmt.Errorf("checksum mismatch")
	}
	
	// Read footer
	footer := make([]byte, 4)
	if _, err := io.ReadFull(r.file, footer); err != nil {
		return nil, fmt.Errorf("failed to read footer: %w", err)
	}
	if string(footer) != TagIndexFooter {
		return nil, fmt.Errorf("invalid footer: %s", footer)
	}
	
	logger.Info("Loaded %d tag entries from index in %v", entriesRead, time.Since(startTime))
	return tagIndex, nil
}

// Close closes the reader
func (r *TagIndexReader) Close() error {
	return r.file.Close()
}

// SaveTagIndex saves the current tag index to disk
func SaveTagIndex(dataFile string, tagIndex map[string][]string) error {
	writer, err := NewTagIndexWriter(dataFile)
	if err != nil {
		return err
	}
	defer writer.Close()
	
	// Write header
	if err := writer.WriteHeader(uint64(len(tagIndex))); err != nil {
		return err
	}
	
	// Write entries
	for tag, entityIDs := range tagIndex {
		if err := writer.WriteEntry(tag, entityIDs); err != nil {
			return fmt.Errorf("failed to write entry for tag %s: %w", tag, err)
		}
	}
	
	return nil
}

// LoadTagIndex loads tag index from disk
func LoadTagIndex(dataFile string) (map[string][]string, error) {
	reader, err := NewTagIndexReader(dataFile)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	
	return reader.ReadAllEntries()
}