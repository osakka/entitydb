package binary

import (
	"entitydb/models"
	"entitydb/logger"
	"fmt"
	"sync"
	"time"
)

// WALOnlyRepository provides O(1) writes by keeping everything in WAL until compaction
type WALOnlyRepository struct {
	*EntityRepository
	walEntities map[string]*models.Entity  // Recent writes from WAL
	walMutex    sync.RWMutex
	lastCompact time.Time
}

// NewWALOnlyRepository creates a repository optimized for write performance
func NewWALOnlyRepository(dataPath string) (*WALOnlyRepository, error) {
	baseRepo, err := NewEntityRepository(dataPath)
	if err != nil {
		return nil, err
	}
	
	repo := &WALOnlyRepository{
		EntityRepository: baseRepo,
		walEntities:     make(map[string]*models.Entity),
		lastCompact:     time.Now(),
	}
	
	// Load WAL entries into memory instead of replaying to disk
	if err := repo.loadWALToMemory(); err != nil {
		return nil, err
	}
	
	// Start background compaction
	go repo.backgroundCompaction()
	
	return repo, nil
}

// Create with O(1) performance - only writes to WAL
func (r *WALOnlyRepository) Create(entity *models.Entity) error {
	r.walMutex.Lock()
	defer r.walMutex.Unlock()
	
	// Write to WAL only
	if err := r.wal.LogCreate(entity); err != nil {
		return err
	}
	
	// Update in-memory state
	r.walEntities[entity.ID] = entity
	
	// Update indexes immediately
	r.updateIndexesForEntity(entity)
	
	return nil
}

// Update with O(1) performance - only writes to WAL
func (r *WALOnlyRepository) Update(entity *models.Entity) error {
	r.walMutex.Lock()
	defer r.walMutex.Unlock()
	
	// Check existence
	if _, err := r.getEntityUnsafe(entity.ID); err != nil {
		return fmt.Errorf("entity not found: %s", entity.ID)
	}
	
	// Write to WAL only
	if err := r.wal.LogUpdate(entity); err != nil {
		return err
	}
	
	// Update in-memory state
	r.walEntities[entity.ID] = entity
	
	// Update indexes immediately
	r.updateIndexesForEntity(entity)
	
	return nil
}

// Get checks WAL first, then disk
func (r *WALOnlyRepository) Get(id string) (*models.Entity, error) {
	r.walMutex.RLock()
	defer r.walMutex.RUnlock()
	
	return r.getEntityUnsafe(id)
}

// getEntityUnsafe gets entity without locking (caller must hold lock)
func (r *WALOnlyRepository) getEntityUnsafe(id string) (*models.Entity, error) {
	// Check WAL entries first
	if entity, exists := r.walEntities[id]; exists {
		return entity, nil
	}
	
	// Fall back to disk
	if entity, exists := r.entities[id]; exists {
		return entity, nil
	}
	
	return nil, fmt.Errorf("entity not found: %s", id)
}

// updateIndexesForEntity updates indexes for a single entity
func (r *WALOnlyRepository) updateIndexesForEntity(entity *models.Entity) {
	// Remove old entries if updating
	if oldEntity, exists := r.entities[entity.ID]; exists {
		for _, tag := range oldEntity.Tags {
			r.removeFromTagIndex(tag, entity.ID)
		}
	}
	
	// Add new entries
	for _, tag := range entity.Tags {
		r.addToTagIndex(tag, entity.ID)
	}
}

// loadWALToMemory loads WAL entries without applying to disk
func (r *WALOnlyRepository) loadWALToMemory() error {
	// TODO: Implement WAL reading to populate walEntities
	// For now, WAL replay happens normally
	return nil
}

// backgroundCompaction periodically merges WAL to disk
func (r *WALOnlyRepository) backgroundCompaction() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		if time.Since(r.lastCompact) > 5*time.Minute && len(r.walEntities) > 100 {
			logger.Info("Starting background compaction of %d WAL entries", len(r.walEntities))
			if err := r.compact(); err != nil {
				logger.Error("Compaction failed: %v", err)
			}
		}
	}
}

// compact merges WAL entries to disk
func (r *WALOnlyRepository) compact() error {
	r.walMutex.Lock()
	defer r.walMutex.Unlock()
	
	if len(r.walEntities) == 0 {
		return nil
	}
	
	startTime := time.Now()
	
	// Apply WAL entries to base repository
	for _, entity := range r.walEntities {
		r.entities[entity.ID] = entity
	}
	
	// Write to disk using base repository method
	if err := r.EntityRepository.saveEntities(); err != nil {
		return err
	}
	
	// Clear WAL entries and reset WAL
	r.walEntities = make(map[string]*models.Entity)
	r.lastCompact = time.Now()
	
	// Truncate WAL file
	if err := r.wal.Truncate(); err != nil {
		logger.Error("Failed to truncate WAL: %v", err)
	}
	
	logger.Info("Compaction completed in %v, merged %d entries", time.Since(startTime), len(r.walEntities))
	return nil
}

// ForceCompact allows manual compaction trigger
func (r *WALOnlyRepository) ForceCompact() error {
	return r.compact()
}