package handlers

import (
	"g4s-crm/api/internal/models"
	"g4s-crm/api/pkg/pagination"
	"g4s-crm/api/pkg/response"
	v "g4s-crm/api/pkg/validator"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OpportunityHandler struct{ db *gorm.DB }

func NewOpportunityHandler(db *gorm.DB) *OpportunityHandler { return &OpportunityHandler{db} }
func (h *OpportunityHandler) List(c *gin.Context) {
	params := pagination.GetParams(c, "title", "stage", "estimated_value", "expected_close_date")
	items := []models.Opportunity{}
	var total int64
	q := h.db.Model(&models.Opportunity{})
	if term := c.Query("q"); term != "" {
		q = q.Where("title ILIKE ?", "%"+term+"%")
	}
	for param, column := range map[string]string{"stage": "stage", "customerId": "customer_id", "salesExecutiveId": "sales_executive_id", "preSalesId": "pre_sales_id"} {
		if value := c.Query(param); value != "" {
			q = q.Where(column+" = ?", value)
		}
	}
	if err := q.Count(&total).Error; err != nil {
		apiError(c, err)
		return
	}
	if err := q.Preload("Customer").Preload("SalesExecutive").Preload("PreSales").Preload("Quotes").Order(params.Sort + " " + params.Order).Scopes(pagination.Paginate(params)).Find(&items).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OKWithMeta(c, items, pagination.BuildMeta(params, total))
}
func (h *OpportunityHandler) Get(c *gin.Context) {
	var item models.Opportunity
	if err := h.db.Preload("Customer").Preload("SalesExecutive").Preload("PreSales").Preload("Quotes").First(&item, "id = ?", c.Param("id")).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OK(c, item)
}

type opportunityInput struct {
	Title             string                  `json:"title"`
	CustomerID        uuid.UUID               `json:"customerId"`
	Stage             models.OpportunityStage `json:"stage"`
	ServiceTypes      pq.StringArray          `json:"serviceTypes"`
	EstimatedValue    float64                 `json:"estimatedValue"`
	EstimatedCost     float64                 `json:"estimatedCost"`
	WinProbability    int                     `json:"winProbability"`
	SalesExecutiveID  *uuid.UUID              `json:"salesExecutiveId"`
	PreSalesID        *uuid.UUID              `json:"preSalesId"`
	ExpectedCloseDate *string                 `json:"expectedCloseDate"`
	Notes             string                  `json:"notes"`
}

var opportunityRules = map[string]string{"title": "required,max=255", "customerId": "required", "stage": "required,oneof=qualification proposal negotiation closed-won closed-lost", "serviceTypes": "max=10,dive,oneof=cctv access-control intrusion-detection fire-alarm networking it-solutions guarding monitoring maintenance consulting", "estimatedValue": moneyRule, "estimatedCost": moneyRule, "winProbability": "gte=0,lte=100", "salesExecutiveId": "", "preSalesId": "", "expectedCloseDate": "", "notes": "max=20000"}

func (h *OpportunityHandler) Create(c *gin.Context) { h.save(c, true) }
func (h *OpportunityHandler) Update(c *gin.Context) { h.save(c, false) }
func (h *OpportunityHandler) save(c *gin.Context, create bool) {
	var item models.Opportunity
	var input opportunityInput
	if create {
		input.Stage = models.StageQualification
	} else {
		if err := h.db.First(&item, "id = ?", c.Param("id")).Error; err != nil {
			apiError(c, err)
			return
		}
		input = opportunityInput{Title: item.Title, CustomerID: item.CustomerID, Stage: item.Stage, ServiceTypes: item.ServiceTypes, EstimatedValue: item.EstimatedValue, EstimatedCost: item.EstimatedCost, WinProbability: item.WinProbability, SalesExecutiveID: item.SalesExecutiveID, PreSalesID: item.PreSalesID, Notes: item.Notes}
		if item.ExpectedCloseDate != nil {
			date := item.ExpectedCloseDate.Format("2006-01-02")
			input.ExpectedCloseDate = &date
		}
	}
	if !bindFields(c, &input, opportunityRules) {
		return
	}
	date, err := parseDate(input.ExpectedCloseDate)
	if err != nil {
		apiError(c, err)
		return
	}
	err = h.db.Transaction(func(tx *gorm.DB) error {
		if err := exists(tx, &models.Customer{}, input.CustomerID); err != nil {
			return err
		}
		for _, id := range []*uuid.UUID{input.SalesExecutiveID, input.PreSalesID} {
			if id != nil {
				var user models.User
				if err := tx.Where("is_active = true").First(&user, "id = ?", id).Error; err != nil {
					return invalid("Assigned user is unavailable")
				}
			}
		}
		if !create {
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&item, "id = ?", item.ID).Error; err != nil {
				return err
			}
		}
		oldStage := item.Stage
		item.Title = input.Title
		item.CustomerID = input.CustomerID
		item.Stage = input.Stage
		item.ServiceTypes = input.ServiceTypes
		item.EstimatedValue = input.EstimatedValue
		item.EstimatedCost = input.EstimatedCost
		item.WinProbability = input.WinProbability
		item.SalesExecutiveID = input.SalesExecutiveID
		item.PreSalesID = input.PreSalesID
		item.ExpectedCloseDate = date
		item.Notes = input.Notes
		item.EstimatedMargin = 0
		if item.EstimatedValue > 0 {
			item.EstimatedMargin = (item.EstimatedValue - item.EstimatedCost) / item.EstimatedValue * 100
		}
		if err := tx.Omit(clause.Associations).Save(&item).Error; err != nil {
			return err
		}
		action := "updated"
		if create {
			action = "created"
		} else if oldStage != item.Stage {
			action = "stage: " + string(oldStage) + " → " + string(item.Stage)
		}
		return recordActivity(tx, c, "opportunity", item.ID, action)
	})
	if err != nil {
		apiError(c, err)
		return
	}
	h.db.Preload("Customer").Preload("SalesExecutive").Preload("PreSales").Preload("Quotes").First(&item, "id = ?", item.ID)
	if create {
		response.Created(c, item)
	} else {
		response.OK(c, item)
	}
}
func (h *OpportunityHandler) Delete(c *gin.Context) { deleteRecord(c, h.db, &models.Opportunity{}) }
func (h *OpportunityHandler) UpdateStage(c *gin.Context) {
	var req struct {
		Stage models.OpportunityStage `json:"stage" validate:"required,oneof=qualification proposal negotiation closed-won closed-lost"`
		Notes string                  `json:"notes" validate:"max=20000"`
	}
	if !v.BindStrict(c, &req) {
		return
	}
	var item models.Opportunity
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&item, "id = ?", c.Param("id")).Error; err != nil {
			return err
		}
		previous := item.Stage
		item.Stage = req.Stage
		if req.Notes != "" {
			item.Notes += "\n" + req.Notes
		}
		if err := tx.Omit(clause.Associations).Save(&item).Error; err != nil {
			return err
		}
		return recordActivity(tx, c, "opportunity", item.ID, "stage: "+string(previous)+" → "+string(item.Stage))
	})
	if err != nil {
		apiError(c, err)
		return
	}
	h.Get(c)
}
func (h *OpportunityHandler) Pipeline(c *gin.Context) {
	type stageCount struct {
		Stage models.OpportunityStage `json:"stage"`
		Count int64                   `json:"count"`
		Value float64                 `json:"value"`
	}
	rows := []stageCount{}
	if err := h.db.Model(&models.Opportunity{}).Select("stage, count(*) as count, COALESCE(sum(estimated_value),0) as value").Group("stage").Find(&rows).Error; err != nil {
		apiError(c, err)
		return
	}
	result := []stageCount{}
	for _, stage := range []models.OpportunityStage{models.StageQualification, models.StageProposal, models.StageNegotiation, models.StageClosedWon, models.StageClosedLost} {
		row := stageCount{Stage: stage}
		for _, found := range rows {
			if found.Stage == stage {
				row = found
			}
		}
		result = append(result, row)
	}
	response.OK(c, result)
}
