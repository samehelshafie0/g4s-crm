package integration

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"g4s-crm/api/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// All business records are created through the public API. The DB is read only
// for final reconciliation so this cannot pass by bypassing workflow transitions.
func testFullCycle(t *testing.T, db *gorm.DB, r *gin.Engine, token string) {
	t.Run("customer to catalog quote mixed stock procurement damage replacement and delivery", func(t *testing.T) {
		m := func(pairs ...any) map[string]any {
			out := map[string]any{}
			for i := 0; i < len(pairs); i += 2 {
				out[pairs[i].(string)] = pairs[i+1]
			}
			return out
		}
		callAs := func(actor, method, path string, body any, want int) map[string]any {
			t.Helper()
			raw, _ := json.Marshal(body)
			req := httptest.NewRequest(method, "/api/v1"+path, bytes.NewReader(raw))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+actor)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != want {
				t.Fatalf("%s %s got %d want %d: %s", method, path, w.Code, want, w.Body.String())
			}
			out := map[string]any{}
			must(t, json.Unmarshal(w.Body.Bytes(), &out))
			if data, ok := out["data"].(map[string]any); ok {
				return data
			}
			return out
		}
		actors := map[string]string{}
		for _, role := range []string{"sales_executive", "sales_manager", "procurement_officer", "procurement_manager", "warehouse_manager", "project_manager", "viewer"} {
			email := "cycle-" + role + "@example.test"
			callAs(token, "POST", "/users", m("firstName", "Cycle", "lastName", role, "email", email, "password", "CycleAcceptance123!", "role", role, "department", "operations"), 201)
			session := callAs("", "POST", "/auth/login", m("email", email, "password", "CycleAcceptance123!"), 200)
			actors[role] = session["accessToken"].(string)
		}
		// Every business write uses its operational role; admin only provisions accounts.
		call := func(method, path string, body any, want int) map[string]any {
			role := "sales_executive"
			switch {
			case strings.HasPrefix(path, "/products"), strings.HasPrefix(path, "/manufacturers"):
				role = "procurement_manager"
			case strings.HasPrefix(path, "/inventory"), strings.HasPrefix(path, "/procurement/goods-receipts"):
				role = "warehouse_manager"
			case strings.HasPrefix(path, "/procurement"):
				role = "procurement_officer"
				if strings.HasSuffix(path, "/approve") {
					role = "procurement_manager"
				}
			case strings.HasSuffix(path, "/approve"), strings.HasSuffix(path, "/convert-to-contract"), strings.HasPrefix(path, "/contracts"):
				role = "sales_manager"
			case strings.HasSuffix(path, "/delivery"):
				role = "project_manager"
			}
			return callAs(actors[role], method, path, body, want)
		}

		customer := call("POST", "/customers", m("companyName", "Full Cycle Customer", "sector", "government", "region", "Riyadh", "status", "prospect", "type", "get"), 201)
		customerID := customer["id"].(string)
		site := call("POST", "/customers/"+customerID+"/sites", m("name", "Full Cycle Site", "city", "Riyadh"), 201)
		call("POST", "/customers/"+customerID+"/contacts", m("name", "Cycle Buyer", "isPrimary", true, "siteId", site["id"]), 201)
		maker := call("POST", "/manufacturers", m("name", "Cycle Maker", "code", "CYCLE", "vendorType", "both"), 201)
		category := call("POST", "/manufacturers/"+maker["id"].(string)+"/categories", m("name", "Security"), 201)
		manual := call("POST", "/products", m("sku", "CYCLE-MANUAL", "name", "Manual recorder", "manufacturerId", maker["id"], "categoryId", category["id"], "productType", "local", "originCurrency", "USD", "fxRate", 3.75, "unitCostOrigin", 50, "sellingPrice", 100, "supplierName", "Cycle Supplier"), 201)
		if manual["originCurrency"] != "SAR" || manual["fxRate"] != float64(1) || manual["landedCostSAR"] != float64(50) {
			t.Fatal("local product currency/cost mismatch", manual)
		}
		importBody := m("source", "vendor-catalog", "sourceRef", "g4s-product-import.xlsx", "productType", "import", "originCurrency", "USD", "fxRate", 3.75, "targetMarginPercent", 25, "updateExisting", true,
			"rows", []any{m("sku", "CYCLE-IMPORT", "name", "Imported camera", "unitCost", 100, "vendorName", "Cycle Supplier", "manufacturerName", "cycle", "categoryName", "Security", "currency", "SAR"), m("sku", "CYCLE-EUR", "name", "Euro sensor", "unitCost", 20, "vendorName", "Other Supplier", "manufacturerName", "Cycle Maker", "categoryName", "Security", "currency", "EUR", "fxRate", 4.2)})
		imported := call("POST", "/products/import", importBody, 201)
		if imported["created"] != float64(2) {
			t.Fatal(imported)
		}
		productID := imported["rows"].([]any)[0].(map[string]any)["productId"].(string)
		euroID := imported["rows"].([]any)[1].(map[string]any)["productId"].(string)
		euro := call("GET", "/products/"+euroID, nil, 200)
		if euro["landedCostSAR"] != float64(84) || euro["manufacturerName"] != "Cycle Maker" || euro["categoryName"] != "Security" {
			t.Fatal(euro)
		}
		importBody["updateExisting"] = false
		repeated := call("POST", "/products/import", importBody, 201)
		if repeated["skipped"] != float64(2) {
			t.Fatal(repeated)
		}
		importBody["updateExisting"] = true
		importBody["rows"] = []any{m("sku", "CYCLE-IMPORT", "name", "Imported camera", "unitCost", 100, "vendorName", "Cycle Supplier", "manufacturerName", "cycle", "currency", "SAR")}
		call("POST", "/products/import", importBody, 201)
		if current := call("GET", "/products/"+productID, nil, 200); current["categoryName"] != "Security" {
			t.Fatal("repricing cleared the existing category", current)
		}
		// One invalid classification rolls back the entire batch; no partial hidden writes.
		importBody["rows"] = []any{m("sku", "CYCLE-ROLLBACK", "name", "Rollback", "unitCost", 1, "vendorName", "Vendor"), m("sku", "CYCLE-INVALID", "name", "Unknown maker", "unitCost", 1, "vendorName", "Vendor", "manufacturerName", "Unknown manufacturer")}
		call("POST", "/products/import", importBody, 400)
		var count int64
		must(t, db.Model(&models.Product{}).Where("sku = ?", "CYCLE-ROLLBACK").Count(&count).Error)
		if count != 0 {
			t.Fatal("batch was partially persisted")
		}
		importBody["rows"] = []any{m("sku", "CYCLE-MISSING-FX", "name", "Missing FX", "unitCost", 1, "vendorName", "Vendor", "currency", "EUR")}
		call("POST", "/products/import", importBody, 400)
		importBody["updateExisting"] = true
		importBody["rows"] = []any{m("sku", "CYCLE-EUR", "name", "Euro sensor", "unitCost", 20, "vendorName", "Other Supplier", "currency", "EUR", "fxRate", 4.3)}
		call("POST", "/products/import", importBody, 201)
		must(t, db.Model(&models.ProductPriceRecord{}).Where("product_id = ?", euroID).Count(&count).Error)
		if count != 2 {
			t.Fatal("FX-only price change was not audited", count)
		}

		// Purchase existing warehouse stock first, rather than seed it by editing the DB.
		order := func(po map[string]any) {
			path := "/procurement/purchase-orders/" + po["id"].(string)
			call("PATCH", path+"/status", m("status", "pending-approval"), 200)
			call("PATCH", path+"/approve", m(), 200)
			call("PATCH", path+"/status", m("status", "ordered"), 200)
		}
		receive := func(po map[string]any, qty, damage int) map[string]any {
			item := po["items"].([]any)[0].(map[string]any)
			condition := "good"
			if damage > 0 {
				condition = "partial-damage"
			}
			return call("POST", "/procurement/goods-receipts", m("poId", po["id"], "receiveDate", "2026-10-07", "items", []any{m("poItemId", item["id"], "productId", item["productId"], "receivedQty", qty, "damagedQty", damage, "condition", condition, "storageLocation", "riyadh-main")}), 201)
		}
		stockPO := call("POST", "/procurement/purchase-orders", m("supplierName", "Cycle Supplier", "currency", "SAR", "items", []any{m("productId", productID, "quantity", 3, "unitCost", 100)}), 201)
		order(stockPO)
		receive(stockPO, 3, 0)
		opportunity := call("POST", "/opportunities", m("title", "Full Cycle Installation", "customerId", customerID), 201)
		oppID := opportunity["id"].(string)
		quote := call("POST", "/quotes", m("customerId", customerID, "opportunityId", oppID, "vatPercent", 15), 201)
		qID := quote["id"].(string)
		row := func(id any, description string, qty, cost, price int) map[string]any {
			return m("rowType", "item", "source", "product", "productId", id, "description", description, "quantity", qty, "multiplier", 1, "unitCost", cost, "unitPrice", price, "isSelected", true, "isPrintable", true)
		}
		callAs(actors["warehouse_manager"], "GET", "/procurement/job-demands/"+qID, nil, 404)

		call("PUT", "/quotes/"+qID+"/builder", m("customerId", customerID, "opportunityId", oppID, "lockVersion", quote["lockVersion"], "currency", "SAR", "vatPercent", 15, "rows", []any{row(productID, "Imported camera", 5, 100, 150), row(manual["id"], "Manual recorder", 2, 50, 100)}), 200)
		call("PATCH", "/quotes/"+qID+"/submit", m(), 200)
		if call("GET", "/quotes/"+qID, nil, 200)["status"] == "pending-approval" {
			call("PATCH", "/quotes/"+qID+"/approve", m(), 200)
		}
		call("PATCH", "/quotes/"+qID+"/send", m(), 200)
		call("PATCH", "/quotes/"+qID+"/accept", m(), 200)
		for _, role := range []string{"procurement_officer", "warehouse_manager"} {
			callAs(actors[role], "GET", "/quotes/"+qID, nil, 403)
			for _, endpoint := range []string{"kpis", "pipeline", "recent-quotes", "expiring-quotes", "sales-performance", "top-customers", "recent-activity", "alerts"} {
				callAs(actors[role], "GET", "/dashboard/"+endpoint, nil, 403)
			}
			demand := callAs(actors[role], "GET", "/procurement/job-demands/"+qID, nil, 200)
			if len(demand["lineItems"].([]any)) != 2 {
				t.Fatal("material demand missing", demand)
			}
			for _, key := range []string{"total", "totalCost", "internalNotes", "notes", "marginPercent", "customer"} {
				if _, exists := demand[key]; exists {
					t.Fatal("operational lookup leaks", key)
				}
			}
			for _, raw := range demand["lineItems"].([]any) {
				for _, key := range []string{"unitPrice", "unitCost", "lineTotal", "marginPercent"} {
					if _, exists := raw.(map[string]any)[key]; exists {
						t.Fatal("operational line leaks", key)
					}
				}
			}
		}
		callAs(actors["sales_executive"], "GET", "/procurement/job-demands", nil, 403)
		callAs(actors["viewer"], "POST", "/procurement/goods-receipts", m(), 403)
		callAs(actors["warehouse_manager"], "POST", "/procurement/purchase-orders", m(), 403)
		callAs(actors["procurement_officer"], "PATCH", "/procurement/purchase-orders/"+stockPO["id"].(string)+"/approve", m(), 403)

		won := call("GET", "/opportunities/"+oppID, nil, 200)
		if won["stage"] != "closed-won" {
			t.Fatal(won)
		}
		quote = call("GET", "/quotes/"+qID, nil, 200)
		if quote["total"] != 1092.5 {
			t.Fatal("quote totals", quote)
		}
		lines := map[string]string{}
		for _, raw := range quote["lineItems"].([]any) {
			line := raw.(map[string]any)
			lines[line["productId"].(string)] = line["id"].(string)
		}
		holdBody := m("productId", productID, "warehouseLocation", "riyadh-main", "qty", 3, "source", "quote", "quoteLineItemId", lines[productID])
		hold := call("POST", "/inventory/reservations", holdBody, 201)
		call("PATCH", "/inventory/reservations/"+hold["id"].(string)+"/release", m("reason", "Test release before reholding"), 200)
		call("POST", "/inventory/reservations", holdBody, 201)
		// Buy only the shortfall through a supplier quotation; retain exact demand links.
		sq := call("POST", "/procurement/supplier-quotes", m("supplierName", "Cycle Supplier", "currency", "SAR", "sourceQuoteId", qID, "opportunityId", oppID, "items", []any{m("productId", productID, "productName", "Imported camera", "quantity", 2, "unitCost", 100)}), 201)
		sqPath := "/procurement/supplier-quotes/" + sq["id"].(string)
		call("PATCH", sqPath+"/status", m("status", "accepted"), 200)
		cameraPO := call("POST", sqPath+"/convert-to-po", m(), 201)
		manualPO := call("POST", "/procurement/purchase-orders", m("supplierName", "Cycle Supplier", "currency", "SAR", "sourceQuoteId", qID, "items", []any{m("productId", manual["id"], "quantity", 2, "unitCost", 50, "quoteLineItemId", lines[manual["id"].(string)])}), 201)
		call("POST", "/inventory/reservations", holdBody, 400)
		order(cameraPO)
		order(manualPO)
		damaged := receive(cameraPO, 2, 1)
		receive(manualPO, 1, 0)
		partial := call("GET", "/procurement/purchase-orders/"+manualPO["id"].(string), nil, 200)
		if partial["status"] != "partial-received" {
			t.Fatal(partial)
		}
		receive(manualPO, 1, 0)
		damagedItem := damaged["items"].([]any)[0].(map[string]any)
		call("PATCH", "/procurement/goods-receipts/"+damaged["id"].(string)+"/items/"+damagedItem["id"].(string)+"/return", m("reference", "CYCLE-RETURN-001"), 200)
		replacement := call("GET", "/procurement/purchase-orders/"+damaged["replacementPoId"].(string), nil, 200)
		order(replacement)
		receive(replacement, 1, 0)
		fulfilment := call("GET", "/opportunities/"+oppID+"/fulfilment", nil, 200)
		for _, raw := range fulfilment["lines"].([]any) {
			line := raw.(map[string]any)
			if line["state"] != "covered" || line["outstanding"] != float64(0) {
				t.Fatal(line)
			}
		}
		// Fulfil every active hold through the API and reconcile the physical stock.
		for _, product := range []string{productID, manual["id"].(string)} {
			var reservations []models.StockReservation
			must(t, db.Where("product_id = ? AND status = 'active'", product).Find(&reservations).Error)
			for _, reservation := range reservations {
				call("PATCH", "/inventory/reservations/"+reservation.ID.String()+"/fulfill", m("reason", "Full cycle customer delivery"), 200)
			}
			var stock models.WarehouseStock
			must(t, db.Where("product_id = ? AND warehouse_location = ?", product, "riyadh-main").First(&stock).Error)
			if stock.OnHandQty != 0 || stock.ReservedQty != 0 || stock.AvailableQty != 0 {
				t.Fatal("stock did not reconcile", stock)
			}
		}
		finalDemand := callAs(actors["warehouse_manager"], "GET", "/procurement/job-demands/"+qID, nil, 200)
		for _, raw := range finalDemand["lineItems"].([]any) {
			if raw.(map[string]any)["remainingQty"] != float64(0) {
				t.Fatal("dispatched demand reappeared", finalDemand)
			}
		}
		dispatched := call("GET", "/opportunities/"+oppID+"/fulfilment", nil, 200)
		for _, raw := range dispatched["lines"].([]any) {
			line := raw.(map[string]any)
			if line["source"] != "service" && line["dispatched"].(float64) < line["quantity"].(float64) {
				t.Fatal("fulfilment does not report the dispatched quantity", line)
			}
		}

		contract := call("POST", "/quotes/"+qID+"/convert-to-contract", m("title", "Cycle sales contract", "type", "sales", "startDate", "2026-10-07", "endDate", "2027-10-07"), 201)
		call("PATCH", "/contracts/"+contract["id"].(string)+"/activate", m(), 200)
		call("PATCH", "/opportunities/"+oppID+"/delivery", m("deliveryStatus", "in-progress", "startDate", "2026-10-07", "targetEndDate", "2026-10-08"), 200)
		completed := call("PATCH", "/opportunities/"+oppID+"/delivery", m("deliveryStatus", "completed", "actualEndDate", "2026-10-08"), 200)
		if completed["deliveryStatus"] != "completed" {
			t.Fatal(completed)
		}
	})
}
