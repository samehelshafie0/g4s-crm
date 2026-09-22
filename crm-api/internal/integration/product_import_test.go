package integration

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"g4s-crm/api/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func testProductImport(t *testing.T, db *gorm.DB, r *gin.Engine, token string) {
	t.Run("vendor file import creates, reprices and audits catalog products", func(t *testing.T) {
		post := func(payload any, want int) map[string]any {
			t.Helper()
			body, _ := json.Marshal(payload)
			req := httptest.NewRequest("POST", "/api/v1/products/import", bytes.NewReader(body))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != want {
				t.Fatalf("import returned %d, want %d: %s", w.Code, want, w.Body.String())
			}
			var out map[string]any
			_ = json.Unmarshal(w.Body.Bytes(), &out)
			return out
		}

		base := map[string]any{
			"vendorName": "Ingram Micro", "source": "vendor-catalog", "sourceRef": "price-list.xlsx",
			"productType": "import", "originCurrency": "USD", "fxRate": 3.75,
			"freightPercent": 4.0, "customsPercent": 5.0, "clearancePercent": 1.5,
			"targetMarginPercent": 30.0, "updateExisting": true,
		}
		withRows := func(rows []map[string]any) map[string]any {
			payload := map[string]any{}
			for k, v := range base {
				payload[k] = v
			}
			payload["rows"] = rows
			return payload
		}

		res := post(withRows([]map[string]any{
			{"sku": "IMP-TEST-1", "name": "Imported camera", "unitCost": 100.0, "qty": 5},
			{"sku": "IMP-TEST-2", "name": "Imported recorder", "unitCost": 200.0},
			{"sku": "IMP-TEST-1", "name": "Duplicate line", "unitCost": 999.0},
		}), 201)
		data := res["data"].(map[string]any)
		if data["created"].(float64) != 2 || data["skipped"].(float64) != 1 {
			t.Fatalf("expected 2 created and 1 duplicate skipped, got %v", data)
		}

		var product models.Product
		must(t, db.First(&product, "sku = ?", "IMP-TEST-1").Error)
		// 100 USD × 3.75 = 375 SAR, plus 10.5% landing = 414.375, margin 30% of the sell price.
		if models.RoundMoney(product.LandedCostSAR) != 414.38 {
			t.Fatalf("landed cost was %v, want 414.38", product.LandedCostSAR)
		}
		if models.RoundMoney(product.SellingPrice) != 591.96 {
			t.Fatalf("selling price was %v, want 591.96", product.SellingPrice)
		}
		if product.SupplierName != "Ingram Micro" || !product.IsActive {
			t.Fatalf("supplier or active flag not set: %+v", product)
		}

		var vendors []models.ProductVendorEntry
		must(t, db.Where("product_id = ?", product.ID).Find(&vendors).Error)
		if len(vendors) != 1 || vendors[0].UnitCost != 100 || vendors[0].CatalogSource != "price-list.xlsx" {
			t.Fatalf("vendor entry not recorded from the file: %+v", vendors)
		}
		var records []models.ProductPriceRecord
		must(t, db.Where("product_id = ?", product.ID).Find(&records).Error)
		if len(records) != 1 || records[0].SourceRef != "price-list.xlsx" {
			t.Fatalf("price history not recorded: %+v", records)
		}

		// A second run at a new cost repriced the product without duplicating it.
		res = post(withRows([]map[string]any{{"sku": "IMP-TEST-1", "name": "Imported camera", "unitCost": 120.0}}), 201)
		data = res["data"].(map[string]any)
		if data["updated"].(float64) != 1 || data["created"].(float64) != 0 {
			t.Fatalf("expected an update, got %v", data)
		}
		must(t, db.First(&product, "sku = ?", "IMP-TEST-1").Error)
		if models.RoundMoney(product.UnitCostOrigin) != 120 {
			t.Fatalf("cost was not refreshed: %v", product.UnitCostOrigin)
		}
		must(t, db.Where("product_id = ?", product.ID).Find(&records).Error)
		if len(records) != 2 {
			t.Fatalf("a changed cost must add one price record, got %d", len(records))
		}
		var count int64
		must(t, db.Model(&models.Product{}).Where("sku = ?", "IMP-TEST-1").Count(&count).Error)
		if count != 1 {
			t.Fatalf("re-import duplicated the product: %d rows", count)
		}

		// Without updateExisting the known SKU is reported rather than silently changed.
		payload := withRows([]map[string]any{{"sku": "IMP-TEST-1", "name": "Imported camera", "unitCost": 500.0}})
		payload["updateExisting"] = false
		payload["rows"] = append(payload["rows"].([]map[string]any), map[string]any{"sku": "IMP-TEST-3", "name": "New line", "unitCost": 50.0})
		res = post(payload, 201)
		data = res["data"].(map[string]any)
		if data["skipped"].(float64) != 1 || data["created"].(float64) != 1 {
			t.Fatalf("expected the known SKU skipped and the new one created, got %v", data)
		}
		must(t, db.First(&product, "sku = ?", "IMP-TEST-1").Error)
		if models.RoundMoney(product.UnitCostOrigin) != 120 {
			t.Fatalf("a skipped row must not change the product: %v", product.UnitCostOrigin)
		}

		// A row with no part number is rejected rather than creating a nameless product.
		post(withRows([]map[string]any{{"sku": "", "name": "No part number", "unitCost": 10.0}}), 400)
	})
}
