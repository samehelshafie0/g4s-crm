package handlers

import (
	"errors"
	"g4s-crm/api/internal/models"
	"g4s-crm/api/pkg/response"
	"g4s-crm/api/pkg/seqgen"
	v "g4s-crm/api/pkg/validator"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type builderRow struct {
	ID                 string     `json:"id"`
	RowType            string     `json:"rowType" validate:"required,oneof=item heading comment subtotal discount"`
	Source             string     `json:"source" validate:"required,oneof=product service recurring write-in"`
	ProductID          *uuid.UUID `json:"productId"`
	ServiceID          *uuid.UUID `json:"serviceId"`
	RecurringServiceID *uuid.UUID `json:"recurringServiceId"`
	SKU                string     `json:"sku" validate:"max=100"`
	Description        string     `json:"description" validate:"max=2000"`
	Manufacturer       string     `json:"manufacturer" validate:"max=255"`
	StockAvailable     *int       `json:"stockAvailable"`
	LeadTimeDays       *int       `json:"leadTimeDays"`
	Quantity           float64    `json:"quantity" validate:"gte=0,lte=1000000"`
	Multiplier         float64    `json:"multiplier" validate:"gt=0,lte=1000000"`
	UnitCost           float64    `json:"unitCost" validate:"gte=0,lte=1000000000"`
	UnitPrice          float64    `json:"unitPrice" validate:"gte=0,lte=1000000000"`
	DiscountPercent    float64    `json:"discountPercent" validate:"gte=0,lte=100"`
	IsOptional         bool       `json:"isOptional"`
	IsSelected         bool       `json:"isSelected"`
	IsPrintable        bool       `json:"isPrintable"`
	HeadingText        string     `json:"headingText" validate:"max=2000"`
	CommentText        string     `json:"commentText" validate:"max=20000"`
	RateType           string     `json:"rateType" validate:"max=100"`
	BillingCycle       string     `json:"billingCycle" validate:"max=100"`
	SiteID             *uuid.UUID `json:"siteId"`
	FulfilmentSource   *string    `json:"fulfilmentSource" validate:"omitempty,oneof=stock purchase service"`
}

var errStaleQuote = errors.New("quote changed")

func (h *QuoteHandler) SaveBuilder(c *gin.Context) {
	var req struct {
		Appendices       *[]appendixInput    `json:"appendices" validate:"omitempty,max=20,dive"`
		LockVersion      int                 `json:"lockVersion" validate:"required,gt=0"`
		CustomerID       uuid.UUID           `json:"customerId" validate:"required"`
		OpportunityID    *uuid.UUID          `json:"opportunityId"`
		Currency         models.Currency     `json:"currency" validate:"required,oneof=SAR USD"`
		ValidUntil       *string             `json:"validUntil"`
		DiscountPercent  float64             `json:"discountPercent" validate:"gte=0,lte=100"`
		VATPercent       float64             `json:"vatPercent" validate:"gte=0,lte=100"`
		Notes            string              `json:"notes" validate:"max=20000"`
		PurchasingNotes  string              `json:"purchasingNotes" validate:"max=20000"`
		StatementOfWork  string              `json:"statementOfWork" validate:"max=20000"`
		InternalNotes    string              `json:"internalNotes" validate:"max=20000"`
		PaymentTerms     string              `json:"paymentTerms" validate:"max=20000"`
		DeliveryTerms    string              `json:"deliveryTerms" validate:"max=20000"`
		IntroductionText string              `json:"introductionText" validate:"max=20000"`
		ClosingText      string              `json:"closingText" validate:"max=20000"`
		SoldTo           models.QuoteAddress `json:"soldTo"`
		ShipTo           models.QuoteAddress `json:"shipTo"`
		Rows             []builderRow        `json:"rows" validate:"max=1000,dive"`
	}
	if !v.BindStrict(c, &req) {
		return
	}
	date, err := parseDate(req.ValidUntil)
	if err != nil {
		apiError(c, err)
		return
	}
	var quote models.Quote
	err = h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&quote, "id = ?", c.Param("id")).Error; err != nil {
			return err
		}
		if !quote.IsCurrent {
			return errQuoteSuperseded
		}
		if quote.Status != models.QuoteStatusDraft {
			return errQuoteState
		}
		if quote.LockVersion != req.LockVersion {
			return errStaleQuote
		}
		// The quote is bound to its job and customer; the builder cannot move it.
		if req.CustomerID != quote.CustomerID {
			return invalid("A quote cannot move to another customer")
		}
		if req.OpportunityID != nil && (quote.OpportunityID == nil || *req.OpportunityID != *quote.OpportunityID) {
			return invalid("A quote cannot move to another opportunity")
		}
		if req.Currency != quote.Currency {
			return invalid("Change the currency with the currency switch, so prices are converted at a pinned rate")
		}
		allowedSites, err := quoteSites(tx, &quote)
		if err != nil {
			return err
		}
		if req.Appendices != nil {
			if err := replaceAppendices(tx, c, &quote, *req.Appendices); err != nil {
				return err
			}
		}
		lines := []models.QuoteLineItem{}
		for index, row := range req.Rows {
			if row.RowType == "item" && (row.Quantity <= 0 || row.Description == "") {
				return invalid("Item rows need a description and positive quantity")
			}
			if (row.Source != "product" && row.ProductID != nil) || (row.Source != "service" && row.ServiceID != nil) || (row.Source != "recurring" && row.RecurringServiceID != nil) {
				return invalid("Catalog reference must match the row source")
			}
			if err := optionalExists(tx, &models.Product{}, row.ProductID); err != nil {
				return err
			}
			if err := optionalExists(tx, &models.CatalogService{}, row.ServiceID); err != nil {
				return err
			}
			if err := optionalExists(tx, &models.RecurringService{}, row.RecurringServiceID); err != nil {
				return err
			}
			if row.SiteID != nil && !allowedSites[*row.SiteID] {
				return invalid("A line names a site this job does not cover")
			}
			category := models.QuoteLineMaterials
			if row.Source == "service" {
				category = models.QuoteLineManpower
			} else if row.Source != "product" {
				category = models.QuoteLineMiscellaneous
			}
			line := models.QuoteLineItem{QuoteID: quote.ID, Category: category, RowType: row.RowType, Source: row.Source, ProductID: row.ProductID, ServiceID: row.ServiceID, RecurringServiceID: row.RecurringServiceID, SKU: row.SKU, Description: row.Description, ManufacturerName: row.Manufacturer, Quantity: row.Quantity, Multiplier: row.Multiplier, UnitCost: row.UnitCost, UnitPrice: row.UnitPrice, DiscountPercent: row.DiscountPercent, IsOptional: row.IsOptional, IsSelected: row.IsSelected, IsPrintable: row.IsPrintable, HeadingText: row.HeadingText, CommentText: row.CommentText, RateType: row.RateType, BillingCycle: row.BillingCycle, LeadTimeDays: row.LeadTimeDays, SortOrder: index, SiteID: row.SiteID, FulfilmentSource: row.FulfilmentSource}
			line.Calculate()
			lines = append(lines, line)
		}
		if err := tx.Where("quote_id = ?", quote.ID).Unscoped().Delete(&models.QuoteLineItem{}).Error; err != nil {
			return err
		}
		// Explicit field selection preserves false optional/print flags despite GORM defaults.
		for i := range lines {
			selected, printable := lines[i].IsSelected, lines[i].IsPrintable
			if err := tx.Create(&lines[i]).Error; err != nil {
				return err
			}
			if err := tx.Model(&lines[i]).Updates(map[string]any{"is_selected": selected, "is_printable": printable}).Error; err != nil {
				return err
			}
			lines[i].IsSelected = selected
			lines[i].IsPrintable = printable
		}
		quote.ValidUntil = date
		quote.DiscountPercent = req.DiscountPercent
		quote.VATPercent = req.VATPercent
		quote.Notes = req.Notes
		quote.InternalNotes = req.InternalNotes
		quote.PurchasingNotes = req.PurchasingNotes
		quote.StatementOfWork = req.StatementOfWork
		quote.PaymentTerms = req.PaymentTerms
		quote.DeliveryTerms = req.DeliveryTerms
		quote.IntroductionText = req.IntroductionText
		quote.ClosingText = req.ClosingText
		quote.SoldTo = req.SoldTo
		quote.ShipTo = req.ShipTo
		quote.LineItems = lines
		quote.LockVersion++
		quote.Recalculate()
		if err := tx.Omit(clause.Associations).Save(&quote).Error; err != nil {
			return err
		}
		return recordActivity(tx, c, "quote", quote.ID, "saved builder")
	})
	if errors.Is(err, errQuoteSuperseded) {
		response.UnprocessableEntity(c, "This version is superseded and read-only; edit the current version")
		return
	}
	if errors.Is(err, errQuoteState) {
		response.UnprocessableEntity(c, "Only draft quotes can be edited")
		return
	}
	if errors.Is(err, errStaleQuote) {
		response.Conflict(c, "Quote changed in another session; reload before saving")
		return
	}
	if err != nil {
		apiError(c, err)
		return
	}
	h.Get(c)
}
func (h *QuoteHandler) ConvertToContract(c *gin.Context) {
	var req struct {
		Title     string              `json:"title" validate:"required,max=255"`
		Type      models.ContractType `json:"type" validate:"required,oneof=sales maintenance service project subscription"`
		StartDate string              `json:"startDate" validate:"required,datetime=2006-01-02"`
		EndDate   string              `json:"endDate" validate:"required,datetime=2006-01-02"`
		Terms     string              `json:"terms" validate:"max=20000"`
	}
	if !v.BindStrict(c, &req) {
		return
	}
	start, _ := time.Parse("2006-01-02", req.StartDate)
	end, _ := time.Parse("2006-01-02", req.EndDate)
	if !end.After(start) {
		response.BadRequest(c, "End date must be after start date")
		return
	}
	var result models.Contract
	reused := false
	err := h.db.Transaction(func(tx *gorm.DB) error {
		var quote models.Quote
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&quote, "id = ?", c.Param("id")).Error; err != nil {
			return err
		}
		if quote.Status != models.QuoteStatusAccepted {
			return errQuoteState
		}
		err := tx.First(&result, "quote_id = ?", quote.ID).Error
		if err == nil {
			reused = true
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		number, err := seqgen.NextNumber(tx, "contract")
		if err != nil {
			return err
		}
		result = models.Contract{Currency: quote.Currency, ContractNumber: number, Title: req.Title, CustomerID: quote.CustomerID, QuoteID: &quote.ID, Type: req.Type, Status: models.ContractStatusDraft, StartDate: &start, EndDate: &end, Value: quote.Total, Terms: req.Terms}
		if err := tx.Create(&result).Error; err != nil {
			return err
		}
		return recordActivity(tx, c, "contract", result.ID, "created from "+quote.QuoteNumber)
	})
	if errors.Is(err, errQuoteState) {
		response.UnprocessableEntity(c, "Only accepted quotes can become contracts")
		return
	}
	if err != nil {
		apiError(c, err)
		return
	}
	if reused {
		response.OK(c, result)
	} else {
		response.Created(c, result)
	}
}
