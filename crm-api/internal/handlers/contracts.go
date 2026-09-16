package handlers

import (
	"time"

	"g4s-crm/api/internal/models"
	"g4s-crm/api/pkg/pagination"
	"g4s-crm/api/pkg/response"
	"g4s-crm/api/pkg/seqgen"
	v "g4s-crm/api/pkg/validator"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ContractHandler struct{ db *gorm.DB }

func NewContractHandler(db *gorm.DB) *ContractHandler { return &ContractHandler{db: db} }

func (h *ContractHandler) List(c *gin.Context) {
	params := pagination.GetParams(c)
	var items []models.Contract
	var total int64
	query := h.db.Model(&models.Contract{}).Preload("Customer")
	if status := c.Query("status"); status != "" { query = query.Where("status = ?", status) }
	if cid := c.Query("customerId"); cid != "" { query = query.Where("customer_id = ?", cid) }
	query.Count(&total)
	query.Order(params.Sort + " " + params.Order).Scopes(pagination.Paginate(params)).Find(&items)
	response.OKWithMeta(c, items, pagination.BuildMeta(params, total))
}

func (h *ContractHandler) Get(c *gin.Context) {
	var item models.Contract
	if err := h.db.Preload("Customer").First(&item, "id = ?", c.Param("id")).Error; err != nil {
		response.NotFound(c, "Contract not found"); return
	}
	response.OK(c, item)
}

func (h *ContractHandler) Create(c *gin.Context) {
	var req struct {
		Title      string              `json:"title" validate:"required"`
		CustomerID string              `json:"customerId" validate:"required"`
		Type       models.ContractType `json:"type" validate:"required"`
		Value      float64             `json:"value"`
		StartDate  *string             `json:"startDate"`
		EndDate    *string             `json:"endDate"`
		AutoRenew  bool                `json:"autoRenew"`
		Terms      string              `json:"terms"`
		Notes      string              `json:"notes"`
	}
	if !v.BindAndValidate(c, &req) { return }

	contractNumber, _ := seqgen.NextNumber(h.db, "contract")
	item := &models.Contract{
		ContractNumber: contractNumber,
		Title:          req.Title,
		Type:           req.Type,
		Status:         models.ContractStatusDraft,
		Value:          req.Value,
		AutoRenew:      req.AutoRenew,
		Terms:          req.Terms,
		Notes:          req.Notes,
	}
	if err := h.db.Create(item).Error; err != nil {
		response.InternalError(c, "Failed to create contract"); return
	}
	response.Created(c, item)
}

func (h *ContractHandler) Update(c *gin.Context) {
	var item models.Contract
	if err := h.db.First(&item, "id = ?", c.Param("id")).Error; err != nil {
		response.NotFound(c, "Contract not found"); return
	}
	var req map[string]interface{}
	c.ShouldBindJSON(&req)
	h.db.Model(&item).Updates(req)
	response.OK(c, item)
}

func (h *ContractHandler) Delete(c *gin.Context) {
	h.db.Delete(&models.Contract{}, "id = ?", c.Param("id"))
	response.NoContent(c)
}

func (h *ContractHandler) Activate(c *gin.Context) {
	h.db.Model(&models.Contract{}).Where("id = ?", c.Param("id")).Update("status", models.ContractStatusActive)
	response.OK(c, gin.H{"status": "active"})
}

func (h *ContractHandler) Terminate(c *gin.Context) {
	h.db.Model(&models.Contract{}).Where("id = ?", c.Param("id")).Update("status", models.ContractStatusTerminated)
	response.OK(c, gin.H{"status": "terminated"})
}

func (h *ContractHandler) Renew(c *gin.Context) {
	var original models.Contract
	if err := h.db.First(&original, "id = ?", c.Param("id")).Error; err != nil {
		response.NotFound(c, "Contract not found"); return
	}

	h.db.Model(&original).Update("status", models.ContractStatusRenewed)

	contractNumber, _ := seqgen.NextNumber(h.db, "contract")
	var newStart *time.Time
	if original.EndDate != nil {
		t := original.EndDate.AddDate(0, 0, 1)
		newStart = &t
	}

	renewed := &models.Contract{
		ContractNumber:    contractNumber,
		Title:             original.Title + " (Renewal)",
		CustomerID:        original.CustomerID,
		Type:              original.Type,
		Status:            models.ContractStatusDraft,
		StartDate:         newStart,
		Value:             original.Value,
		AutoRenew:         original.AutoRenew,
		RenewalNoticeDays: original.RenewalNoticeDays,
		RenewedFromID:     &original.ID,
		Terms:             original.Terms,
	}
	h.db.Create(renewed)
	response.Created(c, renewed)
}

func (h *ContractHandler) Expiring(c *gin.Context) {
	days := 30
	var items []models.Contract
	h.db.Preload("Customer").
		Where("status = 'active' AND end_date <= NOW() + INTERVAL '? days'", days).
		Find(&items)
	response.OK(c, items)
}
