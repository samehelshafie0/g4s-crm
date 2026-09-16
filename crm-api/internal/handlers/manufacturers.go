package handlers

import (
	"g4s-crm/api/internal/models"
	"g4s-crm/api/pkg/pagination"
	"g4s-crm/api/pkg/response"
	v "g4s-crm/api/pkg/validator"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ManufacturerHandler struct{ db *gorm.DB }

func NewManufacturerHandler(db *gorm.DB) *ManufacturerHandler { return &ManufacturerHandler{db: db} }

func (h *ManufacturerHandler) List(c *gin.Context) {
	params := pagination.GetParams(c, "name", "code")
	var items []models.Manufacturer
	var total int64
	query := h.db.Model(&models.Manufacturer{}).Preload("Categories")
	if q := c.Query("q"); q != "" {
		query = query.Where("name ILIKE ?", "%"+q+"%")
	}
	if vt := c.Query("vendorType"); vt != "" {
		query = query.Where("vendor_type = ?", vt)
	}
	query.Count(&total)
	query.Order(params.Sort + " " + params.Order).Scopes(pagination.Paginate(params)).Find(&items)
	response.OKWithMeta(c, items, pagination.BuildMeta(params, total))
}
func (h *ManufacturerHandler) Get(c *gin.Context) {
	var item models.Manufacturer
	if err := h.db.Preload("Categories").First(&item, "id = ?", c.Param("id")).Error; err != nil {
		response.NotFound(c, "Manufacturer not found")
		return
	}
	response.OK(c, item)
}
func (h *ManufacturerHandler) Create(c *gin.Context) {
	var req struct {
		Name       string            `json:"name" validate:"required"`
		Code       string            `json:"code" validate:"required"`
		Country    string            `json:"country"`
		VendorType models.VendorType `json:"vendorType"`
	}
	if !v.BindAndValidate(c, &req) {
		return
	}
	item := &models.Manufacturer{Name: req.Name, Code: req.Code, Country: req.Country, VendorType: req.VendorType, IsActive: true}
	if err := h.db.Create(item).Error; err != nil {
		response.Conflict(c, "Code may already exist")
		return
	}
	response.Created(c, item)
}
func (h *ManufacturerHandler) Update(c *gin.Context) {
	var item models.Manufacturer
	if err := h.db.First(&item, "id = ?", c.Param("id")).Error; err != nil {
		response.NotFound(c, "Not found")
		return
	}
	var req map[string]interface{}
	c.ShouldBindJSON(&req)
	h.db.Model(&item).Updates(req)
	response.OK(c, item)
}
func (h *ManufacturerHandler) Delete(c *gin.Context) {
	h.db.Delete(&models.Manufacturer{}, "id = ?", c.Param("id"))
	response.NoContent(c)
}
