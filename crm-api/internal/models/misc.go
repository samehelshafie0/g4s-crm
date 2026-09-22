package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

// ─── Price Books ───────────────────────────────────────────

type PriceBookType string

const (
	PriceBookStandard         PriceBookType = "standard"
	PriceBookVolume           PriceBookType = "volume"
	PriceBookContract         PriceBookType = "contract"
	PriceBookPromotional      PriceBookType = "promotional"
	PriceBookCustomerSpecific PriceBookType = "customer-specific"
)

type PriceBook struct {
	Base
	Name        string           `gorm:"not null" json:"name"`
	Type        PriceBookType    `gorm:"not null" json:"type"`
	Description string           `json:"description"`
	CustomerID  *uuid.UUID       `gorm:"type:uuid" json:"customerId,omitempty"`
	Customer    *Customer        `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	ContractID  *uuid.UUID       `gorm:"type:uuid" json:"contractId,omitempty"`
	ValidFrom   *time.Time       `json:"validFrom,omitempty"`
	ValidTo     *time.Time       `json:"validTo,omitempty"`
	IsActive    bool             `gorm:"default:true" json:"isActive"`
	Entries     []PriceBookEntry `gorm:"foreignKey:PriceBookID" json:"entries,omitempty"`
}

type PriceBookEntry struct {
	Base
	PriceBookID        uuid.UUID         `gorm:"type:uuid;not null;index" json:"priceBookId"`
	ProductID          *uuid.UUID        `gorm:"type:uuid" json:"productId,omitempty"`
	Kind               string            `json:"kind"`
	ServiceID          *uuid.UUID        `gorm:"type:uuid" json:"serviceId,omitempty"`
	Service            *CatalogService   `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
	RecurringServiceID *uuid.UUID        `gorm:"type:uuid" json:"recurringServiceId,omitempty"`
	RecurringService   *RecurringService `gorm:"foreignKey:RecurringServiceID" json:"recurringService,omitempty"`
	Product            *Product          `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	StandardPrice      float64           `json:"standardPrice"`
	CustomPrice        float64           `json:"customPrice"`
	DiscountPercent    float64           `json:"discountPercent"`
}

// ─── Exchange Rates ────────────────────────────────────────

type ExchangeRate struct {
	Base
	FromCurrency  Currency              `gorm:"not null;uniqueIndex:idx_fx_pair" json:"fromCurrency"`
	ToCurrency    string                `gorm:"not null;default:'SAR';uniqueIndex:idx_fx_pair" json:"toCurrency"`
	CurrentRate   float64               `gorm:"not null" json:"currentRate"`
	EffectiveDate time.Time             `gorm:"not null" json:"effectiveDate"`
	History       []ExchangeRateHistory `gorm:"foreignKey:ExchangeRateID" json:"history,omitempty"`
}

type ExchangeRateHistory struct {
	Base
	ExchangeRateID uuid.UUID `gorm:"type:uuid;not null;index" json:"exchangeRateId"`
	Rate           float64   `gorm:"not null" json:"rate"`
	EffectiveDate  time.Time `gorm:"not null" json:"effectiveDate"`
}

// ─── Recurring Services ────────────────────────────────────

type RecurringServiceType string
type BillingFrequency string

const (
	RecurringGuarding           RecurringServiceType = "guarding"
	RecurringMaintenance        RecurringServiceType = "maintenance"
	RecurringMonitoring         RecurringServiceType = "monitoring"
	RecurringPatrol             RecurringServiceType = "patrol"
	RecurringFacilityManagement RecurringServiceType = "facility-management"
	RecurringRental             RecurringServiceType = "rental"

	BillingMonthly   BillingFrequency = "monthly"
	BillingQuarterly BillingFrequency = "quarterly"
	BillingAnnually  BillingFrequency = "annually"
)

type RecurringService struct {
	Base
	LineItems           []RecurringLine      `gorm:"serializer:json;type:jsonb" json:"lineItems"`
	Name                string               `gorm:"not null" json:"name"`
	ServiceType         RecurringServiceType `json:"serviceType"`
	Description         string               `json:"description"`
	MonthlyCost         float64              `json:"monthlyCost"`
	MonthlyPrice        float64              `json:"monthlyPrice"`
	AnnualCost          float64              `json:"annualCost"`
	AnnualPrice         float64              `json:"annualPrice"`
	TargetMarginPercent float64              `json:"targetMarginPercent"`
	BillingFrequency    BillingFrequency     `json:"billingFrequency"`
	IsActive            bool                 `gorm:"default:true" json:"isActive"`
}

// ─── Documents ─────────────────────────────────────────────

type DocumentCategory string
type DocumentType string

const (
	DocCategoryContract   DocumentCategory = "contract"
	DocCategoryQuote      DocumentCategory = "quote"
	DocCategoryGeneral    DocumentCategory = "general"
	DocCategoryCompliance DocumentCategory = "compliance"
	DocCategoryLegal      DocumentCategory = "legal"

	DocTypeTerms    DocumentType = "terms"
	DocTypeDelivery DocumentType = "delivery"
	DocTypeTech     DocumentType = "technical"
	DocTypeWarranty DocumentType = "warranty"
	DocTypeSLA      DocumentType = "sla"
)

type Document struct {
	Base
	Name         string           `gorm:"not null" json:"name"`
	Category     DocumentCategory `json:"category"`
	DocumentType DocumentType     `json:"documentType"`
	Tags         pq.StringArray   `gorm:"type:text[]" json:"tags"`
	Version      string           `json:"version"`
	FileName     string           `json:"fileName"`
	FileSize     int64            `json:"fileSize"`
	FileType     string           `json:"fileType"`
	FilePath     string           `json:"-"`
	UploadedByID *uuid.UUID       `gorm:"type:uuid" json:"uploadedById,omitempty"`
	UploadedBy   *User            `gorm:"foreignKey:UploadedByID" json:"uploadedBy,omitempty"`
	Links        []DocumentLink   `gorm:"foreignKey:DocumentID" json:"links,omitempty"`
}

type DocumentLink struct {
	Base
	DocumentID uuid.UUID `gorm:"type:uuid;not null;index" json:"documentId"`
	EntityType string    `gorm:"not null" json:"entityType"`
	EntityID   uuid.UUID `gorm:"type:uuid;not null" json:"entityId"`
	EntityName string    `json:"entityName"`
}

// ─── Activity Log ──────────────────────────────────────────

type ActivityLog struct {
	Base
	UserID      *uuid.UUID `gorm:"type:uuid;index" json:"userId,omitempty"`
	User        *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	EntityType  string     `gorm:"not null;index" json:"entityType"`
	EntityID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"entityId"`
	Action      string     `gorm:"not null" json:"action"`
	Description string     `json:"description"`
	OldValues   *string    `gorm:"type:jsonb" json:"oldValues,omitempty"`
	NewValues   *string    `gorm:"type:jsonb" json:"newValues,omitempty"`
}

// ─── Sequence ──────────────────────────────────────────────

type Sequence struct {
	Name    string `gorm:"primaryKey" json:"name"`
	Prefix  string `gorm:"not null" json:"prefix"`
	Year    int    `gorm:"not null" json:"year"`
	Current int    `gorm:"not null;default:0" json:"current"`
}

type DocumentVersion struct {
	Base
	DocumentID   uuid.UUID  `gorm:"type:uuid" json:"documentId"`
	Version      string     `json:"version"`
	FileName     string     `json:"fileName"`
	FileSize     int64      `json:"fileSize"`
	FileType     string     `json:"fileType"`
	FilePath     string     `json:"-"`
	UploadedByID *uuid.UUID `gorm:"type:uuid" json:"uploadedById,omitempty"`
}

func (ExchangeRateHistory) TableName() string { return "exchange_rate_history" }

type RecurringLine struct {
	ID          string     `json:"id" validate:"max=100"`
	Source      string     `json:"source" validate:"required,oneof=product service vendor write-in"`
	SourceID    *uuid.UUID `json:"sourceId,omitempty"`
	SKU         string     `json:"sku" validate:"max=100"`
	Name        string     `json:"name" validate:"required,max=255"`
	Description string     `json:"description" validate:"max=2000"`
	Qty         float64    `json:"qty" validate:"gt=0,lte=1000000"`
	UnitCost    float64    `json:"unitCost" validate:"gte=0,lte=100000000"`
	UnitPrice   float64    `json:"unitPrice" validate:"gte=0,lte=100000000"`
	VendorName  string     `json:"vendorName" validate:"max=255"`
	VendorFile  string     `json:"vendorFile" validate:"max=255"`
}
