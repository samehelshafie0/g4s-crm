package handlers

import (
	"g4s-crm/api/internal/models"
	"g4s-crm/api/pkg/pagination"
	"g4s-crm/api/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Operational demand deliberately excludes customer pricing, margins and notes.
type jobDemandLine struct {
	ID           uuid.UUID `json:"id"`
	ProductID    uuid.UUID `json:"productId"`
	SKU          string    `json:"sku"`
	Description  string    `json:"description"`
	Category     string    `json:"category"`
	Quantity     float64   `json:"quantity"`
	Multiplier   float64   `json:"multiplier"`
	RemainingQty int       `json:"remainingQty"`
}
type jobDemand struct {
	ID               uuid.UUID       `json:"id"`
	QuoteNumber      string          `json:"quoteNumber"`
	Version          int             `json:"version"`
	Status           string          `json:"status"`
	CustomerName     string          `json:"customerName"`
	OpportunityID    *uuid.UUID      `json:"opportunityId,omitempty"`
	OpportunityTitle string          `json:"opportunityTitle"`
	LineItems        []jobDemandLine `json:"lineItems"`
}

func (h *ProcurementHandler) JobDemands(c *gin.Context) {
	params := pagination.GetParams(c, "quote_number")
	query := h.db.Model(&models.Quote{}).Where("status = ? AND is_current", models.QuoteStatusAccepted)
	if id := c.Param("id"); id != "" {
		parsed, err := uuid.Parse(id)
		if err != nil {
			response.NotFound(c, "Accepted job not found")
			return
		}
		query = query.Where("id = ?", parsed)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		apiError(c, err)
		return
	}
	var quotes []models.Quote
	if err := query.Preload("Customer").Preload("Opportunity").Preload("LineItems.Product").Order("quote_number ASC").Scopes(pagination.Paginate(params)).Find(&quotes).Error; err != nil {
		apiError(c, err)
		return
	}
	result := make([]jobDemand, 0, len(quotes))
	for _, q := range quotes {
		job := jobDemand{ID: q.ID, QuoteNumber: q.QuoteNumber, Version: q.Version, Status: "accepted", OpportunityID: q.OpportunityID, LineItems: []jobDemandLine{}}
		if q.Customer != nil {
			job.CustomerName = q.Customer.CompanyName
		}
		if q.Opportunity != nil {
			job.OpportunityTitle = q.Opportunity.Title
		}
		for _, line := range q.LineItems {
			if line.ProductID == nil || line.Category != models.QuoteLineMaterials || (line.RowType != "" && line.RowType != "item") || (line.IsOptional && !line.IsSelected) || (line.FulfilmentSource != nil && *line.FulfilmentSource == "service") {
				continue
			}
			remaining, err := lineRemaining(h.db, line, uuid.Nil)
			if err != nil {
				apiError(c, err)
				return
			}
			sku := line.SKU
			if sku == "" && line.Product != nil {
				sku = line.Product.SKU
			}
			job.LineItems = append(job.LineItems, jobDemandLine{ID: line.ID, ProductID: *line.ProductID, SKU: sku, Description: line.Description, Category: "materials", Quantity: line.Quantity, Multiplier: line.Multiplier, RemainingQty: remaining})
		}
		result = append(result, job)
	}
	if c.Param("id") != "" {
		if len(result) == 0 {
			response.NotFound(c, "Accepted job not found")
			return
		}
		response.OK(c, result[0])
		return
	}
	response.OKWithMeta(c, result, pagination.BuildMeta(params, total))
}
