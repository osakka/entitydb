package models

// RepositoryQueryWrapper wraps an EntityRepository to provide Query method
type RepositoryQueryWrapper struct {
	EntityRepository
}

// NewRepositoryQueryWrapper creates a new wrapper
func NewRepositoryQueryWrapper(repo EntityRepository) *RepositoryQueryWrapper {
	return &RepositoryQueryWrapper{repo}
}

// Query creates a new query builder
func (r *RepositoryQueryWrapper) Query() *EntityQuery {
	return &EntityQuery{repo: r.EntityRepository}
}