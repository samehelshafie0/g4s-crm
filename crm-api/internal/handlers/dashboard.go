package handlers

import (
	"g4s-crm/api/internal/models"
	"g4s-crm/api/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DashboardHandler struct{ db *gorm.DB }

func NewDashboardHandler(db *gorm.DB) *DashboardHandler { return &DashboardHandler{db: db} }

func (h *DashboardHandler) KPIs(c *gin.Context) {
	var totalRevenue, openQuotesValue float64
	var activeOpps, openQuotes, activeContracts, expiringContracts int64

	h.db.Model(&models.Quote{}).Where("status = 'accepted' AND currency = 'SAR'").Select("COALESCE(SUM(total), 0)").Scan(&totalRevenue)
	h.db.Model(&models.Opportunity{}).Where("stage NOT IN ('closed-won','closed-lost')").Count(&activeOpps)
	h.db.Model(&models.Quote{}).Where("status IN ('draft','pending-approval','approved','sent')").Count(&openQuotes)
	h.db.Model(&models.Quote{}).Where("status IN ('draft','pending-approval','approved','sent') AND currency = 'SAR'").Select("COALESCE(SUM(total), 0)").Scan(&openQuotesValue)
	h.db.Model(&models.Contract{}).Where("status = 'active'").Count(&activeContracts)
	h.db.Model(&models.Contract{}).Where("status = 'active' AND end_date >= CURRENT_DATE AND end_date <= CURRENT_DATE + INTERVAL '30 days'").Count(&expiringContracts)

	response.OK(c, gin.H{
		"totalRevenue":        gin.H{"value": totalRevenue},
		"activeOpportunities": gin.H{"value": activeOpps},
		"openQuotes":          gin.H{"value": openQuotes, "totalValue": openQuotesValue},
		"activeContracts":     gin.H{"value": activeContracts, "expiringIn30Days": expiringContracts},
	})
}

func (h *DashboardHandler) Pipeline(c *gin.Context) {
	type StageResult struct {
		Stage models.OpportunityStage `json:"stage"`
		Count int64                   `json:"count"`
		Value float64                 `json:"value"`
	}
	var stages []StageResult
	h.db.Model(&models.Opportunity{}).
		Select("stage, COUNT(*) as count, COALESCE(SUM(estimated_value),0) as value").
		Group("stage").
		Find(&stages)
	response.OK(c, gin.H{"stages": stages})
}

func (h *DashboardHandler) RecentActivity(c *gin.Context) {
	limit := 20
	var activities []models.ActivityLog
	h.db.Preload("User").Order("created_at DESC").Limit(limit).Find(&activities)
	response.OK(c, activities)
}

func (h *DashboardHandler) TopCustomers(c *gin.Context) {
	type TopCustomer struct {
		CustomerID   string  `json:"customerId"`
		CompanyName  string  `json:"companyName"`
		TotalRevenue float64 `json:"totalRevenue"`
		QuoteCount   int64   `json:"quoteCount"`
	}
	var results []TopCustomer
	h.db.Model(&models.Quote{}).
		Select("customer_id, customers.company_name, SUM(total) as total_revenue, COUNT(*) as quote_count").
		Joins("JOIN customers ON customers.id = quotes.customer_id").
		Where("quotes.status = 'accepted' AND quotes.currency = 'SAR' AND customers.deleted_at IS NULL").
		Group("quotes.customer_id, customers.company_name").
		Order("total_revenue DESC").
		Limit(10).
		Scan(&results)
	response.OK(c, results)
}

func (h *DashboardHandler) Alerts(c *gin.Context) {
	var alerts []gin.H

	var expiringContracts int64
	h.db.Model(&models.Contract{}).
		Where("status = 'active' AND end_date >= CURRENT_DATE AND end_date <= CURRENT_DATE + INTERVAL '30 days'").
		Count(&expiringContracts)
	if expiringContracts > 0 {
		alerts = append(alerts, gin.H{
			"type": "contract_expiring", "severity": "warning",
			"message": "contracts expiring in the next 30 days", "count": expiringContracts,
		})
	}

	var pendingApprovals int64
	h.db.Model(&models.Quote{}).Where("status = 'pending-approval'").Count(&pendingApprovals)
	if pendingApprovals > 0 {
		alerts = append(alerts, gin.H{
			"type": "pending_approvals", "severity": "action",
			"message": "quotes awaiting approval", "count": pendingApprovals,
		})
	}

	if alerts == nil {
		alerts = []gin.H{}
	}
	response.OK(c, alerts)
}

func (h *DashboardHandler) RecentQuotes(c *gin.Context) {
	items := []models.Quote{}
	if err := h.db.Preload("Customer").Order("created_at DESC").Limit(10).Find(&items).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OK(c, items)
}
func (h *DashboardHandler) ExpiringQuotes(c *gin.Context) {
	items := []models.Quote{}
	if err := h.db.Preload("Customer").Where("status IN ('approved','sent') AND valid_until >= CURRENT_DATE AND valid_until <= CURRENT_DATE + INTERVAL '30 days'").Order("valid_until").Find(&items).Error; err != nil {
		apiError(c, err)
		return
	}
	response.OK(c, items)
}
func (h *DashboardHandler) SalesPerformance(c *gin.Context) {
	type entry struct {
		Name          string  `json:"name"`
		QuotesCreated int     `json:"quotesCreated"`
		QuotesWon     int     `json:"quotesWon"`
		Revenue       float64 `json:"revenue"`
		WinRate       float64 `json:"winRate"`
	}
	items := []entry{}
	err := h.db.Model(&models.Quote{}).Select("users.first_name || ' ' || users.last_name AS name, COUNT(*) AS quotes_created, COUNT(*) FILTER (WHERE quotes.status = 'accepted') AS quotes_won, COALESCE(SUM(quotes.total) FILTER (WHERE quotes.status = 'accepted' AND quotes.currency = 'SAR'),0) AS revenue").Joins("JOIN users ON users.id=quotes.created_by_id AND users.deleted_at IS NULL").Where("quotes.created_at >= date_trunc('quarter', CURRENT_DATE)").Group("users.id,users.first_name,users.last_name").Order("revenue DESC").Scan(&items).Error
	if err != nil {
		apiError(c, err)
		return
	}
	for i := range items {
		if items[i].QuotesCreated > 0 {
			items[i].WinRate = float64(items[i].QuotesWon) / float64(items[i].QuotesCreated) * 100
		}
	}
	response.OK(c, items)
}
