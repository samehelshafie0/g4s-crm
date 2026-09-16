package models

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string
type Department string

const (
	RoleAdmin               UserRole = "admin"
	RoleSalesManager        UserRole = "sales_manager"
	RoleSalesExecutive      UserRole = "sales_executive"
	RolePreSales            UserRole = "pre_sales"
	RoleProcurementManager  UserRole = "procurement_manager"
	RoleProcurementOfficer  UserRole = "procurement_officer"
	RoleWarehouseManager    UserRole = "warehouse_manager"
	RoleProjectManager      UserRole = "project_manager"
	RoleViewer              UserRole = "viewer"
)

const (
	DeptSales       Department = "sales"
	DeptPreSales    Department = "pre-sales"
	DeptTechnical   Department = "technical"
	DeptSupport     Department = "support"
	DeptMarketing   Department = "marketing"
	DeptManagement  Department = "management"
	DeptOperations  Department = "operations"
)

type User struct {
	Base
	Email        string     `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string     `gorm:"not null" json:"-"`
	FirstName    string     `gorm:"not null" json:"firstName"`
	LastName     string     `gorm:"not null" json:"lastName"`
	Phone        string     `json:"phone"`
	Role         UserRole   `gorm:"not null;default:'viewer'" json:"role"`
	Department   Department `json:"department"`
	TeamID       *uuid.UUID `gorm:"type:uuid" json:"teamId,omitempty"`
	Team         *Team      `gorm:"foreignKey:TeamID" json:"team,omitempty"`
	IsActive     bool       `gorm:"default:true" json:"isActive"`
	LastLoginAt  *time.Time `json:"lastLoginAt,omitempty"`
}

func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

// RefreshToken stores hashed refresh tokens
type RefreshToken struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index"`
	TokenHash string     `gorm:"not null"`
	ExpiresAt time.Time  `gorm:"not null"`
	CreatedAt time.Time
	RevokedAt *time.Time
}
