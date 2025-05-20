package binary

import (
	"entitydb/models"
	"fmt"
	"sync"
	"entitydb/logger"
)

// WriterManager manages a single writer instance for thread-safe access
type WriterManager struct {
	mu       sync.Mutex
	writer   *Writer
	dataFile string
	refCount int
}

// NewWriterManager creates a new writer manager
func NewWriterManager(dataFile string) *WriterManager {
	return &WriterManager{
		dataFile: dataFile,
	}
}

// GetWriter gets the singleton writer instance
func (wm *WriterManager) GetWriter() (*Writer, error) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	
	if wm.writer == nil {
		writer, err := NewWriter(wm.dataFile)
		if err != nil {
			return nil, err
		}
		wm.writer = writer
	}
	
	wm.refCount++
	return wm.writer, nil
}

// ReleaseWriter releases the writer (doesn't close it)
func (wm *WriterManager) ReleaseWriter() {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	
	wm.refCount--
	if wm.refCount < 0 {
		wm.refCount = 0
	}
}

// WriteEntity writes an entity using the managed writer and ensures it's persisted to disk
func (wm *WriterManager) WriteEntity(entity *models.Entity) error {
	logger.Debug("WriterManager.WriteEntity called for entity %s", entity.ID)
	
	writer, err := wm.GetWriter()
	if err != nil {
		logger.Error("Failed to get writer: %v", err)
		return err
	}
	defer wm.ReleaseWriter()
	
	err = writer.WriteEntity(entity)
	if err != nil {
		logger.Error("Failed to write entity %s: %v", entity.ID, err)
		return err
	}
	
	// Always sync file to disk immediately to ensure persistence
	if err := writer.file.Sync(); err != nil {
		logger.Error("Failed to sync entity %s to disk: %v", entity.ID, err)
		return fmt.Errorf("failed to sync entity to disk: %w", err)
	}
	
	// Checkpoint after each write to ensure durability
	logger.Debug("Starting checkpoint after write")
	if err := wm.Checkpoint(); err != nil {
		logger.Error("Failed to checkpoint after write: %v", err)
		// Don't fail the write, just log the error
	} else {
		logger.Debug("Checkpoint completed successfully")
	}
	
	logger.Debug("WriterManager.WriteEntity completed for entity %s", entity.ID)
	return nil
}

// Flush immediately flushes all pending writes to disk
func (wm *WriterManager) Flush() error {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	
	logger.Debug("WriterManager.Flush called")
	
	if wm.writer != nil {
		logger.Debug("Flushing writer to disk")
		if err := wm.writer.file.Sync(); err != nil {
			logger.Error("Failed to sync file: %v", err)
			return err
		}
		logger.Debug("Writer flushed successfully")
		return nil
	}
	
	logger.Debug("No writer exists, nothing to flush")
	return nil
}

// Checkpoint performs a checkpoint operation
func (wm *WriterManager) Checkpoint() error {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	
	logger.Debug("Checkpoint called")
	
	if wm.writer != nil {
		logger.Debug("Writer exists, starting checkpoint process")
		
		// Force a flush
		if err := wm.writer.file.Sync(); err != nil {
			logger.Error("Failed to sync file: %v", err)
			return err
		}
		logger.Debug("File synced")
		
		// Rewrite header and index
		logger.Debug("Calling writer.Close to write index")
		if err := wm.writer.Close(); err != nil {
			logger.Error("Failed to close writer: %v", err)
			return err
		}
		logger.Debug("Writer closed, index should be written")
		
		// Reopen the writer
		logger.Debug("Reopening writer")
		writer, err := NewWriter(wm.dataFile)
		if err != nil {
			logger.Error("Failed to reopen writer: %v", err)
			return err
		}
		wm.writer = writer
		logger.Debug("Writer reopened successfully")
	} else {
		logger.Debug("No writer exists, skipping checkpoint")
	}
	
	return nil
}

// Close closes the writer manager
func (wm *WriterManager) Close() error {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	
	if wm.writer != nil {
		// First close the indexes and flush data
		if err := wm.writer.Close(); err != nil {
			logger.Debug("Error closing writer indexes: %v", err)
		}
		
		// Actually close the file
		if wm.writer.file != nil {
			if err := wm.writer.file.Close(); err != nil {
				logger.Debug("Error closing writer file: %v", err)
				return err
			}
		}
		wm.writer = nil
	}
	
	return nil
}