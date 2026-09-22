package integration

import (
	"net/http/httptest"
	"testing"

	"g4s-crm/api/internal/models"
	"g4s-crm/api/internal/seed"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func testSeedCatalog(t *testing.T, db *gorm.DB, r *gin.Engine, token string) {
	t.Run("catalog seed is complete, priced and idempotent", func(t *testing.T) {
		first, err := seed.Catalog(db)
		must(t, err)
		if first.Manufacturers < 20 || first.Products < 50 || first.Services < 10 || first.Recurring < 5 || first.Rentals < 5 {
			t.Fatalf("seed too small: %+v", first)
		}
		count := func(model any) int64 {
			t.Helper()
			var n int64
			must(t, db.Model(model).Count(&n).Error)
			return n
		}
		// Rates created earlier in the suite are kept, so check the table rather than the run.
		if count(&models.ExchangeRate{}) < 5 {
			t.Fatal("expected USD, EUR, GBP, AED and CNY rates")
		}
		before := []int64{count(&models.Manufacturer{}), count(&models.ManufacturerCategory{}), count(&models.Product{}), count(&models.CatalogService{}), count(&models.RecurringService{}), count(&models.ExchangeRate{}), count(&models.ProductPriceRecord{})}

		// Re-running updates in place and must not duplicate or re-add history.
		second, err := seed.Catalog(db)
		must(t, err)
		if second.Categories != 0 || second.ExchangeRates != 0 {
			t.Fatalf("second run created new categories or rates: %+v", second)
		}
		after := []int64{count(&models.Manufacturer{}), count(&models.ManufacturerCategory{}), count(&models.Product{}), count(&models.CatalogService{}), count(&models.RecurringService{}), count(&models.ExchangeRate{}), count(&models.ProductPriceRecord{})}
		for i := range before {
			if before[i] != after[i] {
				t.Fatalf("re-seed changed row counts: before %v after %v", before, after)
			}
		}

		var products []models.Product
		must(t, db.Preload("Manufacturer").Preload("Category").Where("id IN (SELECT product_id FROM product_price_records WHERE source_ref = ?)", "seed catalog").Find(&products).Error)
		if len(products) != first.Products {
			t.Fatalf("expected %d seeded products, found %d", first.Products, len(products))
		}
		for _, p := range products {
			if p.Manufacturer == nil || p.Category == nil || p.LandedCostSAR <= 0 || p.SellingPrice <= p.LandedCostSAR {
				t.Fatalf("product %s is not fully priced or linked: landed=%v selling=%v", p.SKU, p.LandedCostSAR, p.SellingPrice)
			}
			if p.ProductType == models.ProductTypeImport && p.FXRate <= 1 {
				t.Fatalf("imported product %s has no FX applied", p.SKU)
			}
		}

		// Rentals price hardware per month, so a rental must never cost its purchase price.
		var rentals []models.RecurringService
		must(t, db.Where("service_type = ?", models.RecurringRental).Find(&rentals).Error)
		if len(rentals) != first.Rentals {
			t.Fatalf("expected %d rentals, found %d", first.Rentals, len(rentals))
		}
		for _, rental := range rentals {
			if len(rental.LineItems) == 0 {
				t.Fatalf("rental %q has no line items", rental.Name)
			}
			if rental.MonthlyPrice <= rental.MonthlyCost || rental.AnnualPrice != models.RoundMoney(rental.MonthlyPrice*12) {
				t.Fatalf("rental %q is mispriced: cost=%v price=%v annual=%v", rental.Name, rental.MonthlyCost, rental.MonthlyPrice, rental.AnnualPrice)
			}
			for _, line := range rental.LineItems {
				if line.SourceID == nil || line.UnitPrice <= line.UnitCost {
					t.Fatalf("rental %q line %q is not linked or not marked up", rental.Name, line.SKU)
				}
				if line.Source == "product" {
					var product models.Product
					must(t, db.First(&product, "id = ?", *line.SourceID).Error)
					if line.UnitCost >= product.LandedCostSAR {
						t.Fatalf("rental %q charges the full purchase cost for %s monthly", rental.Name, line.SKU)
					}
				}
			}
		}

		// The quote builder's catalog endpoint must expose the seeded items.
		req := httptest.NewRequest("GET", "/api/v1/catalog", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("catalog returned %d: %s", w.Code, w.Body.String())
		}
	})
}
