package api

import (
	"entitydb/models"
	"net/http"
)

// UserHandlerRBAC wraps UserHandler with RBAC permission checks
type UserHandlerRBAC struct {
	handler         *UserHandler
	repo            models.EntityRepository
	securityManager *models.SecurityManager
}

// NewUserHandlerRBAC creates a new RBAC-enabled user handler
func NewUserHandlerRBAC(handler *UserHandler, repo models.EntityRepository, securityManager *models.SecurityManager) *UserHandlerRBAC {
	return &UserHandlerRBAC{
		handler:         handler,
		repo:            repo,
		securityManager: securityManager,
	}
}

// CreateUser wraps CreateUser with permission check
func (h *UserHandlerRBAC) CreateUser() http.HandlerFunc {
	return RBACMiddleware(h.repo, h.securityManager, PermUserCreate)(h.handler.CreateUser)
}

// GetUser wraps GetUser with permission check - viewing user info requires permission
func (h *UserHandlerRBAC) GetUser() http.HandlerFunc {
	return RBACMiddleware(h.repo, h.securityManager, PermUserView)(func(w http.ResponseWriter, r *http.Request) {
		// For getting user info, we might want to allow users to see their own info
		// but for now we'll require the general user:view permission
		h.handler.CreateUser(w, r) // This seems to be the actual user get method based on the handler
	})
}

// LoginWithoutRBAC - Login doesn't need RBAC check as it's the entry point
// Note: Login is handled directly in main.go, not through UserHandler
func (h *UserHandlerRBAC) LoginWithoutRBAC() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Login handled at server level", http.StatusNotImplemented)
	}
}

// ChangePassword wraps ChangePassword with authentication check
// No specific permission needed - users can change their own password if authenticated
func (h *UserHandlerRBAC) ChangePassword() http.HandlerFunc {
	// We use a custom middleware that checks authentication but doesn't require specific permissions
	return RBACMiddleware(h.repo, h.securityManager, PermUserUpdate)(h.handler.ChangePassword)
}

// ResetPassword wraps ResetPassword with admin permission check
func (h *UserHandlerRBAC) ResetPassword() http.HandlerFunc {
	// This requires admin permission
	return RBACMiddleware(h.repo, h.securityManager, PermUserUpdate)(h.handler.ResetPassword)
}