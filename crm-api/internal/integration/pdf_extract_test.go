package integration

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func testPDFExtract(t *testing.T, r *gin.Engine, token string) {
	t.Run("a vendor PDF is read into priced table rows", func(t *testing.T) {
		if os.Getenv("PDF_PYTHON") == "" {
			t.Skip("set PDF_PYTHON to an interpreter with pymupdf to run PDF extraction tests")
		}
		fixture, err := os.ReadFile("testdata/price-list.pdf")
		must(t, err)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", "vendor-price-list.pdf")
		must(t, err)
		_, err = part.Write(fixture)
		must(t, err)
		must(t, writer.Close())

		req := httptest.NewRequest("POST", "/api/v1/extract/pdf-tables", body)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Fatalf("extraction returned %d: %s", w.Code, w.Body.String())
		}

		var out struct {
			Data struct {
				FileName string `json:"fileName"`
				Tables   []struct {
					LineItemRows int        `json:"lineItemRows"`
					Rows         [][]string `json:"rows"`
				} `json:"tables"`
			} `json:"data"`
		}
		must(t, json.Unmarshal(w.Body.Bytes(), &out))
		if len(out.Data.Tables) == 0 {
			t.Fatalf("no tables extracted: %s", w.Body.String())
		}
		table := out.Data.Tables[0]
		if table.LineItemRows < 2 {
			t.Fatalf("expected the priced lines to be found, got %d: %v", table.LineItemRows, table.Rows)
		}
		flat := ""
		for _, row := range table.Rows {
			for _, cell := range row {
				flat += cell + "|"
			}
		}
		for _, want := range []string{"CAM-4MP-DOME", "1250.00", "Recorder"} {
			if !bytes.Contains([]byte(flat), []byte(want)) {
				t.Fatalf("extracted rows are missing %q: %s", want, flat)
			}
		}

		// A non-PDF is refused rather than handed to the reader.
		other := &bytes.Buffer{}
		w2 := multipart.NewWriter(other)
		part2, err := w2.CreateFormFile("file", "list.xlsx")
		must(t, err)
		_, err = part2.Write([]byte("not a pdf"))
		must(t, err)
		must(t, w2.Close())
		req2 := httptest.NewRequest("POST", "/api/v1/extract/pdf-tables", other)
		req2.Header.Set("Authorization", "Bearer "+token)
		req2.Header.Set("Content-Type", w2.FormDataContentType())
		rec2 := httptest.NewRecorder()
		r.ServeHTTP(rec2, req2)
		if rec2.Code != 400 {
			t.Fatalf("a spreadsheet should be refused here, got %d", rec2.Code)
		}
	})
}
