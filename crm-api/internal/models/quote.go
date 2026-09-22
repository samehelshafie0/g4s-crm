package models

import (
	"math"
	"time"

	"github.com/google/uuid"
)

type QuoteStatus string
type QuoteLineCategory string

const (
	QuoteStatusDraft           QuoteStatus = "draft"
	QuoteStatusPendingApproval QuoteStatus = "pending-approval"
	QuoteStatusApproved        QuoteStatus = "approved"
	QuoteStatusSent            QuoteStatus = "sent"
	QuoteStatusAccepted        QuoteStatus = "accepted"
	QuoteStatusDeclined        QuoteStatus = "declined"
	QuoteStatusExpired         QuoteStatus = "expired"

	QuoteLineMaterials     QuoteLineCategory = "materials"
	QuoteLineManpower      QuoteLineCategory = "manpower"
	QuoteLineMiscellaneous QuoteLineCategory = "miscellaneous"
)

type Quote struct {
	Appendices []QuoteAppendix `gorm:"foreignKey:QuoteID" json:"appendices"`
	Base
	LockVersion      int          `gorm:"default:1" json:"lockVersion"`
	ParentQuoteID    *uuid.UUID   `gorm:"type:uuid" json:"parentQuoteId,omitempty"`
	PaymentTerms     string       `json:"paymentTerms"`
	DeliveryTerms    string       `json:"deliveryTerms"`
	IntroductionText string       `json:"introductionText"`
	ClosingText      string       `json:"closingText"`
	PurchasingNotes  string       `json:"purchasingNotes"`
	StatementOfWork  string       `json:"statementOfWork"`
	PriceBookID      *uuid.UUID   `gorm:"type:uuid" json:"priceBookId,omitempty"`
	InternalNotes    string       `json:"internalNotes"`
	SoldTo           QuoteAddress `gorm:"serializer:json;type:jsonb" json:"soldTo"`
	ShipTo           QuoteAddress `gorm:"serializer:json;type:jsonb" json:"shipTo"`

	QuoteNumber           string          `gorm:"uniqueIndex;not null" json:"quoteNumber"`
	OpportunityID         *uuid.UUID      `gorm:"type:uuid" json:"opportunityId,omitempty"`
	Opportunity           *Opportunity    `gorm:"foreignKey:OpportunityID" json:"opportunity,omitempty"`
	CustomerID            uuid.UUID       `gorm:"type:uuid;not null;index" json:"customerId"`
	Customer              *Customer       `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	Version               int             `gorm:"default:1" json:"version"`
	Status                QuoteStatus     `gorm:"default:'draft'" json:"status"`
	LineItems             []QuoteLineItem `gorm:"foreignKey:QuoteID" json:"lineItems,omitempty"`
	Subtotal              float64         `json:"subtotal"`
	DiscountPercent       float64         `gorm:"default:0" json:"discountPercent"`
	DiscountAmount        float64         `json:"discountAmount"`
	SubtotalAfterDiscount float64         `json:"subtotalAfterDiscount"`
	VATPercent            float64         `gorm:"default:15" json:"vatPercent"`
	VATAmount             float64         `json:"vatAmount"`
	Total                 float64         `json:"total"`
	TotalCost             float64         `json:"totalCost"`
	MarginAmount          float64         `json:"marginAmount"`
	MarginPercent         float64         `json:"marginPercent"`
	ValidUntil            *time.Time      `json:"validUntil,omitempty"`
	Currency              Currency        `gorm:"default:'SAR'" json:"currency"`
	Notes                 string          `json:"notes"`
	ApprovedByID          *uuid.UUID      `gorm:"type:uuid" json:"approvedById,omitempty"`
	ApprovedBy            *User           `gorm:"foreignKey:ApprovedByID" json:"approvedBy,omitempty"`
	ApprovedAt            *time.Time      `json:"approvedAt,omitempty"`
	CreatedByID           *uuid.UUID      `gorm:"type:uuid" json:"createdById,omitempty"`
	CreatedBy             *User           `gorm:"foreignKey:CreatedByID" json:"createdBy,omitempty"`
}

func RoundMoney(value float64) float64 { return math.Round(value*100) / 100 }
func (q *Quote) Recalculate() {
	q.Subtotal = 0
	q.TotalCost = 0
	q.MarginPercent = 0
	for _, item := range q.LineItems {
		if item.RowType != "" && item.RowType != "item" {
			continue
		}
		if item.IsOptional && !item.IsSelected {
			continue
		}
		q.Subtotal += item.LineTotal
		multiplier := item.Multiplier
		if multiplier == 0 {
			multiplier = 1
		}
		q.TotalCost += RoundMoney(item.Quantity * multiplier * item.UnitCost)
	}
	q.Subtotal = RoundMoney(q.Subtotal)
	q.TotalCost = RoundMoney(q.TotalCost)
	q.DiscountAmount = RoundMoney(q.Subtotal * q.DiscountPercent / 100)
	q.SubtotalAfterDiscount = RoundMoney(q.Subtotal - q.DiscountAmount)
	q.VATAmount = RoundMoney(q.SubtotalAfterDiscount * q.VATPercent / 100)
	q.Total = RoundMoney(q.SubtotalAfterDiscount + q.VATAmount)
	q.MarginAmount = RoundMoney(q.SubtotalAfterDiscount - q.TotalCost)
	if q.SubtotalAfterDiscount > 0 {
		q.MarginPercent = q.MarginAmount / q.SubtotalAfterDiscount * 100
	}
}

type QuoteLineItem struct {
	Base
	RowType            string     `gorm:"default:'item'" json:"rowType"`
	Source             string     `gorm:"default:'product'" json:"source"`
	Multiplier         float64    `gorm:"default:1" json:"multiplier"`
	DiscountPercent    float64    `json:"discountPercent"`
	IsOptional         bool       `json:"isOptional"`
	IsSelected         bool       `gorm:"default:true" json:"isSelected"`
	IsPrintable        bool       `gorm:"default:true" json:"isPrintable"`
	HeadingText        string     `json:"headingText"`
	CommentText        string     `json:"commentText"`
	RateType           string     `json:"rateType"`
	BillingCycle       string     `json:"billingCycle"`
	ServiceID          *uuid.UUID `gorm:"type:uuid" json:"serviceId,omitempty"`
	RecurringServiceID *uuid.UUID `gorm:"type:uuid" json:"recurringServiceId,omitempty"`

	QuoteID          uuid.UUID         `gorm:"type:uuid;not null;index" json:"quoteId"`
	Category         QuoteLineCategory `gorm:"not null" json:"category"`
	ProductID        *uuid.UUID        `gorm:"type:uuid" json:"productId,omitempty"`
	Product          *Product          `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	SKU              string            `json:"sku"`
	Description      string            `gorm:"not null" json:"description"`
	ManufacturerName string            `json:"manufacturerName"`
	StockAvailable   *int              `json:"stockAvailable,omitempty"`
	LeadTimeDays     *int              `json:"leadTimeDays,omitempty"`
	Quantity         float64           `gorm:"not null" json:"quantity"`
	UnitCost         float64           `json:"unitCost"`
	UnitPrice        float64           `json:"unitPrice"`
	LineTotal        float64           `json:"lineTotal"`
	MarginPercent    float64           `json:"marginPercent"`
	SortOrder        int               `gorm:"default:0" json:"sortOrder"`
}

func (i *QuoteLineItem) Calculate() {
	i.LineTotal = 0
	i.MarginPercent = 0
	if i.RowType != "" && i.RowType != "item" {
		return
	}
	if i.Multiplier == 0 {
		i.Multiplier = 1
	}
	i.LineTotal = RoundMoney(i.Quantity * i.Multiplier * i.UnitPrice * (1 - i.DiscountPercent/100))
	if i.LineTotal > 0 {
		i.MarginPercent = (i.LineTotal - i.Quantity*i.Multiplier*i.UnitCost) / i.LineTotal * 100
	}
}

type QuoteAddress struct {
	ContactName string `json:"contactName"`
	Company     string `json:"company"`
	Address     string `json:"address"`
	City        string `json:"city"`
	Country     string `json:"country"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
}
type CatalogService struct {
	Base
	SKU         string     `json:"sku"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Department  Department `json:"department"`
	RateType    string     `json:"rateType"`
	UnitCost    float64    `json:"unitCost"`
	UnitPrice   float64    `json:"unitPrice"`
	IsActive    bool       `json:"isActive"`
}

// QuoteAppendix references an immutable document version; labels and order belong to the quote.
type QuoteAppendix struct {
	Base
	QuoteID           uuid.UUID `gorm:"type:uuid" json:"quoteId"`
	DocumentID        uuid.UUID `gorm:"type:uuid" json:"documentId"`
	DocumentVersionID uuid.UUID `gorm:"type:uuid" json:"documentVersionId"`
	Label             string    `json:"label"`
	DocumentName      string    `json:"documentName"`
	Version           string    `json:"version"`
	FileName          string    `json:"fileName"`
	FileType          string    `json:"fileType"`
	FileSize          int64     `json:"fileSize"`
	SortOrder         int       `json:"sortOrder"`
}

func (QuoteAppendix) TableName() string { return "quote_appendices" }
