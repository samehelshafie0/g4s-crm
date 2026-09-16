package handlers

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"g4s-crm/api/internal/middleware"
	"g4s-crm/api/internal/models"
	"g4s-crm/api/pkg/pagination"
	"g4s-crm/api/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DocumentHandler struct {
	db          *gorm.DB
	storageRoot string
}

func NewDocumentHandler(db *gorm.DB, storageRoot string) *DocumentHandler {
	return &DocumentHandler{db: db, storageRoot: storageRoot}
}

func (h *DocumentHandler) List(c *gin.Context) {
	params := pagination.GetParams(c)
	var items []models.Document
	var total int64
	query := h.db.Model(&models.Document{})
	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}
	if q := c.Query("q"); q != "" {
		query = query.Where("name ILIKE ?", "%"+q+"%")
	}
	query.Count(&total)
	query.Order(params.Sort + " " + params.Order).Scopes(pagination.Paginate(params)).Find(&items)
	response.OKWithMeta(c, items, pagination.BuildMeta(params, total))
}

func (h *DocumentHandler) Get(c *gin.Context) {
	var item models.Document
	if err := h.db.Preload("Links").First(&item, "id = ?", c.Param("id")).Error; err != nil {
		response.NotFound(c, "Document not found")
		return
	}
	response.OK(c, item)
}

func (h *DocumentHandler) Upload(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(50 << 20); err != nil {
		response.BadRequest(c, "Failed to parse form")
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "File is required")
		return
	}
	defer file.Close()

	name := c.PostForm("name")
	if name == "" {
		name = header.Filename
	}
	category := c.PostForm("category")
	docType := c.PostForm("documentType")
	version := c.PostForm("version")

	filePath, err := h.saveFile(file, header, category)
	if err != nil {
		response.InternalError(c, "Failed to save file")
		return
	}

	userID := middleware.GetCurrentUserID(c)
	doc := &models.Document{
		Name:         name,
		Category:     models.DocumentCategory(category),
		DocumentType: models.DocumentType(docType),
		Version:      version,
		FileName:     header.Filename,
		FileSize:     header.Size,
		FileType:     header.Header.Get("Content-Type"),
		FilePath:     filePath,
		UploadedByID: &userID,
	}

	if err := h.db.Create(doc).Error; err != nil {
		os.Remove(filepath.Join(h.storageRoot, filePath))
		response.InternalError(c, "Failed to save document metadata")
		return
	}

	response.Created(c, doc)
}

func (h *DocumentHandler) Download(c *gin.Context) {
	var doc models.Document
	if err := h.db.First(&doc, "id = ?", c.Param("id")).Error; err != nil {
		response.NotFound(c, "Document not found")
		return
	}

	fullPath := filepath.Join(h.storageRoot, doc.FilePath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		response.NotFound(c, "File not found on disk")
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, doc.FileName))
	c.File(fullPath)
}

func (h *DocumentHandler) Delete(c *gin.Context) {
	var doc models.Document
	if err := h.db.First(&doc, "id = ?", c.Param("id")).Error; err != nil {
		response.NotFound(c, "Document not found")
		return
	}
	fullPath := filepath.Join(h.storageRoot, doc.FilePath)
	os.Remove(fullPath)
	h.db.Delete(&doc)
	response.NoContent(c)
}

func (h *DocumentHandler) saveFile(file multipart.File, header *multipart.FileHeader, category string) (string, error) {
	subDir := filepath.Join(h.storageRoot, "documents", category)
	if err := os.MkdirAll(subDir, 0750); err != nil {
		return "", err
	}

	ext := filepath.Ext(header.Filename)
	fileName := uuid.New().String() + ext
	fullPath := filepath.Join(subDir, fileName)

	dst, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	return filepath.Join("documents", category, fileName), nil
}
