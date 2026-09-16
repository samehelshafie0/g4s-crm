package handlers

import (
	"g4s-crm/api/internal/middleware"
	"g4s-crm/api/internal/models"
	"g4s-crm/api/pkg/pagination"
	"g4s-crm/api/pkg/response"
	"g4s-crm/api/pkg/seqgen"
	v "g4s-crm/api/pkg/validator"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type QuoteHandler struct{ db *gorm.DB }

func NewQuoteHandler(db *gorm.DB) *QuoteHandler { return &QuoteHandler{db: db} }

func (h *QuoteHandler) List(c *gin.Context) {
	params := pagination.GetParams(c)
	var quotes []models.Quote
	var total int64

	query := h.db.Model(&models.Quote{}).Preload("Customer")
	if q := c.Query("q"); q != "" {
		query = query.Where("quote_number ILIKE ?", "%"+q+"%")
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if cid := c.Query("customerId"); cid != "" {
		query = query.Where("customer_id = ?", cid)
	}

	query.Count(&total)
	query.Order(params.Sort + " " + params.Order).Scopes(pagination.Paginate(params)).Find(&quotes)
	response.OKWithMeta(c, quotes, pagination.BuildMeta(params, total))
}

func (h *QuoteHandler) Get(c *gin.Context) {
	var quote models.Quote
	if err := h.db.Preload("LineItems.Product").Preload("Customer").Preload("Opportunity").
		First(&quote, "id = ?", c.Param("id")).Error; err != nil {
		response.NotFound(c, "Quote not found")
		return
	}
	response.OK(c, quote)
}

func (h *QuoteHandler) Create(c *gin.Context) {
	var req struct {
		OpportunityID   *string                    `json:"opportunityId"`
		CustomerID      string                     `json:"customerId" validate:"required"`
		Currency        models.Currency            `json:"currency"`
		ValidUntil      *string                    `json:"validUntil"`
		DiscountPercent float64                    `json:"discountPercent"`
		VATPercent      float64                    `json:"vatPercent"`
		Notes           string                     `json:"notes"`
		LineItems       []struct {
			Category    models.QuoteLineCategory `json:"category" validate:"required"`
			ProductID   *string                  `json:"productId"`
			SKU         string                   `json:"sku"`
			Description string                   `json:"description" validate:"required"`
			Quantity    int                      `json:"quantity" validate:"required,gt=0"`
			UnitCost    float64                  `json:"unitCost"`
			UnitPrice   float64                  `json:"unitPrice"`
		} `json:"lineItems"`
	}
	if !v.BindAndValidate(c, &req) {
		return
	}

	quoteNumber, err := seqgen.NextNumber(h.db, "quote")
	if err != nil {
		response.InternalError(c, "Failed to generate quote number")
		return
	}

	vatPct := req.VATPercent
	if vatPct == 0 {
		vatPct = 15
	}
	currency := req.Currency
	if currency == "" {
		currency = models.CurrencySAR
	}

	userID := middleware.GetCurrentUserID(c)
	quote := &models.Quote{
		QuoteNumber:     quoteNumber,
		Status:          models.QuoteStatusDraft,
		Currency:        currency,
		DiscountPercent: req.DiscountPercent,
		VATPercent:      vatPct,
		Notes:           req.Notes,
		CreatedByID:     &userID,
	}

	if err := h.db.Create(quote).Error; err != nil {
		response.InternalError(c, "Failed to create quote")
		return
	}

	for i, item := range req.LineItems {
		li := &models.QuoteLineItem{
			QuoteID:     quote.ID,
			Category:    item.Category,
			Description: item.Description,
			Quantity:    item.Quantity,
			UnitCost:    item.UnitCost,
			UnitPrice:   item.UnitPrice,
			SortOrder:   i,
		}
		li.Calculate()
		h.db.Create(li)
		quote.LineItems = append(quote.LineItems, *li)
	}

	quote.Recalculate()
	h.db.Save(quote)

	response.Created(c, quote)
}

func (h *QuoteHandler) Update(c *gin.Context) {
	var quote models.Quote
	if err := h.db.First(&quote, "id = ?", c.Param("id")).Error; err != nil {
		response.NotFound(c, "Quote not found")
		return
	}
	if quote.Status != models.QuoteStatusDraft {
		response.UnprocessableEntity(c, "Only draft quotes can be edited")
		return
	}
	var req map[string]interface{}
	c.ShouldBindJSON(&req)
	h.db.Model(&quote).Updates(req)
	response.OK(c, quote)
}

func (h *QuoteHandler) Delete(c *gin.Context) {
	var quote models.Quote
	if err := h.db.First(&quote, "id = ?", c.Param("id")).Error; err != nil {
		response.NotFound(c, "Quote not found")
		return
	}
	if quote.Status != models.QuoteStatusDraft {
		response.UnprocessableEntity(c, "Only draft quotes can be deleted")
		return
	}
	h.db.Delete(&quote)
	response.NoContent(c)
}

func (h *QuoteHandler) Submit(c *gin.Context) {
	h.changeStatus(c, models.QuoteStatusDraft, models.QuoteStatusPendingApproval)
}

func (h *QuoteHandler) Approve(c *gin.Context) {
	var quote models.Quote
	if err := h.db.First(&quote, "id = ?", c.Param("id")).Error; err != nil {
		response.NotFound(c, "Quote not found")
		return
	}
	if quote.Status != models.QuoteStatusPendingApproval {
		response.UnprocessableEntity(c, "Quote must be pending approval to approve")
		return
	}
	userID := middleware.GetCurrentUserID(c)
	h.db.Model(&quote).Updates(map[string]interface{}{
		"status": models.QuoteStatusApproved, "approved_by_id": userID,
	})
	response.OK(c, quote)
}

func (h *QuoteHandler) Reject(c *gin.Context) {
	h.changeStatus(c, models.QuoteStatusPendingApproval, models.QuoteStatusDraft)
}

func (h *QuoteHandler) Send(c *gin.Context) {
	h.changeStatus(c, models.QuoteStatusApproved, models.QuoteStatusSent)
}

func (h *QuoteHandler) Accept(c *gin.Context) {
	h.changeStatus(c, models.QuoteStatusSent, models.QuoteStatusAccepted)
}

func (h *QuoteHandler) Decline(c *gin.Context) {
	h.changeStatus(c, models.QuoteStatusSent, models.QuoteStatusDeclined)
}

func (h *QuoteHandler) Recalculate(c *gin.Context) {
	var quote models.Quote
	if err := h.db.Preload("LineItems").First(&quote, "id = ?", c.Param("id")).Error; err != nil {
		response.NotFound(c, "Quote not found")
		return
	}
	for i := range quote.LineItems {
		quote.LineItems[i].Calculate()
		h.db.Save(&quote.LineItems[i])
	}
	quote.Recalculate()
	h.db.Save(&quote)
	response.OK(c, quote)
}

func (h *QuoteHandler) changeStatus(c *gin.Context, from, to models.QuoteStatus) {
	var quote models.Quote
	if err := h.db.First(&quote, "id = ?", c.Param("id")).Error; err != nil {
		response.NotFound(c, "Quote not found")
		return
	}
	if quote.Status != from {
		response.UnprocessableEntity(c, "Invalid status transition")
		return
	}
	h.db.Model(&quote).Update("status", to)
	response.OK(c, quote)
}
