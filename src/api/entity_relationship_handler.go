package api

import (
	"entitydb/logger"
	"entitydb/models"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// EntityRelationshipHandler provides advanced relationship discovery APIs
type EntityRelationshipHandler struct {
	repo models.EntityRepository
}

// NewEntityRelationshipHandler creates a new relationship handler
func NewEntityRelationshipHandler(repo models.EntityRepository) *EntityRelationshipHandler {
	return &EntityRelationshipHandler{repo: repo}
}

// RelationshipDiscovery represents discovered relationships
type RelationshipDiscovery struct {
	EntityID      string                 `json:"entity_id"`
	DirectRefs    []EntityReference      `json:"direct_references"`
	TagBasedRefs  []TagBasedRelationship `json:"tag_based_relationships"`
	SharedTags    []SharedTagRelation    `json:"shared_tag_relations"`
	NetworkDepth  int                    `json:"network_depth"`
	TotalFound    int                    `json:"total_found"`
	GeneratedAt   time.Time              `json:"generated_at"`
}

// EntityReference represents a direct entity reference
type EntityReference struct {
	SourceID     string    `json:"source_id"`
	SourceName   string    `json:"source_name"`
	SourceType   string    `json:"source_type"`
	TargetID     string    `json:"target_id"`
	TargetName   string    `json:"target_name"`
	TargetType   string    `json:"target_type"`
	RelationType string    `json:"relation_type"`
	ViaTag       string    `json:"via_tag"`
	Confidence   float64   `json:"confidence"`
	CreatedAt    time.Time `json:"created_at"`
}

// TagBasedRelationship represents relationships discovered through tag analysis
type TagBasedRelationship struct {
	RelatedEntityID   string   `json:"related_entity_id"`
	RelatedEntityName string   `json:"related_entity_name"`
	RelatedEntityType string   `json:"related_entity_type"`
	SharedTags        []string `json:"shared_tags"`
	TagSimilarity     float64  `json:"tag_similarity"`
	RelationStrength  float64  `json:"relation_strength"`
}

// SharedTagRelation represents entities sharing specific tags
type SharedTagRelation struct {
	Tag          string   `json:"tag"`
	EntityCount  int      `json:"entity_count"`
	RelatedIDs   []string `json:"related_entity_ids"`
	TagValue     string   `json:"tag_value"`
	TagNamespace string   `json:"tag_namespace"`
}

// NetworkNode represents a node in the entity network
type NetworkNode struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Type         string            `json:"type"`
	Tags         []string          `json:"tags"`
	IsFocus      bool              `json:"is_focus"`
	Distance     int               `json:"distance"`
	Connections  int               `json:"connections"`
	Properties   map[string]string `json:"properties"`
}

// NetworkEdge represents a connection between entities
type NetworkEdge struct {
	Source       string  `json:"source"`
	Target       string  `json:"target"`
	Type         string  `json:"type"`
	Weight       float64 `json:"weight"`
	Bidirectional bool   `json:"bidirectional"`
	Properties   map[string]string `json:"properties"`
}

// EntityNetwork represents the complete network graph
type EntityNetwork struct {
	FocusEntity string        `json:"focus_entity"`
	Nodes       []NetworkNode `json:"nodes"`
	Edges       []NetworkEdge `json:"edges"`
	Depth       int           `json:"depth"`
	TotalNodes  int           `json:"total_nodes"`
	TotalEdges  int           `json:"total_edges"`
	GeneratedAt time.Time     `json:"generated_at"`
}

// DiscoverRelationships performs comprehensive relationship discovery for an entity
func (h *EntityRelationshipHandler) DiscoverRelationships(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	entityID := vars["id"]

	if entityID == "" {
		RespondError(w, http.StatusBadRequest, "Entity ID is required")
		return
	}

	logger.Trace("relationship", "Discovering relationships for entity: %s", entityID)

	// Get the focus entity
	focusEntity, err := h.repo.GetByID(entityID)
	if err != nil {
		logger.Warn("Failed to get focus entity %s: %v", entityID, err)
		RespondError(w, http.StatusNotFound, "Entity not found")
		return
	}

	discovery := &RelationshipDiscovery{
		EntityID:    entityID,
		GeneratedAt: time.Now(),
	}

	// Discover direct references (entities that reference this entity in their tags)
	discovery.DirectRefs = h.discoverDirectReferences(entityID)

	// Discover tag-based relationships (entities with similar/related tags)
	discovery.TagBasedRefs = h.discoverTagBasedRelationships(focusEntity)

	// Discover shared tag relationships
	discovery.SharedTags = h.discoverSharedTagRelationships(focusEntity)

	// Calculate totals
	discovery.TotalFound = len(discovery.DirectRefs) + len(discovery.TagBasedRefs)
	discovery.NetworkDepth = 1 // Default depth

	logger.Info("Relationship discovery completed for %s: %d direct, %d tag-based, %d shared",
		entityID, len(discovery.DirectRefs), len(discovery.TagBasedRefs), len(discovery.SharedTags))

	RespondJSON(w, http.StatusOK, discovery)
}

// GetEntityNetwork returns a network graph for an entity up to specified depth
func (h *EntityRelationshipHandler) GetEntityNetwork(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	entityID := vars["id"]
	depthStr := vars["depth"]

	if entityID == "" {
		RespondError(w, http.StatusBadRequest, "Entity ID is required")
		return
	}

	depth := 1
	if depthStr != "" {
		if d, err := strconv.Atoi(depthStr); err == nil && d > 0 && d <= 5 {
			depth = d
		}
	}

	logger.Trace("relationship", "Building network graph for entity: %s, depth: %d", entityID, depth)

	network := h.buildEntityNetwork(entityID, depth)
	RespondJSON(w, http.StatusOK, network)
}

// GetRelatedByTags returns entities related through specific tag relationships
func (h *EntityRelationshipHandler) GetRelatedByTags(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	entityID := vars["id"]

	// Get query parameters
	tagPattern := r.URL.Query().Get("tag_pattern")
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	entity, err := h.repo.GetByID(entityID)
	if err != nil {
		RespondError(w, http.StatusNotFound, "Entity not found")
		return
	}

	relatedEntities := h.findRelatedByTags(entity, tagPattern, limit)
	RespondJSON(w, http.StatusOK, relatedEntities)
}

// Helper methods for relationship discovery

func (h *EntityRelationshipHandler) discoverDirectReferences(entityID string) []EntityReference {
	references := []EntityReference{}

	// Search for entities that contain this entity ID in their tags
	shortID := entityID
	if len(entityID) > 8 {
		shortID = entityID[:8]
	}

	// Search by entity ID patterns
	searchPatterns := []string{
		entityID,
		shortID,
		"ref:" + entityID,
		"relates_to:" + entityID,
		"parent:" + entityID,
		"child:" + entityID,
		"depends_on:" + entityID,
	}

	for _, pattern := range searchPatterns {
		entities, err := h.repo.ListByTags([]string{pattern}, false)
		if err != nil {
			continue
		}

		for _, entity := range entities {
			if entity.ID == entityID {
				continue // Skip self-references
			}

			// Determine relationship type from tag pattern
			relationType := h.inferRelationshipType(pattern, entity)

			ref := EntityReference{
				SourceID:     entity.ID,
				SourceName:   h.getEntityName(entity),
				SourceType:   h.getEntityType(entity),
				TargetID:     entityID,
				RelationType: relationType,
				ViaTag:       pattern,
				Confidence:   h.calculateConfidence(pattern, entity),
				CreatedAt:    time.Unix(0, entity.CreatedAt),
			}

			references = append(references, ref)
		}
	}

	return references
}

func (h *EntityRelationshipHandler) discoverTagBasedRelationships(focusEntity *models.Entity) []TagBasedRelationship {
	relationships := []TagBasedRelationship{}
	focusTags := h.parseEntityTags(focusEntity)

	// Find entities with overlapping tags
	for _, tag := range focusTags[:minRel(len(focusTags), 10)] { // Limit to first 10 tags for performance
		entities, err := h.repo.ListByTags([]string{tag}, false)
		if err != nil {
			continue
		}

		for _, entity := range entities {
			if entity.ID == focusEntity.ID {
				continue
			}

			entityTags := h.parseEntityTags(entity)
			sharedTags := h.findSharedTags(focusTags, entityTags)

			if len(sharedTags) > 0 {
				similarity := float64(len(sharedTags)) / float64(maxRel(len(focusTags), len(entityTags)))
				strength := similarity * float64(len(sharedTags)) // Weight by number of shared tags

				relationship := TagBasedRelationship{
					RelatedEntityID:   entity.ID,
					RelatedEntityName: h.getEntityName(entity),
					RelatedEntityType: h.getEntityType(entity),
					SharedTags:        sharedTags,
					TagSimilarity:     similarity,
					RelationStrength:  strength,
				}

				relationships = append(relationships, relationship)
			}
		}
	}

	// Remove duplicates and sort by strength
	relationships = h.deduplicateAndSortRelationships(relationships)

	return relationships[:minRel(len(relationships), 50)] // Limit results
}

func (h *EntityRelationshipHandler) discoverSharedTagRelationships(focusEntity *models.Entity) []SharedTagRelation {
	relations := []SharedTagRelation{}
	focusTags := h.parseEntityTags(focusEntity)

	for _, tag := range focusTags[:minRel(len(focusTags), 5)] { // Limit for performance
		entities, err := h.repo.ListByTags([]string{tag}, false)
		if err != nil {
			continue
		}

		relatedIDs := []string{}
		for _, entity := range entities {
			if entity.ID != focusEntity.ID {
				relatedIDs = append(relatedIDs, entity.ID)
			}
		}

		if len(relatedIDs) > 0 {
			namespace, value := h.parseTag(tag)
			relation := SharedTagRelation{
				Tag:          tag,
				EntityCount:  len(relatedIDs),
				RelatedIDs:   relatedIDs[:minRel(len(relatedIDs), 20)], // Limit IDs
				TagValue:     value,
				TagNamespace: namespace,
			}
			relations = append(relations, relation)
		}
	}

	return relations
}

func (h *EntityRelationshipHandler) buildEntityNetwork(focusEntityID string, depth int) *EntityNetwork {
	network := &EntityNetwork{
		FocusEntity: focusEntityID,
		Nodes:       []NetworkNode{},
		Edges:       []NetworkEdge{},
		Depth:       depth,
		GeneratedAt: time.Now(),
	}

	visited := make(map[string]bool)
	nodeQueue := []string{focusEntityID}
	currentDepth := 0

	for currentDepth < depth && len(nodeQueue) > 0 {
		nextQueue := []string{}

		for _, entityID := range nodeQueue {
			if visited[entityID] {
				continue
			}
			visited[entityID] = true

			// Get entity details
			entity, err := h.repo.GetByID(entityID)
			if err != nil {
				continue
			}

			// Add node
			node := NetworkNode{
				ID:       entity.ID,
				Name:     h.getEntityName(entity),
				Type:     h.getEntityType(entity),
				Tags:     h.parseEntityTags(entity),
				IsFocus:  entity.ID == focusEntityID,
				Distance: currentDepth,
				Properties: map[string]string{
					"created_at": time.Unix(0, entity.CreatedAt).Format(time.RFC3339),
				},
			}

			network.Nodes = append(network.Nodes, node)

			// Find connected entities for next depth level
			if currentDepth < depth-1 {
				connected := h.findConnectedEntities(entity)
				for _, connectedID := range connected {
					if !visited[connectedID] {
						nextQueue = append(nextQueue, connectedID)

						// Add edge
						edge := NetworkEdge{
							Source: entity.ID,
							Target: connectedID,
							Type:   "related",
							Weight: 1.0,
						}
						network.Edges = append(network.Edges, edge)
					}
				}
			}
		}

		nodeQueue = nextQueue
		currentDepth++
	}

	// Calculate connection counts
	for i := range network.Nodes {
		connections := 0
		for _, edge := range network.Edges {
			if edge.Source == network.Nodes[i].ID || edge.Target == network.Nodes[i].ID {
				connections++
			}
		}
		network.Nodes[i].Connections = connections
	}

	network.TotalNodes = len(network.Nodes)
	network.TotalEdges = len(network.Edges)

	return network
}

// Utility helper methods

func (h *EntityRelationshipHandler) parseEntityTags(entity *models.Entity) []string {
	cleanTags := []string{}
	for _, tag := range entity.Tags {
		// Remove temporal timestamp prefix if present
		if strings.Contains(tag, "|") {
			parts := strings.Split(tag, "|")
			if len(parts) >= 2 {
				cleanTags = append(cleanTags, parts[1])
			}
		} else {
			cleanTags = append(cleanTags, tag)
		}
	}
	return cleanTags
}

func (h *EntityRelationshipHandler) getEntityName(entity *models.Entity) string {
	tags := h.parseEntityTags(entity)
	for _, tag := range tags {
		if strings.HasPrefix(tag, "name:") {
			return strings.TrimPrefix(tag, "name:")
		}
	}
	return entity.ID[:minRel(len(entity.ID), 8)]
}

func (h *EntityRelationshipHandler) getEntityType(entity *models.Entity) string {
	tags := h.parseEntityTags(entity)
	for _, tag := range tags {
		if strings.HasPrefix(tag, "type:") {
			return strings.TrimPrefix(tag, "type:")
		}
	}
	return "unknown"
}

func (h *EntityRelationshipHandler) inferRelationshipType(pattern string, entity *models.Entity) string {
	if strings.HasPrefix(pattern, "ref:") {
		return "references"
	}
	if strings.HasPrefix(pattern, "relates_to:") {
		return "relates_to"
	}
	if strings.HasPrefix(pattern, "parent:") {
		return "parent_of"
	}
	if strings.HasPrefix(pattern, "child:") {
		return "child_of"
	}
	if strings.HasPrefix(pattern, "depends_on:") {
		return "depends_on"
	}
	return "related"
}

func (h *EntityRelationshipHandler) calculateConfidence(pattern string, entity *models.Entity) float64 {
	// Higher confidence for explicit relationship tags
	if strings.Contains(pattern, ":") {
		return 0.9
	}
	// Lower confidence for ID-based matches
	return 0.6
}

func (h *EntityRelationshipHandler) findSharedTags(tags1, tags2 []string) []string {
	tagSet := make(map[string]bool)
	for _, tag := range tags1 {
		tagSet[tag] = true
	}

	shared := []string{}
	for _, tag := range tags2 {
		if tagSet[tag] {
			shared = append(shared, tag)
		}
	}
	return shared
}

func (h *EntityRelationshipHandler) parseTag(tag string) (namespace, value string) {
	parts := strings.SplitN(tag, ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", tag
}

func (h *EntityRelationshipHandler) findConnectedEntities(entity *models.Entity) []string {
	connected := []string{}
	tags := h.parseEntityTags(entity)

	for _, tag := range tags {
		// Look for entity ID references in tags
		if strings.Contains(tag, ":") && len(tag) > 30 { // Likely contains entity ID
			parts := strings.Split(tag, ":")
			if len(parts) >= 2 {
				possibleID := parts[1]
				if len(possibleID) >= 8 { // Could be entity ID
					connected = append(connected, possibleID)
				}
			}
		}
	}

	return connected
}

func (h *EntityRelationshipHandler) findRelatedByTags(entity *models.Entity, tagPattern string, limit int) []TagBasedRelationship {
	// Implementation for tag-based entity discovery
	return h.discoverTagBasedRelationships(entity)[:minRel(limit, len(h.discoverTagBasedRelationships(entity)))]
}

func (h *EntityRelationshipHandler) deduplicateAndSortRelationships(relationships []TagBasedRelationship) []TagBasedRelationship {
	seen := make(map[string]bool)
	unique := []TagBasedRelationship{}

	for _, rel := range relationships {
		if !seen[rel.RelatedEntityID] {
			seen[rel.RelatedEntityID] = true
			unique = append(unique, rel)
		}
	}

	// Simple sort by relation strength (descending)
	for i := 0; i < len(unique); i++ {
		for j := i + 1; j < len(unique); j++ {
			if unique[i].RelationStrength < unique[j].RelationStrength {
				unique[i], unique[j] = unique[j], unique[i]
			}
		}
	}

	return unique
}

// Utility functions for relationship handler
func minRel(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxRel(a, b int) int {
	if a > b {
		return a
	}
	return b
}