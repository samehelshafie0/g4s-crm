package models

import (
	"github.com/google/uuid"
	"time"
)

// Append-only price evidence. Supplier editors replace live line IDs; records
// survive those edits and soft deletion of the original supplier quotation.
type SupplierPriceRecord struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	RecordedAt          time.Time  `json:"recordedAt"`
	SupplierQuoteID     uuid.UUID  `json:"supplierQuoteId"`
	SourceItemID        uuid.UUID  `json:"sourceItemId"`
	Revision            int        `json:"revision"`
	ProductID           uuid.UUID  `json:"productId"`
	SKU                 string     `json:"sku"`
	ProductName         string     `json:"productName"`
	SupplierName        string     `json:"supplierName"`
	SupplierRef         string     `json:"supplierRef"`
	SupplierQuoteNumber string     `json:"supplierQuoteNumber"`
	Currency            Currency   `json:"currency"`
	UnitCost            float64    `json:"unitCost"`
	Quantity            int        `json:"quantity"`
	ValidUntil          *time.Time `json:"validUntil,omitempty"`
	SourceQuoteID       *uuid.UUID `json:"sourceQuoteId,omitempty"`
}
