package integration

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"sync"
	"testing"

	"g4s-crm/api/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func testFulfilmentCloseout(t *testing.T, db *gorm.DB, r *gin.Engine, token, customerID string) {
	request := func(method, path string, body any) (int, map[string]any) {
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest(method, "/api/v1"+path, bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		result := map[string]any{}
		_ = json.Unmarshal(w.Body.Bytes(), &result)
		if data, ok := result["data"].(map[string]any); ok {
			return w.Code, data
		}
		return w.Code, result
	}
	call := func(method, path string, body any, want int) map[string]any {
		t.Helper()
		code, result := request(method, path, body)
		if code != want {
			t.Fatalf("%s %s: got %d want %d: %v", method, path, code, want, result)
		}
		return result
	}
	m := func(pairs ...any) map[string]any {
		out := map[string]any{}
		for i := 0; i < len(pairs); i += 2 {
			out[pairs[i].(string)] = pairs[i+1]
		}
		return out
	}
	mfr := models.Manufacturer{Name: "Closeout maker", Code: "CLOSEOUT", VendorType: models.VendorTypeBoth, IsActive: true}
	must(t, db.Create(&mfr).Error)
	category := models.ManufacturerCategory{ManufacturerID: mfr.ID, Name: "Sensors"}
	must(t, db.Create(&category).Error)
	product := models.Product{SKU: "CLOSEOUT-SENSOR", Name: "Closeout sensor", ManufacturerID: &mfr.ID, CategoryID: &category.ID, ProductType: models.ProductTypeLocal, OriginCurrency: models.CurrencySAR, FXRate: 1}
	must(t, db.Create(&product).Error)
	opp := models.Opportunity{Title: "Closeout demand", CustomerID: uuid.MustParse(customerID), Stage: models.StageClosedWon}
	must(t, db.Create(&opp).Error)
	quote := models.Quote{QuoteNumber: "QT-CLOSEOUT", OpportunityID: &opp.ID, CustomerID: opp.CustomerID, Status: models.QuoteStatusAccepted, Currency: models.CurrencySAR, FXRate: 1, IsCurrent: true}
	must(t, db.Create(&quote).Error)
	line := models.QuoteLineItem{QuoteID: quote.ID, ProductID: &product.ID, Description: "Sensor", Category: models.QuoteLineMaterials, RowType: "item", Source: "product", Quantity: 2.5, Multiplier: 2, IsSelected: true}
	must(t, db.Create(&line).Error)
	call("POST", "/inventory/movements/adjustment", m("productId", product.ID, "warehouseLocation", "riyadh-main", "qty", 20, "unitCost", 10, "reason", "Closeout fixture"), 200)
	hold := func(qty int) map[string]any {
		return m("productId", product.ID, "warehouseLocation", "riyadh-main", "qty", qty, "source", "quote", "quoteLineItemId", line.ID)
	}
	poBody := func(qty int) map[string]any {
		return m("supplierName", "alfanar Electric (Riyadh)", "currency", "SAR", "sourceQuoteId", quote.ID, "items", []any{m("productId", product.ID, "quantity", qty, "unitCost", 10, "quoteLineItemId", line.ID)})
	}
	t.Run("inventory classifications and quote holds", func(t *testing.T) {
		listed := call("GET", "/inventory/stock?productId="+product.ID.String(), nil, 200)
		stock := listed["data"].([]any)[0].(map[string]any)
		if stock["manufacturerName"] != "Closeout maker" || stock["categoryName"] != "Sensors" {
			t.Fatal(stock)
		}
		unlinked := hold(1)
		delete(unlinked, "quoteLineItemId")
		call("POST", "/inventory/reservations", unlinked, 400)
		held := call("POST", "/inventory/reservations", hold(2), 201)
		if held["sourceRef"] != "QT-CLOSEOUT v1" || held["quoteLineItemId"] != line.ID.String() {
			t.Fatal(held)
		}
		call("POST", "/inventory/reservations", hold(4), 400)
		call("PATCH", "/inventory/reservations/"+held["id"].(string)+"/fulfill", m("reason", "Delivery 1"), 200)
	})
	var po map[string]any
	t.Run("combined caps direct conversion duplicate rows and draft edits", func(t *testing.T) {
		call("POST", "/procurement/purchase-orders", poBody(4), 400)
		duplicate := poBody(2)
		duplicate["items"] = append(duplicate["items"].([]any), duplicate["items"].([]any)[0])
		call("POST", "/procurement/purchase-orders", duplicate, 400)
		po = call("POST", "/procurement/purchase-orders", poBody(3), 201)
		if po["supplierName"] != "alfanar Electric" {
			t.Fatal(po)
		}
		call("PATCH", "/procurement/purchase-orders/"+po["id"].(string), poBody(3), 200)
		po = call("GET", "/procurement/purchase-orders/"+po["id"].(string), nil, 200)
		call("POST", "/inventory/reservations", hold(1), 400)
		call("POST", "/procurement/purchase-orders", poBody(1), 400)
		sq := call("POST", "/procurement/supplier-quotes", m("supplierName", "Vendor", "currency", "SAR", "sourceQuoteId", quote.ID, "items", []any{m("productId", product.ID, "productName", "Sensor", "quantity", 1, "unitCost", 10)}), 201)
		path := "/procurement/supplier-quotes/" + sq["id"].(string)
		call("PATCH", path+"/status", m("status", "accepted"), 200)
		call("POST", path+"/convert-to-po", m(), 400)
	})
	t.Run("partial damage creates return and replacement without double counting", func(t *testing.T) {
		path := "/procurement/purchase-orders/" + po["id"].(string)
		call("PATCH", path+"/status", m("status", "pending-approval"), 200)
		call("PATCH", path+"/approve", m(), 200)
		call("PATCH", path+"/status", m("status", "ordered"), 200)
		item := po["items"].([]any)[0].(map[string]any)
		receiptLine := m("poItemId", item["id"], "productId", product.ID, "receivedQty", 3, "damagedQty", 1, "condition", "partial-damage", "storageLocation", "dammam-branch")
		body := m("poId", po["id"], "receiveDate", "2026-10-07", "items", []any{receiptLine})
		receiptLine["damagedQty"] = 0
		call("POST", "/procurement/goods-receipts", body, 400)
		receiptLine["damagedQty"] = 1
		gr := call("POST", "/procurement/goods-receipts", body, 201)
		replacementID, ok := gr["replacementPoId"].(string)
		if !ok {
			t.Fatal(gr)
		}
		replacement := call("GET", "/procurement/purchase-orders/"+replacementID, nil, 200)
		if replacement["status"] != "draft" || replacement["items"].([]any)[0].(map[string]any)["quantity"] != float64(1) {
			t.Fatal(replacement)
		}
		gri := gr["items"].([]any)[0].(map[string]any)
		if gri["returnStatus"] != "pending" || gri["damagedQty"] != float64(1) {
			t.Fatal(gri)
		}
		listed := call("GET", "/inventory/stock?productId="+product.ID.String()+"&warehouseLocation=dammam-branch", nil, 200)
		stock := listed["data"].([]any)[0].(map[string]any)
		if stock["onHandQty"] != float64(2) || stock["reservedQty"] != float64(2) || stock["totalValue"] != float64(20) {
			t.Fatal(stock)
		}
		call("POST", "/procurement/goods-receipts", body, 400)
		retPath := "/procurement/goods-receipts/" + gr["id"].(string) + "/items/" + gri["id"].(string) + "/return"
		call("PATCH", retPath, m("reference", "RTN-1"), 200)
		call("PATCH", retPath, m("reference", "RTN-1"), 400)
		var persisted models.WarehouseStock
		must(t, db.Where("id = ?", stock["id"]).First(&persisted).Error)
		if persisted.OnHandQty != 2 || persisted.ReservedQty != 2 {
			t.Fatal(persisted)
		}
		call("POST", "/procurement/purchase-orders", poBody(1), 400)
		rp := "/procurement/purchase-orders/" + replacementID
		call("PATCH", rp+"/status", m("status", "pending-approval"), 200)
		call("PATCH", rp+"/approve", m(), 200)
		call("PATCH", rp+"/status", m("status", "ordered"), 200)
		ri := replacement["items"].([]any)[0].(map[string]any)
		call("POST", "/procurement/goods-receipts", m("poId", replacementID, "receiveDate", "2026-10-07", "items", []any{m("poItemId", ri["id"], "productId", product.ID, "receivedQty", 1, "condition", "good", "storageLocation", "dammam-branch")}), 201)
		f := call("GET", "/opportunities/"+opp.ID.String()+"/fulfilment", nil, 200)["lines"].([]any)[0].(map[string]any)
		if f["state"] != "covered" || f["reserved"] != float64(5) || f["damaged"] != float64(1) || f["outstanding"] != float64(0) {
			t.Fatal(f)
		}
	})
	t.Run("simultaneous holds and orders cannot overcommit", func(t *testing.T) {
		// New independent line shares the same product, with demand for one unit.
		line.Base = models.Base{}
		line.Quantity = 1
		line.Multiplier = 1
		must(t, db.Create(&line).Error)
		var wg sync.WaitGroup
		codes := make(chan int, 2)
		for _, kind := range []string{"hold", "order"} {
			wg.Add(1)
			go func(kind string) {
				defer wg.Done()
				path := "/inventory/reservations"
				body := hold(1)
				if kind == "order" {
					path = "/procurement/purchase-orders"
					body = poBody(1)
				}
				code, _ := request("POST", path, body)
				codes <- code
			}(kind)
		}
		wg.Wait()
		close(codes)
		success, rejected := 0, 0
		for code := range codes {
			if code == 201 {
				success++
			} else if code == 400 {
				rejected++
			} else {
				t.Fatalf("unexpected concurrent status %d", code)
			}
		}
		if success != 1 || rejected != 1 {
			t.Fatalf("success=%d rejected=%d", success, rejected)
		}
	})
	t.Run("legacy unlinked job order cannot bypass caps at submission", func(t *testing.T) {
		legacy := models.PurchaseOrder{PONumber: "PO-LEGACY-CLOSEOUT", SupplierName: "Vendor", SourceQuoteID: &quote.ID, Status: models.POStatusDraft, Currency: models.CurrencySAR, Items: []models.PurchaseOrderItem{{ProductID: product.ID, Quantity: 100, UnitCost: 10}}}
		must(t, db.Create(&legacy).Error)
		call("PATCH", "/procurement/purchase-orders/"+legacy.ID.String()+"/status", m("status", "pending-approval"), 400)
		call("PATCH", "/procurement/purchase-orders/"+legacy.ID.String()+"/status", m("status", "cancelled"), 200)
	})
	t.Run("incomplete vendor cannot return through import or procurement", func(t *testing.T) {
		body := poBody(1)
		body["supplierName"] = " in "
		call("POST", "/procurement/purchase-orders", body, 400)
		call("POST", "/products/import", m("vendorName", "in", "source", "manual", "productType", "local", "originCurrency", "SAR", "fxRate", 1, "rows", []any{m("sku", "BAD-VENDOR", "name", "Bad vendor", "unitCost", 1)}), 400)
	})
}
