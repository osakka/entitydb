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
	// SmallBufferPool for small operations (4KB)
	SmallBufferPool = NewSafeBufferPool(4 * 1024)
	
	// MediumBufferPool for medium operations (64KB) 
	MediumBufferPool = NewSafeBufferPool(64 * 1024)
	
	// LargeBufferPool for large operations (1MB)
	LargeBufferPool = NewSafeBufferPool(1024 * 1024)
)