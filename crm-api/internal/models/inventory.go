package models

import (
	"time"

	"github.com/google/uuid"
)

type WarehouseLocation string
type ReservationSource string
type ReservationStatus string
type MovementType string

const (
	WarehouseRiyadh  WarehouseLocation = "riyadh-main"
	WarehouseJeddah  WarehouseLocation = "jeddah-branch"
	WarehouseDammam  WarehouseLocation = "dammam-branch"

	ReservationSourceQuote   ReservationSource = "quote"
	ReservationSourceProject ReservationSource = "project"
	ReservationSourceManual  ReservationSource = "manual"

	ReservationStatusActive    ReservationStatus = "active"
	ReservationStatusReleased  ReservationStatus = "released"
	ReservationStatusFulfilled ReservationStatus = "fulfilled"

	MovementTransfer   MovementType = "transfer"
	MovementAdjustment MovementType = "adjustment"
	MovementAllocation MovementType = "allocation"
	MovementReceipt    MovementType = "receipt"
	MovementRelease    MovementType = "release"
	MovementWriteOff   MovementType = "write-off"
)

type WarehouseStock struct {
	Base
	ProductID         uuid.UUID         `gorm:"type:uuid;not null;uniqueIndex:idx_stock_product_location" json:"productId"`
	Product           *Product          `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	WarehouseLocation WarehouseLocation `gorm:"not null;uniqueIndex:idx_stock_product_location" json:"warehouseLocation"`
	OnHandQty         int               `gorm:"default:0;check:on_hand_qty >= 0" json:"onHandQty"`
	ReservedQty       int               `gorm:"default:0;check:reserved_qty >= 0" json:"reservedQty"`
	AvailableQty      int               `json:"availableQty"`
	UnitCost          float64           `json:"unitCost"`
	TotalValue        float64           `json:"totalValue"`
	ReorderLevel      *int              `json:"reorderLevel,omitempty"`
}

func (w *WarehouseStock) Recalculate() {
	w.AvailableQty = w.OnHandQty - w.ReservedQty
	w.TotalValue = float64(w.OnHandQty) * w.UnitCost
}

type StockReservation struct {
	Base
	ProductID         uuid.UUID         `gorm:"type:uuid;not null;index" json:"productId"`
	Product           *Product          `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	WarehouseLocation WarehouseLocation `gorm:"not null" json:"warehouseLocation"`
	Qty               int               `gorm:"not null" json:"qty"`
	Source            ReservationSource `gorm:"not null" json:"source"`
	SourceRef         string            `json:"sourceRef"`
	SourceLabel       string            `json:"sourceLabel"`
	CustomerName      string            `json:"customerName"`
	ReservedByID      *uuid.UUID        `gorm:"type:uuid" json:"reservedById,omitempty"`
	ReservedBy        *User             `gorm:"foreignKey:ReservedByID" json:"reservedBy,omitempty"`
	ReservedAt        time.Time         `json:"reservedAt"`
	ReleaseDate       *time.Time        `json:"releaseDate,omitempty"`
	ReleasedByID      *uuid.UUID        `gorm:"type:uuid" json:"releasedById,omitempty"`
	ReleaseReason     string            `json:"releaseReason"`
	Status            ReservationStatus `gorm:"default:'active'" json:"status"`
	Notes             string            `json:"notes"`
}

type InventoryMovement struct {
	Base
	ProductID         uuid.UUID          `gorm:"type:uuid;not null;index" json:"productId"`
	Product           *Product           `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	MovementType      MovementType       `gorm:"not null" json:"movementType"`
	Qty               int                `gorm:"not null" json:"qty"`
	FromWarehouse     *WarehouseLocation `json:"fromWarehouse,omitempty"`
	ToWarehouse       *WarehouseLocation `json:"toWarehouse,omitempty"`
	Reference         string             `json:"reference"`
	Reason            string             `json:"reason"`
	PerformedByID     *uuid.UUID         `gorm:"type:uuid" json:"performedById,omitempty"`
	PerformedBy       *User              `gorm:"foreignKey:PerformedByID" json:"performedBy,omitempty"`
	PerformedAt       time.Time          `json:"performedAt"`
	Notes             string             `json:"notes"`
}
