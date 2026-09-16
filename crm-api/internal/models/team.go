package models

import "github.com/google/uuid"

type Team struct {
	Base
	Name        string     `gorm:"not null" json:"name"`
	Department  Department `gorm:"not null" json:"department"`
	Description string     `json:"description"`
	LeaderID    *uuid.UUID `gorm:"type:uuid" json:"leaderId,omitempty"`
	Leader      *User      `gorm:"foreignKey:LeaderID" json:"leader,omitempty"`
	Members     []User     `gorm:"foreignKey:TeamID" json:"members,omitempty"`
	IsActive    bool       `gorm:"default:true" json:"isActive"`
}
