package handlers

import (
	"g4s-crm/api/internal/middleware"
	"g4s-crm/api/internal/models"
	"g4s-crm/api/pkg/pagination"
	"g4s-crm/api/pkg/response"
	v "g4s-crm/api/pkg/validator"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
)

type CustomerHandler struct {
	db *gorm.DB
}

func NewCustomerHandler(db *gorm.DB) *CustomerHandler {
	return &CustomerHandler{db: db}
}

// List godoc
// @Summary      List customers
// @Tags         customers
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  response.Response
// @Router       /customers [get]
func (h *CustomerHandler) List(c *gin.Context) {
	params := pagination.GetParams(c, "company_name", "sector", "region", "status", "type")
	q := c.Query("q")
	sector := c.Query("sector")
	status := c.Query("status")
	ctype := c.Query("type")
	region := c.Query("region")

	var customers []models.Customer
	var total int64

	query := h.db.Model(&models.Customer{}).Preload("Sites").Preload("Contacts")

	if q != "" {
		query = query.Where("company_name ILIKE ? OR cr_number ILIKE ?", "%"+q+"%", "%"+q+"%")
	}
	if sector != "" {
		query = query.Where("sector = ?", sector)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if ctype != "" {
		query = query.Where("type = ?", ctype)
	}
	if region != "" {
		query = query.Where("region ILIKE ?", "%"+region+"%")
	}

	query.Count(&total)
	if err := query.Order(params.Sort + " " + params.Order).
		Scopes(pagination.Paginate(params)).
		Find(&customers).Error; err != nil {
		response.InternalError(c, "Failed to fetch customers")
		return
	}

	response.OKWithMeta(c, customers, pagination.BuildMeta(params, total))
}

// Get godoc
// @Summary      Get customer by ID
// @Tags         customers
// @Security     BearerAuth
// @Produce      json
// @Param        id  path  string  true  "Customer ID"
// @Success      200  {object}  response.Response
// @Router       /customers/{id} [get]
func (h *CustomerHandler) Get(c *gin.Context) {
	id := c.Param("id")
	var customer models.Customer
	if err := h.db.Preload("Sites.Contacts").Preload("Contacts").
		First(&customer, "id = ?", id).Error; err != nil {
		response.NotFound(c, "Customer not found")
		return
	}
	response.OK(c, customer)
}

type createCustomerRequest struct {
	CompanyName string                `json:"companyName" validate:"required"`
	Sector      models.Sector         `json:"sector" validate:"required,oneof=government healthcare education retail banking oil-gas telecom hospitality real-estate other"`
	Region      string                `json:"region" validate:"required"`
	Status      models.CustomerStatus `json:"status" validate:"required,oneof=active inactive prospect"`
	Type        models.CustomerType   `json:"type" validate:"required,oneof=get grow"`
	CRNumber    string                `json:"crNumber"`
	VATNumber   string                `json:"vatNumber"`
	Notes       string                `json:"notes"`
}

// Create godoc
// @Summary      Create a new customer
// @Tags         customers
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body createCustomerRequest true "Customer data"
// @Success      201  {object}  response.Response
// @Router       /customers [post]
func (h *CustomerHandler) Create(c *gin.Context) {
	var req createCustomerRequest
	if !v.BindAndValidate(c, &req) {
		return
	}

	userID := middleware.GetCurrentUserID(c)
	customer := &models.Customer{
		CompanyName: req.CompanyName,
		Sector:      req.Sector,
		Region:      req.Region,
		Status:      req.Status,
		Type:        req.Type,
		CRNumber:    strings.TrimSpace(req.CRNumber),
		VATNumber:   req.VATNumber,
		Notes:       req.Notes,
		CreatedByID: &userID,
	}

	if err := h.db.Create(customer).Error; err != nil {
		response.Conflict(c, "Failed to create customer. CR number may already exist.")
		return
	}

	response.Created(c, customer)
}

// Update godoc
// @Summary      Update customer
// @Tags         customers
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id  path  string  true  "Customer ID"
// @Success      200  {object}  response.Response
// @Router       /customers/{id} [patch]
func (h *CustomerHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var customer models.Customer
	if err := h.db.First(&customer, "id = ?", id).Error; err != nil {
		response.NotFound(c, "Customer not found")
		return
	}

	var req struct {
		CompanyName *string                `json:"companyName" validate:"omitempty,min=1,max=255"`
		Sector      *models.Sector         `json:"sector" validate:"omitempty,oneof=government healthcare education retail banking oil-gas telecom hospitality real-estate other"`
		Region      *string                `json:"region" validate:"omitempty,min=1,max=100"`
		Status      *models.CustomerStatus `json:"status" validate:"omitempty,oneof=active inactive prospect"`
		Type        *models.CustomerType   `json:"type" validate:"omitempty,oneof=get grow"`
		CRNumber    *string                `json:"crNumber" validate:"omitempty,max=50"`
		VATNumber   *string                `json:"vatNumber" validate:"omitempty,max=50"`
		Notes       *string                `json:"notes" validate:"omitempty,max=20000"`
	}
	if !v.BindStrict(c, &req) {
		return
	}
	// Empty CR means NULL; it must not participate in the unique identifier check.
	updates := map[string]interface{}{}
	if req.CompanyName != nil {
		updates["company_name"] = *req.CompanyName
	}
	if req.Sector != nil {
		updates["sector"] = *req.Sector
	}
	if req.Region != nil {
		updates["region"] = *req.Region
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Type != nil {
		updates["type"] = *req.Type
	}
	if req.CRNumber != nil {
		value := strings.TrimSpace(*req.CRNumber)
		if value == "" {
			updates["cr_number"] = nil
		} else {
			updates["cr_number"] = value
		}
	}
	if req.VATNumber != nil {
		updates["vat_number"] = *req.VATNumber
	}
	if req.Notes != nil {
		updates["notes"] = *req.Notes
	}
	if err := h.db.Model(&customer).Updates(updates).Error; err != nil {
		response.Conflict(c, "Customer could not be updated; registration number may already exist")
		return
	}
	h.Get(c)
}

// Delete godoc
// @Summary      Delete customer
// @Tags         customers
// @Security     BearerAuth
// @Param        id  path  string  true  "Customer ID"
// @Success      204
// @Router       /customers/{id} [delete]
func (h *CustomerHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.db.Delete(&models.Customer{}, "id = ?", id).Error; err != nil {
		response.NotFound(c, "Customer not found")
		return
	}
	response.NoContent(c)
}

func (h *CustomerHandler) parent(tx *gorm.DB, c *gin.Context) (models.Customer, error) {
	var customer models.Customer
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&customer, "id = ?", c.Param("id")).Error
	return customer, err
}
func (h *CustomerHandler) ListSites(c *gin.Context) {
	var customer models.Customer
	if err := h.db.Preload("Sites.Contacts").First(&customer, "id = ?", c.Param("id")).Error; err != nil {
		apiError(c, err)
		return
	}
	if customer.Sites == nil {
		customer.Sites = []models.CustomerSite{}
	}
	response.OK(c, customer.Sites)
}
func (h *CustomerHandler) ListContacts(c *gin.Context) {
	var customer models.Customer
	if err := h.db.Preload("Contacts").First(&customer, "id = ?", c.Param("id")).Error; err != nil {
		apiError(c, err)
		return
	}
	if customer.Contacts == nil {
		customer.Contacts = []models.CustomerContact{}
	}
	response.OK(c, customer.Contacts)
}
func (h *CustomerHandler) AddSite(c *gin.Context)    { h.saveSite(c, true) }
func (h *CustomerHandler) UpdateSite(c *gin.Context) { h.saveSite(c, false) }
func (h *CustomerHandler) saveSite(c *gin.Context, create bool) {
	var site models.CustomerSite
	if !create {
		if err := h.db.First(&site, "id = ? AND customer_id = ?", c.Param("siteId"), c.Param("id")).Error; err != nil {
			apiError(c, err)
			return
		}
	}
	if !bindFields(c, &site, map[string]string{"name": "required,max=255", "address": "max=2000", "city": "max=100", "region": "max=100"}) {
		return
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		customer, err := h.parent(tx, c)
		if err != nil {
			return err
		}
		site.CustomerID = customer.ID
		return tx.Omit(clause.Associations).Save(&site).Error
	})
	if err != nil {
		apiError(c, err)
		return
	}
	if create {
		response.Created(c, site)
	} else {
		response.OK(c, site)
	}
}
func (h *CustomerHandler) DeleteSite(c *gin.Context) {
	err := h.db.Transaction(func(tx *gorm.DB) error {
		customer, err := h.parent(tx, c)
		if err != nil {
			return err
		}
		var site models.CustomerSite
		if err := tx.First(&site, "id = ? AND customer_id = ?", c.Param("siteId"), customer.ID).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.CustomerContact{}).Where("site_id = ?", site.ID).Update("site_id", nil).Error; err != nil {
			return err
		}
		return tx.Delete(&site).Error
	})
	if err != nil {
		apiError(c, err)
		return
	}
	response.NoContent(c)
}
func (h *CustomerHandler) AddContact(c *gin.Context)    { h.saveContact(c, true) }
func (h *CustomerHandler) UpdateContact(c *gin.Context) { h.saveContact(c, false) }
func (h *CustomerHandler) saveContact(c *gin.Context, create bool) {
	var contact models.CustomerContact
	if !create {
		if err := h.db.First(&contact, "id = ? AND customer_id = ?", c.Param("contactId"), c.Param("id")).Error; err != nil {
			apiError(c, err)
			return
		}
	}
	if !bindFields(c, &contact, map[string]string{"name": "required,max=255", "email": "omitempty,email,max=255", "phone": "max=100", "position": "max=255", "isPrimary": "", "siteId": ""}) {
		return
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		customer, err := h.parent(tx, c)
		if err != nil {
			return err
		}
		contact.CustomerID = customer.ID
		if contact.SiteID != nil {
			var site models.CustomerSite
			if err := tx.First(&site, "id = ? AND customer_id = ?", contact.SiteID, customer.ID).Error; err != nil {
				return invalid("Site must belong to this customer")
			}
		}
		if contact.IsPrimary {
			if err := tx.Model(&models.CustomerContact{}).Where("customer_id = ? AND id <> ?", customer.ID, contact.ID).Update("is_primary", false).Error; err != nil {
				return err
			}
		}
		return tx.Omit(clause.Associations).Save(&contact).Error
	})
	if err != nil {
		apiError(c, err)
		return
	}
	if create {
		response.Created(c, contact)
	} else {
		response.OK(c, contact)
	}
}
func (h *CustomerHandler) DeleteContact(c *gin.Context) {
	result := h.db.Where("id = ? AND customer_id = ?", c.Param("contactId"), c.Param("id")).Delete(&models.CustomerContact{})
	if result.Error != nil {
		apiError(c, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		response.NotFound(c, "Contact not found")
		return
	}
	response.NoContent(c)
}
func (h *CustomerHandler) Lookup(c *gin.Context) {
	items := []models.Customer{}
	query := h.db.Select("id", "company_name", "status")
	if q := c.Query("q"); q != "" {
		query = query.Where("company_name ILIKE ?", "%"+q+"%")
	}
	if err := query.Order("company_name").Limit(100).Find(&items).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OK(c, items)
}
