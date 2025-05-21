package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// FileHeader represents the EBF file header
type FileHeader struct {
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

// IndexEntry represents an entry in the entity index
type IndexEntry struct {
	EntityID [36]byte
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

// Magic number "EBDF" (EntityDB Format)
const MagicNumber uint32 = 0x45424446

// BinaryAnalyzer analyzes binary entity files
type BinaryAnalyzer struct {
	filename           string
	outputDir          string
	extractEntities    bool
	validateChecksums  bool
	repairMode         bool
	verbose            bool
	
	file               *os.File
	header             FileHeader
	index              map[string]IndexEntry
	tagDict            map[uint32]string
	
	// Statistics
	stats              map[string]interface{}
	entitySizes        []uint32
	corruptEntities    []string
	suspiciousEntities []string
}

// NewBinaryAnalyzer creates a new binary analyzer
func NewBinaryAnalyzer(filename string, outputDir string, options map[string]bool) (*BinaryAnalyzer, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	
	ba := &BinaryAnalyzer{
		filename:           filename,
		outputDir:          outputDir,
		extractEntities:    options["extract"],
		validateChecksums:  options["validate"],
		repairMode:         options["repair"],
		verbose:            options["verbose"],
		file:               file,
		index:              make(map[string]IndexEntry),
		tagDict:            make(map[uint32]string),
		stats:              make(map[string]interface{}),
		entitySizes:        make([]uint32, 0),
		corruptEntities:    make([]string, 0),
		suspiciousEntities: make([]string, 0),
	}
	
	return ba, nil
}

// Analyze performs a complete analysis of the binary file
func (ba *BinaryAnalyzer) Analyze() error {
	startTime := time.Now()
	
	fmt.Printf("Analyzing binary file: %s\n", ba.filename)
	
	// Create output directory if extracting entities
	if ba.extractEntities {
		if err := os.MkdirAll(ba.outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %v", err)
		}
	}
	
	// Read and validate header
	if err := ba.readHeader(); err != nil {
		return err
	}
	
	// Read tag dictionary
	if err := ba.readTagDictionary(); err != nil {
		return err
	}
	
	// Read entity index
	if err := ba.readEntityIndex(); err != nil {
		return err
	}
	
	// Validate and analyze entities
	if err := ba.analyzeEntities(); err != nil {
		return err
	}
	
	// Calculate statistics
	ba.calculateStatistics()
	
	// Print summary
	ba.printSummary()
	
	fmt.Printf("Analysis completed in %.2f seconds\n", time.Since(startTime).Seconds())
	
	return nil
}

// readHeader reads and validates the file header
func (ba *BinaryAnalyzer) readHeader() error {
	fmt.Println("Reading file header...")
	
	// Get file size
	info, err := ba.file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}
	fileSize := info.Size()
	
	// Read header
	if err := binary.Read(ba.file, binary.LittleEndian, &ba.header); err != nil {
		return fmt.Errorf("failed to read header: %v", err)
	}
	
	// Validate magic number
	if ba.header.Magic != MagicNumber {
		return fmt.Errorf("invalid magic number: 0x%x (expected 0x%x)", ba.header.Magic, MagicNumber)
	}
	
	// Validate file size
	if uint64(fileSize) != ba.header.FileSize {
		fmt.Printf("Warning: File size mismatch: %d bytes on disk, %d bytes in header\n", fileSize, ba.header.FileSize)
	}
	
	// Print header info
	fmt.Printf("Header:\n")
	fmt.Printf("  Magic: 0x%x\n", ba.header.Magic)
	fmt.Printf("  Version: %d\n", ba.header.Version)
	fmt.Printf("  File Size: %d bytes\n", ba.header.FileSize)
	fmt.Printf("  Entity Count: %d\n", ba.header.EntityCount)
	fmt.Printf("  Last Modified: %s\n", time.Unix(0, ba.header.LastModified).Format(time.RFC3339))
	fmt.Printf("  Tag Dictionary: Offset=%d, Size=%d\n", ba.header.TagDictOffset, ba.header.TagDictSize)
	fmt.Printf("  Entity Index: Offset=%d, Size=%d\n", ba.header.EntityIndexOffset, ba.header.EntityIndexSize)
	
	return nil
}

// readTagDictionary reads the tag dictionary
func (ba *BinaryAnalyzer) readTagDictionary() error {
	if ba.header.TagDictOffset == 0 || ba.header.TagDictSize == 0 {
		fmt.Println("No tag dictionary found")
		return nil
	}
	
	fmt.Println("Reading tag dictionary...")
	
	// Seek to tag dictionary
	if _, err := ba.file.Seek(int64(ba.header.TagDictOffset), io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek to tag dictionary: %v", err)
	}
	
	// Read dictionary count
	var count uint32
	if err := binary.Read(ba.file, binary.LittleEndian, &count); err != nil {
		return fmt.Errorf("failed to read tag dictionary count: %v", err)
	}
	
	fmt.Printf("Tag dictionary contains %d entries\n", count)
	
	// Read dictionary entries
	for i := uint32(0); i < count; i++ {
		var id uint32
		var length uint16
		
		if err := binary.Read(ba.file, binary.LittleEndian, &id); err != nil {
			return fmt.Errorf("failed to read tag ID: %v", err)
		}
		
		if err := binary.Read(ba.file, binary.LittleEndian, &length); err != nil {
			return fmt.Errorf("failed to read tag length: %v", err)
		}
		
		tag := make([]byte, length)
		if _, err := io.ReadFull(ba.file, tag); err != nil {
			return fmt.Errorf("failed to read tag: %v", err)
		}
		
		ba.tagDict[id] = string(tag)
		
		if ba.verbose {
			fmt.Printf("  Tag %d: %s\n", id, string(tag))
		}
	}
	
	return nil
}

// readEntityIndex reads the entity index
func (ba *BinaryAnalyzer) readEntityIndex() error {
	if ba.header.EntityIndexOffset == 0 || ba.header.EntityCount == 0 {
		fmt.Println("No entity index found")
		return nil
	}
	
	fmt.Println("Reading entity index...")
	
	// Seek to entity index
	if _, err := ba.file.Seek(int64(ba.header.EntityIndexOffset), io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek to entity index: %v", err)
	}
	
	// Read index entries
	for i := uint64(0); i < ba.header.EntityCount; i++ {
		var entry IndexEntry
		
		if err := binary.Read(ba.file, binary.LittleEndian, &entry); err != nil {
			return fmt.Errorf("failed to read index entry %d: %v", i, err)
		}
		
		entityID := string(entry.EntityID[:])
		ba.index[entityID] = entry
		ba.entitySizes = append(ba.entitySizes, entry.Size)
		
		if ba.verbose && i < 10 {
			fmt.Printf("  Entity %d: ID=%s, Offset=%d, Size=%d\n", i, entityID, entry.Offset, entry.Size)
		}
	}
	
	return nil
}

// analyzeEntities analyzes all entities in the file
func (ba *BinaryAnalyzer) analyzeEntities() error {
	fmt.Printf("Analyzing %d entities...\n", len(ba.index))
	
	// Sort entities by offset for sequential access
	type entityOffset struct {
		ID     string
		Offset uint64
	}
	
	entities := make([]entityOffset, 0, len(ba.index))
	for id, entry := range ba.index {
		entities = append(entities, entityOffset{ID: id, Offset: entry.Offset})
	}
	
	sort.Slice(entities, func(i, j int) bool {
		return entities[i].Offset < entities[j].Offset
	})
	
	// Analyze each entity
	analysisProgressInterval := len(entities) / 20 // 5% increments
	if analysisProgressInterval == 0 {
		analysisProgressInterval = 1
	}
	
	for i, entity := range entities {
		if i % analysisProgressInterval == 0 {
			fmt.Printf("  Progress: %.1f%% (%d/%d)\n", float64(i) * 100.0 / float64(len(entities)), i, len(entities))
		}
		
		entry := ba.index[entity.ID]
		
		// Skip if entry size is too large (likely corrupted)
		if entry.Size > 100 * 1024 * 1024 { // 100 MB sanity check
			fmt.Printf("Warning: Entity %s has a suspicious size of %d bytes (skipping)\n", entity.ID, entry.Size)
			ba.suspiciousEntities = append(ba.suspiciousEntities, entity.ID)
			continue
		}
		
		// Read entity data
		data, err := ba.readEntityData(entity.ID, entry)
		if err != nil {
			fmt.Printf("Error: Failed to read entity %s: %v\n", entity.ID, err)
			ba.corruptEntities = append(ba.corruptEntities, entity.ID)
			continue
		}
		
		// Validate entity data
		if ba.validateChecksums {
			if err := ba.validateEntity(entity.ID, data); err != nil {
				fmt.Printf("Error: Entity %s failed validation: %v\n", entity.ID, err)
				ba.corruptEntities = append(ba.corruptEntities, entity.ID)
				continue
			}
		}
		
		// Extract entity if requested
		if ba.extractEntities {
			if err := ba.extractEntity(entity.ID, data); err != nil {
				fmt.Printf("Error: Failed to extract entity %s: %v\n", entity.ID, err)
			}
		}
	}
	
	return nil
}

// readEntityData reads raw entity data from the file
func (ba *BinaryAnalyzer) readEntityData(entityID string, entry IndexEntry) ([]byte, error) {
	// Seek to entity data
	if _, err := ba.file.Seek(int64(entry.Offset), io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to seek to entity data: %v", err)
	}
	
	// Read entity data
	data := make([]byte, entry.Size)
	if _, err := io.ReadFull(ba.file, data); err != nil {
		return nil, fmt.Errorf("failed to read entity data: %v", err)
	}
	
	return data, nil
}

// validateEntity performs basic validation on entity data
func (ba *BinaryAnalyzer) validateEntity(entityID string, data []byte) error {
	if len(data) < 16 {
		return fmt.Errorf("entity data too small: %d bytes", len(data))
	}
	
	// Read entity header
	var header EntityHeader
	reader := strings.NewReader(string(data[:16]))
	if err := binary.Read(reader, binary.LittleEndian, &header); err != nil {
		return fmt.Errorf("failed to read entity header: %v", err)
	}
	
	// Basic validation
	if header.TagCount == 0 {
		return fmt.Errorf("entity has no tags")
	}
	
	// Skip to content for more detail
	tagSectionSize := 2 + int(header.TagCount) * 4 // 2-byte count + 4 bytes per tag ID
	if len(data) < tagSectionSize {
		return fmt.Errorf("entity data truncated in tag section")
	}
	
	// Calculate simple checksum for integrity checking
	checksum := crc32.ChecksumIEEE(data)
	
	if ba.verbose {
		// Examine first few bytes of content
		contentOffset := tagSectionSize
		contentPreviewSize := 32
		if contentOffset+contentPreviewSize > len(data) {
			contentPreviewSize = len(data) - contentOffset
		}
		
		fmt.Printf("Entity %s: %d tags, checksum: %08x\n", entityID, header.TagCount, checksum)
		if contentPreviewSize > 0 {
			fmt.Printf("  Content preview (hex): % x\n", data[contentOffset:contentOffset+contentPreviewSize])
		}
	}
	
	return nil
}

// extractEntity extracts an entity to a file
func (ba *BinaryAnalyzer) extractEntity(entityID string, data []byte) error {
	outPath := filepath.Join(ba.outputDir, fmt.Sprintf("entity_%s.bin", entityID))
	
	return os.WriteFile(outPath, data, 0644)
}

// calculateStatistics calculates statistics about the entity file
func (ba *BinaryAnalyzer) calculateStatistics() {
	// Basic statistics
	ba.stats["file_size"] = ba.header.FileSize
	ba.stats["entity_count"] = ba.header.EntityCount
	ba.stats["tag_count"] = len(ba.tagDict)
	ba.stats["corrupt_entities"] = len(ba.corruptEntities)
	ba.stats["suspicious_entities"] = len(ba.suspiciousEntities)
	
	// Entity size statistics
	if len(ba.entitySizes) > 0 {
		sort.Slice(ba.entitySizes, func(i, j int) bool {
			return ba.entitySizes[i] < ba.entitySizes[j]
		})
		
		var totalSize uint64
		for _, size := range ba.entitySizes {
			totalSize += uint64(size)
		}
		
		ba.stats["entity_size_min"] = ba.entitySizes[0]
		ba.stats["entity_size_max"] = ba.entitySizes[len(ba.entitySizes)-1]
		ba.stats["entity_size_avg"] = float64(totalSize) / float64(len(ba.entitySizes))
		ba.stats["entity_size_median"] = ba.entitySizes[len(ba.entitySizes)/2]
		ba.stats["entity_size_total"] = totalSize
		
		// Calculate size distribution
		var small, medium, large, xlarge int
		for _, size := range ba.entitySizes {
			switch {
			case size < 1024: // < 1 KB
				small++
			case size < 10*1024: // < 10 KB
				medium++
			case size < 100*1024: // < 100 KB
				large++
			default: // >= 100 KB
				xlarge++
			}
		}
		
		ba.stats["entity_size_distribution"] = map[string]int{
			"small (<1KB)":      small,
			"medium (1-10KB)":   medium, 
			"large (10-100KB)":  large,
			"xlarge (>=100KB)":  xlarge,
		}
	}
	
	// Integrity statistics
	ba.stats["corrupt_percentage"] = float64(len(ba.corruptEntities)) * 100.0 / float64(ba.header.EntityCount)
	ba.stats["suspicious_percentage"] = float64(len(ba.suspiciousEntities)) * 100.0 / float64(ba.header.EntityCount)
	
	// Save stats to a file
	statsJSON, _ := json.MarshalIndent(ba.stats, "", "  ")
	statsPath := filepath.Join(ba.outputDir, "analysis_stats.json")
	os.WriteFile(statsPath, statsJSON, 0644)
}

// printSummary prints a summary of the analysis
func (ba *BinaryAnalyzer) printSummary() {
	fmt.Println("\nAnalysis Summary:")
	fmt.Printf("  File Size: %d bytes (%.2f MB)\n", ba.header.FileSize, float64(ba.header.FileSize) / 1024.0 / 1024.0)
	fmt.Printf("  Entity Count: %d\n", ba.header.EntityCount)
	fmt.Printf("  Tag Dictionary: %d entries\n", len(ba.tagDict))
	
	if len(ba.entitySizes) > 0 {
		fmt.Printf("  Entity Size Statistics:\n")
		fmt.Printf("    Minimum: %d bytes\n", ba.stats["entity_size_min"])
		fmt.Printf("    Maximum: %d bytes\n", ba.stats["entity_size_max"])
		fmt.Printf("    Average: %.2f bytes\n", ba.stats["entity_size_avg"])
		fmt.Printf("    Median: %d bytes\n", ba.stats["entity_size_median"])
		
		dist := ba.stats["entity_size_distribution"].(map[string]int)
		fmt.Printf("    Size Distribution:\n")
		fmt.Printf("      Small (<1KB): %d (%.1f%%)\n", dist["small (<1KB)"], float64(dist["small (<1KB)"]) * 100.0 / float64(ba.header.EntityCount))
		fmt.Printf("      Medium (1-10KB): %d (%.1f%%)\n", dist["medium (1-10KB)"], float64(dist["medium (1-10KB)"]) * 100.0 / float64(ba.header.EntityCount))
		fmt.Printf("      Large (10-100KB): %d (%.1f%%)\n", dist["large (10-100KB)"], float64(dist["large (10-100KB)"]) * 100.0 / float64(ba.header.EntityCount))
		fmt.Printf("      X-Large (>=100KB): %d (%.1f%%)\n", dist["xlarge (>=100KB)"], float64(dist["xlarge (>=100KB)"]) * 100.0 / float64(ba.header.EntityCount))
	}
	
	if len(ba.corruptEntities) > 0 {
		fmt.Printf("  Corrupt Entities: %d (%.1f%%)\n", len(ba.corruptEntities), ba.stats["corrupt_percentage"])
		if ba.verbose {
			fmt.Printf("    Corrupt Entity IDs:\n")
			for i, id := range ba.corruptEntities {
				if i < 10 { // Show at most 10 IDs to avoid flooding output
					fmt.Printf("      %s\n", id)
				} else {
					fmt.Printf("      ... and %d more\n", len(ba.corruptEntities) - 10)
					break
				}
			}
		}
	} else {
		fmt.Printf("  No corrupt entities found\n")
	}
	
	if len(ba.suspiciousEntities) > 0 {
		fmt.Printf("  Suspicious Entities: %d (%.1f%%)\n", len(ba.suspiciousEntities), ba.stats["suspicious_percentage"])
	}
	
	fmt.Println("\nAnalysis results saved to:", filepath.Join(ba.outputDir, "analysis_stats.json"))
}

// Close closes the analyzer
func (ba *BinaryAnalyzer) Close() error {
	return ba.file.Close()
}

func main() {
	// Parse command-line flags
	filename := flag.String("file", "", "Path to the binary entity file")
	outputDir := flag.String("output", "./analysis_output", "Directory to output analysis results")
	extract := flag.Bool("extract", false, "Extract entities to individual files")
	validate := flag.Bool("validate", true, "Validate entity data")
	repair := flag.Bool("repair", false, "Attempt to repair corrupted entities")
	verbose := flag.Bool("verbose", false, "Enable verbose output")
	
	flag.Parse()
	
	if *filename == "" {
		fmt.Println("Error: -file parameter is required")
		flag.Usage()
		os.Exit(1)
	}
	
	options := map[string]bool{
		"extract":  *extract,
		"validate": *validate,
		"repair":   *repair,
		"verbose":  *verbose,
	}
	
	analyzer, err := NewBinaryAnalyzer(*filename, *outputDir, options)
	if err != nil {
		fmt.Printf("Error creating analyzer: %v\n", err)
		os.Exit(1)
	}
	defer analyzer.Close()
	
	if err := analyzer.Analyze(); err != nil {
		fmt.Printf("Analysis failed: %v\n", err)
		os.Exit(1)
	}
}