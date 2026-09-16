package models

import "github.com/google/uuid"

type CustomerStatus string
type CustomerType string
type Sector string

const (
	CustomerStatusActive   CustomerStatus = "active"
	CustomerStatusInactive CustomerStatus = "inactive"
	CustomerStatusProspect CustomerStatus = "prospect"

	CustomerTypeGet  CustomerType = "get"
	CustomerTypeGrow CustomerType = "grow"

	SectorGovernment  Sector = "government"
	SectorHealthcare  Sector = "healthcare"
	SectorEducation   Sector = "education"
	SectorRetail      Sector = "retail"
	SectorBanking     Sector = "banking"
	SectorOilGas      Sector = "oil-gas"
	SectorTelecom     Sector = "telecom"
	SectorHospitality Sector = "hospitality"
	SectorRealEstate  Sector = "real-estate"
	SectorOther       Sector = "other"
)

type Customer struct {
	Base
	CompanyName string         `gorm:"not null" json:"companyName"`
	Sector      Sector         `gorm:"not null" json:"sector"`
	Region      string         `gorm:"not null" json:"region"`
	Status      CustomerStatus `gorm:"not null;default:'prospect'" json:"status"`
	Type        CustomerType   `gorm:"not null;default:'get'" json:"type"`
	CRNumber    string         `gorm:"uniqueIndex" json:"crNumber"`
	VATNumber   string         `json:"vatNumber"`
	Notes       string         `json:"notes"`
	CreatedByID *uuid.UUID     `gorm:"type:uuid" json:"createdById,omitempty"`
	CreatedBy   *User          `gorm:"foreignKey:CreatedByID" json:"createdBy,omitempty"`
	Sites       []CustomerSite    `gorm:"foreignKey:CustomerID" json:"sites,omitempty"`
	Contacts    []CustomerContact `gorm:"foreignKey:CustomerID" json:"contacts,omitempty"`
}

type CustomerSite struct {
	Base
	CustomerID uuid.UUID         `gorm:"type:uuid;not null;index" json:"customerId"`
	Name       string            `gorm:"not null" json:"name"`
	Address    string            `json:"address"`
	City       string            `json:"city"`
	Region     string            `json:"region"`
	Contacts   []CustomerContact `gorm:"foreignKey:SiteID" json:"contacts,omitempty"`
}

type CustomerContact struct {
	Base
	CustomerID uuid.UUID  `gorm:"type:uuid;not null;index" json:"customerId"`
	SiteID     *uuid.UUID `gorm:"type:uuid" json:"siteId,omitempty"`
	Name       string     `gorm:"not null" json:"name"`
	Email      string     `json:"email"`
	Phone      string     `json:"phone"`
	Position   string     `json:"position"`
	IsPrimary  bool       `gorm:"default:false" json:"isPrimary"`
}
