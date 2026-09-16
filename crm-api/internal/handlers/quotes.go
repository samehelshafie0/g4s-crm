package handlers

import (
	"errors"
	"g4s-crm/api/internal/middleware"
	"g4s-crm/api/internal/models"
	"g4s-crm/api/pkg/pagination"
	"g4s-crm/api/pkg/response"
	"g4s-crm/api/pkg/seqgen"
	v "g4s-crm/api/pkg/validator"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type QuoteHandler struct{ db *gorm.DB }

func NewQuoteHandler(db *gorm.DB) *QuoteHandler { return &QuoteHandler{db: db} }

func (h *QuoteHandler) List(c *gin.Context) {
	params := pagination.GetParams(c, "quote_number", "status", "total", "valid_until")
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

type quoteLineRequest struct {
	Category    models.QuoteLineCategory `json:"category" validate:"required,oneof=materials manpower miscellaneous"`
	ProductID   *uuid.UUID               `json:"productId"`
	SKU         string                   `json:"sku"`
	Description string                   `json:"description" validate:"required,max=2000"`
	Quantity    int                      `json:"quantity" validate:"gt=0,lte=1000000"`
	UnitCost    float64                  `json:"unitCost" validate:"gte=0,lte=1000000000"`
	UnitPrice   float64                  `json:"unitPrice" validate:"gte=0,lte=1000000000"`
}
type createQuoteRequest struct {
	CustomerID      uuid.UUID          `json:"customerId" validate:"required"`
	OpportunityID   *uuid.UUID         `json:"opportunityId"`
	Currency        models.Currency    `json:"currency" validate:"omitempty,oneof=SAR USD EUR GBP AED CNY"`
	ValidUntil      *string            `json:"validUntil" validate:"omitempty,datetime=2006-01-02"`
	DiscountPercent float64            `json:"discountPercent" validate:"gte=0,lte=100"`
	VATPercent      *float64           `json:"vatPercent" validate:"omitempty,gte=0,lte=100"`
	Notes           string             `json:"notes" validate:"max=20000"`
	LineItems       []quoteLineRequest `json:"lineItems" validate:"max=1000,dive"`
}
type updateQuoteRequest struct {
	Notes           *string  `json:"notes" validate:"omitempty,max=20000"`
	ValidUntil      *string  `json:"validUntil" validate:"omitempty,datetime=2006-01-02"`
	DiscountPercent *float64 `json:"discountPercent" validate:"omitempty,gte=0,lte=100"`
	VATPercent      *float64 `json:"vatPercent" validate:"omitempty,gte=0,lte=100"`
}

func (h *QuoteHandler) Create(c *gin.Context) {
	var req createQuoteRequest
	if !v.BindStrict(c, &req) {
		return
	}
	var quote models.Quote
	err := h.db.Transaction(func(tx *gorm.DB) error {
		var customer models.Customer
		if err := tx.First(&customer, "id = ?", req.CustomerID).Error; err != nil {
			return errInvalidRelationship
		}
		if req.OpportunityID != nil {
			var opportunity models.Opportunity
			if err := tx.First(&opportunity, "id = ? AND customer_id = ?", req.OpportunityID, req.CustomerID).Error; err != nil {
				return errInvalidRelationship
			}
		}
		number, err := seqgen.NextNumber(tx, "quote")
		if err != nil {
			return err
		}
		userID := middleware.GetCurrentUserID(c)
		quote = models.Quote{QuoteNumber: number, CustomerID: req.CustomerID, OpportunityID: req.OpportunityID, Status: models.QuoteStatusDraft, Currency: req.Currency, DiscountPercent: req.DiscountPercent, VATPercent: 15, Notes: req.Notes, CreatedByID: &userID}
		if quote.Currency == "" {
			quote.Currency = models.CurrencySAR
		}
		if req.VATPercent != nil {
			quote.VATPercent = *req.VATPercent
		}
		if req.ValidUntil != nil {
			date, _ := time.Parse("2006-01-02", *req.ValidUntil)
			quote.ValidUntil = &date
		}
		for index, input := range req.LineItems {
			if input.ProductID != nil {
				var product models.Product
				if err := tx.First(&product, "id = ?", input.ProductID).Error; err != nil {
					return errInvalidRelationship
				}
			}
			line := models.QuoteLineItem{Category: input.Category, ProductID: input.ProductID, SKU: input.SKU, Description: input.Description, Quantity: input.Quantity, UnitCost: input.UnitCost, UnitPrice: input.UnitPrice, SortOrder: index}
			line.Calculate()
			quote.LineItems = append(quote.LineItems, line)
		}
		quote.Recalculate()
		if err := tx.Create(&quote).Error; err != nil {
			return err
		}
		// GORM defaults otherwise replace an explicit zero VAT with 15.
		if req.VATPercent != nil && *req.VATPercent == 0 {
			quote.VATPercent = 0
			return tx.Model(&quote).Update("vat_percent", 0).Error
		}
		return nil
	})
	if errors.Is(err, errInvalidRelationship) {
		response.BadRequest(c, "Customer, opportunity or product does not exist or does not belong to this quote")
		return
	}
	if err != nil {
		response.InternalError(c, "Failed to create quote")
		return
	}
	response.Created(c, quote)
}

var errInvalidRelationship = errors.New("invalid relationship")
var errQuoteState = errors.New("invalid quote state")

// Every mutation locks the same quote row so edits cannot race approval/send.
func (h *QuoteHandler) mutate(c *gin.Context, expected models.QuoteStatus, fn func(*gorm.DB, *models.Quote) error) {
	var quote models.Quote
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("LineItems").First(&quote, "id = ?", c.Param("id")).Error; err != nil {
			return err
		}
		if quote.Status != expected {
			return errQuoteState
		}
		return fn(tx, &quote)
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.NotFound(c, "Quote not found")
		return
	}
	if errors.Is(err, errQuoteState) {
		response.UnprocessableEntity(c, "Quote cannot be changed in its current state")
		return
	}
	if err != nil {
		response.InternalError(c, "Failed to update quote")
		return
	}
	response.OK(c, quote)
}

func (h *QuoteHandler) Update(c *gin.Context) {
	var req updateQuoteRequest
	if !v.BindStrict(c, &req) {
		return
	}
	h.mutate(c, models.QuoteStatusDraft, func(tx *gorm.DB, quote *models.Quote) error {
		if req.Notes != nil {
			quote.Notes = *req.Notes
		}
		if req.ValidUntil != nil {
			date, _ := time.Parse("2006-01-02", *req.ValidUntil)
			quote.ValidUntil = &date
		}
		if req.DiscountPercent != nil {
			quote.DiscountPercent = *req.DiscountPercent
		}
		if req.VATPercent != nil {
			quote.VATPercent = *req.VATPercent
		}
		quote.Recalculate()
		return tx.Omit(clause.Associations).Save(quote).Error
	})
}
func (h *QuoteHandler) Delete(c *gin.Context) {
	h.mutate(c, models.QuoteStatusDraft, func(tx *gorm.DB, quote *models.Quote) error { return tx.Delete(quote).Error })
}
func (h *QuoteHandler) Submit(c *gin.Context) {
	h.changeStatus(c, models.QuoteStatusDraft, models.QuoteStatusPendingApproval)
}
func (h *QuoteHandler) Approve(c *gin.Context) {
	h.changeStatus(c, models.QuoteStatusPendingApproval, models.QuoteStatusApproved)
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
	h.mutate(c, models.QuoteStatusDraft, func(tx *gorm.DB, quote *models.Quote) error {
		for i := range quote.LineItems {
			quote.LineItems[i].Calculate()
			if err := tx.Save(&quote.LineItems[i]).Error; err != nil {
				return err
			}
		}
		quote.Recalculate()
		return tx.Omit(clause.Associations).Save(quote).Error
	})
}
func (h *QuoteHandler) changeStatus(c *gin.Context, from, to models.QuoteStatus) {
	h.mutate(c, from, func(tx *gorm.DB, quote *models.Quote) error {
		if to == models.QuoteStatusPendingApproval && len(quote.LineItems) == 0 {
			return errQuoteState
		}
		quote.Status = to
		if to == models.QuoteStatusApproved {
			userID := middleware.GetCurrentUserID(c)
			now := time.Now()
			quote.ApprovedByID = &userID
			quote.ApprovedAt = &now
		}
		if to == models.QuoteStatusDraft {
			quote.ApprovedByID = nil
			quote.ApprovedAt = nil
		}
		return tx.Omit(clause.Associations).Save(quote).Error
	})
}
