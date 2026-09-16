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

	h.db.Model(&models.Quote{}).Where("status = 'accepted'").Select("COALESCE(SUM(total), 0)").Scan(&totalRevenue)
	h.db.Model(&models.Opportunity{}).Where("stage NOT IN ('closed-won','closed-lost')").Count(&activeOpps)
	h.db.Model(&models.Quote{}).Where("status IN ('draft','pending-approval','approved','sent')").Count(&openQuotes)
	h.db.Model(&models.Quote{}).Where("status IN ('draft','pending-approval','approved','sent')").Select("COALESCE(SUM(total), 0)").Scan(&openQuotesValue)
	h.db.Model(&models.Contract{}).Where("status = 'active'").Count(&activeContracts)
	h.db.Model(&models.Contract{}).Where("status = 'active' AND end_date <= NOW() + INTERVAL '30 days'").Count(&expiringContracts)

	response.OK(c, gin.H{
		"totalRevenue":       gin.H{"value": totalRevenue},
		"activeOpportunities": gin.H{"value": activeOpps},
		"openQuotes":         gin.H{"value": openQuotes, "totalValue": openQuotesValue},
		"activeContracts":    gin.H{"value": activeContracts, "expiringIn30Days": expiringContracts},
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
		Where("quotes.status = 'accepted'").
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
		Where("status = 'active' AND end_date <= NOW() + INTERVAL '30 days'").
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
