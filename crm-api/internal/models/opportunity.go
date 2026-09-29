package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type OpportunityStage string
type ServiceType string

const (
	StageQualification OpportunityStage = "qualification"
	StageProposal      OpportunityStage = "proposal"
	StageNegotiation   OpportunityStage = "negotiation"
	StageClosedWon     OpportunityStage = "closed-won"
	StageClosedLost    OpportunityStage = "closed-lost"

	ServiceCCTV               ServiceType = "cctv"
	ServiceAccessControl      ServiceType = "access-control"
	ServiceIntrusionDetection ServiceType = "intrusion-detection"
	ServiceFireAlarm          ServiceType = "fire-alarm"
	ServiceNetworking         ServiceType = "networking"
	ServiceITSolutions        ServiceType = "it-solutions"
	ServiceGuarding           ServiceType = "guarding"
	ServiceMonitoring         ServiceType = "monitoring"
	ServiceMaintenance        ServiceType = "maintenance"
	ServiceConsulting         ServiceType = "consulting"
)

type Opportunity struct {
	Base
	Title             string           `gorm:"not null" json:"title"`
	CustomerID        uuid.UUID        `gorm:"type:uuid;not null;index" json:"customerId"`
	Customer          *Customer        `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	Stage             OpportunityStage `gorm:"not null;default:'qualification'" json:"stage"`
	ServiceTypes      pq.StringArray   `gorm:"type:text[]" json:"serviceTypes"`
	EstimatedValue    float64          `json:"estimatedValue"`
	EstimatedCost     float64          `json:"estimatedCost"`
	EstimatedMargin   float64          `json:"estimatedMargin"`
	WinProbability    int              `gorm:"check:win_probability >= 0 AND win_probability <= 100" json:"winProbability"`
	SalesExecutiveID  *uuid.UUID       `gorm:"type:uuid" json:"salesExecutiveId,omitempty"`
	SalesExecutive    *User            `gorm:"foreignKey:SalesExecutiveID" json:"salesExecutive,omitempty"`
	PreSalesID        *uuid.UUID       `gorm:"type:uuid" json:"preSalesId,omitempty"`
	PreSales          *User            `gorm:"foreignKey:PreSalesID" json:"preSales,omitempty"`
	ExpectedCloseDate *time.Time       `json:"expectedCloseDate,omitempty"`
	Notes             string           `json:"notes"`
	Quotes            []Quote          `gorm:"foreignKey:OpportunityID" json:"quotes,omitempty"`
	Sites             []CustomerSite   `gorm:"many2many:opportunity_sites;joinForeignKey:OpportunityID;joinReferences:SiteID" json:"sites"`

	// Delivery: the opportunity is the project once it is won.
	DeliveryStatus   *DeliveryStatus `json:"deliveryStatus,omitempty"`
	ProjectManagerID *uuid.UUID      `gorm:"type:uuid" json:"projectManagerId,omitempty"`
	ProjectManager   *User           `gorm:"foreignKey:ProjectManagerID" json:"projectManager,omitempty"`
	StartDate        *time.Time      `json:"startDate,omitempty"`
	TargetEndDate    *time.Time      `json:"targetEndDate,omitempty"`
	ActualEndDate    *time.Time      `json:"actualEndDate,omitempty"`
	ActualValue      *float64        `json:"actualValue,omitempty"`
	ActualCost       *float64        `json:"actualCost,omitempty"`
}
