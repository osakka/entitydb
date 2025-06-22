package binary

import (
	"entitydb/logger"
	"entitydb/models"
	"strings"
)

// DirectRepositoryWrapper provides non-instrumented access to repository operations
// This prevents circular dependencies when metrics collection needs to store data
type DirectRepositoryWrapper struct {
	underlying     models.EntityRepository
	metricsBackend *MetricsBackend
}

// NewDirectRepositoryWrapper creates a new direct repository wrapper
func NewDirectRepositoryWrapper(repo models.EntityRepository) DirectRepository {
	// Get data path and config from repository if possible
	dataPath := "/opt/entitydb/var" // fallback default
	var metricsBackend *MetricsBackend
	
	if entityRepo, ok := repo.(*EntityRepository); ok {
		dataPath = entityRepo.dataPath
		// Use complete metrics path from config if available
		if entityRepo.config != nil {
			metricsBackend = NewMetricsBackendWithPath(dataPath, entityRepo.config.MetricsFilename)
		}
	}
	
	// Fallback to old method if config not available
	if metricsBackend == nil {
		metricsBackend = NewMetricsBackend(dataPath)
	}
	
	return &DirectRepositoryWrapper{
		underlying:     repo,
		metricsBackend: metricsBackend,
	}
}

// CreateDirect creates an entity without triggering any metrics instrumentation
func (drw *DirectRepositoryWrapper) CreateDirect(entity *models.Entity) error {
	logger.Trace("DirectRepositoryWrapper operating without instrumentation: CreateDirect")
	
	// CRITICAL: Use direct storage without checkpoints or metrics to prevent recursion
	// This bypasses all instrumentation that could trigger more metrics collection
	if entityRepo, ok := drw.underlying.(*EntityRepository); ok {
		return drw.createDirectStorage(entityRepo, entity)
	}
	
	// Fallback for other repository types (should not happen in production)
	return drw.underlying.Create(entity)
}

// createDirectStorage performs direct storage operations without checkpoints or metrics
func (drw *DirectRepositoryWrapper) createDirectStorage(repo *EntityRepository, entity *models.Entity) error {
	// Directly write to storage without any checkpoints or metrics instrumentation
	repo.mu.Lock()
	defer repo.mu.Unlock()
	
	// Add to cache
	repo.entityCache.Put(entity.ID, entity)
	repo.loadedEntityCount++
	
	// Update indexes without metrics tracking
	repo.updateIndexes(entity)
	
	// Write to storage using writer manager directly (no checkpoints)
	if err := repo.writerManager.WriteEntity(entity); err != nil {
		// Remove from cache on failure
		repo.entityCache.Delete(entity.ID)
		repo.loadedEntityCount--
		return err
	}
	
	return nil
}

// GetByIDDirect gets an entity by ID without triggering any metrics instrumentation
func (drw *DirectRepositoryWrapper) GetByIDDirect(id string) (*models.Entity, error) {
	logger.Trace("DirectRepositoryWrapper operating without instrumentation: GetByIDDirect")
	
	// DECOMMISSIONED: All entities now use single source of truth (main database)
	// No more isolated metrics backend - everything goes to main repository
	return drw.underlying.GetByID(id)
}

// ListByTagDirect lists entities by tag without triggering any metrics instrumentation
func (drw *DirectRepositoryWrapper) ListByTagDirect(tag string) ([]*models.Entity, error) {
	logger.Trace("DirectRepositoryWrapper operating without instrumentation: ListByTagDirect")
	
	// DECOMMISSIONED: All entities now use single source of truth (main database)
	// No more isolated metrics backend - everything goes to main repository  
	return drw.underlying.ListByTag(tag)
}

// AddTagDirect adds a tag to an entity without triggering any metrics instrumentation
func (drw *DirectRepositoryWrapper) AddTagDirect(entityID, tag string) error {
	logger.Trace("DirectRepositoryWrapper operating without instrumentation: AddTagDirect")
	
	// CRITICAL: Use direct storage without checkpoints or metrics to prevent recursion
	// This bypasses all instrumentation that could trigger more metrics collection
	if entityRepo, ok := drw.underlying.(*EntityRepository); ok {
		return drw.addTagDirectStorage(entityRepo, entityID, tag)
	}
	
	// Fallback for other repository types (should not happen in production)
	return drw.underlying.AddTag(entityID, tag)
}

// addTagDirectStorage performs direct tag addition without checkpoints or metrics
func (drw *DirectRepositoryWrapper) addTagDirectStorage(repo *EntityRepository, entityID, tag string) error {
	// Get entity from cache or storage
	entity, err := repo.GetByID(entityID)
	if err != nil {
		return err
	}
	
	// Add the temporal tag directly without any instrumentation
	temporalTag := models.FormatTemporalTag(tag)
	entity.Tags = append(entity.Tags, temporalTag)
	entity.UpdatedAt = models.Now()
	
	// Update the entity in cache and storage without checkpoints
	repo.mu.Lock()
	defer repo.mu.Unlock()
	
	// Update cache
	repo.entityCache.Put(entity.ID, entity)
	
	// Update indexes without metrics tracking
	repo.updateIndexes(entity)
	
	// Write to storage using writer manager directly (no checkpoints)
	return repo.writerManager.WriteEntity(entity)
}

// isMetricsEntity checks if an entity is metrics-related
func (drw *DirectRepositoryWrapper) isMetricsEntity(entity *models.Entity) bool {
	for _, tag := range entity.Tags {
		cleanTag := tag
		if pipePos := strings.Index(tag, "|"); pipePos != -1 {
			cleanTag = tag[pipePos+1:]
		}
		
		if strings.HasPrefix(cleanTag, "type:metric") {
			return true
		}
	}
	return false
}

// isMetricsEntityID checks if an entity ID is metrics-related
func (drw *DirectRepositoryWrapper) isMetricsEntityID(entityID string) bool {
	return strings.HasPrefix(entityID, "metric_") || strings.Contains(entityID, "metric")
}

// isMetricsTag checks if a tag is metrics-related
func (drw *DirectRepositoryWrapper) isMetricsTag(tag string) bool {
	cleanTag := tag
	if pipePos := strings.Index(tag, "|"); pipePos != -1 {
		cleanTag = tag[pipePos+1:]
	}
	
	return strings.HasPrefix(cleanTag, "name:") && 
		   (strings.Contains(cleanTag, "metric") || 
		    strings.Contains(cleanTag, "storage") ||
		    strings.Contains(cleanTag, "request") ||
		    strings.Contains(cleanTag, "performance"))
}