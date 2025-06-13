//go:build tool
package main

import (
	"fmt"
	"strings"

	"entitydb/storage/binary"
)

func main() {
	fmt.Println("Cleaning up static value tags from metrics")

	// Open repository
	repo, err := binary.NewEntityRepository("/opt/entitydb/var/entities.ebf")
	if err != nil {
		panic(err)
	}
	defer repo.Close()

	// List all metric entities
	metrics, err := repo.ListByTag("type:metric")
	if err != nil {
		panic(fmt.Sprintf("Failed to list metrics: %v", err))
	}

	fmt.Printf("Found %d metric entities\n", len(metrics))

	cleaned := 0
	for _, metric := range metrics {
		hasStaticValue := false
		var newTags []string

		// Check for static value tags
		for _, tag := range metric.Tags {
			// Skip timestamps
			parts := strings.SplitN(tag, "|", 2)
			tagContent := tag
			if len(parts) == 2 {
				tagContent = parts[1]
			}

			if strings.HasPrefix(tagContent, "value:") && len(parts) == 1 {
				// This is a static value tag
				hasStaticValue = true
				fmt.Printf("Removing static value tag '%s' from %s\n", tag, metric.ID)
			} else {
				newTags = append(newTags, tag)
			}
		}

		if hasStaticValue {
			// Update entity with cleaned tags
			metric.Tags = newTags
			if err := repo.Update(metric); err != nil {
				fmt.Printf("Failed to update %s: %v\n", metric.ID, err)
			} else {
				cleaned++
			}
		}
	}

	fmt.Printf("Cleaned %d metrics\n", cleaned)
}