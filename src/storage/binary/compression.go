package binary

import (
	"bytes"
	"compress/gzip"
	"entitydb/logger"
	"fmt"
	"io"
)

// CompressionType defines the type of compression used
type CompressionType byte

const (
	CompressionNone CompressionType = 0
	CompressionGzip CompressionType = 1
	// Reserved for future: CompressionZstd CompressionType = 2
	
	// Compression threshold - only compress if content is larger than this
	CompressionThreshold = 1024 // 1KB
)

// CompressedContent represents content that may be compressed
type CompressedContent struct {
	Type       CompressionType
	Data       []byte
	OriginalSize int
}

// CompressContent compresses content if it's above the threshold
func CompressContent(content []byte) (*CompressedContent, error) {
	if len(content) < CompressionThreshold {
		// Don't compress small content
		return &CompressedContent{
			Type:         CompressionNone,
			Data:         content,
			OriginalSize: len(content),
		}, nil
	}
	
	// Use gzip compression
	var compressed bytes.Buffer
	gw := gzip.NewWriter(&compressed)
	
	n, err := gw.Write(content)
	if err != nil {
		return nil, fmt.Errorf("compression write failed: %w", err)
	}
	
	if err := gw.Close(); err != nil {
		return nil, fmt.Errorf("compression close failed: %w", err)
	}
	
	// Only use compression if it actually saves space
	if compressed.Len() >= len(content) {
		logger.Trace("Compression not beneficial for content of size %d (compressed: %d)", 
			len(content), compressed.Len())
		return &CompressedContent{
			Type:         CompressionNone,
			Data:         content,
			OriginalSize: len(content),
		}, nil
	}
	
	logger.Trace("Compressed %d bytes to %d bytes (%.1f%% reduction)", 
		n, compressed.Len(), float64(len(content)-compressed.Len())/float64(len(content))*100)
	
	return &CompressedContent{
		Type:         CompressionGzip,
		Data:         compressed.Bytes(),
		OriginalSize: len(content),
	}, nil
}

// DecompressContent decompresses content based on its type
func DecompressContent(cc *CompressedContent) ([]byte, error) {
	switch cc.Type {
	case CompressionNone:
		return cc.Data, nil
		
	case CompressionGzip:
		gr, err := gzip.NewReader(bytes.NewReader(cc.Data))
		if err != nil {
			return nil, fmt.Errorf("gzip reader creation failed: %w", err)
		}
		defer gr.Close()
		
		// Pre-allocate buffer based on original size if known
		var decompressed bytes.Buffer
		if cc.OriginalSize > 0 {
			decompressed.Grow(cc.OriginalSize)
		}
		
		_, err = io.Copy(&decompressed, gr)
		if err != nil {
			return nil, fmt.Errorf("decompression failed: %w", err)
		}
		
		result := decompressed.Bytes()
		if cc.OriginalSize > 0 && len(result) != cc.OriginalSize {
			logger.Warn("Decompressed size mismatch: expected %d, got %d", 
				cc.OriginalSize, len(result))
		}
		
		return result, nil
		
	default:
		return nil, fmt.Errorf("unsupported compression type: %d", cc.Type)
	}
}

// CompressWithPool compresses content using pooled buffers
func CompressWithPool(content []byte) (*CompressedContent, error) {
	if len(content) < CompressionThreshold {
		return &CompressedContent{
			Type:         CompressionNone,
			Data:         content,
			OriginalSize: len(content),
		}, nil
	}
	
	// Get buffer from pool
	compressed := GetSmallBuffer()
	defer PutSmallBuffer(compressed)
	compressed.Reset()
	
	gw := gzip.NewWriter(compressed)
	n, err := gw.Write(content)
	if err != nil {
		return nil, fmt.Errorf("compression write failed: %w", err)
	}
	
	if err := gw.Close(); err != nil {
		return nil, fmt.Errorf("compression close failed: %w", err)
	}
	
	// Check if compression is beneficial
	if compressed.Len() >= len(content) {
		return &CompressedContent{
			Type:         CompressionNone,
			Data:         content,
			OriginalSize: len(content),
		}, nil
	}
	
	// Make a copy of the compressed data
	compressedData := make([]byte, compressed.Len())
	copy(compressedData, compressed.Bytes())
	
	logger.Trace("Compressed %d bytes to %d bytes (%.1f%% reduction)", 
		n, len(compressedData), float64(len(content)-len(compressedData))/float64(len(content))*100)
	
	return &CompressedContent{
		Type:         CompressionGzip,
		Data:         compressedData,
		OriginalSize: len(content),
	}, nil
}

// DecompressWithPool decompresses gzip data using pooled buffers
func DecompressWithPool(compressedData []byte) ([]byte, error) {
	gr, err := gzip.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return nil, fmt.Errorf("gzip reader creation failed: %w", err)
	}
	defer gr.Close()
	
	// Get buffer from pool
	decompressed := GetSmallBuffer()
	defer PutSmallBuffer(decompressed)
	decompressed.Reset()
	
	_, err = io.Copy(decompressed, gr)
	if err != nil {
		return nil, fmt.Errorf("decompression failed: %w", err)
	}
	
	// Make a copy of the decompressed data
	result := make([]byte, decompressed.Len())
	copy(result, decompressed.Bytes())
	
	return result, nil
}