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
	"strconv"
	"time"
)

type QuoteHandler struct{ db *gorm.DB }

func NewQuoteHandler(db *gorm.DB) *QuoteHandler { return &QuoteHandler{db: db} }

func (h *QuoteHandler) List(c *gin.Context) {
	params := pagination.GetParams(c, "quote_number", "status", "total", "valid_until")
	var quotes []models.Quote
	var total int64

	query := h.db.Model(&models.Quote{}).Preload("Customer").Preload("LineItems", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC") })
	if q := c.Query("q"); q != "" {
		query = query.Where("quote_number ILIKE ?", "%"+q+"%")
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if cid := c.Query("customerId"); cid != "" {
		query = query.Where("customer_id = ?", cid)
	}
	if oid := c.Query("opportunityId"); oid != "" {
		query = query.Where("opportunity_id = ?", oid)
	}
	if kid := c.Query("contractId"); kid != "" {
		query = query.Where("contract_id = ?", kid)
	}
	if c.Query("versions") != "all" {
		query = query.Where("is_current")
	}

	if err := query.Count(&total).Error; err != nil {
		apiError(c, err)
		return
	}
	if err := query.Preload("Opportunity").Order(params.Sort + " " + params.Order).Scopes(pagination.Paginate(params)).Find(&quotes).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OKWithMeta(c, quotes, pagination.BuildMeta(params, total))
}

func (h *QuoteHandler) Get(c *gin.Context) {
	var quote models.Quote
	if err := h.db.Preload("Appendices", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC") }).Preload("LineItems", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC") }).Preload("LineItems.Product").Preload("Customer").Preload("ApprovedBy").Preload("Opportunity").
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
	PriceBookID     *uuid.UUID         `json:"priceBookId"`
	CustomerID      uuid.UUID          `json:"customerId" validate:"required"`
	OpportunityID   *uuid.UUID         `json:"opportunityId"`
	ContractID      *uuid.UUID         `json:"contractId"`
	Currency        models.Currency    `json:"currency" validate:"omitempty,oneof=SAR USD"`
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
		if (req.OpportunityID == nil) == (req.ContractID == nil) {
			return invalid("A quote belongs to exactly one opportunity or one contract")
		}
		if req.OpportunityID != nil {
			var opportunity models.Opportunity
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&opportunity, "id = ? AND customer_id = ?", req.OpportunityID, req.CustomerID).Error; err != nil {
				return errInvalidRelationship
			}
			// One opportunity, one quote: further pricing is a new version of it.
			var existing models.Quote
			err := tx.Where("opportunity_id = ? AND is_current", opportunity.ID).First(&existing).Error
			if err == nil {
				return errQuoteExists{number: existing.QuoteNumber, id: existing.ID}
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}
		if req.ContractID != nil {
			var contract models.Contract
			if err := tx.First(&contract, "id = ? AND customer_id = ?", req.ContractID, req.CustomerID).Error; err != nil {
				return errInvalidRelationship
			}
			if contract.Status == models.ContractStatusTerminated || contract.Status == models.ContractStatusExpired {
				return invalid("Quotes cannot be raised inside a terminated or expired contract")
			}
		}
		if err := eligiblePriceBook(tx, req.PriceBookID, req.CustomerID); err != nil {
			return errInvalidRelationship
		}
		number, err := seqgen.NextNumber(tx, "quote")
		if err != nil {
			return err
		}
		userID := middleware.GetCurrentUserID(c)
		quote = models.Quote{PriceBookID: req.PriceBookID, QuoteNumber: number, Version: 1, IsCurrent: true, CustomerID: req.CustomerID, OpportunityID: req.OpportunityID, ContractID: req.ContractID, Status: models.QuoteStatusDraft, Currency: req.Currency, DiscountPercent: req.DiscountPercent, VATPercent: 15, Notes: req.Notes, CreatedByID: &userID}
		if quote.Currency == "" {
			quote.Currency = models.CurrencySAR
		}
		rate, err := pinRate(tx, quote.Currency)
		if err != nil {
			return err
		}
		quote.FXRate = rate
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
			line := models.QuoteLineItem{Category: input.Category, ProductID: input.ProductID, SKU: input.SKU, Description: input.Description, Quantity: float64(input.Quantity), UnitCost: input.UnitCost, UnitPrice: input.UnitPrice, SortOrder: index}
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
	var exists errQuoteExists
	if errors.As(err, &exists) {
		c.JSON(409, gin.H{"success": false, "error": gin.H{"message": "This opportunity already has quote " + exists.number + "; revise it instead of starting another", "quoteId": exists.id}})
		return
	}
	if errors.Is(err, errInvalidRelationship) {
		response.BadRequest(c, "Customer, opportunity, contract or product does not exist or does not belong to this quote")
		return
	}
	if err != nil {
		apiError(c, err)
		return
	}
	response.Created(c, quote)
}

var errInvalidRelationship = errors.New("invalid relationship")

// errQuoteExists names the quote an opportunity already has.
type errQuoteExists struct {
	number string
	id     uuid.UUID
}

func (e errQuoteExists) Error() string { return "opportunity already has quote " + e.number }

var errQuoteState = errors.New("invalid quote state")

// Every mutation locks the same quote row so edits cannot race approval/send.
func (h *QuoteHandler) mutate(c *gin.Context, expected models.QuoteStatus, fn func(*gorm.DB, *models.Quote) error) {
	var quote models.Quote
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("LineItems").First(&quote, "id = ?", c.Param("id")).Error; err != nil {
			return err
		}
		if !quote.IsCurrent {
			return errQuoteSuperseded
		}
		if quote.Status != expected {
			return errQuoteState
		}
		quote.LockVersion++
		if err := fn(tx, &quote); err != nil {
			return err
		}
		return recordActivity(tx, c, "quote", quote.ID, string(quote.Status))
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.NotFound(c, "Quote not found")
		return
	}
	if errors.Is(err, errQuoteSuperseded) {
		response.UnprocessableEntity(c, "This version is superseded and read-only")
		return
	}
	if errors.Is(err, errQuoteState) {
		response.UnprocessableEntity(c, "Quote cannot be changed in its current state")
		return
	}
	if err != nil {
		apiError(c, err)
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
	h.mutate(c, models.QuoteStatusDraft, func(tx *gorm.DB, quote *models.Quote) error {
		if err := tx.Delete(quote).Error; err != nil {
			return err
		}
		if quote.ParentQuoteID == nil {
			return nil
		}
		return tx.Model(&models.Quote{}).Where("id = ?", quote.ParentQuoteID).Updates(map[string]any{"is_current": true, "superseded_at": nil}).Error
	})
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
		if to == models.QuoteStatusPendingApproval {
			hasItem := false
			for _, line := range quote.LineItems {
				if (line.RowType == "item" || line.RowType == "") && (!line.IsOptional || line.IsSelected) && line.Quantity > 0 {
					hasItem = true
					break
				}
			}
			if !hasItem {
				return errQuoteState
			}
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
		if err := tx.Omit(clause.Associations).Save(quote).Error; err != nil {
			return err
		}
		if to == models.QuoteStatusAccepted && quote.OpportunityID != nil {
			var opportunity models.Opportunity
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&opportunity, "id = ?", quote.OpportunityID).Error; err != nil {
				return err
			}
			if opportunity.Stage != models.StageClosedWon {
				previous := opportunity.Stage
				opportunity.Stage = models.StageClosedWon
				syncDelivery(&opportunity)
				if err := tx.Omit(clause.Associations).Save(&opportunity).Error; err != nil {
					return err
				}
				return recordActivity(tx, c, "opportunity", opportunity.ID, "stage: "+string(previous)+" → closed-won (quote "+quote.QuoteNumber+" v"+strconv.Itoa(quote.Version)+" accepted)")
			}
		}
		return nil
	})
}
