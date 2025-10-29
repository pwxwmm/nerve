// Package security provides fine-grained permission control and RBAC functionality.
//
// Author: mmwei3 (2025-10-28)
// Wethers: cloudWays
package security

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

// Permission represents a permission
type Permission struct {
	Resource string   `json:"resource"`
	Actions  []string `json:"actions"`
}

// Role represents a user role
type Role struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions"`
}

// User represents a user
type User struct {
	ID       string   `json:"id"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
	IsActive bool     `json:"is_active"`
}

// PermissionManager manages permissions and roles
type PermissionManager struct {
	roles       map[string]*Role
	users       map[string]*User
	permissions map[string]map[string]bool // resource -> action -> allowed
	mutex       sync.RWMutex
}

// NewPermissionManager creates a new permission manager
func NewPermissionManager() *PermissionManager {
	pm := &PermissionManager{
		roles:       make(map[string]*Role),
		users:       make(map[string]*User),
		permissions: make(map[string]map[string]bool),
	}

	// Initialize default roles
	pm.initializeDefaultRoles()

	return pm
}

// initializeDefaultRoles creates default roles
func (pm *PermissionManager) initializeDefaultRoles() {
	// Admin role
	adminRole := &Role{
		ID:          "admin",
		Name:        "Administrator",
		Description: "Full system access",
		Permissions: []Permission{
			{Resource: "*", Actions: []string{"*"}},
		},
	}
	pm.roles["admin"] = adminRole

	// Agent role
	agentRole := &Role{
		ID:          "agent",
		Name:        "Agent",
		Description: "Agent operations",
		Permissions: []Permission{
			{Resource: "agents", Actions: []string{"read", "update"}},
			{Resource: "tasks", Actions: []string{"read", "execute"}},
			{Resource: "system_info", Actions: []string{"read", "update"}},
		},
	}
	pm.roles["agent"] = agentRole

	// Operator role
	operatorRole := &Role{
		ID:          "operator",
		Name:        "Operator",
		Description: "System operations",
		Permissions: []Permission{
			{Resource: "agents", Actions: []string{"read", "create", "update", "delete"}},
			{Resource: "tasks", Actions: []string{"read", "create", "update", "delete"}},
			{Resource: "clusters", Actions: []string{"read", "create", "update", "delete"}},
			{Resource: "alerts", Actions: []string{"read", "create", "update", "delete"}},
		},
	}
	pm.roles["operator"] = operatorRole

	// Viewer role
	viewerRole := &Role{
		ID:          "viewer",
		Name:        "Viewer",
		Description: "Read-only access",
		Permissions: []Permission{
			{Resource: "agents", Actions: []string{"read"}},
			{Resource: "tasks", Actions: []string{"read"}},
			{Resource: "clusters", Actions: []string{"read"}},
			{Resource: "alerts", Actions: []string{"read"}},
		},
	}
	pm.roles["viewer"] = viewerRole

	// Build permission map
	pm.buildPermissionMap()
}

// buildPermissionMap builds the permission lookup map
func (pm *PermissionManager) buildPermissionMap() {
	pm.permissions = make(map[string]map[string]bool)

	for _, role := range pm.roles {
		for _, perm := range role.Permissions {
			if perm.Resource == "*" {
				// Wildcard resource
				if pm.permissions["*"] == nil {
					pm.permissions["*"] = make(map[string]bool)
				}
				for _, action := range perm.Actions {
					if action == "*" {
						pm.permissions["*"]["*"] = true
					} else {
						pm.permissions["*"][action] = true
					}
				}
			} else {
				if pm.permissions[perm.Resource] == nil {
					pm.permissions[perm.Resource] = make(map[string]bool)
				}
				for _, action := range perm.Actions {
					if action == "*" {
						pm.permissions[perm.Resource]["*"] = true
					} else {
						pm.permissions[perm.Resource][action] = true
					}
				}
			}
		}
	}
}

// AddRole adds a new role
func (pm *PermissionManager) AddRole(role *Role) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if _, exists := pm.roles[role.ID]; exists {
		return fmt.Errorf("role %s already exists", role.ID)
	}

	pm.roles[role.ID] = role
	pm.buildPermissionMap()

	return nil
}

// GetRole retrieves a role by ID
func (pm *PermissionManager) GetRole(roleID string) (*Role, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	role, exists := pm.roles[roleID]
	if !exists {
		return nil, fmt.Errorf("role %s not found", roleID)
	}

	return role, nil
}

// ListRoles returns all roles
func (pm *PermissionManager) ListRoles() []*Role {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	roles := make([]*Role, 0, len(pm.roles))
	for _, role := range pm.roles {
		roles = append(roles, role)
	}

	return roles
}

// AddUser adds a new user
func (pm *PermissionManager) AddUser(user *User) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	if _, exists := pm.users[user.ID]; exists {
		return fmt.Errorf("user %s already exists", user.ID)
	}

	// Validate roles
	for _, roleID := range user.Roles {
		if _, exists := pm.roles[roleID]; !exists {
			return fmt.Errorf("role %s not found", roleID)
		}
	}

	pm.users[user.ID] = user
	return nil
}

// GetUser retrieves a user by ID
func (pm *PermissionManager) GetUser(userID string) (*User, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	user, exists := pm.users[userID]
	if !exists {
		return nil, fmt.Errorf("user %s not found", userID)
	}

	return user, nil
}

// ListUsers returns all users
func (pm *PermissionManager) ListUsers() []*User {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	users := make([]*User, 0, len(pm.users))
	for _, user := range pm.users {
		users = append(users, user)
	}

	return users
}

// CheckPermission checks if a user has permission for a resource and action
func (pm *PermissionManager) CheckPermission(userID, resource, action string) bool {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	user, exists := pm.users[userID]
	if !exists || !user.IsActive {
		return false
	}

	// Check wildcard permissions first
	if wildcardPerms, exists := pm.permissions["*"]; exists {
		if wildcardPerms["*"] || wildcardPerms[action] {
			return true
		}
	}

	// Check resource-specific permissions
	if resourcePerms, exists := pm.permissions[resource]; exists {
		if resourcePerms["*"] || resourcePerms[action] {
			return true
		}
	}

	// Check user roles
	for _, roleID := range user.Roles {
		role, exists := pm.roles[roleID]
		if !exists {
			continue
		}

		for _, perm := range role.Permissions {
			if pm.matchesResource(perm.Resource, resource) && pm.matchesAction(perm.Actions, action) {
				return true
			}
		}
	}

	return false
}

// matchesResource checks if a permission resource matches the requested resource
func (pm *PermissionManager) matchesResource(permResource, requestedResource string) bool {
	if permResource == "*" {
		return true
	}

	// Check for exact match
	if permResource == requestedResource {
		return true
	}

	// Check for wildcard match (e.g., "agents/*" matches "agents/123")
	if strings.HasSuffix(permResource, "/*") {
		prefix := strings.TrimSuffix(permResource, "/*")
		return strings.HasPrefix(requestedResource, prefix+"/")
	}

	return false
}

// matchesAction checks if permission actions include the requested action
func (pm *PermissionManager) matchesAction(permActions []string, requestedAction string) bool {
	for _, action := range permActions {
		if action == "*" || action == requestedAction {
			return true
		}
	}
	return false
}

// GetUserPermissions returns all permissions for a user
func (pm *PermissionManager) GetUserPermissions(userID string) ([]Permission, error) {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	user, exists := pm.users[userID]
	if !exists {
		return nil, fmt.Errorf("user %s not found", userID)
	}

	var permissions []Permission
	for _, roleID := range user.Roles {
		role, exists := pm.roles[roleID]
		if !exists {
			continue
		}

		permissions = append(permissions, role.Permissions...)
	}

	return permissions, nil
}

// UpdateUserRoles updates user roles
func (pm *PermissionManager) UpdateUserRoles(userID string, roles []string) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	user, exists := pm.users[userID]
	if !exists {
		return fmt.Errorf("user %s not found", userID)
	}

	// Validate roles
	for _, roleID := range roles {
		if _, exists := pm.roles[roleID]; !exists {
			return fmt.Errorf("role %s not found", roleID)
		}
	}

	user.Roles = roles
	return nil
}

// PermissionMiddleware creates a middleware for permission checking
func PermissionMiddleware(permManager *PermissionManager) func(resource, action string) func(c *gin.Context) {
	return func(resource, action string) func(c *gin.Context) {
		return func(c *gin.Context) {
			userID, exists := c.Get("user_id")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
				c.Abort()
				return
			}

			if !permManager.CheckPermission(userID.(string), resource, action) {
				c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
				c.Abort()
				return
			}

			c.Next()
		}
	}
}

