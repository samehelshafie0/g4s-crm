package handlers

import (
	"errors"
	"strconv"
	"time"

	"g4s-crm/api/internal/middleware"
	"g4s-crm/api/internal/models"
	"g4s-crm/api/pkg/response"
	v "g4s-crm/api/pkg/validator"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var errQuoteSuperseded = errors.New("quote version superseded")

// pinRate returns the SAR rate a customer quote in this currency is priced at.
// Customers are quoted in SAR or USD only; the USD rate comes from the exchange
// rate table as it stands now and is then fixed on the version.
func pinRate(tx *gorm.DB, currency models.Currency) (float64, error) {
	switch currency {
	case models.CurrencySAR, "":
		return 1, nil
	case models.CurrencyUSD:
		var rate models.ExchangeRate
		if err := tx.Where("from_currency = ? AND to_currency = 'SAR'", models.CurrencyUSD).First(&rate).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return 0, invalid("There is no USD rate; add one under Exchange Rates before quoting in USD")
			}
			return 0, err
		}
		return rate.CurrentRate, nil
	}
	return 0, invalid("Customers are quoted in SAR or USD only")
}

// quoteSites is the set of sites a quote's lines may name: the sites its job
// covers, or for a contract quote, any active site of the customer.
func quoteSites(tx *gorm.DB, quote *models.Quote) (map[uuid.UUID]bool, error) {
	allowed := map[uuid.UUID]bool{}
	var ids []uuid.UUID
	var err error
	if quote.OpportunityID != nil {
		err = tx.Raw("SELECT site_id FROM opportunity_sites WHERE opportunity_id = ?", quote.OpportunityID).Scan(&ids).Error
	} else {
		err = tx.Raw("SELECT id FROM customer_sites WHERE customer_id = ? AND deleted_at IS NULL", quote.CustomerID).Scan(&ids).Error
	}
	for _, id := range ids {
		allowed[id] = true
	}
	return allowed, err
}

// Versions lists every version of a quote, newest first. Superseded versions
// are read-only history; each still opens and exports exactly as it stood.
func (h *QuoteHandler) Versions(c *gin.Context) {
	var quote models.Quote
	if err := h.db.First(&quote, "id = ?", c.Param("id")).Error; err != nil {
		apiError(c, err)
		return
	}
	versions := []models.Quote{}
	if err := h.db.Preload("CreatedBy").Where("quote_number = ?", quote.QuoteNumber).Order("version DESC").Find(&versions).Error; err != nil {
		apiError(c, err)
		return
	}
	out := make([]gin.H, 0, len(versions))
	for _, version := range versions {
		author := ""
		if version.CreatedBy != nil {
			author = version.CreatedBy.FullName()
		}
		out = append(out, gin.H{
			"id": version.ID, "quoteNumber": version.QuoteNumber, "version": version.Version,
			"isCurrent": version.IsCurrent, "status": version.Status, "total": version.Total,
			"currency": version.Currency, "fxRate": version.FXRate, "marginPercent": version.MarginPercent,
			"revisionReason": version.RevisionReason, "createdAt": version.CreatedAt,
			"supersededAt": version.SupersededAt, "createdBy": author,
		})
	}
	response.OK(c, out)
}

// revisable is where a version can be revised from. A draft is simply edited;
// a version awaiting approval is rejected back to draft first; an accepted
// version is the deal and is not reopened.
var revisable = map[models.QuoteStatus]bool{
	models.QuoteStatusApproved: true, models.QuoteStatusSent: true,
	models.QuoteStatusDeclined: true, models.QuoteStatusExpired: true,
}

// Revise creates the next version of a quote. The number stays; the version
// rises; the reason is recorded; the version it replaces is frozen as history.
func (h *QuoteHandler) Revise(c *gin.Context) {
	var req struct {
		Reason string `json:"reason" validate:"required,max=500"`
	}
	if !v.BindStrict(c, &req) {
		return
	}
	var result models.Quote
	err := h.db.Transaction(func(tx *gorm.DB) error {
		var source models.Quote
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("LineItems", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC") }).Preload("Appendices", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC") }).First(&source, "id = ?", c.Param("id")).Error; err != nil {
			return err
		}
		if !source.IsCurrent {
			return errQuoteSuperseded
		}
		if !revisable[source.Status] {
			return errQuoteState
		}
		var latest int
		if err := tx.Model(&models.Quote{}).Unscoped().Where("quote_number = ?", source.QuoteNumber).Select("COALESCE(MAX(version), 0)").Scan(&latest).Error; err != nil {
			return err
		}
		// Freeze the old version first: only one version of a number may be current.
		now := time.Now()
		if err := tx.Model(&models.Quote{}).Where("id = ?", source.ID).Updates(map[string]any{"is_current": false, "superseded_at": now}).Error; err != nil {
			return err
		}
		userID := middleware.GetCurrentUserID(c)
		result = source
		result.Base = models.Base{}
		result.ParentQuoteID = &source.ID
		result.Version = latest + 1
		result.IsCurrent = true
		result.SupersededAt = nil
		result.RevisionReason = req.Reason
		result.LockVersion = 1
		result.Status = models.QuoteStatusDraft
		result.ApprovedByID, result.ApprovedAt, result.ApprovedBy = nil, nil, nil
		result.Customer, result.Opportunity, result.CreatedBy = nil, nil, nil
		result.CreatedByID = &userID
		appendices, lines := result.Appendices, result.LineItems
		result.Appendices, result.LineItems = nil, nil
		if err := tx.Omit(clause.Associations).Create(&result).Error; err != nil {
			return err
		}
		// Column defaults would otherwise replace these explicit values.
		if err := tx.Model(&result).Updates(map[string]any{"vat_percent": source.VATPercent, "fx_rate": source.FXRate}).Error; err != nil {
			return err
		}
		for _, appendix := range appendices {
			appendix.Base = models.Base{}
			appendix.QuoteID = result.ID
			if err := tx.Create(&appendix).Error; err != nil {
				return err
			}
			result.Appendices = append(result.Appendices, appendix)
		}
		for _, line := range lines {
			selected, printable := line.IsSelected, line.IsPrintable
			line.Base = models.Base{}
			line.QuoteID = result.ID
			line.Product = nil
			if err := tx.Create(&line).Error; err != nil {
				return err
			}
			if err := tx.Model(&line).Updates(map[string]any{"is_selected": selected, "is_printable": printable}).Error; err != nil {
				return err
			}
			line.IsSelected, line.IsPrintable = selected, printable
			result.LineItems = append(result.LineItems, line)
		}
		return recordActivity(tx, c, "quote", result.ID, "revised to v"+strconv.Itoa(result.Version)+": "+req.Reason)
	})
	if errors.Is(err, errQuoteSuperseded) {
		response.UnprocessableEntity(c, "Only the current version can be revised")
		return
	}
	if errors.Is(err, errQuoteState) {
		response.UnprocessableEntity(c, "Revise an approved, sent, declined or expired version; edit a draft directly, and reject a pending one first")
		return
	}
	if err != nil {
		apiError(c, err)
		return
	}
	response.Created(c, result)
}

// ChangeCurrency moves a draft between SAR and USD. Every line is converted
// through SAR at the old and new pinned rates, so the customer-facing figures
// change and the margin does not.
func (h *QuoteHandler) ChangeCurrency(c *gin.Context) {
	var req struct {
		Currency models.Currency `json:"currency" validate:"required,oneof=SAR USD"`
	}
	if !v.BindStrict(c, &req) {
		return
	}
	var quote models.Quote
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("LineItems").First(&quote, "id = ?", c.Param("id")).Error; err != nil {
			return err
		}
		if !quote.IsCurrent {
			return errQuoteSuperseded
		}
		if quote.Status != models.QuoteStatusDraft {
			return errQuoteState
		}
		rate, err := pinRate(tx, req.Currency)
		if err != nil {
			return err
		}
		factor := quote.FXRate / rate
		if quote.Currency == req.Currency {
			// Same currency: re-pin today's rate without touching prices.
			factor = 1
		}
		for i := range quote.LineItems {
			line := &quote.LineItems[i]
			line.UnitCost = models.RoundMoney(line.UnitCost * factor)
			line.UnitPrice = models.RoundMoney(line.UnitPrice * factor)
			line.Calculate()
			if err := tx.Model(line).Updates(map[string]any{"unit_cost": line.UnitCost, "unit_price": line.UnitPrice, "line_total": line.LineTotal, "margin_percent": line.MarginPercent}).Error; err != nil {
				return err
			}
		}
		previous := quote.Currency
		quote.Currency = req.Currency
		quote.FXRate = rate
		quote.LockVersion++
		quote.Recalculate()
		if err := tx.Omit(clause.Associations).Save(&quote).Error; err != nil {
			return err
		}
		return recordActivity(tx, c, "quote", quote.ID, "currency "+string(previous)+" → "+string(req.Currency)+" at "+strconv.FormatFloat(rate, 'f', 4, 64))
	})
	if errors.Is(err, errQuoteSuperseded) {
		response.UnprocessableEntity(c, "This version is superseded and read-only")
		return
	}
	if errors.Is(err, errQuoteState) {
		response.UnprocessableEntity(c, "Only a draft can change currency")
		return
	}
	if err != nil {
		apiError(c, err)
		return
	}
	h.Get(c)
}
