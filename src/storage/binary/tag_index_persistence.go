package binary

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	
	"entitydb/logger"
)

const (
	TagIndexMagic   = "TDX2"
	TagIndexVersion = 2
)

// SaveTagIndex saves the tag index in a simple, reliable format
func SaveTagIndex(dataFile string, tagIndex map[string][]string) error {
	// Create index filename
	dir := filepath.Dir(dataFile)
	base := filepath.Base(dataFile)
	idxFile := filepath.Join(dir, strings.TrimSuffix(base, ".ebf") + ".idx")
	tempFile := idxFile + ".tmp"
	
	// Create temp file
	file, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create index file: %w", err)
	}
	
	// Write magic and version
	if _, err := file.Write([]byte(TagIndexMagic)); err != nil {
		file.Close()
		os.Remove(tempFile)
		return err
	}
	
	if err := binary.Write(file, binary.LittleEndian, uint16(TagIndexVersion)); err != nil {
		file.Close()
		os.Remove(tempFile)
		return err
	}
	
	// Write entry count
	if err := binary.Write(file, binary.LittleEndian, uint64(len(tagIndex))); err != nil {
		file.Close()
		os.Remove(tempFile)
		return err
	}
	
	// Write each tag entry
	for tag, entityIDs := range tagIndex {
		// Write tag length and tag
		tagBytes := []byte(tag)
		if err := binary.Write(file, binary.LittleEndian, uint32(len(tagBytes))); err != nil {
			file.Close()
			os.Remove(tempFile)
			return err
		}
		if _, err := file.Write(tagBytes); err != nil {
			file.Close()
			os.Remove(tempFile)
			return err
		}
		
		// Write entity count and IDs
		if err := binary.Write(file, binary.LittleEndian, uint32(len(entityIDs))); err != nil {
			file.Close()
			os.Remove(tempFile)
			return err
		}
		
		for _, id := range entityIDs {
			// Write ID length and ID (to handle variable length IDs)
			idBytes := []byte(id)
			if err := binary.Write(file, binary.LittleEndian, uint16(len(idBytes))); err != nil {
				file.Close()
				os.Remove(tempFile)
				return err
			}
			if _, err := file.Write(idBytes); err != nil {
				file.Close()
				os.Remove(tempFile)
				return err
			}
		}
	}
	
	// Close and rename
	if err := file.Close(); err != nil {
		os.Remove(tempFile)
		return err
	}
	
	if err := os.Rename(tempFile, idxFile); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("failed to rename index file: %w", err)
	}
	
	logger.Info("Tag index saved to %s (%d tags)", idxFile, len(tagIndex))
	return nil
}

// LoadTagIndex loads the tag index from disk
func LoadTagIndex(dataFile string) (map[string][]string, error) {
	// Create index filename
	dir := filepath.Dir(dataFile)
	base := filepath.Base(dataFile)
	idxFile := filepath.Join(dir, strings.TrimSuffix(base, ".ebf") + ".idx")
	
	// Open file
	file, err := os.Open(idxFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("index file does not exist: %s", idxFile)
		}
		return nil, err
	}
	defer file.Close()
	
	// Read and verify magic
	magic := make([]byte, 4)
	if _, err := io.ReadFull(file, magic); err != nil {
		return nil, fmt.Errorf("failed to read magic: %w", err)
	}
	if string(magic) != TagIndexMagic {
		return nil, fmt.Errorf("invalid magic: %s", magic)
	}
	
	// Read version
	var version uint16
	if err := binary.Read(file, binary.LittleEndian, &version); err != nil {
		return nil, fmt.Errorf("failed to read version: %w", err)
	}
	if version != TagIndexVersion {
		return nil, fmt.Errorf("unsupported version: %d", version)
	}
	
	// Read entry count
	var entryCount uint64
	if err := binary.Read(file, binary.LittleEndian, &entryCount); err != nil {
		return nil, fmt.Errorf("failed to read entry count: %w", err)
	}
	
	// Read entries
	tagIndex := make(map[string][]string)
	startTime := time.Now()
	
	for i := uint64(0); i < entryCount; i++ {
		// Read tag length
		var tagLen uint32
		if err := binary.Read(file, binary.LittleEndian, &tagLen); err != nil {
			return nil, fmt.Errorf("failed to read tag length: %w", err)
		}
		
		// Read tag
		tagBytes := make([]byte, tagLen)
		if _, err := io.ReadFull(file, tagBytes); err != nil {
			return nil, fmt.Errorf("failed to read tag: %w", err)
		}
		tag := string(tagBytes)
		
		// Read entity count
		var entityCount uint32
		if err := binary.Read(file, binary.LittleEndian, &entityCount); err != nil {
			return nil, fmt.Errorf("failed to read entity count: %w", err)
		}
		
		// Read entity IDs
		entityIDs := make([]string, entityCount)
		for j := uint32(0); j < entityCount; j++ {
			// Read ID length
			var idLen uint16
			if err := binary.Read(file, binary.LittleEndian, &idLen); err != nil {
				return nil, fmt.Errorf("failed to read ID length: %w", err)
			}
			
			// Read ID
			idBytes := make([]byte, idLen)
			if _, err := io.ReadFull(file, idBytes); err != nil {
				return nil, fmt.Errorf("failed to read ID: %w", err)
			}
			entityIDs[j] = string(idBytes)
		}
		
		tagIndex[tag] = entityIDs
		
		// Log progress
		if (i+1) % 10000 == 0 {
			logger.Debug("Loaded %d/%d tag entries", i+1, entryCount)
		}
	}
	
	logger.Info("Loaded %d tags from index in %v", len(tagIndex), time.Since(startTime))
	return tagIndex, nil
}