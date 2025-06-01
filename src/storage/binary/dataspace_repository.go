package binary

import (
	"entitydb/logger"
	"entitydb/models"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"os"
)

// DataspaceRepository implements per-dataspace index isolation
type DataspaceRepository struct {
	*EntityRepository // Embed base repository for entity storage
	
	// Dataspace-specific indexes
	dataspaceIndexes map[string]*DataspaceIndexImpl
	dataspaceConfigs map[string]*models.Dataspace
	dsLock          sync.RWMutex
	
	// Dataspace index directory
	indexPath string
}

// DataspaceIndexImpl implements the DataspaceIndex interface
type DataspaceIndexImpl struct {
	name      string
	indexPath string
	
	// In-memory indexes
	entities  map[string]bool           // Entity IDs in this dataspace
	tagIndex  map[string][]string       // tag -> entity IDs
	
	// Statistics
	stats     models.DataspaceStats
	
	// Synchronization
	mu        sync.RWMutex
}

// NewDataspaceRepository creates a repository with dataspace isolation
func NewDataspaceRepository(dataPath string) (*DataspaceRepository, error) {
	baseRepo, err := NewEntityRepository(dataPath)
	if err != nil {
		return nil, err
	}
	
	indexPath := filepath.Join(dataPath, "dataspaces")
	if err := os.MkdirAll(indexPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create dataspace index directory: %w", err)
	}
	
	repo := &DataspaceRepository{
		EntityRepository:  baseRepo,
		dataspaceIndexes: make(map[string]*DataspaceIndexImpl),
		dataspaceConfigs: make(map[string]*models.Dataspace),
		indexPath:        indexPath,
	}
	
	// Load existing dataspace indexes
	if err := repo.loadDataspaceIndexes(); err != nil {
		logger.Error("Failed to load dataspace indexes: %v", err)
	}
	
	// Rebuild dataspace indexes from existing entities
	if err := repo.rebuildDataspaceIndexes(); err != nil {
		logger.Error("Failed to rebuild dataspace indexes: %v", err)
	}
	
	// Create default dataspace if none exists
	if _, exists := repo.dataspaceIndexes["default"]; !exists {
		defaultDs := &models.Dataspace{
			Name: "default",
			Config: models.DataspaceConfig{
				IndexStrategy: models.IndexStrategyBTree,
				OptimizeFor:   models.OptimizeForReads,
			},
		}
		if err := repo.CreateDataspace(defaultDs); err != nil {
			logger.Error("Failed to create default dataspace: %v", err)
		}
	}
	
	return repo, nil
}

// CreateDataspace creates a new dataspace with its own index
func (r *DataspaceRepository) CreateDataspace(dataspace *models.Dataspace) error {
	r.dsLock.Lock()
	defer r.dsLock.Unlock()
	
	if _, exists := r.dataspaceIndexes[dataspace.Name]; exists {
		return fmt.Errorf("dataspace %s already exists", dataspace.Name)
	}
	
	// Create dataspace index (use .ebf extension so SaveTagIndexV2 creates .idx)
	dsIndex := &DataspaceIndexImpl{
		name:      dataspace.Name,
		indexPath: filepath.Join(r.indexPath, dataspace.Name+".ebf"),
		entities:  make(map[string]bool),
		tagIndex:  make(map[string][]string),
	}
	
	r.dataspaceIndexes[dataspace.Name] = dsIndex
	r.dataspaceConfigs[dataspace.Name] = dataspace
	
	// Save index to disk
	if err := dsIndex.SaveToFile(dsIndex.indexPath); err != nil {
		return fmt.Errorf("failed to save dataspace index: %w", err)
	}
	
	logger.Info("Created dataspace: %s", dataspace.Name)
	return nil
}

// Create entity in a specific dataspace
func (r *DataspaceRepository) Create(entity *models.Entity) error {
	// Extract dataspace from tags
	dataspaceName := r.extractDataspace(entity)
	logger.Debug("Creating entity %s in dataspace: %s", entity.ID, dataspaceName)
	
	// Create in base repository
	if err := r.EntityRepository.Create(entity); err != nil {
		return err
	}
	
	// Add to dataspace index
	r.dsLock.RLock()
	dsIndex, exists := r.dataspaceIndexes[dataspaceName]
	r.dsLock.RUnlock()
	
	if !exists && dataspaceName != "default" {
		// Create dataspace if it doesn't exist
		logger.Info("Creating dataspace '%s' on demand", dataspaceName)
		newDataspace := &models.Dataspace{
			Name: dataspaceName,
			Config: models.DataspaceConfig{
				IndexStrategy: models.IndexStrategyBTree,
				OptimizeFor:   models.OptimizeForReads,
			},
		}
		if err := r.CreateDataspace(newDataspace); err != nil {
			logger.Error("Failed to create dataspace: %v", err)
		} else {
			// Get the newly created index
			r.dsLock.RLock()
			dsIndex, exists = r.dataspaceIndexes[dataspaceName]
			r.dsLock.RUnlock()
		}
	}
	
	if exists {
		if err := dsIndex.AddEntity(entity); err != nil {
			logger.Error("Failed to add entity to dataspace index: %v", err)
		} else {
			// Save index after adding entity
			if err := dsIndex.SaveToFile(dsIndex.indexPath); err != nil {
				logger.Error("Failed to save dataspace index: %v", err)
			}
		}
		logger.Debug("Added entity %s to dataspace '%s' index", entity.ID, dataspaceName)
	}
	
	return nil
}

// ListByTags with dataspace awareness
func (r *DataspaceRepository) ListByTags(tags []string, matchAll bool) ([]*models.Entity, error) {
	// Check if query is dataspace-specific
	dataspaceName := ""
	filteredTags := []string{}
	
	for _, tag := range tags {
		// Handle temporal tags
		actualTag := tag
		if parts := strings.SplitN(tag, "|", 2); len(parts) == 2 {
			actualTag = parts[1]
		}
		
		if strings.HasPrefix(actualTag, "dataspace:") || strings.HasPrefix(actualTag, "dataspace:") {
			dataspaceName = strings.TrimPrefix(strings.TrimPrefix(actualTag, "dataspace:"), "dataspace:")
		} else {
			filteredTags = append(filteredTags, actualTag)
		}
	}
	
	// If dataspace-specific, use dataspace index
	if dataspaceName != "" {
		logger.Debug("Dataspace query for '%s' with tags: %v", dataspaceName, filteredTags)
		
		r.dsLock.RLock()
		dsIndex, exists := r.dataspaceIndexes[dataspaceName]
		r.dsLock.RUnlock()
		
		if !exists {
			logger.Debug("Dataspace '%s' not found", dataspaceName)
			return []*models.Entity{}, nil
		}
		
		entityIDs, err := dsIndex.QueryByTags(filteredTags, matchAll)
		if err != nil {
			return nil, err
		}
		
		logger.Debug("Found %d entities in dataspace '%s'", len(entityIDs), dataspaceName)
		
		// Fetch entities using embedded repository's GetByID method
		entities := make([]*models.Entity, 0, len(entityIDs))
		for _, id := range entityIDs {
			// Use the proper GetByID method instead of direct map access
			entity, err := r.EntityRepository.GetByID(id)
			if err != nil {
				logger.Trace("Entity %s not found in dataspace '%s': %v", id, dataspaceName, err)
				continue
			}
			if entity != nil {
				entities = append(entities, entity)
			}
		}
		
		logger.Debug("Retrieved %d entities from dataspace '%s'", len(entities), dataspaceName)
		return entities, nil
	}
	
	// Special handling for system queries
	if len(tags) == 1 {
		tag := tags[0]
		// Handle temporal tags
		actualTag := tag
		if parts := strings.SplitN(tag, "|", 2); len(parts) == 2 {
			actualTag = parts[1]
		}
		
		// System entities queries (users, permissions, etc.) should query _system dataspace
		if strings.HasPrefix(actualTag, "identity:username:") || 
		   strings.HasPrefix(actualTag, "type:user") ||
		   strings.HasPrefix(actualTag, "type:permission") ||
		   strings.HasPrefix(actualTag, "type:role") ||
		   strings.HasPrefix(actualTag, "type:group") ||
		   strings.HasPrefix(actualTag, "type:session") ||
		   strings.HasPrefix(actualTag, "token:") {
			logger.Debug("System query detected, using _system dataspace for: %s", actualTag)
			dataspaceName = "_system"
			
			r.dsLock.RLock()
			dsIndex, exists := r.dataspaceIndexes[dataspaceName]
			r.dsLock.RUnlock()
			
			if !exists {
				logger.Debug("System dataspace not found")
				return []*models.Entity{}, nil
			}
			
			entityIDs, err := dsIndex.QueryByTags([]string{actualTag}, matchAll)
			if err != nil {
				return nil, err
			}
			
			// Fetch entities
			entities := make([]*models.Entity, 0, len(entityIDs))
			for _, id := range entityIDs {
				entity, err := r.EntityRepository.GetByID(id)
				if err != nil {
					continue
				}
				if entity != nil {
					entities = append(entities, entity)
				}
			}
			
			return entities, nil
		}
	}
	
	// No fallback to global search - enforce dataspace isolation
	logger.Trace("No dataspace tag found in query, returning empty result for isolation")
	return []*models.Entity{}, nil
}

// extractDataspace determines which dataspace an entity belongs to
func (r *DataspaceRepository) extractDataspace(entity *models.Entity) string {
	for _, tag := range entity.Tags {
		// Handle temporal tags (TIMESTAMP|tag format)
		actualTag := tag
		if parts := strings.SplitN(tag, "|", 2); len(parts) == 2 {
			actualTag = parts[1]
		}
		
		if strings.HasPrefix(actualTag, "dataspace:") {
			return strings.TrimPrefix(actualTag, "dataspace:")
		}
	}
	return "default"
}

// loadDataspaceIndexes loads all dataspace indexes from disk
func (r *DataspaceRepository) loadDataspaceIndexes() error {
	entries, err := os.ReadDir(r.indexPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".idx") {
			name := strings.TrimSuffix(entry.Name(), ".idx")
			// Use .ebf extension for consistency with SaveTagIndexV2
			indexPath := filepath.Join(r.indexPath, name+".ebf")
			
			dsIndex := &DataspaceIndexImpl{
				name:      name,
				indexPath: indexPath,
				entities:  make(map[string]bool),
				tagIndex:  make(map[string][]string),
			}
			
			if err := dsIndex.LoadFromFile(indexPath); err != nil {
				logger.Error("Failed to load dataspace index %s: %v", name, err)
				continue
			}
			
			r.dataspaceIndexes[name] = dsIndex
			logger.Info("Loaded dataspace index: %s", name)
		}
	}
	
	return nil
}

// DataspaceIndex Implementation

// AddEntity adds an entity to the dataspace index
func (d *DataspaceIndexImpl) AddEntity(entity *models.Entity) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	
	d.entities[entity.ID] = true
	
	// Update tag index
	for _, tag := range entity.Tags {
		// Handle temporal tags
		actualTag := tag
		if parts := strings.SplitN(tag, "|", 2); len(parts) == 2 {
			actualTag = parts[1]
		}
		
		// Skip dataspace tags
		if strings.HasPrefix(actualTag, "dataspace:") {
			continue
		}
		
		// Use the actual tag (without timestamp) for indexing
		if _, exists := d.tagIndex[actualTag]; !exists {
			d.tagIndex[actualTag] = []string{}
		}
		
		// Add entity ID if not already present
		found := false
		for _, id := range d.tagIndex[actualTag] {
			if id == entity.ID {
				found = true
				break
			}
		}
		if !found {
			d.tagIndex[actualTag] = append(d.tagIndex[actualTag], entity.ID)
		}
	}
	
	d.stats.EntityCount++
	return nil
}

// QueryByTags queries entities by tags within the dataspace
func (d *DataspaceIndexImpl) QueryByTags(tags []string, matchAll bool) ([]string, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	
	if len(tags) == 0 {
		// Return all entities in dataspace
		result := make([]string, 0, len(d.entities))
		for id := range d.entities {
			result = append(result, id)
		}
		return result, nil
	}
	
	resultSet := make(map[string]int)
	
	for _, tag := range tags {
		if entityIDs, exists := d.tagIndex[tag]; exists {
			for _, id := range entityIDs {
				resultSet[id]++
			}
		}
	}
	
	// Filter based on matchAll
	result := []string{}
	requiredCount := len(tags)
	
	for id, count := range resultSet {
		if matchAll && count == requiredCount {
			result = append(result, id)
		} else if !matchAll && count > 0 {
			result = append(result, id)
		}
	}
	
	d.stats.QueryCount++
	return result, nil
}

// SaveToFile persists the dataspace index
func (d *DataspaceIndexImpl) SaveToFile(filepath string) error {
	// For now, use the same format as tag index persistence
	// TODO: Implement custom binary format for dataspace indexes
	return SaveTagIndexV2(filepath, d.tagIndex)
}

// LoadFromFile loads the dataspace index
func (d *DataspaceIndexImpl) LoadFromFile(filepath string) error {
	// For now, use the same format as tag index persistence
	tagIndex, err := LoadTagIndexV2(filepath)
	if err != nil {
		return err
	}
	
	d.tagIndex = tagIndex
	
	// Rebuild entity set from tag index
	d.entities = make(map[string]bool)
	for _, ids := range tagIndex {
		for _, id := range ids {
			d.entities[id] = true
		}
	}
	
	d.stats.EntityCount = int64(len(d.entities))
	return nil
}

// rebuildDataspaceIndexes rebuilds dataspace indexes from existing entities
func (r *DataspaceRepository) rebuildDataspaceIndexes() error {
	logger.Info("Rebuilding dataspace indexes from existing entities...")
	
	// Get all entities from the base repository
	allEntities, err := r.EntityRepository.List()
	if err != nil {
		return fmt.Errorf("failed to list entities: %w", err)
	}
	
	logger.Info("Found %d entities to index", len(allEntities))
	
	// Index each entity into its dataspace
	for _, entity := range allEntities {
		dataspaceName := r.extractDataspace(entity)
		
		// Ensure dataspace index exists
		r.dsLock.RLock()
		dsIndex, exists := r.dataspaceIndexes[dataspaceName]
		r.dsLock.RUnlock()
		
		if !exists {
			// Create dataspace on demand
			logger.Info("Creating dataspace '%s' during rebuild", dataspaceName)
			newDataspace := &models.Dataspace{
				Name: dataspaceName,
				Config: models.DataspaceConfig{
					IndexStrategy: models.IndexStrategyBTree,
					OptimizeFor:   models.OptimizeForReads,
				},
			}
			if err := r.CreateDataspace(newDataspace); err != nil {
				logger.Error("Failed to create dataspace during rebuild: %v", err)
				continue
			}
			
			// Get the newly created index
			r.dsLock.RLock()
			dsIndex, exists = r.dataspaceIndexes[dataspaceName]
			r.dsLock.RUnlock()
		}
		
		if exists {
			if err := dsIndex.AddEntity(entity); err != nil {
				logger.Error("Failed to add entity %s to dataspace %s: %v", entity.ID, dataspaceName, err)
			}
		}
	}
	
	// Save all dataspace indexes
	r.dsLock.RLock()
	defer r.dsLock.RUnlock()
	
	for name, dsIndex := range r.dataspaceIndexes {
		if err := dsIndex.SaveToFile(dsIndex.indexPath); err != nil {
			logger.Error("Failed to save dataspace index %s: %v", name, err)
		} else {
			logger.Info("Saved dataspace index %s with %d entities", name, len(dsIndex.entities))
		}
	}
	
	return nil
}