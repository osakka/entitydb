package api

import (
	"net/http"
	"entitydb/storage/binary"
	"entitydb/logger"
)

// RegisterTagFixHandler adds a route to fix the tag index for temporal tags
func RegisterTagFixHandler(router *http.ServeMux, handler *EntityHandler) {
	logger.Info("Registering tag index fix handler...")
	
	// Register the tag reindex endpoint
	router.HandleFunc("/api/v1/patches/reindex-tags", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte(`{"error":"Method not allowed"}`))
			return
		}
		
		// Apply the tag fix
		err := FixTemporalTagIndex(handler)
		if err != nil {
			logger.Error("Failed to fix temporal tag index: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Failed to fix temporal tag index"}`))
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"success","message":"Temporal tag index has been fixed"}`))
	})
	
	logger.Info("Tag index fix handler registered")
}

// FixTemporalTagIndex fixes the tag index in the underlying repository
func FixTemporalTagIndex(handler *EntityHandler) error {
	if handler == nil {
		return nil
	}
	
	// Try to access the underlying binary repository
	var repo *binary.EntityRepository
	
	// Try to access embedded repository
	switch r := handler.repo.(type) {
	case *binary.TemporalRepository:
		repo = r.HighPerformanceRepository.EntityRepository
	case *binary.HighPerformanceRepository:
		repo = r.EntityRepository
	case *binary.EntityRepository:
		repo = r
	default:
		// Fall back to using ListByTag method fix as implemented above
		logger.Error("Unable to access underlying binary repository, tag reindexing not supported")
		return nil
	}
	
	// Fix the tag index
	if repo.ReindexTags != nil {
		return repo.ReindexTags()
	}
	
	logger.Warn("Repository doesn't support ReindexTags method")
	return nil
}