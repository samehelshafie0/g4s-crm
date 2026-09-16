package handlers

import (
	"g4s-crm/api/internal/models"
	"g4s-crm/api/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

func eligiblePriceBook(db *gorm.DB, id *uuid.UUID, customerID uuid.UUID) error {
	if id == nil {
		return nil
	}
	var book models.PriceBook
	if err := db.First(&book, "id = ?", id).Error; err != nil {
		return err
	}
	today := time.Now().Truncate(24 * time.Hour)
	if !book.IsActive || (book.CustomerID != nil && *book.CustomerID != customerID) || (book.ValidFrom != nil && book.ValidFrom.After(today)) || (book.ValidTo != nil && book.ValidTo.Before(today)) {
		return invalid("Price book is not eligible for this customer or date")
	}
	return nil
}
func (h *ProductHandler) Catalog(c *gin.Context) {
	products := []models.Product{}
	services := []models.CatalogService{}
	recurring := []models.RecurringService{}
	stocks := []models.WarehouseStock{}
	for _, err := range []error{productQuery(h.db).Where("is_active = true").Find(&products).Error, h.db.Where("is_active = true").Find(&services).Error, h.db.Where("is_active = true").Find(&recurring).Error, h.db.Find(&stocks).Error} {
		if err != nil {
			apiError(c, err)
			return
		}
	}
	available := map[uuid.UUID]int{}
	for _, stock := range stocks {
		available[stock.ProductID] += stock.AvailableQty
	}
	items := []gin.H{}
	for _, p := range products {
		mfr := ""
		if p.Manufacturer != nil {
			mfr = p.Manufacturer.Name
		}
		history := []gin.H{}
		for _, price := range p.PriceHistory {
			history = append(history, gin.H{"id": price.ID, "source": price.Source, "supplier": price.VendorName, "date": price.Date, "cost": price.LandingCost, "price": p.SellingPrice, "qty": price.Qty, "batchRef": price.SourceRef})
		}
		items = append(items, gin.H{"id": p.ID, "sku": p.SKU, "name": p.Name, "manufacturer": mfr, "unitCost": p.LandedCostSAR, "unitPrice": p.SellingPrice, "stockAvailable": available[p.ID], "leadTimeDays": p.LeadTimeDays, "priceHistory": history})
	}
	recurringItems := []gin.H{}
	for _, r := range recurring {
		recurringItems = append(recurringItems, gin.H{"id": r.ID, "sku": "REC-" + r.ID.String()[:8], "name": r.Name, "billingCycle": r.BillingFrequency, "monthlyCost": r.MonthlyCost, "monthlyPrice": r.MonthlyPrice})
	}
	response.OK(c, gin.H{"products": items, "services": services, "recurring": recurringItems})
}
func (h *QuoteHandler) Activity(c *gin.Context) {
	var quote models.Quote
	if err := h.db.First(&quote, "id = ?", c.Param("id")).Error; err != nil {
		apiError(c, err)
		return
	}
	items := []models.ActivityLog{}
	if err := h.db.Preload("User").Where("entity_type = 'quote' AND entity_id = ?", quote.ID).Order("created_at ASC").Find(&items).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OK(c, items)
}
