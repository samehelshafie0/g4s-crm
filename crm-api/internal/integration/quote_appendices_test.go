package integration

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"g4s-crm/api/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func testQuoteAppendices(t *testing.T, db *gorm.DB, r *gin.Engine, token, customerID, storageRoot string) {
	t.Run("quote appendix order labels pinned versions and immutable revisions", func(t *testing.T) {
		call := func(method, path string, payload any, want int) map[string]any {
			t.Helper()
			body, _ := json.Marshal(payload)
			req := httptest.NewRequest(method, "/api/v1"+path, bytes.NewReader(body))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != want {
				t.Fatalf("%s %s: %d want %d: %s", method, path, w.Code, want, w.Body.String())
			}
			result := map[string]any{}
			if w.Body.Len() > 0 {
				must(t, json.Unmarshal(w.Body.Bytes(), &result))
			}
			if data, ok := result["data"].(map[string]any); ok {
				return data
			}
			return result
		}
		fixture, err := os.ReadFile("testdata/appendix.pdf")
		must(t, err)
		must(t, os.WriteFile(filepath.Join(storageRoot, "appendix.pdf"), fixture, 0600))
		doc := models.Document{Name: "Customer specification", Category: "general", DocumentType: "technical", Version: "1", FileName: "specification.pdf", FilePath: "appendix.pdf", FileType: "application/pdf", FileSize: int64(len(fixture))}
		must(t, db.Create(&doc).Error)
		version := models.DocumentVersion{DocumentID: doc.ID, Version: "1", FileName: doc.FileName, FilePath: doc.FilePath, FileType: doc.FileType, FileSize: doc.FileSize}
		must(t, db.Create(&version).Error)
		doc2 := models.Document{Name: "Agreement", Category: "contract", DocumentType: "terms", Version: "1", FileName: "agreement.pdf", FilePath: "appendix.pdf", FileType: "application/pdf", FileSize: int64(len(fixture))}
		must(t, db.Create(&doc2).Error)
		version2 := models.DocumentVersion{DocumentID: doc2.ID, Version: "1", FileName: doc2.FileName, FilePath: doc2.FilePath, FileType: doc2.FileType, FileSize: doc2.FileSize}
		must(t, db.Create(&version2).Error)
		q := call("POST", "/quotes", map[string]any{"customerId": customerID}, 201)
		path := "/quotes/" + q["id"].(string)
		payload := map[string]any{"lockVersion": q["lockVersion"], "customerId": customerID, "currency": "SAR", "vatPercent": 15, "rows": []any{map[string]any{"rowType": "item", "source": "write-in", "description": "Installation", "quantity": 2, "multiplier": 1, "unitPrice": 100, "unitCost": 50, "isPrintable": true, "isSelected": true}}, "appendices": []any{map[string]any{"documentVersionId": version2.ID, "label": "Customer agreement"}, map[string]any{"documentVersionId": version.ID, "label": "Technical specification"}}}
		saved := call("PUT", path+"/builder", payload, 200)
		got := call("GET", path+"/builder", nil, 200)
		items := got["appendices"].([]any)
		if len(items) != 2 || items[0].(map[string]any)["documentVersionId"] != version2.ID.String() || items[1].(map[string]any)["label"] != "Technical specification" {
			t.Fatal(items)
		}
		call("PUT", path+"/builder", payload, 409)
		payload["lockVersion"] = saved["lockVersion"]
		valid := payload["appendices"]
		payload["appendices"] = []any{map[string]any{"documentVersionId": version.ID, "label": "One"}, map[string]any{"documentVersionId": version.ID, "label": "Duplicate"}}
		call("PUT", path+"/builder", payload, 400)
		if current := call("GET", path+"/builder", nil, 200); current["lockVersion"] != saved["lockVersion"] || len(current["appendices"].([]any)) != 2 {
			t.Fatal("Invalid save changed the quote", current)
		}
		payload["appendices"] = valid
		// Omitting the optional appendices field preserves selections for older clients.
		delete(payload, "appendices")
		saved = call("PUT", path+"/builder", payload, 200)
		if len(saved["appendices"].([]any)) != 2 {
			t.Fatal("Old client save removed appendices")
		}
		must(t, db.Model(&doc).Updates(map[string]any{"version": "2", "name": "New library title"}).Error)
		pinned := call("GET", path+"/builder", nil, 200)["appendices"].([]any)[1].(map[string]any)
		if pinned["version"] != "1" || pinned["documentName"] != "Customer specification" {
			t.Fatal(pinned)
		}
		clone := call("POST", path+"/revisions", nil, 201)
		cloned := call("GET", "/quotes/"+clone["id"].(string)+"/builder", nil, 200)
		if len(cloned["appendices"].([]any)) != 2 {
			t.Fatal(cloned)
		}
		call("PATCH", path+"/submit", nil, 200)
		payload["appendices"] = []any{}
		payload["lockVersion"] = saved["lockVersion"]
		call("PUT", path+"/builder", payload, 422)
		call("GET", path+"/pdf?lockVersion=1", nil, 409)
		// Library soft deletion does not invalidate files in an approved quote snapshot.
		must(t, db.Delete(&doc).Error)
		if os.Getenv("PDF_PYTHON") != "" {
			req := httptest.NewRequest("GET", "/api/v1"+path+"/pdf", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != 200 || !bytes.HasPrefix(w.Body.Bytes(), []byte("%PDF")) || w.Header().Get("Content-Type") != "application/pdf" {
				t.Fatal(w.Code, w.Body.String())
			}
		}
		req := httptest.NewRequest("GET", "/api/v1"+path+"/pdf", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != 401 {
			t.Fatal("Anonymous PDF export allowed", w.Code)
		}
		// A new quote cannot select a removed document, though its pinned copy remains usable.
		clonePath := "/quotes/" + clone["id"].(string)
		payload["lockVersion"] = cloned["lockVersion"]
		payload["appendices"] = []any{}
		cleared := call("PUT", clonePath+"/builder", payload, 200)
		payload["lockVersion"] = cleared["lockVersion"]
		payload["appendices"] = valid
		call("PUT", clonePath+"/builder", payload, 400)
	})
}
