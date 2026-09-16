package handlers

import (
	"time"

	"g4s-crm/api/internal/models"
	"g4s-crm/api/pkg/pagination"
	"g4s-crm/api/pkg/response"
	"g4s-crm/api/pkg/seqgen"
	v "g4s-crm/api/pkg/validator"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ContractHandler struct{ db *gorm.DB }

func NewContractHandler(db *gorm.DB) *ContractHandler { return &ContractHandler{db: db} }

func (h *ContractHandler) List(c *gin.Context) {
	params := pagination.GetParams(c, "contract_number", "title", "status", "value", "start_date", "end_date")
	var items []models.Contract
	var total int64
	query := h.db.Model(&models.Contract{}).Preload("Customer")
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if cid := c.Query("customerId"); cid != "" {
		query = query.Where("customer_id = ?", cid)
	}
	query.Count(&total)
	query.Order(params.Sort + " " + params.Order).Scopes(pagination.Paginate(params)).Find(&items)
	response.OKWithMeta(c, items, pagination.BuildMeta(params, total))
}

func (h *ContractHandler) Get(c *gin.Context) {
	var item models.Contract
	if err := h.db.Preload("Customer").First(&item, "id = ?", c.Param("id")).Error; err != nil {
		response.NotFound(c, "Contract not found")
		return
	}
	response.OK(c, item)
}

func (h *ContractHandler) Create(c *gin.Context) {
	var req struct {
		Title      string              `json:"title" validate:"required"`
		CustomerID uuid.UUID           `json:"customerId" validate:"required"`
		Type       models.ContractType `json:"type" validate:"required,oneof=sales maintenance service project subscription"`
		Value      float64             `json:"value" validate:"gte=0"`
		StartDate  *string             `json:"startDate" validate:"omitempty,datetime=2006-01-02"`
		EndDate    *string             `json:"endDate" validate:"omitempty,datetime=2006-01-02"`
		AutoRenew  bool                `json:"autoRenew"`
		Terms      string              `json:"terms"`
		Notes      string              `json:"notes"`
	}
	if !v.BindAndValidate(c, &req) {
		return
	}

	contractNumber, err := seqgen.NextNumber(h.db, "contract")
	if err != nil {
		response.InternalError(c, "Failed to generate contract number")
		return
	}
	item := &models.Contract{
		ContractNumber: contractNumber,
		CustomerID:     req.CustomerID,
		Title:          req.Title,
		Type:           req.Type,
		Status:         models.ContractStatusDraft,
		Value:          req.Value,
		AutoRenew:      req.AutoRenew,
		Terms:          req.Terms,
		Notes:          req.Notes,
	}
	if req.StartDate != nil {
		date, _ := time.Parse("2006-01-02", *req.StartDate)
		item.StartDate = &date
	}
	if req.EndDate != nil {
		date, _ := time.Parse("2006-01-02", *req.EndDate)
		item.EndDate = &date
	}
	if item.StartDate != nil && item.EndDate != nil && !item.EndDate.After(*item.StartDate) {
		response.BadRequest(c, "End date must be after start date")
		return
	}
	var customer models.Customer
	if err := h.db.First(&customer, "id = ?", req.CustomerID).Error; err != nil {
		response.BadRequest(c, "Customer does not exist")
		return
	}
	if err := h.db.Create(item).Error; err != nil {
		response.InternalError(c, "Failed to create contract")
		return
	}
	response.Created(c, item)
}

func (h *ContractHandler) Update(c *gin.Context) {
	var item models.Contract
	if err := h.db.First(&item, "id = ?", c.Param("id")).Error; err != nil {
		response.NotFound(c, "Contract not found")
		return
	}
	var req struct {
		Title     *string `json:"title" validate:"omitempty,min=1,max=255"`
		Terms     *string `json:"terms" validate:"omitempty,max=20000"`
		Notes     *string `json:"notes" validate:"omitempty,max=20000"`
		AutoRenew *bool   `json:"autoRenew"`
	}
	if !v.BindStrict(c, &req) {
		return
	}
	updates := map[string]interface{}{}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Terms != nil {
		updates["terms"] = *req.Terms
	}
	if req.Notes != nil {
		updates["notes"] = *req.Notes
	}
	if req.AutoRenew != nil {
		updates["auto_renew"] = *req.AutoRenew
	}
	if item.Status != models.ContractStatusDraft {
		response.UnprocessableEntity(c, "Only draft contracts can be edited")
		return
	}
	if len(updates) > 0 {
		result := h.db.Model(&item).Where("status = ?", models.ContractStatusDraft).Updates(updates)
		if result.Error != nil {
			response.InternalError(c, "Failed to update contract")
			return
		}
		if result.RowsAffected != 1 {
			response.Conflict(c, "Contract changed; reload before editing")
			return
		}
	}
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
		response.NotFound(c, "Contract not found")
		return
	}

	h.db.Model(&original).Update("status", models.ContractStatusRenewed)

	contractNumber, err := seqgen.NextNumber(h.db, "contract")
	if err != nil {
		response.InternalError(c, "Failed to generate contract number")
		return
	}
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
		Where("status = ? AND end_date >= ? AND end_date <= ?", models.ContractStatusActive, time.Now(), time.Now().AddDate(0, 0, days)).
		Find(&items)
	response.OK(c, items)
}
