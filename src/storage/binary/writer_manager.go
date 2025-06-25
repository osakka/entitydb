package binary

import (
	"entitydb/config"
	"entitydb/models"
	"fmt"
	"os"
	"sync"
	"entitydb/logger"
)

// WriterManager manages a single writer instance for thread-safe access
type WriterManager struct {
	mu              sync.Mutex
	writer          *Writer
	dataFile        string
	config          *config.Config
	refCount        int
	atomicFileManager *AtomicFileManager // Atomic file operations for corruption prevention
	useAtomicOps    bool                 // Feature flag for atomic operations
}

// NewWriterManager creates a new writer manager
func NewWriterManager(dataFile string, cfg *config.Config) *WriterManager {
	wm := &WriterManager{
		dataFile:          dataFile,
		config:            cfg,
		atomicFileManager: NewAtomicFileManager(),
		useAtomicOps:      true, // Default to enabled for corruption prevention
	}
	
	// Start atomic file manager
	if err := wm.atomicFileManager.Start(); err != nil {
		logger.Warn("Failed to start atomic file manager: %v", err)
		wm.useAtomicOps = false
	} else {
		logger.Debug("Atomic file manager started for corruption prevention")
	}
	
	return wm
}

// GetWriter gets the singleton writer instance
func (wm *WriterManager) GetWriter() (*Writer, error) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	
	if wm.writer == nil {
		writer, err := NewWriter(wm.dataFile, wm.config)
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

// Checkpoint performs a checkpoint operation with HeaderSync protection
// This implements a three-layer corruption prevention system
func (wm *WriterManager) Checkpoint() error {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	
	logger.Debug("Checkpoint called with HeaderSync protection")
	
	if wm.writer != nil {
		logger.Debug("Writer exists, starting protected checkpoint process")
		
		// LAYER 1: Preserve HeaderSync state before checkpoint
		headerSnapshot := wm.writer.headerSync.CreateSnapshot()
		logger.Debug("HeaderSync snapshot created: WALOffset=%d, EntityCount=%d", 
			headerSnapshot.Header.WALOffset, headerSnapshot.EntityCount)
		
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
		logger.Debug("Reopening writer with validation")
		writer, err := NewWriter(wm.dataFile, wm.config)
		if err != nil {
			logger.Error("Failed to reopen writer: %v", err)
			return err
		}
		
		// LAYER 2: Validate header integrity after reopen
		if !writer.headerSync.ValidateHeader() {
			logger.Warn("HeaderSync validation failed after reopen, corruption detected")
			
			// LAYER 3: Fallback recovery using preserved snapshot
			logger.Info("Attempting HeaderSync recovery from snapshot")
			if err := writer.RestoreHeaderSync(headerSnapshot); err != nil {
				logger.Error("Failed to restore HeaderSync from snapshot: %v", err)
				return fmt.Errorf("checkpoint recovery failed: %w", err)
			}
			logger.Info("HeaderSync successfully recovered from snapshot")
		} else {
			logger.Debug("HeaderSync validation passed after reopen")
		}
		
		wm.writer = writer
		logger.Debug("Writer reopened successfully with HeaderSync protection")
	} else {
		logger.Debug("No writer exists, skipping checkpoint")
	}
	
	return nil
}

// WriteEntityAtomic writes an entity using atomic file operations for corruption prevention
func (wm *WriterManager) WriteEntityAtomic(entity *models.Entity) error {
	if !wm.useAtomicOps || wm.atomicFileManager == nil {
		// Fallback to regular write if atomic operations not available
		return wm.WriteEntity(entity)
	}
	
	logger.Debug("WriteEntityAtomic called for entity %s", entity.ID)
	
	// Get current file data
	currentData, err := os.ReadFile(wm.dataFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read current data: %w", err)
	}
	
	// Create a temporary writer to build the new file content
	tempFile := wm.dataFile + ".atomic.tmp"
	writer, err := NewWriter(tempFile, wm.config)
	if err != nil {
		return fmt.Errorf("failed to create temporary writer: %w", err)
	}
	
	// If we have existing data, we need to read existing entities and add the new one
	if len(currentData) > 0 {
		// Read existing entities
		reader, err := NewReader(wm.dataFile)
		if err != nil {
			writer.Close()
			os.Remove(tempFile)
			return fmt.Errorf("failed to create reader: %w", err)
		}
		
		existingEntities, err := reader.GetAllEntities()
		reader.Close()
		if err != nil {
			writer.Close()
			os.Remove(tempFile)
			return fmt.Errorf("failed to read existing entities: %w", err)
		}
		
		// Write all existing entities
		for _, existingEntity := range existingEntities {
			if err := writer.WriteEntity(existingEntity); err != nil {
				writer.Close()
				os.Remove(tempFile)
				return fmt.Errorf("failed to write existing entity: %w", err)
			}
		}
	}
	
	// Write the new entity
	if err := writer.WriteEntity(entity); err != nil {
		writer.Close()
		os.Remove(tempFile)
		return fmt.Errorf("failed to write new entity: %w", err)
	}
	
	// Close the temporary writer to finalize the file
	if err := writer.Close(); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("failed to close temporary writer: %w", err)
	}
	
	// Read the completed temporary file
	newData, err := os.ReadFile(tempFile)
	if err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("failed to read temporary file: %w", err)
	}
	
	// Atomically update the main file
	if err := wm.atomicFileManager.AtomicUpdate(wm.dataFile, newData); err != nil {
		os.Remove(tempFile)
		return fmt.Errorf("atomic update failed: %w", err)
	}
	
	// Clean up temporary file
	os.Remove(tempFile)
	
	// Invalidate our writer instance so it gets recreated
	wm.mu.Lock()
	if wm.writer != nil {
		wm.writer.Close()
		wm.writer = nil
	}
	wm.mu.Unlock()
	
	logger.Debug("WriteEntityAtomic completed for entity %s", entity.ID)
	return nil
}

// Close closes the writer manager
func (wm *WriterManager) Close() error {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	
	// Stop atomic file manager
	if wm.atomicFileManager != nil {
		if err := wm.atomicFileManager.Stop(); err != nil {
			logger.Warn("Failed to stop atomic file manager: %v", err)
		}
	}
	
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