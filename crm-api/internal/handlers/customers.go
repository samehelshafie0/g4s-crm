package handlers

import (
	"g4s-crm/api/internal/middleware"
	"g4s-crm/api/internal/models"
	"g4s-crm/api/pkg/pagination"
	"g4s-crm/api/pkg/response"
	v "g4s-crm/api/pkg/validator"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
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
	params := pagination.GetParams(c)
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
	CompanyName string                 `json:"companyName" validate:"required"`
	Sector      models.Sector          `json:"sector" validate:"required"`
	Region      string                 `json:"region" validate:"required"`
	Status      models.CustomerStatus  `json:"status" validate:"required,oneof=active inactive prospect"`
	Type        models.CustomerType    `json:"type" validate:"required,oneof=get grow"`
	CRNumber    string                 `json:"crNumber"`
	VATNumber   string                 `json:"vatNumber"`
	Notes       string                 `json:"notes"`
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
		CRNumber:    req.CRNumber,
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

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.db.Model(&customer).Updates(req).Error; err != nil {
		response.InternalError(c, "Failed to update customer")
		return
	}

	response.OK(c, customer)
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

// AddSite adds a site to a customer
func (h *CustomerHandler) AddSite(c *gin.Context) {
	customerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid customer ID")
		return
	}

	var req struct {
		Name    string `json:"name" validate:"required"`
		Address string `json:"address"`
		City    string `json:"city"`
		Region  string `json:"region"`
	}
	if !v.BindAndValidate(c, &req) {
		return
	}

	site := &models.CustomerSite{
		CustomerID: customerID,
		Name:       req.Name,
		Address:    req.Address,
		City:       req.City,
		Region:     req.Region,
	}

	if err := h.db.Create(site).Error; err != nil {
		response.InternalError(c, "Failed to add site")
		return
	}
	response.Created(c, site)
}

// UpdateSite updates a specific site
func (h *CustomerHandler) UpdateSite(c *gin.Context) {
	siteID := c.Param("siteId")
	var site models.CustomerSite
	if err := h.db.First(&site, "id = ? AND customer_id = ?", siteID, c.Param("id")).Error; err != nil {
		response.NotFound(c, "Site not found")
		return
	}
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	h.db.Model(&site).Updates(req)
	response.OK(c, site)
}

// DeleteSite removes a site
func (h *CustomerHandler) DeleteSite(c *gin.Context) {
	siteID := c.Param("siteId")
	h.db.Delete(&models.CustomerSite{}, "id = ? AND customer_id = ?", siteID, c.Param("id"))
	response.NoContent(c)
}

// AddContact adds a contact to a customer
func (h *CustomerHandler) AddContact(c *gin.Context) {
	customerID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid customer ID")
		return
	}

	var req struct {
		Name      string `json:"name" validate:"required"`
		Email     string `json:"email"`
		Phone     string `json:"phone"`
		Position  string `json:"position"`
		IsPrimary bool   `json:"isPrimary"`
		SiteID    *uuid.UUID `json:"siteId"`
	}
	if !v.BindAndValidate(c, &req) {
		return
	}

	contact := &models.CustomerContact{
		CustomerID: customerID,
		SiteID:     req.SiteID,
		Name:       req.Name,
		Email:      req.Email,
		Phone:      req.Phone,
		Position:   req.Position,
		IsPrimary:  req.IsPrimary,
	}

	if err := h.db.Create(contact).Error; err != nil {
		response.InternalError(c, "Failed to add contact")
		return
	}
	response.Created(c, contact)
}

// UpdateContact updates a contact
func (h *CustomerHandler) UpdateContact(c *gin.Context) {
	contactID := c.Param("contactId")
	var contact models.CustomerContact
	if err := h.db.First(&contact, "id = ? AND customer_id = ?", contactID, c.Param("id")).Error; err != nil {
		response.NotFound(c, "Contact not found")
		return
	}
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	h.db.Model(&contact).Updates(req)
	response.OK(c, contact)
}

// DeleteContact removes a contact
func (h *CustomerHandler) DeleteContact(c *gin.Context) {
	contactID := c.Param("contactId")
	h.db.Delete(&models.CustomerContact{}, "id = ? AND customer_id = ?", contactID, c.Param("id"))
	response.NoContent(c)
}
