package models

import (
	"time"

	"github.com/google/uuid"
)

type ProjectStatus string
type ProjectPriority string

const (
	ProjectStatusPlanning   ProjectStatus = "planning"
	ProjectStatusInProgress ProjectStatus = "in-progress"
	ProjectStatusOnHold     ProjectStatus = "on-hold"
	ProjectStatusCompleted  ProjectStatus = "completed"
	ProjectStatusCancelled  ProjectStatus = "cancelled"

	PriorityLow      ProjectPriority = "low"
	PriorityMedium   ProjectPriority = "medium"
	PriorityHigh     ProjectPriority = "high"
	PriorityCritical ProjectPriority = "critical"
)

type Project struct {
	Base
	Currency         Currency        `gorm:"default:'SAR'" json:"currency"`
	ProjectNumber    string          `gorm:"uniqueIndex;not null" json:"projectNumber"`
	Name             string          `gorm:"not null" json:"name"`
	CustomerID       uuid.UUID       `gorm:"type:uuid;not null;index" json:"customerId"`
	Customer         *Customer       `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	QuoteID          *uuid.UUID      `gorm:"type:uuid" json:"quoteId,omitempty"`
	Quote            *Quote          `gorm:"foreignKey:QuoteID" json:"quote,omitempty"`
	Status           ProjectStatus   `gorm:"default:'planning'" json:"status"`
	Priority         ProjectPriority `gorm:"default:'medium'" json:"priority"`
	StartDate        *time.Time      `json:"startDate,omitempty"`
	TargetEndDate    *time.Time      `json:"targetEndDate,omitempty"`
	ActualEndDate    *time.Time      `json:"actualEndDate,omitempty"`
	ProjectManagerID *uuid.UUID      `gorm:"type:uuid" json:"projectManagerId,omitempty"`
	ProjectManager   *User           `gorm:"foreignKey:ProjectManagerID" json:"projectManager,omitempty"`
	TotalValue       float64         `json:"totalValue"`
	TotalCost        float64         `json:"totalCost"`
	MarginPercent    float64         `json:"marginPercent"`
	Notes            string          `json:"notes"`
}
