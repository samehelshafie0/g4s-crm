package handlers

import (
	"time"

	"g4s-crm/api/internal/middleware"
	"g4s-crm/api/internal/models"
	"g4s-crm/api/pkg/pagination"
	"g4s-crm/api/pkg/response"
	v "g4s-crm/api/pkg/validator"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type InventoryHandler struct{ db *gorm.DB }

func NewInventoryHandler(db *gorm.DB) *InventoryHandler { return &InventoryHandler{db: db} }

func (h *InventoryHandler) ListStock(c *gin.Context) {
	params := pagination.GetParams(c)
	var items []models.WarehouseStock
	var total int64
	query := h.db.Model(&models.WarehouseStock{}).Preload("Product")
	if loc := c.Param("location"); loc != "" { query = query.Where("warehouse_location = ?", loc) }
	if pid := c.Param("productId"); pid != "" { query = query.Where("product_id = ?", pid) }
	query.Count(&total)
	query.Scopes(pagination.Paginate(params)).Find(&items)
	response.OKWithMeta(c, items, pagination.BuildMeta(params, total))
}

func (h *InventoryHandler) LowStock(c *gin.Context) {
	var items []models.WarehouseStock
	h.db.Preload("Product").
		Where("reorder_level IS NOT NULL AND on_hand_qty <= reorder_level").
		Find(&items)
	response.OK(c, items)
}

func (h *InventoryHandler) CreateReservation(c *gin.Context) {
	var req struct {
		ProductID         string                    `json:"productId" validate:"required"`
		WarehouseLocation models.WarehouseLocation  `json:"warehouseLocation" validate:"required"`
		Qty               int                       `json:"qty" validate:"required,gt=0"`
		Source            models.ReservationSource  `json:"source" validate:"required"`
		SourceRef         string                    `json:"sourceRef"`
		SourceLabel       string                    `json:"sourceLabel"`
		CustomerName      string                    `json:"customerName"`
		Notes             string                    `json:"notes"`
	}
	if !v.BindAndValidate(c, &req) { return }

	var stock models.WarehouseStock
	if err := h.db.Where("product_id = ? AND warehouse_location = ?", req.ProductID, req.WarehouseLocation).
		First(&stock).Error; err != nil {
		response.NotFound(c, "Stock record not found for this product/warehouse")
		return
	}
	if stock.AvailableQty < req.Qty {
		response.UnprocessableEntity(c, "Insufficient available stock")
		return
	}

	userID := middleware.GetCurrentUserID(c)
	reservation := &models.StockReservation{
		WarehouseLocation: req.WarehouseLocation,
		Qty:               req.Qty,
		Source:            req.Source,
		SourceRef:         req.SourceRef,
		SourceLabel:       req.SourceLabel,
		CustomerName:      req.CustomerName,
		ReservedByID:      &userID,
		ReservedAt:        time.Now(),
		Status:            models.ReservationStatusActive,
		Notes:             req.Notes,
	}

	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(reservation).Error; err != nil { return err }
		return tx.Model(&stock).Updates(map[string]interface{}{
			"reserved_qty":  gorm.Expr("reserved_qty + ?", req.Qty),
			"available_qty": gorm.Expr("available_qty - ?", req.Qty),
		}).Error
	})

	if err != nil { response.InternalError(c, "Failed to create reservation"); return }
	response.Created(c, reservation)
}

func (h *InventoryHandler) Transfer(c *gin.Context) {
	var req struct {
		ProductID     string                   `json:"productId" validate:"required"`
		Qty           int                      `json:"qty" validate:"required,gt=0"`
		FromWarehouse models.WarehouseLocation `json:"fromWarehouse" validate:"required"`
		ToWarehouse   models.WarehouseLocation `json:"toWarehouse" validate:"required"`
		Reason        string                   `json:"reason" validate:"required"`
		Notes         string                   `json:"notes"`
	}
	if !v.BindAndValidate(c, &req) { return }

	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.WarehouseStock{}).
			Where("product_id = ? AND warehouse_location = ? AND available_qty >= ?", req.ProductID, req.FromWarehouse, req.Qty).
			Update("on_hand_qty", gorm.Expr("on_hand_qty - ?", req.Qty)).Error; err != nil {
			return err
		}
		// Upsert destination stock
		tx.Model(&models.WarehouseStock{}).
			Where("product_id = ? AND warehouse_location = ?", req.ProductID, req.ToWarehouse).
			Update("on_hand_qty", gorm.Expr("on_hand_qty + ?", req.Qty))
		// Record movement
		userID := middleware.GetCurrentUserID(c)
		from := req.FromWarehouse
		to := req.ToWarehouse
		movement := &models.InventoryMovement{
			MovementType:  models.MovementTransfer,
			Qty:           req.Qty,
			FromWarehouse: &from,
			ToWarehouse:   &to,
			Reason:        req.Reason,
			PerformedByID: &userID,
			PerformedAt:   time.Now(),
			Notes:         req.Notes,
		}
		return tx.Create(movement).Error
	})
	if err != nil { response.InternalError(c, "Transfer failed"); return }
	response.OK(c, gin.H{"message": "Transfer completed"})
}

func (h *InventoryHandler) ListMovements(c *gin.Context) {
	params := pagination.GetParams(c)
	var items []models.InventoryMovement
	var total int64
	h.db.Model(&models.InventoryMovement{}).Count(&total)
	h.db.Preload("Product").Preload("PerformedBy").
		Order("performed_at DESC").Scopes(pagination.Paginate(params)).Find(&items)
	response.OKWithMeta(c, items, pagination.BuildMeta(params, total))
}
