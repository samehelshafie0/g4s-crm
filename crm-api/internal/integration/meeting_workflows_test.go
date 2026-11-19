package integration

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"mime/multipart"
	"net/http/httptest"
	"testing"
)

func testMeetingWorkflows(t *testing.T, db *gorm.DB, r *gin.Engine, admin, customerID string) {
	t.Run("meeting quotation pricing permissions supplier links and multiple teams", func(t *testing.T) {
		call := func(token, method, path string, body any, want int) map[string]any {
			t.Helper()
			raw, err := json.Marshal(body)
			must(t, err)
			req := httptest.NewRequest(method, "/api/v1"+path, bytes.NewReader(raw))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != want {
				t.Fatalf("%s %s: %d want %d: %s", method, path, w.Code, want, w.Body.String())
			}
			out := map[string]any{}
			if w.Body.Len() > 0 {
				must(t, json.Unmarshal(w.Body.Bytes(), &out))
			}
			if data, ok := out["data"].(map[string]any); ok {
				return data
			}
			return out
		}
		m := func(pairs ...any) map[string]any {
			out := map[string]any{}
			for i := 0; i < len(pairs); i += 2 {
				out[pairs[i].(string)] = pairs[i+1]
			}
			return out
		}
		newUser := func(email string) map[string]any {
			return call(admin, "POST", "/users", m("firstName", "Meeting", "lastName", "User", "email", email, "password", "MeetingTest123!", "role", "sales_manager", "department", "sales"), 201)
		}
		leader := newUser("meeting-leader@example.test")
		member := newUser("meeting-member@example.test")
		login := func(email string) string {
			return call("", "POST", "/auth/login", m("email", email, "password", "MeetingTest123!"), 200)["accessToken"].(string)
		}
		call(admin, "PATCH", "/users/"+leader["id"].(string), m("role", "pre_sales", "quotePermissions", m("quotes:approve", true)), 200)
		leaderToken, memberToken := login(leader["email"].(string)), login(member["email"].(string))
		team := call(admin, "POST", "/teams", m("name", "Meeting team", "department", "sales", "leaderId", leader["id"]), 201)
		second := call(admin, "POST", "/teams", m("name", "Second meeting team", "department", "sales", "leaderId", leader["id"]), 201)
		teamID, secondID := team["id"].(string), second["id"].(string)
		for _, id := range []string{teamID, secondID} {
			call(admin, "POST", "/teams/"+id+"/members", m("userId", member["id"]), 200)
		}
		call(admin, "DELETE", "/teams/"+secondID+"/members/"+member["id"].(string), nil, 204)
		options := call(leaderToken, "GET", "/quotes/team-options", nil, 200)["data"].([]any)
		if len(options) != 2 {
			t.Fatal("leader could not load own quotation teams", options)
		}
		call(leaderToken, "GET", "/teams", nil, 403)
		kept := call(admin, "GET", "/teams/"+teamID, nil, 200)
		if len(kept["members"].([]any)) != 2 {
			t.Fatal("removing one team removed another membership", kept)
		}
		vendor := call(admin, "POST", "/manufacturers", m("name", "Meeting Vendor", "code", "MEET-V", "vendorType", "both"), 201)
		call(admin, "POST", "/manufacturers", m("name", "  meeting   vendor ", "code", "MEET-DUP", "vendorType", "supplier"), 409)
		product := call(admin, "POST", "/products", m("sku", "MEETING-ITEM", "name", "Meeting camera", "manufacturerId", vendor["id"], "unitCostOrigin", 80, "sellingPrice", 100), 201)
		productID := product["id"].(string)
		call(admin, "POST", "/products/"+productID+"/vendors", m("vendorName", "Meeting Vendor", "unitCost", 80, "currency", "SAR", "moq", 1), 201)
		call(admin, "POST", "/products/"+productID+"/vendors", m("vendorName", "meeting vendor", "unitCost", 80, "currency", "SAR", "moq", 1), 409)
		opportunity := call(admin, "POST", "/opportunities", m("title", "Meeting project", "customerId", customerID), 201)
		quote := call(admin, "POST", "/quotes", m("customerId", customerID, "opportunityId", opportunity["id"], "teamId", teamID), 201)
		quoteID := quote["id"].(string)
		row := m("id", "new", "rowType", "item", "source", "product", "productId", productID, "description", "Meeting camera", "quantity", 2, "multiplier", 1, "unitCost", 80, "unitPrice", 100, "isSelected", true, "isPrintable", true)
		payload := m("lockVersion", quote["lockVersion"], "customerId", customerID, "currency", "SAR", "vatPercent", 15, "rows", []any{row})
		saved := call(admin, "PUT", "/quotes/"+quoteID+"/builder", payload, 200)
		line := saved["lineItems"].([]any)[0].(map[string]any)
		row["id"] = line["id"]
		row["unitCost"] = 1
		row["unitPrice"] = 200
		payload["lockVersion"] = saved["lockVersion"]
		saved = call(admin, "PUT", "/quotes/"+quoteID+"/builder", payload, 200)
		line = saved["lineItems"].([]any)[0].(map[string]any)
		if line["unitCost"] != float64(80) || line["unitPrice"] != float64(200) {
			t.Fatal("selling-price edit changed fixed cost", line)
		}
		// Reject unknown capabilities and self-escalation. Denials take effect on an existing token.
		call(admin, "PATCH", "/users/"+member["id"].(string), m("quotePermissions", m("quotes:god-mode", true)), 400)
		call(memberToken, "PATCH", "/auth/me", m("quotePermissions", m("quotes:approve", true)), 400)
		overrides := m("quotes:view-margins", false, "quotes:pricing", false, "quotes:export", false, "quotes:documents", false)
		call(admin, "PATCH", "/users/"+member["id"].(string), m("quotePermissions", overrides), 200)
		hidden := call(memberToken, "GET", "/quotes/"+quoteID, nil, 200)
		if _, exists := hidden["marginPercent"]; exists {
			t.Fatal("margin leaked", hidden)
		}
		hiddenLine := hidden["lineItems"].([]any)[0].(map[string]any)
		if _, exists := hiddenLine["unitCost"]; exists {
			t.Fatal("cost leaked", hiddenLine)
		}
		payload["lockVersion"] = saved["lockVersion"]
		row["unitPrice"] = 210
		call(memberToken, "PUT", "/quotes/"+quoteID+"/builder", payload, 403)
		call(memberToken, "GET", "/quotes/"+quoteID+"/pdf", nil, 403)
		call(memberToken, "POST", "/quotes/"+quoteID+"/currency", m("currency", "USD"), 403)
		history := call(memberToken, "GET", "/customers/"+customerID+"/price-history?productId="+productID, nil, 200)["data"].([]any)
		if len(history) != 1 {
			t.Fatal("customer history missing", history)
		}
		if _, exists := history[0].(map[string]any)["unitCost"]; exists {
			t.Fatal("history leaked cost")
		}
		supplier := call(admin, "POST", "/procurement/supplier-quotes", m("supplierName", "Meeting Vendor", "currency", "SAR", "items", []any{m("productId", productID, "productName", "Meeting camera", "quantity", 2, "unitCost", 70)}), 201)
		call(memberToken, "POST", "/quotes/"+quoteID+"/supplier-quotes", m("supplierQuoteId", supplier["id"]), 403)
		call(admin, "POST", "/quotes/"+quoteID+"/supplier-quotes", m("supplierQuoteId", supplier["id"]), 200)
		offers := call(admin, "GET", "/products/"+productID+"/quote-offers", nil, 200)["data"].([]any)
		if len(offers) != 1 || offers[0].(map[string]any)["quoteNumber"] != quote["quoteNumber"] {
			t.Fatal("supplier offer provenance missing", offers)
		}
		supplier = call(admin, "GET", "/procurement/supplier-quotes/"+supplier["id"].(string), nil, 200)
		supplierItem := supplier["items"].([]any)[0].(map[string]any)
		row["id"] = "supplier-new"
		row["supplierQuoteItemId"] = supplierItem["id"]
		row["unitCost"] = 999
		row["unitPrice"] = 200
		payload["lockVersion"] = saved["lockVersion"]
		saved = call(admin, "PUT", "/quotes/"+quoteID+"/builder", payload, 200)
		if saved["lineItems"].([]any)[0].(map[string]any)["unitCost"] != float64(70) {
			t.Fatal("supplier cost was not resolved on server", saved)
		}
		// A quote-only document grant can upload and attach, without granting
		// edits to customer links or unrelated document library records.
		clerk := newUser("meeting-documents@example.test")
		call(admin, "PATCH", "/users/"+clerk["id"].(string), m("role", "sales_executive", "quotePermissions", m("quotes:documents", true)), 200)
		call(admin, "POST", "/teams/"+teamID+"/members", m("userId", clerk["id"]), 200)
		clerkToken := login(clerk["email"].(string))
		var uploadBody bytes.Buffer
		form := multipart.NewWriter(&uploadBody)
		must(t, form.WriteField("category", "quote"))
		part, err := form.CreateFormFile("file", "meeting.txt")
		must(t, err)
		_, err = part.Write([]byte("Supplier supporting document"))
		must(t, err)
		must(t, form.Close())
		uploadRequest := httptest.NewRequest("POST", "/api/v1/documents", &uploadBody)
		uploadRequest.Header.Set("Authorization", "Bearer "+clerkToken)
		uploadRequest.Header.Set("Content-Type", form.FormDataContentType())
		uploadResponse := httptest.NewRecorder()
		r.ServeHTTP(uploadResponse, uploadRequest)
		if uploadResponse.Code != 201 {
			t.Fatalf("quote-only upload: %d %s", uploadResponse.Code, uploadResponse.Body.String())
		}
		var uploaded map[string]any
		must(t, json.Unmarshal(uploadResponse.Body.Bytes(), &uploaded))
		documentID := uploaded["data"].(map[string]any)["id"].(string)
		docPath := "/documents/" + documentID
		link := call(clerkToken, "POST", docPath+"/links", m("entityType", "quote", "entityId", quoteID), 201)
		call(clerkToken, "POST", docPath+"/links", m("entityType", "supplier-quote", "entityId", supplier["id"]), 201)
		call(clerkToken, "POST", docPath+"/links", m("entityType", "customer", "entityId", customerID), 403)
		call(clerkToken, "DELETE", docPath+"/links/"+link["id"].(string), nil, 204)
		call(admin, "PATCH", "/users/"+clerk["id"].(string), m("quotePermissions", m("quotes:documents", false)), 200)
		call(clerkToken, "POST", docPath+"/links", m("entityType", "quote", "entityId", quoteID), 403)
		// A book limits new catalog rows, including direct API and write-in bypass attempts.
		book := call(admin, "POST", "/price-books", m("name", "Meeting negotiated prices", "type", "customer-specific", "customerId", customerID), 201)
		payload["lockVersion"] = saved["lockVersion"]
		payload["priceBookId"] = book["id"]
		row["id"] = "outside-book"
		call(admin, "PUT", "/quotes/"+quoteID+"/builder", payload, 400)
		call(admin, "POST", "/price-books/"+book["id"].(string)+"/entries", m("kind", "product", "productId", productID, "customPrice", 150), 200)
		saved = call(admin, "PUT", "/quotes/"+quoteID+"/builder", payload, 200)
		payload["lockVersion"] = saved["lockVersion"]
		row["id"] = saved["lineItems"].([]any)[0].(map[string]any)["id"]
		row["discountPercent"] = 10
		payload["discountPercent"] = 5
		optional := m("id", "optional", "rowType", "item", "source", "product", "productId", productID, "description", "Optional camera", "quantity", 1, "multiplier", 1, "unitCost", 80, "unitPrice", 999, "isOptional", true, "isSelected", false, "isPrintable", true)
		payload["rows"] = []any{row, optional}
		saved = call(admin, "PUT", "/quotes/"+quoteID+"/builder", payload, 200)
		call(memberToken, "PATCH", "/quotes/"+quoteID+"/submit", nil, 200)
		call(memberToken, "PATCH", "/quotes/"+quoteID+"/approve", nil, 403)
		call(leaderToken, "PATCH", "/quotes/"+quoteID+"/approve", nil, 200)
		historyPath := "/products/" + productID + "/sales-history?customerId=" + customerID
		if len(call(admin, "GET", historyPath, nil, 200)["data"].([]any)) != 0 {
			t.Fatal("approved quote counted as sold")
		}
		call(admin, "PATCH", "/quotes/"+quoteID+"/send", nil, 200)
		accepted := call(admin, "PATCH", "/quotes/"+quoteID+"/accept", nil, 200)
		if accepted["acceptedAt"] == nil || accepted["acceptedAtEstimated"] != false {
			t.Fatal("acceptance timestamp missing", accepted)
		}
		sales := call(admin, "GET", historyPath, nil, 200)["data"].([]any)
		if len(sales) != 1 || sales[0].(map[string]any)["netUnitPrice"] != float64(171) {
			t.Fatal("sale price omitted discounts or included optional item", sales)
		}
		sold := call(admin, "GET", "/customers/"+customerID+"/price-history?scope=sold&productId="+productID, nil, 200)["data"].([]any)
		if len(sold) != 1 {
			t.Fatal("customer sale record missing", sold)
		}
		// Supplier edits and deletion keep both offers and the accepted line's cost.
		call(admin, "PATCH", "/procurement/supplier-quotes/"+supplier["id"].(string), m("supplierName", "Meeting Vendor", "currency", "SAR", "items", []any{m("productId", productID, "productName", "Meeting camera", "quantity", 2, "unitCost", 90)}), 200)
		archive := call(admin, "GET", "/products/"+productID+"/quote-offers", nil, 200)["data"].([]any)
		if len(archive) != 2 || archive[0].(map[string]any)["unitCost"] != float64(90) || archive[1].(map[string]any)["unitCost"] != float64(70) || archive[1].(map[string]any)["isCurrent"] != false {
			t.Fatal("supplier history overwritten", archive)
		}
		frozen := call(admin, "GET", "/quotes/"+quoteID, nil, 200)
		if frozen["lineItems"].([]any)[0].(map[string]any)["unitCost"] != float64(70) {
			t.Fatal("accepted cost was rewritten")
		}
		call(admin, "DELETE", "/procurement/supplier-quotes/"+supplier["id"].(string), nil, 204)
		if len(call(admin, "GET", "/products/"+productID+"/quote-offers", nil, 200)["data"].([]any)) != 2 {
			t.Fatal("deleting supplier quote erased offer history")
		}
		otherCustomer := call(admin, "POST", "/customers", m("companyName", "Other price-history customer", "sector", "other", "region", "Riyadh", "status", "active", "type", "grow"), 201)
		otherID := otherCustomer["id"].(string)
		otherPath := "/products/" + productID + "/sales-history?customerId=" + otherID + "&scope=others"
		compared := call(memberToken, "GET", otherPath, nil, 200)["data"].([]any)
		if len(compared) != 1 || compared[0].(map[string]any)["customerId"] != customerID {
			t.Fatal("other-customer comparison missing", compared)
		}
		if len(call(admin, "GET", "/products/"+productID+"/sales-history?customerId="+otherID, nil, 200)["data"].([]any)) != 0 {
			t.Fatal("customer history mixed other customers")
		}
		call(admin, "PATCH", "/users/"+member["id"].(string), m("quotePermissions", m("quotes:compare-prices", false)), 200)
		call(memberToken, "GET", otherPath, nil, 403)
		call(memberToken, "GET", historyPath, nil, 200)
		// Losing membership removes that team's historic prices as well.
		call(admin, "DELETE", "/teams/"+teamID+"/members/"+clerk["id"].(string), nil, 204)
		if len(call(clerkToken, "GET", otherPath, nil, 200)["data"].([]any)) != 0 {
			t.Fatal("cross-team sale price leaked")
		}
		call(admin, "PATCH", "/users/"+member["id"].(string), m("quotePermissions", m("quotes:read", false)), 200)
		call(memberToken, "GET", "/quotes/"+quoteID, nil, 403)
	})
}
