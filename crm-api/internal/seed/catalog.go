// Package seed loads reference catalog data (manufacturers, products, services,
// recurring services and exchange rates) so quotes can be built on a fresh
// database. It is idempotent: records are matched by manufacturer code, product
// SKU, service SKU, recurring-service name or currency pair and updated in place,
// so it can be re-run after editing catalog.json without creating duplicates.
package seed

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"g4s-crm/api/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

//go:embed catalog.json
var catalogJSON []byte

type catalogFile struct {
	ExchangeRates []struct {
		Currency models.Currency `json:"currency"`
		Rate     float64         `json:"rate"`
	} `json:"exchangeRates"`
	Manufacturers []struct {
		Code       string            `json:"code"`
		Name       string            `json:"name"`
		Country    string            `json:"country"`
		Website    string            `json:"website"`
		VendorType models.VendorType `json:"vendorType"`
		Categories []string          `json:"categories"`
	} `json:"manufacturers"`
	Products []struct {
		SKU                 string             `json:"sku"`
		Name                string             `json:"name"`
		Description         string             `json:"description"`
		Manufacturer        string             `json:"manufacturer"`
		Category            string             `json:"category"`
		ProductType         models.ProductType `json:"productType"`
		OriginCurrency      models.Currency    `json:"originCurrency"`
		UnitCostOrigin      float64            `json:"unitCostOrigin"`
		FreightPercent      float64            `json:"freightPercent"`
		CustomsPercent      float64            `json:"customsPercent"`
		ClearancePercent    float64            `json:"clearancePercent"`
		TargetMarginPercent float64            `json:"targetMarginPercent"`
		LeadTimeDays        int                `json:"leadTimeDays"`
		SupplierName        string             `json:"supplierName"`
	} `json:"products"`
	Services []struct {
		SKU         string            `json:"sku"`
		Name        string            `json:"name"`
		Description string            `json:"description"`
		Department  models.Department `json:"department"`
		RateType    string            `json:"rateType"`
		UnitCost    float64           `json:"unitCost"`
		UnitPrice   float64           `json:"unitPrice"`
	} `json:"services"`
	Rentals []struct {
		Name                string                  `json:"name"`
		Description         string                  `json:"description"`
		BillingFrequency    models.BillingFrequency `json:"billingFrequency"`
		TermMonths          float64                 `json:"termMonths"`
		TargetMarginPercent float64                 `json:"targetMarginPercent"`
		Lines               []struct {
			SKU string  `json:"sku"`
			Qty float64 `json:"qty"`
		} `json:"lines"`
		Services []struct {
			SKU string  `json:"sku"`
			Qty float64 `json:"qty"`
		} `json:"services"`
	} `json:"rentals"`
	Recurring []struct {
		Name                string                      `json:"name"`
		ServiceType         models.RecurringServiceType `json:"serviceType"`
		Description         string                      `json:"description"`
		BillingFrequency    models.BillingFrequency     `json:"billingFrequency"`
		MonthlyCost         float64                     `json:"monthlyCost"`
		MonthlyPrice        float64                     `json:"monthlyPrice"`
		TargetMarginPercent float64                     `json:"targetMarginPercent"`
	} `json:"recurring"`
}

// Summary reports what the seed run created or updated.
type Summary struct {
	ExchangeRates, Manufacturers, Categories, Products, Services, Recurring, Rentals int
}

// Catalog upserts the embedded catalog inside one transaction.
func Catalog(db *gorm.DB) (Summary, error) {
	var file catalogFile
	if err := json.Unmarshal(catalogJSON, &file); err != nil {
		return Summary{}, fmt.Errorf("catalog.json: %w", err)
	}
	var summary Summary
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(473246533)").Error; err != nil {
			return err
		}
		now := time.Now()

		// Exchange rates: existing pairs are left alone so operator updates survive re-seeding.
		rates := map[models.Currency]float64{models.CurrencySAR: 1}
		for _, r := range file.ExchangeRates {
			var existing models.ExchangeRate
			err := tx.Where("from_currency = ? AND to_currency = 'SAR'", r.Currency).First(&existing).Error
			switch {
			case err == nil:
				rates[r.Currency] = existing.CurrentRate
			case err == gorm.ErrRecordNotFound:
				rate := models.ExchangeRate{FromCurrency: r.Currency, ToCurrency: "SAR", CurrentRate: r.Rate, EffectiveDate: now}
				if err := tx.Create(&rate).Error; err != nil {
					return err
				}
				if err := tx.Create(&models.ExchangeRateHistory{ExchangeRateID: rate.ID, Rate: r.Rate, EffectiveDate: now}).Error; err != nil {
					return err
				}
				rates[r.Currency] = r.Rate
				summary.ExchangeRates++
			default:
				return err
			}
		}

		manufacturers := map[string]uuid.UUID{}
		categories := map[string]uuid.UUID{} // "CODE/Category name"
		for _, m := range file.Manufacturers {
			var item models.Manufacturer
			err := tx.Where("code = ?", m.Code).First(&item).Error
			if err != nil && err != gorm.ErrRecordNotFound {
				return err
			}
			item.Code, item.Name, item.Country, item.Website, item.VendorType, item.IsActive = m.Code, m.Name, m.Country, m.Website, m.VendorType, true
			if err := tx.Save(&item).Error; err != nil {
				return err
			}
			manufacturers[m.Code] = item.ID
			summary.Manufacturers++
			for _, name := range m.Categories {
				var category models.ManufacturerCategory
				err := tx.Where("manufacturer_id = ? AND name = ?", item.ID, name).First(&category).Error
				if err == gorm.ErrRecordNotFound {
					category = models.ManufacturerCategory{ManufacturerID: item.ID, Name: name}
					if err := tx.Create(&category).Error; err != nil {
						return err
					}
					summary.Categories++
				} else if err != nil {
					return err
				}
				categories[m.Code+"/"+name] = category.ID
			}
		}

		for _, p := range file.Products {
			manufacturerID, ok := manufacturers[p.Manufacturer]
			if !ok {
				return fmt.Errorf("product %s references unknown manufacturer %q", p.SKU, p.Manufacturer)
			}
			categoryID, ok := categories[p.Manufacturer+"/"+p.Category]
			if !ok {
				return fmt.Errorf("product %s references unknown category %q for %s", p.SKU, p.Category, p.Manufacturer)
			}
			fx, ok := rates[p.OriginCurrency]
			if !ok {
				return fmt.Errorf("product %s uses currency %s with no exchange rate", p.SKU, p.OriginCurrency)
			}
			var item models.Product
			created := false
			err := tx.Where("sku = ?", p.SKU).First(&item).Error
			if err == gorm.ErrRecordNotFound {
				created = true
			} else if err != nil {
				return err
			}
			item.SKU, item.Name, item.Description = p.SKU, p.Name, p.Description
			item.ManufacturerID, item.CategoryID = &manufacturerID, &categoryID
			item.ProductType, item.OriginCurrency, item.UnitCostOrigin, item.FXRate = p.ProductType, p.OriginCurrency, p.UnitCostOrigin, fx
			item.FreightPercent, item.CustomsPercent, item.ClearancePercent = p.FreightPercent, p.CustomsPercent, p.ClearancePercent
			item.TargetMarginPercent, item.LeadTimeDays, item.SupplierName, item.IsActive = p.TargetMarginPercent, p.LeadTimeDays, p.SupplierName, true
			item.RecalculateCosts()
			// Same formula as the product editor: margin is a share of the selling price.
			item.SellingPrice = math.Round(item.LandedCostSAR/(1-p.TargetMarginPercent/100)*100) / 100
			item.RecalculateCosts()
			if err := tx.Save(&item).Error; err != nil {
				return err
			}
			summary.Products++
			if created {
				// One vendor line and one catalog price record so the quote builder's
				// price history and the product's vendor tab are not empty.
				if err := tx.Create(&models.ProductVendorEntry{ProductID: item.ID, VendorName: p.SupplierName, VendorSKU: p.SKU, UnitCost: p.UnitCostOrigin, Currency: p.OriginCurrency, MOQ: 1, LeadTimeDays: p.LeadTimeDays, LastQuoteDate: &now, CatalogSource: "seed catalog"}).Error; err != nil {
					return err
				}
				if err := tx.Create(&models.ProductPriceRecord{ProductID: item.ID, Date: now, Source: models.PriceSourceVendorCatalog, SourceRef: "seed catalog", VendorName: p.SupplierName, UnitCost: p.UnitCostOrigin, Currency: p.OriginCurrency, LandingCost: item.LandedCostSAR, Qty: 1}).Error; err != nil {
					return err
				}
			}
		}

		serviceLines := map[string]models.CatalogService{}
		for _, s := range file.Services {
			var item models.CatalogService
			err := tx.Where("sku = ?", s.SKU).First(&item).Error
			if err != nil && err != gorm.ErrRecordNotFound {
				return err
			}
			item.SKU, item.Name, item.Description, item.Department, item.RateType = s.SKU, s.Name, s.Description, s.Department, s.RateType
			item.UnitCost, item.UnitPrice, item.IsActive = s.UnitCost, s.UnitPrice, true
			if err := tx.Save(&item).Error; err != nil {
				return err
			}
			summary.Services++
			serviceLines[s.SKU] = item
		}

		// Rentals: the customer pays monthly instead of buying. Each line's monthly
		// cost is the landed hardware cost spread over the contract term; the monthly
		// price adds the target margin. Attached service lines (maintenance visits)
		// are already monthly figures and are not amortised.
		for _, r := range file.Rentals {
			if r.TermMonths <= 0 {
				return fmt.Errorf("rental %q needs a positive termMonths", r.Name)
			}
			lines := []models.RecurringLine{}
			var monthlyCost float64
			for i, l := range r.Lines {
				var product models.Product
				if err := tx.Where("sku = ?", l.SKU).First(&product).Error; err != nil {
					return fmt.Errorf("rental %q references unknown product %s: %w", r.Name, l.SKU, err)
				}
				id := product.ID
				unitCost := models.RoundMoney(product.LandedCostSAR / r.TermMonths)
				unitPrice := models.RoundMoney(unitCost / (1 - r.TargetMarginPercent/100))
				lines = append(lines, models.RecurringLine{
					ID: fmt.Sprintf("rental-%d", i+1), Source: "product", SourceID: &id,
					SKU: product.SKU, Name: product.Name,
					Description: fmt.Sprintf("Rental, hardware amortised over %.0f months", r.TermMonths),
					Qty:         l.Qty, UnitCost: unitCost, UnitPrice: unitPrice,
				})
				monthlyCost += models.RoundMoney(l.Qty * unitCost)
			}
			for i, l := range r.Services {
				service, ok := serviceLines[l.SKU]
				if !ok {
					return fmt.Errorf("rental %q references unknown service %s", r.Name, l.SKU)
				}
				id := service.ID
				lines = append(lines, models.RecurringLine{
					ID: fmt.Sprintf("rental-svc-%d", i+1), Source: "write-in", SourceID: &id,
					SKU: service.SKU, Name: service.Name, Description: "Included maintenance, monthly share",
					Qty: l.Qty, UnitCost: service.UnitCost, UnitPrice: service.UnitPrice,
				})
				monthlyCost += models.RoundMoney(l.Qty * service.UnitCost)
			}
			monthlyCost = models.RoundMoney(monthlyCost)
			monthlyPrice := models.RoundMoney(monthlyCost / (1 - r.TargetMarginPercent/100))

			var item models.RecurringService
			err := tx.Where("name = ?", r.Name).First(&item).Error
			if err != nil && err != gorm.ErrRecordNotFound {
				return err
			}
			item.Name, item.ServiceType, item.Description = r.Name, models.RecurringRental, r.Description
			item.BillingFrequency, item.LineItems, item.IsActive = r.BillingFrequency, lines, true
			item.MonthlyCost, item.MonthlyPrice = monthlyCost, monthlyPrice
			item.AnnualCost, item.AnnualPrice = models.RoundMoney(monthlyCost*12), models.RoundMoney(monthlyPrice*12)
			item.TargetMarginPercent = r.TargetMarginPercent
			if err := tx.Omit(clause.Associations).Save(&item).Error; err != nil {
				return err
			}
			summary.Rentals++
		}

		for _, r := range file.Recurring {
			var item models.RecurringService
			err := tx.Where("name = ?", r.Name).First(&item).Error
			if err != nil && err != gorm.ErrRecordNotFound {
				return err
			}
			item.Name, item.ServiceType, item.Description, item.BillingFrequency = r.Name, r.ServiceType, r.Description, r.BillingFrequency
			item.MonthlyCost, item.MonthlyPrice, item.TargetMarginPercent, item.IsActive = r.MonthlyCost, r.MonthlyPrice, r.TargetMarginPercent, true
			item.AnnualCost, item.AnnualPrice = r.MonthlyCost*12, r.MonthlyPrice*12
			if item.LineItems == nil {
				item.LineItems = []models.RecurringLine{{ID: "seed-1", Source: "write-in", Name: r.Name, Qty: 1, UnitCost: r.MonthlyCost, UnitPrice: r.MonthlyPrice}}
			}
			if err := tx.Save(&item).Error; err != nil {
				return err
			}
			summary.Recurring++
		}
		return nil
	})
	return summary, err
}
