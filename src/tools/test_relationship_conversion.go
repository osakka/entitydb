package main

import (
	"entitydb/logger"
	"entitydb/storage/binary"
	"os"
)

func main() {
	logger.Info("=== Test Relationship Conversion ===")
	
	dataDir := "var/"
	if len(os.Args) > 1 {
		dataDir = os.Args[1]
	}
	
	repo, err := binary.NewEntityRepository(dataDir)
	if err != nil {
		logger.Error("Failed to create repository: %v", err)
		os.Exit(1)
	}
	
	// Test a specific relationship ID we know exists
	relID := "rel_9de03ca57e584fd7a658663cff67f297"
	logger.Info("Testing GetRelationshipByID for: %s", relID)
	
	rel, err := repo.GetRelationshipByID(relID)
	if err != nil {
		logger.Error("GetRelationshipByID failed: %v", err)
	} else {
		logger.Info("GetRelationshipByID succeeded: %v", rel)
	}
}