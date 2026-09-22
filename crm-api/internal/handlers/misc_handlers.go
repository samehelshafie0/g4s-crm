package handlers

import (
	"g4s-crm/api/internal/models"
	"g4s-crm/api/pkg/pagination"
	"g4s-crm/api/pkg/response"
	"g4s-crm/api/pkg/seqgen"
	v "g4s-crm/api/pkg/validator"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

// Each resource has an explicit write whitelist; derived values and associations
// are saved by resource-specific operations below.

type TeamHandler struct{ db *gorm.DB }

func NewTeamHandler(db *gorm.DB) *TeamHandler { return &TeamHandler{db: db} }
func (h *TeamHandler) List(c *gin.Context) {
	params := pagination.GetParams(c, "name")
	items := []models.Team{}
	var total int64
	q := h.db.Model(&models.Team{})
	if search := c.Query("q"); search != "" {
		q = q.Where("name ILIKE ?", "%"+search+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		apiError(c, err)
		return
	}
	if err := q.Preload("Members").Preload("Leader").Order(params.Sort + " " + params.Order).Scopes(pagination.Paginate(params)).Find(&items).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OKWithMeta(c, items, pagination.BuildMeta(params, total))
}
func (h *TeamHandler) Get(c *gin.Context) {
	var item models.Team
	if err := h.db.Preload("Members").Preload("Leader").First(&item, "id = ?", c.Param("id")).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OK(c, item)
}
func (h *TeamHandler) Create(c *gin.Context) { h.save(c, true) }
func (h *TeamHandler) Update(c *gin.Context) { h.save(c, false) }
func (h *TeamHandler) save(c *gin.Context, create bool) {
	item := models.Team{IsActive: true}
	if !create {
		if err := h.db.First(&item, "id = ?", c.Param("id")).Error; err != nil {
			apiError(c, err)
			return
		}
	}
	if !bindFields(c, &item, map[string]string{"name": "required,max=200", "department": departmentRule, "description": "max=5000", "leaderId": "", "isActive": ""}) {
		return
	}
	active := item.IsActive
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if item.LeaderID != nil {
			var leader models.User
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("is_active = true").First(&leader, "id = ?", item.LeaderID).Error; err != nil {
				return err
			}
			if leader.TeamID != nil && *leader.TeamID != item.ID {
				return invalid("Leader already belongs to another team")
			}
		}
		if err := tx.Omit(clause.Associations).Save(&item).Error; err != nil {
			return err
		}
		if err := tx.Model(&item).Update("is_active", active).Error; err != nil {
			return err
		}
		if item.LeaderID != nil {
			return tx.Model(&models.User{}).Where("id = ?", item.LeaderID).Update("team_id", item.ID).Error
		}
		return nil
	})
	if err != nil {
		apiError(c, err)
		return
	}
	if err := h.db.Preload("Members").Preload("Leader").First(&item, "id = ?", item.ID).Error; err != nil {
		apiError(c, err)
		return
	}
	if create {
		response.Created(c, item)
	} else {
		response.OK(c, item)
	}
}
func (h *TeamHandler) Delete(c *gin.Context) {
	err := h.db.Transaction(func(tx *gorm.DB) error {
		var team models.Team
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&team, "id = ?", c.Param("id")).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.User{}).Where("team_id = ?", team.ID).Update("team_id", nil).Error; err != nil {
			return err
		}
		return tx.Delete(&team).Error
	})
	if err != nil {
		apiError(c, err)
		return
	}
	response.NoContent(c)
}

type PriceBookHandler struct{ db *gorm.DB }

func NewPriceBookHandler(db *gorm.DB) *PriceBookHandler { return &PriceBookHandler{db: db} }
func (h *PriceBookHandler) List(c *gin.Context) {
	params := pagination.GetParams(c, "name")
	items := []models.PriceBook{}
	var total int64
	q := h.db.Model(&models.PriceBook{})
	if search := c.Query("q"); search != "" {
		q = q.Where("name ILIKE ?", "%"+search+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		apiError(c, err)
		return
	}
	if err := q.Preload("Entries.Product").Preload("Entries.Service").Preload("Entries.RecurringService").Preload("Customer").Order(params.Sort + " " + params.Order).Scopes(pagination.Paginate(params)).Find(&items).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OKWithMeta(c, items, pagination.BuildMeta(params, total))
}
func (h *PriceBookHandler) Get(c *gin.Context) {
	var item models.PriceBook
	if err := h.db.Preload("Entries.Product").Preload("Entries.Service").Preload("Entries.RecurringService").Preload("Customer").First(&item, "id = ?", c.Param("id")).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OK(c, item)
}
func (h *PriceBookHandler) Create(c *gin.Context) { h.save(c, true) }
func (h *PriceBookHandler) Update(c *gin.Context) { h.save(c, false) }
func (h *PriceBookHandler) save(c *gin.Context, create bool) {
	item := models.PriceBook{IsActive: true, Type: models.PriceBookStandard}
	if !create {
		if err := h.db.First(&item, "id = ?", c.Param("id")).Error; err != nil {
			apiError(c, err)
			return
		}
	}
	if !bindFields(c, &item, map[string]string{"name": "required,max=200", "type": "required,oneof=standard volume contract promotional customer-specific", "description": "max=5000", "customerId": "", "contractId": "", "validFrom": "", "validTo": "", "isActive": ""}) {
		return
	}
	active := item.IsActive
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := optionalExists(tx, &models.Customer{}, item.CustomerID); err != nil {
			return err
		}
		if err := optionalExists(tx, &models.Contract{}, item.ContractID); err != nil {
			return err
		}
		if item.ValidFrom != nil && item.ValidTo != nil && item.ValidTo.Before(*item.ValidFrom) {
			return invalid("Valid to must follow valid from")
		}
		if item.ContractID != nil {
			var contract models.Contract
			if err := tx.First(&contract, "id = ?", item.ContractID).Error; err != nil {
				return err
			}
			if item.CustomerID != nil && *item.CustomerID != contract.CustomerID {
				return invalid("Contract belongs to another customer")
			}
			item.CustomerID = &contract.CustomerID
		}
		if err := tx.Omit(clause.Associations).Save(&item).Error; err != nil {
			return err
		}
		if err := tx.Model(&item).Update("is_active", active).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		apiError(c, err)
		return
	}
	if err := h.db.Preload("Entries.Product").Preload("Entries.Service").Preload("Entries.RecurringService").Preload("Customer").First(&item, "id = ?", item.ID).Error; err != nil {
		apiError(c, err)
		return
	}
	if create {
		response.Created(c, item)
	} else {
		response.OK(c, item)
	}
}
func (h *PriceBookHandler) Delete(c *gin.Context) { deleteRecord(c, h.db, &models.PriceBook{}) }

type RecurringServiceHandler struct{ db *gorm.DB }

func NewRecurringServiceHandler(db *gorm.DB) *RecurringServiceHandler {
	return &RecurringServiceHandler{db: db}
}
func (h *RecurringServiceHandler) List(c *gin.Context) {
	params := pagination.GetParams(c, "name")
	items := []models.RecurringService{}
	var total int64
	q := h.db.Model(&models.RecurringService{})
	if search := c.Query("q"); search != "" {
		q = q.Where("name ILIKE ?", "%"+search+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		apiError(c, err)
		return
	}
	if err := q.Order(params.Sort + " " + params.Order).Scopes(pagination.Paginate(params)).Find(&items).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OKWithMeta(c, items, pagination.BuildMeta(params, total))
}
func (h *RecurringServiceHandler) Get(c *gin.Context) {
	var item models.RecurringService
	if err := h.db.First(&item, "id = ?", c.Param("id")).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OK(c, item)
}
func (h *RecurringServiceHandler) Create(c *gin.Context) { h.save(c, true) }
func (h *RecurringServiceHandler) Update(c *gin.Context) { h.save(c, false) }
func (h *RecurringServiceHandler) save(c *gin.Context, create bool) {
	item := models.RecurringService{IsActive: true, BillingFrequency: models.BillingMonthly, ServiceType: models.RecurringGuarding}
	if !create {
		if err := h.db.First(&item, "id = ?", c.Param("id")).Error; err != nil {
			apiError(c, err)
			return
		}
	}
	if !bindFields(c, &item, map[string]string{"name": "required,max=200", "serviceType": "required,oneof=guarding maintenance monitoring patrol facility-management rental", "description": "max=5000", "lineItems": "max=1000,dive", "monthlyCost": moneyRule, "monthlyPrice": moneyRule, "targetMarginPercent": percentRule, "billingFrequency": "required,oneof=monthly quarterly annually", "isActive": ""}) {
		return
	}
	active := item.IsActive
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if item.LineItems == nil {
			item.LineItems = []models.RecurringLine{}
		}
		if len(item.LineItems) > 0 {
			item.MonthlyCost = 0
			for _, line := range item.LineItems {
				if line.Source == "product" {
					if line.SourceID == nil {
						return invalid("Product source ID required")
					}
					if err := exists(tx, &models.Product{}, *line.SourceID); err != nil {
						return err
					}
				}
				if line.Source == "service" {
					if line.SourceID == nil || *line.SourceID == item.ID {
						return invalid("Choose a different source service")
					}
					if err := exists(tx, &models.RecurringService{}, *line.SourceID); err != nil {
						return err
					}
				}
				item.MonthlyCost += models.RoundMoney(line.Qty * line.UnitCost)
			}
			item.MonthlyCost = models.RoundMoney(item.MonthlyCost)
		}
		item.AnnualCost = models.RoundMoney(item.MonthlyCost * 12)
		item.AnnualPrice = models.RoundMoney(item.MonthlyPrice * 12)
		if err := tx.Omit(clause.Associations).Save(&item).Error; err != nil {
			return err
		}
		if err := tx.Model(&item).Update("is_active", active).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		apiError(c, err)
		return
	}
	if err := h.db.First(&item, "id = ?", item.ID).Error; err != nil {
		apiError(c, err)
		return
	}
	if create {
		response.Created(c, item)
	} else {
		response.OK(c, item)
	}
}
func (h *RecurringServiceHandler) Delete(c *gin.Context) {
	deleteRecord(c, h.db, &models.RecurringService{})
}

type CatalogServiceHandler struct{ db *gorm.DB }

func NewCatalogServiceHandler(db *gorm.DB) *CatalogServiceHandler {
	return &CatalogServiceHandler{db: db}
}
func (h *CatalogServiceHandler) List(c *gin.Context) {
	params := pagination.GetParams(c, "name")
	items := []models.CatalogService{}
	var total int64
	q := h.db.Model(&models.CatalogService{})
	if search := c.Query("q"); search != "" {
		q = q.Where("name ILIKE ?", "%"+search+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		apiError(c, err)
		return
	}
	if err := q.Order(params.Sort + " " + params.Order).Scopes(pagination.Paginate(params)).Find(&items).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OKWithMeta(c, items, pagination.BuildMeta(params, total))
}
func (h *CatalogServiceHandler) Get(c *gin.Context) {
	var item models.CatalogService
	if err := h.db.First(&item, "id = ?", c.Param("id")).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OK(c, item)
}
func (h *CatalogServiceHandler) Create(c *gin.Context) { h.save(c, true) }
func (h *CatalogServiceHandler) Update(c *gin.Context) { h.save(c, false) }
func (h *CatalogServiceHandler) save(c *gin.Context, create bool) {
	item := models.CatalogService{IsActive: true, Department: models.DeptTechnical, RateType: "hour"}
	if !create {
		if err := h.db.First(&item, "id = ?", c.Param("id")).Error; err != nil {
			apiError(c, err)
			return
		}
	}
	if !bindFields(c, &item, map[string]string{"sku": "required,max=100", "name": "required,max=200", "description": "max=5000", "department": departmentRule, "rateType": "required,max=50", "unitCost": moneyRule, "unitPrice": moneyRule, "isActive": ""}) {
		return
	}
	active := item.IsActive
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit(clause.Associations).Save(&item).Error; err != nil {
			return err
		}
		if err := tx.Model(&item).Update("is_active", active).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		apiError(c, err)
		return
	}
	if err := h.db.First(&item, "id = ?", item.ID).Error; err != nil {
		apiError(c, err)
		return
	}
	if create {
		response.Created(c, item)
	} else {
		response.OK(c, item)
	}
}
func (h *CatalogServiceHandler) Delete(c *gin.Context) {
	deleteRecord(c, h.db, &models.CatalogService{})
}

type ProjectHandler struct{ db *gorm.DB }

func NewProjectHandler(db *gorm.DB) *ProjectHandler { return &ProjectHandler{db: db} }
func (h *ProjectHandler) List(c *gin.Context) {
	params := pagination.GetParams(c, "name")
	items := []models.Project{}
	var total int64
	q := h.db.Model(&models.Project{})
	if search := c.Query("q"); search != "" {
		q = q.Where("name ILIKE ?", "%"+search+"%")
	}
	if err := q.Count(&total).Error; err != nil {
		apiError(c, err)
		return
	}
	if err := q.Preload("Customer").Preload("Quote.LineItems").Preload("ProjectManager").Order(params.Sort + " " + params.Order).Scopes(pagination.Paginate(params)).Find(&items).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OKWithMeta(c, items, pagination.BuildMeta(params, total))
}
func (h *ProjectHandler) Get(c *gin.Context) {
	var item models.Project
	if err := h.db.Preload("Customer").Preload("Quote.LineItems").Preload("ProjectManager").First(&item, "id = ?", c.Param("id")).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OK(c, item)
}
func (h *ProjectHandler) Create(c *gin.Context) { h.save(c, true) }
func (h *ProjectHandler) Update(c *gin.Context) { h.save(c, false) }
func (h *ProjectHandler) save(c *gin.Context, create bool) {
	item := models.Project{Status: models.ProjectStatusPlanning, Priority: models.PriorityMedium}
	if !create {
		if err := h.db.First(&item, "id = ?", c.Param("id")).Error; err != nil {
			apiError(c, err)
			return
		}
	}
	if !bindFields(c, &item, map[string]string{"name": "required,max=200", "customerId": "required", "quoteId": "", "status": "required,oneof=planning in-progress on-hold completed cancelled", "priority": "required,oneof=low medium high critical", "startDate": "", "targetEndDate": "", "actualEndDate": "", "projectManagerId": "", "notes": "max=20000"}) {
		return
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := exists(tx, &models.Customer{}, item.CustomerID); err != nil {
			return err
		}
		if err := optionalExists(tx, &models.User{}, item.ProjectManagerID); err != nil {
			return err
		}
		if item.StartDate != nil && item.TargetEndDate != nil && item.TargetEndDate.Before(*item.StartDate) {
			return invalid("Target end date must follow start date")
		}
		if item.QuoteID != nil {
			var quote models.Quote
			if err := tx.First(&quote, "id = ?", item.QuoteID).Error; err != nil {
				return err
			}
			if quote.CustomerID != item.CustomerID {
				return invalid("Quote belongs to another customer")
			}
			if quote.Status != models.QuoteStatusAccepted {
				return invalid("Project requires an accepted quote")
			}
			item.Currency = quote.Currency
			item.TotalValue = quote.Total
			item.TotalCost = quote.TotalCost
			item.MarginPercent = quote.MarginPercent
		}
		if create {
			var err error
			item.ProjectNumber, err = seqgen.NextNumber(tx, "project")
			if err != nil {
				return err
			}
		}
		if err := tx.Omit(clause.Associations).Save(&item).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		apiError(c, err)
		return
	}
	if err := h.db.Preload("Customer").Preload("Quote.LineItems").Preload("ProjectManager").First(&item, "id = ?", item.ID).Error; err != nil {
		apiError(c, err)
		return
	}
	if create {
		response.Created(c, item)
	} else {
		response.OK(c, item)
	}
}
func (h *ProjectHandler) Delete(c *gin.Context) { deleteRecord(c, h.db, &models.Project{}) }

func (h *TeamHandler) AddMember(c *gin.Context) {
	var req struct {
		UserID uuid.UUID `json:"userId" validate:"required"`
	}
	if !v.BindStrict(c, &req) {
		return
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		var team models.Team
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&team, "id = ?", c.Param("id")).Error; err != nil {
			return err
		}
		var user models.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, "id = ? AND is_active = true", req.UserID).Error; err != nil {
			return err
		}
		if user.TeamID != nil && *user.TeamID != team.ID {
			return invalid("User already belongs to another team")
		}
		return tx.Model(&user).Update("team_id", team.ID).Error
	})
	if err != nil {
		apiError(c, err)
		return
	}
	h.Get(c)
}
func (h *TeamHandler) RemoveMember(c *gin.Context) {
	err := h.db.Transaction(func(tx *gorm.DB) error {
		var team models.Team
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&team, "id = ?", c.Param("id")).Error; err != nil {
			return err
		}
		result := tx.Model(&models.User{}).Where("id = ? AND team_id = ?", c.Param("userId"), team.ID).Update("team_id", nil)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return gorm.ErrRecordNotFound
		}
		if team.LeaderID != nil && team.LeaderID.String() == c.Param("userId") {
			return tx.Model(&team).Update("leader_id", nil).Error
		}
		return nil
	})
	if err != nil {
		apiError(c, err)
		return
	}
	response.NoContent(c)
}
func (h *PriceBookHandler) Entries(c *gin.Context) {
	var book models.PriceBook
	if err := h.db.Preload("Entries.Product").Preload("Entries.Service").Preload("Entries.RecurringService").First(&book, "id = ?", c.Param("id")).Error; err != nil {
		apiError(c, err)
		return
	}
	if book.Entries == nil {
		book.Entries = []models.PriceBookEntry{}
	}
	response.OK(c, book.Entries)
}
func (h *PriceBookHandler) SaveEntry(c *gin.Context) {
	item := models.PriceBookEntry{Kind: "product"}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		var book models.PriceBook
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&book, "id = ?", c.Param("id")).Error; err != nil {
			return err
		}
		item.PriceBookID = book.ID
		if c.Param("entryId") != "" {
			if err := tx.First(&item, "id = ? AND price_book_id = ?", c.Param("entryId"), book.ID).Error; err != nil {
				return err
			}
		}
		if !bindFields(c, &item, map[string]string{"kind": "required,oneof=product service recurring", "productId": "", "serviceId": "", "recurringServiceId": "", "customPrice": moneyRule}) {
			return invalid("Invalid entry")
		}
		if err := priceEntry(tx, &item); err != nil {
			return err
		}
		return tx.Omit(clause.Associations).Save(&item).Error
	})
	if err != nil {
		if !c.Writer.Written() {
			apiError(c, err)
		}
		return
	}
	h.Entries(c)
}
func priceEntry(tx *gorm.DB, item *models.PriceBookEntry) error {
	count := 0
	for _, id := range []*uuid.UUID{item.ProductID, item.ServiceID, item.RecurringServiceID} {
		if id != nil {
			count++
		}
	}
	if count != 1 {
		return invalid("Select exactly one catalog item")
	}
	switch item.Kind {
	case "product":
		if item.ProductID == nil {
			return invalid("Product required")
		}
		var product models.Product
		if err := tx.First(&product, "id = ?", item.ProductID).Error; err != nil {
			return err
		}
		item.StandardPrice = product.SellingPrice
	case "service":
		if item.ServiceID == nil {
			return invalid("Service required")
		}
		var service models.CatalogService
		if err := tx.First(&service, "id = ?", item.ServiceID).Error; err != nil {
			return err
		}
		item.StandardPrice = service.UnitPrice
	case "recurring":
		if item.RecurringServiceID == nil {
			return invalid("Recurring service required")
		}
		var service models.RecurringService
		if err := tx.First(&service, "id = ?", item.RecurringServiceID).Error; err != nil {
			return err
		}
		item.StandardPrice = service.MonthlyPrice
	default:
		return invalid("Invalid item kind")
	}
	item.DiscountPercent = 0
	if item.StandardPrice > 0 {
		item.DiscountPercent = (1 - item.CustomPrice/item.StandardPrice) * 100
	}
	return nil
}
func (h *PriceBookHandler) ReplaceEntries(c *gin.Context) {
	var req struct {
		Entries []struct {
			Kind               string     `json:"kind" validate:"required,oneof=product service recurring"`
			ProductID          *uuid.UUID `json:"productId"`
			ServiceID          *uuid.UUID `json:"serviceId"`
			RecurringServiceID *uuid.UUID `json:"recurringServiceId"`
			CustomPrice        float64    `json:"customPrice" validate:"gte=0,lte=100000000000"`
		} `json:"entries" validate:"max=1000,dive"`
	}
	if !v.BindStrict(c, &req) {
		return
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		var book models.PriceBook
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&book, "id = ?", c.Param("id")).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("price_book_id = ?", book.ID).Delete(&models.PriceBookEntry{}).Error; err != nil {
			return err
		}
		seen := map[string]bool{}
		for _, input := range req.Entries {
			item := models.PriceBookEntry{PriceBookID: book.ID, Kind: input.Kind, ProductID: input.ProductID, ServiceID: input.ServiceID, RecurringServiceID: input.RecurringServiceID, CustomPrice: input.CustomPrice}
			if err := priceEntry(tx, &item); err != nil {
				return err
			}
			id := item.ProductID
			if id == nil {
				id = item.ServiceID
			}
			if id == nil {
				id = item.RecurringServiceID
			}
			key := item.Kind + id.String()
			if seen[key] {
				return invalid("Duplicate item in price book")
			}
			seen[key] = true
			if err := tx.Create(&item).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		apiError(c, err)
		return
	}
	h.Get(c)
}
func (h *PriceBookHandler) DeleteEntry(c *gin.Context) {
	result := h.db.Where("price_book_id = ?", c.Param("id")).Delete(&models.PriceBookEntry{}, "id = ?", c.Param("entryId"))
	if result.Error != nil {
		apiError(c, result.Error)
		return
	}
	if result.RowsAffected != 1 {
		response.NotFound(c, "Entry not found")
		return
	}
	response.NoContent(c)
}

type ExchangeRateHandler struct{ db *gorm.DB }

func NewExchangeRateHandler(db *gorm.DB) *ExchangeRateHandler { return &ExchangeRateHandler{db: db} }
func (h *ExchangeRateHandler) List(c *gin.Context) {
	items := []models.ExchangeRate{}
	if err := h.db.Preload("History", func(db *gorm.DB) *gorm.DB { return db.Order("effective_date DESC") }).Order("from_currency").Find(&items).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OK(c, items)
}
func (h *ExchangeRateHandler) Get(c *gin.Context) {
	var item models.ExchangeRate
	if err := h.db.Preload("History").First(&item, "id = ?", c.Param("id")).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OK(c, item)
}
func (h *ExchangeRateHandler) Create(c *gin.Context) { h.save(c, true) }
func (h *ExchangeRateHandler) Update(c *gin.Context) { h.save(c, false) }
func (h *ExchangeRateHandler) save(c *gin.Context, create bool) {
	item := models.ExchangeRate{ToCurrency: "SAR", EffectiveDate: time.Now()}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if !create {
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&item, "id = ?", c.Param("id")).Error; err != nil {
				return err
			}
		}
		fields := map[string]string{"currentRate": "required,gt=0,lte=1000000", "effectiveDate": "required"}
		if create {
			fields["fromCurrency"] = currencyRule
			fields["toCurrency"] = currencyRule
		}
		if !bindFields(c, &item, fields) {
			return invalid("Invalid exchange rate")
		}
		if string(item.FromCurrency) == item.ToCurrency && item.CurrentRate != 1 {
			return invalid("Same-currency rate must be one")
		}
		if err := tx.Omit(clause.Associations).Save(&item).Error; err != nil {
			return err
		}
		return tx.Create(&models.ExchangeRateHistory{ExchangeRateID: item.ID, Rate: item.CurrentRate, EffectiveDate: item.EffectiveDate}).Error
	})
	if err != nil {
		if !c.Writer.Written() {
			apiError(c, err)
		}
		return
	}
	if err := h.db.Preload("History").First(&item, "id = ?", item.ID).Error; err != nil {
		apiError(c, err)
		return
	}
	if create {
		response.Created(c, item)
	} else {
		response.OK(c, item)
	}
}
