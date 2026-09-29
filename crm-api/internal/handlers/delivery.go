package handlers

import (
	"time"

	"g4s-crm/api/internal/models"
	"g4s-crm/api/pkg/pagination"
	"g4s-crm/api/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Delivery lists won opportunities: the jobs owed to customers. The
// opportunity is the project, so this replaces the separate projects list and
// is authorized by the same delivery permission.
func (h *OpportunityHandler) Delivery(c *gin.Context) {
	params := pagination.GetParams(c, "title", "target_end_date", "start_date", "delivery_status")
	items := []models.Opportunity{}
	var total int64
	q := h.db.Model(&models.Opportunity{}).Where("stage = ?", models.StageClosedWon)
	if term := c.Query("q"); term != "" {
		q = q.Where("title ILIKE ?", "%"+term+"%")
	}
	for param, column := range map[string]string{"deliveryStatus": "delivery_status", "projectManagerId": "project_manager_id", "customerId": "customer_id"} {
		if value := c.Query(param); value != "" {
			q = q.Where(column+" = ?", value)
		}
	}
	if c.Query("open") == "true" {
		q = q.Where("delivery_status NOT IN ('completed', 'cancelled')")
	}
	if err := q.Count(&total).Error; err != nil {
		apiError(c, err)
		return
	}
	if err := q.Preload("Customer").Preload("ProjectManager").Preload("SalesExecutive").Preload("Sites").Preload("Quotes", currentFirst).Order(params.Sort + " " + params.Order).Scopes(pagination.Paginate(params)).Find(&items).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OKWithMeta(c, items, pagination.BuildMeta(params, total))
}

type deliveryInput struct {
	DeliveryStatus   models.DeliveryStatus `json:"deliveryStatus"`
	ProjectManagerID *uuid.UUID            `json:"projectManagerId"`
	StartDate        *string               `json:"startDate"`
	TargetEndDate    *string               `json:"targetEndDate"`
	ActualEndDate    *string               `json:"actualEndDate"`
	ActualValue      *float64              `json:"actualValue"`
	ActualCost       *float64              `json:"actualCost"`
}

var deliveryRules = map[string]string{
	"deliveryStatus":   "required,oneof=planning in-progress on-hold completed cancelled",
	"projectManagerId": "", "startDate": "", "targetEndDate": "", "actualEndDate": "",
	"actualValue": "omitempty,gte=0,lte=1000000000", "actualCost": "omitempty,gte=0,lte=1000000000",
}

func dateText(value *time.Time) *string {
	if value == nil {
		return nil
	}
	text := value.Format("2006-01-02")
	return &text
}

// UpdateDelivery records how a won job is progressing. Only a won opportunity
// has delivery; everything before that is the sales pipeline.
func (h *OpportunityHandler) UpdateDelivery(c *gin.Context) {
	var item models.Opportunity
	if err := h.db.First(&item, "id = ?", c.Param("id")).Error; err != nil {
		apiError(c, err)
		return
	}
	input := deliveryInput{ProjectManagerID: item.ProjectManagerID, StartDate: dateText(item.StartDate), TargetEndDate: dateText(item.TargetEndDate), ActualEndDate: dateText(item.ActualEndDate), ActualValue: item.ActualValue, ActualCost: item.ActualCost}
	if item.DeliveryStatus != nil {
		input.DeliveryStatus = *item.DeliveryStatus
	}
	if !bindFields(c, &input, deliveryRules) {
		return
	}
	start, err := parseDate(input.StartDate)
	if err != nil {
		apiError(c, err)
		return
	}
	target, err := parseDate(input.TargetEndDate)
	if err != nil {
		apiError(c, err)
		return
	}
	actual, err := parseDate(input.ActualEndDate)
	if err != nil {
		apiError(c, err)
		return
	}
	if start != nil && target != nil && target.Before(*start) {
		response.BadRequest(c, "The target end date must be after the start date")
		return
	}
	if start != nil && actual != nil && actual.Before(*start) {
		response.BadRequest(c, "The actual end date must be after the start date")
		return
	}
	err = h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&item, "id = ?", item.ID).Error; err != nil {
			return err
		}
		if item.Stage != models.StageClosedWon {
			return invalid("Delivery can only be recorded once the opportunity is won")
		}
		if input.ProjectManagerID != nil {
			var user models.User
			if err := tx.Where("is_active = true").First(&user, "id = ?", input.ProjectManagerID).Error; err != nil {
				return invalid("Project manager is unavailable")
			}
		}
		previous := ""
		if item.DeliveryStatus != nil {
			previous = string(*item.DeliveryStatus)
		}
		status := input.DeliveryStatus
		item.DeliveryStatus = &status
		item.ProjectManagerID = input.ProjectManagerID
		item.StartDate, item.TargetEndDate, item.ActualEndDate = start, target, actual
		item.ActualValue, item.ActualCost = input.ActualValue, input.ActualCost
		if err := tx.Omit(clause.Associations).Save(&item).Error; err != nil {
			return err
		}
		action := "delivery updated"
		if previous != string(status) {
			action = "delivery: " + previous + " → " + string(status)
		}
		return recordActivity(tx, c, "opportunity", item.ID, action)
	})
	if err != nil {
		apiError(c, err)
		return
	}
	h.Get(c)
}

type fulfilmentLine struct {
	QuoteLineItemID uuid.UUID     `json:"quoteLineItemId"`
	SKU             string        `json:"sku"`
	Description     string        `json:"description"`
	SiteID          *uuid.UUID    `json:"siteId,omitempty"`
	Quantity        float64       `json:"quantity"`
	Source          string        `json:"source"`
	Reserved        int           `json:"reserved"`
	Ordered         int           `json:"ordered"`
	Received        int           `json:"received"`
	Orders          []orderedLine `json:"orders"`
	State           string        `json:"state"`
}

type orderedLine struct {
	PurchaseOrderID string `json:"purchaseOrderId"`
	PONumber        string `json:"poNumber"`
	Status          string `json:"status"`
	Quantity        int    `json:"quantity"`
	Received        int    `json:"received"`
	Expected        string `json:"expectedDelivery,omitempty"`
}

// Fulfilment reports, line by line, how the job's quote is being filled: what
// is held in stock, what is on order and what is still open. It reads the
// accepted version when there is one, otherwise the current version as a plan.
func (h *OpportunityHandler) Fulfilment(c *gin.Context) {
	var opportunity models.Opportunity
	if err := h.db.First(&opportunity, "id = ?", c.Param("id")).Error; err != nil {
		apiError(c, err)
		return
	}
	var quote models.Quote
	err := h.db.Where("opportunity_id = ? AND status = ?", opportunity.ID, models.QuoteStatusAccepted).Order("version DESC").First(&quote).Error
	basis := "accepted"
	if err == gorm.ErrRecordNotFound {
		basis = "current"
		err = h.db.Where("opportunity_id = ? AND is_current", opportunity.ID).First(&quote).Error
	}
	if err == gorm.ErrRecordNotFound {
		response.OK(c, gin.H{"basis": "none", "lines": []fulfilmentLine{}})
		return
	}
	if err != nil {
		apiError(c, err)
		return
	}
	var items []models.QuoteLineItem
	if err := h.db.Where("quote_id = ? AND (row_type = 'item' OR row_type = '')", quote.ID).Order("sort_order").Find(&items).Error; err != nil {
		apiError(c, err)
		return
	}
	lines := []fulfilmentLine{}
	for _, item := range items {
		if item.IsOptional && !item.IsSelected {
			continue
		}
		line := fulfilmentLine{QuoteLineItemID: item.ID, SKU: item.SKU, Description: item.Description, SiteID: item.SiteID, Quantity: item.Quantity * item.Multiplier, Orders: []orderedLine{}}
		switch {
		case item.FulfilmentSource != nil:
			line.Source = *item.FulfilmentSource
		case item.Source == "service" || item.Source == "recurring":
			line.Source = "service"
		default:
			line.Source = "unassigned"
		}
		if err := h.db.Model(&models.StockReservation{}).Where("quote_line_item_id = ? AND status IN ('active', 'fulfilled')", item.ID).Select("COALESCE(SUM(qty), 0)").Scan(&line.Reserved).Error; err != nil {
			apiError(c, err)
			return
		}
		type row struct {
			PurchaseOrderID  string
			PONumber         string
			Status           string
			Quantity         int
			ReceivedQty      int
			ExpectedDelivery *time.Time
		}
		var rows []row
		if err := h.db.Table("purchase_order_items").Select("purchase_orders.id AS purchase_order_id, purchase_orders.po_number, purchase_orders.status, purchase_order_items.quantity, purchase_order_items.received_qty, purchase_orders.expected_delivery").
			Joins("JOIN purchase_orders ON purchase_orders.id = purchase_order_items.purchase_order_id AND purchase_orders.deleted_at IS NULL").
			Where("purchase_order_items.quote_line_item_id = ? AND purchase_order_items.deleted_at IS NULL AND purchase_orders.status <> 'cancelled'", item.ID).Scan(&rows).Error; err != nil {
			apiError(c, err)
			return
		}
		for _, r := range rows {
			order := orderedLine{PurchaseOrderID: r.PurchaseOrderID, PONumber: r.PONumber, Status: r.Status, Quantity: r.Quantity, Received: r.ReceivedQty}
			if r.ExpectedDelivery != nil {
				order.Expected = r.ExpectedDelivery.Format("2006-01-02")
			}
			line.Orders = append(line.Orders, order)
			line.Ordered += r.Quantity
			line.Received += r.ReceivedQty
		}
		needed := int(line.Quantity + 0.999)
		switch {
		case line.Source == "service":
			line.State = "not-applicable"
		case line.Reserved+line.Received >= needed:
			line.State = "covered"
		case line.Reserved+line.Ordered >= needed:
			line.State = "on-order"
		case line.Reserved+line.Ordered > 0:
			line.State = "partial"
		default:
			line.State = "open"
		}
		lines = append(lines, line)
	}
	response.OK(c, gin.H{"basis": basis, "quoteId": quote.ID, "quoteNumber": quote.QuoteNumber, "version": quote.Version, "lines": lines})
}

// Activity is the opportunity's audit trail: stage moves, delivery updates and
// the like, oldest first.
func (h *OpportunityHandler) Activity(c *gin.Context) {
	var item models.Opportunity
	if err := h.db.First(&item, "id = ?", c.Param("id")).Error; err != nil {
		apiError(c, err)
		return
	}
	items := []models.ActivityLog{}
	if err := h.db.Preload("User").Where("entity_type = 'opportunity' AND entity_id = ?", item.ID).Order("created_at ASC").Find(&items).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OK(c, items)
}
