package models

import "github.com/google/uuid"

type VendorType string

const (
	VendorTypeManufacturer VendorType = "manufacturer"
	VendorTypeSupplier     VendorType = "supplier"
	VendorTypeBoth         VendorType = "both"
)

type Manufacturer struct {
	Base
	Name         string                  `gorm:"not null" json:"name"`
	Code         string                  `gorm:"uniqueIndex" json:"code"`
	Country      string                  `json:"country"`
	ContactEmail string                  `json:"contactEmail"`
	ContactPhone string                  `json:"contactPhone"`
	Website      string                  `json:"website"`
	VendorType   VendorType              `json:"vendorType"`
	IsActive     bool                    `gorm:"default:true" json:"isActive"`
	Categories   []ManufacturerCategory  `gorm:"foreignKey:ManufacturerID" json:"categories,omitempty"`
}

type ManufacturerCategory struct {
	Base
	ManufacturerID uuid.UUID `gorm:"type:uuid;not null;index" json:"manufacturerId"`
	Name           string    `gorm:"not null" json:"name"`
	Description    string    `json:"description"`
}
