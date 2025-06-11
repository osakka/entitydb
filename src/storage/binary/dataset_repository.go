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

// DatasetRepository implements per-dataset index isolation
type DatasetRepository struct {
	*EntityRepository // Embed base repository for entity storage
	
	// Dataset-specific indexes
	datasetIndexes map[string]*DatasetIndexImpl
	datasetConfigs map[string]*models.Dataset
	dsLock          sync.RWMutex
	
	// Dataset index directory
	indexPath string
}

// DatasetIndexImpl implements the DatasetIndex interface
type DatasetIndexImpl struct {
	name      string
	indexPath string
	
	// In-memory indexes
	entities  map[string]bool           // Entity IDs in this dataset
	tagIndex  map[string][]string       // tag -> entity IDs
	
	// Statistics
	stats     models.DatasetStats
	
	// Synchronization
	mu        sync.RWMutex
}

// NewDatasetRepository creates a repository with dataset isolation
func NewDatasetRepository(dataPath string) (*DatasetRepository, error) {
	baseRepo, err := NewEntityRepository(dataPath)
	if err != nil {
		return nil, err
	}
	
	indexPath := filepath.Join(dataPath, "datasets")
	if err := os.MkdirAll(indexPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create dataset index directory: %w", err)
	}
	
	repo := &DatasetRepository{
		EntityRepository:  baseRepo,
		datasetIndexes: make(map[string]*DatasetIndexImpl),
		datasetConfigs: make(map[string]*models.Dataset),
		indexPath:        indexPath,
	}
	
	// Load existing dataset indexes
	if err := repo.loadDatasetIndexes(); err != nil {
		logger.Error("Failed to load dataset indexes: %v", err)
	}
	
	// Rebuild dataset indexes from existing entities
	if err := repo.rebuildDatasetIndexes(); err != nil {
		logger.Error("Failed to rebuild dataset indexes: %v", err)
	}
	
	// Create default dataset if none exists
	if _, exists := repo.datasetIndexes["default"]; !exists {
		defaultDs := &models.Dataset{
			Name: "default",
			Config: models.DatasetConfig{
				IndexStrategy: models.IndexStrategyBTree,
				OptimizeFor:   models.OptimizeForReads,
			},
		}
		if err := repo.CreateDataset(defaultDs); err != nil {
			logger.Error("Failed to create default dataset: %v", err)
		}
	}
	
	return repo, nil
}

// CreateDataset creates a new dataset with its own index
func (r *DatasetRepository) CreateDataset(dataset *models.Dataset) error {
	r.dsLock.Lock()
	defer r.dsLock.Unlock()
	
	if _, exists := r.datasetIndexes[dataset.Name]; exists {
		return fmt.Errorf("dataset %s already exists", dataset.Name)
	}
	
	// Create dataset index (use .ebf extension so SaveTagIndexV2 creates .idx)
	dsIndex := &DatasetIndexImpl{
		name:      dataset.Name,
		indexPath: filepath.Join(r.indexPath, dataset.Name+".ebf"),
		entities:  make(map[string]bool),
		tagIndex:  make(map[string][]string),
	}
	
	r.datasetIndexes[dataset.Name] = dsIndex
	r.datasetConfigs[dataset.Name] = dataset
	
	// Save index to disk
	if err := dsIndex.SaveToFile(dsIndex.indexPath); err != nil {
		return fmt.Errorf("failed to save dataset index: %w", err)
	}
	
	logger.Info("Created dataset: %s", dataset.Name)
	return nil
}

// Create entity in a specific dataset
func (r *DatasetRepository) Create(entity *models.Entity) error {
	// Extract dataset from tags
	datasetName := r.extractDataset(entity)
	logger.Debug("Creating entity %s in dataset: %s", entity.ID, datasetName)
	
	// Create in base repository
	if err := r.EntityRepository.Create(entity); err != nil {
		return err
	}
	
	// Add to dataset index
	r.dsLock.RLock()
	dsIndex, exists := r.datasetIndexes[datasetName]
	r.dsLock.RUnlock()
	
	if !exists && datasetName != "default" {
		// Create dataset if it doesn't exist
		logger.Info("Creating dataset '%s' on demand", datasetName)
		newDataset := &models.Dataset{
			Name: datasetName,
			Config: models.DatasetConfig{
				IndexStrategy: models.IndexStrategyBTree,
				OptimizeFor:   models.OptimizeForReads,
			},
		}
		if err := r.CreateDataset(newDataset); err != nil {
			logger.Error("Failed to create dataset: %v", err)
		} else {
			// Get the newly created index
			r.dsLock.RLock()
			dsIndex, exists = r.datasetIndexes[datasetName]
			r.dsLock.RUnlock()
		}
	}
	
	if exists {
		if err := dsIndex.AddEntity(entity); err != nil {
			logger.Error("Failed to add entity to dataset index: %v", err)
		} else {
			// Save index after adding entity
			if err := dsIndex.SaveToFile(dsIndex.indexPath); err != nil {
				logger.Error("Failed to save dataset index: %v", err)
			}
		}
		logger.Debug("Added entity %s to dataset '%s' index", entity.ID, datasetName)
	}
	
	return nil
}

// ListByTags with dataset awareness
func (r *DatasetRepository) ListByTags(tags []string, matchAll bool) ([]*models.Entity, error) {
	// Check if query is dataset-specific
	datasetName := ""
	filteredTags := []string{}
	
	for _, tag := range tags {
		// Handle temporal tags
		actualTag := tag
		if parts := strings.SplitN(tag, "|", 2); len(parts) == 2 {
			actualTag = parts[1]
		}
		
		if strings.HasPrefix(actualTag, "dataset:") || strings.HasPrefix(actualTag, "dataset:") {
			datasetName = strings.TrimPrefix(strings.TrimPrefix(actualTag, "dataset:"), "dataset:")
		} else {
			filteredTags = append(filteredTags, actualTag)
		}
	}
	
	// If dataset-specific, use dataset index
	if datasetName != "" {
		logger.Debug("Dataset query for '%s' with tags: %v", datasetName, filteredTags)
		
		r.dsLock.RLock()
		dsIndex, exists := r.datasetIndexes[datasetName]
		r.dsLock.RUnlock()
		
		if !exists {
			logger.Debug("Dataset '%s' not found", datasetName)
			return []*models.Entity{}, nil
		}
		
		entityIDs, err := dsIndex.QueryByTags(filteredTags, matchAll)
		if err != nil {
			return nil, err
		}
		
		logger.Debug("Found %d entities in dataset '%s'", len(entityIDs), datasetName)
		
		// Fetch entities using embedded repository's GetByID method
		entities := make([]*models.Entity, 0, len(entityIDs))
		for _, id := range entityIDs {
			// Use the proper GetByID method instead of direct map access
			entity, err := r.EntityRepository.GetByID(id)
			if err != nil {
				logger.Trace("Entity %s not found in dataset '%s': %v", id, datasetName, err)
				continue
			}
			if entity != nil {
				entities = append(entities, entity)
			}
		}
		
		logger.Debug("Retrieved %d entities from dataset '%s'", len(entities), datasetName)
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
		
		// System entities queries (users, permissions, etc.) should query _system dataset
		if strings.HasPrefix(actualTag, "identity:username:") || 
		   strings.HasPrefix(actualTag, "type:user") ||
		   strings.HasPrefix(actualTag, "type:permission") ||
		   strings.HasPrefix(actualTag, "type:role") ||
		   strings.HasPrefix(actualTag, "type:group") ||
		   strings.HasPrefix(actualTag, "type:session") ||
		   strings.HasPrefix(actualTag, "token:") {
			logger.Debug("System query detected, using _system dataset for: %s", actualTag)
			datasetName = "_system"
			
			r.dsLock.RLock()
			dsIndex, exists := r.datasetIndexes[datasetName]
			r.dsLock.RUnlock()
			
			if !exists {
				logger.Debug("System dataset not found")
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
	
	// No fallback to global search - enforce dataset isolation
	logger.Trace("No dataset tag found in query, returning empty result for isolation")
	return []*models.Entity{}, nil
}

// extractDataset determines which dataset an entity belongs to
func (r *DatasetRepository) extractDataset(entity *models.Entity) string {
	for _, tag := range entity.Tags {
		// Handle temporal tags (TIMESTAMP|tag format)
		actualTag := tag
		if parts := strings.SplitN(tag, "|", 2); len(parts) == 2 {
			actualTag = parts[1]
		}
		
		if strings.HasPrefix(actualTag, "dataset:") {
			return strings.TrimPrefix(actualTag, "dataset:")
		}
	}
	return "default"
}

// loadDatasetIndexes loads all dataset indexes from disk
func (r *DatasetRepository) loadDatasetIndexes() error {
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
			
			dsIndex := &DatasetIndexImpl{
				name:      name,
				indexPath: indexPath,
				entities:  make(map[string]bool),
				tagIndex:  make(map[string][]string),
			}
			
			if err := dsIndex.LoadFromFile(indexPath); err != nil {
				logger.Error("Failed to load dataset index %s: %v", name, err)
				continue
			}
			
			r.datasetIndexes[name] = dsIndex
			logger.Info("Loaded dataset index: %s", name)
		}
	}
	
	return nil
}

// DatasetIndex Implementation

// AddEntity adds an entity to the dataset index
func (d *DatasetIndexImpl) AddEntity(entity *models.Entity) error {
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
		
		// Skip dataset tags
		if strings.HasPrefix(actualTag, "dataset:") {
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

// QueryByTags queries entities by tags within the dataset
func (d *DatasetIndexImpl) QueryByTags(tags []string, matchAll bool) ([]string, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	
	if len(tags) == 0 {
		// Return all entities in dataset
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

// SaveToFile persists the dataset index
func (d *DatasetIndexImpl) SaveToFile(filepath string) error {
	// For now, use the same format as tag index persistence
	// TODO: Implement custom binary format for dataset indexes
	return SaveTagIndexV2(filepath, d.tagIndex)
}

// LoadFromFile loads the dataset index
func (d *DatasetIndexImpl) LoadFromFile(filepath string) error {
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

// rebuildDatasetIndexes rebuilds dataset indexes from existing entities
func (r *DatasetRepository) rebuildDatasetIndexes() error {
	logger.Info("Rebuilding dataset indexes from existing entities...")
	
	// Get all entities from the base repository
	allEntities, err := r.EntityRepository.List()
	if err != nil {
		return fmt.Errorf("failed to list entities: %w", err)
	}
	
	logger.Info("Found %d entities to index", len(allEntities))
	
	// Index each entity into its dataset
	for _, entity := range allEntities {
		datasetName := r.extractDataset(entity)
		
		// Ensure dataset index exists
		r.dsLock.RLock()
		dsIndex, exists := r.datasetIndexes[datasetName]
		r.dsLock.RUnlock()
		
		if !exists {
			// Create dataset on demand
			logger.Info("Creating dataset '%s' during rebuild", datasetName)
			newDataset := &models.Dataset{
				Name: datasetName,
				Config: models.DatasetConfig{
					IndexStrategy: models.IndexStrategyBTree,
					OptimizeFor:   models.OptimizeForReads,
				},
			}
			if err := r.CreateDataset(newDataset); err != nil {
				logger.Error("Failed to create dataset during rebuild: %v", err)
				continue
			}
			
			// Get the newly created index
			r.dsLock.RLock()
			dsIndex, exists = r.datasetIndexes[datasetName]
			r.dsLock.RUnlock()
		}
		
		if exists {
			if err := dsIndex.AddEntity(entity); err != nil {
				logger.Error("Failed to add entity %s to dataset %s: %v", entity.ID, datasetName, err)
			}
		}
	}
	
	// Save all dataset indexes
	r.dsLock.RLock()
	defer r.dsLock.RUnlock()
	
	for name, dsIndex := range r.datasetIndexes {
		if err := dsIndex.SaveToFile(dsIndex.indexPath); err != nil {
			logger.Error("Failed to save dataset index %s: %v", name, err)
		} else {
			logger.Info("Saved dataset index %s with %d entities", name, len(dsIndex.entities))
		}
	}
	
	return nil
}