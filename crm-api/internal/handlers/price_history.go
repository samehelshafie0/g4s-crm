package handlers

import (
	"g4s-crm/api/internal/middleware"
	"g4s-crm/api/internal/models"
	"g4s-crm/api/pkg/pagination"
	"g4s-crm/api/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

// Called inside the supplier quote's write transaction and parent row lock.
func recordSupplierPrices(tx *gorm.DB, sq *models.SupplierQuote) error {
	var revision int
	if err := tx.Model(&models.SupplierPriceRecord{}).Where("supplier_quote_id = ?", sq.ID).Select("COALESCE(MAX(revision),0)+1").Scan(&revision).Error; err != nil {
		return err
	}
	now := time.Now()
	for _, item := range sq.Items {
		if item.ProductID == nil {
			continue
		}
		valid := item.ValidUntil
		if valid == nil {
			valid = sq.ValidUntil
		}
		record := models.SupplierPriceRecord{RecordedAt: now, SupplierQuoteID: sq.ID, SourceItemID: item.ID, Revision: revision, ProductID: *item.ProductID, SKU: item.ProductSKU, ProductName: item.ProductName, SupplierName: sq.SupplierName, SupplierRef: sq.SupplierRef, SupplierQuoteNumber: sq.SQNumber, Currency: sq.Currency, UnitCost: item.UnitCost, Quantity: item.Quantity, ValidUntil: valid, SourceQuoteID: sq.SourceQuoteID}
		if err := tx.Create(&record).Error; err != nil {
			return err
		}
	}
	return nil
}

// Accepted quotations are immutable, so their original lines are the sale-price
// record. Optional unselected items and all unaccepted revisions are excluded.
func (h *ProductHandler) SalesHistory(c *gin.Context) {
	customerID, err := uuid.Parse(c.Query("customerId"))
	if err != nil {
		response.BadRequest(c, "Choose a customer for price comparison")
		return
	}
	scope := c.DefaultQuery("scope", "customer")
	if scope != "customer" && scope != "others" {
		response.BadRequest(c, "Choose customer or others")
		return
	}
	if scope == "others" && !middleware.Can(c, "quotes:compare-prices") {
		response.Forbidden(c, "Other-customer price comparison permission is required")
		return
	}
	q := acceptedPriceQuery(h.db, c).Where("line.product_id = ?", c.Param("id"))
	if scope == "customer" {
		q = q.Where("quotes.customer_id = ?", customerID)
	} else {
		q = q.Where("quotes.customer_id <> ?", customerID)
	}
	sendSalePrices(c, q)
}
func acceptedPriceQuery(db *gorm.DB, c *gin.Context) *gorm.DB {
	return quoteScope(db, c).Table("quote_line_items AS line").Joins("JOIN quotes ON quotes.id = line.quote_id AND quotes.deleted_at IS NULL").Joins("JOIN customers ON customers.id = quotes.customer_id AND customers.deleted_at IS NULL").Where("quotes.status = 'accepted' AND line.deleted_at IS NULL AND line.row_type = 'item' AND (NOT line.is_optional OR line.is_selected)")
}
func sendSalePrices(c *gin.Context, q *gorm.DB) {
	params := pagination.GetParams(c)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		apiError(c, err)
		return
	}
	items := []map[string]any{}
	fields := `line.id, line.product_id AS "productId", line.sku, line.description, line.quantity, line.multiplier, line.unit_price AS "unitPrice", line.discount_percent AS "discountPercent", quotes.discount_percent AS "quoteDiscountPercent", ROUND((line.unit_price * (1-line.discount_percent/100) * (1-quotes.discount_percent/100))::numeric,2) AS "netUnitPrice", quotes.id AS "quoteId", quotes.quote_number AS "quoteNumber", quotes.team_id AS "teamId", quotes.version, quotes.currency, quotes.fx_rate AS "fxRate", quotes.customer_id AS "customerId", customers.company_name AS "customerName", quotes.accepted_at AS "acceptedAt", quotes.accepted_at_estimated AS "acceptedAtEstimated"`
	if err := q.Select(fields).Order("quotes.accepted_at DESC, line.id").Scopes(pagination.Paginate(params)).Find(&items).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OKWithMeta(c, items, pagination.BuildMeta(params, total))
}
