package integration

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"g4s-crm/api/internal/bootstrap"
	"g4s-crm/api/internal/config"
	"g4s-crm/api/internal/middleware"
	"g4s-crm/api/internal/models"
	"g4s-crm/api/internal/router"
	"g4s-crm/api/internal/services"
	"g4s-crm/api/migrations"
	"g4s-crm/api/pkg/seqgen"
	v "g4s-crm/api/pkg/validator"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// Only an explicitly supplied disposable database is used. Each run owns a schema.
func TestFoundation(t *testing.T) {
	dsn := os.Getenv("CRM_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set CRM_TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	must(t, err)
	sqlDB, err := db.DB()
	must(t, err)
	defer sqlDB.Close()
	schema := "crm_test_" + uuid.New().String()[:8]
	must(t, db.Exec("CREATE SCHEMA "+schema).Error)
	defer db.Exec("DROP SCHEMA " + schema + " CASCADE")
	// A single connection keeps search_path fixed for the initial setup. The separate
	// test pool below configures it per connection so concurrent tests use this schema.
	testDSN := dsn
	separator := "?"
	if bytes.Contains([]byte(dsn), []byte("?")) {
		separator = "&"
	}
	testDSN += separator + "search_path=" + schema + ",public"
	scoped, err := gorm.Open(postgres.Open(testDSN), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	must(t, err)
	scopedSQL, err := scoped.DB()
	must(t, err)
	defer scopedSQL.Close()
	ctx := context.Background()
	must(t, migrations.Up(ctx, scopedSQL))
	must(t, migrations.Up(ctx, scopedSQL))
	must(t, migrations.Ready(ctx, scopedSQL))
	db = scoped
	dir := t.TempDir()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	must(t, err)
	privatePath := filepath.Join(dir, "private.pem")
	publicPath := filepath.Join(dir, "public.pem")
	must(t, os.WriteFile(privatePath, pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}), 0600))
	pub, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	must(t, err)
	must(t, os.WriteFile(publicPath, pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pub}), 0600))
	cfg := &config.Config{JWT: config.JWTConfig{PrivateKeyPath: privatePath, PublicKeyPath: publicPath, AccessExpiry: time.Hour, RefreshExpiry: 24 * time.Hour}, Storage: config.StorageConfig{Root: dir}}
	gin.SetMode(gin.TestMode)
	v.Setup()
	must(t, middleware.LoadPublicKey(publicPath))
	r := router.Setup(db, cfg)
	must(t, bootstrap.Admin(db, "admin@example.test", "TestPassword123!", "Test", "Admin"))
	if err := bootstrap.Admin(db, "second@example.test", "TestPassword123!", "Second", "Admin"); err == nil {
		t.Fatal("bootstrap must not create a second admin")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("TestPassword123!"), bcrypt.MinCost)
	must(t, err)
	sales := models.User{Email: "sales@example.test", FirstName: "Sales", LastName: "User", PasswordHash: string(hash), Role: models.RoleSalesExecutive, Department: models.DeptSales, IsActive: true}
	must(t, db.Create(&sales).Error)
	call := func(method, path, body, token string) (int, map[string]any) {
		request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
		request.Header.Set("Content-Type", "application/json")
		if token != "" {
			request.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, request)
		result := map[string]any{}
		if w.Body.Len() > 0 {
			must(t, json.Unmarshal(w.Body.Bytes(), &result))
		}
		return w.Code, result
	}
	expect := func(t *testing.T, code, want int, result map[string]any) {
		t.Helper()
		if code != want {
			t.Fatalf("HTTP %d, want %d: %v", code, want, result)
		}
	}
	login := func(email string) string {
		code, res := call("POST", "/api/v1/auth/login", fmt.Sprintf(`{"email":%q,"password":"TestPassword123!"}`, email), "")
		expect(t, code, 200, res)
		return res["data"].(map[string]any)["accessToken"].(string)
	}
	adminToken := login("admin@example.test")
	salesToken := login("sales@example.test")
	t.Run("required validation and full user profile", func(t *testing.T) {
		code, res := call("POST", "/api/v1/auth/login", `{}`, "")
		expect(t, code, 400, res)
		code, res = call("GET", "/api/v1/auth/me", "", salesToken)
		expect(t, code, 200, res)
		u := res["data"].(map[string]any)
		if u["firstName"] != "Sales" || u["lastName"] != "User" || u["passwordHash"] != nil {
			t.Fatal(u)
		}
	})
	var customerID string
	t.Run("multiple optional CRs and empty collections", func(t *testing.T) {
		for i := 0; i < 2; i++ {
			code, res := call("POST", "/api/v1/customers", `{"companyName":"Test Customer","sector":"government","region":"Riyadh","status":"prospect","type":"get"}`, salesToken)
			expect(t, code, 201, res)
			customer := res["data"].(map[string]any)
			customerID = customer["id"].(string)
			for _, name := range []string{"contacts", "sites"} {
				if rows, ok := customer[name].([]any); !ok || len(rows) != 0 {
					t.Fatalf("%s must be []: %v", name, customer)
				}
			}
		}
		code, res := call("GET", "/api/v1/customers?sort=created_at%3BSELECT%201&order=asc", "", salesToken)
		expect(t, code, 200, res)
	})
	t.Run("customer updates preserve arrays and reject ownership changes", func(t *testing.T) {
		code, res := call("PATCH", "/api/v1/customers/"+customerID, `{"companyName":"Renamed customer","crNumber":""}`, salesToken)
		expect(t, code, 200, res)
		if res["data"].(map[string]any)["companyName"] != "Renamed customer" {
			t.Fatal(res)
		}
		code, res = call("PATCH", "/api/v1/customers/"+customerID, `{"createdById":"anything"}`, salesToken)
		expect(t, code, 400, res)
		code, res = call("PATCH", "/api/v1/customers/"+customerID, `{"sector":"invalid"}`, salesToken)
		expect(t, code, 400, res)
	})
	t.Run("contract creation preserves relationships and rejects generic activation", func(t *testing.T) {
		body := fmt.Sprintf(`{"title":"Test contract","customerId":%q,"type":"sales","startDate":"2026-09-16","endDate":"2027-09-16"}`, customerID)
		code, res := call("POST", "/api/v1/contracts", body, adminToken)
		expect(t, code, 201, res)
		contract := res["data"].(map[string]any)
		id := contract["id"].(string)
		if contract["customerId"] != customerID || contract["startDate"] == nil || contract["endDate"] == nil {
			t.Fatal(contract)
		}
		code, res = call("PATCH", "/api/v1/contracts/"+id, `{"status":"active"}`, adminToken)
		expect(t, code, 400, res)
	})
	t.Run("purchase orders require approval before ordering", func(t *testing.T) {
		po := models.PurchaseOrder{PONumber: "TEST-PO", SupplierName: "Test supplier", Status: models.POStatusDraft, Currency: models.CurrencySAR}
		must(t, db.Create(&po).Error)
		path := "/api/v1/procurement/purchase-orders/" + po.ID.String()
		code, res := call("PATCH", path+"/status", `{"status":"approved"}`, adminToken)
		expect(t, code, 400, res)
		code, res = call("PATCH", path+"/status", `{"status":"received"}`, adminToken)
		expect(t, code, 400, res)
		code, res = call("PATCH", path+"/status", `{"status":"ordered"}`, adminToken)
		expect(t, code, 409, res)
		code, res = call("PATCH", path+"/status", `{"status":"pending-approval"}`, adminToken)
		expect(t, code, 200, res)
		code, res = call("PATCH", path+"/approve", `{}`, adminToken)
		expect(t, code, 200, res)
		code, res = call("PATCH", path+"/status", `{"status":"ordered"}`, adminToken)
		expect(t, code, 200, res)
	})

	var quoteID string
	t.Run("atomic quote create, nested validation and explicit zero VAT", func(t *testing.T) {
		body := fmt.Sprintf(`{"customerId":%q,"vatPercent":0,"lineItems":[{"category":"materials","description":"Test item","quantity":2,"unitCost":25,"unitPrice":100}]}`, customerID)
		code, res := call("POST", "/api/v1/quotes", body, salesToken)
		expect(t, code, 201, res)
		q := res["data"].(map[string]any)
		quoteID = q["id"].(string)
		if q["total"] != float64(200) || q["vatPercent"] != float64(0) {
			t.Fatal(q)
		}
		invalid := fmt.Sprintf(`{"customerId":%q,"lineItems":[{"category":"materials","description":"Bad qty","quantity":0}]}`, customerID)
		code, res = call("POST", "/api/v1/quotes", invalid, salesToken)
		expect(t, code, 400, res)
		invalid = fmt.Sprintf(`{"customerId":%q,"lineItems":[{"category":"materials","description":"Bad FK","quantity":1,"productId":%q}]}`, customerID, uuid.New().String())
		code, res = call("POST", "/api/v1/quotes", invalid, salesToken)
		expect(t, code, 400, res)
		var count int64
		must(t, db.Model(&models.Quote{}).Count(&count).Error)
		if count != 1 {
			t.Fatalf("partial quote persisted: %d", count)
		}
	})
	t.Run("approval bypasses are rejected and transitions freeze edits", func(t *testing.T) {
		for _, payload := range []string{`{"status":"approved"}`, `{"total":999}`, `{"approved_by_id":"anything"}`, `{"approvedAt":"2026-01-01"}`, `{"notes":"ok"} {"status":"approved"}`} {
			code, res := call("PATCH", "/api/v1/quotes/"+quoteID, payload, salesToken)
			expect(t, code, 400, res)
		}
		code, res := call("PATCH", "/api/v1/quotes/"+quoteID, `{"notes":"A persisted note","discountPercent":10}`, salesToken)
		expect(t, code, 200, res)
		if res["data"].(map[string]any)["total"] != float64(180) {
			t.Fatal(res)
		}
		code, res = call("PATCH", "/api/v1/quotes/"+quoteID+"/approve", `{}`, salesToken)
		expect(t, code, 403, res)
		code, res = call("PATCH", "/api/v1/quotes/"+quoteID+"/submit", `{}`, salesToken)
		expect(t, code, 200, res)
		code, res = call("PATCH", "/api/v1/quotes/"+quoteID+"/approve", `{}`, adminToken)
		expect(t, code, 200, res)
		if res["data"].(map[string]any)["approvedAt"] == nil {
			t.Fatal("approval date missing")
		}
		code, res = call("PATCH", "/api/v1/quotes/"+quoteID, `{"notes":"late"}`, salesToken)
		expect(t, code, 422, res)
		code, res = call("POST", "/api/v1/quotes/"+quoteID+"/recalculate", `{}`, adminToken)
		expect(t, code, 422, res)
	})
	t.Run("all roles reject quote approval through generic updates", func(t *testing.T) {
		roles := []models.UserRole{models.RoleAdmin, models.RoleSalesManager, models.RoleSalesExecutive, models.RolePreSales, models.RoleProcurementManager, models.RoleProcurementOfficer, models.RoleWarehouseManager, models.RoleProjectManager, models.RoleViewer}
		for _, role := range roles {
			must(t, db.Model(&sales).Update("role", role).Error)
			code, res := call("PATCH", "/api/v1/quotes/"+quoteID, `{"status":"approved","total":999}`, salesToken)
			if code != 400 && code != 403 {
				t.Fatalf("role %s: %d %v", role, code, res)
			}
		}
		must(t, db.Model(&sales).Update("role", models.RoleSalesExecutive).Error)
		var quote models.Quote
		must(t, db.First(&quote, "id = ?", quoteID).Error)
		if quote.Total == 999 {
			t.Fatal("protected total changed")
		}
	})

	testModuleAPIs(t, db, r, adminToken, customerID)
	testRoleWorkflows(t, db, r, adminToken)
	testQuoteAppendices(t, db, r, adminToken, customerID, dir)
	testSeedCatalog(t, db, r, adminToken)
	testProductImport(t, db, r, adminToken)
	t.Run("current role and deactivation take effect for existing access token", func(t *testing.T) {
		must(t, db.Model(&sales).Update("role", models.RoleViewer).Error)
		code, res := call("POST", "/api/v1/customers", `{}`, salesToken)
		expect(t, code, 403, res)
		must(t, db.Model(&sales).Update("is_active", false).Error)
		code, res = call("GET", "/api/v1/auth/me", "", salesToken)
		expect(t, code, 401, res)
	})
	t.Run("refresh is single use under concurrency", func(t *testing.T) {
		svc, err := services.NewAuthService(db, &cfg.JWT)
		must(t, err)
		session, err := svc.Login("admin@example.test", "TestPassword123!")
		must(t, err)
		results := make(chan bool, 8)
		var group sync.WaitGroup
		for i := 0; i < 8; i++ {
			group.Add(1)
			go func() { defer group.Done(); _, err := svc.Refresh(session.RefreshToken); results <- err == nil }()
		}
		group.Wait()
		close(results)
		successes := 0
		for ok := range results {
			if ok {
				successes++
			}
		}
		if successes != 1 {
			t.Fatalf("%d rotations succeeded", successes)
		}
	})
	t.Run("year rollover and concurrent numbering", func(t *testing.T) {
		must(t, db.Exec("UPDATE sequences SET year = ?, current = 99 WHERE name = 'project'", time.Now().Year()-1).Error)
		first, err := seqgen.NextNumber(db, "project")
		must(t, err)
		if first != fmt.Sprintf("PRJ-%d-0001", time.Now().Year()) {
			t.Fatal(first)
		}
		results := make(chan string, 24)
		failures := make(chan error, 24)
		var group sync.WaitGroup
		for i := 0; i < 24; i++ {
			group.Add(1)
			go func() {
				defer group.Done()
				number, err := seqgen.NextNumber(db, "project")
				if err != nil {
					failures <- err
				} else {
					results <- number
				}
			}()
		}
		group.Wait()
		close(results)
		close(failures)
		for err := range failures {
			t.Fatal(err)
		}
		seen := map[string]bool{}
		for number := range results {
			if seen[number] {
				t.Fatalf("duplicate %s", number)
			}
			seen[number] = true
		}
		if len(seen) != 24 {
			t.Fatal(len(seen))
		}
	})
	t.Run("readiness detects altered migrations", func(t *testing.T) {
		code, res := call("GET", "/ready", "", "")
		expect(t, code, 200, res)
		must(t, db.Exec("UPDATE crm_schema_migrations SET checksum='changed' WHERE name='000002_optional_customer_cr.up.sql'").Error)
		if migrations.Up(ctx, scopedSQL) == nil {
			t.Fatal("changed applied migration accepted")
		}
		code, res = call("GET", "/ready", "", "")
		expect(t, code, 503, res)
	})
}
func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
