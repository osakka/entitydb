package models

import (
	"time"
)

// Dataset represents an isolated data universe with its own indexes and rules
type Dataset struct {
	Name        string    `json:"name"`
	Created     time.Time `json:"created"`
	Description string    `json:"description"`
	Config      DatasetConfig `json:"config"`
}

// DatasetConfig defines the behavior and optimization strategy for a dataset
type DatasetConfig struct {
	// Index strategy for this dataset
	IndexStrategy IndexStrategyType `json:"index_strategy"`
	
	// Whether to keep indexes in memory
	InMemoryIndexes bool `json:"in_memory_indexes"`
	
	// Custom index fields for this dataset
	CustomIndexes []string `json:"custom_indexes,omitempty"`
	
	// Retention policy (0 = keep forever)
	RetentionDays int `json:"retention_days"`
	
	// Performance hints
	OptimizeFor OptimizationType `json:"optimize_for"`
}

// IndexStrategyType defines how data is indexed in a dataset
type IndexStrategyType string

const (
	IndexStrategyBTree      IndexStrategyType = "btree"      // Balanced tree for general purpose
	IndexStrategyHash       IndexStrategyType = "hash"       // Hash table for exact matches
	IndexStrategyTimeSeries IndexStrategyType = "timeseries" // Optimized for time-based data
	IndexStrategyGraph      IndexStrategyType = "graph"      // Optimized for relationships
)

// OptimizationType hints at the primary use case for optimization
type OptimizationType string

const (
	OptimizeForWrites      OptimizationType = "writes"       // Optimize for write throughput
	OptimizeForReads       OptimizationType = "reads"        // Optimize for query performance
	OptimizeForSpace       OptimizationType = "space"        // Optimize for storage efficiency
	OptimizeForConcurrency OptimizationType = "concurrency"  // Optimize for concurrent access
)

// DatasetStats provides metrics about a dataset
type DatasetStats struct {
	EntityCount   int64     `json:"entity_count"`
	IndexSize     int64     `json:"index_size_bytes"`
	DataSize      int64     `json:"data_size_bytes"`
	LastModified  time.Time `json:"last_modified"`
	QueryCount    int64     `json:"query_count"`
	WriteCount    int64     `json:"write_count"`
}

// DatasetManager defines the interface for managing datasets
type DatasetManager interface {
	// Dataset lifecycle
	CreateDataset(dataset *Dataset) error
	GetDataset(name string) (*Dataset, error)
	ListDatasets() ([]*Dataset, error)
	DeleteDataset(name string) error
	
	// Dataset statistics
	GetDatasetStats(name string) (*DatasetStats, error)
	
	// Check if entity belongs to a dataset
	GetEntityDataset(entityID string) (string, error)
}

// DatasetIndex defines the interface for dataset-specific indexes
type DatasetIndex interface {
	// Index operations
	AddEntity(entity *Entity) error
	RemoveEntity(entityID string) error
	UpdateEntity(entity *Entity) error
	
	// Query operations
	QueryByTag(tag string) ([]string, error)
	QueryByTags(tags []string, matchAll bool) ([]string, error)
	
	// Persistence
	SaveToFile(filepath string) error
	LoadFromFile(filepath string) error
	
	// Maintenance
	Optimize() error
	GetStats() *DatasetStats
}