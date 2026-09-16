package handlers

import (
	"g4s-crm/api/internal/models"
	"g4s-crm/api/pkg/pagination"
	"g4s-crm/api/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type ProductHandler struct{ db *gorm.DB }

func NewProductHandler(db *gorm.DB) *ProductHandler { return &ProductHandler{db} }
func productQuery(db *gorm.DB) *gorm.DB {
	return db.Preload("Manufacturer").Preload("Category").Preload("VendorEntries").Preload("PriceHistory", func(tx *gorm.DB) *gorm.DB { return tx.Order("date DESC") }).Preload("Documents")
}
func (h *ProductHandler) List(c *gin.Context) {
	params := pagination.GetParams(c, "sku", "name", "selling_price", "landed_cost_sar")
	items := []models.Product{}
	var total int64
	q := h.db.Model(&models.Product{})
	if term := c.Query("q"); term != "" {
		q = q.Where("name ILIKE ? OR sku ILIKE ?", "%"+term+"%", "%"+term+"%")
	}
	for param, column := range map[string]string{"manufacturerId": "manufacturer_id", "categoryId": "category_id", "productType": "product_type", "isActive": "is_active"} {
		if value := c.Query(param); value != "" {
			q = q.Where(column+" = ?", value)
		}
	}
	if err := q.Count(&total).Error; err != nil {
		apiError(c, err)
		return
	}
	if err := productQuery(q).Order(params.Sort + " " + params.Order).Scopes(pagination.Paginate(params)).Find(&items).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OKWithMeta(c, items, pagination.BuildMeta(params, total))
}
func (h *ProductHandler) Get(c *gin.Context) {
	var item models.Product
	if err := productQuery(h.db).First(&item, "id = ?", c.Param("id")).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OK(c, item)
}

var productRules = map[string]string{"sku": "required,max=100", "name": "required,max=255", "description": "max=20000", "manufacturerId": "", "categoryId": "", "productType": "required,oneof=import local", "originCurrency": currencyRule, "unitCostOrigin": moneyRule, "fxRate": "gt=0,lte=1000000", "freightPercent": percentRule, "customsPercent": percentRule, "clearancePercent": percentRule, "targetMarginPercent": "gte=0,lt=100", "sellingPrice": moneyRule, "leadTimeDays": "gte=0,lte=3650", "supplierName": "max=255", "isActive": ""}

func validateProduct(tx *gorm.DB, p *models.Product) error {
	if err := optionalExists(tx, &models.Manufacturer{}, p.ManufacturerID); err != nil {
		return err
	}
	if p.CategoryID != nil {
		var category models.ManufacturerCategory
		if err := tx.First(&category, "id = ?", p.CategoryID).Error; err != nil {
			return invalid("Category does not exist")
		}
		if p.ManufacturerID == nil || category.ManufacturerID != *p.ManufacturerID {
			return invalid("Category must belong to the manufacturer")
		}
	}
	p.RecalculateCosts()
	return nil
}
func (h *ProductHandler) Create(c *gin.Context) { h.save(c, true) }
func (h *ProductHandler) Update(c *gin.Context) { h.save(c, false) }
func (h *ProductHandler) save(c *gin.Context, create bool) {
	item := models.Product{ProductType: models.ProductTypeLocal, OriginCurrency: models.CurrencySAR, FXRate: 1, IsActive: true}
	if !create {
		if err := h.db.First(&item, "id = ?", c.Param("id")).Error; err != nil {
			apiError(c, err)
			return
		}
	}
	oldCost := item.UnitCostOrigin
	if !bindFields(c, &item, productRules) {
		return
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := validateProduct(tx, &item); err != nil {
			return err
		}
		active := item.IsActive
		if err := tx.Omit(clause.Associations).Save(&item).Error; err != nil {
			return err
		}
		if err := tx.Model(&item).Update("is_active", active).Error; err != nil {
			return err
		}

		if create || oldCost != item.UnitCostOrigin {
			if err := tx.Create(&models.ProductPriceRecord{ProductID: item.ID, Date: time.Now(), Source: models.PriceSourceManual, UnitCost: item.UnitCostOrigin, Currency: item.OriginCurrency, LandingCost: item.LandedCostSAR, Qty: 1, VendorName: item.SupplierName}).Error; err != nil {
				return err
			}
		}
		return recordActivity(tx, c, "product", item.ID, "saved")
	})
	if err != nil {
		apiError(c, err)
		return
	}
	if err := productQuery(h.db).First(&item, "id = ?", item.ID).Error; err != nil {
		apiError(c, err)
		return
	}
	if create {
		response.Created(c, item)
	} else {
		response.OK(c, item)
	}
}
func (h *ProductHandler) Delete(c *gin.Context) { deleteRecord(c, h.db, &models.Product{}) }
func (h *ProductHandler) RecalculateCosts(c *gin.Context) {
	count := 0
	err := h.db.Transaction(func(tx *gorm.DB) error {
		items := []models.Product{}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Find(&items).Error; err != nil {
			return err
		}
		for i := range items {
			items[i].RecalculateCosts()
			if err := tx.Omit(clause.Associations).Save(&items[i]).Error; err != nil {
				return err
			}
			count++
		}
		return nil
	})
	if err != nil {
		apiError(c, err)
		return
	}
	response.OK(c, gin.H{"updated": count})
}
func (h *ProductHandler) Vendors(c *gin.Context) {
	items := []models.ProductVendorEntry{}
	if err := h.db.Where("product_id = ?", c.Param("id")).Find(&items).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OK(c, items)
}
func (h *ProductHandler) SaveVendor(c *gin.Context) {
	var product models.Product
	if err := h.db.First(&product, "id = ?", c.Param("id")).Error; err != nil {
		apiError(c, err)
		return
	}
	item := models.ProductVendorEntry{ProductID: product.ID, MOQ: 1, Currency: models.CurrencySAR}
	create := c.Param("vendorId") == ""
	if !create {
		if err := h.db.First(&item, "id = ? AND product_id = ?", c.Param("vendorId"), product.ID).Error; err != nil {
			apiError(c, err)
			return
		}
	}
	if !bindFields(c, &item, map[string]string{"vendorName": "required,max=255", "vendorSku": "max=100", "unitCost": moneyRule, "currency": currencyRule, "moq": "gte=1", "leadTimeDays": "gte=0", "catalogSource": "max=255", "notes": "max=20000"}) {
		return
	}
	now := time.Now()
	item.LastQuoteDate = &now
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
func (h *ProductHandler) DeleteVendor(c *gin.Context) {
	result := h.db.Where("id = ? AND product_id = ?", c.Param("vendorId"), c.Param("id")).Delete(&models.ProductVendorEntry{})
	if result.Error != nil {
		apiError(c, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		response.NotFound(c, "Vendor entry not found")
		return
	}
	response.NoContent(c)
}
func (h *ProductHandler) PriceHistory(c *gin.Context) {
	items := []models.ProductPriceRecord{}
	if err := h.db.Where("product_id = ?", c.Param("id")).Order("date DESC").Find(&items).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OK(c, items)
}
func (h *ProductHandler) AddPrice(c *gin.Context) {
	var product models.Product
	if err := h.db.First(&product, "id = ?", c.Param("id")).Error; err != nil {
		apiError(c, err)
		return
	}
	item := models.ProductPriceRecord{ProductID: product.ID, Date: time.Now(), Source: models.PriceSourceManual, Currency: models.CurrencySAR, Qty: 1}
	if !bindFields(c, &item, map[string]string{"date": "required", "sourceRef": "max=255", "vendorName": "max=255", "unitCost": moneyRule, "currency": currencyRule, "landingCost": moneyRule, "qty": "gte=1", "notes": "max=20000"}) {
		return
	}
	if err := h.db.Create(&item).Error; err != nil {
		apiError(c, err)
		return
	}
	response.Created(c, item)
}

func (h *ProductHandler) Documents(c *gin.Context) {
	items := []models.ProductDocument{}
	if err := h.db.Preload("UploadedBy").Where("product_id = ?", c.Param("id")).Find(&items).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OK(c, items)
}
func (h *ProductHandler) AddDocument(c *gin.Context) {
	var item models.ProductDocument
	if !bindFields(c, &item, map[string]string{"documentId": "required", "name": "required,max=255", "docType": "required,oneof=datasheet manual certificate vendor-quote catalog image other", "notes": "max=5000"}) {
		return
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		var p models.Product
		if err := tx.First(&p, "id = ?", c.Param("id")).Error; err != nil {
			return err
		}
		var doc models.Document
		if err := tx.First(&doc, "id = ?", item.DocumentID).Error; err != nil {
			return err
		}
		item.ProductID = p.ID
		item.FileName = doc.FileName
		item.FileSize = doc.FileSize
		item.UploadedByID = doc.UploadedByID
		return tx.Create(&item).Error
	})
	if err != nil {
		apiError(c, err)
		return
	}
	response.Created(c, item)
}
func (h *ProductHandler) DeleteDocument(c *gin.Context) {
	result := h.db.Where("product_id = ?", c.Param("id")).Delete(&models.ProductDocument{}, "id = ?", c.Param("documentId"))
	if result.Error != nil {
		apiError(c, result.Error)
		return
	}
	if result.RowsAffected != 1 {
		response.NotFound(c, "Document not found")
		return
	}
	response.NoContent(c)
}
