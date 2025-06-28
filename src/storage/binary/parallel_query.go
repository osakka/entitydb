package binary

import (
	"entitydb/models"
	"runtime"
	"sync"
)

// ParallelQueryProcessor handles concurrent query execution
type ParallelQueryProcessor struct {
	repo       *EntityRepository
	workerPool *WorkerPool
}

// WorkerPool manages a pool of goroutines for parallel processing
type WorkerPool struct {
	workers   int
	taskQueue chan Task
	wg        sync.WaitGroup
}

// Task represents a unit of work
type Task struct {
	EntityIDs []string
	Filter    func(*models.Entity) bool
	Result    chan<- *models.Entity
}

// NewParallelQueryProcessor creates a new parallel query processor
func NewParallelQueryProcessor(repo *EntityRepository) *ParallelQueryProcessor {
	numWorkers := runtime.NumCPU() * 2
	
	pool := &WorkerPool{
		workers:   numWorkers,
		taskQueue: make(chan Task, 1000),
	}
	
	// Start worker goroutines
	for i := 0; i < numWorkers; i++ {
		pool.wg.Add(1)
		go pool.worker(repo)
	}
	
	return &ParallelQueryProcessor{
		repo:       repo,
		workerPool: pool,
	}
}

// worker processes tasks from the queue
func (wp *WorkerPool) worker(repo *EntityRepository) {
	defer wp.wg.Done()
	
	// Use bounded reader pool to prevent file descriptor exhaustion
	// This is critical for preventing OS-level Seek() race conditions
	
	for task := range wp.taskQueue {
		// Get reader from bounded pool for each task to prevent FD exhaustion
		reader, err := repo.readerPool.Get()
		if err != nil {
			continue // Skip this task if can't get reader
		}
		
		for _, entityID := range task.EntityIDs {
			entity, err := reader.GetEntity(entityID)
			if err != nil {
				continue
			}
			
			if task.Filter == nil || task.Filter(entity) {
				task.Result <- entity
			}
		}
		
		// Return reader to pool after processing this task
		repo.readerPool.Put(reader)
	}
}

// QueryParallel executes a query in parallel
func (p *ParallelQueryProcessor) QueryParallel(entityIDs []string, filter func(*models.Entity) bool) ([]*models.Entity, error) {
	resultChan := make(chan *models.Entity, len(entityIDs))
	
	// Split work into chunks
	chunkSize := len(entityIDs) / p.workerPool.workers
	if chunkSize < 10 {
		chunkSize = 10
	}
	
	tasksSubmitted := 0
	for i := 0; i < len(entityIDs); i += chunkSize {
		end := i + chunkSize
		if end > len(entityIDs) {
			end = len(entityIDs)
		}
		
		task := Task{
			EntityIDs: entityIDs[i:end],
			Filter:    filter,
			Result:    resultChan,
		}
		
		p.workerPool.taskQueue <- task
		tasksSubmitted++
	}
	
	// Collect results
	results := make([]*models.Entity, 0, len(entityIDs))
	expectedResults := len(entityIDs)
	
	for i := 0; i < expectedResults; i++ {
		select {
		case entity := <-resultChan:
			if entity != nil {
				results = append(results, entity)
			}
		default:
			// Continue if no result available
		}
	}
	
	return results, nil
}

// Close shuts down the worker pool
func (p *ParallelQueryProcessor) Close() {
	close(p.workerPool.taskQueue)
	p.workerPool.wg.Wait()
}