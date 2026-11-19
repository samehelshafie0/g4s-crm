package response

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestQuotationVisibility(t *testing.T) {
	ctx := func(path string) *gin.Context {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest("GET", path, nil)
		c.Set("hideQuoteMargins", true)
		return c
	}
	quote := map[string]any{"quoteNumber": "QT-1", "margin": 25, "lineItems": []any{map[string]any{"unitCost": 75, "unitPrice": 100}}}
	for _, path := range []string{"/api/v1/quotes/1", "/api/v1/opportunities/1"} {
		out := visibleData(ctx(path), quote).(map[string]any)
		if _, exists := out["margin"]; exists {
			t.Fatal("quote margin leaked", path)
		}
		line := out["lineItems"].([]any)[0].(map[string]any)
		if _, exists := line["unitCost"]; exists {
			t.Fatal("quote cost leaked", path)
		}
		if line["unitPrice"] != float64(100) {
			t.Fatal("selling price lost", path)
		}
	}
	catalog := map[string]any{"unitCost": 75, "unitPrice": 100}
	if _, exists := visibleData(ctx("/api/v1/catalog"), catalog).(map[string]any)["unitCost"]; exists {
		t.Fatal("builder catalog cost leaked")
	}
	if _, exists := visibleData(ctx("/api/v1/products/1"), catalog).(map[string]any)["unitCost"]; !exists {
		t.Fatal("quote override changed product module permissions")
	}
	c := ctx("/api/v1/contracts/1/quotes")
	c.Set("quoteTeams", map[string]bool{"mine": true})
	scoped := visibleData(c, []any{map[string]any{"quoteNumber": "QT-1", "teamId": "other"}, map[string]any{"quoteNumber": "QT-2", "teamId": "mine"}}).([]any)
	if len(scoped) != 1 {
		t.Fatal("another team's quotation leaked", scoped)
	}
	c.Set("hideQuotes", true)
	if len(visibleData(c, []any{quote}).([]any)) != 0 {
		t.Fatal("read-denied quotation leaked")
	}
	c.Set("operationalQuoteResponse", true)
	if visibleData(c, map[string]any{"quoteNumber": "QT-1", "quantity": 2}) == nil {
		t.Fatal("operational quantity record removed")
	}
}
