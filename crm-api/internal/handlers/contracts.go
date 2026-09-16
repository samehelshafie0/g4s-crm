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
	"gorm.io/gorm/clause"
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
		RenewalNoticeDays *int                `json:"renewalNoticeDays" validate:"omitempty,gte=0,lte=3650"`
		Currency          models.Currency     `json:"currency" validate:"omitempty,oneof=SAR USD EUR GBP AED CNY"`
		Title             string              `json:"title" validate:"required"`
		CustomerID        uuid.UUID           `json:"customerId" validate:"required"`
		Type              models.ContractType `json:"type" validate:"required,oneof=sales maintenance service project subscription"`
		Value             float64             `json:"value" validate:"gte=0"`
		StartDate         *string             `json:"startDate" validate:"omitempty,datetime=2006-01-02"`
		EndDate           *string             `json:"endDate" validate:"omitempty,datetime=2006-01-02"`
		AutoRenew         bool                `json:"autoRenew"`
		Terms             string              `json:"terms"`
		Notes             string              `json:"notes"`
	}
	if !v.BindStrict(c, &req) {
		return
	}

	contractNumber, err := seqgen.NextNumber(h.db, "contract")
	if err != nil {
		response.InternalError(c, "Failed to generate contract number")
		return
	}
	item := &models.Contract{
		ContractNumber: contractNumber,
		Currency:       req.Currency,
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
	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(item).Error; err != nil {
			return err
		}
		if req.RenewalNoticeDays != nil {
			item.RenewalNoticeDays = *req.RenewalNoticeDays
			if err := tx.Model(item).Update("renewal_notice_days", item.RenewalNoticeDays).Error; err != nil {
				return err
			}
		}
		return recordActivity(tx, c, "contract", item.ID, "created")
	}); err != nil {
		response.InternalError(c, "Failed to create contract")
		return
	}
	response.Created(c, item)
}

func (h *ContractHandler) Update(c *gin.Context) {
	var item models.Contract
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&item, "id = ?", c.Param("id")).Error; err != nil {
			return err
		}
		if item.Status != models.ContractStatusDraft {
			return invalid("Only draft contracts can be edited")
		}
		rules := map[string]string{"title": "required,max=255", "terms": "max=20000", "notes": "max=20000", "autoRenew": "", "renewalNoticeDays": "gte=0,lte=3650", "startDate": "", "endDate": ""}
		if item.QuoteID == nil {
			rules["value"] = moneyRule
			rules["type"] = "required,oneof=sales maintenance service project subscription"
		}
		if !bindFields(c, &item, rules) {
			return invalid("Invalid contract")
		}
		if item.StartDate != nil && item.EndDate != nil && !item.EndDate.After(*item.StartDate) {
			return invalid("End date must follow start date")
		}
		if err := tx.Omit(clause.Associations).Save(&item).Error; err != nil {
			return err
		}
		return recordActivity(tx, c, "contract", item.ID, "updated")
	})
	if err != nil {
		if !c.Writer.Written() {
			apiError(c, err)
		}
		return
	}
	h.Get(c)
}
func (h *ContractHandler) Delete(c *gin.Context) {
	result := h.db.Where("status = ?", models.ContractStatusDraft).Delete(&models.Contract{}, "id = ?", c.Param("id"))
	if result.Error != nil {
		apiError(c, result.Error)
		return
	}
	if result.RowsAffected != 1 {
		response.Conflict(c, "Only an existing draft contract can be deleted")
		return
	}
	response.NoContent(c)
}
func (h *ContractHandler) Activate(c *gin.Context)  { h.transition(c, true) }
func (h *ContractHandler) Terminate(c *gin.Context) { h.transition(c, false) }
func (h *ContractHandler) transition(c *gin.Context, activate bool) {
	err := h.db.Transaction(func(tx *gorm.DB) error {
		var item models.Contract
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&item, "id = ?", c.Param("id")).Error; err != nil {
			return err
		}
		if activate {
			if item.Status != models.ContractStatusDraft && item.Status != models.ContractStatusPendingApproval {
				return invalid("Contract cannot be activated from this status")
			}
			if item.StartDate == nil || item.EndDate == nil || !item.EndDate.After(*item.StartDate) {
				return invalid("Set valid start and end dates before activation")
			}
			item.Status = models.ContractStatusActive
		} else {
			if item.Status != models.ContractStatusActive {
				return invalid("Only active contracts can be terminated")
			}
			item.Status = models.ContractStatusTerminated
		}
		if err := tx.Omit(clause.Associations).Save(&item).Error; err != nil {
			return err
		}
		return recordActivity(tx, c, "contract", item.ID, string(item.Status))
	})
	if err != nil {
		apiError(c, err)
		return
	}
	h.Get(c)
}
func (h *ContractHandler) Renew(c *gin.Context) {
	var result models.Contract
	reused := false
	err := h.db.Transaction(func(tx *gorm.DB) error {
		var original models.Contract
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&original, "id = ?", c.Param("id")).Error; err != nil {
			return err
		}
		if original.Status == models.ContractStatusRenewed {
			reused = true
			return tx.First(&result, "renewed_from_id = ?", original.ID).Error
		}
		if original.Status != models.ContractStatusActive && original.Status != models.ContractStatusExpired {
			return invalid("Only active or expired contracts can be renewed")
		}
		if original.StartDate == nil || original.EndDate == nil {
			return invalid("Original contract needs a term before renewal")
		}
		number, err := seqgen.NextNumber(tx, "contract")
		if err != nil {
			return err
		}
		start := original.EndDate.AddDate(0, 0, 1)
		end := start.Add(original.EndDate.Sub(*original.StartDate))
		result = original
		result.Base = models.Base{}
		result.ContractNumber = number
		result.Title = original.Title + " (Renewal)"
		result.Status = models.ContractStatusDraft
		result.StartDate = &start
		result.EndDate = &end
		result.QuoteID = nil
		result.RenewedFromID = &original.ID
		result.Customer = nil
		if err := tx.Omit(clause.Associations).Create(&result).Error; err != nil {
			return err
		}
		if err := tx.Model(&original).Update("status", models.ContractStatusRenewed).Error; err != nil {
			return err
		}
		return recordActivity(tx, c, "contract", result.ID, "renewed from "+original.ContractNumber)
	})
	if err != nil {
		apiError(c, err)
		return
	}
	if reused {
		response.OK(c, result)
	} else {
		response.Created(c, result)
	}
}

func (h *ContractHandler) Expiring(c *gin.Context) {
	days := 30
	var items []models.Contract
	h.db.Preload("Customer").
		Where("status = ? AND end_date >= ? AND end_date <= ?", models.ContractStatusActive, time.Now(), time.Now().AddDate(0, 0, days)).
		Find(&items)
	response.OK(c, items)
}
