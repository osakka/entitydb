// Package binary provides a high-performance binary storage implementation for EntityDB.
//
// This package implements a custom binary format (EBF - Entity Binary Format) with:
//   - Write-Ahead Logging (WAL) for durability
//   - Memory-mapped file access for performance
//   - B-tree temporal indexes for time-based queries
//   - Bloom filters for efficient tag lookups
//   - Automatic checkpointing and recovery
//
// The binary format is designed for optimal performance with temporal data,
// supporting nanosecond-precision timestamps and efficient range queries.
package binary

import (
	"entitydb/models"
	"entitydb/cache"
	"entitydb/logger"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// EntityRepository implements models.EntityRepository using a custom binary format.
// It provides high-performance storage with temporal capabilities, supporting
// both standard file-based and memory-mapped access patterns.
type EntityRepository struct {
	dataPath string
	mu       sync.RWMutex  // Still keep this for backward compatibility
	
	// In-memory indexes for queries
	tagIndex     map[string][]string  // tag -> entity IDs (legacy)
	contentIndex map[string][]string  // content -> entity IDs
	
	// Sharded tag index for improved concurrency
	shardedTagIndex *ShardedTagIndex
	useShardedIndex bool // Feature flag for gradual rollout
	
	// In-memory entity storage
	entities     map[string]*models.Entity // id -> entity
	
	// Locking and transaction support
	lockManager *LockManager
	wal         *WAL
	
	// File handle management
	readerPool    sync.Pool      // Pool of readers for concurrent access
	writerManager *WriterManager // Manages single writer instance
	currentFile   *os.File       // Current file handle
	
	// Query cache for performance
	cache *cache.QueryCache
	
	// Temporal index for efficient temporal queries
	temporalIndex *TemporalIndex
	
	// Namespace index for efficient namespace queries
	namespaceIndex *NamespaceIndex
	
	// Tag index persistence
	tagIndexDirty         bool        // Whether tag index needs to be saved
	lastIndexSave         time.Time   // Last time index was saved
	
	// WAL checkpoint management
	walOperationCount     int64       // Count of operations since last checkpoint
	lastCheckpoint        time.Time   // Time of last checkpoint
	checkpointMu          sync.Mutex  // Protect checkpoint operations
	persistentIndexLoaded bool        // Whether persistent index was loaded successfully
	
	// Recovery manager
	recovery *RecoveryManager
}

// NewEntityRepository creates a new binary entity repository
func NewEntityRepository(dataPath string) (*EntityRepository, error) {
	// Ensure data directory exists
	if err := os.MkdirAll(dataPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}
	
	// Check environment variable for sharded index feature flag
	// Default to true for better performance, can be disabled by setting to "false"
	useSharded := os.Getenv("ENTITYDB_USE_SHARDED_INDEX") != "false"
	
	repo := &EntityRepository{
		dataPath:        dataPath,
		tagIndex:        make(map[string][]string),
		contentIndex:    make(map[string][]string),
		entities:        make(map[string]*models.Entity),
		lockManager:     NewLockManager(),
		writerManager:   NewWriterManager(filepath.Join(dataPath, "entities.ebf")),
		cache:           cache.NewQueryCache(1000, 5*time.Minute), // Cache up to 1000 queries for 5 minutes
		temporalIndex:   NewTemporalIndex(),
		namespaceIndex:  NewNamespaceIndex(),
		lastCheckpoint:  time.Now(),  // Initialize checkpoint time
		shardedTagIndex: NewShardedTagIndex(),
		useShardedIndex: useSharded,
	}
	
	if useSharded {
		logger.Info("Using sharded tag index for improved concurrency")
	}
	
	// Ensure the data file exists with a proper header before trying to read it
	dataFile := repo.getDataFile()
	if _, err := os.Stat(dataFile); os.IsNotExist(err) {
		// Use the writerManager to create the initial file with header
		_, err := repo.writerManager.GetWriter()
		if err != nil {
			return nil, fmt.Errorf("error creating initial data file: %w", err)
		}
		// The writer creates the file with a header when it's first created
		repo.writerManager.ReleaseWriter()
	}
	
	// Initialize reader pool with binary format readers
	repo.readerPool = sync.Pool{
		New: func() interface{} {
			reader, err := NewReader(repo.getDataFile())
			if err != nil {
				logger.Error("Failed to create reader: %v", err)
				return nil
			}
			return reader
		},
	}
	
	// Initialize WAL
	wal, err := NewWAL(dataPath)
	if err != nil {
		return nil, fmt.Errorf("error creating WAL: %w", err)
	}
	repo.wal = wal
	
	// Initialize recovery manager
	repo.recovery = NewRecoveryManager(dataPath)
	
	// Ensure data file exists before building indexes
	if _, err := os.Stat(repo.getDataFile()); os.IsNotExist(err) {
		_, err := repo.writerManager.GetWriter()
		if err != nil {
			return nil, fmt.Errorf("error creating initial data file: %w", err)
		}
		repo.writerManager.ReleaseWriter()
	}
	
	// Build initial indexes
	if err := repo.buildIndexes(); err != nil {
		logger.Warn("Failed to build initial indexes: %v", err)
		// Don't fail initialization - we can still write entities
	}
	
	// Log entity count after building indexes
	logger.Info("Initialized: %d entities, %d tag index entries", 
		len(repo.entities), len(repo.tagIndex))
	
	// WAL replay is already handled during WAL initialization
	// No need for additional replay here as entities already include WAL data
	
	// Log entity count after WAL replay
	logger.Info("WAL replay complete: %d entities, %d tag entries", 
		len(repo.entities), len(repo.tagIndex))
	
	// Verify index health
	if err := repo.VerifyIndexHealth(); err != nil {
		logger.Warn("Index health check failed: %v", err)
		// Attempt to repair the index
		logger.Info("Attempting automatic index repair")
		if repairErr := repo.RepairIndexes(); repairErr != nil {
			logger.Error("Index repair failed: %v", repairErr)
		} else {
			logger.Info("Index repair completed successfully")
			// Verify again after repair
			if verifyErr := repo.VerifyIndexHealth(); verifyErr != nil {
				logger.Error("Index still unhealthy after repair: %v", verifyErr)
			}
		}
	}
	
	// Open the current file for use with readers
	repo.currentFile, err = os.OpenFile(repo.getDataFile(), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("error opening data file: %w", err)
	}
	
	return repo, nil
}

// getDataFile returns the path to the current data file
func (r *EntityRepository) getDataFile() string {
	return filepath.Join(r.dataPath, "entities.ebf")
}

// buildIndexes reads the entire file and builds in-memory indexes
func (r *EntityRepository) buildIndexes() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Clear existing indexes
	r.tagIndex = make(map[string][]string)
	r.contentIndex = make(map[string][]string)
	r.temporalIndex = NewTemporalIndex()
	r.namespaceIndex = NewNamespaceIndex()
	
	// Try to load persisted index first
	indexFile := r.getDataFile() + ".idx"
	if _, err := os.Stat(indexFile); err == nil {
		logger.Debug("Loading persisted tag index from %s", indexFile)
		if loadedIndex, err := LoadTagIndexV2(r.getDataFile()); err == nil {
			r.tagIndex = loadedIndex
			r.persistentIndexLoaded = true
			
			// If using sharded index, populate it from loaded index
			if r.useShardedIndex {
				logger.Debug("Populating sharded index from loaded data")
				for tag, entities := range loadedIndex {
					for _, entityID := range entities {
						r.shardedTagIndex.AddTag(tag, entityID)
					}
				}
			}
			
			logger.Info("Loaded persisted tag index with %d tags", len(loadedIndex))
			// Still need to load entities into memory
			// Continue to load entities below
		} else {
			logger.Warn("Failed to load persisted index: %v, will rebuild", err)
		}
	}
	
	// Always need to read entities from disk
	logger.Debug("Building indexes from entities")
	
	reader, err := NewReader(r.getDataFile())
	if err != nil {
		return err
	}
	defer reader.Close()
	
	// Read all entities
	entities, err := reader.GetAllEntities()
	if err != nil {
		return err
	}
	
	// Load entities and optionally build indexes
	logger.Debug("Loading entities: %d found", len(entities))
	for _, entity := range entities {
		// Add to entity cache
		r.entities[entity.ID] = entity
		
		// Log entity details for debugging
		if strings.HasPrefix(entity.ID, "rel_") {
			logger.Trace("Indexing relationship: %s", entity.ID)
		}
		
		// Only update indexes if we didn't load from persistent index
		if !r.persistentIndexLoaded {
			// Update tag index
			for _, tag := range entity.Tags {
				if r.useShardedIndex {
					// Add to sharded index
					r.shardedTagIndex.AddTag(tag, entity.ID)
					
					// Also index the non-timestamped version for temporal tags
					if strings.Contains(tag, "|") {
						parts := strings.SplitN(tag, "|", 2)
						if len(parts) == 2 {
							r.shardedTagIndex.AddTag(parts[1], entity.ID)
						}
					}
				} else {
					// Legacy indexing
					r.tagIndex[tag] = append(r.tagIndex[tag], entity.ID)
					
					// Add to temporal index if it's a temporal tag
					if strings.Contains(tag, "|") {
						parts := strings.SplitN(tag, "|", 2)
						if len(parts) == 2 {
							// ALSO index the non-timestamped version for easier searching
							// This is critical for authentication and tag lookups
							actualTag := parts[1]
							r.tagIndex[actualTag] = append(r.tagIndex[actualTag], entity.ID)
							
							// Log relationship tag indexing
							if strings.HasPrefix(actualTag, "_source:") || strings.HasPrefix(actualTag, "_target:") || strings.HasPrefix(actualTag, "_relationship:") {
								logger.Trace("Indexed relationship tag: %s for %s", actualTag, entity.ID)
							}
						}
					}
				}
				
				// Add to temporal index if it's a temporal tag (regardless of sharding)
				if strings.Contains(tag, "|") {
					parts := strings.SplitN(tag, "|", 2)
					if len(parts) == 2 {
						// Try to parse timestamp - it's stored as Unix nanoseconds
						if timestampNanos, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
							timestamp := time.Unix(0, timestampNanos)
							r.temporalIndex.AddEntry(entity.ID, tag, timestamp)
						}
					}
				}
				
				// Add to namespace index
				r.namespaceIndex.AddTag(entity.ID, tag)
			}
		}
		
		// Update content index - store content as string for searching
		if len(entity.Content) > 0 {
			contentStr := string(entity.Content)
			r.contentIndex[contentStr] = append(r.contentIndex[contentStr], entity.ID)
		}
	}
	
	return nil
}

// updateIndexes updates in-memory indexes for a new or updated entity
// This function MUST be called with the mutex already locked
func (r *EntityRepository) updateIndexes(entity *models.Entity) {
	logger.Trace("Updating indexes for entity %s (%d tags)", entity.ID, len(entity.Tags))
	
	// CRITICAL: First remove all existing index entries for this entity
	// This prevents duplicate entries when updating
	if existingEntity, exists := r.entities[entity.ID]; exists {
		for _, tag := range existingEntity.Tags {
			// Remove from tag index
			if ids, ok := r.tagIndex[tag]; ok {
				newIDs := make([]string, 0, len(ids))
				for _, id := range ids {
					if id != entity.ID {
						newIDs = append(newIDs, id)
					}
				}
				if len(newIDs) > 0 {
					r.tagIndex[tag] = newIDs
				} else {
					delete(r.tagIndex, tag)
				}
			}
			
			// Also remove non-timestamped version
			if strings.Contains(tag, "|") {
				parts := strings.SplitN(tag, "|", 2)
				if len(parts) == 2 {
					actualTag := parts[1]
					if ids, ok := r.tagIndex[actualTag]; ok {
						newIDs := make([]string, 0, len(ids))
						for _, id := range ids {
							if id != entity.ID {
								newIDs = append(newIDs, id)
							}
						}
						if len(newIDs) > 0 {
							r.tagIndex[actualTag] = newIDs
						} else {
							delete(r.tagIndex, actualTag)
						}
					}
				}
			}
		}
	}
	
	// Helper function to add entity ID to tag if not already present
	addEntityToTag := func(tag, entityID string) {
		// Check if entity is already indexed for this tag
		existing := r.tagIndex[tag]
		for _, id := range existing {
			if id == entityID {
				return // Already indexed
			}
		}
		r.tagIndex[tag] = append(r.tagIndex[tag], entityID)
	}
	
	// Update tag index
	for _, tag := range entity.Tags {
		// Always index the full tag (with timestamp)
		logger.Trace("Indexing tag: %s for entity %s", tag, entity.ID)
		addEntityToTag(tag, entity.ID)
		
		// Also index the non-timestamped version for easier searching
		if strings.Contains(tag, "|") {
			parts := strings.SplitN(tag, "|", 2)
			if len(parts) == 2 {
				// Try to parse timestamp as nanosecond epoch
				if timestampNanos, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
					timestamp := time.Unix(0, timestampNanos)
					r.temporalIndex.AddEntry(entity.ID, tag, timestamp)
					logger.Trace("Added to temporal index: %s", entity.ID)
				} else {
					logger.Trace("Failed to parse timestamp in tag: %s", tag)
				}
				
				// Index the actual tag part too
				actualTag := parts[1]
				logger.Trace("Indexing non-timestamped: %s for %s", actualTag, entity.ID)
				addEntityToTag(actualTag, entity.ID)
			}
		}
		
		// Add to namespace index
		r.namespaceIndex.AddTag(entity.ID, tag)
	}
	
	// Mark tag index as dirty
	r.tagIndexDirty = true
	
	// Update content index - store content as string for searching
	if len(entity.Content) > 0 {
		contentStr := string(entity.Content)
		r.contentIndex[contentStr] = append(r.contentIndex[contentStr], entity.ID)
		logger.Trace("Indexed %d bytes of content for %s", len(contentStr), entity.ID)
	}
	
	// Dump tag index for debugging - removed as too verbose
}

// Create creates a new entity with strong durability guarantees
func (r *EntityRepository) Create(entity *models.Entity) error {
	startTime := time.Now()
	
	// Generate UUID only if no ID is provided
	if entity.ID == "" {
		entity.ID = models.GenerateUUID()
	}
	
	entity.CreatedAt = models.Now()
	entity.UpdatedAt = entity.CreatedAt
	
	// Ensure all tags have timestamps (temporal-only system)
	timestampedTags := []string{}
	for _, tag := range entity.Tags {
		if !strings.Contains(tag, "|") {
			// Add timestamp if not present (temporal-only system requires all tags to have timestamps)
			timestampedTags = append(timestampedTags, fmt.Sprintf("%s|%s", models.NowString(), tag))
		} else {
			// Keep existing timestamped tags
			timestampedTags = append(timestampedTags, tag)
		}
	}
	entity.Tags = timestampedTags
	
	// Note: Checksum generation disabled - was causing systematic validation failures
	// without providing real security value. Can be re-implemented properly if needed.
	logger.Trace("Entity prepared for storage: %s (%d bytes content)", entity.ID, len(entity.Content))
	
	// Log to WAL first
	if err := r.wal.LogCreate(entity); err != nil {
		return fmt.Errorf("error logging to WAL: %w", err)
	}
	
	// Write entity with locking
	r.lockManager.AcquireEntityLock(entity.ID, WriteLock)
	defer r.lockManager.ReleaseEntityLock(entity.ID, WriteLock)
	
	// CRITICAL: Update indexes BEFORE writing to ensure atomicity
	// This prevents the entity from being written without indexes
	r.mu.Lock()
	r.updateIndexes(entity)
	// Store entity in-memory as well
	r.entities[entity.ID] = entity
	r.mu.Unlock()
	
	// Write entity using WriterManager (which handles checkpoints)
	if err := r.writerManager.WriteEntity(entity); err != nil {
		// Rollback index changes on write failure
		r.mu.Lock()
		delete(r.entities, entity.ID)
		// Remove from all indexes
		for _, tag := range entity.Tags {
			r.removeFromTagIndex(tag, entity.ID)
			// Also remove non-timestamped version
			if strings.Contains(tag, "|") {
				parts := strings.SplitN(tag, "|", 2)
				if len(parts) == 2 {
					r.removeFromTagIndex(parts[1], entity.ID)
				}
			}
		}
		r.mu.Unlock()
		return err
	}
	
	// Invalidate cache
	r.cache.Clear()
	
	// Explicitly sync to disk to ensure persistence
	if err := r.writerManager.Flush(); err != nil {
		logger.Error("Failed to flush writes to disk: %v", err)
		return fmt.Errorf("failed to flush entity to disk: %w", err)
	}
	
	// Force a checkpoint to ensure data is fully persisted
	if err := r.writerManager.Checkpoint(); err != nil {
		logger.Error("Failed to checkpoint after create: %v", err)
		// Don't fail the write, just log the error
	}
	
	logger.Debug("Created entity: %s", entity.ID)
	
	// Save tag index periodically
	if err := r.SaveTagIndexIfNeeded(); err != nil {
		logger.Warn("Failed to save tag index: %v", err)
	}
	
	// Check if we need to perform checkpoint
	r.checkAndPerformCheckpoint()
	
	// Track write metrics (skip metric entities to avoid recursion)
	if storageMetrics != nil && !strings.HasPrefix(entity.ID, "metric_") {
		duration := time.Since(startTime)
		size := int64(len(entity.Content))
		storageMetrics.TrackWrite("create_entity", size, duration, nil)
	}
	
	return nil
}

// GetByID gets an entity by ID with improved reliability from in-memory cache
func (r *EntityRepository) GetByID(id string) (*models.Entity, error) {
	startTime := time.Now()
	logger.Trace("GetByID: %s", id)
	
	// First check in-memory cache for the entity
	r.mu.RLock()
	entity, exists := r.entities[id]
	r.mu.RUnlock()
	
	if exists {
		logger.Trace("Found in memory cache: %s", id)
		// Skip metrics for metric entities to avoid recursion
		if storageMetrics != nil && !strings.HasPrefix(id, "metric_") {
			storageMetrics.TrackCacheOperation("entity", true)
		}
		return entity, nil
	}
	
	// Cache miss - skip metrics for metric entities to avoid recursion
	if storageMetrics != nil && !strings.HasPrefix(id, "metric_") {
		storageMetrics.TrackCacheOperation("entity", false)
	}
	
	// First check if entity exists in indexes
	r.mu.RLock()
	found := false
	for _, ids := range r.tagIndex {
		for _, entityID := range ids {
			if entityID == id {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	r.mu.RUnlock()
	
	if !found {
		logger.Trace("Not found in index, checking disk: %s", id)
		// Attempt to detect and fix index corruption
		if err := r.detectAndFixIndexCorruption(id); err != nil {
			logger.Warn("Failed to fix potential index corruption for %s: %v", id, err)
		}
	}
	
	// Skip flush and checkpoint for metric entities to avoid infinite recursion
	// For non-metric entities, force a flush and checkpoint
	if !strings.HasPrefix(id, "metric_") {
		// Force a flush and checkpoint of any pending writes before attempting to read
		// This ensures that we can read immediately after writing
		if err := r.writerManager.Flush(); err != nil {
			logger.Error("Failed to flush writes: %v", err)
			// Continue anyway as we might still find the entity
		}
		
		// Also force a checkpoint to ensure index is updated
		if err := r.writerManager.Checkpoint(); err != nil {
			logger.Error("Failed to checkpoint: %v", err)
			// Continue anyway as we might still find the entity
		}
	}
	
	// Acquire read lock for the entity
	r.lockManager.AcquireEntityLock(id, ReadLock)
	defer r.lockManager.ReleaseEntityLock(id, ReadLock)
	
	// Get a reader from the pool
	readerInterface := r.readerPool.Get()
	if readerInterface == nil {
		logger.Trace("Creating new reader for %s", id)
		reader, err := NewReader(r.getDataFile())
		if err != nil {
			logger.Error("Failed to create reader: %v", err)
			return nil, err
		}
		defer reader.Close()
		
		readStart := time.Now()
		entity, err := reader.GetEntity(id)
		readDuration := time.Since(readStart)
		
		// Track read metrics (skip metric entities to avoid recursion)
		if storageMetrics != nil && !strings.HasPrefix(id, "metric_") {
			size := int64(0)
			if entity != nil {
				size = int64(len(entity.Content))
			}
			storageMetrics.TrackRead("get_entity", size, readDuration, err)
		}
		
		if err != nil {
			logger.Error("Failed to get entity %s: %v", id, err)
			
			// Try recovery if read failed
			logger.Info("Attempting recovery for entity %s", id)
			if recoveredEntity, recErr := r.recovery.RecoverCorruptedEntity(r, id); recErr == nil {
				logger.Info("Successfully recovered entity %s", id)
				// Store recovered entity
				r.mu.Lock()
				r.entities[id] = recoveredEntity
				r.mu.Unlock()
				
				// Track overall operation time including recovery
				if storageMetrics != nil {
					totalDuration := time.Since(startTime)
					storageMetrics.TrackRead("get_entity_with_recovery", int64(len(recoveredEntity.Content)), totalDuration, nil)
				}
				
				return recoveredEntity, nil
			} else {
				logger.Error("Recovery failed for entity %s: %v", id, recErr)
			}
			
			return nil, err
		}
		
		if entity != nil {
			logger.Trace("Found entity %s: %d bytes, %d tags", 
				id, len(entity.Content), len(entity.Tags))
			
			// Store in memory for future fast access
			r.mu.Lock()
			r.entities[id] = entity
			r.mu.Unlock()
		}
		return entity, nil
	}
	
	logger.Trace("Using pooled reader for %s", id)
	reader := readerInterface.(*Reader)
	defer r.readerPool.Put(reader)
	
	entity, err := reader.GetEntity(id)
	if err != nil {
		logger.Error("Failed to get entity %s: %v", id, err)
		
		// Try recovery if read failed
		logger.Info("Attempting recovery for entity %s", id)
		if recoveredEntity, recErr := r.recovery.RecoverCorruptedEntity(r, id); recErr == nil {
			logger.Info("Successfully recovered entity %s", id)
			// Store recovered entity
			r.mu.Lock()
			r.entities[id] = recoveredEntity
			r.mu.Unlock()
			return recoveredEntity, nil
		} else {
			logger.Error("Recovery failed for entity %s: %v", id, recErr)
		}
		
		return nil, err
	}
	
	if entity != nil {
		logger.Trace("Found entity %s: %d bytes, %d tags",
			id, len(entity.Content), len(entity.Tags))
		
		// Store in memory for future fast access
		r.mu.Lock()
		r.entities[id] = entity
		r.mu.Unlock()
	} else {
		logger.Trace("Entity %s not found", id)
	}
	
	return entity, nil
}

// Update updates an existing entity
func (r *EntityRepository) Update(entity *models.Entity) error {
	startTime := time.Now()
	
	if entity.ID == "" {
		return fmt.Errorf("entity ID is required for update")
	}
	
	// Verify the entity exists (prevents ID manipulation)
	existingEntity, err := r.GetByID(entity.ID)
	if err != nil {
		return fmt.Errorf("entity not found: %w", err)
	}
	
	// Preserve the original ID (make it immutable)
	entity.ID = existingEntity.ID
	entity.CreatedAt = existingEntity.CreatedAt // Also preserve creation time
	
	entity.UpdatedAt = models.Now()
	
	// Ensure all tags have timestamps (temporal-only system)
	timestampedTags := []string{}
	for _, tag := range entity.Tags {
		if !strings.Contains(tag, "|") {
			// Add timestamp if not present (temporal-only system requires all tags to have timestamps)
			timestampedTags = append(timestampedTags, fmt.Sprintf("%s|%s", models.NowString(), tag))
		} else {
			// Keep existing timestamped tags
			timestampedTags = append(timestampedTags, tag)
		}
	}
	entity.Tags = timestampedTags
	
	// Content in the new model is just binary data - no timestamps needed
	
	// Log to WAL first
	if err := r.wal.LogUpdate(entity); err != nil {
		return fmt.Errorf("error logging to WAL: %w", err)
	}
	
	// Acquire write lock
	r.lockManager.AcquireEntityLock(entity.ID, WriteLock)
	defer r.lockManager.ReleaseEntityLock(entity.ID, WriteLock)
	
	// Create temporary file for writing
	tempPath := r.getDataFile() + ".tmp"
	writer, err := NewWriter(tempPath)
	if err != nil {
		return err
	}
	defer writer.Close()
	
	// Read all entities and update the target
	reader, err := NewReader(r.getDataFile())
	if err != nil {
		return err
	}
	defer reader.Close()
	
	entities, err := reader.GetAllEntities()
	if err != nil {
		return err
	}
	
	// Write updated entities
	updated := false
	for _, e := range entities {
		if e.ID == entity.ID {
			if err := writer.WriteEntity(entity); err != nil {
				return err
			}
			updated = true
		} else {
			if err := writer.WriteEntity(e); err != nil {
				return err
			}
		}
	}
	
	if !updated {
		return fmt.Errorf("entity not found: %s", entity.ID)
	}
	
	writer.Close()
	
	// Replace the original file with the temporary file
	if err := os.Rename(tempPath, r.getDataFile()); err != nil {
		return err
	}
	
	// Rebuild indexes
	r.buildIndexes()
	
	// Invalidate cache
	r.cache.Clear()
	
	// Save tag index periodically
	if err := r.SaveTagIndexIfNeeded(); err != nil {
		logger.Warn("Failed to save tag index: %v", err)
	}
	
	// Check if we need to perform checkpoint
	r.checkAndPerformCheckpoint()
	
	// Track write metrics (skip metric entities to avoid recursion)
	if storageMetrics != nil && !strings.HasPrefix(entity.ID, "metric_") {
		duration := time.Since(startTime)
		size := int64(len(entity.Content))
		storageMetrics.TrackWrite("update_entity", size, duration, nil)
	}
	
	return nil
}

// VerifyIndexIntegrity checks for index consistency issues
func (r *EntityRepository) VerifyIndexIntegrity() []error {
	var errors []error
	
	reader, err := NewReader(r.getDataFile())
	if err != nil {
		errors = append(errors, fmt.Errorf("failed to open reader: %v", err))
		return errors
	}
	defer reader.Close()
	
	// Check 1: Every index entry points to valid data
	for id, entry := range reader.index {
		// Try to read the entity
		entity, err := reader.GetEntity(id)
		if err != nil {
			errors = append(errors, fmt.Errorf("index entry %s points to unreadable data: %v", id, err))
			continue
		}
		
		// Verify ID matches
		if entity.ID != id {
			errors = append(errors, fmt.Errorf("index entry %s points to entity with ID %s", id, entity.ID))
		}
		
		// Verify offset and size are reasonable
		if entry.Offset == 0 || entry.Size == 0 {
			errors = append(errors, fmt.Errorf("index entry %s has invalid offset/size: %d/%d", id, entry.Offset, entry.Size))
		}
	}
	
	// Check 2: Header count matches index entries
	actualCount := len(reader.index)
	if uint64(actualCount) != reader.header.EntityCount {
		errors = append(errors, fmt.Errorf("header claims %d entities but index has %d", reader.header.EntityCount, actualCount))
	}
	
	return errors
}

// FindOrphanedEntries finds entries in index that don't exist in data
func (r *EntityRepository) FindOrphanedEntries() []string {
	var orphaned []string
	
	reader, err := NewReader(r.getDataFile())
	if err != nil {
		return orphaned
	}
	defer reader.Close()
	
	// Check each index entry
	for id := range reader.index {
		_, err := reader.GetEntity(id)
		if err != nil {
			orphaned = append(orphaned, id)
		}
	}
	
	return orphaned
}

// RebuildIndex rebuilds the index from scratch
func (r *EntityRepository) RebuildIndex() error {
	logger.Info("Rebuilding index from data file")
	
	// Create a new temporary file
	tempPath := r.getDataFile() + ".rebuild"
	newWriter, err := NewWriter(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create temp writer: %v", err)
	}
	
	// Read all valid entities
	reader, err := NewReader(r.getDataFile())
	if err != nil {
		newWriter.Close()
		os.Remove(tempPath)
		return fmt.Errorf("failed to create reader: %v", err)
	}
	
	validCount := 0
	allEntities, _ := reader.GetAllEntities()
	for _, entity := range allEntities {
		// Try to read each entity
		if entity != nil && entity.ID != "" {
			if err := newWriter.WriteEntity(entity); err != nil {
				logger.Warn("Failed to write entity %s during rebuild: %v", entity.ID, err)
			} else {
				validCount++
			}
		}
	}
	
	reader.Close()
	newWriter.Close()
	
	// Backup old file
	backupPath := r.getDataFile() + ".backup." + time.Now().Format("20060102150405")
	if err := os.Rename(r.getDataFile(), backupPath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to backup old file: %v", err)
	}
	
	// Move new file into place
	if err := os.Rename(tempPath, r.getDataFile()); err != nil {
		// Try to restore backup
		os.Rename(backupPath, r.getDataFile())
		return fmt.Errorf("failed to move rebuilt file: %v", err)
	}
	
	logger.Info("Index rebuilt: %d valid entities", validCount)
	
	// Rebuild in-memory indexes
	r.buildIndexes()
	
	return nil
}

// RemoveFromIndex removes an entry from the index
func (r *EntityRepository) RemoveFromIndex(id string) error {
	// This would require modifying the Writer to support index removal
	// For now, we'll need to rebuild the entire index
	logger.Warn("RemoveFromIndex triggering full rebuild for %s", id)
	return r.RebuildIndex()
}

// GetBaseRepository returns the underlying EntityRepository (for high-performance wrapper)
func (r *EntityRepository) GetBaseRepository() *EntityRepository {
	return r
}

// Delete deletes an entity
func (r *EntityRepository) Delete(id string) error {
	// Log to WAL first
	if err := r.wal.LogDelete(id); err != nil {
		return fmt.Errorf("error logging to WAL: %w", err)
	}
	
	// Acquire write lock
	r.lockManager.AcquireEntityLock(id, WriteLock)
	defer r.lockManager.ReleaseEntityLock(id, WriteLock)
	
	// Create temporary file
	tempPath := r.getDataFile() + ".tmp"
	writer, err := NewWriter(tempPath)
	if err != nil {
		return err
	}
	defer writer.Close()
	
	// Read all entities and skip the deleted one
	reader, err := NewReader(r.getDataFile())
	if err != nil {
		return err
	}
	defer reader.Close()
	
	entities, err := reader.GetAllEntities()
	if err != nil {
		return err
	}
	
	// Write all entities except the deleted one
	found := false
	for _, e := range entities {
		if e.ID == id {
			found = true
			continue
		}
		if err := writer.WriteEntity(e); err != nil {
			return err
		}
	}
	
	if !found {
		return fmt.Errorf("entity not found: %s", id)
	}
	
	writer.Close()
	
	// Replace the original file
	if err := os.Rename(tempPath, r.getDataFile()); err != nil {
		return err
	}
	
	// Rebuild indexes
	r.buildIndexes()
	
	// Invalidate cache
	r.cache.Clear()
	
	return nil
}

// Transaction starts a new transaction (currently returns self as transactions are implicit with WAL)
func (r *EntityRepository) Transaction(fn func(tx interface{}) error) error {
	// For simplicity, we'll just execute the function with the repository itself
	return fn(r)
}

// Commit commits the transaction (handled automatically via WAL)
func (r *EntityRepository) Commit(tx interface{}) error {
	// Checkpoint the WAL
	return r.wal.LogCheckpoint()
}

// Rollback rolls back the transaction (handled via WAL replay)
func (r *EntityRepository) Rollback(tx interface{}) error {
	// In case of error, rely on WAL replay during recovery
	return nil
}

// Query operations

// List lists all entities
func (r *EntityRepository) List() ([]*models.Entity, error) {
	startTime := time.Now()
	
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	reader, err := NewReader(r.getDataFile())
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	
	entities, err := reader.GetAllEntities()
	
	// Track read metrics (skip if we're listing metric entities)
	if storageMetrics != nil {
		// Don't track metrics for metric operations to avoid recursion
		hasMetricEntities := false
		if entities != nil {
			for _, entity := range entities {
				if strings.HasPrefix(entity.ID, "metric_") {
					hasMetricEntities = true
					break
				}
			}
		}
		
		if !hasMetricEntities {
			totalSize := int64(0)
			if entities != nil {
				for _, entity := range entities {
					totalSize += int64(len(entity.Content))
				}
			}
			storageMetrics.TrackRead("list_entities", totalSize, time.Since(startTime), err)
		}
	}
	
	return entities, err
}

// ListByTag lists entities with a specific tag
func (r *EntityRepository) ListByTag(tag string) ([]*models.Entity, error) {
	startTime := time.Now()
	logger.Trace("ListByTag: %s", tag)
	
	// Check cache first
	cacheKey := fmt.Sprintf("tag:%s", tag)
	if cached, found := r.cache.Get(cacheKey); found {
		logger.Trace("Cache hit for tag: %s", tag)
		entities := cached.([]*models.Entity)
		
		// Track cache hit (skip metric-related tags to avoid recursion)
		if storageMetrics != nil && !strings.HasPrefix(tag, "name:") && !strings.HasPrefix(tag, "type:metric") {
			storageMetrics.TrackCacheOperation("tag_query", true)
			// Still track the read with 0 duration since it was from cache
			totalSize := int64(0)
			for _, entity := range entities {
				totalSize += int64(len(entity.Content))
			}
			storageMetrics.TrackRead("list_by_tag_cached", totalSize, time.Since(startTime), nil)
		}
		
		return entities, nil
	}
	
	logger.Trace("Cache miss for tag: %s", tag)
	
	var matchingEntityIDs []string
	
	if r.useShardedIndex {
		// Use sharded index for better concurrency
		logger.Trace("Using sharded index for tag: %s", tag)
		
		// Direct lookup first
		directMatches := r.shardedTagIndex.GetEntitiesForTag(tag)
		
		// Also check for temporal versions of the tag
		// This requires a scan, but it's parallelized across shards
		temporalMatches := r.shardedTagIndex.OptimizedListByTag(tag, true)
		
		// Combine and deduplicate
		seen := make(map[string]bool)
		for _, id := range directMatches {
			if !seen[id] {
				seen[id] = true
				matchingEntityIDs = append(matchingEntityIDs, id)
			}
		}
		for _, id := range temporalMatches {
			if !seen[id] {
				seen[id] = true
				matchingEntityIDs = append(matchingEntityIDs, id)
			}
		}
		
		logger.Trace("Sharded index found %d matches", len(matchingEntityIDs))
		
	} else {
		// Legacy path with global lock
		r.mu.RLock()
		
		// For non-temporal searches, we need to find tags that match the requested tag
		// regardless of the timestamp prefix
		matchingEntityIDs = make([]string, 0)
		uniqueEntityIDs := make(map[string]bool)
		
		// Remove verbose debug output
		
		// First check for exact tag match
		if entityIDs, exists := r.tagIndex[tag]; exists {
			logger.Trace("Found exact match: %d entities", len(entityIDs))
			for _, entityID := range entityIDs {
				if !uniqueEntityIDs[entityID] {
					uniqueEntityIDs[entityID] = true
					matchingEntityIDs = append(matchingEntityIDs, entityID)
				}
			}
		}
		
		// Then check for temporal tags with timestamp prefix
		matchCount := 0
		for indexedTag, entityIDs := range r.tagIndex {
			// Skip if this is exactly the tag we already processed
			if indexedTag == tag {
				continue
			}
			
			// Extract the actual tag part (after the timestamp)
			tagParts := strings.SplitN(indexedTag, "|", 2)
			if len(tagParts) == 2 {
				actualTag := tagParts[1]
				
				// Check if the actual tag matches our search tag
				if actualTag == tag {
					matchCount++
					for _, entityID := range entityIDs {
						if !uniqueEntityIDs[entityID] {
							uniqueEntityIDs[entityID] = true
							matchingEntityIDs = append(matchingEntityIDs, entityID)
						}
					}
				}
			}
		}
		logger.Trace("Found %d temporal matches", matchCount)
		
		r.mu.RUnlock()
	}
	
	logger.Debug("ListByTag: %s found %d entities", 
		tag, len(matchingEntityIDs))
	
	if len(matchingEntityIDs) == 0 {
		return []*models.Entity{}, nil
	}
	
	// NOTE: Removed bulk lock acquisition here to prevent deadlocks
	// Locks are properly acquired one-at-a-time in fetchEntitiesWithReader
	
	// Get a reader from the pool
	readerInterface := r.readerPool.Get()
	if readerInterface == nil {
		logger.Trace("Creating new reader")
		reader, err := NewReader(r.getDataFile())
		if err != nil {
			logger.Error("Failed to create reader: %v", err)
			return nil, err
		}
		defer reader.Close()
		return r.fetchEntitiesWithReader(reader, matchingEntityIDs)
	}
	
	logger.Trace("Using pooled reader")
	reader := readerInterface.(*Reader)
	defer r.readerPool.Put(reader)
	
	entities, err := r.fetchEntitiesWithReader(reader, matchingEntityIDs)
	if err != nil {
		logger.Error("Failed to fetch entities: %v", err)
		return nil, err
	}
	
	logger.Trace("Fetched %d entities", len(entities))
	
	// Track metrics (skip metric-related tags to avoid recursion)
	if storageMetrics != nil && !strings.HasPrefix(tag, "name:") && !strings.HasPrefix(tag, "type:metric") {
		storageMetrics.TrackCacheOperation("tag_query", false) // Cache miss
		totalSize := int64(0)
		if entities != nil {
			for _, entity := range entities {
				totalSize += int64(len(entity.Content))
			}
		}
		storageMetrics.TrackRead("list_by_tag", totalSize, time.Since(startTime), err)
	}
	
	// Cache the result
	r.cache.Set(cacheKey, entities)
	return entities, err
}

// fetchEntitiesWithReader is a helper to fetch multiple entities
func (r *EntityRepository) fetchEntitiesWithReader(reader *Reader, entityIDs []string) ([]*models.Entity, error) {
	if len(entityIDs) == 0 {
		return []*models.Entity{}, nil
	}
	
	// For small sets, use sequential processing
	if len(entityIDs) <= 5 {
		entities := make([]*models.Entity, 0, len(entityIDs))
		for _, id := range entityIDs {
			entity, err := reader.GetEntity(id)
			if err == nil {
				entities = append(entities, entity)
			}
		}
		return entities, nil
	}
	
	// For larger sets, use concurrent processing
	entities := make([]*models.Entity, 0, len(entityIDs))
	results := make(chan *models.Entity, len(entityIDs))
	errors := make(chan error, len(entityIDs))
	
	// Use a worker pool to limit concurrency
	const maxWorkers = 10
	numWorkers := maxWorkers
	if len(entityIDs) < maxWorkers {
		numWorkers = len(entityIDs)
	}
	
	// Create work queue
	workQueue := make(chan string, len(entityIDs))
	for _, id := range entityIDs {
		workQueue <- id
	}
	close(workQueue)
	
	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Each worker gets its own reader from the pool
			readerInterface := r.readerPool.Get()
			var workerReader *Reader
			if readerInterface != nil {
				workerReader = readerInterface.(*Reader)
				defer r.readerPool.Put(workerReader)
			} else {
				// Create a new reader if pool is empty
				newReader, err := NewReader(r.getDataFile())
				if err != nil {
					errors <- err
					return
				}
				workerReader = newReader
				defer newReader.Close()
			}
			
			// Process work items
			for id := range workQueue {
				entity, err := workerReader.GetEntity(id)
				if err != nil {
					errors <- err
				} else {
					results <- entity
				}
			}
		}()
	}
	
	// Wait for all workers to complete
	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()
	
	// Collect results
	for entity := range results {
		if entity != nil {
			entities = append(entities, entity)
		}
	}
	
	// Check for any errors
	var firstError error
	for err := range errors {
		if firstError == nil && err != nil {
			firstError = err
		}
	}
	
	return entities, firstError
}

// ListByTags retrieves entities with all specified tags
func (r *EntityRepository) ListByTags(tags []string, matchAll bool) ([]*models.Entity, error) {
	logger.Trace("ListByTags: %d tags, matchAll=%v", len(tags), matchAll)
	
	if len(tags) == 0 {
		return r.List()
	}
	
	// Use index intersection for better performance
	r.mu.RLock()
	
	var entityIDs []string
	
	if matchAll {
		// Get entity IDs for first tag
		logger.Trace("Looking for first tag: %s", tags[0])
		
		// Helper function to find entities by tag (including temporal matches)
		findEntitiesByTag := func(searchTag string) map[string]bool {
			entitySet := make(map[string]bool)
			
			// First check for exact tag match
			if ids, exists := r.tagIndex[searchTag]; exists {
				logger.Trace("Found exact match: %d entities", len(ids))
				for _, id := range ids {
					entitySet[id] = true
				}
			}
			
			// Then check for temporal tags with timestamp prefix
			for indexedTag, ids := range r.tagIndex {
				if indexedTag == searchTag {
					continue // Skip if already processed
				}
				
				// Extract the actual tag part (after the timestamp)
				tagParts := strings.SplitN(indexedTag, "|", 2)
				if len(tagParts) == 2 && tagParts[1] == searchTag {
					logger.Trace("Found temporal match: %d entities", len(ids))
					for _, id := range ids {
						entitySet[id] = true
					}
				}
			}
			
			return entitySet
		}
		
		// Get entities for first tag
		firstTagEntities := findEntitiesByTag(tags[0])
		if len(firstTagEntities) == 0 {
			r.mu.RUnlock()
			return []*models.Entity{}, nil
		}
		
		// Convert to slice for processing
		entityIDs = make([]string, 0, len(firstTagEntities))
		for id := range firstTagEntities {
			entityIDs = append(entityIDs, id)
		}
		logger.Trace("Found %d entities for first tag", len(entityIDs))
		
		// Intersect with remaining tags
		for i := 1; i < len(tags) && len(entityIDs) > 0; i++ {
			tagEntities := findEntitiesByTag(tags[i])
			if len(tagEntities) == 0 {
				r.mu.RUnlock()
				return []*models.Entity{}, nil
			}
			
			// Filter to keep only common IDs
			filtered := make([]string, 0)
			for _, id := range entityIDs {
				if tagEntities[id] {
					filtered = append(filtered, id)
				}
			}
			entityIDs = filtered
			logger.Trace("After intersecting tag %d: %d entities remain", i, len(entityIDs))
		}
	} else {
		// For matchAny, create a set to collect unique entity IDs
		entitySet := make(map[string]bool)
		for _, tag := range tags {
			// First check for exact tag match
			if tagIDs, exists := r.tagIndex[tag]; exists {
				for _, id := range tagIDs {
					entitySet[id] = true
				}
			}
			
			// Then check for temporal tags with timestamp prefix
			for indexedTag, ids := range r.tagIndex {
				if indexedTag == tag {
					continue // Skip if already processed
				}
				
				// Extract the actual tag part (after the timestamp)
				tagParts := strings.SplitN(indexedTag, "|", 2)
				if len(tagParts) == 2 && tagParts[1] == tag {
					for _, id := range ids {
						entitySet[id] = true
					}
				}
			}
		}
		
		// Convert set to slice
		entityIDs = make([]string, 0, len(entitySet))
		for id := range entitySet {
			entityIDs = append(entityIDs, id)
		}
	}
	
	r.mu.RUnlock()
	
	// Fetch the entities
	if len(entityIDs) == 0 {
		return []*models.Entity{}, nil
	}
	
	readerInterface := r.readerPool.Get()
	if readerInterface == nil {
		reader, err := NewReader(r.getDataFile())
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		return r.fetchEntitiesWithReader(reader, entityIDs)
	}
	
	reader := readerInterface.(*Reader)
	defer r.readerPool.Put(reader)
	
	return r.fetchEntitiesWithReader(reader, entityIDs)
}

// Query methods using in-memory indexes

func (r *EntityRepository) ListByTagSQL(tag string) ([]*models.Entity, error) {
	// Binary format doesn't use SQL, just delegate to ListByTag
	return r.ListByTag(tag)
}

func (r *EntityRepository) ListByTagWildcard(pattern string) ([]*models.Entity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	// Convert pattern to prefix matching
	prefix := strings.TrimSuffix(pattern, "*")
	
	var matchingIDs []string
	for tag, ids := range r.tagIndex {
		// For temporal tags, check the part after the pipe
		actualTag := tag
		if parts := strings.SplitN(tag, "|", 2); len(parts) == 2 {
			actualTag = parts[1]
		}
		
		if strings.HasPrefix(actualTag, prefix) {
			matchingIDs = append(matchingIDs, ids...)
		}
	}
	
	// Remove duplicates
	idSet := make(map[string]bool)
	for _, id := range matchingIDs {
		idSet[id] = true
	}
	
	reader, err := NewReader(r.getDataFile())
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	
	entities := make([]*models.Entity, 0, len(idSet))
	for id := range idSet {
		entity, err := reader.GetEntity(id)
		if err == nil {
			entities = append(entities, entity)
		}
	}
	
	return entities, nil
}

func (r *EntityRepository) SearchContent(searchText string) ([]*models.Entity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	searchLower := strings.ToLower(searchText)
	matchingIDs := make(map[string]bool)
	
	// Search in content index
	for key, ids := range r.contentIndex {
		if strings.Contains(strings.ToLower(key), searchLower) {
			for _, id := range ids {
				matchingIDs[id] = true
			}
		}
	}
	
	reader, err := NewReader(r.getDataFile())
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	
	entities := make([]*models.Entity, 0, len(matchingIDs))
	for id := range matchingIDs {
		entity, err := reader.GetEntity(id)
		if err == nil {
			entities = append(entities, entity)
		}
	}
	
	return entities, nil
}

func (r *EntityRepository) SearchContentByType(contentType string) ([]*models.Entity, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	matchingIDs := make(map[string]bool)
	
	// Search in content index for the given type
	for key, ids := range r.contentIndex {
		if strings.HasPrefix(key, contentType+":") {
			for _, id := range ids {
				matchingIDs[id] = true
			}
		}
	}
	
	reader, err := NewReader(r.getDataFile())
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	
	entities := make([]*models.Entity, 0, len(matchingIDs))
	for id := range matchingIDs {
		entity, err := reader.GetEntity(id)
		if err == nil {
			entities = append(entities, entity)
		}
	}
	
	return entities, nil
}

func (r *EntityRepository) QueryAdvanced(conditions map[string]interface{}) ([]*models.Entity, error) {
	// Simple implementation - just filter all entities
	entities, err := r.List()
	if err != nil {
		return nil, err
	}
	
	result := make([]*models.Entity, 0)
	for _, entity := range entities {
		if r.matchesConditions(entity, conditions) {
			result = append(result, entity)
		}
	}
	
	return result, nil
}

func (r *EntityRepository) ListByNamespace(namespace string) ([]*models.Entity, error) {
	// Use namespace index for efficient lookup
	entityIDs := r.namespaceIndex.GetByNamespace(namespace)
	
	if len(entityIDs) == 0 {
		return []*models.Entity{}, nil
	}
	
	// Fetch entities efficiently using reader pool
	readerInterface := r.readerPool.Get()
	if readerInterface == nil {
		reader, err := NewReader(r.getDataFile())
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		return r.fetchEntitiesWithReader(reader, entityIDs)
	}
	
	reader := readerInterface.(*Reader)
	defer r.readerPool.Put(reader)
	
	return r.fetchEntitiesWithReader(reader, entityIDs)
}

// AddContent adds content to an entity
func (r *EntityRepository) AddContent(entityID, contentType, content string) error {
	entity, err := r.GetByID(entityID)
	if err != nil {
		return err
	}
	
	// For the new model, we'll store content as JSON
	var contentData map[string]interface{}
	if len(entity.Content) > 0 {
		json.Unmarshal(entity.Content, &contentData)
	} else {
		contentData = make(map[string]interface{})
	}
	
	contentData[contentType] = content
	jsonData, _ := json.Marshal(contentData)
	entity.Content = jsonData
	entity.AddTag("content:type:" + contentType)
	
	err = r.Update(entity)
	return err
}

// AddTag adds a tag to an entity efficiently without full entity rewrite
func (r *EntityRepository) AddTag(entityID, tag string) error {
	logger.Debug("AddTag: adding tag '%s' to entity %s", tag, entityID)
	
	// Get current entity to verify it exists and check for duplicate tags
	entity, err := r.GetByID(entityID)
	if err != nil {
		logger.Error("Entity %s not found for AddTag: %v", entityID, err)
		return err
	}
	
	// Ensure tag has timestamp (temporal-only system)
	timestampedTag := tag
	if !strings.Contains(tag, "|") {
		timestampedTag = fmt.Sprintf("%s|%s", models.NowString(), tag)
	}
	
	// Check if tag already exists (check both timestamped and non-timestamped versions)
	for _, existingTag := range entity.Tags {
		if existingTag == timestampedTag || existingTag == tag {
			logger.Debug("Tag '%s' already exists on entity %s", tag, entityID)
			return nil // Tag already exists
		}
		// For value tags, allow multiple instances with different timestamps
		// This is essential for temporal metrics tracking
		if strings.HasPrefix(tag, "value:") {
			continue // Allow duplicate value tags with different timestamps
		}
		// For non-value tags, check if the tag content matches (ignoring timestamp)
		if strings.Contains(existingTag, "|") {
			parts := strings.SplitN(existingTag, "|", 2)
			if len(parts) == 2 && parts[1] == tag {
				logger.Debug("Tag content '%s' already exists on entity %s with different timestamp", tag, entityID)
				return nil // Tag content already exists
			}
		}
	}
	
	// Log to WAL first for durability
	entity.Tags = append(entity.Tags, timestampedTag)
	entity.UpdatedAt = models.Now()
	
	if err := r.wal.LogUpdate(entity); err != nil {
		logger.Error("Failed to log AddTag to WAL for entity %s: %v", entityID, err)
		return fmt.Errorf("error logging to WAL: %w", err)
	}
	
	// Acquire write lock
	r.lockManager.AcquireEntityLock(entityID, WriteLock)
	defer r.lockManager.ReleaseEntityLock(entityID, WriteLock)
	
	// Update in-memory entity
	r.mu.Lock()
	if cachedEntity, exists := r.entities[entityID]; exists {
		cachedEntity.Tags = append(cachedEntity.Tags, timestampedTag)
		cachedEntity.UpdatedAt = entity.UpdatedAt
	}
	r.mu.Unlock()
	
	// Update indexes efficiently without database rewrite
	if r.useShardedIndex {
		// Use sharded index for better concurrency
		r.shardedTagIndex.AddTag(timestampedTag, entityID)
		// Also index the non-timestamped version for easier searching
		if strings.Contains(timestampedTag, "|") {
			parts := strings.SplitN(timestampedTag, "|", 2)
			if len(parts) == 2 {
				r.shardedTagIndex.AddTag(parts[1], entityID)
			}
		}
	} else {
		// Legacy path with global lock
		r.mu.Lock()
		r.addToTagIndex(timestampedTag, entityID)
		// Also index the non-timestamped version for easier searching
		if strings.Contains(timestampedTag, "|") {
			parts := strings.SplitN(timestampedTag, "|", 2)
			if len(parts) == 2 {
				r.addToTagIndex(parts[1], entityID)
			}
		}
		r.tagIndexDirty = true
		r.mu.Unlock()
	}
	
	// Add to temporal index if applicable
	if strings.Contains(timestampedTag, "|") {
		parts := strings.SplitN(timestampedTag, "|", 2)
		if len(parts) == 2 {
			if timestampNanos, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
				timestamp := time.Unix(0, timestampNanos)
				r.temporalIndex.AddEntry(entityID, timestampedTag, timestamp)
			}
		}
	}
	
	// Add to namespace index
	r.namespaceIndex.AddTag(entityID, timestampedTag)
	
	// Invalidate cache
	r.cache.Clear()
	
	logger.Debug("AddTag completed successfully for entity %s, tag '%s'", entityID, tag)
	
	// Check if we need to perform checkpoint
	r.checkAndPerformCheckpoint()
	
	return nil
}

// checkAndPerformCheckpoint checks if checkpoint is needed and performs it
func (r *EntityRepository) checkAndPerformCheckpoint() {
	r.checkpointMu.Lock()
	defer r.checkpointMu.Unlock()
	
	// Increment operation count
	r.walOperationCount++
	
	// Check conditions for checkpoint:
	// 1. Every 1000 operations
	// 2. Every 5 minutes
	// 3. WAL file size > 100MB
	shouldCheckpoint := false
	checkpointReason := ""
	
	if r.walOperationCount >= 1000 {
		shouldCheckpoint = true
		checkpointReason = fmt.Sprintf("operation count reached %d", r.walOperationCount)
	} else if time.Since(r.lastCheckpoint) > 5*time.Minute {
		shouldCheckpoint = true
		checkpointReason = fmt.Sprintf("time elapsed: %v", time.Since(r.lastCheckpoint))
	} else {
		// Check WAL file size
		walPath := filepath.Join(r.dataPath, "entitydb.wal")
		if info, err := os.Stat(walPath); err == nil && info.Size() > 100*1024*1024 { // 100MB
			shouldCheckpoint = true
			checkpointReason = fmt.Sprintf("WAL size: %d bytes", info.Size())
		}
	}
	
	if shouldCheckpoint {
		logger.Info("Performing WAL checkpoint (reason: %s)", checkpointReason)
		
		// Track checkpoint metrics
		startTime := time.Now()
		var walSizeBefore int64
		walPath := filepath.Join(r.dataPath, "entitydb.wal")
		if info, err := os.Stat(walPath); err == nil {
			walSizeBefore = info.Size()
		}
		
		// Log checkpoint operation
		if err := r.wal.LogCheckpoint(); err != nil {
			logger.Error("Failed to log checkpoint: %v", err)
			r.storeCheckpointMetric("failed", 0, walSizeBefore, walSizeBefore, checkpointReason)
			return
		}
		
		// Persist all WAL entries to binary file before truncating
		logger.Debug("Persisting WAL entries to binary file")
		if err := r.persistWALEntries(); err != nil {
			logger.Error("Failed to persist WAL entries: %v", err)
			r.storeCheckpointMetric("failed", time.Since(startTime), walSizeBefore, walSizeBefore, checkpointReason)
			return
		}
		
		// Flush all pending writes
		if err := r.writerManager.Flush(); err != nil {
			logger.Error("Failed to flush writes during checkpoint: %v", err)
			r.storeCheckpointMetric("failed", time.Since(startTime), walSizeBefore, walSizeBefore, checkpointReason)
			return
		}
		
		// Force checkpoint to persist everything
		if err := r.writerManager.Checkpoint(); err != nil {
			logger.Error("Failed to checkpoint during WAL truncation: %v", err)
			r.storeCheckpointMetric("failed", time.Since(startTime), walSizeBefore, walSizeBefore, checkpointReason)
			return
		}
		
		// Truncate the WAL
		if err := r.wal.Truncate(); err != nil {
			logger.Error("Failed to truncate WAL: %v", err)
			r.storeCheckpointMetric("failed", time.Since(startTime), walSizeBefore, walSizeBefore, checkpointReason)
			return
		}
		
		// Get WAL size after checkpoint
		var walSizeAfter int64
		if info, err := os.Stat(walPath); err == nil {
			walSizeAfter = info.Size()
		}
		
		// Reset counters
		r.walOperationCount = 0
		r.lastCheckpoint = time.Now()
		
		// Store successful checkpoint metrics
		duration := time.Since(startTime)
		r.storeCheckpointMetric("success", duration, walSizeBefore, walSizeAfter, checkpointReason)
		
		logger.Info("WAL checkpoint completed successfully (duration: %v, size reduced: %d -> %d bytes)", 
			duration, walSizeBefore, walSizeAfter)
	}
}

// storeCheckpointMetric stores WAL checkpoint metrics
func (r *EntityRepository) storeCheckpointMetric(status string, duration time.Duration, sizeBefore, sizeAfter int64, reason string) {
	// Store multiple checkpoint-related metrics
	
	// 1. Checkpoint count
	metricID := fmt.Sprintf("metric_wal_checkpoint_%s_total", status)
	if entity, err := r.GetByID(metricID); err != nil {
		// Create new metric entity
		tags := []string{
			"type:metric",
			"dataset:system",
			fmt.Sprintf("name:wal_checkpoint_%s_total", status),
			"unit:count",
			fmt.Sprintf("description:Total WAL checkpoints with status %s", status),
			"value:1",
			"retention:count:1000",
			"retention:period:86400", // 24 hours
		}
		
		newEntity := &models.Entity{
			ID:      metricID,
			Tags:    tags,
			Content: []byte{},
		}
		
		if err := r.Create(newEntity); err != nil {
			logger.Error("Failed to create checkpoint metric: %v", err)
		}
	} else {
		// Increment counter
		currentValue := 0.0
		for _, tag := range entity.GetTagsWithoutTimestamp() {
			if strings.HasPrefix(tag, "value:") {
				if val, err := strconv.ParseFloat(strings.TrimPrefix(tag, "value:"), 64); err == nil {
					currentValue = val
					break
				}
			}
		}
		valueTag := fmt.Sprintf("value:%.0f", currentValue+1)
		if err := r.AddTag(metricID, valueTag); err != nil {
			logger.Error("Failed to update checkpoint count: %v", err)
		}
	}
	
	// 2. Checkpoint duration (only for successful checkpoints)
	if status == "success" && duration > 0 {
		durationMetricID := "metric_wal_checkpoint_duration_ms"
		durationMillis := float64(duration.Milliseconds())
		
		if _, err := r.GetByID(durationMetricID); err != nil {
			// Create duration metric
			tags := []string{
				"type:metric",
				"dataset:system",
				"name:wal_checkpoint_duration_ms",
				"unit:milliseconds",
				"description:WAL checkpoint duration",
				fmt.Sprintf("value:%.0f", durationMillis),
				"retention:count:100",
				"retention:period:3600",
			}
			
			newEntity := &models.Entity{
				ID:      durationMetricID,
				Tags:    tags,
				Content: []byte{},
			}
			
			if err := r.Create(newEntity); err != nil {
				logger.Error("Failed to create checkpoint duration metric: %v", err)
			}
		} else {
			// Add temporal value
			valueTag := fmt.Sprintf("value:%.0f", durationMillis)
			if err := r.AddTag(durationMetricID, valueTag); err != nil {
				logger.Error("Failed to update checkpoint duration: %v", err)
			}
		}
		
		// 3. Size reduction metric
		if sizeBefore > 0 && sizeAfter >= 0 {
			sizeReductionID := "metric_wal_checkpoint_size_reduction_bytes"
			reduction := float64(sizeBefore - sizeAfter)
			
			if _, err := r.GetByID(sizeReductionID); err != nil {
				tags := []string{
					"type:metric",
					"dataset:system",
					"name:wal_checkpoint_size_reduction_bytes",
					"unit:bytes",
					"description:Bytes freed by WAL checkpoint",
					fmt.Sprintf("value:%.0f", reduction),
					"retention:count:100",
					"retention:period:3600",
				}
				
				newEntity := &models.Entity{
					ID:      sizeReductionID,
					Tags:    tags,
					Content: []byte{},
				}
				
				if err := r.Create(newEntity); err != nil {
					logger.Error("Failed to create size reduction metric: %v", err)
				}
			} else {
				valueTag := fmt.Sprintf("value:%.0f", reduction)
				if err := r.AddTag(sizeReductionID, valueTag); err != nil {
					logger.Error("Failed to update size reduction: %v", err)
				}
			}
		}
	}
}

// RemoveTag removes a tag from an entity
func (r *EntityRepository) RemoveTag(entityID, tag string) error {
	entity, err := r.GetByID(entityID)
	if err != nil {
		return err
	}
	
	// Remove tag
	filtered := make([]string, 0)
	for _, existingTag := range entity.Tags {
		if existingTag != tag {
			filtered = append(filtered, existingTag)
		}
	}
	
	entity.Tags = filtered
	err = r.Update(entity)
	return err
}

// Stub implementations for unimplemented methods

func (r *EntityRepository) ListByExpression(expression string) ([]*models.Entity, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *EntityRepository) ListByMetadata(key, value string) ([]*models.Entity, error) {
	return nil, fmt.Errorf("not implemented")
}

func (r *EntityRepository) ListContentByType(contentType string) ([]models.ContentItem, error) {
	entities, err := r.List()
	if err != nil {
		return nil, err
	}
	
	var content []models.ContentItem
	// In the new model, we need to check tags for content type
	for _, entity := range entities {
		hasType := false
		for _, tag := range entity.Tags {
			if tag == "content:type:" + contentType {
				hasType = true
				break
			}
		}
		if hasType && len(entity.Content) > 0 {
			// Create a ContentItem for backward compatibility
			content = append(content, models.ContentItem{
				Type:  contentType,
				Value: string(entity.Content),
			})
		}
	}
	
	return content, nil
}

// Relationship operations (stub implementations)

func (r *EntityRepository) CreateRelationship(rel interface{}) error {
	relationship, ok := rel.(*models.EntityRelationship)
	if !ok {
		return fmt.Errorf("invalid relationship type")
	}
	if relationship.ID == "" {
		relationship.ID = "rel_" + models.GenerateUUID()
	}
	relationship.CreatedAt = models.Now()
	
	// Store as special entity
	entity := &models.Entity{
		ID:        relationship.ID,
		Tags:      []string{},
		CreatedAt: models.Now(),
		UpdatedAt: models.Now(),
	}
	entity.AddTag("type:relationship")
	entity.AddTagWithValue("_relationship", relationship.RelationshipType)
	entity.AddTagWithValue("_source", relationship.SourceID)
	entity.AddTagWithValue("_target", relationship.TargetID)
	
	// Store relationship data as JSON content
	relData := map[string]string{
		"relationship_type": relationship.RelationshipType,
		"source_id":         relationship.SourceID,
		"target_id":         relationship.TargetID,
	}
	jsonData, _ := json.Marshal(relData)
	entity.Content = jsonData
	entity.AddTag("content:type:relationship")
	
	err := r.Create(entity)
	return err
}

func (r *EntityRepository) GetRelationshipByID(id string) (interface{}, error) {
	logger.Trace("GetRelationshipByID: %s", id)
	entity, err := r.GetByID(id)
	if err != nil {
		logger.Trace("Failed to get entity: %v", err)
		return nil, err
	}
	
	hasTypeRelationship := r.hasTag(entity, "type:relationship")
	hasRelationshipWildcard := r.hasTag(entity, "_relationship:*")
	
	if !hasTypeRelationship && !hasRelationshipWildcard {
		logger.Trace("Entity is not a relationship")
		return nil, fmt.Errorf("entity is not a relationship")
	}
	
	rel := &models.EntityRelationship{
		ID: entity.ID,
	}
	
	// First, try to extract from JSON content if available
	if len(entity.Content) > 0 {
		var relData map[string]interface{}
		if err := json.Unmarshal(entity.Content, &relData); err == nil {
			// Extract fields from JSON
			if v, ok := relData["relationship_type"].(string); ok {
				rel.RelationshipType = v
			}
			if v, ok := relData["source_id"].(string); ok {
				rel.SourceID = v
			}
			if v, ok := relData["target_id"].(string); ok {
				rel.TargetID = v
			}
		}
	}
	
	// Extract relationship data from entity tags (can override JSON if needed)
	for _, tag := range entity.Tags {
		// Handle temporal tags by extracting the actual tag part
		actualTag := tag
		parts := strings.SplitN(tag, "|", 2)
		if len(parts) == 2 {
			actualTag = parts[1]
		}
		
		if strings.HasPrefix(actualTag, "_relationship:") {
			relType := strings.TrimPrefix(actualTag, "_relationship:")
			if relType != "" {
				rel.RelationshipType = relType
			}
		} else if strings.HasPrefix(actualTag, "_source:") {
			sourceID := strings.TrimPrefix(actualTag, "_source:")
			if sourceID != "" {
				rel.SourceID = sourceID
			}
		} else if strings.HasPrefix(actualTag, "_target:") {
			targetID := strings.TrimPrefix(actualTag, "_target:")
			if targetID != "" {
				rel.TargetID = targetID
			}
		}
	}
	
	logger.Trace("Converted to relationship: %s -> %s (%s)", 
		rel.SourceID, rel.TargetID, rel.RelationshipType)
	
	// Set the Type field as well (for compatibility)
	if rel.Type == "" {
		rel.Type = rel.RelationshipType
	}
	
	// Validate that we have all required fields
	if rel.SourceID == "" || rel.TargetID == "" {
		return nil, fmt.Errorf("relationship missing required fields: source=%s, target=%s", 
			rel.SourceID, rel.TargetID)
	}
	
	return rel, nil
}

func (r *EntityRepository) GetRelationshipsBySource(sourceID string) ([]interface{}, error) {
	searchTag := "_source:" + sourceID
	logger.Debug("GetRelationshipsBySource: %s", sourceID)
	logger.Debug("[RELATIONSHIPS] About to search for tag: %s", searchTag)
	
	entities, err := r.ListByTag(searchTag)
	logger.Debug("[RELATIONSHIPS] ListByTag returned, err=%v", err)
	if err != nil {
		logger.Error("ListByTag failed: %v", err)
		return nil, err
	}
	
	logger.Debug("Found %d entities with source %s", len(entities), sourceID)
	
	relationships := make([]interface{}, 0)
	for _, entity := range entities {
		rel, err := r.GetRelationshipByID(entity.ID)
		if err == nil {
			relationships = append(relationships, rel)
		} else {
			logger.Trace("Failed to convert entity %s: %v", entity.ID, err)
		}
	}
	
	logger.Debug("Returning %d relationships for source %s", len(relationships), sourceID)
	return relationships, nil
}

func (r *EntityRepository) GetRelationshipsByTarget(targetID string) ([]interface{}, error) {
	entities, err := r.ListByTag("_target:" + targetID)
	if err != nil {
		return nil, err
	}
	
	relationships := make([]interface{}, 0)
	for _, entity := range entities {
		rel, err := r.GetRelationshipByID(entity.ID)
		if err == nil {
			relationships = append(relationships, rel)
		}
	}
	
	return relationships, nil
}

func (r *EntityRepository) DeleteRelationship(id string) error {
	// Simply delete the entity
	return r.Delete(id)
}

// Helper functions

func (r *EntityRepository) hasTag(entity *models.Entity, tag string) bool {
	if strings.HasSuffix(tag, "*") {
		// Wildcard matching - check if tag (after timestamp) starts with prefix
		prefix := strings.TrimSuffix(tag, "*")
		for _, t := range entity.Tags {
			// Extract the tag part after the timestamp
			parts := strings.SplitN(t, "|", 2)
			actualTag := t
			if len(parts) == 2 {
				actualTag = parts[1]
			}
			if strings.HasPrefix(actualTag, prefix) {
				return true
			}
		}
		return false
	} else {
		// Exact matching - check if tag (after timestamp) matches exactly
		for _, t := range entity.Tags {
			// For temporal tags, check the part after the pipe
			if strings.HasSuffix(t, "|"+tag) {
				return true
			}
			// Also check exact match for backward compatibility
			if t == tag {
				return true
			}
		}
		return false
	}
}

func (r *EntityRepository) matchesConditions(entity *models.Entity, conditions map[string]interface{}) bool {
	for key, value := range conditions {
		switch key {
		case "tag":
			if v, ok := value.(string); ok && !r.hasTag(entity, v) {
				return false
			}
		case "content_type":
			if v, ok := value.(string); ok {
				found := false
				// Check for content type in tags
				for _, tag := range entity.Tags {
					if tag == "content:type:" + v {
						found = true
						break
					}
				}
				if !found {
					return false
				}
			}
		}
	}
	return true
}

// Temporal operations
func (r *EntityRepository) GetEntityAsOf(id string, timestamp time.Time) (*models.Entity, error) {
	// Get current entity
	entity, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}
	
	// Get tags as of timestamp
	temporalTags := r.temporalIndex.GetEntityAsOf(id, timestamp)
	if temporalTags != nil {
		// Build entity snapshot
		snapshot := &models.Entity{
			ID:        entity.ID,
			Tags:      temporalTags,
			Content:   entity.Content,
			CreatedAt: entity.CreatedAt,
			UpdatedAt: entity.UpdatedAt,
		}
		return snapshot, nil
	}
	
	// Fallback to current entity
	return entity, nil
}

func (r *EntityRepository) GetEntityHistory(id string, limit int) ([]*models.EntityChange, error) {
	// Get temporal entries for this entity
	// For now, just use the current time as the end point
	to := time.Now()
	from := to.Add(-24 * 365 * time.Hour) // Go back one year
	entries := r.temporalIndex.GetEntityHistory(id, from, to)
	
	// Convert temporal entries to EntityChange objects
	changes := make([]*models.EntityChange, 0, len(entries))
	
	for i, entry := range entries {
		if i >= limit && limit > 0 {
			break
		}
		
		change := &models.EntityChange{
			Type:      "tag_change",
			Timestamp: models.Now(),
			NewValue:  entry.Tag,
		}
		
		// Try to find the previous value
		if i > 0 {
			change.OldValue = entries[i-1].Tag
		}
		
		changes = append(changes, change)
	}
	
	return changes, nil
}

func (r *EntityRepository) GetRecentChanges(limit int) ([]*models.EntityChange, error) {
	// Get entity IDs that changed recently (within the last day)
	since := time.Now().Add(-24 * time.Hour)
	entityIDs := r.temporalIndex.GetRecentChanges(since)
	
	// Fetch the entities efficiently
	if len(entityIDs) == 0 {
		return []*models.EntityChange{}, nil
	}
	
	// Convert entity IDs to EntityChange objects
	changes := make([]*models.EntityChange, 0, len(entityIDs))
	
	// Get temporal entries for these entities
	for i, entityID := range entityIDs {
		if i >= limit && limit > 0 {
			break
		}
		
		// Get the most recent change for this entity
		entries := r.temporalIndex.GetEntityHistory(entityID, since, time.Now())
		if len(entries) > 0 {
			// Take the most recent entry
			entry := entries[len(entries)-1]
			change := &models.EntityChange{
				Type:      "tag_change",
				Timestamp: models.Now(),
				NewValue:  entry.Tag,
			}
			changes = append(changes, change)
		}
	}
	
	return changes, nil
}

func (r *EntityRepository) GetEntityDiff(id string, t1, t2 time.Time) (*models.Entity, *models.Entity, error) {
	// Get entity states at both timestamps
	before, err := r.GetEntityAsOf(id, t1)
	if err != nil {
		return nil, nil, err
	}
	
	after, err := r.GetEntityAsOf(id, t2)
	if err != nil {
		return nil, nil, err
	}
	
	// Return both states
	return before, after, nil
}

// InitializeWAL initializes the WAL for crash recovery
func (r *EntityRepository) InitializeWAL(path string) error {
	// This is already done in NewEntityRepository
	return nil
}

// ReplayWAL replays the WAL entries for crash recovery
func (r *EntityRepository) ReplayWAL() error {
	if r.wal == nil {
		return fmt.Errorf("WAL not initialized")
	}
	
	logger.Debug("Replaying WAL entries")
	
	count := 0
	err := r.wal.Replay(func(entry WALEntry) error {
		count++
		
		switch entry.OpType {
		case WALOpCreate, WALOpUpdate:
			// Reconstruct entity and write it
			if entry.Entity != nil {
				// Get the writer
				writer, err := NewWriter(r.getDataFile())
				if err != nil {
					return err
				}
				defer writer.Close()
				
				if err := writer.WriteEntity(entry.Entity); err != nil {
					return err
				}
				
				// Update indexes
				r.updateIndexes(entry.Entity)
			}
			
		case WALOpDelete:
			// Handle deletions if implemented
			logger.Debug("Delete operation not yet implemented for entity %s", entry.EntityID)
			
		case WALOpCheckpoint:
			// Checkpoint reached, can truncate WAL up to this point
			logger.Debug("Checkpoint reached")
		}
		
		return nil
	})
	
	if err != nil {
		return fmt.Errorf("error replaying WAL: %w", err)
	}
	
	logger.Info("Replayed %d WAL entries", count)
	
	// After successful replay, checkpoint and truncate
	if count > 0 {
		if err := r.wal.LogCheckpoint(); err != nil {
			logger.Error("Failed to log checkpoint: %v", err)
		}
		
		if err := r.wal.Truncate(); err != nil {
			logger.Error("Failed to truncate WAL: %v", err)
		}
	}
	
	return nil
}

// Query returns a new EntityQuery builder
func (r *EntityRepository) Query() *models.EntityQuery {
	return models.NewEntityQuery(r)
}

// RepairIndex attempts to fix corrupted index entries
func (r *EntityRepository) RepairIndex() error {
	writer, err := r.writerManager.GetWriter()
	if err != nil {
		return fmt.Errorf("failed to get writer: %w", err)
	}
	defer r.writerManager.ReleaseWriter()
	
	return writer.RepairIndex()
}

// ReindexTags rebuilds all tag indexes from scratch
func (r *EntityRepository) ReindexTags() error {
	logger.Info("Starting tag reindexing")
	
	// Acquire write lock to prevent concurrent access during reindexing
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Clear existing indexes
	r.tagIndex = make(map[string][]string)
	r.contentIndex = make(map[string][]string)
	r.temporalIndex = NewTemporalIndex()
	r.namespaceIndex = NewNamespaceIndex()
	r.entities = make(map[string]*models.Entity)
	
	logger.Trace("Cleared existing indexes")
	
	// Create a new reader to read all entities
	reader, err := NewReader(r.getDataFile())
	if err != nil {
		logger.Error("Failed to create reader for reindexing: %v", err)
		return fmt.Errorf("failed to create reader: %w", err)
	}
	defer reader.Close()
	
	// Read all entities from disk
	entities, err := reader.GetAllEntities()
	if err != nil {
		logger.Error("Failed to read entities for reindexing: %v", err)
		return fmt.Errorf("failed to read entities: %w", err)
	}
	
	logger.Info("Read %d entities for reindexing", len(entities))
	
	// Rebuild indexes for each entity
	for i, entity := range entities {
		// Store entity in memory cache
		r.entities[entity.ID] = entity
		
		// Update tag index
		for _, tag := range entity.Tags {
			// Always index the full tag (with timestamp)
			r.tagIndex[tag] = append(r.tagIndex[tag], entity.ID)
			
			// Handle temporal tags
			if strings.Contains(tag, "|") {
				parts := strings.SplitN(tag, "|", 2)
				if len(parts) == 2 {
					// Try to parse timestamp for temporal index as nanosecond epoch
					if timestampNanos, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
						timestamp := time.Unix(0, timestampNanos)
						r.temporalIndex.AddEntry(entity.ID, tag, timestamp)
					}
					
					// Index the actual tag part too (without timestamp)
					actualTag := parts[1]
					r.tagIndex[actualTag] = append(r.tagIndex[actualTag], entity.ID)
				}
			}
			
			// Add to namespace index
			r.namespaceIndex.AddTag(entity.ID, tag)
		}
		
		// Update content index
		if len(entity.Content) > 0 {
			contentStr := string(entity.Content)
			r.contentIndex[contentStr] = append(r.contentIndex[contentStr], entity.ID)
		}
		
		// Log progress for large datasets
		if (i+1)%1000 == 0 {
			logger.Info("Reindexed %d/%d entities", i+1, len(entities))
		}
	}
	
	// Clear the query cache since indexes have changed
	r.cache.Clear()
	
	logger.Info("Tag reindexing complete: %d entities indexed", len(entities))
	return nil
}

// replayWAL replays the WAL to rebuild indexes for any operations not yet in the data file
func (r *EntityRepository) replayWAL() error {
	if r.wal == nil {
		return fmt.Errorf("WAL not initialized")
	}
	
	entitiesReplayed := 0
	err := r.wal.Replay(func(entry WALEntry) error {
		switch entry.OpType {
		case WALOpCreate, WALOpUpdate:
			if entry.Entity != nil {
				// Add to in-memory cache
				r.mu.Lock()
				r.entities[entry.EntityID] = entry.Entity
				
				// Update tag index - use the updateIndexes method for consistency
				r.updateIndexes(entry.Entity)
				
				r.mu.Unlock()
				entitiesReplayed++
			}
			
		case WALOpDelete:
			// Remove from indexes
			r.mu.Lock()
			if entity, exists := r.entities[entry.EntityID]; exists {
				// Remove from tag index
				for _, tag := range entity.Tags {
					if ids, ok := r.tagIndex[tag]; ok {
						newIDs := make([]string, 0, len(ids))
						for _, id := range ids {
							if id != entry.EntityID {
								newIDs = append(newIDs, id)
							}
						}
						if len(newIDs) > 0 {
							r.tagIndex[tag] = newIDs
						} else {
							delete(r.tagIndex, tag)
						}
					}
				}
				
				// Remove from cache
				delete(r.entities, entry.EntityID)
			}
			r.mu.Unlock()
		}
		return nil
	})
	
	if err != nil {
		return err
	}
	
	logger.Info("WAL replay complete: %d entities processed", entitiesReplayed)
	return nil
}

// persistWALEntries persists all Write-Ahead Log entries to the binary storage file.
// This is a critical function that ensures durability by writing all WAL entries
// to permanent storage before the WAL can be truncated.
//
// The function performs the following steps:
//   1. Obtains a writer instance directly (bypassing automatic checkpoints)
//   2. Replays all WAL entries sequentially
//   3. Writes each Create/Update operation to the binary file
//   4. Skips Delete operations (handled separately via tombstones)
//   5. Syncs the writer to ensure all data is on disk
//
// This function is called during checkpoint operations to prevent data loss.
// It's essential that this completes successfully before WAL truncation.
func (r *EntityRepository) persistWALEntries() error {
	if r.wal == nil {
		return fmt.Errorf("WAL not initialized")
	}
	
	logger.Debug("Starting WAL persistence")
	entitiesPersisted := 0
	
	// Get the writer directly to avoid checkpoint recursion
	// We must not trigger another checkpoint while persisting WAL entries
	writer, err := r.writerManager.GetWriter()
	if err != nil {
		return fmt.Errorf("failed to get writer: %w", err)
	}
	defer r.writerManager.ReleaseWriter()
	
	// Track which entities need to be persisted
	entitiesToPersist := make(map[string]bool)
	
	// First pass: identify all entities mentioned in the WAL
	err = r.wal.Replay(func(entry WALEntry) error {
		switch entry.OpType {
		case WALOpCreate, WALOpUpdate:
			entitiesToPersist[entry.EntityID] = true
		case WALOpDelete:
			// Mark for deletion
			entitiesToPersist[entry.EntityID] = false
		}
		return nil
	})
	
	if err != nil {
		return fmt.Errorf("failed to scan WAL: %w", err)
	}
	
	// Second pass: persist the current in-memory state of each entity
	for entityID, shouldPersist := range entitiesToPersist {
		if !shouldPersist {
			// Handle deletions if needed
			logger.Trace("Skipping deleted entity %s", entityID)
			continue
		}
		
		// Get the current in-memory state with all accumulated tags
		r.mu.RLock()
		currentEntity, exists := r.entities[entityID]
		r.mu.RUnlock()
		
		if !exists {
			logger.Warn("Entity %s in WAL but not in memory, skipping", entityID)
			continue
		}
		
		// Write the current state to binary file
		if err := writer.WriteEntity(currentEntity); err != nil {
			logger.Error("Failed to persist entity %s: %v", entityID, err)
			return fmt.Errorf("failed to persist entity %s: %w", entityID, err)
		}
		entitiesPersisted++
		logger.Trace("Persisted entity %s with %d tags (current state)", entityID, len(currentEntity.Tags))
	}
	
	// Sync the writer to ensure all data is on disk
	if err := writer.file.Sync(); err != nil {
		return fmt.Errorf("failed to sync writer: %w", err)
	}
	
	logger.Info("WAL persistence complete: %d entities persisted", entitiesPersisted)
	return nil
}

// VerifyIndexHealth checks if the tag index is consistent with the entities
func (r *EntityRepository) VerifyIndexHealth() error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	// Count unique entities in tag index with detailed tracking
	indexedEntities := make(map[string]int) // entity -> tag count
	totalTagEntries := 0
	
	if r.useShardedIndex && r.shardedTagIndex != nil {
		// Use sharded index
		allTags := r.shardedTagIndex.GetAllTags()
		for tag, entityIDs := range allTags {
			for _, id := range entityIDs {
				indexedEntities[id]++
				totalTagEntries++
			}
			logger.Trace("Tag %s: %d entities", tag, len(entityIDs))
		}
	} else {
		// Use regular index
		for tag, entityIDs := range r.tagIndex {
			for _, id := range entityIDs {
				indexedEntities[id]++
				totalTagEntries++
			}
			logger.Trace("Tag %s: %d entities", tag, len(entityIDs))
		}
	}
	
	// Count entities in repository
	entityCount := len(r.entities)
	indexCount := len(indexedEntities)
	
	logger.Info("Index health: %d entities, %d in tag index, %d tag entries", 
		entityCount, indexCount, totalTagEntries)
		
	// Debug: Show first few entities in memory
	debugCount := 0
	for id := range r.entities {
		if debugCount < 5 {
			logger.Trace("Entity in memory: %s", id)
			debugCount++
		} else {
			break
		}
	}
	
	if entityCount != indexCount {
		logger.Error("Index mismatch: %d entities, %d in index", entityCount, indexCount)
		
		// Find entities that are missing from index
		missingFromIndex := 0
		for entityID := range r.entities {
			if _, exists := indexedEntities[entityID]; !exists {
				logger.Error("Entity %s not in tag index", entityID)
				missingFromIndex++
			}
		}
		
		// Find entities in index but not in repository
		missingFromRepo := 0
		for entityID := range indexedEntities {
			if _, exists := r.entities[entityID]; !exists {
				logger.Error("Entity %s in index but not in repository", entityID)
				missingFromRepo++
			}
		}
		
		return fmt.Errorf("index mismatch: %d entities, %d in index (missing from index: %d, from repo: %d)", 
			entityCount, indexCount, missingFromIndex, missingFromRepo)
	}
	
	logger.Info("Index health check passed: %d entities indexed, all repository layers synchronized", entityCount)
	return nil
}

// RepairIndexes rebuilds the tag indexes from the entity data
func (r *EntityRepository) RepairIndexes() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	logger.Info("Starting index repair")
	
	// Clear existing indexes
	r.tagIndex = make(map[string][]string)
	r.contentIndex = make(map[string][]string)
	r.temporalIndex = NewTemporalIndex()
	r.namespaceIndex = NewNamespaceIndex()
	
	// Clear sharded index if enabled
	if r.useShardedIndex {
		r.shardedTagIndex = NewShardedTagIndex()
	}
	
	// Rebuild from entities in memory
	for entityID, entity := range r.entities {
		logger.Trace("Re-indexing entity %s", entityID)
		
		// Re-index all tags
		for _, tag := range entity.Tags {
			if r.useShardedIndex {
				// Add to sharded index
				r.shardedTagIndex.AddTag(tag, entity.ID)
				
				// Also index the non-timestamped version for temporal tags
				if strings.Contains(tag, "|") {
					parts := strings.SplitN(tag, "|", 2)
					if len(parts) == 2 {
						r.shardedTagIndex.AddTag(parts[1], entity.ID)
					}
				}
			} else {
				// Legacy indexing
				r.tagIndex[tag] = append(r.tagIndex[tag], entity.ID)
				
				// CRITICAL: Also index the non-timestamped version for temporal tags
				if strings.Contains(tag, "|") {
					parts := strings.SplitN(tag, "|", 2)
					if len(parts) == 2 {
						actualTag := parts[1]
						r.tagIndex[actualTag] = append(r.tagIndex[actualTag], entity.ID)
						
						// Add to temporal index
						if timestampNanos, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
							timestamp := time.Unix(0, timestampNanos)
							r.temporalIndex.AddEntry(entity.ID, tag, timestamp)
						}
					}
				}
			}
			
			// Add to namespace index
			r.namespaceIndex.AddTag(entity.ID, tag)
		}
		
		// Update content index
		if len(entity.Content) > 0 {
			contentStr := string(entity.Content)
			r.contentIndex[contentStr] = append(r.contentIndex[contentStr], entity.ID)
		}
	}
	
	// Mark index as dirty to force persistence
	r.tagIndexDirty = true
	
	// Save the repaired index immediately
	if err := r.SaveTagIndex(); err != nil {
		logger.Error("Failed to save repaired index: %v", err)
		return fmt.Errorf("failed to save repaired index: %w", err)
	}
	
	logger.Info("Index repair completed: %d entities re-indexed", len(r.entities))
	return nil
}

// detectAndFixIndexCorruption attempts to detect and fix index corruption for a specific entity
func (r *EntityRepository) detectAndFixIndexCorruption(entityID string) error {
	logger.Debug("Checking for index corruption for entity %s", entityID)
	
	// Try to read the entity from disk
	reader, err := NewReader(r.getDataFile())
	if err != nil {
		return fmt.Errorf("failed to create reader: %w", err)
	}
	defer reader.Close()
	
	entity, err := reader.GetEntity(entityID)
	if err != nil || entity == nil {
		// Entity doesn't exist on disk either
		return nil
	}
	
	logger.Warn("Entity %s found on disk but missing from indexes - fixing corruption", entityID)
	
	// Lock for index updates
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Add entity to memory cache
	r.entities[entityID] = entity
	
	// Re-index the entity
	for _, tag := range entity.Tags {
		// Add to tag index
		if r.useShardedIndex {
			r.shardedTagIndex.AddTag(tag, entity.ID)
			// Also index the non-timestamped version for temporal tags
			if strings.Contains(tag, "|") {
				parts := strings.SplitN(tag, "|", 2)
				if len(parts) == 2 {
					r.shardedTagIndex.AddTag(parts[1], entity.ID)
				}
			}
		} else {
			r.addToTagIndex(tag, entity.ID)
			// Also index the non-timestamped version for temporal tags
			if strings.Contains(tag, "|") {
				parts := strings.SplitN(tag, "|", 2)
				if len(parts) == 2 {
					r.addToTagIndex(parts[1], entity.ID)
				}
			}
		}
		
		// Add to temporal index if applicable
		if strings.Contains(tag, "|") {
			parts := strings.SplitN(tag, "|", 2)
			if len(parts) == 2 {
				if timestampNanos, err := strconv.ParseInt(parts[0], 10, 64); err == nil {
					timestamp := time.Unix(0, timestampNanos)
					r.temporalIndex.AddEntry(entity.ID, tag, timestamp)
				}
			}
		}
		
		// Add to namespace index
		r.namespaceIndex.AddTag(entity.ID, tag)
	}
	
	// Update content index
	if len(entity.Content) > 0 {
		contentStr := string(entity.Content)
		r.contentIndex[contentStr] = append(r.contentIndex[contentStr], entity.ID)
	}
	
	// Mark index as dirty
	r.tagIndexDirty = true
	
	logger.Info("Fixed index corruption for entity %s - entity re-indexed", entityID)
	return nil
}

// SaveTagIndex persists the current tag index to disk
func (r *EntityRepository) SaveTagIndex() error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	if !r.tagIndexDirty {
		logger.Trace("Tag index not dirty, skipping save")
		return nil
	}
	
	startTime := time.Now()
	logger.Debug("Saving tag index")
	
	// If using sharded index, convert it to regular index for persistence
	indexToSave := r.tagIndex
	if r.useShardedIndex && r.shardedTagIndex != nil {
		logger.Debug("Converting sharded index to regular index for persistence")
		indexToSave = r.shardedTagIndex.ToMap()
	}
	
	if err := SaveTagIndexV2(r.getDataFile(), indexToSave); err != nil {
		return fmt.Errorf("failed to save tag index: %w", err)
	}
	
	r.tagIndexDirty = false
	r.lastIndexSave = time.Now()
	
	logger.Info("Tag index saved in %v", time.Since(startTime))
	return nil
}

// SaveTagIndexIfNeeded saves the tag index if it's dirty and enough time has passed
func (r *EntityRepository) SaveTagIndexIfNeeded() error {
	// Save every 5 minutes if dirty
	if r.tagIndexDirty && time.Since(r.lastIndexSave) > 5*time.Minute {
		return r.SaveTagIndex()
	}
	return nil
}

// Close closes the repository and saves any pending data
func (r *EntityRepository) Close() error {
	logger.Debug("Closing entity repository")
	
	// Save tag index if dirty
	if err := r.SaveTagIndex(); err != nil {
		logger.Error("Failed to save tag index on close: %v", err)
	}
	
	// Close WAL
	if r.wal != nil {
		if err := r.wal.Close(); err != nil {
			logger.Error("Failed to close WAL: %v", err)
		}
	}
	
	// Close current file
	if r.currentFile != nil {
		if err := r.currentFile.Close(); err != nil {
			logger.Error("Failed to close current file: %v", err)
		}
	}
	
	// Clear reader pool
	r.readerPool = sync.Pool{}
	
	logger.Debug("Entity repository closed")
	return nil
}

// saveEntities writes all entities to disk - exposed for WALOnlyRepository
func (r *EntityRepository) saveEntities() error {
	dataFile := r.getDataFile()
	tempFile := dataFile + ".tmp"
	writer, err := NewWriter(tempFile)
	if err != nil {
		return err
	}
	defer writer.Close()
	
	// Write all entities
	for _, entity := range r.entities {
		if err := writer.WriteEntity(entity); err != nil {
			os.Remove(tempFile)
			return err
		}
	}
	
	// Close writer to finalize the file
	if err := writer.Close(); err != nil {
		os.Remove(tempFile)
		return err
	}
	
	// Atomically replace the old file
	if err := os.Rename(tempFile, dataFile); err != nil {
		os.Remove(tempFile)
		return err
	}
	
	return nil
}

// addToTagIndex adds an entity ID to a tag's index
func (r *EntityRepository) addToTagIndex(tag, entityID string) {
	if ids, exists := r.tagIndex[tag]; exists {
		// Check if already present
		for _, id := range ids {
			if id == entityID {
				return
			}
		}
		r.tagIndex[tag] = append(ids, entityID)
	} else {
		r.tagIndex[tag] = []string{entityID}
	}
}

// removeFromTagIndex removes an entity ID from a tag's index  
func (r *EntityRepository) removeFromTagIndex(tag, entityID string) {
	if ids, exists := r.tagIndex[tag]; exists {
		newIDs := make([]string, 0, len(ids))
		for _, id := range ids {
			if id != entityID {
				newIDs = append(newIDs, id)
			}
		}
		if len(newIDs) > 0 {
			r.tagIndex[tag] = newIDs
		} else {
			delete(r.tagIndex, tag)
		}
	}
}

// RepairWAL repairs the WAL using the recovery manager
func (r *EntityRepository) RepairWAL() error {
	return r.recovery.RepairWAL()
}

// ValidateEntityChecksum validates the checksum of an entity
func (r *EntityRepository) ValidateEntityChecksum(entity *models.Entity) (bool, string) {
	return r.recovery.ValidateChecksum(entity)
}

// CreateEntityBackup creates a backup of an entity
func (r *EntityRepository) CreateEntityBackup(entity *models.Entity) error {
	return r.recovery.CreateBackup(entity)
}

