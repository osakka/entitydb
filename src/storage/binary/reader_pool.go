package binary

import (
	"entitydb/logger"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

// ReaderPool manages a pool of file readers to avoid repeated opens/closes
type ReaderPool struct {
	dataFile    string
	maxSize     int
	minSize     int
	
	// Pool of available readers
	available   chan *Reader
	// All readers (for cleanup)
	allReaders  []*Reader
	mu          sync.Mutex
	
	// Metrics
	created     int64
	borrowed    int64
	returned    int64
	closed      chan bool
}

// NewReaderPool creates a new reader pool
func NewReaderPool(dataFile string, minSize, maxSize int) (*ReaderPool, error) {
	if minSize <= 0 {
		minSize = 2
	}
	if maxSize < minSize {
		maxSize = minSize * 2
	}
	
	pool := &ReaderPool{
		dataFile:   dataFile,
		minSize:    minSize,
		maxSize:    maxSize,
		available:  make(chan *Reader, maxSize),
		allReaders: make([]*Reader, 0, maxSize),
		closed:     make(chan bool),
	}
	
	// Pre-create minimum readers
	for i := 0; i < minSize; i++ {
		reader, err := NewReader(dataFile)
		if err != nil {
			// Clean up any created readers
			pool.Close()
			return nil, fmt.Errorf("failed to create reader %d: %v", i, err)
		}
		pool.allReaders = append(pool.allReaders, reader)
		pool.available <- reader
		pool.created++
	}
	
	logger.Info("Created ReaderPool with min=%d, max=%d readers", minSize, maxSize)
	
	// Start metrics reporter
	go pool.reportMetrics()
	
	return pool, nil
}

// Get borrows a reader from the pool
func (p *ReaderPool) Get() (*Reader, error) {
	p.borrowed++
	
	select {
	case reader := <-p.available:
		// Got one from the pool
		return reader, nil
		
	default:
		// Pool is empty, try to create a new one
		p.mu.Lock()
		if len(p.allReaders) < p.maxSize {
			reader, err := NewReader(p.dataFile)
			if err != nil {
				p.mu.Unlock()
				return nil, fmt.Errorf("failed to create new reader: %v", err)
			}
			p.allReaders = append(p.allReaders, reader)
			p.created++
			p.mu.Unlock()
			return reader, nil
		}
		p.mu.Unlock()
		
		// Pool is at max size, wait for one to become available
		select {
		case reader := <-p.available:
			return reader, nil
		case <-time.After(5 * time.Second):
			return nil, fmt.Errorf("reader pool timeout: all %d readers in use", p.maxSize)
		}
	}
}

// Put returns a reader to the pool
func (p *ReaderPool) Put(reader *Reader) {
	if reader == nil {
		return
	}
	
	p.returned++
	
	select {
	case p.available <- reader:
		// Successfully returned to pool
	default:
		// Pool is full (shouldn't happen), close the reader
		reader.Close()
		logger.Warn("ReaderPool full, closing excess reader")
	}
}

// WithReader executes a function with a pooled reader
func (p *ReaderPool) WithReader(fn func(*Reader) error) error {
	reader, err := p.Get()
	if err != nil {
		return err
	}
	defer p.Put(reader)
	
	return fn(reader)
}

// Close closes all readers in the pool
func (p *ReaderPool) Close() error {
	close(p.closed)
	
	// Close all readers
	p.mu.Lock()
	defer p.mu.Unlock()
	
	for _, reader := range p.allReaders {
		if reader != nil {
			reader.Close()
		}
	}
	
	// Drain the available channel
	close(p.available)
	for range p.available {
		// Drain
	}
	
	logger.Info("Closed ReaderPool: created=%d, borrowed=%d, returned=%d", 
		p.created, p.borrowed, p.returned)
	
	return nil
}

// reportMetrics periodically logs pool metrics
func (p *ReaderPool) reportMetrics() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			p.mu.Lock()
			available := len(p.available)
			total := len(p.allReaders)
			p.mu.Unlock()
			
			logger.Debug("ReaderPool stats: %d/%d available, borrowed=%d, returned=%d",
				available, total, p.borrowed, p.returned)
				
		case <-p.closed:
			return
		}
	}
}

// Global reader pool instance
var globalReaderPool *ReaderPool
var poolOnce sync.Once

// GetGlobalReaderPool returns the global reader pool instance
func GetGlobalReaderPool(dataFile string) (*ReaderPool, error) {
	var initErr error
	
	poolOnce.Do(func() {
		// Get pool size from environment
		minSize := 4
		maxSize := 16
		
		if envMin := os.Getenv("ENTITYDB_READER_POOL_MIN"); envMin != "" {
			if n, err := strconv.Atoi(envMin); err == nil && n > 0 {
				minSize = n
			}
		}
		
		if envMax := os.Getenv("ENTITYDB_READER_POOL_MAX"); envMax != "" {
			if n, err := strconv.Atoi(envMax); err == nil && n > 0 {
				maxSize = n
			}
		}
		
		globalReaderPool, initErr = NewReaderPool(dataFile, minSize, maxSize)
	})
	
	if initErr != nil {
		return nil, initErr
	}
	
	return globalReaderPool, nil
}