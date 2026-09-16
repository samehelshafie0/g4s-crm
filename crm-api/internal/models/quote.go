package models

import (
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
	Base
	QuoteNumber           string            `gorm:"uniqueIndex;not null" json:"quoteNumber"`
	OpportunityID         *uuid.UUID        `gorm:"type:uuid" json:"opportunityId,omitempty"`
	Opportunity           *Opportunity      `gorm:"foreignKey:OpportunityID" json:"opportunity,omitempty"`
	CustomerID            uuid.UUID         `gorm:"type:uuid;not null;index" json:"customerId"`
	Customer              *Customer         `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	Version               int               `gorm:"default:1" json:"version"`
	Status                QuoteStatus       `gorm:"default:'draft'" json:"status"`
	LineItems             []QuoteLineItem   `gorm:"foreignKey:QuoteID" json:"lineItems,omitempty"`
	Subtotal              float64           `json:"subtotal"`
	DiscountPercent       float64           `gorm:"default:0" json:"discountPercent"`
	DiscountAmount        float64           `json:"discountAmount"`
	SubtotalAfterDiscount float64           `json:"subtotalAfterDiscount"`
	VATPercent            float64           `gorm:"default:15" json:"vatPercent"`
	VATAmount             float64           `json:"vatAmount"`
	Total                 float64           `json:"total"`
	TotalCost             float64           `json:"totalCost"`
	MarginAmount          float64           `json:"marginAmount"`
	MarginPercent         float64           `json:"marginPercent"`
	ValidUntil            *time.Time        `json:"validUntil,omitempty"`
	Currency              Currency          `gorm:"default:'SAR'" json:"currency"`
	Notes                 string            `json:"notes"`
	ApprovedByID          *uuid.UUID        `gorm:"type:uuid" json:"approvedById,omitempty"`
	ApprovedBy            *User             `gorm:"foreignKey:ApprovedByID" json:"approvedBy,omitempty"`
	ApprovedAt            *time.Time        `json:"approvedAt,omitempty"`
	CreatedByID           *uuid.UUID        `gorm:"type:uuid" json:"createdById,omitempty"`
	CreatedBy             *User             `gorm:"foreignKey:CreatedByID" json:"createdBy,omitempty"`
}

func (q *Quote) Recalculate() {
	q.Subtotal = 0
	q.TotalCost = 0

	for _, item := range q.LineItems {
		q.Subtotal += item.LineTotal
		q.TotalCost += float64(item.Quantity) * item.UnitCost
	}

	q.DiscountAmount = q.Subtotal * q.DiscountPercent / 100
	q.SubtotalAfterDiscount = q.Subtotal - q.DiscountAmount
	q.VATAmount = q.SubtotalAfterDiscount * q.VATPercent / 100
	q.Total = q.SubtotalAfterDiscount + q.VATAmount
	q.MarginAmount = q.SubtotalAfterDiscount - q.TotalCost
	if q.SubtotalAfterDiscount > 0 {
		q.MarginPercent = (q.MarginAmount / q.SubtotalAfterDiscount) * 100
	}
}

type QuoteLineItem struct {
	Base
	QuoteID          uuid.UUID         `gorm:"type:uuid;not null;index" json:"quoteId"`
	Category         QuoteLineCategory `gorm:"not null" json:"category"`
	ProductID        *uuid.UUID        `gorm:"type:uuid" json:"productId,omitempty"`
	Product          *Product          `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	SKU              string            `json:"sku"`
	Description      string            `gorm:"not null" json:"description"`
	ManufacturerName string            `json:"manufacturerName"`
	StockAvailable   *int              `json:"stockAvailable,omitempty"`
	LeadTimeDays     *int              `json:"leadTimeDays,omitempty"`
	Quantity         int               `gorm:"not null" json:"quantity"`
	UnitCost         float64           `json:"unitCost"`
	UnitPrice        float64           `json:"unitPrice"`
	LineTotal        float64           `json:"lineTotal"`
	MarginPercent    float64           `json:"marginPercent"`
	SortOrder        int               `gorm:"default:0" json:"sortOrder"`
}

func (i *QuoteLineItem) Calculate() {
	i.LineTotal = float64(i.Quantity) * i.UnitPrice
	if i.UnitPrice > 0 {
		i.MarginPercent = ((i.UnitPrice - i.UnitCost) / i.UnitPrice) * 100
	}
}
