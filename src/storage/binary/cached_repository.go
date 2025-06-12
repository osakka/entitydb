package binary

import (
	"entitydb/logger"
	"entitydb/models"
	"strings"
	"sync"
	"time"
)

// CachedRepository wraps any repository with an in-memory cache
type CachedRepository struct {
	models.EntityRepository
	
	// Entity cache with TTL
	entityCache     *sync.Map // map[string]*cachedEntity
	tagCache        *sync.Map // map[string]*cachedTagResult
	cacheTTL        time.Duration
	
	// Performance metrics
	cacheHits   uint64
	cacheMisses uint64
	
	// Background cleaner
	cleanupTicker *time.Ticker
	done          chan bool
}

type cachedEntity struct {
	entity    *models.Entity
	timestamp time.Time
}

type cachedTagResult struct {
	entities  []*models.Entity
	timestamp time.Time
}

// NewCachedRepository wraps a repository with caching
func NewCachedRepository(baseRepo models.EntityRepository, ttl time.Duration) *CachedRepository {
	if ttl == 0 {
		ttl = 5 * time.Minute // Default 5 minute TTL
	}
	
	repo := &CachedRepository{
		EntityRepository: baseRepo,
		entityCache:     &sync.Map{},
		tagCache:        &sync.Map{},
		cacheTTL:        ttl,
		cleanupTicker:   time.NewTicker(ttl / 2),
		done:            make(chan bool),
	}
	
	// Start background cleanup
	go repo.cleanupExpired()
	
	logger.Info("Created CachedRepository with TTL: %v", ttl)
	return repo
}

// GetByID with caching
func (r *CachedRepository) GetByID(id string) (*models.Entity, error) {
	// Check cache first
	if cached, ok := r.entityCache.Load(id); ok {
		ce := cached.(*cachedEntity)
		if time.Since(ce.timestamp) < r.cacheTTL {
			r.cacheHits++
			return ce.entity, nil
		}
		// Expired, remove it
		r.entityCache.Delete(id)
	}
	
	r.cacheMisses++
	
	// Get from underlying repository
	entity, err := r.EntityRepository.GetByID(id)
	if err != nil {
		return nil, err
	}
	
	// Cache the result
	r.entityCache.Store(id, &cachedEntity{
		entity:    entity,
		timestamp: time.Now(),
	})
	
	return entity, nil
}

// Create invalidates relevant caches
func (r *CachedRepository) Create(entity *models.Entity) error {
	err := r.EntityRepository.Create(entity)
	if err != nil {
		return err
	}
	
	// Add to cache
	r.entityCache.Store(entity.ID, &cachedEntity{
		entity:    entity,
		timestamp: time.Now(),
	})
	
	// Invalidate tag caches that might be affected
	r.invalidateTagCaches(entity.Tags)
	
	return nil
}

// Update invalidates caches
func (r *CachedRepository) Update(entity *models.Entity) error {
	// Get old entity to know which tags to invalidate
	oldEntity, _ := r.EntityRepository.GetByID(entity.ID)
	
	err := r.EntityRepository.Update(entity)
	if err != nil {
		return err
	}
	
	// Update cache
	r.entityCache.Store(entity.ID, &cachedEntity{
		entity:    entity,
		timestamp: time.Now(),
	})
	
	// Invalidate affected tag caches
	if oldEntity != nil {
		r.invalidateTagCaches(oldEntity.Tags)
	}
	r.invalidateTagCaches(entity.Tags)
	
	return nil
}

// Delete invalidates caches
func (r *CachedRepository) Delete(id string) error {
	// Get entity to know which tags to invalidate
	entity, _ := r.EntityRepository.GetByID(id)
	
	err := r.EntityRepository.Delete(id)
	if err != nil {
		return err
	}
	
	// Remove from cache
	r.entityCache.Delete(id)
	
	// Invalidate affected tag caches
	if entity != nil {
		r.invalidateTagCaches(entity.Tags)
	}
	
	return nil
}

// ListByTags with caching
func (r *CachedRepository) ListByTags(tags []string, matchAll bool) ([]*models.Entity, error) {
	// Create cache key
	cacheKey := formatTagCacheKey(tags, matchAll)
	
	// Check cache
	if cached, ok := r.tagCache.Load(cacheKey); ok {
		cr := cached.(*cachedTagResult)
		if time.Since(cr.timestamp) < r.cacheTTL {
			r.cacheHits++
			return cr.entities, nil
		}
		// Expired
		r.tagCache.Delete(cacheKey)
	}
	
	r.cacheMisses++
	
	// Get from underlying repository
	entities, err := r.EntityRepository.ListByTags(tags, matchAll)
	if err != nil {
		return nil, err
	}
	
	// Cache the result
	r.tagCache.Store(cacheKey, &cachedTagResult{
		entities:  entities,
		timestamp: time.Now(),
	})
	
	return entities, nil
}

// ListByTag with caching
func (r *CachedRepository) ListByTag(tag string) ([]*models.Entity, error) {
	// Call the underlying repository's ListByTag directly to use sharded index
	return r.EntityRepository.ListByTag(tag)
}

// List with caching
func (r *CachedRepository) List() ([]*models.Entity, error) {
	// Check if we have it cached
	cacheKey := "_all_entities_"
	
	if cached, ok := r.tagCache.Load(cacheKey); ok {
		cr := cached.(*cachedTagResult)
		if time.Since(cr.timestamp) < r.cacheTTL {
			r.cacheHits++
			return cr.entities, nil
		}
		r.tagCache.Delete(cacheKey)
	}
	
	r.cacheMisses++
	
	// Get from underlying repository
	entities, err := r.EntityRepository.List()
	if err != nil {
		return nil, err
	}
	
	// Cache the result
	r.tagCache.Store(cacheKey, &cachedTagResult{
		entities:  entities,
		timestamp: time.Now(),
	})
	
	// Also populate entity cache while we have the data
	for _, entity := range entities {
		r.entityCache.Store(entity.ID, &cachedEntity{
			entity:    entity,
			timestamp: time.Now(),
		})
	}
	
	return entities, nil
}

// Helper functions

func (r *CachedRepository) invalidateTagCaches(tags []string) {
	// Clear any cached results that might include these tags
	r.tagCache.Range(func(key, value interface{}) bool {
		// For simplicity, clear all tag caches when tags change
		// In production, you'd want more granular invalidation
		if key.(string) != "_all_entities_" {
			r.tagCache.Delete(key)
		}
		return true
	})
	
	// Always invalidate the all entities cache
	r.tagCache.Delete("_all_entities_")
}

func formatTagCacheKey(tags []string, matchAll bool) string {
	mode := "any"
	if matchAll {
		mode = "all"
	}
	return mode + ":" + strings.Join(tags, ",")
}

func (r *CachedRepository) cleanupExpired() {
	for {
		select {
		case <-r.cleanupTicker.C:
			now := time.Now()
			
			// Clean entity cache
			r.entityCache.Range(func(key, value interface{}) bool {
				ce := value.(*cachedEntity)
				if now.Sub(ce.timestamp) > r.cacheTTL {
					r.entityCache.Delete(key)
				}
				return true
			})
			
			// Clean tag cache
			r.tagCache.Range(func(key, value interface{}) bool {
				cr := value.(*cachedTagResult)
				if now.Sub(cr.timestamp) > r.cacheTTL {
					r.tagCache.Delete(key)
				}
				return true
			})
			
		case <-r.done:
			return
		}
	}
}

// GetCacheStats returns cache performance metrics
func (r *CachedRepository) GetCacheStats() (hits, misses uint64) {
	return r.cacheHits, r.cacheMisses
}

// Close stops the background cleanup
func (r *CachedRepository) Close() error {
	r.cleanupTicker.Stop()
	close(r.done)
	return nil
}

// GetUnderlying returns the underlying repository
func (r *CachedRepository) GetUnderlying() models.EntityRepository {
	return r.EntityRepository
}