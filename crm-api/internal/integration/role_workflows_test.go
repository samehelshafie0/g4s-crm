package integration

import (
	"bytes"
	"encoding/json"
	"g4s-crm/api/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http/httptest"
	"testing"
)

func testRoleWorkflows(t *testing.T, db *gorm.DB, r *gin.Engine, admin string) {
	call := func(t *testing.T, token, method, path string, body any, expected int) map[string]any {
		t.Helper()
		payload, err := json.Marshal(body)
		must(t, err)
		req := httptest.NewRequest(method, "/api/v1"+path, bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		res := httptest.NewRecorder()
		r.ServeHTTP(res, req)
		if res.Code != expected {
			t.Fatalf("%s %s: %d want %d: %s", method, path, res.Code, expected, res.Body.String())
		}
		result := map[string]any{}
		if res.Body.Len() > 0 {
			must(t, json.Unmarshal(res.Body.Bytes(), &result))
		}
		if data, ok := result["data"].(map[string]any); ok {
			return data
		}
		return result
	}
	m := func(pairs ...any) map[string]any {
		result := map[string]any{}
		for i := 0; i < len(pairs); i += 2 {
			result[pairs[i].(string)] = pairs[i+1]
		}
		return result
	}
	password := "RoleWorkflow123!"
	createUser := func(t *testing.T, email, role string) map[string]any {
		return call(t, admin, "POST", "/users", m("firstName", "Workflow", "lastName", role, "email", email, "password", password, "role", role, "department", "sales"), 201)
	}
	login := func(t *testing.T, email, secret string) map[string]any {
		return call(t, "", "POST", "/auth/login", m("email", email, "password", secret), 200)
	}
	var sales, manager, warehouse string
	var salesID string
	t.Run("admin user management profile and password sessions", func(t *testing.T) {
		user := createUser(t, "workflow-sales@example.test", "sales_executive")
		salesID = user["id"].(string)
		createUser(t, "workflow-manager@example.test", "sales_manager")
		createUser(t, "workflow-warehouse@example.test", "warehouse_manager")
		session := login(t, "WORKFLOW-SALES@example.test", password)
		sales = session["accessToken"].(string)
		manager = login(t, "workflow-manager@example.test", password)["accessToken"].(string)
		warehouse = login(t, "workflow-warehouse@example.test", password)["accessToken"].(string)
		permissions := session["user"].(map[string]any)["permissions"].([]any)
		for _, permission := range permissions {
			if permission == "quotes:approve" || permission == "users:update" {
				t.Fatal("Sales received administrative permissions")
			}
		}
		call(t, sales, "PATCH", "/auth/me", m("role", "admin"), 400)
		profile := call(t, sales, "PATCH", "/auth/me", m("phone", "+966500000001"), 200)
		if profile["phone"] != "+966500000001" {
			t.Fatal(profile)
		}
		call(t, sales, "PATCH", "/users/"+salesID, m("role", "admin"), 403)
		call(t, manager, "POST", "/users", m(), 403)
		call(t, admin, "PATCH", "/users/"+salesID+"/password", m("password", "NewRolePassword123!"), 204)
		call(t, "", "POST", "/auth/refresh", m("refreshToken", session["refreshToken"]), 401)
		call(t, "", "POST", "/auth/login", m("email", "workflow-sales@example.test", "password", password), 401)
		session = login(t, "workflow-sales@example.test", "NewRolePassword123!")
		sales = session["accessToken"].(string)
		call(t, sales, "PATCH", "/auth/change-password", m("currentPassword", "wrong-password", "newPassword", password), 400)
		call(t, sales, "PATCH", "/auth/change-password", m("currentPassword", "NewRolePassword123!", "newPassword", password), 200)
		call(t, "", "POST", "/auth/refresh", m("refreshToken", session["refreshToken"]), 401)
		sales = login(t, "workflow-sales@example.test", password)["accessToken"].(string)
		team := call(t, admin, "POST", "/teams", m("name", "Workflow team", "department", "sales"), 201)
		call(t, admin, "PATCH", "/users/"+salesID, m("teamId", team["id"]), 200)
		call(t, admin, "DELETE", "/users/"+salesID, nil, 204)
		call(t, sales, "GET", "/auth/me", nil, 401)
		call(t, admin, "PATCH", "/users/"+salesID, m("isActive", true), 200)
		sales = login(t, "workflow-sales@example.test", password)["accessToken"].(string)
	})
	t.Run("separate sales and manager complete the persisted sales journey", func(t *testing.T) {
		customer := call(t, sales, "POST", "/customers", m("companyName", "Role workflow customer", "sector", "other", "region", "Riyadh", "status", "prospect", "type", "get"), 201)
		opp := call(t, sales, "POST", "/opportunities", m("title", "Role acceptance deal", "customerId", customer["id"], "salesExecutiveId", salesID, "estimatedValue", 1000), 201)
		service := call(t, admin, "POST", "/services", m("sku", "ROLE-SVC", "name", "Installation", "unitCost", 50, "unitPrice", 100), 201)
		call(t, sales, "POST", "/services", m("sku", "DENIED", "name", "Denied"), 403)
		q := call(t, sales, "POST", "/quotes", m("customerId", customer["id"], "opportunityId", opp["id"], "vatPercent", 15), 201)
		path := "/quotes/" + q["id"].(string)
		payload := m("lockVersion", q["lockVersion"], "customerId", customer["id"], "opportunityId", opp["id"], "currency", "SAR", "vatPercent", 15, "rows", []any{m("rowType", "item", "source", "service", "serviceId", service["id"], "description", "Installation", "quantity", 2, "multiplier", 1, "unitCost", 50, "unitPrice", 100, "isSelected", true, "isPrintable", true)})
		saved := call(t, sales, "PUT", path+"/builder", payload, 200)
		reread := call(t, manager, "GET", path+"/builder", nil, 200)
		if reread["total"] != float64(230) || len(reread["lineItems"].([]any)) != 1 {
			t.Fatal(reread)
		}
		call(t, sales, "PUT", path+"/builder", payload, 409)
		call(t, sales, "PATCH", path, m("status", "approved"), 400)
		call(t, sales, "PATCH", path+"/submit", m(), 200)
		call(t, sales, "PATCH", path+"/approve", m(), 403)
		call(t, manager, "PATCH", path+"/reject", m(), 200)
		call(t, sales, "PATCH", path+"/submit", m(), 200)
		call(t, manager, "PATCH", path+"/approve", m(), 200)
		payload["lockVersion"] = saved["lockVersion"]
		call(t, sales, "PUT", path+"/builder", payload, 422)
		call(t, sales, "PATCH", path+"/send", m(), 200)
		call(t, sales, "PATCH", path+"/accept", m(), 200)
		dates := m("title", "Role workflow contract", "type", "sales", "startDate", "2026-09-17", "endDate", "2027-09-17")
		call(t, sales, "POST", path+"/convert-to-contract", dates, 403)
		contract := call(t, manager, "POST", path+"/convert-to-contract", dates, 201)
		again := call(t, manager, "POST", path+"/convert-to-contract", dates, 200)
		if contract["id"] != again["id"] || contract["value"] != float64(230) {
			t.Fatal(contract, again)
		}
		cp := "/contracts/" + contract["id"].(string)
		call(t, sales, "PATCH", cp+"/activate", m(), 403)
		call(t, manager, "PATCH", cp+"/activate", m(), 200)
		active := call(t, sales, "GET", cp, nil, 200)
		if active["status"] != "active" {
			t.Fatal(active)
		}
		call(t, manager, "GET", path+"/activity", nil, 200)
	})
	t.Run("warehouse fulfillment and reorder policy enforce stock integrity", func(t *testing.T) {
		product := call(t, admin, "POST", "/products", m("sku", "ROLE-STOCK", "name", "Role stock"), 201)
		stock := call(t, warehouse, "POST", "/inventory/movements/adjustment", m("productId", product["id"], "warehouseLocation", "riyadh-main", "qty", 3, "unitCost", 25, "reason", "Opening balance"), 200)
		call(t, warehouse, "PATCH", "/inventory/stock/"+stock["id"].(string)+"/reorder-level", m("reorderLevel", 2), 200)
		hold := call(t, warehouse, "POST", "/inventory/reservations", m("productId", product["id"], "warehouseLocation", "riyadh-main", "qty", 2, "source", "manual", "sourceRef", "ROLE-TEST", "customerName", "Role workflow customer"), 201)
		path := "/inventory/reservations/" + hold["id"].(string) + "/fulfill"
		call(t, sales, "PATCH", path, m("reason", "Unauthorized dispatch"), 403)
		call(t, warehouse, "PATCH", path, m(), 400)
		fulfilled := call(t, warehouse, "PATCH", path, m("reason", "Dispatch ROLE-TEST"), 200)
		if fulfilled["status"] != "fulfilled" {
			t.Fatal(fulfilled)
		}
		call(t, warehouse, "PATCH", path, m("reason", "Retry"), 400)
		var after models.WarehouseStock
		must(t, db.First(&after, "id = ?", stock["id"]).Error)
		if after.OnHandQty != 1 || after.ReservedQty != 0 || after.AvailableQty != 1 || *after.ReorderLevel != 2 {
			t.Fatal(after)
		}
	})
}
