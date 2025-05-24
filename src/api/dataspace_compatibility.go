package api

import (
	"net/http"
	"strings"
)

// DataspaceCompatibilityMiddleware provides backward compatibility for hub endpoints
func DataspaceCompatibilityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Convert hub URLs to dataspace URLs
		if strings.Contains(r.URL.Path, "/hubs/") {
			newPath := strings.Replace(r.URL.Path, "/hubs/", "/dataspaces/", 1)
			r.URL.Path = newPath
			
			// Add deprecation header
			w.Header().Set("X-Deprecation-Warning", "Hub endpoints are deprecated. Use /dataspaces/ instead.")
		}
		
		// Convert hub query parameters to dataspace
		if hub := r.URL.Query().Get("hub"); hub != "" {
			q := r.URL.Query()
			q.Del("hub")
			q.Set("dataspace", hub)
			r.URL.RawQuery = q.Encode()
		}
		
		// Convert hub: tags to dataspace: tags in the context
		// This would be handled at the repository level
		
		next.ServeHTTP(w, r)
	})
}

// ConvertHubTagsToDataspace converts hub: prefixed tags to dataspace: prefixed tags
func ConvertHubTagsToDataspace(tags []string) []string {
	converted := make([]string, len(tags))
	for i, tag := range tags {
		if strings.HasPrefix(tag, "hub:") {
			converted[i] = "dataspace:" + strings.TrimPrefix(tag, "hub:")
		} else {
			converted[i] = tag
		}
	}
	return converted
}

// ConvertDataspaceTagsToHub converts dataspace: tags back to hub: for backward compatibility
func ConvertDataspaceTagsToHub(tags []string) []string {
	converted := make([]string, len(tags))
	for i, tag := range tags {
		if strings.HasPrefix(tag, "dataspace:") {
			converted[i] = "hub:" + strings.TrimPrefix(tag, "dataspace:")
		} else {
			converted[i] = tag
		}
	}
	return converted
}

// IsHubRelatedTag checks if a tag is hub or dataspace related
func IsHubRelatedTag(tag string) bool {
	return strings.HasPrefix(tag, "hub:") || strings.HasPrefix(tag, "dataspace:")
}

// ExtractDataspaceName extracts the dataspace name from hub: or dataspace: tags
func ExtractDataspaceName(tags []string) string {
	for _, tag := range tags {
		if strings.HasPrefix(tag, "dataspace:") {
			return strings.TrimPrefix(tag, "dataspace:")
		}
		if strings.HasPrefix(tag, "hub:") {
			return strings.TrimPrefix(tag, "hub:")
		}
	}
	return ""
}