package middleware

import (
	"g4s-crm/api/internal/models"
	"g4s-crm/api/pkg/response"
	"github.com/gin-gonic/gin"
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
		"customers:read": true,
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
		"quotes:read": true, "procurement:read": true, "inventory:read": true,
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

		perms, exists := rolePermissions[role]
		if !exists || !perms[permission] {
			response.Forbidden(c, "You do not have permission to perform this action")
			c.Abort()
			return
		}

		c.Next()
	}
}

// HasPermission checks if a role has a given permission (for use in services).
func HasPermission(role models.UserRole, permission string) bool {
	perms, exists := rolePermissions[role]
	return exists && perms[permission]
}
