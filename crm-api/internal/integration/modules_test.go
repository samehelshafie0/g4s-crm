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
	"strings"
	"sync"
	"testing"

	"github.com/xuri/excelize/v2"
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
	var productID, mfrID, quoteID, oppID string
	t.Run("catalog nested writes and relationships", func(t *testing.T) {
		manufacturer := call(t, "POST", "/manufacturers", m("name", "Integration Manufacturer", "code", "INTEGRATION", "vendorType", "both"), 201)
		mfrID = manufacturer["id"].(string)
		category := call(t, "POST", "/manufacturers/"+mfrID+"/categories", m("name", "Cameras"), 201)
		call(t, "POST", "/manufacturers/"+mfrID+"/categories", m("name", " cameras "), 409)
		product := call(t, "POST", "/products", m("sku", "TEST-CAMERA", "name", "Integration Camera", "manufacturerId", mfrID, "categoryId", category["id"], "unitCostOrigin", 25, "sellingPrice", 100), 201)
		productID = product["id"].(string)
		if product["manufacturerName"] != "Integration Manufacturer" || product["landedCostSAR"] != float64(25) {
			t.Fatal(product)
		}
		call(t, "PATCH", "/products/"+productID, m("total", 12), 400)
		call(t, "DELETE", "/manufacturers/"+mfrID+"/categories/"+category["id"].(string), nil, 409)
	})
	t.Run("product import template lists the catalog's manufacturers and categories", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/products/import-template", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != 200 || !strings.Contains(w.Header().Get("Content-Disposition"), "g4s-product-import-") {
			t.Fatal(w.Code, w.Header())
		}
		book, err := excelize.OpenReader(bytes.NewReader(w.Body.Bytes()))
		must(t, err)
		if sheets := book.GetSheetList(); len(sheets) != 3 || sheets[0] != "Products" {
			t.Fatal(sheets)
		}
		if visible, _ := book.GetSheetVisible("Lists"); visible {
			t.Fatal("the lookup sheet should be hidden")
		}
		rows, err := book.GetRows("Lists")
		must(t, err)
		found := false
		for _, row := range rows {
			if len(row) >= 2 && row[0] == "Integration Manufacturer" && row[1] == "Cameras" {
				found = true
			}
		}
		if !found {
			t.Fatal("manufacturer/category pair missing from the template lists", rows)
		}
		validations, err := book.GetDataValidations("Products")
		must(t, err)
		ranges := map[string]string{}
		for _, dv := range validations {
			ranges[dv.Sqref] = dv.Formula1
		}
		if !strings.Contains(ranges["E2:E501"], "Lists!$D$2") || !strings.Contains(ranges["F2:F501"], "OFFSET(Lists!$B$1") {
			t.Fatal("manufacturer and category drop-downs missing", ranges)
		}
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
		oppID = opp["id"].(string)
	})
	t.Run("builder save stale edit lifecycle revision and conversion", func(t *testing.T) {
		service := call(t, "POST", "/services", m("sku", "SVC-TEST", "name", "Engineering", "unitCost", 10, "unitPrice", 30), 201)
		call(t, "POST", "/quotes", m("customerId", customerID, "vatPercent", 0), 400)
		quote := call(t, "POST", "/quotes", m("customerId", customerID, "opportunityId", oppID, "vatPercent", 0), 201)
		quoteID = quote["id"].(string)
		if quote["version"] != float64(1) || quote["isCurrent"] != true {
			t.Fatal(quote)
		}
		// One opportunity, one quote: a second is refused and names the first.
		call(t, "POST", "/quotes", m("customerId", customerID, "opportunityId", oppID), 409)
		rows := []any{m("rowType", "heading", "source", "write-in", "multiplier", 1, "headingText", "Equipment", "isPrintable", true), m("rowType", "item", "source", "product", "productId", productID, "description", "Camera", "quantity", 2.5, "multiplier", 2, "unitCost", 25, "unitPrice", 100, "discountPercent", 10, "isSelected", true, "isPrintable", true), m("rowType", "item", "source", "service", "serviceId", service["id"], "description", "Optional engineering", "quantity", 10, "multiplier", 1, "unitCost", 10, "unitPrice", 30, "isOptional", true, "isSelected", false, "isPrintable", false)}
		payload := m("customerId", customerID, "opportunityId", oppID, "lockVersion", quote["lockVersion"], "currency", "SAR", "vatPercent", 0, "discountPercent", 10, "rows", rows, "soldTo", m("company", "Saved address"), "paymentTerms", "Net 60")
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

		// Revising keeps the number, raises the version and freezes the old one.
		call(t, "POST", "/quotes/"+quoteID+"/revisions", m(), 400)
		revision := call(t, "POST", "/quotes/"+quoteID+"/revisions", m("reason", "Customer asked for a second camera"), 201)
		if revision["status"] != "draft" || revision["vatPercent"] != float64(0) || revision["version"] != float64(2) || revision["quoteNumber"] != quote["quoteNumber"] || revision["revisionReason"] != "Customer asked for a second camera" {
			t.Fatal(revision)
		}
		call(t, "PATCH", "/quotes/"+quoteID+"/accept", m(), 422)
		call(t, "POST", "/quotes/"+quoteID+"/revisions", m("reason", "again"), 422)
		versions := call(t, "GET", "/quotes/"+quoteID+"/versions", nil, 200)
		_ = versions
		listed := call(t, "GET", "/quotes?opportunityId="+oppID, nil, 200)
		_ = listed
		oldID := quoteID
		quoteID = revision["id"].(string)
		call(t, "PATCH", "/quotes/"+quoteID+"/submit", m(), 200)
		if current := call(t, "GET", "/quotes/"+quoteID, nil, 200); current["status"] == "pending-approval" {
			call(t, "PATCH", "/quotes/"+quoteID+"/approve", m(), 200)
		}
		call(t, "PATCH", "/quotes/"+quoteID+"/send", m(), 200)
		call(t, "PATCH", "/quotes/"+quoteID+"/accept", m(), 200)
		// Accepting the job's quote wins the job and starts delivery.
		won := call(t, "GET", "/opportunities/"+oppID, nil, 200)
		if won["stage"] != "closed-won" || won["deliveryStatus"] != "planning" {
			t.Fatal(won)
		}
		current := won["currentQuote"].(map[string]any)
		if current["version"] != float64(2) || current["id"] != quoteID {
			t.Fatal(current)
		}
		// The accepted version is the deal; it is not reopened.
		call(t, "POST", "/quotes/"+quoteID+"/revisions", m("reason", "too late"), 422)
		superseded := call(t, "GET", "/quotes/"+oldID, nil, 200)
		if superseded["isCurrent"] != false || superseded["supersededAt"] == nil {
			t.Fatal(superseded)
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
	t.Run("supplier quote carries its customer quote and opportunity into the purchase order", func(t *testing.T) {
		sq := call(t, "POST", "/procurement/supplier-quotes", m(
			"supplierId", mfrID, "supplierName", "Integration Manufacturer", "currency", "SAR",
			"sourceQuoteId", quoteID, "opportunityId", oppID,
			"items", []any{m("productId", productID, "productName", "Camera", "quantity", 4, "unitCost", 30)}), 201)
		if sq["sourceQuoteId"] != quoteID || sq["opportunityId"] != oppID {
			t.Fatalf("supplier quote did not keep its customer links: %v", sq)
		}
		sqPath := "/procurement/supplier-quotes/" + sq["id"].(string)
		call(t, "PATCH", sqPath+"/status", m("status", "accepted"), 200)

		po := call(t, "POST", sqPath+"/convert-to-po", m(), 201)
		if po["sourceQuoteId"] != quoteID || po["opportunityId"] != oppID {
			t.Fatalf("conversion dropped the customer quote or opportunity: %v", po)
		}
		reloaded := call(t, "GET", "/procurement/purchase-orders/"+po["id"].(string), nil, 200)
		if reloaded["sourceQuoteNumber"] == "" || reloaded["opportunityTitle"] != "API opportunity" {
			t.Fatalf("purchase order does not report its links by name: %v", reloaded)
		}

		// The supplier's own paperwork must be attachable to both ends of the chain.
		doc := models.Document{Name: "Supplier quotation", Category: "general", DocumentType: "terms", Version: "1", FileName: "vendor-quote.pdf", FilePath: "vendor-quote.pdf", FileType: "application/pdf", FileSize: 1024}
		must(t, db.Create(&doc).Error)
		docPath := "/documents/" + doc.ID.String() + "/links"
		call(t, "POST", docPath, m("entityType", "supplier-quote", "entityId", sq["id"]), 201)
		call(t, "POST", docPath, m("entityType", "purchase-order", "entityId", po["id"]), 201)
		listed := call(t, "GET", "/documents?entityId="+po["id"].(string), nil, 200)
		if listed["data"] == nil {
			t.Fatalf("documents could not be listed by linked record: %v", listed)
		}
		call(t, "PATCH", "/procurement/purchase-orders/"+po["id"].(string)+"/status", m("status", "cancelled"), 200)
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
		call(t, "POST", "/price-books", m("name", "No customer", "type", "customer-specific"), 400)
		book := call(t, "POST", "/price-books", m("name", "Customer pricing", "type", "customer-specific", "customerId", customerID), 201)
		call(t, "POST", "/price-books/"+book["id"].(string)+"/entries", m("productId", productID, "customPrice", 90), 200)
		// The won opportunity is the project: delivery is recorded on it directly.
		delivery := call(t, "PATCH", "/opportunities/"+oppID+"/delivery", m("deliveryStatus", "in-progress", "projectManagerId", user["id"], "startDate", "2026-10-01", "targetEndDate", "2026-12-15"), 200)
		if delivery["deliveryStatus"] != "in-progress" || delivery["projectManagerName"] != "Test Admin" {
			t.Fatal(delivery)
		}
		call(t, "PATCH", "/opportunities/"+oppID+"/delivery", m("startDate", "2026-12-01", "targetEndDate", "2026-10-01"), 400)
		open := call(t, "GET", "/delivery?open=true", nil, 200)
		_ = open
		lost := call(t, "POST", "/opportunities", m("title", "Not won", "customerId", customerID), 201)
		call(t, "PATCH", "/opportunities/"+lost["id"].(string)+"/delivery", m("deliveryStatus", "planning"), 400)
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
		for _, path := range []string{"/customers", "/customers/lookup", "/opportunities", "/quotes", "/contracts", "/delivery", "/manufacturers", "/products", "/services", "/price-books", "/recurring-services", "/exchange-rates", "/teams", "/users/lookup", "/documents", "/inventory/stock", "/inventory/reservations", "/inventory/movements", "/procurement/purchase-orders", "/procurement/supplier-quotes", "/procurement/goods-receipts", "/procurement/supplier-items"} {
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
	t.Run("quote numbers follow their opportunity or contract", func(t *testing.T) {
		first := call(t, "POST", "/opportunities", m("title", "Numbered job", "customerId", customerID), 201)
		second := call(t, "POST", "/opportunities", m("title", "Next numbered job", "customerId", customerID), 201)
		number := first["opportunityNumber"].(string)
		if !strings.HasPrefix(number, "OPP-") || len(number) != len("OPP-2026-0001") || second["opportunityNumber"] == number {
			t.Fatal("opportunities are not numbered uniquely", first["opportunityNumber"], second["opportunityNumber"])
		}
		want := "QT-" + strings.TrimPrefix(number, "OPP-")
		quote := call(t, "POST", "/quotes", m("customerId", customerID, "opportunityId", first["id"]), 201)
		if quote["quoteNumber"] != want {
			t.Fatalf("quote %v should share opportunity %s", quote["quoteNumber"], number)
		}
		// Deleting the draft frees the job to be quoted again under the same number.
		call(t, "DELETE", "/quotes/"+quote["id"].(string), nil, 200)
		again := call(t, "POST", "/quotes", m("customerId", customerID, "opportunityId", first["id"]), 201)
		if again["quoteNumber"] != want {
			t.Fatal("re-quoting a job must reuse its number", again["quoteNumber"])
		}
		found := call(t, "GET", "/opportunities?q="+number, nil, 200)
		if items, ok := found["data"].([]any); !ok || len(items) != 1 {
			t.Fatal("opportunities are not searchable by number", found)
		}
		contract := call(t, "POST", "/contracts", m("title", "Maintenance period", "customerId", customerID, "type", "maintenance", "value", 1000, "startDate", "2026-10-01", "endDate", "2027-09-30"), 201)
		call(t, "PATCH", "/contracts/"+contract["id"].(string)+"/activate", m(), 200)
		base := "QT-" + contract["contractNumber"].(string)
		one := call(t, "POST", "/quotes", m("customerId", customerID, "contractId", contract["id"]), 201)
		two := call(t, "POST", "/quotes", m("customerId", customerID, "contractId", contract["id"]), 201)
		if one["quoteNumber"] != base+"-01" || two["quoteNumber"] != base+"-02" {
			t.Fatal("contract quotes should count up under the contract", one["quoteNumber"], two["quoteNumber"])
		}
	})
	t.Run("a draft quote's price book can be set, kept and cleared in the builder", func(t *testing.T) {
		opp := call(t, "POST", "/opportunities", m("title", "Price book job", "customerId", customerID), 201)
		quote := call(t, "POST", "/quotes", m("customerId", customerID, "opportunityId", opp["id"]), 201)
		path := "/quotes/" + quote["id"].(string) + "/builder"
		general := call(t, "POST", "/price-books", m("name", "General list", "type", "standard"), 201)
		other := call(t, "POST", "/customers", m("companyName", "Other customer", "sector", "government", "region", "Riyadh", "status", "prospect", "type", "get"), 201)
		foreign := call(t, "POST", "/price-books", m("name", "Other customer's book", "type", "customer-specific", "customerId", other["id"]), 201)
		save := func(extra map[string]any, want int) map[string]any {
			current := call(t, "GET", path, nil, 200)
			body := m("customerId", customerID, "opportunityId", opp["id"], "lockVersion", current["lockVersion"], "currency", "SAR", "vatPercent", 15, "rows", []any{})
			for k, v := range extra {
				body[k] = v
			}
			return call(t, "PUT", path, body, want)
		}
		if saved := save(m("priceBookId", general["id"]), 200); saved["priceBookId"] != general["id"] {
			t.Fatal("price book not set", saved)
		}
		if saved := save(m(), 200); saved["priceBookId"] != general["id"] {
			t.Fatal("omitting the field must keep the price book", saved)
		}
		save(m("priceBookId", foreign["id"]), 400)
		if saved := save(m("priceBookId", nil), 200); saved["priceBookId"] != nil {
			t.Fatal("null must switch to standard pricing", saved)
		}
	})
	t.Run("job purchases are held on receipt and fulfilment counts only held stock", func(t *testing.T) {
		cameraLine := func() map[string]any {
			f := call(t, "GET", "/opportunities/"+oppID+"/fulfilment", nil, 200)
			for _, raw := range f["lines"].([]any) {
				if line := raw.(map[string]any); line["description"] == "Camera" {
					return line
				}
			}
			t.Fatal(f)
			return nil
		}
		before := cameraLine()
		lineID := before["quoteLineItemId"].(string)

		// Converting a supplier quote raised for the job links its lines to the job.
		sq := call(t, "POST", "/procurement/supplier-quotes", m("supplierName", "Job vendor", "currency", "SAR", "sourceQuoteId", quoteID,
			"items", []any{m("productId", productID, "productName", "Camera", "quantity", 5, "unitCost", 20)}), 201)
		if sq["opportunityId"] != oppID {
			t.Fatalf("supplier quote did not take the quote's opportunity: %v", sq)
		}
		sqPath := "/procurement/supplier-quotes/" + sq["id"].(string)
		call(t, "PATCH", sqPath+"/status", m("status", "accepted"), 200)
		po := call(t, "POST", sqPath+"/convert-to-po", m(), 201)
		item := po["items"].([]any)[0].(map[string]any)
		if item["quoteLineItemId"] != lineID || po["opportunityId"] != oppID {
			t.Fatalf("converted order is not tied to the job's quote line: %v", po)
		}
		poPath := "/procurement/purchase-orders/" + po["id"].(string)
		call(t, "PATCH", poPath+"/status", m("status", "pending-approval"), 200)
		call(t, "PATCH", poPath+"/approve", m(), 200)
		call(t, "PATCH", poPath+"/status", m("status", "ordered"), 200)
		ordered := cameraLine()
		if ordered["outstanding"].(float64)-before["outstanding"].(float64) != 5 {
			t.Fatalf("order not counted as outstanding: %v", ordered)
		}

		receive := func(qty int, condition string) {
			call(t, "POST", "/procurement/goods-receipts", m("poId", po["id"], "receiveDate", "2026-09-29", "items", []any{m("poItemId", item["id"], "productId", productID, "receivedQty", qty, "storageLocation", "dammam-branch", "condition", condition)}), 201)
		}
		receive(3, "good")
		var held int64
		must(t, db.Model(&models.StockReservation{}).Where("quote_line_item_id = ? AND status = 'active' AND warehouse_location = 'dammam-branch'", lineID).Select("COALESCE(SUM(qty), 0)").Scan(&held).Error)
		var stock models.WarehouseStock
		must(t, db.Where("product_id = ? AND warehouse_location = 'dammam-branch'", productID).First(&stock).Error)
		need := int64(before["quantity"].(float64) + 0.999)
		want := min(3, need-int64(before["reserved"].(float64)))
		if held != want || stock.AvailableQty != stock.OnHandQty-int(want) {
			t.Fatalf("good receipt was not held for the job: held %d want %d, stock %+v", held, want, stock)
		}

		receive(2, "damaged")
		after := cameraLine()
		if after["damaged"].(float64)-before["damaged"].(float64) != 2 || after["received"].(float64)-before["received"].(float64) != 3 {
			t.Fatalf("damaged units counted as received: %v", after)
		}
		if after["reserved"].(float64) < float64(need) && after["state"] == "covered" {
			t.Fatalf("line shows covered without enough held stock: %v", after)
		}
	})
}
