package handlers

import (
	"github.com/google/uuid"
	"gorm.io/gorm/clause"
	"sort"
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
	query := h.db.Model(&models.PurchaseOrder{}).Preload("Items.Product.Manufacturer").Preload("ApprovedBy")
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)
	query.Order(params.Sort + " " + params.Order).Scopes(pagination.Paginate(params)).Find(&items)
	response.OKWithMeta(c, items, pagination.BuildMeta(params, total))
}

func (h *ProcurementHandler) GetPO(c *gin.Context) {
	var item models.PurchaseOrder
	if err := h.db.Preload("Items.Product.Manufacturer").Preload("ApprovedBy").First(&item, "id = ?", c.Param("id")).Error; err != nil {
		response.NotFound(c, "Purchase order not found")
		return
	}
	response.OK(c, item)
}

type poLineInput struct {
	ProductID    uuid.UUID `json:"productId" validate:"required"`
	Quantity     int       `json:"quantity" validate:"required,gt=0,lte=100000000"`
	UnitCost     float64   `json:"unitCost" validate:"gte=0,lte=100000000"`
	LeadTimeDays int       `json:"leadTimeDays" validate:"gte=0,lte=3650"`
}
type poInput struct {
	SupplierID       *uuid.UUID      `json:"supplierId"`
	SupplierName     string          `json:"supplierName" validate:"required,max=255"`
	Currency         models.Currency `json:"currency" validate:"required,oneof=SAR USD EUR GBP AED CNY"`
	ShippingCost     float64         `json:"shippingCost" validate:"gte=0,lte=100000000"`
	CustomsDuty      float64         `json:"customsDuty" validate:"gte=0,lte=100000000"`
	ExpectedDelivery *string         `json:"expectedDelivery"`
	SourceQuoteID    *uuid.UUID      `json:"sourceQuoteId"`
	Notes            string          `json:"notes" validate:"max=20000"`
	Items            []poLineInput   `json:"items" validate:"max=1000,dive"`
}

func (h *ProcurementHandler) CreatePO(c *gin.Context) { h.savePO(c, true) }
func (h *ProcurementHandler) UpdatePO(c *gin.Context) { h.savePO(c, false) }

// PUT-like PATCH replaces the editable header and all draft lines atomically.
func (h *ProcurementHandler) savePO(c *gin.Context, create bool) {
	req := poInput{Currency: models.CurrencySAR}
	if !v.BindStrict(c, &req) {
		return
	}
	date, err := parseDate(req.ExpectedDelivery)
	if err != nil {
		apiError(c, err)
		return
	}
	var po models.PurchaseOrder
	err = h.db.Transaction(func(tx *gorm.DB) error {
		if !create {
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&po, "id = ?", c.Param("id")).Error; err != nil {
				return err
			}
			if po.Status != models.POStatusDraft {
				return invalid("Only draft purchase orders can be edited")
			}
		}
		if err := optionalExists(tx, &models.Manufacturer{}, req.SupplierID); err != nil {
			return err
		}
		if err := optionalExists(tx, &models.Quote{}, req.SourceQuoteID); err != nil {
			return err
		}
		if create {
			number, err := seqgen.NextNumber(tx, "purchase_order")
			if err != nil {
				return err
			}
			user := middleware.GetCurrentUserID(c)
			po.PONumber = number
			po.CreatedByID = &user
			po.Status = models.POStatusDraft
		}
		po.SupplierID = req.SupplierID
		po.SupplierName = req.SupplierName
		po.Currency = req.Currency
		po.ShippingCost = req.ShippingCost
		po.CustomsDuty = req.CustomsDuty
		po.ExpectedDelivery = date
		po.SourceQuoteID = req.SourceQuoteID
		po.Notes = req.Notes
		po.Subtotal = 0
		po.Items = nil
		for _, line := range req.Items {
			if err := exists(tx, &models.Product{}, line.ProductID); err != nil {
				return err
			}
			total := models.RoundMoney(float64(line.Quantity) * line.UnitCost)
			po.Subtotal += total
			po.Items = append(po.Items, models.PurchaseOrderItem{ProductID: line.ProductID, Quantity: line.Quantity, UnitCost: line.UnitCost, Total: total, LeadTimeDays: line.LeadTimeDays})
		}
		po.Subtotal = models.RoundMoney(po.Subtotal)
		po.Total = models.RoundMoney(po.Subtotal + po.ShippingCost + po.CustomsDuty)
		if err := tx.Omit(clause.Associations).Save(&po).Error; err != nil {
			return err
		}
		if !create {
			if err := tx.Unscoped().Where("purchase_order_id = ?", po.ID).Delete(&models.PurchaseOrderItem{}).Error; err != nil {
				return err
			}
		}
		for i := range po.Items {
			po.Items[i].PurchaseOrderID = po.ID
			if err := tx.Create(&po.Items[i]).Error; err != nil {
				return err
			}
		}
		return recordActivity(tx, c, "purchase-order", po.ID, "saved")
	})
	if err != nil {
		apiError(c, err)
		return
	}
	if err := h.db.Preload("Items.Product.Manufacturer").Preload("ApprovedBy").First(&po, "id = ?", po.ID).Error; err != nil {
		apiError(c, err)
		return
	}
	if create {
		response.Created(c, po)
	} else {
		response.OK(c, po)
	}
}
func (h *ProcurementHandler) DeletePO(c *gin.Context) {
	result := h.db.Where("status = ?", models.POStatusDraft).Delete(&models.PurchaseOrder{}, "id = ?", c.Param("id"))
	if result.Error != nil {
		apiError(c, result.Error)
		return
	}
	if result.RowsAffected != 1 {
		response.Conflict(c, "Only an existing draft PO can be deleted")
		return
	}
	response.NoContent(c)
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
	result := h.db.Model(&po).Where("status = ?", models.POStatusPendingApproval).Updates(map[string]interface{}{
		"status": models.POStatusApproved, "approved_by_id": userID, "approved_at": now,
	})
	if result.Error != nil {
		response.InternalError(c, "Failed to approve PO")
		return
	}
	if result.RowsAffected != 1 {
		response.Conflict(c, "PO changed; reload before approving")
		return
	}
	h.GetPO(c)
}

func (h *ProcurementHandler) UpdatePOStatus(c *gin.Context) {
	var req struct {
		Status models.PurchaseOrderStatus `json:"status" validate:"required,oneof=pending-approval ordered cancelled"`
	}
	if !v.BindStrict(c, &req) {
		return
	}
	var allowed []models.PurchaseOrderStatus
	switch req.Status {
	case models.POStatusPendingApproval:
		allowed = []models.PurchaseOrderStatus{models.POStatusDraft}
	case models.POStatusOrdered:
		allowed = []models.PurchaseOrderStatus{models.POStatusApproved}
	case models.POStatusCancelled:
		allowed = []models.PurchaseOrderStatus{models.POStatusDraft, models.POStatusPendingApproval, models.POStatusApproved}
	}
	result := h.db.Model(&models.PurchaseOrder{}).Where("id = ? AND status IN ?", c.Param("id"), allowed).Update("status", req.Status)
	if result.Error != nil {
		response.InternalError(c, "Failed to update PO")
		return
	}
	if result.RowsAffected != 1 {
		response.Conflict(c, "PO is missing or this transition is not allowed")
		return
	}
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

type sqLineInput struct {
	ProductID        *uuid.UUID `json:"productId"`
	ProductSKU       string     `json:"productSku" validate:"max=100"`
	ProductName      string     `json:"productName" validate:"required,max=255"`
	ManufacturerName string     `json:"manufacturerName" validate:"max=255"`
	Quantity         int        `json:"quantity" validate:"required,gt=0,lte=100000000"`
	UnitCost         float64    `json:"unitCost" validate:"gte=0,lte=100000000"`
	LeadTimeDays     int        `json:"leadTimeDays" validate:"gte=0,lte=3650"`
	MOQ              int        `json:"moq" validate:"gte=0,lte=100000000"`
	ValidUntil       *string    `json:"validUntil"`
	Notes            string     `json:"notes" validate:"max=5000"`
}

func (h *ProcurementHandler) CreateSQ(c *gin.Context) { h.saveSQ(c, true) }
func (h *ProcurementHandler) UpdateSQ(c *gin.Context) { h.saveSQ(c, false) }
func (h *ProcurementHandler) saveSQ(c *gin.Context, create bool) {
	var req struct {
		SupplierID    *uuid.UUID      `json:"supplierId"`
		SupplierName  string          `json:"supplierName" validate:"required,max=255"`
		SupplierRef   string          `json:"supplierRef" validate:"max=255"`
		Currency      models.Currency `json:"currency" validate:"required,oneof=SAR USD EUR GBP AED CNY"`
		ValidFrom     *string         `json:"validFrom"`
		ValidUntil    *string         `json:"validUntil"`
		ContactName   string          `json:"contactName" validate:"max=255"`
		ContactEmail  string          `json:"contactEmail" validate:"omitempty,email"`
		PaymentTerms  string          `json:"paymentTerms" validate:"max=5000"`
		DeliveryTerms string          `json:"deliveryTerms" validate:"max=5000"`
		Notes         string          `json:"notes" validate:"max=20000"`
		Items         []sqLineInput   `json:"items" validate:"max=1000,dive"`
	}
	req.Currency = models.CurrencyUSD
	if !v.BindStrict(c, &req) {
		return
	}
	from, err := parseDate(req.ValidFrom)
	if err != nil {
		apiError(c, err)
		return
	}
	until, err := parseDate(req.ValidUntil)
	if err != nil {
		apiError(c, err)
		return
	}
	if from != nil && until != nil && until.Before(*from) {
		response.BadRequest(c, "Invalid validity dates")
		return
	}
	var sq models.SupplierQuote
	err = h.db.Transaction(func(tx *gorm.DB) error {
		if !create {
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&sq, "id = ?", c.Param("id")).Error; err != nil {
				return err
			}
			if sq.Status != models.SQStatusReceived && sq.Status != models.SQStatusUnderReview {
				return invalid("Only received or under-review quotes can be edited")
			}
		}
		if err := optionalExists(tx, &models.Manufacturer{}, req.SupplierID); err != nil {
			return err
		}
		if create {
			number, err := seqgen.NextNumber(tx, "supplier_quote")
			if err != nil {
				return err
			}
			sq.SQNumber = number
			sq.Status = models.SQStatusReceived
		}
		sq.SupplierID = req.SupplierID
		sq.SupplierName = req.SupplierName
		sq.SupplierRef = req.SupplierRef
		sq.Currency = req.Currency
		sq.ValidFrom = from
		sq.ValidUntil = until
		sq.ContactName = req.ContactName
		sq.ContactEmail = req.ContactEmail
		sq.PaymentTerms = req.PaymentTerms
		sq.DeliveryTerms = req.DeliveryTerms
		sq.Notes = req.Notes
		sq.Subtotal = 0
		sq.Items = nil
		for _, line := range req.Items {
			if err := optionalExists(tx, &models.Product{}, line.ProductID); err != nil {
				return err
			}
			valid, err := parseDate(line.ValidUntil)
			if err != nil {
				return err
			}
			total := models.RoundMoney(float64(line.Quantity) * line.UnitCost)
			sq.Subtotal += total
			sq.Items = append(sq.Items, models.SupplierQuoteItem{ProductID: line.ProductID, ProductSKU: line.ProductSKU, ProductName: line.ProductName, ManufacturerName: line.ManufacturerName, Quantity: line.Quantity, UnitCost: line.UnitCost, Total: total, LeadTimeDays: line.LeadTimeDays, MOQ: line.MOQ, ValidUntil: valid, Notes: line.Notes})
		}
		sq.Subtotal = models.RoundMoney(sq.Subtotal)
		if err := tx.Omit(clause.Associations).Save(&sq).Error; err != nil {
			return err
		}
		if !create {
			if err := tx.Unscoped().Where("supplier_quote_id = ?", sq.ID).Delete(&models.SupplierQuoteItem{}).Error; err != nil {
				return err
			}
		}
		for i := range sq.Items {
			sq.Items[i].SupplierQuoteID = sq.ID
			if err := tx.Create(&sq.Items[i]).Error; err != nil {
				return err
			}
		}
		return recordActivity(tx, c, "supplier-quote", sq.ID, "saved")
	})
	if err != nil {
		apiError(c, err)
		return
	}
	if create {
		response.Created(c, sq)
	} else {
		response.OK(c, sq)
	}
}
func (h *ProcurementHandler) UpdateSQStatus(c *gin.Context) {
	var req struct {
		Status models.SupplierQuoteStatus `json:"status" validate:"required,oneof=under-review accepted rejected expired"`
	}
	if !v.BindStrict(c, &req) {
		return
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		var sq models.SupplierQuote
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Items").First(&sq, "id = ?", c.Param("id")).Error; err != nil {
			return err
		}
		if sq.Status != models.SQStatusReceived && sq.Status != models.SQStatusUnderReview {
			return invalid("Supplier quote is already finalized")
		}
		if req.Status == models.SQStatusAccepted {
			if len(sq.Items) == 0 {
				return invalid("Add items before accepting")
			}
			if sq.ValidUntil != nil && sq.ValidUntil.Before(time.Now().Truncate(24*time.Hour)) {
				return invalid("Supplier quote has expired")
			}
		}
		if err := tx.Model(&sq).Update("status", req.Status).Error; err != nil {
			return err
		}
		return recordActivity(tx, c, "supplier-quote", sq.ID, string(req.Status))
	})
	if err != nil {
		apiError(c, err)
		return
	}
	h.GetSQ(c)
}
func (h *ProcurementHandler) DeleteSQ(c *gin.Context) {
	result := h.db.Where("status IN ?", []models.SupplierQuoteStatus{models.SQStatusReceived, models.SQStatusRejected, models.SQStatusExpired}).Delete(&models.SupplierQuote{}, "id = ?", c.Param("id"))
	if result.Error != nil {
		apiError(c, result.Error)
		return
	}
	if result.RowsAffected != 1 {
		response.Conflict(c, "Supplier quote is missing or cannot be deleted")
		return
	}
	response.NoContent(c)
}

func (h *ProcurementHandler) ConvertToPO(c *gin.Context) {
	var po models.PurchaseOrder
	created := false
	err := h.db.Transaction(func(tx *gorm.DB) error {
		var sq models.SupplierQuote
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Items").First(&sq, "id = ?", c.Param("id")).Error; err != nil {
			return err
		}
		if sq.Status != models.SQStatusAccepted {
			return invalid("Supplier quote must be accepted")
		}
		result := tx.Where("supplier_quote_id = ?", sq.ID).First(&po)
		if result.Error == nil {
			return nil
		}
		if result.Error != gorm.ErrRecordNotFound {
			return result.Error
		}
		if len(sq.Items) == 0 {
			return invalid("Supplier quote has no items")
		}
		number, err := seqgen.NextNumber(tx, "purchase_order")
		if err != nil {
			return err
		}
		user := middleware.GetCurrentUserID(c)
		po = models.PurchaseOrder{PONumber: number, SupplierQuoteID: &sq.ID, SupplierID: sq.SupplierID, SupplierName: sq.SupplierName, Currency: sq.Currency, CreatedByID: &user, Status: models.POStatusDraft, Subtotal: sq.Subtotal, Total: sq.Subtotal}
		for _, line := range sq.Items {
			if line.ProductID == nil {
				return invalid("Link every supplier quote item to a catalog product before conversion")
			}
			if err := exists(tx, &models.Product{}, *line.ProductID); err != nil {
				return err
			}
			po.Items = append(po.Items, models.PurchaseOrderItem{ProductID: *line.ProductID, Quantity: line.Quantity, UnitCost: line.UnitCost, Total: line.Total, LeadTimeDays: line.LeadTimeDays})
		}
		if err := tx.Create(&po).Error; err != nil {
			return err
		}
		created = true
		return recordActivity(tx, c, "purchase-order", po.ID, "converted")
	})
	if err != nil {
		apiError(c, err)
		return
	}
	if err := h.db.Preload("Items.Product.Manufacturer").Preload("ApprovedBy").First(&po, "id = ?", po.ID).Error; err != nil {
		apiError(c, err)
		return
	}
	if created {
		response.Created(c, po)
	} else {
		response.OK(c, po)
	}
}

func (h *ProcurementHandler) ListGRs(c *gin.Context) {
	params := pagination.GetParams(c)
	var items []models.GoodsReceipt
	var total int64
	h.db.Model(&models.GoodsReceipt{}).Count(&total)
	h.db.Preload("Items.Product").Preload("PO").Preload("ReceivedBy").Order(params.Sort + " " + params.Order).
		Scopes(pagination.Paginate(params)).Find(&items)
	response.OKWithMeta(c, items, pagination.BuildMeta(params, total))
}

func (h *ProcurementHandler) GetGR(c *gin.Context) {
	var item models.GoodsReceipt
	if err := h.db.Preload("Items.Product").Preload("PO").Preload("ReceivedBy").First(&item, "id = ?", c.Param("id")).Error; err != nil {
		response.NotFound(c, "Goods receipt not found")
		return
	}
	response.OK(c, item)
}

func (h *ProcurementHandler) CreateGR(c *gin.Context) {
	var req struct {
		POID        uuid.UUID `json:"poId" validate:"required"`
		ReceiveDate string    `json:"receiveDate" validate:"required,datetime=2006-01-02"`
		Notes       string    `json:"notes" validate:"max=20000"`
		Items       []struct {
			POItemID        uuid.UUID                `json:"poItemId" validate:"required"`
			ProductID       uuid.UUID                `json:"productId" validate:"required"`
			ReceivedQty     int                      `json:"receivedQty" validate:"required,gt=0,lte=100000000"`
			StorageLocation models.WarehouseLocation `json:"storageLocation" validate:"required,oneof=riyadh-main jeddah-branch dammam-branch"`
			Condition       models.GoodsCondition    `json:"condition" validate:"required,oneof=good damaged partial-damage"`
			Notes           string                   `json:"notes" validate:"max=5000"`
		} `json:"items" validate:"required,min=1,max=1000,dive"`
	}
	if !v.BindStrict(c, &req) {
		return
	}
	date, _ := time.Parse("2006-01-02", req.ReceiveDate)
	user := middleware.GetCurrentUserID(c)
	gr := models.GoodsReceipt{POID: req.POID, ReceiveDate: date, Notes: req.Notes, ReceivedByID: &user}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		var po models.PurchaseOrder
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Items").First(&po, "id = ?", req.POID).Error; err != nil {
			return err
		}
		if po.Status != models.POStatusOrdered && po.Status != models.POStatusPartialReceived {
			return invalid("Only ordered or partially received POs can receive goods")
		}
		gr.FXRate = 1
		if po.Currency != models.CurrencySAR {
			var rate models.ExchangeRate
			if err := tx.Where("from_currency = ? AND to_currency = 'SAR'", po.Currency).First(&rate).Error; err != nil {
				return invalid("Configure an exchange rate to SAR before receiving this currency")
			}
			var history models.ExchangeRateHistory
			err := tx.Where("exchange_rate_id = ? AND effective_date <= ?", rate.ID, date).Order("effective_date DESC, created_at DESC").First(&history).Error
			if err == nil {
				gr.FXRate = history.Rate
			} else if err == gorm.ErrRecordNotFound && !rate.EffectiveDate.After(date) {
				gr.FXRate = rate.CurrentRate
			} else {
				return invalid("No exchange rate is effective on this receipt date")
			}
		}
		// Deterministic product lock ordering also protects receipts for different POs.
		ids := map[string]uuid.UUID{}
		for _, line := range req.Items {
			ids[line.ProductID.String()] = line.ProductID
		}
		keys := []string{}
		for key := range ids {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			if err := lockProduct(tx, ids[key]); err != nil {
				return err
			}
		}
		number, err := seqgen.NextNumber(tx, "goods_receipt")
		if err != nil {
			return err
		}
		gr.GRNumber = number
		gr.SupplierName = po.SupplierName
		if err := tx.Create(&gr).Error; err != nil {
			return err
		}
		seen := map[uuid.UUID]bool{}
		for _, line := range req.Items {
			if seen[line.POItemID] {
				return invalid("A PO item may appear only once per receipt")
			}
			seen[line.POItemID] = true
			var item *models.PurchaseOrderItem
			for i := range po.Items {
				if po.Items[i].ID == line.POItemID {
					item = &po.Items[i]
					break
				}
			}
			if item == nil || item.ProductID != line.ProductID {
				return invalid("Receipt item does not match the purchase order")
			}
			if line.ReceivedQty > item.Quantity-item.ReceivedQty {
				return invalid("Received quantity exceeds outstanding quantity")
			}
			shipping, customs := 0.0, 0.0
			if po.Subtotal > 0 {
				ratio := float64(line.ReceivedQty) * item.UnitCost / po.Subtotal
				shipping = models.RoundMoney(po.ShippingCost * ratio * gr.FXRate)
				customs = models.RoundMoney(po.CustomsDuty * ratio * gr.FXRate)
			}
			gri := models.GoodsReceiptItem{GoodsReceiptID: gr.ID, POItemID: item.ID, ProductID: item.ProductID, ReceivedQty: line.ReceivedQty, UnitCost: item.UnitCost, ShippingAlloc: shipping, CustomsAlloc: customs, LandingCost: models.RoundMoney(float64(line.ReceivedQty)*item.UnitCost*gr.FXRate + shipping + customs), StorageLocation: string(line.StorageLocation), Condition: line.Condition, Notes: line.Notes}
			if err := tx.Create(&gri).Error; err != nil {
				return err
			}
			gr.Items = append(gr.Items, gri)
			gr.TotalLandingCost += gri.LandingCost
			item.ReceivedQty += line.ReceivedQty
			if err := tx.Omit(clause.Associations).Save(item).Error; err != nil {
				return err
			}
			// Damaged deliveries remain in the receipt for reconciliation, not usable stock.
			if line.Condition == models.ConditionGood {
				stock, err := stockRow(tx, item.ProductID, line.StorageLocation)
				if err != nil {
					return err
				}
				cost := gri.LandingCost / float64(gri.ReceivedQty)
				stock.UnitCost = (float64(stock.OnHandQty)*stock.UnitCost + gri.LandingCost) / float64(stock.OnHandQty+line.ReceivedQty)
				stock.OnHandQty += line.ReceivedQty
				if err := saveStock(tx, stock); err != nil {
					return err
				}
				if err := stockMovement(tx, c, &models.InventoryMovement{ProductID: item.ProductID, MovementType: models.MovementReceipt, Qty: line.ReceivedQty, ToWarehouse: &line.StorageLocation, Reference: gr.GRNumber, Reason: "Goods receipt"}); err != nil {
					return err
				}
				if err := tx.Create(&models.ProductPriceRecord{ProductID: item.ProductID, Date: date, Currency: models.CurrencySAR, LandingCost: cost, Source: models.PriceSourceGoodsReceipt, SourceRef: gr.GRNumber, UnitCost: cost, Qty: line.ReceivedQty}).Error; err != nil {
					return err
				}
			}
		}
		complete := true
		for _, item := range po.Items {
			if item.ReceivedQty < item.Quantity {
				complete = false
			}
		}
		po.Status = models.POStatusPartialReceived
		if complete {
			po.Status = models.POStatusReceived
			po.ActualDelivery = &date
		}
		if err := tx.Omit(clause.Associations).Save(&po).Error; err != nil {
			return err
		}
		gr.TotalItems = len(gr.Items)
		gr.TotalLandingCost = models.RoundMoney(gr.TotalLandingCost)
		if err := tx.Omit(clause.Associations).Save(&gr).Error; err != nil {
			return err
		}
		return recordActivity(tx, c, "goods-receipt", gr.ID, "received")
	})
	if err != nil {
		apiError(c, err)
		return
	}
	response.Created(c, gr)
}

func (h *ProcurementHandler) SupplierItems(c *gin.Context) {
	items := []models.SupplierItemCatalog{}
	q := h.db.Preload("Supplier").Preload("Product.Manufacturer")
	if id := c.Query("productId"); id != "" {
		q = q.Where("product_id = ?", id)
	}
	if id := c.Query("supplierId"); id != "" {
		q = q.Where("supplier_id = ?", id)
	}
	if err := q.Find(&items).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OK(c, items)
}
func (h *ProcurementHandler) SaveSupplierItem(c *gin.Context) {
	item := models.SupplierItemCatalog{MOQ: 1, CostTrend: "stable"}
	if id := c.Param("id"); id != "" {
		if err := h.db.First(&item, "id = ?", id).Error; err != nil {
			apiError(c, err)
			return
		}
	}
	previous := item.LatestCost
	if !bindFields(c, &item, map[string]string{"supplierId": "required", "productId": "required", "latestCost": moneyRule, "moq": "gte=1", "leadTimeDays": "gte=0,lte=3650", "reliability": "gte=0,lte=100"}) {
		return
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := exists(tx, &models.Manufacturer{}, item.SupplierID); err != nil {
			return err
		}
		if err := exists(tx, &models.Product{}, item.ProductID); err != nil {
			return err
		}
		if c.Param("id") != "" && previous != item.LatestCost {
			item.PreviousCost = &previous
			item.CostTrend = "up"
			if item.LatestCost < previous {
				item.CostTrend = "down"
			}
		}
		return tx.Omit(clause.Associations).Save(&item).Error
	})
	if err != nil {
		apiError(c, err)
		return
	}
	response.OK(c, item)
}
func (h *ProcurementHandler) DeleteSupplierItem(c *gin.Context) {
	deleteRecord(c, h.db, &models.SupplierItemCatalog{})
}
