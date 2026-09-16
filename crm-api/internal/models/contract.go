package models

import (
	"time"

	"github.com/google/uuid"
)

type ContractType string
type ContractStatus string

const (
	ContractTypeSales        ContractType = "sales"
	ContractTypeMaintenance  ContractType = "maintenance"
	ContractTypeService      ContractType = "service"
	ContractTypeProject      ContractType = "project"
	ContractTypeSubscription ContractType = "subscription"

	ContractStatusDraft           ContractStatus = "draft"
	ContractStatusPendingApproval ContractStatus = "pending-approval"
	ContractStatusActive          ContractStatus = "active"
	ContractStatusExpired         ContractStatus = "expired"
	ContractStatusTerminated      ContractStatus = "terminated"
	ContractStatusRenewed         ContractStatus = "renewed"
)

type Contract struct {
	Base
	Currency          Currency       `gorm:"default:'SAR'" json:"currency"`
	ContractNumber    string         `gorm:"uniqueIndex;not null" json:"contractNumber"`
	Title             string         `gorm:"not null" json:"title"`
	CustomerID        uuid.UUID      `gorm:"type:uuid;not null;index" json:"customerId"`
	Customer          *Customer      `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	Type              ContractType   `gorm:"not null" json:"type"`
	Status            ContractStatus `gorm:"default:'draft'" json:"status"`
	StartDate         *time.Time     `json:"startDate,omitempty"`
	EndDate           *time.Time     `json:"endDate,omitempty"`
	Value             float64        `json:"value"`
	AutoRenew         bool           `gorm:"default:false" json:"autoRenew"`
	RenewalNoticeDays int            `gorm:"default:30" json:"renewalNoticeDays"`
	RenewedFromID     *uuid.UUID     `gorm:"type:uuid" json:"renewedFromId,omitempty"`
	QuoteID           *uuid.UUID     `gorm:"type:uuid" json:"quoteId,omitempty"`
	Quote             *Quote         `gorm:"foreignKey:QuoteID" json:"quote,omitempty"`
	Terms             string         `json:"terms"`
	Notes             string         `json:"notes"`
}
