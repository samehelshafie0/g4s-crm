package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"g4s-crm/api/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"mime/multipart"
	"net/http/httptest"
	"sync"
	"testing"
)

func testModuleAPIs(t *testing.T, db *gorm.DB, r *gin.Engine, token, customerID string) {
	call := func(t *testing.T, method, path string, payload any, want int) map[string]any {
		t.Helper()
		body := []byte{}
		if payload != nil {
			var err error
			body, err = json.Marshal(payload)
			must(t, err)
		}
		req := httptest.NewRequest(method, "/api/v1"+path, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != want {
			t.Fatalf("%s %s: %d want %d: %s", method, path, w.Code, want, w.Body.String())
		}
		res := map[string]any{}
		if w.Body.Len() > 0 {
			must(t, json.Unmarshal(w.Body.Bytes(), &res))
		}
		if data, ok := res["data"].(map[string]any); ok {
			return data
		}
		return res
	}
	m := func(pairs ...any) map[string]any {
		result := map[string]any{}
		for i := 0; i < len(pairs); i += 2 {
			result[pairs[i].(string)] = pairs[i+1]
		}
		return result
	}
	var productID, mfrID, quoteID string
	t.Run("catalog nested writes and relationships", func(t *testing.T) {
		manufacturer := call(t, "POST", "/manufacturers", m("name", "Integration Manufacturer", "code", "INTEGRATION", "vendorType", "both"), 201)
		mfrID = manufacturer["id"].(string)
		category := call(t, "POST", "/manufacturers/"+mfrID+"/categories", m("name", "Cameras"), 201)
		product := call(t, "POST", "/products", m("sku", "TEST-CAMERA", "name", "Integration Camera", "manufacturerId", mfrID, "categoryId", category["id"], "unitCostOrigin", 25, "sellingPrice", 100), 201)
		productID = product["id"].(string)
		if product["manufacturerName"] != "Integration Manufacturer" || product["landedCostSAR"] != float64(25) {
			t.Fatal(product)
		}
		call(t, "PATCH", "/products/"+productID, m("total", 12), 400)
		call(t, "DELETE", "/manufacturers/"+mfrID+"/categories/"+category["id"].(string), nil, 409)
	})
	t.Run("contacts sites and assigned opportunities persist", func(t *testing.T) {
		site := call(t, "POST", "/customers/"+customerID+"/sites", m("name", "Main site", "city", "Riyadh"), 201)
		first := call(t, "POST", "/customers/"+customerID+"/contacts", m("name", "First contact", "isPrimary", true, "siteId", site["id"]), 201)
		call(t, "POST", "/customers/"+customerID+"/contacts", m("name", "Second contact", "isPrimary", true), 201)
		customer := call(t, "GET", "/customers/"+customerID, nil, 200)
		primary := 0
		for _, raw := range customer["contacts"].([]any) {
			if raw.(map[string]any)["isPrimary"] == true {
				primary++
			}
		}
		if primary != 1 {
			t.Fatal(customer)
		}
		call(t, "DELETE", "/customers/"+customerID+"/sites/"+site["id"].(string), nil, 204)
		call(t, "PATCH", "/customers/"+customerID+"/contacts/"+first["id"].(string), m("siteId", site["id"]), 400)
		user := call(t, "GET", "/auth/me", nil, 200)
		opp := call(t, "POST", "/opportunities", m("title", "API opportunity", "customerId", customerID, "salesExecutiveId", user["id"], "estimatedValue", 1000, "estimatedCost", 500, "serviceTypes", []string{"cctv"}, "expectedCloseDate", "2027-03-01"), 201)
		if opp["salesExecutiveName"] != "Test Admin" || opp["estimatedMargin"] != float64(50) {
			t.Fatal(opp)
		}
		call(t, "PATCH", "/opportunities/"+opp["id"].(string), m("notes", "Retain owners"), 200)
	})
	t.Run("builder save stale edit lifecycle revision and conversion", func(t *testing.T) {
		service := call(t, "POST", "/services", m("sku", "SVC-TEST", "name", "Engineering", "unitCost", 10, "unitPrice", 30), 201)
		quote := call(t, "POST", "/quotes", m("customerId", customerID, "vatPercent", 0), 201)
		quoteID = quote["id"].(string)
		rows := []any{m("rowType", "heading", "source", "write-in", "multiplier", 1, "headingText", "Equipment", "isPrintable", true), m("rowType", "item", "source", "product", "productId", productID, "description", "Camera", "quantity", 2.5, "multiplier", 2, "unitCost", 25, "unitPrice", 100, "discountPercent", 10, "isSelected", true, "isPrintable", true), m("rowType", "item", "source", "service", "serviceId", service["id"], "description", "Optional engineering", "quantity", 10, "multiplier", 1, "unitCost", 10, "unitPrice", 30, "isOptional", true, "isSelected", false, "isPrintable", false)}
		payload := m("customerId", customerID, "lockVersion", quote["lockVersion"], "currency", "SAR", "vatPercent", 0, "discountPercent", 10, "rows", rows, "soldTo", m("company", "Saved address"), "paymentTerms", "Net 60")
		saved := call(t, "PUT", "/quotes/"+quoteID+"/builder", payload, 200)
		if saved["total"] != float64(405) || saved["totalCost"] != float64(125) || saved["paymentTerms"] != "Net 60" {
			t.Fatal(saved)
		}
		call(t, "PUT", "/quotes/"+quoteID+"/builder", payload, 409)
		reload := call(t, "GET", "/quotes/"+quoteID+"/builder", nil, 200)
		if len(reload["lineItems"].([]any)) != 3 {
			t.Fatal(reload)
		}
		call(t, "PATCH", "/quotes/"+quoteID+"/submit", m(), 200)
		submitted := call(t, "GET", "/quotes/"+quoteID, nil, 200)
		if submitted["status"] == "pending-approval" {
			call(t, "PATCH", "/quotes/"+quoteID+"/approve", m(), 200)
		}
		call(t, "PUT", "/quotes/"+quoteID+"/builder", payload, 422)
		call(t, "PATCH", "/quotes/"+quoteID+"/send", m(), 200)
		call(t, "PATCH", "/quotes/"+quoteID+"/accept", m(), 200)
		revision := call(t, "POST", "/quotes/"+quoteID+"/revisions", m(), 201)
		if revision["status"] != "draft" || revision["vatPercent"] != float64(0) {
			t.Fatal(revision)
		}
		contractInput := m("title", "Converted contract", "type", "sales", "startDate", "2026-09-16", "endDate", "2027-09-16")
		contract := call(t, "POST", "/quotes/"+quoteID+"/convert-to-contract", contractInput, 201)
		duplicate := call(t, "POST", "/quotes/"+quoteID+"/convert-to-contract", contractInput, 200)
		if duplicate["id"] != contract["id"] {
			t.Fatal(duplicate)
		}
		path := "/contracts/" + contract["id"].(string)
		call(t, "PATCH", path+"/activate", m(), 200)
		renewal := call(t, "POST", path+"/renew", m(), 201)
		again := call(t, "POST", path+"/renew", m(), 200)
		if renewal["id"] != again["id"] {
			t.Fatal(again)
		}
	})
	t.Run("procurement lines conversion partial receipts and stock", func(t *testing.T) {
		sq := call(t, "POST", "/procurement/supplier-quotes", m("supplierId", mfrID, "supplierName", "Integration Manufacturer", "currency", "SAR", "items", []any{m("productId", productID, "productName", "Camera", "quantity", 10, "unitCost", 25)}), 201)
		sqPath := "/procurement/supplier-quotes/" + sq["id"].(string)
		call(t, "PATCH", sqPath+"/status", m("status", "accepted"), 200)
		po := call(t, "POST", sqPath+"/convert-to-po", m(), 201)
		again := call(t, "POST", sqPath+"/convert-to-po", m(), 200)
		if again["id"] != po["id"] || po["total"] != float64(250) {
			t.Fatal(po, again)
		}
		poPath := "/procurement/purchase-orders/" + po["id"].(string)
		call(t, "PATCH", poPath+"/status", m("status", "pending-approval"), 200)
		call(t, "PATCH", poPath+"/approve", m(), 200)
		call(t, "PATCH", poPath+"/status", m("status", "ordered"), 200)
		line := po["items"].([]any)[0].(map[string]any)
		receipt := m("poId", po["id"], "receiveDate", "2026-09-16", "items", []any{m("poItemId", line["id"], "productId", productID, "receivedQty", 6, "storageLocation", "riyadh-main", "condition", "good")})
		call(t, "POST", "/procurement/goods-receipts", receipt, 201)
		partial := call(t, "GET", poPath, nil, 200)
		if partial["status"] != "partial-received" {
			t.Fatal(partial)
		}
		call(t, "POST", "/procurement/goods-receipts", receipt, 400)
		receipt["items"] = []any{m("poItemId", line["id"], "productId", productID, "receivedQty", 4, "storageLocation", "riyadh-main", "condition", "good")}
		call(t, "POST", "/procurement/goods-receipts", receipt, 201)
		completed := call(t, "GET", poPath, nil, 200)
		if completed["status"] != "received" {
			t.Fatal(completed)
		}
		call(t, "POST", "/inventory/movements/transfer", m("productId", productID, "qty", 3, "fromWarehouse", "riyadh-main", "toWarehouse", "jeddah-branch", "reason", "Project delivery"), 200)
		call(t, "POST", "/inventory/movements/transfer", m("productId", productID, "qty", 99, "fromWarehouse", "riyadh-main", "toWarehouse", "jeddah-branch", "reason", "Too many"), 400)
		reserve := call(t, "POST", "/inventory/reservations", m("productId", productID, "warehouseLocation", "riyadh-main", "qty", 5, "source", "manual"), 201)
		call(t, "POST", "/inventory/reservations", m("productId", productID, "warehouseLocation", "riyadh-main", "qty", 5, "source", "manual"), 400)
		call(t, "PATCH", "/inventory/reservations/"+reserve["id"].(string)+"/release", m("reason", "No longer needed"), 200)
		call(t, "PATCH", "/inventory/reservations/"+reserve["id"].(string)+"/release", m("reason", "Repeat"), 400)
		reserve = call(t, "POST", "/inventory/reservations", m("productId", productID, "warehouseLocation", "riyadh-main", "qty", 2, "source", "manual"), 201)
		call(t, "PATCH", "/inventory/reservations/"+reserve["id"].(string)+"/fulfill", m("reason", "Delivered"), 200)
		var stocks []models.WarehouseStock
		must(t, db.Where("product_id = ?", productID).Find(&stocks).Error)
		sum := 0
		for _, stock := range stocks {
			sum += stock.OnHandQty
			if stock.AvailableQty != stock.OnHandQty-stock.ReservedQty {
				t.Fatal(stock)
			}
		}
		if sum != 8 {
			t.Fatal(stocks)
		}
	})
	t.Run("concurrent reservations cannot oversell", func(t *testing.T) {
		body := fmt.Sprintf(`{"productId":%q,"warehouseLocation":"jeddah-branch","qty":1,"source":"manual"}`, productID)
		codes := make(chan int, 8)
		var wg sync.WaitGroup
		for i := 0; i < 8; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				req := httptest.NewRequest("POST", "/api/v1/inventory/reservations", bytes.NewBufferString(body))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+token)
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				codes <- w.Code
			}()
		}
		wg.Wait()
		close(codes)
		success := 0
		for code := range codes {
			if code == 201 {
				success++
			} else if code != 400 {
				t.Fatalf("Unexpected status: %d", code)
			}
		}
		if success != 3 {
			t.Fatalf("%d reservations; available was 3", success)
		}
	})
	t.Run("teams recurring FX entries projects and protected user fields", func(t *testing.T) {
		user := call(t, "GET", "/auth/me", nil, 200)
		call(t, "PATCH", "/users/"+user["id"].(string), m("passwordHash", "attacker"), 400)
		call(t, "DELETE", "/users/"+user["id"].(string), nil, 400)
		team := call(t, "POST", "/teams", m("name", "API Team", "department", "sales", "leaderId", user["id"]), 201)
		if team["leaderName"] != "Test Admin" || len(team["members"].([]any)) != 1 {
			t.Fatal(team)
		}
		recurring := call(t, "POST", "/recurring-services", m("name", "Maintenance", "monthlyCost", 50, "monthlyPrice", 100), 201)
		if recurring["annualPrice"] != float64(1200) {
			t.Fatal(recurring)
		}
		updated := call(t, "PATCH", "/recurring-services/"+recurring["id"].(string), m("monthlyPrice", 150), 200)
		if updated["annualPrice"] != float64(1800) {
			t.Fatal(updated)
		}
		// The initial migration seeds common FX pairs; create an unused pair.
		fx := call(t, "POST", "/exchange-rates", m("fromCurrency", "EUR", "toCurrency", "USD", "currentRate", 1.1, "effectiveDate", "2026-09-16"), 201)
		fx = call(t, "PATCH", "/exchange-rates/"+fx["id"].(string), m("currentRate", 1.2, "effectiveDate", "2026-09-17"), 200)
		if len(fx["history"].([]any)) != 2 || fx["currentRate"] != 1.2 {
			t.Fatal(fx)
		}
		book := call(t, "POST", "/price-books", m("name", "Customer pricing", "type", "customer-specific", "customerId", customerID), 201)
		call(t, "POST", "/price-books/"+book["id"].(string)+"/entries", m("productId", productID, "customPrice", 90), 200)
		project := call(t, "POST", "/projects", m("name", "Delivery", "customerId", customerID, "quoteId", quoteID, "projectManagerId", user["id"]), 201)
		if project["totalValue"] != float64(405) || project["projectManager"] != "Test Admin" {
			t.Fatal(project)
		}
	})
	t.Run("mixed price books recurring components inactive flags and reporting", func(t *testing.T) {
		service := call(t, "POST", "/services", m("sku", "SVC-MIXED", "name", "Installation", "unitCost", 10, "unitPrice", 30), 201)
		recurring := call(t, "POST", "/recurring-services", m("name", "Component costing", "monthlyPrice", 100, "isActive", false, "lineItems", []any{m("source", "product", "sourceId", productID, "name", "Camera rental", "qty", 2, "unitCost", 12.5, "unitPrice", 50)}), 201)
		if recurring["isActive"] != false || recurring["monthlyCost"] != float64(25) || recurring["annualCost"] != float64(300) {
			t.Fatal(recurring)
		}
		book := call(t, "POST", "/price-books", m("name", "Mixed pricing", "type", "standard"), 201)
		path := "/price-books/" + book["id"].(string)
		entries := []any{m("kind", "product", "productId", productID, "customPrice", 90), m("kind", "service", "serviceId", service["id"], "customPrice", 25), m("kind", "recurring", "recurringServiceId", recurring["id"], "customPrice", 80)}
		saved := call(t, "PUT", path+"/entries", m("entries", entries), 200)
		if len(saved["entries"].([]any)) != 3 {
			t.Fatal(saved)
		}
		call(t, "PUT", path+"/entries", m("entries", []any{entries[0], entries[0]}), 400)
		reload := call(t, "GET", path, nil, 200)
		if len(reload["entries"].([]any)) != 3 {
			t.Fatal("Invalid replacement removed entries")
		}
		contract := call(t, "POST", "/contracts", m("title", "Notice period", "customerId", customerID, "type", "service", "renewalNoticeDays", 0), 201)
		if contract["renewalNoticeDays"] != float64(0) {
			t.Fatal(contract)
		}
		for _, endpoint := range []string{"/catalog", "/dashboard/kpis", "/dashboard/recent-quotes", "/dashboard/expiring-quotes", "/dashboard/sales-performance", "/dashboard/top-customers"} {
			call(t, "GET", endpoint, nil, 200)
		}
		team := call(t, "POST", "/teams", m("name", "Disposable team", "department", "sales"), 201)
		call(t, "DELETE", "/teams/"+team["id"].(string), nil, 204)
	})
	t.Run("foreign currency receipts value warehouse stock in SAR", func(t *testing.T) {
		call(t, "POST", "/exchange-rates", m("fromCurrency", "CNY", "toCurrency", "SAR", "currentRate", 0.5, "effectiveDate", "2026-01-01"), 201)
		product := call(t, "POST", "/products", m("sku", "FX-RECEIPT", "name", "Imported camera", "productType", "import"), 201)
		po := call(t, "POST", "/procurement/purchase-orders", m("supplierName", "Overseas supplier", "currency", "CNY", "shippingCost", 20, "customsDuty", 10, "items", []any{m("productId", product["id"], "quantity", 2, "unitCost", 100)}), 201)
		path := "/procurement/purchase-orders/" + po["id"].(string)
		call(t, "PATCH", path+"/status", m("status", "pending-approval"), 200)
		call(t, "PATCH", path+"/approve", m(), 200)
		call(t, "PATCH", path+"/status", m("status", "ordered"), 200)
		line := po["items"].([]any)[0].(map[string]any)
		gr := call(t, "POST", "/procurement/goods-receipts", m("poId", po["id"], "receiveDate", "2026-09-16", "items", []any{m("poItemId", line["id"], "productId", product["id"], "receivedQty", 2, "storageLocation", "riyadh-main", "condition", "good")}), 201)
		if gr["fxRate"] != 0.5 || gr["totalLandingCost"] != float64(115) {
			t.Fatal(gr)
		}
		var stock models.WarehouseStock
		must(t, db.First(&stock, "product_id = ?", product["id"]).Error)
		if stock.OnHandQty != 2 || stock.UnitCost != 57.5 || stock.TotalValue != 115 {
			t.Fatal(stock)
		}
	})
	t.Run("frontend collection routes return arrays", func(t *testing.T) {
		for _, path := range []string{"/customers", "/customers/lookup", "/opportunities", "/quotes", "/contracts", "/projects", "/manufacturers", "/products", "/services", "/price-books", "/recurring-services", "/exchange-rates", "/teams", "/users/lookup", "/documents", "/inventory/stock", "/inventory/reservations", "/inventory/movements", "/procurement/purchase-orders", "/procurement/supplier-quotes", "/procurement/goods-receipts", "/procurement/supplier-items"} {
			result := call(t, "GET", path, nil, 200)
			if _, ok := result["data"].([]any); !ok {
				t.Fatalf("%s is not a collection: %#v", path, result)
			}
		}
	})
	t.Run("documents versioned upload authenticated download and containment", func(t *testing.T) {
		upload := func(path, category, content string, want int) map[string]any {
			t.Helper()
			var body bytes.Buffer
			form := multipart.NewWriter(&body)
			must(t, form.WriteField("category", category))
			part, err := form.CreateFormFile("file", "test.txt")
			must(t, err)
			_, err = part.Write([]byte(content))
			must(t, err)
			must(t, form.Close())
			req := httptest.NewRequest("POST", "/api/v1"+path, &body)
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", form.FormDataContentType())
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != want {
				t.Fatalf("upload: %d %s", w.Code, w.Body.String())
			}
			var res map[string]any
			must(t, json.Unmarshal(w.Body.Bytes(), &res))
			if data, ok := res["data"].(map[string]any); ok {
				return data
			}
			return res
		}
		upload("/documents", "../../outside", "blocked", 400)
		doc := upload("/documents", "general", "Version one", 201)
		if doc["filePath"] != nil {
			t.Fatal("Storage path exposed")
		}
		path := "/documents/" + doc["id"].(string)
		version := upload(path+"/versions", "general", "Version two", 201)
		if version["version"] != "2" {
			t.Fatal(version)
		}
		link := call(t, "POST", path+"/links", m("entityType", "customer", "entityId", customerID), 201)
		if link["entityName"] == "" {
			t.Fatal("Missing linked customer name")
		}
		call(t, "POST", path+"/links", m("entityType", "quote", "entityId", quoteID), 201)
		history := call(t, "GET", path+"/versions", nil, 200)["data"].([]any)
		if len(history) != 2 {
			t.Fatal(history)
		}
		for _, entry := range history {
			snapshot := entry.(map[string]any)
			req := httptest.NewRequest("GET", "/api/v1"+path+"/versions/"+snapshot["id"].(string)+"/download", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			expected := "Version one"
			if snapshot["version"] == "2" {
				expected = "Version two"
			}
			if w.Code != 200 || w.Body.String() != expected {
				t.Fatal(w.Code, w.Body.String(), snapshot)
			}
		}
		updated := call(t, "PATCH", path, m("name", "Acceptance document", "tags", []string{"reviewed"}), 200)
		if updated["name"] != "Acceptance document" {
			t.Fatal(updated)
		}
		call(t, "DELETE", path+"/links/"+link["id"].(string), nil, 204)
		linked := call(t, "GET", path, nil, 200)["links"].([]any)
		if len(linked) != 1 || linked[0].(map[string]any)["entityType"] != "quote" {
			t.Fatal(linked)
		}
		req := httptest.NewRequest("GET", "/api/v1"+path+"/download", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != 200 || w.Body.String() != "Version two" {
			t.Fatal(w.Code, w.Body.String())
		}
	})
	t.Run("catalog service administration persists active state and deletion", func(t *testing.T) {
		service := call(t, "POST", "/services", m("sku", "ADMIN-SVC", "name", "Admin service", "unitCost", 10, "unitPrice", 25, "isActive", true), 201)
		path := "/services/" + service["id"].(string)
		call(t, "PATCH", path, m("name", "Updated service", "unitPrice", 35, "isActive", false), 200)
		saved := call(t, "GET", path, nil, 200)
		if saved["name"] != "Updated service" || saved["unitPrice"] != float64(35) || saved["isActive"] != false {
			t.Fatal(saved)
		}
		call(t, "DELETE", path, nil, 204)
		call(t, "GET", path, nil, 404)
	})

}
