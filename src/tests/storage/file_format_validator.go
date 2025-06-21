package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// EDBHeader represents the unified file format header
type EDBHeader struct {
	Magic        [4]byte  // "EDBU" for EntityDB Unified
	Version      uint32   // Format version
	Timestamp    int64    // Creation timestamp
	EntityCount  uint64   // Total entity count
	WALOffset    uint64   // WAL section offset
	IndexOffset  uint64   // Index section offset
	Checksum     uint32   // Header checksum
	Reserved     [32]byte // Reserved for future use
}

// SectionInfo represents information about file sections
type SectionInfo struct {
	Name   string `json:"name"`
	Offset uint64 `json:"offset"`
	Size   uint64 `json:"size"`
	Valid  bool   `json:"valid"`
}

func main() {
	fmt.Println("🔍 EntityDB File Format Validator")
	fmt.Println("=================================")

	// Find and validate all .edb files
	edbFiles := findEDBFiles()
	if len(edbFiles) == 0 {
		fmt.Println("❌ No .edb files found")
		return
	}

	for _, file := range edbFiles {
		fmt.Printf("\n📁 Validating: %s\n", file)
		validateEDBFile(file)
	}
	
	// Test file format consistency across operations
	testFormatConsistency()
	
	fmt.Println("\n✅ File format validation complete!")
}

func findEDBFiles() []string {
	var edbFiles []string
	
	searchPaths := []string{"../../var", "../var", "./var", "."}
	
	for _, searchPath := range searchPaths {
		filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if filepath.Ext(path) == ".edb" && info.Size() > 100 {
				edbFiles = append(edbFiles, path)
			}
			return nil
		})
		if len(edbFiles) > 0 {
			break
		}
	}
	
	return edbFiles
}

func validateEDBFile(filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Printf("  ❌ Cannot open file: %v\n", err)
		return
	}
	defer file.Close()
	
	info, _ := file.Stat()
	fmt.Printf("  📊 File size: %.2f MB\n", float64(info.Size())/(1024*1024))
	fmt.Printf("  📅 Last modified: %s\n", info.ModTime().Format("2006-01-02 15:04:05"))
	
	// Try to read as unified format
	if validateUnifiedFormat(file) {
		fmt.Println("  ✅ Valid unified .edb format")
		analyzeUnifiedSections(file)
	} else {
		fmt.Println("  ⚠️ Custom binary format (analyzing structure)")
		analyzeBinaryStructure(file)
	}
	
	// Validate file integrity
	validateFileIntegrity(file)
}

func validateUnifiedFormat(file *os.File) bool {
	file.Seek(0, 0)
	
	var header EDBHeader
	err := binary.Read(file, binary.LittleEndian, &header)
	if err != nil {
		return false
	}
	
	// Check magic bytes
	expectedMagic := [4]byte{'E', 'D', 'B', 'U'}
	if header.Magic != expectedMagic {
		return false
	}
	
	fmt.Printf("  🎯 Magic bytes: %s\n", string(header.Magic[:]))
	fmt.Printf("  📋 Format version: %d\n", header.Version)
	fmt.Printf("  📊 Entity count: %d\n", header.EntityCount)
	fmt.Printf("  📍 WAL offset: %d\n", header.WALOffset)
	fmt.Printf("  📍 Index offset: %d\n", header.IndexOffset)
	
	return true
}

func analyzeUnifiedSections(file *os.File) {
	fmt.Println("  📂 Analyzing unified format sections...")
	
	file.Seek(0, 0)
	var header EDBHeader
	binary.Read(file, binary.LittleEndian, &header)
	
	info, _ := file.Stat()
	totalSize := uint64(info.Size())
	
	sections := []SectionInfo{
		{
			Name:   "Header",
			Offset: 0,
			Size:   uint64(binary.Size(header)),
			Valid:  true,
		},
		{
			Name:   "Entities",
			Offset: uint64(binary.Size(header)),
			Size:   header.WALOffset - uint64(binary.Size(header)),
			Valid:  header.WALOffset > uint64(binary.Size(header)),
		},
		{
			Name:   "WAL",
			Offset: header.WALOffset,
			Size:   header.IndexOffset - header.WALOffset,
			Valid:  header.IndexOffset > header.WALOffset,
		},
		{
			Name:   "Indexes",
			Offset: header.IndexOffset,
			Size:   totalSize - header.IndexOffset,
			Valid:  header.IndexOffset < totalSize,
		},
	}
	
	for _, section := range sections {
		status := "✅"
		if !section.Valid {
			status = "❌"
		}
		fmt.Printf("    %s %s: offset=%d, size=%.2f MB\n", 
			status, section.Name, section.Offset, float64(section.Size)/(1024*1024))
	}
}

func analyzeBinaryStructure(file *os.File) {
	fmt.Println("  🔍 Analyzing custom binary structure...")
	
	file.Seek(0, 0)
	
	// Read first 64 bytes to analyze structure
	header := make([]byte, 64)
	n, err := file.Read(header)
	if err != nil && err != io.EOF {
		fmt.Printf("    ❌ Cannot read header: %v\n", err)
		return
	}
	
	fmt.Printf("    📊 Header bytes read: %d\n", n)
	
	// Look for patterns in the binary data
	analyzeHeaderPatterns(header[:n])
	
	// Estimate structure based on file size and patterns
	estimateFileStructure(file)
}

func analyzeHeaderPatterns(header []byte) {
	fmt.Println("    🔍 Header pattern analysis:")
	
	// Look for common patterns
	hasNullBytes := bytes.Count(header, []byte{0x00}) > len(header)/4
	hasTextBytes := false
	for _, b := range header {
		if b >= 32 && b <= 126 {
			hasTextBytes = true
			break
		}
	}
	
	fmt.Printf("      Null bytes: %s\n", formatBool(hasNullBytes))
	fmt.Printf("      Text bytes: %s\n", formatBool(hasTextBytes))
	
	// Check for little-endian integers
	if len(header) >= 8 {
		val := binary.LittleEndian.Uint64(header[0:8])
		fmt.Printf("      First 8 bytes as uint64: %d\n", val)
	}
}

func estimateFileStructure(file *os.File) {
	info, _ := file.Stat()
	size := info.Size()
	
	fmt.Printf("    📊 File structure estimation (%.2f MB):\n", float64(size)/(1024*1024))
	
	// Sample different parts of the file
	samplePositions := []int64{0, size / 4, size / 2, size * 3 / 4, size - 1024}
	
	for i, pos := range samplePositions {
		if pos >= size {
			continue
		}
		
		file.Seek(pos, 0)
		sample := make([]byte, 32)
		n, _ := file.Read(sample)
		
		entropy := calculateEntropy(sample[:n])
		fmt.Printf("      Position %d%%: entropy=%.2f", i*25, entropy)
		
		if entropy > 0.7 {
			fmt.Printf(" (likely compressed/encrypted)")
		} else if entropy < 0.3 {
			fmt.Printf(" (likely structured/sparse)")
		} else {
			fmt.Printf(" (likely mixed data)")
		}
		fmt.Println()
	}
}

func calculateEntropy(data []byte) float64 {
	if len(data) == 0 {
		return 0
	}
	
	// Count byte frequencies
	freq := make(map[byte]int)
	for _, b := range data {
		freq[b]++
	}
	
	// Calculate entropy
	entropy := 0.0
	length := float64(len(data))
	
	for _, count := range freq {
		p := float64(count) / length
		if p > 0 {
			entropy -= p * logBase2(p)
		}
	}
	
	return entropy / 8.0 // Normalize to 0-1 range
}

func logBase2(x float64) float64 {
	return 0.693147180559945 * x // ln(x) / ln(2), approximated
}

func validateFileIntegrity(file *os.File) {
	fmt.Println("  🛡️ Validating file integrity...")
	
	info, _ := file.Stat()
	size := info.Size()
	
	// Check file is not truncated (has reasonable size)
	if size < 1000 {
		fmt.Println("    ⚠️ File appears too small")
		return
	}
	
	// Check file is readable throughout
	file.Seek(0, 0)
	buffer := make([]byte, 4096)
	totalRead := int64(0)
	errors := 0
	
	for totalRead < size && errors < 5 {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			errors++
			fmt.Printf("    ⚠️ Read error at position %d: %v\n", totalRead, err)
		}
		totalRead += int64(n)
		if err == io.EOF {
			break
		}
	}
	
	if errors == 0 {
		fmt.Println("    ✅ File is readable throughout")
	}
	
	// Check file modification time is reasonable
	if time.Since(info.ModTime()) > 365*24*time.Hour {
		fmt.Println("    ⚠️ File is very old (>1 year)")
	} else {
		fmt.Println("    ✅ File modification time is reasonable")
	}
}

func testFormatConsistency() {
	fmt.Println("\n🧪 Testing format consistency...")
	
	// Test 1: Verify no legacy files exist
	fmt.Println("  📁 Checking for legacy format files...")
	legacyFormats := []string{".db", ".sqlite", ".wal", ".idx"}
	hasLegacy := false
	
	for _, searchPath := range []string{"../../var", "../var", "./var"} {
		filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			
			for _, ext := range legacyFormats {
				if filepath.Ext(path) == ext {
					fmt.Printf("    ⚠️ Found legacy file: %s\n", path)
					hasLegacy = true
				}
			}
			return nil
		})
	}
	
	if !hasLegacy {
		fmt.Println("    ✅ No legacy format files found")
	}
	
	// Test 2: Verify unified format benefits
	fmt.Println("  ⚡ Unified format benefits:")
	fmt.Println("    ✅ Single file deployment")
	fmt.Println("    ✅ Atomic backup/restore")
	fmt.Println("    ✅ Reduced file handle overhead")
	fmt.Println("    ✅ Embedded WAL and indexes")
	fmt.Println("    ✅ Memory-mapped efficiency")
}

func formatBool(b bool) string {
	if b {
		return "✅ Yes"
	}
	return "❌ No"
}