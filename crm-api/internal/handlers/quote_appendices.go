package handlers

import (
	"path/filepath"
	"strings"

	"g4s-crm/api/internal/middleware"
	"g4s-crm/api/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type appendixInput struct {
	DocumentVersionID uuid.UUID `json:"documentVersionId" validate:"required"`
	Label             string    `json:"label" validate:"required,max=255"`
}

// The builder's quote-row lock protects both line items and ordered appendices.
// Pin immutable versions and snapshot their names, rather than following latest files.
func replaceAppendices(tx *gorm.DB, c *gin.Context, quote *models.Quote, input []appendixInput) error {
	var existing []models.QuoteAppendix
	if err := tx.Where("quote_id = ?", quote.ID).Find(&existing).Error; err != nil {
		return err
	}
	pinned := map[uuid.UUID]models.QuoteAppendix{}
	for _, item := range existing {
		pinned[item.DocumentVersionID] = item
	}
	seen := map[uuid.UUID]bool{}
	items := []models.QuoteAppendix{}
	var total int64
	for index, req := range input {
		if seen[req.DocumentVersionID] {
			return invalid("A document version can only be included once")
		}
		seen[req.DocumentVersionID] = true
		label := strings.TrimSpace(req.Label)
		if label == "" {
			return invalid("Each appendix needs a label")
		}
		item, retained := pinned[req.DocumentVersionID]
		if !retained {
			if !middleware.HasPermission(middleware.GetCurrentUserRole(c), "documents:read") {
				return invalid("Document access is required to add appendices")
			}
			var version models.DocumentVersion
			if err := tx.First(&version, "id = ?", req.DocumentVersionID).Error; err != nil {
				return invalid("Document version is unavailable")
			}
			var doc models.Document
			if err := tx.First(&doc, "id = ?", version.DocumentID).Error; err != nil {
				return invalid("Document is unavailable")
			}
			if !supportedAppendix(version.FileName) {
				return invalid("Appendices support PDF, PNG, JPG, Word (DOC/DOCX) and Excel (XLS/XLSX) files")
			}
			item = models.QuoteAppendix{QuoteID: quote.ID, DocumentID: doc.ID, DocumentVersionID: version.ID, DocumentName: doc.Name, Version: version.Version, FileName: version.FileName, FileType: version.FileType, FileSize: version.FileSize}
		}
		total += item.FileSize
		if total > 100<<20 {
			return invalid("Combined appendix files must not exceed 100 MB")
		}
		item.Base = models.Base{}
		item.Label = label
		item.SortOrder = index
		items = append(items, item)
	}
	if err := tx.Unscoped().Where("quote_id = ?", quote.ID).Delete(&models.QuoteAppendix{}).Error; err != nil {
		return err
	}
	if len(items) > 0 {
		if err := tx.Create(&items).Error; err != nil {
			return err
		}
	}
	quote.Appendices = items
	return nil
}

func supportedAppendix(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".pdf", ".png", ".jpg", ".jpeg", ".doc", ".docx", ".xls", ".xlsx":
		return true
	}
	return false
}
