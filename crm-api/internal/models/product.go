package models

import (
	"time"

	"github.com/google/uuid"
)

type ProductType string
type Currency string
type PriceSource string
type ProductDocType string

const (
	ProductTypeImport ProductType = "import"
	ProductTypeLocal  ProductType = "local"

	CurrencySAR Currency = "SAR"
	CurrencyUSD Currency = "USD"
	CurrencyEUR Currency = "EUR"
	CurrencyGBP Currency = "GBP"
	CurrencyAED Currency = "AED"
	CurrencyCNY Currency = "CNY"

	PriceSourceVendorCatalog  PriceSource = "vendor-catalog"
	PriceSourcePurchaseOrder  PriceSource = "purchase-order"
	PriceSourceSupplierQuote  PriceSource = "supplier-quote"
	PriceSourceGoodsReceipt   PriceSource = "goods-receipt"
	PriceSourceManual         PriceSource = "manual"

	DocTypeDatasheet   ProductDocType = "datasheet"
	DocTypeManual      ProductDocType = "manual"
	DocTypeCertificate ProductDocType = "certificate"
	DocTypeVendorQuote ProductDocType = "vendor-quote"
	DocTypeCatalog     ProductDocType = "catalog"
	DocTypeImage       ProductDocType = "image"
	DocTypeOther       ProductDocType = "other"
)

type Product struct {
	Base
	SKU                  string                `gorm:"uniqueIndex;not null" json:"sku"`
	Name                 string                `gorm:"not null" json:"name"`
	Description          string                `json:"description"`
	ManufacturerID       *uuid.UUID            `gorm:"type:uuid" json:"manufacturerId,omitempty"`
	Manufacturer         *Manufacturer         `gorm:"foreignKey:ManufacturerID" json:"manufacturer,omitempty"`
	CategoryID           *uuid.UUID            `gorm:"type:uuid" json:"categoryId,omitempty"`
	Category             *ManufacturerCategory `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	ProductType          ProductType           `gorm:"not null" json:"productType"`
	OriginCurrency       Currency              `gorm:"not null;default:'USD'" json:"originCurrency"`
	UnitCostOrigin       float64               `json:"unitCostOrigin"`
	FXRate               float64               `gorm:"default:1" json:"fxRate"`
	CostInSAR            float64               `json:"costInSAR"`
	FreightPercent       float64               `gorm:"default:0" json:"freightPercent"`
	CustomsPercent       float64               `gorm:"default:0" json:"customsPercent"`
	ClearancePercent     float64               `gorm:"default:0" json:"clearancePercent"`
	LandedCostSAR        float64               `json:"landedCostSAR"`
	TargetMarginPercent  float64               `json:"targetMarginPercent"`
	SellingPrice         float64               `json:"sellingPrice"`
	MarginAmount         float64               `json:"marginAmount"`
	LeadTimeDays         int                   `json:"leadTimeDays"`
	SupplierName         string                `json:"supplierName"`
	IsActive             bool                  `gorm:"default:true" json:"isActive"`
	VendorEntries        []ProductVendorEntry  `gorm:"foreignKey:ProductID" json:"vendorEntries,omitempty"`
	PriceHistory         []ProductPriceRecord  `gorm:"foreignKey:ProductID" json:"priceHistory,omitempty"`
	Documents            []ProductDocument     `gorm:"foreignKey:ProductID" json:"documents,omitempty"`
}

func (p *Product) RecalculateCosts() {
	p.CostInSAR = p.UnitCostOrigin * p.FXRate
	freightAmt := p.CostInSAR * p.FreightPercent / 100
	customsAmt := p.CostInSAR * p.CustomsPercent / 100
	clearanceAmt := p.CostInSAR * p.ClearancePercent / 100
	p.LandedCostSAR = p.CostInSAR + freightAmt + customsAmt + clearanceAmt
	p.MarginAmount = p.SellingPrice - p.LandedCostSAR
}

type ProductVendorEntry struct {
	Base
	ProductID    uuid.UUID `gorm:"type:uuid;not null;index" json:"productId"`
	VendorName   string    `gorm:"not null" json:"vendorName"`
	VendorSKU    string    `json:"vendorSku"`
	UnitCost     float64   `json:"unitCost"`
	Currency     Currency  `json:"currency"`
	MOQ          int       `gorm:"default:1" json:"moq"`
	LeadTimeDays int       `json:"leadTimeDays"`
	LastQuoteDate *time.Time `json:"lastQuoteDate,omitempty"`
	CatalogSource string   `json:"catalogSource"`
	Notes        string    `json:"notes"`
}

type ProductPriceRecord struct {
	Base
	ProductID   uuid.UUID   `gorm:"type:uuid;not null;index" json:"productId"`
	Date        time.Time   `gorm:"not null" json:"date"`
	Source      PriceSource `json:"source"`
	SourceRef   string      `json:"sourceRef"`
	VendorName  string      `json:"vendorName"`
	UnitCost    float64     `json:"unitCost"`
	Currency    Currency    `json:"currency"`
	LandingCost float64     `json:"landingCost"`
	Qty         int         `json:"qty"`
	Notes       string      `json:"notes"`
}

type ProductDocument struct {
	Base
	ProductID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"productId"`
	Name        string         `gorm:"not null" json:"name"`
	DocType     ProductDocType `gorm:"not null" json:"docType"`
	FileName    string         `json:"fileName"`
	FileSize    int64          `json:"fileSize"`
	FilePath    string         `json:"filePath"`
	UploadedByID *uuid.UUID    `gorm:"type:uuid" json:"uploadedById,omitempty"`
	UploadedBy  *User          `gorm:"foreignKey:UploadedByID" json:"uploadedBy,omitempty"`
	Notes       string         `json:"notes"`
}
