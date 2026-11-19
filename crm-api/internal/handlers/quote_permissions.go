package handlers

import (
	"errors"
	"g4s-crm/api/internal/middleware"
	"g4s-crm/api/internal/models"
	"g4s-crm/api/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var errQuotePermission = errors.New("quotation permission denied")

func sameID(a, b *uuid.UUID) bool {
	return (a == nil && b == nil) || (a != nil && b != nil && *a == *b)
}
func validateQuoteTeam(tx *gorm.DB, c *gin.Context, id *uuid.UUID) error {
	if id == nil {
		return nil
	}
	var team models.Team
	if err := tx.Where("is_active = true").First(&team, "id = ?", id).Error; err != nil {
		return invalid("Choose an active team")
	}
	if middleware.GetCurrentUserRole(c) == models.RoleAdmin {
		return nil
	}
	var count int64
	if err := tx.Table("team_members").Where("team_id = ? AND user_id = ?", id, middleware.GetCurrentUserID(c)).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errQuotePermission
	}
	return nil
}

// Team assignment scopes quotation access; per-user capabilities still apply.
func (h *QuoteHandler) Access(c *gin.Context) {
	if !middleware.Can(c, "quotes:read") {
		response.Forbidden(c, "Quotation view permission is required")
		c.Abort()
		return
	}
	if c.Param("id") != "" {
		var quote models.Quote
		if err := h.db.First(&quote, "id = ?", c.Param("id")).Error; err != nil {
			apiError(c, err)
			c.Abort()
			return
		}
		if err := validateQuoteTeam(h.db, c, quote.TeamID); err != nil && middleware.GetCurrentUserRole(c) != models.RoleAdmin {
			response.Forbidden(c, "This quotation belongs to another or inactive team")
			c.Abort()
			return
		}
	}
	c.Next()
}
func quoteScope(db *gorm.DB, c *gin.Context) *gorm.DB {
	if middleware.GetCurrentUserRole(c) == models.RoleAdmin {
		return db
	}
	return db.Where("(quotes.team_id IS NULL OR quotes.team_id IN (SELECT tm.team_id FROM team_members tm JOIN teams t ON t.id = tm.team_id WHERE tm.user_id = ? AND t.is_active AND t.deleted_at IS NULL))", middleware.GetCurrentUserID(c))
}
func bookItemQuery(tx *gorm.DB, book *uuid.UUID, source string, product, service, recurring *uuid.UUID) *gorm.DB {
	q := tx.Model(&models.PriceBookEntry{}).Where("price_book_id = ?", book)
	switch source {
	case "product":
		return q.Where("product_id = ?", product)
	case "service":
		return q.Where("service_id = ?", service)
	case "recurring":
		return q.Where("recurring_service_id = ?", recurring)
	}
	return q.Where("false")
}
func validateBookItem(tx *gorm.DB, book *uuid.UUID, source string, product, service, recurring *uuid.UUID) error {
	if book == nil {
		return nil
	}
	var count int64
	if err := bookItemQuery(tx, book, source, product, service, recurring).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return invalid("Choose an item from the selected price book")
	}
	return nil
}
func quoteCatalogPrice(tx *gorm.DB, quote *models.Quote, line *models.QuoteLineItem) (float64, float64, error) {
	var cost, price float64
	switch line.Source {
	case "product":
		var item models.Product
		if line.ProductID == nil {
			return 0, 0, invalid("Select a catalog product")
		}
		if err := tx.First(&item, "id = ?", line.ProductID).Error; err != nil {
			return 0, 0, err
		}
		cost, price = item.LandedCostSAR, item.SellingPrice
	case "service":
		var item models.CatalogService
		if line.ServiceID == nil {
			return 0, 0, invalid("Select a catalog service")
		}
		if err := tx.First(&item, "id = ?", line.ServiceID).Error; err != nil {
			return 0, 0, err
		}
		cost, price = item.UnitCost, item.UnitPrice
	case "recurring":
		var item models.RecurringService
		if line.RecurringServiceID == nil {
			return 0, 0, invalid("Select a recurring service")
		}
		if err := tx.First(&item, "id = ?", line.RecurringServiceID).Error; err != nil {
			return 0, 0, err
		}
		cost, price = item.MonthlyCost, item.MonthlyPrice
	default:
		return 0, 0, invalid("Write-in items require pricing and cost visibility permissions")
	}
	if quote.PriceBookID != nil {
		var entry models.PriceBookEntry
		if err := bookItemQuery(tx, quote.PriceBookID, line.Source, line.ProductID, line.ServiceID, line.RecurringServiceID).First(&entry).Error; err != nil {
			return 0, 0, err
		}
		price = entry.CustomPrice
	}
	if quote.Currency != models.CurrencySAR {
		cost /= quote.FXRate
		price /= quote.FXRate
	}
	return models.RoundMoney(cost), models.RoundMoney(price), nil
}

func editableQuoteDocuments(tx *gorm.DB, c *gin.Context, id uuid.UUID) error {
	if !middleware.Can(c, "quotes:documents") || !middleware.Can(c, "quotes:read") {
		return errQuotePermission
	}
	var quote models.Quote
	if err := tx.First(&quote, "id = ?", id).Error; err != nil {
		return err
	}
	return validateQuoteTeam(tx, c, quote.TeamID)
}

// Team choices and approver identity are available with quotation access,
// without granting access to the team directory or its member profiles.
func (h *QuoteHandler) TeamOptions(c *gin.Context) {
	type option struct {
		ID       uuid.UUID  `json:"id"`
		Name     string     `json:"name"`
		LeaderID *uuid.UUID `json:"leaderId,omitempty"`
		IsActive bool       `json:"isActive"`
	}
	items := []option{}
	q := h.db.Model(&models.Team{}).Select("id, name, leader_id, is_active").Where("is_active")
	if middleware.GetCurrentUserRole(c) != models.RoleAdmin {
		q = q.Where("id IN (SELECT team_id FROM team_members WHERE user_id = ?)", middleware.GetCurrentUserID(c))
	}
	if err := q.Order("name").Find(&items).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OK(c, items)
}
