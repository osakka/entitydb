package main

import (
	"entitydb/storage/binary"
	"flag"
	"fmt"
	"os"
)

func main() {
	var (
		dataPath     = flag.String("data", "./var", "Path to EntityDB data directory")
		operation    = flag.String("op", "check", "Operation: check, repair-index, repair-wal, validate-checksums")
		entityID     = flag.String("entity", "", "Entity ID for specific operations")
	)
	flag.Parse()
	
	// Initialize logger (logger is already initialized)
	
	// Create recovery manager
	recovery := binary.NewRecoveryManager(*dataPath)
	
	switch *operation {
	case "check":
		checkIntegrity(*dataPath)
		
	case "repair-index":
		repairIndex(*dataPath)
		
	case "repair-wal":
		if err := recovery.RepairWAL(); err != nil {
			fmt.Fprintf(os.Stderr, "WAL repair failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("WAL repair completed successfully")
		
	case "validate-checksums":
		validateChecksums(*dataPath)
		
	case "recover-entity":
		if *entityID == "" {
			fmt.Fprintf(os.Stderr, "Entity ID required for recovery\n")
			os.Exit(1)
		}
		recoverEntity(*dataPath, *entityID)
		
	default:
		fmt.Fprintf(os.Stderr, "Unknown operation: %s\n", *operation)
		flag.Usage()
		os.Exit(1)
	}
}

func checkIntegrity(dataPath string) {
	fmt.Println("Checking database integrity...")
	
	repo, err := binary.NewEntityRepository(dataPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open repository: %v\n", err)
		os.Exit(1)
	}
	defer repo.Close()
	
	// Check index health
	if err := repo.VerifyIndexHealth(); err != nil {
		fmt.Printf("Index health check failed: %v\n", err)
		fmt.Println("Run with -op=repair-index to fix")
	} else {
		fmt.Println("Index health check passed")
	}
	
	// Check for orphaned entries
	orphaned := repo.FindOrphanedEntries()
	if len(orphaned) > 0 {
		fmt.Printf("Found %d orphaned entries\n", len(orphaned))
		for _, id := range orphaned {
			fmt.Printf("  - %s\n", id)
		}
	} else {
		fmt.Println("No orphaned entries found")
	}
}

func repairIndex(dataPath string) {
	fmt.Println("Repairing index...")
	
	repo, err := binary.NewEntityRepository(dataPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open repository: %v\n", err)
		os.Exit(1)
	}
	defer repo.Close()
	
	if err := repo.RepairIndex(); err != nil {
		fmt.Fprintf(os.Stderr, "Index repair failed: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Println("Index repair completed successfully")
}

func validateChecksums(dataPath string) {
	fmt.Println("Validating entity checksums...")
	
	repo, err := binary.NewEntityRepository(dataPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open repository: %v\n", err)
		os.Exit(1)
	}
	defer repo.Close()
	
	entities, err := repo.List()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to list entities: %v\n", err)
		os.Exit(1)
	}
	
	valid := 0
	invalid := 0
	missing := 0
	
	for _, entity := range entities {
		isValid, expectedChecksum := repo.ValidateEntityChecksum(entity)
		if expectedChecksum == "" {
			missing++
			fmt.Printf("Entity %s: No checksum found\n", entity.ID)
		} else if !isValid {
			invalid++
			fmt.Printf("Entity %s: Checksum mismatch!\n", entity.ID)
		} else {
			valid++
		}
	}
	
	fmt.Printf("\nChecksum validation summary:\n")
	fmt.Printf("  Valid:   %d\n", valid)
	fmt.Printf("  Invalid: %d\n", invalid)
	fmt.Printf("  Missing: %d\n", missing)
	fmt.Printf("  Total:   %d\n", len(entities))
}

func recoverEntity(dataPath string, entityID string) {
	fmt.Printf("Attempting to recover entity %s...\n", entityID)
	
	repo, err := binary.NewEntityRepository(dataPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open repository: %v\n", err)
		os.Exit(1)
	}
	defer repo.Close()
	
	// First try normal read
	entity, err := repo.GetByID(entityID)
	if err == nil {
		fmt.Printf("Entity %s is readable, no recovery needed\n", entityID)
		return
	}
	
	fmt.Printf("Normal read failed: %v\n", err)
	fmt.Println("Recovery process should have been triggered automatically")
	
	// Try reading again after recovery attempt
	entity, err = repo.GetByID(entityID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Recovery failed: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("Entity recovered successfully\n")
	fmt.Printf("  ID: %s\n", entity.ID)
	fmt.Printf("  Tags: %d\n", len(entity.Tags))
	fmt.Printf("  Content size: %d bytes\n", len(entity.Content))
}