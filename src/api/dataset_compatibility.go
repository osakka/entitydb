package api

import (
	"net/http"
	"strings"
)

// DatasetCompatibilityMiddleware provides backward compatibility for hub endpoints
func DatasetCompatibilityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Convert hub URLs to dataset URLs
		if strings.Contains(r.URL.Path, "/datasets/") {
			newPath := strings.Replace(r.URL.Path, "/datasets/", "/datasets/", 1)
			r.URL.Path = newPath
			
			// Add deprecation header
			w.Header().Set("X-Deprecation-Warning", "Hub endpoints are deprecated. Use /datasets/ instead.")
		}
		
		// Convert hub query parameters to dataset
		if hub := r.URL.Query().Get("hub"); hub != "" {
			q := r.URL.Query()
			q.Del("hub")
			q.Set("dataset", hub)
			r.URL.RawQuery = q.Encode()
		}
		
		// Convert hub: tags to dataset: tags in the context
		// This would be handled at the repository level
		
		next.ServeHTTP(w, r)
	})
}

// ConvertHubTagsToDataset converts hub: prefixed tags to dataset: prefixed tags
func ConvertHubTagsToDataset(tags []string) []string {
	converted := make([]string, len(tags))
	for i, tag := range tags {
		if strings.HasPrefix(tag, "dataset:") {
			converted[i] = "dataset:" + strings.TrimPrefix(tag, "dataset:")
		} else {
			converted[i] = tag
		}
	}
	return converted
}

// ConvertDatasetTagsToHub converts dataset: tags back to hub: for backward compatibility
func ConvertDatasetTagsToHub(tags []string) []string {
	converted := make([]string, len(tags))
	for i, tag := range tags {
		if strings.HasPrefix(tag, "dataset:") {
			converted[i] = "dataset:" + strings.TrimPrefix(tag, "dataset:")
		} else {
			converted[i] = tag
		}
	}
	return converted
}

// IsHubRelatedTag checks if a tag is hub or dataset related
func IsHubRelatedTag(tag string) bool {
	return strings.HasPrefix(tag, "dataset:") || strings.HasPrefix(tag, "dataset:")
}

// ExtractDatasetName extracts the dataset name from hub: or dataset: tags
func ExtractDatasetName(tags []string) string {
	for _, tag := range tags {
		if strings.HasPrefix(tag, "dataset:") {
			return strings.TrimPrefix(tag, "dataset:")
		}
		if strings.HasPrefix(tag, "dataset:") {
			return strings.TrimPrefix(tag, "dataset:")
		}
	}
	return ""
}