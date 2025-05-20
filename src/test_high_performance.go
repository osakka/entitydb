package main

import (
	"entitydb/storage/binary"
	"entitydb/models"
	"fmt"
	"time"
)

func main() {
	// Test that HighPerformanceRepository implements EntityRepository interface
	var repo models.EntityRepository
	
	// Create high-performance repository
	highPerfRepo, err := binary.NewHighPerformanceRepository("./test_data")
	if err != nil {
		fmt.Printf("Error creating high-performance repository: %v\n", err)
		return
	}
	
	// Assign to interface to verify it implements all methods
	repo = highPerfRepo
	
	// Test basic operations
	entity := models.NewEntity()
	entity.AddTag("test:tag")
	
	// Test Create
	created, err := repo.Create(entity)
	if err != nil {
		fmt.Printf("Error creating entity: %v\n", err)
		return
	}
	
	// Test GetByID
	retrieved, err := repo.GetByID(created.ID)
	if err != nil {
		fmt.Printf("Error retrieving entity: %v\n", err)
		return
	}
	
	// Test Query
	query := repo.Query()
	if query == nil {
		fmt.Printf("Query returns nil\n")
		return
	}
	
	fmt.Println("HighPerformanceRepository successfully implements EntityRepository interface!")
}