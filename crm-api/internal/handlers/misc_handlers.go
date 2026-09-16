package handlers

import (
	"g4s-crm/api/internal/models"
	"g4s-crm/api/pkg/pagination"
	"g4s-crm/api/pkg/response"
	"g4s-crm/api/pkg/seqgen"
	v "g4s-crm/api/pkg/validator"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ─── Teams ─────────────────────────────────────────────────

type TeamHandler struct{ db *gorm.DB }

func NewTeamHandler(db *gorm.DB) *TeamHandler { return &TeamHandler{db: db} }

func (h *TeamHandler) List(c *gin.Context) {
	params := pagination.GetParams(c)
	var items []models.Team; var total int64
	h.db.Model(&models.Team{}).Count(&total)
	h.db.Preload("Members").Preload("Leader").Order(params.Sort+" "+params.Order).Scopes(pagination.Paginate(params)).Find(&items)
	response.OKWithMeta(c, items, pagination.BuildMeta(params, total))
}
func (h *TeamHandler) Get(c *gin.Context) {
	var item models.Team
	if err := h.db.Preload("Members").Preload("Leader").First(&item, "id = ?", c.Param("id")).Error; err != nil {
		response.NotFound(c, "Team not found"); return
	}
	response.OK(c, item)
}
func (h *TeamHandler) Create(c *gin.Context) {
	var req struct {
		Name       string            `json:"name" validate:"required"`
		Department models.Department `json:"department" validate:"required"`
		Description string           `json:"description"`
	}
	if !v.BindAndValidate(c, &req) { return }
	item := &models.Team{Name: req.Name, Department: req.Department, Description: req.Description, IsActive: true}
	h.db.Create(item); response.Created(c, item)
}
func (h *TeamHandler) Update(c *gin.Context) {
	var item models.Team
	if err := h.db.First(&item, "id = ?", c.Param("id")).Error; err != nil { response.NotFound(c, "Team not found"); return }
	var req map[string]interface{}; c.ShouldBindJSON(&req); h.db.Model(&item).Updates(req); response.OK(c, item)
}
func (h *TeamHandler) Delete(c *gin.Context) {
	h.db.Delete(&models.Team{}, "id = ?", c.Param("id")); response.NoContent(c)
}

// ─── Price Books ───────────────────────────────────────────

type PriceBookHandler struct{ db *gorm.DB }

func NewPriceBookHandler(db *gorm.DB) *PriceBookHandler { return &PriceBookHandler{db: db} }

func (h *PriceBookHandler) List(c *gin.Context) {
	params := pagination.GetParams(c)
	var items []models.PriceBook; var total int64
	h.db.Model(&models.PriceBook{}).Count(&total)
	h.db.Preload("Entries").Order(params.Sort+" "+params.Order).Scopes(pagination.Paginate(params)).Find(&items)
	response.OKWithMeta(c, items, pagination.BuildMeta(params, total))
}
func (h *PriceBookHandler) Get(c *gin.Context) {
	var item models.PriceBook
	if err := h.db.Preload("Entries.Product").First(&item, "id = ?", c.Param("id")).Error; err != nil {
		response.NotFound(c, "Price book not found"); return
	}
	response.OK(c, item)
}
func (h *PriceBookHandler) Create(c *gin.Context) {
	var req struct {
		Name        string               `json:"name" validate:"required"`
		Type        models.PriceBookType `json:"type" validate:"required"`
		Description string               `json:"description"`
	}
	if !v.BindAndValidate(c, &req) { return }
	item := &models.PriceBook{Name: req.Name, Type: req.Type, Description: req.Description, IsActive: true}
	h.db.Create(item); response.Created(c, item)
}
func (h *PriceBookHandler) Update(c *gin.Context) {
	var item models.PriceBook
	if err := h.db.First(&item, "id = ?", c.Param("id")).Error; err != nil { response.NotFound(c, "Not found"); return }
	var req map[string]interface{}; c.ShouldBindJSON(&req); h.db.Model(&item).Updates(req); response.OK(c, item)
}
func (h *PriceBookHandler) Delete(c *gin.Context) {
	h.db.Delete(&models.PriceBook{}, "id = ?", c.Param("id")); response.NoContent(c)
}

// ─── Exchange Rates ────────────────────────────────────────

type ExchangeRateHandler struct{ db *gorm.DB }

func NewExchangeRateHandler(db *gorm.DB) *ExchangeRateHandler { return &ExchangeRateHandler{db: db} }

func (h *ExchangeRateHandler) List(c *gin.Context) {
	var items []models.ExchangeRate
	h.db.Find(&items)
	response.OK(c, items)
}
func (h *ExchangeRateHandler) Get(c *gin.Context) {
	var item models.ExchangeRate
	if err := h.db.Preload("History").First(&item, "id = ?", c.Param("id")).Error; err != nil {
		response.NotFound(c, "Exchange rate not found"); return
	}
	response.OK(c, item)
}
func (h *ExchangeRateHandler) Update(c *gin.Context) {
	var item models.ExchangeRate
	if err := h.db.First(&item, "id = ?", c.Param("id")).Error; err != nil { response.NotFound(c, "Not found"); return }
	var req struct{ CurrentRate float64 `json:"currentRate" validate:"required,gt=0"`; EffectiveDate string `json:"effectiveDate"` }
	if !v.BindAndValidate(c, &req) { return }
	h.db.Create(&models.ExchangeRateHistory{ExchangeRateID: item.ID, Rate: item.CurrentRate, EffectiveDate: item.EffectiveDate})
	h.db.Model(&item).Update("current_rate", req.CurrentRate)
	response.OK(c, item)
}

// ─── Recurring Services ────────────────────────────────────

type RecurringServiceHandler struct{ db *gorm.DB }

func NewRecurringServiceHandler(db *gorm.DB) *RecurringServiceHandler { return &RecurringServiceHandler{db: db} }

func (h *RecurringServiceHandler) List(c *gin.Context) {
	params := pagination.GetParams(c)
	var items []models.RecurringService; var total int64
	h.db.Model(&models.RecurringService{}).Count(&total)
	h.db.Order(params.Sort+" "+params.Order).Scopes(pagination.Paginate(params)).Find(&items)
	response.OKWithMeta(c, items, pagination.BuildMeta(params, total))
}
func (h *RecurringServiceHandler) Get(c *gin.Context) {
	var item models.RecurringService
	if err := h.db.First(&item, "id = ?", c.Param("id")).Error; err != nil { response.NotFound(c, "Not found"); return }
	response.OK(c, item)
}
func (h *RecurringServiceHandler) Create(c *gin.Context) {
	var req struct {
		Name             string                      `json:"name" validate:"required"`
		ServiceType      models.RecurringServiceType `json:"serviceType"`
		Description      string                      `json:"description"`
		MonthlyCost      float64                     `json:"monthlyCost"`
		MonthlyPrice     float64                     `json:"monthlyPrice"`
		BillingFrequency models.BillingFrequency     `json:"billingFrequency"`
	}
	if !v.BindAndValidate(c, &req) { return }
	item := &models.RecurringService{
		Name: req.Name, ServiceType: req.ServiceType, Description: req.Description,
		MonthlyCost: req.MonthlyCost, MonthlyPrice: req.MonthlyPrice,
		AnnualCost: req.MonthlyCost * 12, AnnualPrice: req.MonthlyPrice * 12,
		BillingFrequency: req.BillingFrequency, IsActive: true,
	}
	h.db.Create(item); response.Created(c, item)
}
func (h *RecurringServiceHandler) Update(c *gin.Context) {
	var item models.RecurringService
	if err := h.db.First(&item, "id = ?", c.Param("id")).Error; err != nil { response.NotFound(c, "Not found"); return }
	var req map[string]interface{}; c.ShouldBindJSON(&req); h.db.Model(&item).Updates(req); response.OK(c, item)
}
func (h *RecurringServiceHandler) Delete(c *gin.Context) {
	h.db.Delete(&models.RecurringService{}, "id = ?", c.Param("id")); response.NoContent(c)
}

// ─── Projects ──────────────────────────────────────────────

type ProjectHandler struct{ db *gorm.DB }

func NewProjectHandler(db *gorm.DB) *ProjectHandler { return &ProjectHandler{db: db} }

func (h *ProjectHandler) List(c *gin.Context) {
	params := pagination.GetParams(c)
	var items []models.Project; var total int64
	query := h.db.Model(&models.Project{}).Preload("Customer").Preload("ProjectManager")
	if status := c.Query("status"); status != "" { query = query.Where("status = ?", status) }
	query.Count(&total)
	query.Order(params.Sort+" "+params.Order).Scopes(pagination.Paginate(params)).Find(&items)
	response.OKWithMeta(c, items, pagination.BuildMeta(params, total))
}
func (h *ProjectHandler) Get(c *gin.Context) {
	var item models.Project
	if err := h.db.Preload("Customer").Preload("Quote").Preload("ProjectManager").First(&item, "id = ?", c.Param("id")).Error; err != nil {
		response.NotFound(c, "Project not found"); return
	}
	response.OK(c, item)
}
func (h *ProjectHandler) Create(c *gin.Context) {
	var req struct {
		QuoteID          string  `json:"quoteId" validate:"required"`
		Name             string  `json:"name" validate:"required"`
		ProjectManagerID *string `json:"projectManagerId"`
		Priority         models.ProjectPriority `json:"priority"`
		Notes            string  `json:"notes"`
	}
	if !v.BindAndValidate(c, &req) { return }

	var quote models.Quote
	if err := h.db.First(&quote, "id = ?", req.QuoteID).Error; err != nil { response.NotFound(c, "Quote not found"); return }

	projectNumber, _ := seqgen.NextNumber(h.db, "project")
	priority := req.Priority
	if priority == "" { priority = models.PriorityMedium }

	item := &models.Project{
		ProjectNumber: projectNumber,
		Name:          req.Name,
		CustomerID:    quote.CustomerID,
		QuoteID:       &quote.ID,
		Status:        models.ProjectStatusPlanning,
		Priority:      priority,
		TotalValue:    quote.Total,
		TotalCost:     quote.TotalCost,
		MarginPercent: quote.MarginPercent,
		Notes:         req.Notes,
	}
	h.db.Create(item); response.Created(c, item)
}
func (h *ProjectHandler) Update(c *gin.Context) {
	var item models.Project
	if err := h.db.First(&item, "id = ?", c.Param("id")).Error; err != nil { response.NotFound(c, "Not found"); return }
	var req map[string]interface{}; c.ShouldBindJSON(&req); h.db.Model(&item).Updates(req); response.OK(c, item)
}
func (h *ProjectHandler) Delete(c *gin.Context) {
	h.db.Delete(&models.Project{}, "id = ?", c.Param("id")); response.NoContent(c)
}
