package api

import (
	"sync"
	"time"
	
	"entitydb/logger"
)

// MetricsWorkerPool manages a pool of workers for processing metrics
type MetricsWorkerPool struct {
	workers   int
	taskQueue chan MetricsTask
	wg        sync.WaitGroup
	shutdown  chan struct{}
}

// MetricsTask represents a metrics storage task
type MetricsTask struct {
	Handler func()
}

// NewMetricsWorkerPool creates a new worker pool
func NewMetricsWorkerPool(workers int, queueSize int) *MetricsWorkerPool {
	pool := &MetricsWorkerPool{
		workers:   workers,
		taskQueue: make(chan MetricsTask, queueSize),
		shutdown:  make(chan struct{}),
	}
	pool.start()
	return pool
}

// start initializes the worker goroutines
func (p *MetricsWorkerPool) start() {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}
	logger.Info("Started metrics worker pool with %d workers", p.workers)
}

// worker processes tasks from the queue
func (p *MetricsWorkerPool) worker(id int) {
	defer p.wg.Done()
	
	for {
		select {
		case task := <-p.taskQueue:
			// Add panic recovery
			func() {
				defer func() {
					if r := recover(); r != nil {
						logger.Error("Worker %d panic: %v", id, r)
					}
				}()
				task.Handler()
			}()
		case <-p.shutdown:
			logger.Debug("Worker %d shutting down", id)
			return
		}
	}
}

// Submit adds a task to the queue
func (p *MetricsWorkerPool) Submit(handler func()) bool {
	select {
	case p.taskQueue <- MetricsTask{Handler: handler}:
		return true
	case <-time.After(10 * time.Millisecond):
		// Queue is full, drop the task
		logger.Warn("Metrics queue full, dropping task")
		return false
	}
}

// Shutdown gracefully stops the worker pool
func (p *MetricsWorkerPool) Shutdown() {
	close(p.shutdown)
	p.wg.Wait()
	close(p.taskQueue)
	logger.Info("Metrics worker pool shut down")
}

// QueueSize returns the current queue size
func (p *MetricsWorkerPool) QueueSize() int {
	return len(p.taskQueue)
}