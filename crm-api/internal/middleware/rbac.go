package middleware

import (
	"g4s-crm/api/internal/models"
	"g4s-crm/api/pkg/response"
	"github.com/gin-gonic/gin"
	"sort"
)

// rolePermissions maps each role to the set of permissions it holds.
var rolePermissions = map[models.UserRole]map[string]bool{
	models.RoleAdmin: {
		"users:read": true, "users:create": true, "users:update": true, "users:delete": true,
		"teams:read": true, "teams:create": true, "teams:update": true, "teams:delete": true,
		"customers:read": true, "customers:create": true, "customers:update": true, "customers:delete": true, "customers:export": true,
		"opportunities:read": true, "opportunities:create": true, "opportunities:update": true, "opportunities:delete": true, "opportunities:export": true,
		"manufacturers:read": true, "manufacturers:create": true, "manufacturers:update": true, "manufacturers:delete": true,
		"products:read": true, "products:create": true, "products:update": true, "products:delete": true, "products:export": true,
		"inventory:read": true, "inventory:create": true, "inventory:update": true, "inventory:delete": true,
		"quotes:read": true, "quotes:create": true, "quotes:update": true, "quotes:delete": true, "quotes:approve": true, "quotes:export": true,
		"projects:read": true, "projects:create": true, "projects:update": true, "projects:delete": true,
		"price-books:read": true, "price-books:create": true, "price-books:update": true, "price-books:delete": true,
		"exchange-rates:read": true, "exchange-rates:create": true, "exchange-rates:update": true,
		"contracts:read": true, "contracts:create": true, "contracts:update": true, "contracts:delete": true, "contracts:approve": true, "contracts:export": true,
		"recurring-services:read": true, "recurring-services:create": true, "recurring-services:update": true, "recurring-services:delete": true,
		"procurement:read": true, "procurement:create": true, "procurement:update": true, "procurement:delete": true, "procurement:approve": true, "procurement:export": true,
		"documents:read": true, "documents:create": true, "documents:update": true, "documents:delete": true,
		"dashboard:read": true,
	},
	models.RoleSalesManager: {
		"customers:read": true, "customers:create": true, "customers:update": true, "customers:export": true,
		"opportunities:read": true, "opportunities:create": true, "opportunities:update": true, "opportunities:delete": true, "opportunities:export": true,
		"quotes:read": true, "quotes:create": true, "quotes:update": true, "quotes:approve": true, "quotes:export": true,
		"contracts:read": true, "contracts:create": true, "contracts:update": true, "contracts:approve": true, "contracts:export": true,
		"products:read": true, "price-books:read": true, "exchange-rates:read": true,
		"projects:read": true, "documents:read": true, "documents:create": true,
		"teams:read": true, "users:read": true, "dashboard:read": true,
		"recurring-services:read": true,
	},
	models.RoleSalesExecutive: {
		"customers:read": true, "customers:create": true, "customers:update": true,
		"opportunities:read": true, "opportunities:create": true, "opportunities:update": true,
		"quotes:read": true, "quotes:create": true, "quotes:update": true,
		"products:read": true, "price-books:read": true, "exchange-rates:read": true,
		"contracts:read": true, "projects:read": true, "documents:read": true,
		"teams:read": true, "users:read": true, "dashboard:read": true,
	},
	models.RolePreSales: {
		"customers:read":     true,
		"opportunities:read": true, "opportunities:update": true,
		"quotes:read": true, "quotes:create": true, "quotes:update": true,
		"products:read": true, "products:create": true, "products:update": true,
		"manufacturers:read": true, "price-books:read": true, "exchange-rates:read": true,
		"procurement:read": true, "documents:read": true, "documents:create": true,
		"dashboard:read": true, "users:read": true,
	},
	models.RoleProcurementManager: {
		"procurement:read": true, "procurement:create": true, "procurement:update": true, "procurement:delete": true, "procurement:approve": true, "procurement:export": true,
		"products:read": true, "products:create": true, "products:update": true,
		"manufacturers:read": true, "manufacturers:create": true, "manufacturers:update": true,
		"inventory:read": true, "inventory:create": true, "inventory:update": true,
		"exchange-rates:read": true, "exchange-rates:update": true,
		"documents:read": true, "documents:create": true,
		"dashboard:read": true, "users:read": true,
	},
	models.RoleProcurementOfficer: {
		"procurement:read": true, "procurement:create": true, "procurement:update": true,
		"products:read": true, "manufacturers:read": true,
		"inventory:read": true, "inventory:create": true,
		"exchange-rates:read": true, "documents:read": true, "dashboard:read": true,
	},
	models.RoleWarehouseManager: {
		"inventory:read": true, "inventory:create": true, "inventory:update": true, "inventory:delete": true,
		"products:read": true, "procurement:read": true,
		"documents:read": true, "dashboard:read": true, "users:read": true,
	},
	models.RoleProjectManager: {
		"projects:read": true, "projects:create": true, "projects:update": true,
		"opportunities:read": true,
		"quotes:read":        true, "procurement:read": true, "inventory:read": true,
		"customers:read": true, "products:read": true,
		"documents:read": true, "documents:create": true,
		"dashboard:read": true, "users:read": true,
	},
	models.RoleViewer: {
		"customers:read": true, "opportunities:read": true, "quotes:read": true,
		"products:read": true, "manufacturers:read": true,
		"inventory:read": true, "procurement:read": true,
		"contracts:read": true, "projects:read": true,
		"price-books:read": true, "exchange-rates:read": true,
		"recurring-services:read": true, "documents:read": true,
		"teams:read": true, "users:read": true, "dashboard:read": true,
	},
}

// Authorize returns a middleware that checks if the current user has the required permission.
func Authorize(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := GetCurrentUserRole(c)
		if role == "" {
			response.Unauthorized(c, "Authentication required")
			c.Abort()
			return
		}

		if !Can(c, permission) {
			response.Forbidden(c, "You do not have permission to perform this action")
			c.Abort()
			return
		}

		c.Next()
	}
}

// AuthorizeAny allows a request holding at least one of the listed permissions,
// for endpoints shared by more than one module.
func AuthorizeAny(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := GetCurrentUserRole(c)
		if role == "" {
			response.Unauthorized(c, "Authentication required")
			c.Abort()
			return
		}
		for _, permission := range permissions {
			if Can(c, permission) {
				c.Next()
				return
			}
		}
		response.Forbidden(c, "You do not have permission to perform this action")
		c.Abort()
	}
}

// HasPermission checks if a role has a given permission (for use in services).
func HasPermission(role models.UserRole, permission string) bool {
	perms, exists := rolePermissions[role]
	if exists && perms[permission] {
		return true
	}
	if inherited, ok := QuotePermissionDefaults[permission]; ok {
		return exists && perms[inherited]
	}
	return false
}

// Permissions is the same authoritative policy used by Authorize.
func Permissions(role models.UserRole) []string {
	permissions := []string{}
	for permission, allowed := range rolePermissions[role] {
		if allowed {
			permissions = append(permissions, permission)
		}
	}
	for permission := range QuotePermissionDefaults {
		if HasPermission(role, permission) {
			permissions = append(permissions, permission)
		}
	}
	sort.Strings(permissions)
	return permissions
}

// Existing roles retain their behavior until an administrator sets an override.
var QuotePermissionDefaults = map[string]string{
	"quotes:submit": "quotes:update", "quotes:send": "quotes:update",
	"quotes:accept": "quotes:update", "quotes:revise": "quotes:create",
	"quotes:pricing": "quotes:update", "quotes:view-margins": "products:read",
	"quotes:documents": "quotes:update", "quotes:compare-prices": "quotes:read",
}

func ValidQuotePermission(key string) bool {
	if _, ok := QuotePermissionDefaults[key]; ok {
		return true
	}
	switch key {
	case "quotes:read", "quotes:create", "quotes:update", "quotes:delete", "quotes:approve", "quotes:export":
		return true
	}
	return false
}
func UserCan(user models.User, permission string) bool {
	if allowed, exists := user.QuotePermissions[permission]; exists && ValidQuotePermission(permission) {
		return allowed
	}
	return HasPermission(user.Role, permission)
}
func UserPermissions(user models.User) []string {
	found := map[string]bool{}
	for _, permission := range Permissions(user.Role) {
		found[permission] = true
	}
	for permission, allowed := range user.QuotePermissions {
		if ValidQuotePermission(permission) {
			found[permission] = allowed
		}
	}
	out := []string{}
	for permission, allowed := range found {
		if allowed {
			out = append(out, permission)
		}
	}
	sort.Strings(out)
	return out
}
func Can(c *gin.Context, permission string) bool {
	if value, ok := c.Get("currentUser"); ok {
		if user, ok := value.(models.User); ok {
			return UserCan(user, permission)
		}
	}
	return HasPermission(GetCurrentUserRole(c), permission)
}
