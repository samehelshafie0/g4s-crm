package handlers

import (
	"g4s-crm/api/internal/models"
	"g4s-crm/api/pkg/response"
	v "g4s-crm/api/pkg/validator"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Supplier documents and item offers belong to the job, including earlier quote versions.
func (h *QuoteHandler) SupplierQuotes(c *gin.Context) {
	var quote models.Quote
	if err := h.db.First(&quote, "id = ?", c.Param("id")).Error; err != nil {
		apiError(c, err)
		return
	}
	items := []models.SupplierQuote{}
	q := h.db.Preload("Items").Preload("SourceQuote").Preload("Opportunity")
	if c.Query("available") == "true" {
		q = q.Where("source_quote_id IS NULL AND (opportunity_id IS NULL OR opportunity_id = ?)", quote.OpportunityID)
	} else {
		q = q.Where("source_quote_id IN (SELECT id FROM quotes WHERE quote_number = ? AND deleted_at IS NULL)", quote.QuoteNumber)
	}
	if err := q.Order("created_at DESC").Find(&items).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OK(c, items)
}
func (h *QuoteHandler) LinkSupplierQuote(c *gin.Context) {
	var req struct {
		SupplierQuoteID uuid.UUID `json:"supplierQuoteId" validate:"required"`
	}
	if !v.BindStrict(c, &req) {
		return
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		var quote models.Quote
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&quote, "id = ?", c.Param("id")).Error; err != nil {
			return err
		}
		if !quote.IsCurrent || quote.Status != models.QuoteStatusDraft {
			return errQuoteState
		}
		var supplier models.SupplierQuote
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&supplier, "id = ?", req.SupplierQuoteID).Error; err != nil {
			return err
		}
		if supplier.SourceQuoteID != nil && *supplier.SourceQuoteID != quote.ID {
			return invalid("This supplier quotation is already linked to another quotation")
		}
		if supplier.OpportunityID != nil && !sameID(supplier.OpportunityID, quote.OpportunityID) {
			return invalid("This supplier quotation belongs to another opportunity")
		}
		if err := tx.Model(&supplier).Updates(map[string]any{"source_quote_id": quote.ID, "opportunity_id": quote.OpportunityID}).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.SupplierPriceRecord{}).Where("supplier_quote_id = ? AND source_quote_id IS NULL", supplier.ID).Update("source_quote_id", quote.ID).Error; err != nil {
			return err
		}
		return recordActivity(tx, c, "quote", quote.ID, "linked supplier quotation "+supplier.SQNumber)
	})
	if err != nil {
		apiError(c, err)
		return
	}
	h.SupplierQuotes(c)
}

func (h *ProductHandler) QuoteOffers(c *gin.Context) {
	items := []map[string]any{}
	q := quoteScope(h.db, c).Table("supplier_price_records record").Select(`record.id, record.unit_cost AS "unitCost", record.quantity, record.currency, record.supplier_quote_id AS "supplierQuoteId", record.supplier_quote_number AS "supplierQuoteNumber", record.supplier_name AS "supplierName", record.supplier_ref AS "supplierRef", record.revision, record.recorded_at AS "recordedAt", record.valid_until AS "validUntil", quotes.id AS "quoteId", quotes.quote_number AS "quoteNumber", quotes.version, quotes.customer_id AS "customerId", customers.company_name AS "customerName", opportunities.title AS "projectName", EXISTS (SELECT 1 FROM supplier_quote_items live JOIN supplier_quotes sq ON sq.id = live.supplier_quote_id WHERE live.id = record.source_item_id AND live.deleted_at IS NULL AND sq.deleted_at IS NULL) AS "isCurrent"`).Joins("LEFT JOIN quotes ON quotes.id = record.source_quote_id AND quotes.deleted_at IS NULL").Joins("LEFT JOIN customers ON customers.id = quotes.customer_id").Joins("LEFT JOIN opportunities ON opportunities.id = quotes.opportunity_id").Where("record.product_id = ? AND (record.source_quote_id IS NULL OR quotes.id IS NOT NULL)", c.Param("id"))
	if err := q.Order("record.recorded_at DESC, record.id").Find(&items).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OK(c, items)
}
func (h *CustomerHandler) PriceHistory(c *gin.Context) {
	if c.Query("scope") == "sold" {
		q := acceptedPriceQuery(h.db, c).Where("quotes.customer_id = ?", c.Param("id"))
		if id := c.Query("productId"); id != "" {
			q = q.Where("line.product_id = ?", id)
		}
		sendSalePrices(c, q)
		return
	}

	items := []map[string]any{}
	q := quoteScope(h.db, c).Table("quote_line_items AS line").Select("line.id, line.product_id AS \"productId\", line.sku, line.description, line.quantity, line.unit_price AS \"unitPrice\", line.unit_cost AS \"unitCost\", line.discount_percent AS \"discountPercent\", quotes.id AS \"quoteId\", quotes.quote_number AS \"quoteNumber\", quotes.version, quotes.currency, quotes.fx_rate AS \"fxRate\", quotes.status, quotes.created_at AS date").Joins("JOIN quotes ON quotes.id = line.quote_id AND quotes.deleted_at IS NULL").Where("quotes.customer_id = ? AND line.deleted_at IS NULL AND line.row_type = 'item'", c.Param("id"))
	if id := c.Query("productId"); id != "" {
		q = q.Where("line.product_id = ?", id)
	}
	if err := q.Order("quotes.created_at DESC, line.sort_order").Find(&items).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OK(c, items)
}
