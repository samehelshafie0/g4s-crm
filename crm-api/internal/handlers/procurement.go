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
	query := h.db.Model(&models.PurchaseOrder{}).Preload("Items.Product.Manufacturer").Preload("ApprovedBy").Preload("CreatedBy").Preload("SourceQuote").Preload("Opportunity")
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)
	query.Order(params.Sort + " " + params.Order).Scopes(pagination.Paginate(params)).Find(&items)
	response.OKWithMeta(c, items, pagination.BuildMeta(params, total))
}

func (h *ProcurementHandler) GetPO(c *gin.Context) {
	var item models.PurchaseOrder
	if err := h.db.Preload("Items.Product.Manufacturer").Preload("ApprovedBy").Preload("CreatedBy").Preload("SourceQuote").Preload("Opportunity").First(&item, "id = ?", c.Param("id")).Error; err != nil {
		response.NotFound(c, "Purchase order not found")
		return
	}
	response.OK(c, item)
}

type poLineInput struct {
	ProductID       uuid.UUID  `json:"productId" validate:"required"`
	Quantity        int        `json:"quantity" validate:"required,gt=0,lte=100000000"`
	UnitCost        float64    `json:"unitCost" validate:"gte=0,lte=100000000"`
	LeadTimeDays    int        `json:"leadTimeDays" validate:"gte=0,lte=3650"`
	QuoteLineItemID *uuid.UUID `json:"quoteLineItemId"`
}
type poInput struct {
	SupplierID       *uuid.UUID      `json:"supplierId"`
	SupplierName     string          `json:"supplierName" validate:"required,max=255"`
	Currency         models.Currency `json:"currency" validate:"required,oneof=SAR USD EUR GBP AED CNY"`
	ShippingCost     float64         `json:"shippingCost" validate:"gte=0,lte=100000000"`
	CustomsDuty      float64         `json:"customsDuty" validate:"gte=0,lte=100000000"`
	ExpectedDelivery *string         `json:"expectedDelivery"`
	SourceQuoteID    *uuid.UUID      `json:"sourceQuoteId"`
	OpportunityID    *uuid.UUID      `json:"opportunityId"`
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
	name, nameErr := vendorName(req.SupplierName)
	if nameErr != nil {
		apiError(c, nameErr)
		return
	}
	if name == "" {
		response.BadRequest(c, "A supplier name is required")
		return
	}
	req.SupplierName = name
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
		if err := optionalExists(tx, &models.Opportunity{}, req.OpportunityID); err != nil {
			return err
		}
		if create {
			user := middleware.GetCurrentUserID(c)
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
		po.OpportunityID = req.OpportunityID
		po.Notes = req.Notes
		po.Subtotal = 0
		po.Items = nil
		opportunityID, accepted, jobLineIDs, err := jobLines(tx, po.OpportunityID, po.SourceQuoteID)
		if err != nil {
			return err
		}
		po.OpportunityID = opportunityID
		for _, line := range req.Items {
			if line.QuoteLineItemID == nil && accepted != nil {
				id, ok := jobLineIDs[line.ProductID]
				if !ok {
					return invalid("Choose a specific accepted quote line for each job product; use a stock PO for extra items")
				}
				line.QuoteLineItemID = &id
			}
			if err := exists(tx, &models.Product{}, line.ProductID); err != nil {
				return err
			}
			total := models.RoundMoney(float64(line.Quantity) * line.UnitCost)
			po.Subtotal += total
			// An order line bought for a job points at the quote line it fulfils,
			// and the order takes that job's opportunity if it has none.
			if line.QuoteLineItemID != nil {
				_, quote, err := committedLine(tx, *line.QuoteLineItemID, line.ProductID)
				if err != nil {
					return err
				}
				if quote.OpportunityID != nil {
					if po.OpportunityID == nil {
						po.OpportunityID = quote.OpportunityID
					} else if *po.OpportunityID != *quote.OpportunityID {
						return invalid("An order line fulfils a quote from a different opportunity")
					}
				}
				if po.SourceQuoteID == nil {
					po.SourceQuoteID = &quote.ID
				}
			}
			po.Items = append(po.Items, models.PurchaseOrderItem{ProductID: line.ProductID, Quantity: line.Quantity, UnitCost: line.UnitCost, Total: total, LeadTimeDays: line.LeadTimeDays, QuoteLineItemID: line.QuoteLineItemID})
		}
		if err := validatePOCommitments(tx, &po); err != nil {
			return err
		}
		if create {
			number, err := seqgen.NextNumber(tx, "purchase_order")
			if err != nil {
				return err
			}
			po.PONumber = number
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
	if err := h.db.Preload("Items.Product.Manufacturer").Preload("ApprovedBy").Preload("CreatedBy").Preload("SourceQuote").Preload("Opportunity").First(&po, "id = ?", po.ID).Error; err != nil {
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
	err := h.db.Transaction(func(tx *gorm.DB) error {
		var po models.PurchaseOrder
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Items").First(&po, "id = ?", c.Param("id")).Error; err != nil {
			return err
		}
		permitted := false
		for _, status := range allowed {
			if po.Status == status {
				permitted = true
			}
		}
		if !permitted {
			return conflict("PO is missing or this transition is not allowed")
		}
		if req.Status != models.POStatusCancelled {
			if err := validatePOCommitments(tx, &po); err != nil {
				return err
			}
		}
		if err := tx.Model(&po).Update("status", req.Status).Error; err != nil {
			return err
		}
		return recordActivity(tx, c, "purchase-order", po.ID, string(req.Status))
	})
	if err != nil {
		apiError(c, err)
		return
	}
	response.OK(c, gin.H{"status": req.Status})
}

// ─── Supplier Quotes ───────────────────────────────────────

func (h *ProcurementHandler) ListSQs(c *gin.Context) {
	params := pagination.GetParams(c)
	var items []models.SupplierQuote
	var total int64
	query := h.db.Model(&models.SupplierQuote{}).Preload("Items").Preload("SourceQuote").Preload("Opportunity")
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)
	query.Order(params.Sort + " " + params.Order).Scopes(pagination.Paginate(params)).Find(&items)
	response.OKWithMeta(c, items, pagination.BuildMeta(params, total))
}

func (h *ProcurementHandler) GetSQ(c *gin.Context) {
	var item models.SupplierQuote
	if err := h.db.Preload("Items.Product").Preload("SourceQuote").Preload("Opportunity").First(&item, "id = ?", c.Param("id")).Error; err != nil {
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
		SourceQuoteID *uuid.UUID      `json:"sourceQuoteId"`
		OpportunityID *uuid.UUID      `json:"opportunityId"`
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
	name, nameErr := vendorName(req.SupplierName)
	if nameErr != nil {
		apiError(c, nameErr)
		return
	}
	if name == "" {
		response.BadRequest(c, "A supplier name is required")
		return
	}
	req.SupplierName = name
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
		if err := optionalExists(tx, &models.Quote{}, req.SourceQuoteID); err != nil {
			return err
		}
		if err := optionalExists(tx, &models.Opportunity{}, req.OpportunityID); err != nil {
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
		sq.SourceQuoteID = req.SourceQuoteID
		sq.OpportunityID = req.OpportunityID
		if sq.OpportunityID == nil && sq.SourceQuoteID != nil {
			var source models.Quote
			if err := tx.First(&source, "id = ?", sq.SourceQuoteID).Error; err != nil {
				return err
			}
			sq.OpportunityID = source.OpportunityID
		}
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
		name, err := vendorName(sq.SupplierName)
		if err != nil {
			return err
		}
		sq.SupplierName = name
		user := middleware.GetCurrentUserID(c)
		po = models.PurchaseOrder{SupplierQuoteID: &sq.ID, SupplierID: sq.SupplierID, SupplierName: sq.SupplierName, Currency: sq.Currency, CreatedByID: &user, Status: models.POStatusDraft, Subtotal: sq.Subtotal, Total: sq.Subtotal, SourceQuoteID: sq.SourceQuoteID, OpportunityID: sq.OpportunityID}
		opportunityID, accepted, lines, err := jobLines(tx, sq.OpportunityID, sq.SourceQuoteID)
		if err != nil {
			return err
		}
		po.OpportunityID = opportunityID
		if accepted != nil {
			po.SourceQuoteID = &accepted.ID
		}
		for _, line := range sq.Items {
			if line.ProductID == nil {
				return invalid("Link every supplier quote item to a catalog product before conversion")
			}
			if err := exists(tx, &models.Product{}, *line.ProductID); err != nil {
				return err
			}
			item := models.PurchaseOrderItem{ProductID: *line.ProductID, Quantity: line.Quantity, UnitCost: line.UnitCost, Total: line.Total, LeadTimeDays: line.LeadTimeDays}
			if lineID, ok := lines[*line.ProductID]; ok {
				item.QuoteLineItemID = &lineID
			} else if accepted != nil {
				return invalid("This product has no unique accepted quote line; create a direct PO and choose the line explicitly")
			}
			po.Items = append(po.Items, item)
		}
		if err := validatePOCommitments(tx, &po); err != nil {
			return err
		}
		number, err := seqgen.NextNumber(tx, "purchase_order")
		if err != nil {
			return err
		}
		po.PONumber = number
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
	if err := h.db.Preload("Items.Product.Manufacturer").Preload("ApprovedBy").Preload("CreatedBy").Preload("SourceQuote").Preload("Opportunity").First(&po, "id = ?", po.ID).Error; err != nil {
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
	h.db.Preload("Items.Product").Preload("PO").Preload("ReplacementPO").Preload("ReceivedBy").Order(params.Sort + " " + params.Order).
		Scopes(pagination.Paginate(params)).Find(&items)
	response.OKWithMeta(c, items, pagination.BuildMeta(params, total))
}

func (h *ProcurementHandler) GetGR(c *gin.Context) {
	var item models.GoodsReceipt
	if err := h.db.Preload("Items.Product").Preload("PO").Preload("ReplacementPO").Preload("ReceivedBy").First(&item, "id = ?", c.Param("id")).Error; err != nil {
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
			DamagedQty      int                      `json:"damagedQty" validate:"gte=0,lte=100000000"`
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
		replacement := models.PurchaseOrder{SupplierID: po.SupplierID, SupplierName: po.SupplierName, Currency: po.Currency, Status: models.POStatusDraft, SourceQuoteID: po.SourceQuoteID, OpportunityID: po.OpportunityID, CreatedByID: &user, Notes: "Replacement for damaged goods on " + gr.GRNumber + "; confirm supplier credit and price before approval."}
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
			damaged := line.DamagedQty
			if line.Condition == models.ConditionDamaged {
				damaged = line.ReceivedQty
			}
			if damaged > line.ReceivedQty || (line.Condition == models.ConditionGood && damaged != 0) || (line.Condition == models.ConditionPartialDamage && (damaged <= 0 || damaged >= line.ReceivedQty)) {
				return invalid("Partial damage requires the exact damaged quantity, less than the received quantity")
			}
			good := line.ReceivedQty - damaged
			returnStatus := "not-required"
			if damaged > 0 {
				returnStatus = "pending"
			}
			shipping, customs := 0.0, 0.0
			if po.Subtotal > 0 {
				ratio := float64(line.ReceivedQty) * item.UnitCost / po.Subtotal
				shipping = models.RoundMoney(po.ShippingCost * ratio * gr.FXRate)
				customs = models.RoundMoney(po.CustomsDuty * ratio * gr.FXRate)
			}
			gri := models.GoodsReceiptItem{DamagedQty: damaged, ReturnStatus: returnStatus, GoodsReceiptID: gr.ID, POItemID: item.ID, ProductID: item.ProductID, ReceivedQty: line.ReceivedQty, UnitCost: item.UnitCost, ShippingAlloc: shipping, CustomsAlloc: customs, LandingCost: models.RoundMoney(float64(line.ReceivedQty)*item.UnitCost*gr.FXRate + shipping + customs), StorageLocation: string(line.StorageLocation), Condition: line.Condition, Notes: line.Notes}
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
			if good > 0 {
				stock, err := stockRow(tx, item.ProductID, line.StorageLocation)
				if err != nil {
					return err
				}
				cost := gri.LandingCost / float64(gri.ReceivedQty)
				stock.UnitCost = (float64(stock.OnHandQty)*stock.UnitCost + cost*float64(good)) / float64(stock.OnHandQty+good)
				stock.OnHandQty += good
				// Goods bought for a job are held for it on arrival, so no other
				// hold can take them before the job does.
				hold, err := jobHold(tx, *item, good, line.StorageLocation)
				if err != nil {
					return err
				}
				if hold != nil {
					stock.ReservedQty += hold.Qty
				}
				if err := saveStock(tx, stock); err != nil {
					return err
				}
				if err := stockMovement(tx, c, &models.InventoryMovement{ProductID: item.ProductID, MovementType: models.MovementReceipt, Qty: good, ToWarehouse: &line.StorageLocation, Reference: gr.GRNumber, Reason: "Goods receipt"}); err != nil {
					return err
				}
				if hold != nil {
					hold.ReservedByID, hold.ReservedAt, hold.Notes = &user, time.Now(), "Held on receipt "+gr.GRNumber
					if err := tx.Create(hold).Error; err != nil {
						return err
					}
					if err := stockMovement(tx, c, &models.InventoryMovement{ProductID: item.ProductID, MovementType: models.MovementAllocation, Qty: hold.Qty, FromWarehouse: &line.StorageLocation, Reference: hold.ID.String(), Reason: "Held for " + hold.SourceRef + " on receipt"}); err != nil {
						return err
					}
				}
				if err := tx.Create(&models.ProductPriceRecord{ProductID: item.ProductID, Date: date, Currency: models.CurrencySAR, LandingCost: cost, Source: models.PriceSourceGoodsReceipt, SourceRef: gr.GRNumber, UnitCost: cost, Qty: good}).Error; err != nil {
					return err
				}
			}
			if damaged > 0 {
				replacement.Items = append(replacement.Items, models.PurchaseOrderItem{ProductID: item.ProductID, Quantity: damaged, UnitCost: item.UnitCost, LeadTimeDays: item.LeadTimeDays, QuoteLineItemID: item.QuoteLineItemID})
			}
		}
		// Receipts have already consumed the original outstanding quantity.
		// Replace only unmet demand for job lines; stock orders replace all damage.
		used := map[uuid.UUID]int{}
		replacements := []models.PurchaseOrderItem{}
		for _, item := range replacement.Items {
			if item.QuoteLineItemID != nil {
				line, _, err := committedLine(tx, *item.QuoteLineItemID, item.ProductID)
				if err != nil {
					return err
				}
				remaining, err := lineRemaining(tx, line, uuid.Nil)
				if err != nil {
					return err
				}
				item.Quantity = min(item.Quantity, max(0, remaining-used[line.ID]))
				used[line.ID] += item.Quantity
			}
			if item.Quantity == 0 {
				continue
			}
			item.Total = models.RoundMoney(float64(item.Quantity) * item.UnitCost)
			replacement.Subtotal += item.Total
			replacements = append(replacements, item)
		}
		replacement.Items = replacements
		if len(replacement.Items) > 0 {
			number, err := seqgen.NextNumber(tx, "purchase_order")
			if err != nil {
				return err
			}
			replacement.PONumber = number
			replacement.Total = models.RoundMoney(replacement.Subtotal)
			replacement.Subtotal = replacement.Total
			if err := tx.Create(&replacement).Error; err != nil {
				return err
			}
			gr.ReplacementPOID = &replacement.ID
			if err := recordActivity(tx, c, "purchase-order", replacement.ID, "replacement draft for "+gr.GRNumber); err != nil {
				return err
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
