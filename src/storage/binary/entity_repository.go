package binary

import (
	"entitydb/models"
	"entitydb/cache"
	"entitydb/logger"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// EntityRepository implements models.EntityRepository for binary format
type EntityRepository struct {
	dataPath string
	mu       sync.RWMutex  // Still keep this for backward compatibility
	
	// In-memory indexes for queries
	tagIndex     map[string][]string  // tag -> entity IDs
	contentIndex map[string][]string  // content -> entity IDs
	
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
}

// NewEntityRepository creates a new binary entity repository
func NewEntityRepository(dataPath string) (*EntityRepository, error) {
	logger.Debug("NewEntityRepository called with dataPath: %s", dataPath)
	repo := &EntityRepository{
		dataPath:      dataPath,
		tagIndex:      make(map[string][]string),
		contentIndex:  make(map[string][]string),
		entities:      make(map[string]*models.Entity),
		lockManager:   NewLockManager(),
		writerManager: NewWriterManager(filepath.Join(dataPath, "entities.ebf")),
		cache:          cache.NewQueryCache(1000, 5*time.Minute), // Cache up to 1000 queries for 5 minutes
		temporalIndex:  NewTemporalIndex(),
		namespaceIndex: NewNamespaceIndex(),
	}
	
	// Ensure the data file exists with a proper header before trying to read it
	dataFile := repo.getDataFile()
	if _, err := os.Stat(dataFile); os.IsNotExist(err) {
		logger.Debug("Data file doesn't exist, creating initial file with header")
		// Use the writerManager to create the initial file with header
		_, err := repo.writerManager.GetWriter()
		if err != nil {
			return nil, fmt.Errorf("error creating initial data file: %w", err)
		}
		// The writer creates the file with a header when it's first created
		repo.writerManager.ReleaseWriter()
		logger.Debug("Initial data file created with header")
	}
	
	// Initialize reader pool with binary format readers
	repo.readerPool = sync.Pool{
		New: func() interface{} {
			reader, err := NewReader(repo.getDataFile())
			if err != nil {
				logger.Debug("Error creating reader: %v", err)
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
	
	// Ensure data file exists before building indexes
	if _, err := os.Stat(repo.getDataFile()); os.IsNotExist(err) {
		logger.Debug("Data file doesn't exist, creating initial file...")
		_, err := repo.writerManager.GetWriter()
		if err != nil {
			return nil, fmt.Errorf("error creating initial data file: %w", err)
		}
		repo.writerManager.ReleaseWriter()
		logger.Debug("Initial data file created")
	}
	
	// Build initial indexes
	if err := repo.buildIndexes(); err != nil {
		logger.Warn("Failed to build initial indexes: %v", err)
		// Don't fail initialization - we can still write entities
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
	
	// Build indexes
	for _, entity := range entities {
		// Update tag index
		for _, tag := range entity.Tags {
			r.tagIndex[tag] = append(r.tagIndex[tag], entity.ID)
			
			// Add to temporal index if it's a temporal tag
			if strings.Contains(tag, "|") {
				parts := strings.SplitN(tag, "|", 2)
				if len(parts) == 2 {
					// Try to parse timestamp
					if timestamp, err := time.Parse(time.RFC3339Nano, parts[0]); err == nil {
						r.temporalIndex.AddEntry(entity.ID, tag, timestamp)
					}
				}
			}
			
			// Add to namespace index
			r.namespaceIndex.AddTag(entity.ID, tag)
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
func (r *EntityRepository) updateIndexes(entity *models.Entity) {
	logger.Debug("Updating indexes for entity %s with %d tags", entity.ID, len(entity.Tags))
	
	// Update tag index
	for _, tag := range entity.Tags {
		// Always index the full tag (with timestamp)
		logger.Debug("Indexing tag: '%s' for entity %s", tag, entity.ID)
		r.tagIndex[tag] = append(r.tagIndex[tag], entity.ID)
		
		// Also index the non-timestamped version for easier searching
		if strings.Contains(tag, "|") {
			parts := strings.SplitN(tag, "|", 2)
			if len(parts) == 2 {
				// Try to parse timestamp
				if timestamp, err := time.Parse(time.RFC3339Nano, parts[0]); err == nil {
					r.temporalIndex.AddEntry(entity.ID, tag, timestamp)
					logger.Debug("Added to temporal index: entity %s, tag %s, timestamp %v", 
						entity.ID, tag, timestamp)
				} else {
					logger.Debug("Failed to parse timestamp in tag '%s': %v", tag, err)
				}
				
				// Index the actual tag part too
				actualTag := parts[1]
				logger.Debug("Also indexing non-timestamped version: '%s' for entity %s", 
					actualTag, entity.ID)
				r.tagIndex[actualTag] = append(r.tagIndex[actualTag], entity.ID)
			}
		}
		
		// Add to namespace index
		r.namespaceIndex.AddTag(entity.ID, tag)
	}
	
	// Update content index - store content as string for searching
	if len(entity.Content) > 0 {
		contentStr := string(entity.Content)
		r.contentIndex[contentStr] = append(r.contentIndex[contentStr], entity.ID)
		logger.Debug("Indexed %d bytes of content for entity %s", len(contentStr), entity.ID)
	}
	
	// Dump tag index for debugging
	logger.Debug("Entity %s now indexed with following tags:", entity.ID)
	for indexedTag, ids := range r.tagIndex {
		for _, id := range ids {
			if id == entity.ID {
				logger.Debug("  - %s", indexedTag)
				break
			}
		}
	}
}

// Create creates a new entity with strong durability guarantees
func (r *EntityRepository) Create(entity *models.Entity) error {
	// Always generate a new UUID for entities
	entity.ID = models.GenerateUUID()
	
	timestamp := time.Now()
	entity.CreatedAt = timestamp.Format(time.RFC3339Nano)
	entity.UpdatedAt = entity.CreatedAt
	
	// Ensure all tags have timestamps (temporal-only system)
	timestampedTags := []string{}
	for _, tag := range entity.Tags {
		if !strings.Contains(tag, "|") {
			// Add timestamp if not present (temporal-only system requires all tags to have timestamps)
			timestampedTags = append(timestampedTags, fmt.Sprintf("%s|%s", timestamp.Format(time.RFC3339Nano), tag))
		} else {
			// Keep existing timestamped tags
			timestampedTags = append(timestampedTags, tag)
		}
	}
	entity.Tags = timestampedTags
	
	// Content in the new model is just binary data - no timestamps needed
	
	// Log to WAL first
	if err := r.wal.LogCreate(entity); err != nil {
		return fmt.Errorf("error logging to WAL: %w", err)
	}
	
	// Write entity with locking
	r.lockManager.AcquireEntityLock(entity.ID, WriteLock)
	defer r.lockManager.ReleaseEntityLock(entity.ID, WriteLock)
	
	// Write entity using WriterManager (which handles checkpoints)
	if err := r.writerManager.WriteEntity(entity); err != nil {
		return err
	}
	
	// Update indexes
	r.mu.Lock()
	r.updateIndexes(entity)
	// Store entity in-memory as well
	r.entities[entity.ID] = entity
	r.mu.Unlock()
	
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
	
	logger.Debug("Entity %s successfully created and persisted", entity.ID)
	
	return nil
}

// GetByID gets an entity by ID with improved reliability from in-memory cache
func (r *EntityRepository) GetByID(id string) (*models.Entity, error) {
	logger.Debug("EntityRepository.GetByID: Looking for entity %s", id)
	
	// First check in-memory cache for the entity
	r.mu.RLock()
	entity, exists := r.entities[id]
	r.mu.RUnlock()
	
	if exists {
		logger.Debug("EntityRepository.GetByID: Found entity %s in memory cache", id)
		return entity, nil
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
		logger.Debug("EntityRepository.GetByID: Entity %s not found in index, trying disk", id)
	}
	
	// Force a flush and checkpoint of any pending writes before attempting to read
	// This ensures that we can read immediately after writing
	if err := r.writerManager.Flush(); err != nil {
		logger.Error("EntityRepository.GetByID: Failed to flush writes before reading: %v", err)
		// Continue anyway as we might still find the entity
	}
	
	// Also force a checkpoint to ensure index is updated
	if err := r.writerManager.Checkpoint(); err != nil {
		logger.Error("EntityRepository.GetByID: Failed to checkpoint: %v", err)
		// Continue anyway as we might still find the entity
	}
	
	// Acquire read lock for the entity
	r.lockManager.AcquireEntityLock(id, ReadLock)
	defer r.lockManager.ReleaseEntityLock(id, ReadLock)
	
	// Get a reader from the pool
	readerInterface := r.readerPool.Get()
	if readerInterface == nil {
		logger.Debug("EntityRepository.GetByID: Creating new reader for %s", id)
		reader, err := NewReader(r.getDataFile())
		if err != nil {
			logger.Error("EntityRepository.GetByID: Failed to create reader: %v", err)
			return nil, err
		}
		defer reader.Close()
		
		entity, err := reader.GetEntity(id)
		if err != nil {
			logger.Error("EntityRepository.GetByID: Failed to get entity %s from new reader: %v", id, err)
			return nil, err
		}
		
		if entity != nil {
			logger.Debug("EntityRepository.GetByID: Found entity %s with new reader, %d bytes content, %d tags", 
				id, len(entity.Content), len(entity.Tags))
			
			// Store in memory for future fast access
			r.mu.Lock()
			r.entities[id] = entity
			r.mu.Unlock()
		}
		return entity, nil
	}
	
	logger.Debug("EntityRepository.GetByID: Using pooled reader for %s", id)
	reader := readerInterface.(*Reader)
	defer r.readerPool.Put(reader)
	
	entity, err := reader.GetEntity(id)
	if err != nil {
		logger.Error("EntityRepository.GetByID: Failed to get entity %s from pooled reader: %v", id, err)
		return nil, err
	}
	
	if entity != nil {
		logger.Debug("EntityRepository.GetByID: Found entity %s with pooled reader, %d bytes content, %d tags",
			id, len(entity.Content), len(entity.Tags))
		
		// Store in memory for future fast access
		r.mu.Lock()
		r.entities[id] = entity
		r.mu.Unlock()
	} else {
		logger.Debug("EntityRepository.GetByID: Entity %s not found", id)
	}
	
	return entity, nil
}

// Update updates an existing entity
func (r *EntityRepository) Update(entity *models.Entity) error {
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
	
	timestamp := time.Now()
	entity.UpdatedAt = timestamp.Format(time.RFC3339Nano)
	
	// Ensure all tags have timestamps (temporal-only system)
	timestampedTags := []string{}
	for _, tag := range entity.Tags {
		if !strings.Contains(tag, "|") {
			// Add timestamp if not present (temporal-only system requires all tags to have timestamps)
			timestampedTags = append(timestampedTags, fmt.Sprintf("%s|%s", timestamp.Format(time.RFC3339Nano), tag))
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
	
	return nil
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
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	reader, err := NewReader(r.getDataFile())
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	
	return reader.GetAllEntities()
}

// ListByTag lists entities with a specific tag
func (r *EntityRepository) ListByTag(tag string) ([]*models.Entity, error) {
	logger.Debug("ListByTag called with tag: %s", tag)
	
	// Check cache first
	cacheKey := fmt.Sprintf("tag:%s", tag)
	if cached, found := r.cache.Get(cacheKey); found {
		logger.Debug("ListByTag cache hit for tag: %s", tag)
		return cached.([]*models.Entity), nil
	}
	
	logger.Debug("ListByTag cache miss for tag: %s", tag)
	
	r.mu.RLock()
	logger.Debug("ListByTag acquired read lock for tag index")
	
	// For non-temporal searches, we need to find tags that match the requested tag
	// regardless of the timestamp prefix
	matchingEntityIDs := make([]string, 0)
	uniqueEntityIDs := make(map[string]bool)
	
	// Dump all tags in the index (for debugging)
	logger.Debug("===== Tag Index Contents =====")
	for indexedTag, ids := range r.tagIndex {
		logger.Debug("Tag: '%s' => %d entities", indexedTag, len(ids))
	}
	logger.Debug("=============================")
	
	// First check for exact tag match
	if entityIDs, exists := r.tagIndex[tag]; exists {
		logger.Debug("Found exact tag match for '%s' with %d entities", tag, len(entityIDs))
		for _, entityID := range entityIDs {
			if !uniqueEntityIDs[entityID] {
				uniqueEntityIDs[entityID] = true
				matchingEntityIDs = append(matchingEntityIDs, entityID)
				logger.Debug("Added entity %s from exact match", entityID)
			}
		}
	} else {
		logger.Debug("No exact match found for tag '%s' in index", tag)
	}
	
	// Then check for temporal tags with timestamp prefix
	logger.Debug("Checking for temporal tag matches for '%s'", tag)
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
			logger.Debug("Checking temporal tag: '%s' (actual part: '%s')", indexedTag, actualTag)
			
			// Check if the actual tag matches our search tag
			if actualTag == tag {
				logger.Debug("Found temporal tag match '%s' in '%s' with %d entities", 
					tag, indexedTag, len(entityIDs))
				matchCount++
				for _, entityID := range entityIDs {
					if !uniqueEntityIDs[entityID] {
						uniqueEntityIDs[entityID] = true
						matchingEntityIDs = append(matchingEntityIDs, entityID)
						logger.Debug("Added entity %s from temporal match", entityID)
					}
				}
			}
		}
	}
	logger.Debug("Found %d temporal tag matches for '%s'", matchCount, tag)
	
	r.mu.RUnlock()
	logger.Debug("ListByTag released read lock")
	
	logger.Debug("ListByTag for '%s' found %d matching entities: %v", 
		tag, len(matchingEntityIDs), matchingEntityIDs)
	
	if len(matchingEntityIDs) == 0 {
		logger.Debug("ListByTag returning empty result for tag: %s", tag)
		return []*models.Entity{}, nil
	}
	
	// Acquire read locks for all matching entities
	for _, id := range matchingEntityIDs {
		r.lockManager.AcquireEntityLock(id, ReadLock)
		defer r.lockManager.ReleaseEntityLock(id, ReadLock)
	}
	
	// Get a reader from the pool
	readerInterface := r.readerPool.Get()
	if readerInterface == nil {
		logger.Debug("Creating new reader for entities")
		reader, err := NewReader(r.getDataFile())
		if err != nil {
			logger.Error("Failed to create reader: %v", err)
			return nil, err
		}
		defer reader.Close()
		return r.fetchEntitiesWithReader(reader, matchingEntityIDs)
	}
	
	logger.Debug("Using reader from pool")
	reader := readerInterface.(*Reader)
	defer r.readerPool.Put(reader)
	
	entities, err := r.fetchEntitiesWithReader(reader, matchingEntityIDs)
	if err != nil {
		logger.Error("Failed to fetch entities: %v", err)
		return nil, err
	}
	
	logger.Debug("Successfully fetched %d entities", len(entities))
	
	// Cache the result
	r.cache.Set(cacheKey, entities)
	return entities, err
}

// fetchEntitiesWithReader is a helper to fetch multiple entities
func (r *EntityRepository) fetchEntitiesWithReader(reader *Reader, entityIDs []string) ([]*models.Entity, error) {
	entities := make([]*models.Entity, 0, len(entityIDs))
	
	for _, id := range entityIDs {
		// Acquire entity read lock for each entity
		r.lockManager.AcquireEntityLock(id, ReadLock)
		entity, err := reader.GetEntity(id)
		r.lockManager.ReleaseEntityLock(id, ReadLock)
		
		if err == nil {
			entities = append(entities, entity)
		}
	}
	
	return entities, nil
}

// ListByTags retrieves entities with all specified tags
func (r *EntityRepository) ListByTags(tags []string, matchAll bool) ([]*models.Entity, error) {
	if len(tags) == 0 {
		return r.List()
	}
	
	// Use index intersection for better performance
	r.mu.RLock()
	
	var entityIDs []string
	
	if matchAll {
		// Get entity IDs for first tag
		if ids, exists := r.tagIndex[tags[0]]; exists {
			entityIDs = make([]string, len(ids))
			copy(entityIDs, ids)
		} else {
			r.mu.RUnlock()
			return []*models.Entity{}, nil
		}
		
		// Intersect with remaining tags
		for i := 1; i < len(tags) && len(entityIDs) > 0; i++ {
			if tagIDs, exists := r.tagIndex[tags[i]]; exists {
				// Create a set for fast lookup
				idSet := make(map[string]bool)
				for _, id := range tagIDs {
					idSet[id] = true
				}
				
				// Filter to keep only common IDs
				filtered := make([]string, 0)
				for _, id := range entityIDs {
					if idSet[id] {
						filtered = append(filtered, id)
					}
				}
				entityIDs = filtered
			} else {
				r.mu.RUnlock()
				return []*models.Entity{}, nil
			}
		}
	} else {
		// For matchAny, create a set to collect unique entity IDs
		entitySet := make(map[string]bool)
		for _, tag := range tags {
			if tagIDs, exists := r.tagIndex[tag]; exists {
				for _, id := range tagIDs {
					entitySet[id] = true
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

// AddTag adds a tag to an entity
func (r *EntityRepository) AddTag(entityID, tag string) error {
	entity, err := r.GetByID(entityID)
	if err != nil {
		return err
	}
	
	// Check if tag already exists
	for _, existingTag := range entity.Tags {
		if existingTag == tag {
			return nil // Tag already exists
		}
	}
	
	entity.Tags = append(entity.Tags, tag)
	err = r.Update(entity)
	return err
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
	relationship.CreatedAt = time.Now()
	
	// Store as special entity
	entity := &models.Entity{
		ID:        relationship.ID,
		Tags:      []string{},
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
	}
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
	entity, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}
	
	if !r.hasTag(entity, "_relationship:*") {
		return nil, fmt.Errorf("entity is not a relationship")
	}
	
	rel := &models.EntityRelationship{
		ID: entity.ID,
	}
	
	// Extract relationship data from entity
	for _, tag := range entity.Tags {
		if strings.HasPrefix(tag, "_relationship:") {
			rel.RelationshipType = strings.TrimPrefix(tag, "_relationship:")
		} else if strings.HasPrefix(tag, "_source:") {
			rel.SourceID = strings.TrimPrefix(tag, "_source:")
		} else if strings.HasPrefix(tag, "_target:") {
			rel.TargetID = strings.TrimPrefix(tag, "_target:")
		}
	}
	
	return rel, nil
}

func (r *EntityRepository) GetRelationshipsBySource(sourceID string) ([]interface{}, error) {
	entities, err := r.ListByTag("_source:" + sourceID)
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
			Timestamp: entry.Timestamp,
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
				Timestamp: entry.Timestamp,
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
	
	logger.Info("Replaying WAL entries...")
	
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
	
	logger.Debug("Replayed %d WAL entries", count)
	
	// After successful replay, checkpoint and truncate
	if count > 0 {
		if err := r.wal.LogCheckpoint(); err != nil {
			logger.Debug("Error logging checkpoint: %v", err)
		}
		
		if err := r.wal.Truncate(); err != nil {
			logger.Debug("Error truncating WAL: %v", err)
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
	logger.Info("Starting tag reindexing...")
	
	// Acquire write lock to prevent concurrent access during reindexing
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Clear existing indexes
	r.tagIndex = make(map[string][]string)
	r.contentIndex = make(map[string][]string)
	r.temporalIndex = NewTemporalIndex()
	r.namespaceIndex = NewNamespaceIndex()
	r.entities = make(map[string]*models.Entity)
	
	logger.Debug("Cleared existing indexes")
	
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
					// Try to parse timestamp for temporal index
					if timestamp, err := time.Parse(time.RFC3339Nano, parts[0]); err == nil {
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
			logger.Debug("Reindexed %d/%d entities", i+1, len(entities))
		}
	}
	
	// Clear the query cache since indexes have changed
	r.cache.Clear()
	
	logger.Info("Tag reindexing completed successfully. Indexed %d entities", len(entities))
	return nil
}