package handlers

import (
	"time"

	"g4s-crm/api/internal/middleware"
	"g4s-crm/api/internal/models"
	"g4s-crm/api/pkg/pagination"
	"g4s-crm/api/pkg/response"
	"g4s-crm/api/pkg/seqgen"
	v "g4s-crm/api/pkg/validator"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProcurementHandler struct{ db *gorm.DB }

func NewProcurementHandler(db *gorm.DB) *ProcurementHandler { return &ProcurementHandler{db: db} }

// ─── Purchase Orders ───────────────────────────────────────

func (h *ProcurementHandler) ListPOs(c *gin.Context) {
	params := pagination.GetParams(c)
	var items []models.PurchaseOrder
	var total int64
	query := h.db.Model(&models.PurchaseOrder{}).Preload("Items.Product")
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)
	query.Order(params.Sort + " " + params.Order).Scopes(pagination.Paginate(params)).Find(&items)
	response.OKWithMeta(c, items, pagination.BuildMeta(params, total))
}

func (h *ProcurementHandler) GetPO(c *gin.Context) {
	var item models.PurchaseOrder
	if err := h.db.Preload("Items.Product").First(&item, "id = ?", c.Param("id")).Error; err != nil {
		response.NotFound(c, "Purchase order not found")
		return
	}
	response.OK(c, item)
}

func (h *ProcurementHandler) CreatePO(c *gin.Context) {
	var req struct {
		SupplierID   *string          `json:"supplierId"`
		SupplierName string           `json:"supplierName" validate:"required"`
		Currency     models.Currency  `json:"currency"`
		Notes        string           `json:"notes"`
	}
	if !v.BindAndValidate(c, &req) {
		return
	}
	poNumber, _ := seqgen.NextNumber(h.db, "purchase_order")
	userID := middleware.GetCurrentUserID(c)
	po := &models.PurchaseOrder{
		PONumber:     poNumber,
		SupplierName: req.SupplierName,
		Currency:     req.Currency,
		Notes:        req.Notes,
		CreatedByID:  &userID,
	}
	h.db.Create(po)
	response.Created(c, po)
}

func (h *ProcurementHandler) ApprovePO(c *gin.Context) {
	var po models.PurchaseOrder
	if err := h.db.First(&po, "id = ?", c.Param("id")).Error; err != nil {
		response.NotFound(c, "PO not found")
		return
	}
	if po.Status != models.POStatusPendingApproval {
		response.UnprocessableEntity(c, "PO must be pending approval")
		return
	}
	userID := middleware.GetCurrentUserID(c)
	now := time.Now()
	h.db.Model(&po).Updates(map[string]interface{}{
		"status": models.POStatusApproved, "approved_by_id": userID, "approved_at": now,
	})
	response.OK(c, po)
}

func (h *ProcurementHandler) UpdatePOStatus(c *gin.Context) {
	var req struct {
		Status models.PurchaseOrderStatus `json:"status" validate:"required"`
	}
	if !v.BindAndValidate(c, &req) {
		return
	}
	h.db.Model(&models.PurchaseOrder{}).Where("id = ?", c.Param("id")).Update("status", req.Status)
	response.OK(c, gin.H{"status": req.Status})
}

// ─── Supplier Quotes ───────────────────────────────────────

func (h *ProcurementHandler) ListSQs(c *gin.Context) {
	params := pagination.GetParams(c)
	var items []models.SupplierQuote
	var total int64
	query := h.db.Model(&models.SupplierQuote{}).Preload("Items")
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)
	query.Order(params.Sort + " " + params.Order).Scopes(pagination.Paginate(params)).Find(&items)
	response.OKWithMeta(c, items, pagination.BuildMeta(params, total))
}

func (h *ProcurementHandler) GetSQ(c *gin.Context) {
	var item models.SupplierQuote
	if err := h.db.Preload("Items.Product").First(&item, "id = ?", c.Param("id")).Error; err != nil {
		response.NotFound(c, "Supplier quote not found")
		return
	}
	response.OK(c, item)
}

func (h *ProcurementHandler) CreateSQ(c *gin.Context) {
	var req struct {
		SupplierName string          `json:"supplierName" validate:"required"`
		Currency     models.Currency `json:"currency"`
		Notes        string          `json:"notes"`
	}
	if !v.BindAndValidate(c, &req) {
		return
	}
	sqNumber, _ := seqgen.NextNumber(h.db, "supplier_quote")
	sq := &models.SupplierQuote{
		SQNumber:     sqNumber,
		SupplierName: req.SupplierName,
		Currency:     req.Currency,
		Notes:        req.Notes,
		Status:       models.SQStatusReceived,
	}
	h.db.Create(sq)
	response.Created(c, sq)
}

// ConvertToPO creates a PO from an accepted supplier quote
func (h *ProcurementHandler) ConvertToPO(c *gin.Context) {
	var sq models.SupplierQuote
	if err := h.db.Preload("Items").First(&sq, "id = ?", c.Param("id")).Error; err != nil {
		response.NotFound(c, "Supplier quote not found")
		return
	}
	if sq.Status != models.SQStatusAccepted {
		response.UnprocessableEntity(c, "Supplier quote must be accepted before converting to PO")
		return
	}
	poNumber, _ := seqgen.NextNumber(h.db, "purchase_order")
	userID := middleware.GetCurrentUserID(c)
	po := &models.PurchaseOrder{
		PONumber:     poNumber,
		SupplierID:   sq.SupplierID,
		SupplierName: sq.SupplierName,
		Currency:     sq.Currency,
		CreatedByID:  &userID,
		Status:       models.POStatusDraft,
	}
	for _, item := range sq.Items {
		if item.ProductID != nil {
			po.Items = append(po.Items, models.PurchaseOrderItem{
				ProductID:    *item.ProductID,
				Quantity:     item.Quantity,
				UnitCost:     item.UnitCost,
				Total:        item.Total,
				LeadTimeDays: item.LeadTimeDays,
			})
		}
	}
	h.db.Create(po)
	response.Created(c, po)
}

// ─── Goods Receipts ────────────────────────────────────────

func (h *ProcurementHandler) ListGRs(c *gin.Context) {
	params := pagination.GetParams(c)
	var items []models.GoodsReceipt
	var total int64
	h.db.Model(&models.GoodsReceipt{}).Count(&total)
	h.db.Preload("Items.Product").Order(params.Sort+" "+params.Order).
		Scopes(pagination.Paginate(params)).Find(&items)
	response.OKWithMeta(c, items, pagination.BuildMeta(params, total))
}

func (h *ProcurementHandler) GetGR(c *gin.Context) {
	var item models.GoodsReceipt
	if err := h.db.Preload("Items.Product").First(&item, "id = ?", c.Param("id")).Error; err != nil {
		response.NotFound(c, "Goods receipt not found")
		return
	}
	response.OK(c, item)
}

func (h *ProcurementHandler) CreateGR(c *gin.Context) {
	var req struct {
		POID        string                   `json:"poId" validate:"required"`
		ReceiveDate string                   `json:"receiveDate" validate:"required"`
		Notes       string                   `json:"notes"`
		Items       []struct {
			POItemID        string                `json:"poItemId" validate:"required"`
			ProductID       string                `json:"productId" validate:"required"`
			ReceivedQty     int                   `json:"receivedQty" validate:"required,gt=0"`
			UnitCost        float64               `json:"unitCost"`
			StorageLocation string                `json:"storageLocation"`
			Condition       models.GoodsCondition `json:"condition"`
		} `json:"items" validate:"required,min=1"`
	}
	if !v.BindAndValidate(c, &req) {
		return
	}

	grNumber, _ := seqgen.NextNumber(h.db, "goods_receipt")
	userID := middleware.GetCurrentUserID(c)

	receiveDate, _ := time.Parse("2006-01-02", req.ReceiveDate)
	gr := &models.GoodsReceipt{
		GRNumber:     grNumber,
		SupplierName: "",
		ReceiveDate:  receiveDate,
		Notes:        req.Notes,
		ReceivedByID: &userID,
	}

	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(gr).Error; err != nil {
			return err
		}
		for _, item := range req.Items {
			condition := item.Condition
			if condition == "" {
				condition = models.ConditionGood
			}
			gri := &models.GoodsReceiptItem{
				GoodsReceiptID:  gr.ID,
				ReceivedQty:     item.ReceivedQty,
				UnitCost:        item.UnitCost,
				StorageLocation: item.StorageLocation,
				Condition:       condition,
			}
			if err := tx.Create(gri).Error; err != nil {
				return err
			}
			// Update stock
			tx.Model(&models.WarehouseStock{}).
				Where("product_id = ?", item.ProductID).
				Update("on_hand_qty", gorm.Expr("on_hand_qty + ?", item.ReceivedQty))
			// Record price history
			tx.Create(&models.ProductPriceRecord{
				Date:      receiveDate,
				Source:    models.PriceSourceGoodsReceipt,
				SourceRef: gr.GRNumber,
				UnitCost:  item.UnitCost,
				Qty:       item.ReceivedQty,
			})
		}
		gr.TotalItems = len(req.Items)
		return tx.Save(gr).Error
	})

	if err != nil {
		response.InternalError(c, "Failed to create goods receipt")
		return
	}
	response.Created(c, gr)
}
