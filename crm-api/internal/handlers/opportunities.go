package handlers

import (
	"g4s-crm/api/internal/models"
	"g4s-crm/api/pkg/pagination"
	"g4s-crm/api/pkg/response"
	v "g4s-crm/api/pkg/validator"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type OpportunityHandler struct{ db *gorm.DB }

func NewOpportunityHandler(db *gorm.DB) *OpportunityHandler {
	return &OpportunityHandler{db: db}
}

func (h *OpportunityHandler) List(c *gin.Context) {
	params := pagination.GetParams(c)
	var opps []models.Opportunity
	var total int64

	query := h.db.Model(&models.Opportunity{}).Preload("Customer").Preload("SalesExecutive").Preload("PreSales")

	if q := c.Query("q"); q != "" {
		query = query.Where("title ILIKE ?", "%"+q+"%")
	}
	if stage := c.Query("stage"); stage != "" {
		query = query.Where("stage = ?", stage)
	}
	if cid := c.Query("customerId"); cid != "" {
		query = query.Where("customer_id = ?", cid)
	}
	if sid := c.Query("salesExecutiveId"); sid != "" {
		query = query.Where("sales_executive_id = ?", sid)
	}

	query.Count(&total)
	query.Order(params.Sort + " " + params.Order).Scopes(pagination.Paginate(params)).Find(&opps)
	response.OKWithMeta(c, opps, pagination.BuildMeta(params, total))
}

func (h *OpportunityHandler) Get(c *gin.Context) {
	var opp models.Opportunity
	if err := h.db.Preload("Customer").Preload("Quotes").First(&opp, "id = ?", c.Param("id")).Error; err != nil {
		response.NotFound(c, "Opportunity not found")
		return
	}
	response.OK(c, opp)
}

func (h *OpportunityHandler) Create(c *gin.Context) {
	var req struct {
		Title             string                  `json:"title" validate:"required"`
		CustomerID        string                  `json:"customerId" validate:"required"`
		Stage             models.OpportunityStage `json:"stage"`
		EstimatedValue    float64                 `json:"estimatedValue"`
		WinProbability    int                     `json:"winProbability"`
		ExpectedCloseDate *string                 `json:"expectedCloseDate"`
		Notes             string                  `json:"notes"`
	}
	if !v.BindAndValidate(c, &req) {
		return
	}

	opp := &models.Opportunity{
		Title:          req.Title,
		EstimatedValue: req.EstimatedValue,
		WinProbability: req.WinProbability,
		Notes:          req.Notes,
	}
	if req.Stage == "" {
		opp.Stage = models.StageQualification
	} else {
		opp.Stage = req.Stage
	}

	if err := h.db.Create(opp).Error; err != nil {
		response.InternalError(c, "Failed to create opportunity")
		return
	}
	response.Created(c, opp)
}

func (h *OpportunityHandler) Update(c *gin.Context) {
	var opp models.Opportunity
	if err := h.db.First(&opp, "id = ?", c.Param("id")).Error; err != nil {
		response.NotFound(c, "Opportunity not found")
		return
	}
	var req map[string]interface{}
	c.ShouldBindJSON(&req)
	h.db.Model(&opp).Updates(req)
	response.OK(c, opp)
}

func (h *OpportunityHandler) Delete(c *gin.Context) {
	h.db.Delete(&models.Opportunity{}, "id = ?", c.Param("id"))
	response.NoContent(c)
}

func (h *OpportunityHandler) UpdateStage(c *gin.Context) {
	var req struct {
		Stage models.OpportunityStage `json:"stage" validate:"required"`
		Notes string                  `json:"notes"`
	}
	if !v.BindAndValidate(c, &req) {
		return
	}

	var opp models.Opportunity
	if err := h.db.First(&opp, "id = ?", c.Param("id")).Error; err != nil {
		response.NotFound(c, "Opportunity not found")
		return
	}

	h.db.Model(&opp).Updates(map[string]interface{}{"stage": req.Stage})
	if req.Notes != "" {
		h.db.Model(&opp).Update("notes", opp.Notes+"\n[Stage: "+string(req.Stage)+"]: "+req.Notes)
	}
	response.OK(c, opp)
}

func (h *OpportunityHandler) Pipeline(c *gin.Context) {
	type StageCount struct {
		Stage models.OpportunityStage `json:"stage"`
		Count int64                   `json:"count"`
		Value float64                 `json:"value"`
	}

	stages := []models.OpportunityStage{
		models.StageQualification, models.StageProposal,
		models.StageNegotiation, models.StageClosedWon, models.StageClosedLost,
	}

	var results []StageCount
	h.db.Model(&models.Opportunity{}).
		Select("stage, COUNT(*) as count, COALESCE(SUM(estimated_value), 0) as value").
		Group("stage").
		Find(&results)

	stageMap := make(map[models.OpportunityStage]StageCount)
	for _, r := range results {
		stageMap[r.Stage] = r
	}

	var pipeline []StageCount
	for _, s := range stages {
		if v, ok := stageMap[s]; ok {
			pipeline = append(pipeline, v)
		} else {
			pipeline = append(pipeline, StageCount{Stage: s})
		}
	}

	response.OK(c, pipeline)
}
