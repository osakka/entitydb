// Package binary provides single writer queue architecture for corruption prevention
//
// This system ensures that only one write operation occurs at a time, preventing
// concurrent write corruption which can lead to:
// - Torn writes (partial data written)
// - Index inconsistency
// - WAL corruption
// - Race conditions in file operations
package binary

import (
	"context"
	"entitydb/logger"
	"entitydb/models"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// WriteOperation represents a queued write operation
type WriteOperation struct {
	Type      WriteOpType
	Entity    *models.Entity
	EntityID  string
	Tag       string
	Timestamp time.Time
	Done      chan error
	Context   context.Context
}

// WriteOpType defines the type of write operation
type WriteOpType int

const (
	OpCreate WriteOpType = iota
	OpUpdate
	OpDelete
	OpAddTag
	OpRemoveTag
	OpCheckpoint
	OpFlush
)

// String returns the string representation of the operation type
func (wt WriteOpType) String() string {
	switch wt {
	case OpCreate:
		return "CREATE"
	case OpUpdate:
		return "UPDATE"
	case OpDelete:
		return "DELETE"
	case OpAddTag:
		return "ADD_TAG"
	case OpRemoveTag:
		return "REMOVE_TAG"
	case OpCheckpoint:
		return "CHECKPOINT"
	case OpFlush:
		return "FLUSH"
	default:
		return "UNKNOWN"
	}
}

// SingleWriterQueue ensures sequential write operations to prevent corruption
type SingleWriterQueue struct {
	// Core components
	queue       chan *WriteOperation
	stopChan    chan struct{}
	wg          sync.WaitGroup
	
	// Repository reference for actual writes
	repo        *EntityRepository
	
	// Statistics
	queueDepth  int64
	processed   int64
	errors      int64
	
	// Configuration
	maxQueueSize int
	timeout      time.Duration
	
	// State
	running     int32
	mu          sync.RWMutex
}

// NewSingleWriterQueue creates a new single writer queue
func NewSingleWriterQueue(repo *EntityRepository, queueSize int) *SingleWriterQueue {
	if queueSize <= 0 {
		queueSize = 1000 // Default queue size
	}
	
	return &SingleWriterQueue{
		queue:        make(chan *WriteOperation, queueSize),
		stopChan:     make(chan struct{}),
		repo:         repo,
		maxQueueSize: queueSize,
		timeout:      30 * time.Second, // Default operation timeout
	}
}

// Start begins processing write operations
func (swq *SingleWriterQueue) Start() error {
	if !atomic.CompareAndSwapInt32(&swq.running, 0, 1) {
		return fmt.Errorf("single writer queue already running")
	}
	
	swq.wg.Add(1)
	go swq.processQueue()
	
	logger.Info("Single writer queue started with size %d", swq.maxQueueSize)
	return nil
}

// Stop gracefully shuts down the queue
func (swq *SingleWriterQueue) Stop() error {
	if !atomic.CompareAndSwapInt32(&swq.running, 1, 0) {
		return fmt.Errorf("single writer queue not running")
	}
	
	// Signal shutdown
	close(swq.stopChan)
	
	// Wait for queue to drain with timeout
	done := make(chan struct{})
	go func() {
		swq.wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		logger.Info("Single writer queue stopped gracefully")
		return nil
	case <-time.After(swq.timeout):
		logger.Warn("Single writer queue stop timeout - forcing shutdown")
		return fmt.Errorf("shutdown timeout")
	}
}

// processQueue is the main worker loop
func (swq *SingleWriterQueue) processQueue() {
	defer swq.wg.Done()
	
	logger.Debug("Single writer queue worker started")
	
	for {
		select {
		case op := <-swq.queue:
			if op == nil {
				continue
			}
			
			// Update queue depth
			atomic.AddInt64(&swq.queueDepth, -1)
			
			// Process the operation
			err := swq.processOperation(op)
			
			// Send result
			select {
			case op.Done <- err:
			case <-time.After(100 * time.Millisecond):
				logger.Warn("Failed to send operation result - client timeout")
			}
			
			// Update statistics
			atomic.AddInt64(&swq.processed, 1)
			if err != nil {
				atomic.AddInt64(&swq.errors, 1)
			}
			
		case <-swq.stopChan:
			// Drain remaining operations
			remaining := len(swq.queue)
			if remaining > 0 {
				logger.Info("Draining %d remaining operations", remaining)
				for i := 0; i < remaining; i++ {
					op := <-swq.queue
					if op != nil {
						op.Done <- fmt.Errorf("queue shutting down")
					}
				}
			}
			return
		}
	}
}

// processOperation executes a single write operation
func (swq *SingleWriterQueue) processOperation(op *WriteOperation) error {
	startTime := time.Now()
	
	logger.Trace("Processing %s operation for entity %s", op.Type, op.EntityID)
	
	// Set operation timeout
	ctx, cancel := context.WithTimeout(op.Context, swq.timeout)
	defer cancel()
	
	// Check context before processing
	select {
	case <-ctx.Done():
		return fmt.Errorf("operation cancelled: %v", ctx.Err())
	default:
	}
	
	// Execute the operation
	var err error
	switch op.Type {
	case OpCreate:
		err = swq.repo.createInternal(op.Entity)
	case OpUpdate:
		err = swq.repo.updateInternal(op.Entity)
	case OpAddTag:
		err = swq.repo.addTagInternal(op.EntityID, op.Tag)
	case OpDelete:
		// TODO: Implement delete when needed
		err = fmt.Errorf("delete operation not implemented")
	case OpCheckpoint:
		err = swq.repo.writerManager.Checkpoint()
	case OpFlush:
		err = swq.repo.writerManager.Flush()
	default:
		err = fmt.Errorf("unknown operation type: %v", op.Type)
	}
	
	duration := time.Since(startTime)
	if err != nil {
		logger.Error("Operation %s failed for %s: %v (duration: %v)", 
			op.Type, op.EntityID, err, duration)
	} else {
		logger.Trace("Operation %s completed for %s (duration: %v)", 
			op.Type, op.EntityID, duration)
	}
	
	return err
}

// Enqueue adds a write operation to the queue
func (swq *SingleWriterQueue) Enqueue(op *WriteOperation) error {
	if atomic.LoadInt32(&swq.running) == 0 {
		return fmt.Errorf("single writer queue not running")
	}
	
	// Check queue capacity
	currentDepth := atomic.LoadInt64(&swq.queueDepth)
	if currentDepth >= int64(swq.maxQueueSize) {
		return fmt.Errorf("write queue full (%d operations)", currentDepth)
	}
	
	// Set defaults
	if op.Done == nil {
		op.Done = make(chan error, 1)
	}
	if op.Context == nil {
		op.Context = context.Background()
	}
	op.Timestamp = time.Now()
	
	// Try to enqueue with timeout
	select {
	case swq.queue <- op:
		atomic.AddInt64(&swq.queueDepth, 1)
		return nil
	case <-time.After(100 * time.Millisecond):
		return fmt.Errorf("failed to enqueue operation - timeout")
	}
}

// CreateEntity queues an entity creation
func (swq *SingleWriterQueue) CreateEntity(entity *models.Entity) error {
	op := &WriteOperation{
		Type:     OpCreate,
		Entity:   entity,
		EntityID: entity.ID,
		Done:     make(chan error, 1),
		Context:  context.Background(),
	}
	
	if err := swq.Enqueue(op); err != nil {
		return err
	}
	
	// Wait for completion
	select {
	case err := <-op.Done:
		return err
	case <-time.After(swq.timeout):
		return fmt.Errorf("create operation timeout")
	}
}

// UpdateEntity queues an entity update
func (swq *SingleWriterQueue) UpdateEntity(entity *models.Entity) error {
	op := &WriteOperation{
		Type:     OpUpdate,
		Entity:   entity,
		EntityID: entity.ID,
		Done:     make(chan error, 1),
		Context:  context.Background(),
	}
	
	if err := swq.Enqueue(op); err != nil {
		return err
	}
	
	// Wait for completion
	select {
	case err := <-op.Done:
		return err
	case <-time.After(swq.timeout):
		return fmt.Errorf("update operation timeout")
	}
}

// AddTag queues a tag addition
func (swq *SingleWriterQueue) AddTag(entityID, tag string) error {
	op := &WriteOperation{
		Type:     OpAddTag,
		EntityID: entityID,
		Tag:      tag,
		Done:     make(chan error, 1),
		Context:  context.Background(),
	}
	
	if err := swq.Enqueue(op); err != nil {
		return err
	}
	
	// Wait for completion
	select {
	case err := <-op.Done:
		return err
	case <-time.After(swq.timeout):
		return fmt.Errorf("add tag operation timeout")
	}
}

// GetStatistics returns queue statistics
func (swq *SingleWriterQueue) GetStatistics() map[string]int64 {
	return map[string]int64{
		"queue_depth": atomic.LoadInt64(&swq.queueDepth),
		"processed":   atomic.LoadInt64(&swq.processed),
		"errors":      atomic.LoadInt64(&swq.errors),
		"max_size":    int64(swq.maxQueueSize),
		"running":     int64(atomic.LoadInt32(&swq.running)),
	}
}

// IsRunning returns true if the queue is processing operations
func (swq *SingleWriterQueue) IsRunning() bool {
	return atomic.LoadInt32(&swq.running) == 1
}