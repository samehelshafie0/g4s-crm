package handlers

import (
	"g4s-crm/api/internal/models"
	"g4s-crm/api/pkg/pagination"
	"g4s-crm/api/pkg/response"
	v "g4s-crm/api/pkg/validator"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProductHandler struct{ db *gorm.DB }

func NewProductHandler(db *gorm.DB) *ProductHandler { return &ProductHandler{db: db} }

func (h *ProductHandler) List(c *gin.Context) {
	params := pagination.GetParams(c)
	var items []models.Product
	var total int64
	query := h.db.Model(&models.Product{}).Preload("Manufacturer").Preload("Category")
	if q := c.Query("q"); q != "" {
		query = query.Where("name ILIKE ? OR sku ILIKE ?", "%"+q+"%", "%"+q+"%")
	}
	if mid := c.Query("manufacturerId"); mid != "" {
		query = query.Where("manufacturer_id = ?", mid)
	}
	if active := c.Query("isActive"); active != "" {
		query = query.Where("is_active = ?", active == "true")
	}
	query.Count(&total)
	query.Order(params.Sort + " " + params.Order).Scopes(pagination.Paginate(params)).Find(&items)
	response.OKWithMeta(c, items, pagination.BuildMeta(params, total))
}

func (h *ProductHandler) Get(c *gin.Context) {
	var item models.Product
	if err := h.db.Preload("Manufacturer").Preload("Category").
		Preload("VendorEntries").Preload("PriceHistory").Preload("Documents").
		First(&item, "id = ?", c.Param("id")).Error; err != nil {
		response.NotFound(c, "Product not found")
		return
	}
	response.OK(c, item)
}

func (h *ProductHandler) Create(c *gin.Context) {
	var req struct {
		SKU             string              `json:"sku" validate:"required"`
		Name            string              `json:"name" validate:"required"`
		Description     string              `json:"description"`
		ManufacturerID  *string             `json:"manufacturerId"`
		ProductType     models.ProductType  `json:"productType" validate:"required"`
		OriginCurrency  models.Currency     `json:"originCurrency"`
		UnitCostOrigin  float64             `json:"unitCostOrigin"`
		FXRate          float64             `json:"fxRate"`
		FreightPercent  float64             `json:"freightPercent"`
		CustomsPercent  float64             `json:"customsPercent"`
		SellingPrice    float64             `json:"sellingPrice"`
		LeadTimeDays    int                 `json:"leadTimeDays"`
	}
	if !v.BindAndValidate(c, &req) {
		return
	}

	item := &models.Product{
		SKU:            req.SKU,
		Name:           req.Name,
		Description:    req.Description,
		ProductType:    req.ProductType,
		OriginCurrency: req.OriginCurrency,
		UnitCostOrigin: req.UnitCostOrigin,
		FXRate:         req.FXRate,
		FreightPercent: req.FreightPercent,
		CustomsPercent: req.CustomsPercent,
		SellingPrice:   req.SellingPrice,
		LeadTimeDays:   req.LeadTimeDays,
		IsActive:       true,
	}
	item.RecalculateCosts()

	if err := h.db.Create(item).Error; err != nil {
		response.Conflict(c, "SKU already exists")
		return
	}
	response.Created(c, item)
}

func (h *ProductHandler) Update(c *gin.Context) {
	var item models.Product
	if err := h.db.First(&item, "id = ?", c.Param("id")).Error; err != nil {
		response.NotFound(c, "Product not found")
		return
	}
	var req map[string]interface{}
	c.ShouldBindJSON(&req)
	h.db.Model(&item).Updates(req)
	item.RecalculateCosts()
	h.db.Save(&item)
	response.OK(c, item)
}

func (h *ProductHandler) Delete(c *gin.Context) {
	h.db.Model(&models.Product{}).Where("id = ?", c.Param("id")).Update("is_active", false)
	response.NoContent(c)
}

func (h *ProductHandler) RecalculateCosts(c *gin.Context) {
	var products []models.Product
	h.db.Where("is_active = true").Find(&products)
	for i := range products {
		products[i].RecalculateCosts()
		h.db.Save(&products[i])
	}
	response.OK(c, gin.H{"updated": len(products)})
}
