package models

import (
	"time"

	"github.com/google/uuid"
)

type PurchaseOrderStatus string
type SupplierQuoteStatus string
type GoodsCondition string

const (
	POStatusDraft           PurchaseOrderStatus = "draft"
	POStatusPendingApproval PurchaseOrderStatus = "pending-approval"
	POStatusApproved        PurchaseOrderStatus = "approved"
	POStatusOrdered         PurchaseOrderStatus = "ordered"
	POStatusPartialReceived PurchaseOrderStatus = "partial-received"
	POStatusReceived        PurchaseOrderStatus = "received"
	POStatusCancelled       PurchaseOrderStatus = "cancelled"

	SQStatusReceived    SupplierQuoteStatus = "received"
	SQStatusUnderReview SupplierQuoteStatus = "under-review"
	SQStatusAccepted    SupplierQuoteStatus = "accepted"
	SQStatusExpired     SupplierQuoteStatus = "expired"
	SQStatusRejected    SupplierQuoteStatus = "rejected"

	ConditionGood          GoodsCondition = "good"
	ConditionDamaged       GoodsCondition = "damaged"
	ConditionPartialDamage GoodsCondition = "partial-damage"
)

type PurchaseOrder struct {
	Base
	PONumber         string              `gorm:"uniqueIndex;not null" json:"poNumber"`
	SupplierID       *uuid.UUID          `gorm:"type:uuid" json:"supplierId,omitempty"`
	Supplier         *Manufacturer       `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
	SupplierName     string              `gorm:"not null" json:"supplierName"`
	Status           PurchaseOrderStatus `gorm:"default:'draft'" json:"status"`
	Items            []PurchaseOrderItem `gorm:"foreignKey:PurchaseOrderID" json:"items,omitempty"`
	Subtotal         float64             `json:"subtotal"`
	ShippingCost     float64             `gorm:"default:0" json:"shippingCost"`
	CustomsDuty      float64             `gorm:"default:0" json:"customsDuty"`
	Total            float64             `json:"total"`
	Currency         Currency            `gorm:"default:'SAR'" json:"currency"`
	ExpectedDelivery *time.Time          `json:"expectedDelivery,omitempty"`
	ActualDelivery   *time.Time          `json:"actualDelivery,omitempty"`
	SupplierQuoteID  *uuid.UUID          `gorm:"type:uuid" json:"supplierQuoteId,omitempty"`
	SourceQuoteID    *uuid.UUID          `gorm:"type:uuid" json:"sourceQuoteId,omitempty"`
	SourceQuote      *Quote              `gorm:"foreignKey:SourceQuoteID" json:"sourceQuote,omitempty"`
	ProjectID        *uuid.UUID          `gorm:"type:uuid" json:"projectId,omitempty"`
	Project          *Project            `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Notes            string              `json:"notes"`
	ApprovedByID     *uuid.UUID          `gorm:"type:uuid" json:"approvedById,omitempty"`
	ApprovedBy       *User               `gorm:"foreignKey:ApprovedByID" json:"approvedBy,omitempty"`
	ApprovedAt       *time.Time          `json:"approvedAt,omitempty"`
	CreatedByID      *uuid.UUID          `gorm:"type:uuid" json:"createdById,omitempty"`
}

type PurchaseOrderItem struct {
	Base
	PurchaseOrderID uuid.UUID `gorm:"type:uuid;not null;index" json:"purchaseOrderId"`
	ProductID       uuid.UUID `gorm:"type:uuid;not null" json:"productId"`
	Product         *Product  `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Quantity        int       `gorm:"not null" json:"quantity"`
	UnitCost        float64   `json:"unitCost"`
	Total           float64   `json:"total"`
	ReceivedQty     int       `gorm:"default:0" json:"receivedQty"`
	LeadTimeDays    int       `json:"leadTimeDays"`
}

type SupplierQuote struct {
	Base
	SQNumber      string              `gorm:"uniqueIndex;not null" json:"sqNumber"`
	SupplierID    *uuid.UUID          `gorm:"type:uuid" json:"supplierId,omitempty"`
	Supplier      *Manufacturer       `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
	SupplierName  string              `gorm:"not null" json:"supplierName"`
	SupplierRef   string              `json:"supplierRef"`
	SourceQuoteID *uuid.UUID          `gorm:"type:uuid" json:"sourceQuoteId,omitempty"`
	SourceQuote   *Quote              `gorm:"foreignKey:SourceQuoteID" json:"sourceQuote,omitempty"`
	ProjectID     *uuid.UUID          `gorm:"type:uuid" json:"projectId,omitempty"`
	Project       *Project            `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Status        SupplierQuoteStatus `gorm:"default:'received'" json:"status"`
	Items         []SupplierQuoteItem `gorm:"foreignKey:SupplierQuoteID" json:"items,omitempty"`
	Subtotal      float64             `json:"subtotal"`
	Currency      Currency            `gorm:"default:'USD'" json:"currency"`
	ValidFrom     *time.Time          `json:"validFrom,omitempty"`
	ValidUntil    *time.Time          `json:"validUntil,omitempty"`
	ContactName   string              `json:"contactName"`
	ContactEmail  string              `json:"contactEmail"`
	PaymentTerms  string              `json:"paymentTerms"`
	DeliveryTerms string              `json:"deliveryTerms"`
	Notes         string              `json:"notes"`
}

type SupplierQuoteItem struct {
	Base
	SupplierQuoteID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"supplierQuoteId"`
	ProductID        *uuid.UUID `gorm:"type:uuid" json:"productId,omitempty"`
	Product          *Product   `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	ProductSKU       string     `json:"productSku"`
	ProductName      string     `json:"productName"`
	ManufacturerName string     `json:"manufacturerName"`
	Quantity         int        `json:"quantity"`
	UnitCost         float64    `json:"unitCost"`
	Total            float64    `json:"total"`
	LeadTimeDays     int        `json:"leadTimeDays"`
	MOQ              int        `json:"moq"`
	ValidUntil       *time.Time `json:"validUntil,omitempty"`
	Notes            string     `json:"notes"`
}

type SupplierItemCatalog struct {
	Base
	SupplierID    uuid.UUID     `gorm:"type:uuid;not null;uniqueIndex:idx_supplier_product" json:"supplierId"`
	Supplier      *Manufacturer `gorm:"foreignKey:SupplierID" json:"supplier,omitempty"`
	ProductID     uuid.UUID     `gorm:"type:uuid;not null;uniqueIndex:idx_supplier_product" json:"productId"`
	Product       *Product      `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	LatestCost    float64       `json:"latestCost"`
	PreviousCost  *float64      `json:"previousCost,omitempty"`
	CostTrend     string        `gorm:"default:'stable'" json:"costTrend"`
	MOQ           int           `gorm:"default:1" json:"moq"`
	LeadTimeDays  int           `json:"leadTimeDays"`
	LastQuoteDate *time.Time    `json:"lastQuoteDate,omitempty"`
	LastPODate    *time.Time    `json:"lastPoDate,omitempty"`
	Reliability   float64       `json:"reliability"`
}

type GoodsReceipt struct {
	Base
	FXRate           float64            `json:"fxRate"`
	GRNumber         string             `gorm:"uniqueIndex;not null" json:"grNumber"`
	POID             uuid.UUID          `gorm:"type:uuid;not null;index" json:"poId"`
	PO               *PurchaseOrder     `gorm:"foreignKey:POID" json:"po,omitempty"`
	SupplierName     string             `json:"supplierName"`
	ReceiveDate      time.Time          `gorm:"not null" json:"receiveDate"`
	Items            []GoodsReceiptItem `gorm:"foreignKey:GoodsReceiptID" json:"items,omitempty"`
	TotalItems       int                `json:"totalItems"`
	TotalLandingCost float64            `json:"totalLandingCost"`
	ReceivedByID     *uuid.UUID         `gorm:"type:uuid" json:"receivedById,omitempty"`
	ReceivedBy       *User              `gorm:"foreignKey:ReceivedByID" json:"receivedBy,omitempty"`
	Notes            string             `json:"notes"`
}

type GoodsReceiptItem struct {
	Base
	GoodsReceiptID  uuid.UUID      `gorm:"type:uuid;not null;index" json:"goodsReceiptId"`
	POItemID        uuid.UUID      `gorm:"type:uuid;not null" json:"poItemId"`
	ProductID       uuid.UUID      `gorm:"type:uuid;not null" json:"productId"`
	Product         *Product       `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	ReceivedQty     int            `gorm:"not null" json:"receivedQty"`
	UnitCost        float64        `json:"unitCost"`
	ShippingAlloc   float64        `json:"shippingAlloc"`
	CustomsAlloc    float64        `json:"customsAlloc"`
	LandingCost     float64        `json:"landingCost"`
	StorageLocation string         `json:"storageLocation"`
	Condition       GoodsCondition `gorm:"default:'good'" json:"condition"`
	Notes           string         `json:"notes"`
}
