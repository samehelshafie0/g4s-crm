package handlers

import (
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"g4s-crm/api/internal/models"
	"g4s-crm/api/internal/services"
	"g4s-crm/api/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var pdfJobs = make(chan struct{}, 2)

type pdfAppendix struct {
	models.QuoteAppendix
	Path string `json:"path"`
}

// ExportQuote is authorized by quotes:read. Pinned files belong to the readable
// quotation even if their library record was subsequently soft-deleted.
func (h *DocumentHandler) ExportQuote(c *gin.Context) {
	select {
	case pdfJobs <- struct{}{}:
		defer func() { <-pdfJobs }()
	default:
		c.JSON(429, gin.H{"success": false, "error": gin.H{"message": "PDF exports are busy. Try again shortly."}})
		return
	}
	var quote models.Quote
	var versions []models.DocumentVersion
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "SHARE"}).Preload("Customer").Preload("LineItems", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC") }).Preload("Appendices", func(db *gorm.DB) *gorm.DB { return db.Order("sort_order ASC") }).First(&quote, "id = ?", c.Param("id")).Error; err != nil {
			return err
		}
		if expected := c.Query("lockVersion"); expected != "" && expected != strconv.Itoa(quote.LockVersion) {
			return errStaleQuote
		}
		for _, item := range quote.Appendices {
			var version models.DocumentVersion
			if err := tx.First(&version, "id = ? AND document_id = ?", item.DocumentVersionID, item.DocumentID).Error; err != nil {
				return err
			}
			versions = append(versions, version)
		}
		return nil
	})
	if err == errStaleQuote {
		response.Conflict(c, "Quote changed; reload it before exporting")
		return
	}
	if err != nil {
		apiError(c, err)
		return
	}
	directory, err := os.MkdirTemp("", "crm-quote-pdf-")
	if err != nil {
		apiError(c, err)
		return
	}
	defer os.RemoveAll(directory)
	appendices := []pdfAppendix{}
	if len(versions) > 0 {
		root, err := os.OpenRoot(h.storageRoot)
		if err != nil {
			apiError(c, err)
			return
		}
		defer root.Close()
		var total int64
		for i, version := range versions {
			if !filepath.IsLocal(version.FilePath) || !supportedAppendix(version.FileName) {
				response.BadRequest(c, "An appendix file is invalid")
				return
			}
			file, err := root.Open(version.FilePath)
			if err != nil {
				response.NotFound(c, "An appendix file is missing; restore the file before exporting")
				return
			}
			path := filepath.Join(directory, fmt.Sprintf("source-%d%s", i, strings.ToLower(filepath.Ext(version.FileName))))
			dst, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0600)
			if err != nil {
				file.Close()
				apiError(c, err)
				return
			}
			n, copyErr := io.Copy(dst, io.LimitReader(file, (50<<20)+1))
			file.Close()
			dst.Close()
			total += n
			if copyErr != nil {
				apiError(c, copyErr)
				return
			}
			if n > 50<<20 || total > 100<<20 {
				response.BadRequest(c, "Appendices exceed the export size limit")
				return
			}
			appendices = append(appendices, pdfAppendix{QuoteAppendix: quote.Appendices[i], Path: path})
		}
	}
	output, err := services.RenderQuotePDF(c.Request.Context(), directory, gin.H{"quote": quote, "appendices": appendices})
	if err != nil {
		c.JSON(422, gin.H{"success": false, "error": gin.H{"message": err.Error()}})
		return
	}
	c.Header("Content-Type", "application/pdf")
	c.Header("Cache-Control", "no-store")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": quote.QuoteNumber + ".pdf"}))
	c.File(output)
}
