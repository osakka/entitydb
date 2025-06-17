package binary

import (
	"encoding/json"
	"entitydb/logger"
	"entitydb/models"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// MetricsBackend provides completely isolated backend for metrics persistence
// This operates independently of the main repository to prevent any deadlocks
type MetricsBackend struct {
	dataPath    string
	initialized bool
	mu          sync.Mutex
	storage     *IsolatedMetricsStorage
}

// NewMetricsBackend creates an isolated metrics backend
func NewMetricsBackend(dataPath string) *MetricsBackend {
	return &MetricsBackend{
		dataPath: dataPath,
	}
}

// IsolatedMetricsStorage provides completely isolated metrics storage
// This operates independently of the main EntityDB repository to prevent deadlocks
type IsolatedMetricsStorage struct {
	dataPath        string
	entities        map[string]*models.Entity
	tagIndex        map[string][]string // tag -> entity IDs
	mu              sync.RWMutex
	metricsFilePath string
	initialized     bool
}

// NewIsolatedMetricsStorage creates a new isolated metrics storage
func NewIsolatedMetricsStorage(dataPath string) *IsolatedMetricsStorage {
	storage := &IsolatedMetricsStorage{
		dataPath:        dataPath,
		entities:        make(map[string]*models.Entity),
		tagIndex:        make(map[string][]string),
		metricsFilePath: filepath.Join(dataPath, "metrics.json"),
	}
	
	// Load existing metrics from file
	storage.load()
	
	return storage
}

// CreateMetricsEntity creates a metrics entity in isolated storage
func (mb *MetricsBackend) CreateMetricsEntity(entity *models.Entity) error {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	
	if !mb.initialized {
		mb.storage = NewIsolatedMetricsStorage(mb.dataPath)
		mb.initialized = true
	}
	
	return mb.storage.CreateEntity(entity)
}

// GetMetricsEntity gets a metrics entity from isolated storage
func (mb *MetricsBackend) GetMetricsEntity(id string) (*models.Entity, error) {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	
	if !mb.initialized {
		mb.storage = NewIsolatedMetricsStorage(mb.dataPath)
		mb.initialized = true
	}
	
	return mb.storage.GetEntity(id)
}

// ListMetricsEntitiesByTag lists metrics entities by tag from isolated storage
func (mb *MetricsBackend) ListMetricsEntitiesByTag(tag string) ([]*models.Entity, error) {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	
	if !mb.initialized {
		mb.storage = NewIsolatedMetricsStorage(mb.dataPath)
		mb.initialized = true
	}
	
	return mb.storage.ListByTag(tag)
}

// AddTagToMetricsEntity adds a tag to a metrics entity in isolated storage
func (mb *MetricsBackend) AddTagToMetricsEntity(entityID, tag string) error {
	mb.mu.Lock()
	defer mb.mu.Unlock()
	
	if !mb.initialized {
		mb.storage = NewIsolatedMetricsStorage(mb.dataPath)
		mb.initialized = true
	}
	
	return mb.storage.AddTag(entityID, tag)
}

// CreateEntity creates an entity in isolated storage
func (ims *IsolatedMetricsStorage) CreateEntity(entity *models.Entity) error {
	ims.mu.Lock()
	defer ims.mu.Unlock()
	
	// Check if entity already exists
	if _, exists := ims.entities[entity.ID]; exists {
		return fmt.Errorf("metrics entity with ID %s already exists", entity.ID)
	}
	
	// Store entity
	ims.entities[entity.ID] = entity
	
	// Update tag index
	for _, tag := range entity.Tags {
		ims.addToTagIndex(entity.ID, tag)
	}
	
	// Persist to file (async to avoid blocking)
	go ims.persist()
	
	logger.Trace("IsolatedMetricsStorage: created entity %s", entity.ID)
	return nil
}

// GetEntity gets an entity from isolated storage
func (ims *IsolatedMetricsStorage) GetEntity(id string) (*models.Entity, error) {
	ims.mu.RLock()
	defer ims.mu.RUnlock()
	
	entity, exists := ims.entities[id]
	if !exists {
		return nil, fmt.Errorf("metrics entity not found: %s", id)
	}
	
	return entity, nil
}

// ListByTag lists entities by tag from isolated storage
func (ims *IsolatedMetricsStorage) ListByTag(tag string) ([]*models.Entity, error) {
	ims.mu.RLock()
	defer ims.mu.RUnlock()
	
	var results []*models.Entity
	
	// Look for exact tag match and partial matches
	for indexTag, entityIDs := range ims.tagIndex {
		if ims.tagMatches(indexTag, tag) {
			for _, entityID := range entityIDs {
				if entity, exists := ims.entities[entityID]; exists {
					results = append(results, entity)
				}
			}
		}
	}
	
	logger.Trace("IsolatedMetricsStorage: found %d entities for tag %s", len(results), tag)
	return results, nil
}

// AddTag adds a tag to an entity in isolated storage
func (ims *IsolatedMetricsStorage) AddTag(entityID, tag string) error {
	ims.mu.Lock()
	defer ims.mu.Unlock()
	
	entity, exists := ims.entities[entityID]
	if !exists {
		return fmt.Errorf("metrics entity not found: %s", entityID)
	}
	
	// Create temporal tag with current timestamp
	timestampedTag := fmt.Sprintf("%d|%s", models.Now(), tag)
	
	// Add tag to entity
	entity.Tags = append(entity.Tags, timestampedTag)
	
	// Update tag index
	ims.addToTagIndex(entityID, timestampedTag)
	
	// Persist to file (async to avoid blocking)
	go ims.persist()
	
	logger.Trace("IsolatedMetricsStorage: added tag %s to entity %s", tag, entityID)
	return nil
}

// addToTagIndex adds an entity ID to the tag index
func (ims *IsolatedMetricsStorage) addToTagIndex(entityID, tag string) {
	// Index both the full temporal tag and the clean tag
	cleanTag := tag
	if pipePos := strings.Index(tag, "|"); pipePos != -1 {
		cleanTag = tag[pipePos+1:]
	}
	
	// Add to index for both versions
	for _, indexTag := range []string{tag, cleanTag} {
		if !ims.containsString(ims.tagIndex[indexTag], entityID) {
			ims.tagIndex[indexTag] = append(ims.tagIndex[indexTag], entityID)
		}
	}
}

// tagMatches checks if a tag matches the search criteria
func (ims *IsolatedMetricsStorage) tagMatches(indexTag, searchTag string) bool {
	// Clean both tags for comparison
	cleanIndexTag := indexTag
	if pipePos := strings.Index(indexTag, "|"); pipePos != -1 {
		cleanIndexTag = indexTag[pipePos+1:]
	}
	
	cleanSearchTag := searchTag
	if pipePos := strings.Index(searchTag, "|"); pipePos != -1 {
		cleanSearchTag = searchTag[pipePos+1:]
	}
	
	return cleanIndexTag == cleanSearchTag || indexTag == searchTag
}

// containsString checks if a slice contains a string
func (ims *IsolatedMetricsStorage) containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// MetricsStorageData represents the JSON structure for persistence
type MetricsStorageData struct {
	Entities map[string]*models.Entity `json:"entities"`
	TagIndex map[string][]string       `json:"tag_index"`
	Version  string                    `json:"version"`
	Updated  time.Time                 `json:"updated"`
}

// persist saves the metrics storage to file (called asynchronously)
func (ims *IsolatedMetricsStorage) persist() {
	ims.mu.RLock()
	data := &MetricsStorageData{
		Entities: ims.entities,
		TagIndex: ims.tagIndex,
		Version:  "1.0",
		Updated:  time.Now(),
	}
	ims.mu.RUnlock()
	
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(ims.metricsFilePath), 0755); err != nil {
		logger.Warn("Failed to create metrics directory: %v", err)
		return
	}
	
	// Write to temp file first, then atomic rename
	tempFile := ims.metricsFilePath + ".tmp"
	
	file, err := os.Create(tempFile)
	if err != nil {
		logger.Warn("Failed to create metrics temp file: %v", err)
		return
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	
	if err := encoder.Encode(data); err != nil {
		logger.Warn("Failed to encode metrics data: %v", err)
		return
	}
	
	if err := file.Sync(); err != nil {
		logger.Warn("Failed to sync metrics file: %v", err)
		return
	}
	
	file.Close()
	
	// Atomic rename
	if err := os.Rename(tempFile, ims.metricsFilePath); err != nil {
		logger.Warn("Failed to rename metrics file: %v", err)
		return
	}
	
	logger.Trace("IsolatedMetricsStorage: persisted %d entities to %s", len(data.Entities), ims.metricsFilePath)
}

// load loads the metrics storage from file
// DECOMMISSIONED: Legacy JSON file metrics loading disabled in favor of single source of truth
func (ims *IsolatedMetricsStorage) load() {
	ims.mu.Lock()
	defer ims.mu.Unlock()
	
	// Initialize empty maps - no more loading from legacy JSON file
	if ims.entities == nil {
		ims.entities = make(map[string]*models.Entity)
	}
	if ims.tagIndex == nil {
		ims.tagIndex = make(map[string][]string)
	}
	
	logger.Debug("IsolatedMetricsStorage: legacy JSON file loading decommissioned, using single source of truth")
}

// GetStats returns statistics about the isolated metrics storage
func (ims *IsolatedMetricsStorage) GetStats() map[string]interface{} {
	ims.mu.RLock()
	defer ims.mu.RUnlock()
	
	return map[string]interface{}{
		"entity_count":   len(ims.entities),
		"tag_index_size": len(ims.tagIndex),
		"storage_path":   ims.metricsFilePath,
	}
}

// Cleanup removes old metrics entities (called periodically)
func (ims *IsolatedMetricsStorage) Cleanup(maxAge time.Duration) {
	ims.mu.Lock()
	defer ims.mu.Unlock()
	
	cutoff := time.Now().Add(-maxAge)
	var toDelete []string
	
	// Find entities to delete based on age
	for entityID, entity := range ims.entities {
		// Check created_at tag
		for _, tag := range entity.Tags {
			cleanTag := tag
			if pipePos := strings.Index(tag, "|"); pipePos != -1 {
				cleanTag = tag[pipePos+1:]
			}
			
			if strings.HasPrefix(cleanTag, "created_at:") {
				createdAtStr := strings.TrimPrefix(cleanTag, "created_at:")
				if createdAt, err := models.ParseStringToNanos(createdAtStr); err == nil {
					if time.Unix(0, createdAt).Before(cutoff) {
						toDelete = append(toDelete, entityID)
						break
					}
				}
			}
		}
	}
	
	// Delete old entities
	for _, entityID := range toDelete {
		if entity, exists := ims.entities[entityID]; exists {
			// Remove from tag index
			for _, tag := range entity.Tags {
				ims.removeFromTagIndex(entityID, tag)
			}
			// Remove entity
			delete(ims.entities, entityID)
		}
	}
	
	if len(toDelete) > 0 {
		logger.Info("IsolatedMetricsStorage: cleaned up %d old metrics entities", len(toDelete))
		go ims.persist()
	}
}

// removeFromTagIndex removes an entity ID from the tag index
func (ims *IsolatedMetricsStorage) removeFromTagIndex(entityID, tag string) {
	cleanTag := tag
	if pipePos := strings.Index(tag, "|"); pipePos != -1 {
		cleanTag = tag[pipePos+1:]
	}
	
	for _, indexTag := range []string{tag, cleanTag} {
		if entityIDs, exists := ims.tagIndex[indexTag]; exists {
			for i, id := range entityIDs {
				if id == entityID {
					ims.tagIndex[indexTag] = append(entityIDs[:i], entityIDs[i+1:]...)
					break
				}
			}
			// Clean up empty tag entries
			if len(ims.tagIndex[indexTag]) == 0 {
				delete(ims.tagIndex, indexTag)
			}
		}
	}
}