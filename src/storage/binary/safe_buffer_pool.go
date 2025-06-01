package binary

import (
	"bytes"
	"sync"
)

// SafeBufferPool provides thread-safe buffer pooling with proper reset
type SafeBufferPool struct {
	pool sync.Pool
}

// NewSafeBufferPool creates a new safe buffer pool
func NewSafeBufferPool(size int) *SafeBufferPool {
	return &SafeBufferPool{
		pool: sync.Pool{
			New: func() interface{} {
				return bytes.NewBuffer(make([]byte, 0, size))
			},
		},
	}
}

// Get retrieves a buffer from the pool, properly reset
func (p *SafeBufferPool) Get() *bytes.Buffer {
	buf := p.pool.Get().(*bytes.Buffer)
	buf.Reset() // Always reset before use
	return buf
}

// Put returns a buffer to the pool
func (p *SafeBufferPool) Put(buf *bytes.Buffer) {
	// Only pool reasonable sized buffers
	if buf.Cap() > 10*1024*1024 { // 10MB limit
		return
	}
	// Reset before putting back
	buf.Reset()
	p.pool.Put(buf)
}

// Global pools for different use cases
var (
	// smallSafeBufferPool for small operations (4KB)
	smallSafeBufferPool = NewSafeBufferPool(4 * 1024)
	
	// mediumSafeBufferPool for medium operations (64KB) 
	mediumSafeBufferPool = NewSafeBufferPool(64 * 1024)
	
	// largeSafeBufferPool for large operations (1MB)
	largeSafeBufferPool = NewSafeBufferPool(1024 * 1024)
)

// GetSmallBuffer gets a small buffer from the safe pool
func GetSmallBuffer() *bytes.Buffer {
	return smallSafeBufferPool.Get()
}

// PutSmallBuffer returns a small buffer to the safe pool
func PutSmallBuffer(buf *bytes.Buffer) {
	smallSafeBufferPool.Put(buf)
}

// GetMediumBuffer gets a medium buffer from the safe pool
func GetMediumBuffer() *bytes.Buffer {
	return mediumSafeBufferPool.Get()
}

// PutMediumBuffer returns a medium buffer to the safe pool
func PutMediumBuffer(buf *bytes.Buffer) {
	mediumSafeBufferPool.Put(buf)
}

// GetLargeSafeBuffer gets a large buffer from the safe pool
func GetLargeSafeBuffer() *bytes.Buffer {
	return largeSafeBufferPool.Get()
}

// PutLargeSafeBuffer returns a large buffer to the safe pool
func PutLargeSafeBuffer(buf *bytes.Buffer) {
	largeSafeBufferPool.Put(buf)
}