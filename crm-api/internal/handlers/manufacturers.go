package handlers

import (
	"g4s-crm/api/internal/models"
	"g4s-crm/api/pkg/pagination"
	"g4s-crm/api/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ManufacturerHandler struct{ db *gorm.DB }

func NewManufacturerHandler(db *gorm.DB) *ManufacturerHandler { return &ManufacturerHandler{db} }
func (h *ManufacturerHandler) List(c *gin.Context) {
	params := pagination.GetParams(c, "name", "code", "country")
	items := []models.Manufacturer{}
	var total int64
	q := h.db.Model(&models.Manufacturer{})
	if term := c.Query("q"); term != "" {
		q = q.Where("name ILIKE ? OR code ILIKE ?", "%"+term+"%", "%"+term+"%")
	}
	if kind := c.Query("vendorType"); kind != "" {
		q = q.Where("vendor_type IN ?", []string{kind, "both"})
	}
	if err := q.Count(&total).Error; err != nil {
		apiError(c, err)
		return
	}
	if err := q.Preload("Categories").Order(params.Sort + " " + params.Order).Scopes(pagination.Paginate(params)).Find(&items).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OKWithMeta(c, items, pagination.BuildMeta(params, total))
}
func (h *ManufacturerHandler) Get(c *gin.Context) {
	var item models.Manufacturer
	if err := h.db.Preload("Categories").First(&item, "id = ?", c.Param("id")).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OK(c, item)
}
func (h *ManufacturerHandler) Create(c *gin.Context) { h.save(c, true) }
func (h *ManufacturerHandler) Update(c *gin.Context) { h.save(c, false) }
func (h *ManufacturerHandler) save(c *gin.Context, create bool) {
	item := models.Manufacturer{VendorType: models.VendorTypeManufacturer, IsActive: true}
	if !create {
		if err := h.db.First(&item, "id = ?", c.Param("id")).Error; err != nil {
			apiError(c, err)
			return
		}
	}
	if !bindFields(c, &item, map[string]string{"name": "required,max=255", "code": "required,max=50", "country": "max=100", "contactEmail": "omitempty,email,max=255", "contactPhone": "max=100", "website": "omitempty,http_url", "vendorType": "required,oneof=manufacturer supplier both", "isActive": ""}) {
		return
	}
	active := item.IsActive
	if err := h.db.Omit(clause.Associations).Save(&item).Error; err != nil {
		apiError(c, err)
		return
	}
	if err := h.db.Model(&item).Update("is_active", active).Error; err != nil {
		apiError(c, err)
		return
	}

	if create {
		response.Created(c, item)
	} else {
		h.Get(c)
	}
}
func (h *ManufacturerHandler) Delete(c *gin.Context) { deleteRecord(c, h.db, &models.Manufacturer{}) }
func (h *ManufacturerHandler) Categories(c *gin.Context) {
	var parent models.Manufacturer
	if err := h.db.Preload("Categories").First(&parent, "id = ?", c.Param("id")).Error; err != nil {
		apiError(c, err)
		return
	}
	if parent.Categories == nil {
		parent.Categories = []models.ManufacturerCategory{}
	}
	response.OK(c, parent.Categories)
}
func (h *ManufacturerHandler) SaveCategory(c *gin.Context) {
	var parent models.Manufacturer
	if err := h.db.First(&parent, "id = ?", c.Param("id")).Error; err != nil {
		apiError(c, err)
		return
	}
	item := models.ManufacturerCategory{ManufacturerID: parent.ID}
	create := c.Param("categoryId") == ""
	if !create {
		if err := h.db.First(&item, "id = ? AND manufacturer_id = ?", c.Param("categoryId"), parent.ID).Error; err != nil {
			apiError(c, err)
			return
		}
	}
	if !bindFields(c, &item, map[string]string{"name": "required,max=255", "description": "max=2000"}) {
		return
	}
	if err := h.db.Save(&item).Error; err != nil {
		apiError(c, err)
		return
	}
	if create {
		response.Created(c, item)
	} else {
		response.OK(c, item)
	}
}
func (h *ManufacturerHandler) DeleteCategory(c *gin.Context) {
	var item models.ManufacturerCategory
	if err := h.db.First(&item, "id = ? AND manufacturer_id = ?", c.Param("categoryId"), c.Param("id")).Error; err != nil {
		apiError(c, err)
		return
	}
	var count int64
	if err := h.db.Model(&models.Product{}).Where("category_id = ?", item.ID).Count(&count).Error; err != nil {
		apiError(c, err)
		return
	}
	if count > 0 {
		response.Conflict(c, "Category is used by products")
		return
	}
	if err := h.db.Delete(&item).Error; err != nil {
		apiError(c, err)
		return
	}
	response.NoContent(c)
}
