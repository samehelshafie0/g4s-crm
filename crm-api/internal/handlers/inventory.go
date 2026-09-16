package handlers

import (
	"g4s-crm/api/internal/middleware"
	"g4s-crm/api/internal/models"
	"g4s-crm/api/pkg/pagination"
	"g4s-crm/api/pkg/response"
	v "g4s-crm/api/pkg/validator"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type InventoryHandler struct{ db *gorm.DB }

func NewInventoryHandler(db *gorm.DB) *InventoryHandler { return &InventoryHandler{db: db} }

const warehouseRule = "required,oneof=riyadh-main jeddah-branch dammam-branch"

// All stock writers lock the product first, including absent warehouse rows.
// This also prevents opposite-direction transfers from acquiring row locks in
// opposite orders. Multi-product callers acquire product IDs in sorted order.
func lockProduct(tx *gorm.DB, id uuid.UUID) error {
	var p models.Product
	return tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&p, "id = ?", id).Error
}
func stockRow(tx *gorm.DB, id uuid.UUID, location models.WarehouseLocation) (*models.WarehouseStock, error) {
	item := models.WarehouseStock{ProductID: id, WarehouseLocation: location}
	if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "product_id"}, {Name: "warehouse_location"}}, DoNothing: true}).Create(&item).Error; err != nil {
		return nil, err
	}
	item.Base = models.Base{}
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&item, "product_id = ? AND warehouse_location = ?", id, location).Error
	return &item, err
}
func saveStock(tx *gorm.DB, stock *models.WarehouseStock) error {
	if stock.OnHandQty < 0 || stock.ReservedQty < 0 || stock.ReservedQty > stock.OnHandQty {
		return invalid("Insufficient available stock")
	}
	stock.Recalculate()
	return tx.Omit(clause.Associations).Save(stock).Error
}
func stockMovement(tx *gorm.DB, c *gin.Context, m *models.InventoryMovement) error {
	user := middleware.GetCurrentUserID(c)
	m.PerformedByID = &user
	m.PerformedAt = time.Now()
	return tx.Create(m).Error
}
func (h *InventoryHandler) ListStock(c *gin.Context) {
	params := pagination.GetParams(c)
	items := []models.WarehouseStock{}
	var total int64
	q := h.db.Model(&models.WarehouseStock{})
	loc := c.Param("location")
	if loc == "" {
		loc = c.Query("warehouseLocation")
	}
	if loc != "" {
		q = q.Where("warehouse_location = ?", loc)
	}
	pid := c.Param("productId")
	if pid == "" {
		pid = c.Query("productId")
	}
	if pid != "" {
		q = q.Where("product_id = ?", pid)
	}
	if err := q.Count(&total).Error; err != nil {
		apiError(c, err)
		return
	}
	if err := q.Preload("Product").Order("created_at DESC").Scopes(pagination.Paginate(params)).Find(&items).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OKWithMeta(c, items, pagination.BuildMeta(params, total))
}
func (h *InventoryHandler) LowStock(c *gin.Context) {
	items := []models.WarehouseStock{}
	if err := h.db.Preload("Product").Where("reorder_level IS NOT NULL AND available_qty <= reorder_level").Find(&items).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OK(c, items)
}
func (h *InventoryHandler) ListReservations(c *gin.Context) {
	items := []models.StockReservation{}
	q := h.db.Preload("Product").Preload("ReservedBy")
	for _, field := range []string{"status", "product_id"} {
		key := field
		if field == "product_id" {
			key = "productId"
		}
		if value := c.Query(key); value != "" {
			q = q.Where(field+" = ?", value)
		}
	}
	if err := q.Order("reserved_at DESC").Find(&items).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OK(c, items)
}
func (h *InventoryHandler) CreateReservation(c *gin.Context) {
	item := models.StockReservation{Source: models.ReservationSourceManual, Status: models.ReservationStatusActive, ReservedAt: time.Now()}
	if !bindFields(c, &item, map[string]string{"productId": "required", "warehouseLocation": warehouseRule, "qty": "required,gt=0,lte=100000000", "source": "required,oneof=quote project manual", "sourceRef": "max=200", "sourceLabel": "max=200", "customerName": "max=200", "notes": "max=5000"}) {
		return
	}
	user := middleware.GetCurrentUserID(c)
	item.ReservedByID = &user
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := lockProduct(tx, item.ProductID); err != nil {
			return err
		}
		stock, err := stockRow(tx, item.ProductID, item.WarehouseLocation)
		if err != nil {
			return err
		}
		stock.ReservedQty += item.Qty
		if err := saveStock(tx, stock); err != nil {
			return err
		}
		if err := tx.Create(&item).Error; err != nil {
			return err
		}
		location := item.WarehouseLocation
		return stockMovement(tx, c, &models.InventoryMovement{ProductID: item.ProductID, MovementType: models.MovementAllocation, Qty: item.Qty, FromWarehouse: &location, Reference: item.ID.String(), Reason: "Stock reserved"})
	})
	if err != nil {
		apiError(c, err)
		return
	}
	response.Created(c, item)
}
func (h *InventoryHandler) Release(c *gin.Context) { h.finishReservation(c, false) }
func (h *InventoryHandler) Fulfill(c *gin.Context) { h.finishReservation(c, true) }
func (h *InventoryHandler) finishReservation(c *gin.Context, fulfill bool) {
	var req struct {
		Reason string `json:"reason" validate:"required,max=1000"`
	}
	if !v.BindStrict(c, &req) {
		return
	}
	var item models.StockReservation
	err := h.db.Transaction(func(tx *gorm.DB) error {
		// Read only the immutable product ID before taking locks in global order.
		if err := tx.First(&item, "id = ?", c.Param("id")).Error; err != nil {
			return err
		}
		if err := lockProduct(tx, item.ProductID); err != nil {
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&item, "id = ?", item.ID).Error; err != nil {
			return err
		}
		if item.Status != models.ReservationStatusActive {
			return invalid("Reservation is no longer active")
		}
		stock, err := stockRow(tx, item.ProductID, item.WarehouseLocation)
		if err != nil {
			return err
		}
		stock.ReservedQty -= item.Qty
		item.Status = models.ReservationStatusReleased
		kind := models.MovementRelease
		if fulfill {
			stock.OnHandQty -= item.Qty
			item.Status = models.ReservationStatusFulfilled
			kind = models.MovementAllocation
		}
		if err := saveStock(tx, stock); err != nil {
			return err
		}
		now := time.Now()
		user := middleware.GetCurrentUserID(c)
		item.ReleaseDate = &now
		item.ReleasedByID = &user
		item.ReleaseReason = req.Reason
		if err := tx.Omit(clause.Associations).Save(&item).Error; err != nil {
			return err
		}
		loc := item.WarehouseLocation
		return stockMovement(tx, c, &models.InventoryMovement{ProductID: item.ProductID, MovementType: kind, Qty: item.Qty, FromWarehouse: &loc, Reference: item.ID.String(), Reason: req.Reason})
	})
	if err != nil {
		apiError(c, err)
		return
	}
	response.OK(c, item)
}
func (h *InventoryHandler) Transfer(c *gin.Context) {
	var req struct {
		ProductID     uuid.UUID                `json:"productId"`
		Qty           int                      `json:"qty"`
		FromWarehouse models.WarehouseLocation `json:"fromWarehouse"`
		ToWarehouse   models.WarehouseLocation `json:"toWarehouse"`
		Reason        string                   `json:"reason"`
		Notes         string                   `json:"notes"`
	}
	if !bindFields(c, &req, map[string]string{"productId": "required", "qty": "required,gt=0,lte=100000000", "fromWarehouse": warehouseRule, "toWarehouse": warehouseRule, "reason": "required,max=1000", "notes": "max=5000"}) {
		return
	}
	if req.FromWarehouse == req.ToWarehouse {
		response.BadRequest(c, "Choose different warehouses")
		return
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := lockProduct(tx, req.ProductID); err != nil {
			return err
		}
		from, err := stockRow(tx, req.ProductID, req.FromWarehouse)
		if err != nil {
			return err
		}
		to, err := stockRow(tx, req.ProductID, req.ToWarehouse)
		if err != nil {
			return err
		}
		from.OnHandQty -= req.Qty
		if err := saveStock(tx, from); err != nil {
			return err
		}
		to.UnitCost = (float64(to.OnHandQty)*to.UnitCost + float64(req.Qty)*from.UnitCost) / float64(to.OnHandQty+req.Qty)
		to.OnHandQty += req.Qty
		if err := saveStock(tx, to); err != nil {
			return err
		}
		return stockMovement(tx, c, &models.InventoryMovement{ProductID: req.ProductID, MovementType: models.MovementTransfer, Qty: req.Qty, FromWarehouse: &req.FromWarehouse, ToWarehouse: &req.ToWarehouse, Reason: req.Reason, Notes: req.Notes})
	})
	if err != nil {
		apiError(c, err)
		return
	}
	response.OK(c, gin.H{"message": "Transfer completed"})
}
func (h *InventoryHandler) Adjust(c *gin.Context) {
	var req struct {
		ProductID         uuid.UUID                `json:"productId"`
		WarehouseLocation models.WarehouseLocation `json:"warehouseLocation"`
		Qty               int                      `json:"qty"`
		UnitCost          *float64                 `json:"unitCost"`
		Reason            string                   `json:"reason"`
	}
	if !bindFields(c, &req, map[string]string{"productId": "required", "warehouseLocation": warehouseRule, "qty": "required,gte=-100000000,lte=100000000", "unitCost": "omitempty," + moneyRule, "reason": "required,max=1000"}) {
		return
	}
	var stock *models.WarehouseStock
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := lockProduct(tx, req.ProductID); err != nil {
			return err
		}
		var err error
		stock, err = stockRow(tx, req.ProductID, req.WarehouseLocation)
		if err != nil {
			return err
		}
		if req.UnitCost != nil && req.Qty > 0 {
			stock.UnitCost = (float64(stock.OnHandQty)*stock.UnitCost + float64(req.Qty)**req.UnitCost) / float64(stock.OnHandQty+req.Qty)
		}
		stock.OnHandQty += req.Qty
		if err := saveStock(tx, stock); err != nil {
			return err
		}
		return stockMovement(tx, c, &models.InventoryMovement{ProductID: req.ProductID, MovementType: models.MovementAdjustment, Qty: req.Qty, ToWarehouse: &req.WarehouseLocation, Reason: req.Reason})
	})
	if err != nil {
		apiError(c, err)
		return
	}
	response.OK(c, stock)
}
func (h *InventoryHandler) ReorderLevel(c *gin.Context) {
	var req struct {
		ReorderLevel *int `json:"reorderLevel" validate:"omitempty,gte=0"`
	}
	if !v.BindStrict(c, &req) {
		return
	}
	result := h.db.Model(&models.WarehouseStock{}).Where("id = ?", c.Param("id")).Update("reorder_level", req.ReorderLevel)
	if result.Error != nil {
		apiError(c, result.Error)
		return
	}
	if result.RowsAffected != 1 {
		response.NotFound(c, "Stock record not found")
		return
	}
	response.OK(c, req)
}
func (h *InventoryHandler) ListMovements(c *gin.Context) {
	params := pagination.GetParams(c)
	items := []models.InventoryMovement{}
	var total int64
	q := h.db.Model(&models.InventoryMovement{})
	if id := c.Query("productId"); id != "" {
		q = q.Where("product_id = ?", id)
	}
	if err := q.Count(&total).Error; err != nil {
		apiError(c, err)
		return
	}
	if err := q.Preload("Product").Preload("PerformedBy").Order("performed_at DESC").Scopes(pagination.Paginate(params)).Find(&items).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OKWithMeta(c, items, pagination.BuildMeta(params, total))
}
