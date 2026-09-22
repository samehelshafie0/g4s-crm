package handlers

import (
	"g4s-crm/api/internal/middleware"
	"g4s-crm/api/internal/models"
	"g4s-crm/api/pkg/pagination"
	"g4s-crm/api/pkg/response"
	v "g4s-crm/api/pkg/validator"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type DocumentHandler struct {
	db          *gorm.DB
	storageRoot string
}

func NewDocumentHandler(db *gorm.DB, root string) *DocumentHandler {
	return &DocumentHandler{db: db, storageRoot: root}
}
func (h *DocumentHandler) List(c *gin.Context) {
	params := pagination.GetParams(c, "name", "category")
	items := []models.Document{}
	var total int64
	q := h.db.Model(&models.Document{})
	if cat := c.Query("category"); cat != "" {
		q = q.Where("category = ?", cat)
	}
	if search := c.Query("q"); search != "" {
		q = q.Where("name ILIKE ?", "%"+search+"%")
	}
	if entity := c.Query("entityId"); entity != "" {
		q = q.Where("id IN (SELECT document_id FROM document_links WHERE entity_id = ? AND deleted_at IS NULL)", entity)
	}
	if err := q.Count(&total).Error; err != nil {
		apiError(c, err)
		return
	}
	if err := q.Preload("Links").Preload("UploadedBy").Order(params.Sort + " " + params.Order).Scopes(pagination.Paginate(params)).Find(&items).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OKWithMeta(c, items, pagination.BuildMeta(params, total))
}
func (h *DocumentHandler) Get(c *gin.Context) {
	var item models.Document
	if err := h.db.Preload("Links").Preload("UploadedBy").First(&item, "id = ?", c.Param("id")).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OK(c, item)
}
func (h *DocumentHandler) Upload(c *gin.Context)        { h.upload(c, false) }
func (h *DocumentHandler) UploadVersion(c *gin.Context) { h.upload(c, true) }
func (h *DocumentHandler) upload(c *gin.Context, newVersion bool) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, (50<<20)+(1<<20))
	if err := c.Request.ParseMultipartForm(2 << 20); err != nil {
		response.BadRequest(c, "Upload must be a multipart form of at most 50 MB")
		return
	}
	defer c.Request.MultipartForm.RemoveAll()
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "File is required")
		return
	}
	defer file.Close()
	if header.Size <= 0 || header.Size > 50<<20 {
		response.BadRequest(c, "File must contain between 1 byte and 50 MB")
		return
	}
	// The original name/category never contributes to the storage path.
	name := c.PostForm("name")
	if name == "" {
		name = filepath.Base(header.Filename)
	}
	category := c.DefaultPostForm("category", "general")
	docType := c.DefaultPostForm("documentType", "technical")
	if !newVersion && (!oneOf(category, "contract", "quote", "general", "compliance", "legal") || !oneOf(docType, "terms", "delivery", "technical", "warranty", "sla")) {
		response.BadRequest(c, "Invalid document category or type")
		return
	}
	tags := pq.StringArray{}
	for _, tag := range strings.Split(c.PostForm("tags"), ",") {
		if t := strings.TrimSpace(tag); t != "" {
			tags = append(tags, t)
		}
	}
	dir := filepath.Join(h.storageRoot, "documents")
	if err := os.MkdirAll(dir, 0750); err != nil {
		apiError(c, err)
		return
	}
	relative := filepath.Join("documents", uuid.NewString())
	full := filepath.Join(h.storageRoot, relative)
	dst, err := os.OpenFile(full, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		apiError(c, err)
		return
	}
	written, copyErr := io.Copy(dst, io.LimitReader(file, (50<<20)+1))
	closeErr := dst.Close()
	if copyErr != nil || closeErr != nil || written > 50<<20 {
		os.Remove(full)
		response.BadRequest(c, "Unable to store upload")
		return
	}
	var doc models.Document
	user := middleware.GetCurrentUserID(c)
	err = h.db.Transaction(func(tx *gorm.DB) error {
		version := "1"
		if newVersion {
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&doc, "id = ?", c.Param("id")).Error; err != nil {
				return err
			}
			var count int64
			if err := tx.Model(&models.DocumentVersion{}).Where("document_id = ?", doc.ID).Count(&count).Error; err != nil {
				return err
			}
			version = strconv.FormatInt(count+1, 10)
		} else {
			doc.Name = name
			doc.Category = models.DocumentCategory(category)
			doc.DocumentType = models.DocumentType(docType)
			doc.Tags = tags
		}
		doc.Version = version
		doc.FileName = filepath.Base(header.Filename)
		doc.FileSize = written
		doc.FileType = header.Header.Get("Content-Type")
		doc.FilePath = relative
		doc.UploadedByID = &user
		if err := tx.Omit(clause.Associations).Save(&doc).Error; err != nil {
			return err
		}
		snapshot := models.DocumentVersion{DocumentID: doc.ID, Version: version, FileName: doc.FileName, FileSize: doc.FileSize, FileType: doc.FileType, FilePath: relative, UploadedByID: &user}
		if err := tx.Create(&snapshot).Error; err != nil {
			return err
		}
		return recordActivity(tx, c, "document", doc.ID, "uploaded version "+version)
	})
	if err != nil {
		os.Remove(full)
		apiError(c, err)
		return
	}
	response.Created(c, doc)
}
func oneOf(value string, options ...string) bool {
	for _, option := range options {
		if value == option {
			return true
		}
	}
	return false
}
func (h *DocumentHandler) Update(c *gin.Context) {
	var item models.Document
	if err := h.db.First(&item, "id = ?", c.Param("id")).Error; err != nil {
		apiError(c, err)
		return
	}
	if !bindFields(c, &item, map[string]string{"name": "required,max=255", "category": "required,oneof=contract quote general compliance legal", "documentType": "required,oneof=terms delivery technical warranty sla", "tags": "max=100,dive,max=100"}) {
		return
	}
	if err := h.db.Omit(clause.Associations).Save(&item).Error; err != nil {
		apiError(c, err)
		return
	}
	h.Get(c)
}
func (h *DocumentHandler) Versions(c *gin.Context) {
	var doc models.Document
	if err := h.db.First(&doc, "id = ?", c.Param("id")).Error; err != nil {
		apiError(c, err)
		return
	}
	items := []models.DocumentVersion{}
	if err := h.db.Where("document_id = ?", doc.ID).Order("created_at DESC").Find(&items).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OK(c, items)
}
func (h *DocumentHandler) Download(c *gin.Context) {
	var doc models.Document
	if err := h.db.First(&doc, "id = ?", c.Param("id")).Error; err != nil {
		apiError(c, err)
		return
	}
	path, name := doc.FilePath, doc.FileName
	if id := c.Param("versionId"); id != "" {
		var version models.DocumentVersion
		if err := h.db.First(&version, "id = ? AND document_id = ?", id, doc.ID).Error; err != nil {
			apiError(c, err)
			return
		}
		path, name = version.FilePath, version.FileName
	}
	if !filepath.IsLocal(path) {
		response.NotFound(c, "File path is invalid")
		return
	}
	root, err := os.OpenRoot(h.storageRoot)
	if err != nil {
		apiError(c, err)
		return
	}
	defer root.Close()
	file, err := root.Open(path)
	if err != nil {
		response.NotFound(c, "File not found")
		return
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || info.IsDir() {
		response.NotFound(c, "File not found")
		return
	}
	c.Header("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": name}))
	c.Header("Content-Type", "application/octet-stream")
	c.Header("X-Content-Type-Options", "nosniff")
	http.ServeContent(c.Writer, c.Request, name, info.ModTime(), file)
}

// Soft deletion retains version files for restore/retention; no irreversible purge here.
func (h *DocumentHandler) Delete(c *gin.Context) { deleteRecord(c, h.db, &models.Document{}) }
func (h *DocumentHandler) AddLink(c *gin.Context) {
	var req struct {
		EntityType string    `json:"entityType" validate:"required,oneof=customer opportunity quote contract project product purchase-order supplier-quote"`
		EntityID   uuid.UUID `json:"entityId" validate:"required"`
	}
	if !v.BindStrict(c, &req) {
		return
	}
	resource := map[string]string{"customer": "customers", "opportunity": "opportunities", "quote": "quotes", "contract": "contracts", "project": "projects", "product": "products", "purchase-order": "procurement", "supplier-quote": "procurement"}[req.EntityType]
	if !middleware.HasPermission(middleware.GetCurrentUserRole(c), resource+":read") {
		response.Forbidden(c, "Cannot link records you cannot read")
		return
	}
	var link models.DocumentLink
	err := h.db.Transaction(func(tx *gorm.DB) error {
		var doc models.Document
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&doc, "id = ?", c.Param("id")).Error; err != nil {
			return err
		}
		model := map[string]any{"customer": &models.Customer{}, "opportunity": &models.Opportunity{}, "quote": &models.Quote{}, "contract": &models.Contract{}, "project": &models.Project{}, "product": &models.Product{}, "purchase-order": &models.PurchaseOrder{}, "supplier-quote": &models.SupplierQuote{}}[req.EntityType]
		if err := exists(tx, model, req.EntityID); err != nil {
			return err
		}
		column := map[string]string{"customer": "company_name", "opportunity": "title", "quote": "quote_number", "contract": "title", "project": "name", "product": "name", "purchase-order": "po_number", "supplier-quote": "sq_number"}[req.EntityType]
		var name string
		if err := tx.Model(model).Select(column).Where("id = ?", req.EntityID).Scan(&name).Error; err != nil {
			return err
		}
		return tx.Where("document_id = ? AND entity_type = ? AND entity_id = ?", doc.ID, req.EntityType, req.EntityID).FirstOrCreate(&link, models.DocumentLink{DocumentID: doc.ID, EntityType: req.EntityType, EntityID: req.EntityID, EntityName: name}).Error
	})
	if err != nil {
		apiError(c, err)
		return
	}
	response.Created(c, link)
}
func (h *DocumentHandler) DeleteLink(c *gin.Context) {
	result := h.db.Where("document_id = ?", c.Param("id")).Delete(&models.DocumentLink{}, "id = ?", c.Param("linkId"))
	if result.Error != nil {
		apiError(c, result.Error)
		return
	}
	if result.RowsAffected != 1 {
		response.NotFound(c, "Link not found")
		return
	}
	response.NoContent(c)
}
