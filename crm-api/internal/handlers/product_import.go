package handlers

import (
	"strings"
	"time"

	"g4s-crm/api/internal/models"
	"g4s-crm/api/pkg/response"
	v "g4s-crm/api/pkg/validator"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// importRow is one line read from a vendor price list or quotation. The client
// parses the file and maps its columns; the server owns validation, costing and
// the audit trail so an import cannot write prices the product editor rejects.
type importRow struct {
	VendorName       string          `json:"vendorName" validate:"max=255"`
	ManufacturerName string          `json:"manufacturerName" validate:"max=255"`
	CategoryName     string          `json:"categoryName" validate:"max=255"`
	Currency         models.Currency `json:"currency" validate:"omitempty,oneof=SAR USD EUR GBP AED CNY"`
	FXRate           *float64        `json:"fxRate" validate:"omitempty,gt=0,lte=1000000"`
	SKU              string          `json:"sku" validate:"required,max=100"`
	Name             string          `json:"name" validate:"required,max=255"`
	Description      string          `json:"description" validate:"max=20000"`
	UnitCost         float64         `json:"unitCost" validate:"gte=0,lte=1000000000"`
	Qty              int             `json:"qty" validate:"gte=0,lte=1000000"`
	LeadTimeDays     int             `json:"leadTimeDays" validate:"gte=0,lte=3650"`
	VendorSKU        string          `json:"vendorSku" validate:"max=100"`
}

type importRequest struct {
	ManufacturerID *uuid.UUID `json:"manufacturerId"`
	CategoryID     *uuid.UUID `json:"categoryId"`
	// SourceRef names the file the rows came from; it is stored on every price
	// record so a later price question can be traced back to the vendor document.
	SourceRef           string             `json:"sourceRef" validate:"max=255"`
	VendorName          string             `json:"vendorName" validate:"max=255"`
	Source              models.PriceSource `json:"source" validate:"required,oneof=vendor-catalog supplier-quote manual"`
	ProductType         models.ProductType `json:"productType" validate:"required,oneof=import local"`
	OriginCurrency      models.Currency    `json:"originCurrency" validate:"required,oneof=SAR USD EUR GBP AED CNY"`
	FXRate              float64            `json:"fxRate" validate:"gt=0,lte=1000000"`
	FreightPercent      float64            `json:"freightPercent" validate:"gte=0,lte=100"`
	CustomsPercent      float64            `json:"customsPercent" validate:"gte=0,lte=100"`
	ClearancePercent    float64            `json:"clearancePercent" validate:"gte=0,lte=100"`
	TargetMarginPercent float64            `json:"targetMarginPercent" validate:"gte=0,lt=100"`
	// UpdateExisting decides what happens to a SKU that is already in the catalog:
	// refresh its cost and price, or leave it untouched and report it as skipped.
	UpdateExisting bool        `json:"updateExisting"`
	Rows           []importRow `json:"rows" validate:"required,min=1,max=2000,dive"`
}

type importOutcome struct {
	SKU     string  `json:"sku"`
	Status  string  `json:"status"` // created | updated | skipped | failed
	Reason  string  `json:"reason,omitempty"`
	Cost    float64 `json:"landedCostSAR,omitempty"`
	Price   float64 `json:"sellingPrice,omitempty"`
	Product string  `json:"productId,omitempty"`
}

// ImportProducts creates or refreshes catalog products from a parsed vendor file
// in one transaction, recording a vendor entry and a price record per row so the
// product's price history shows where each figure came from.
func (h *ProductHandler) ImportProducts(c *gin.Context) {
	var req importRequest
	if !v.BindAndValidate(c, &req) {
		return
	}

	name, nameErr := vendorName(req.VendorName)
	if nameErr != nil {
		apiError(c, nameErr)
		return
	}
	req.VendorName = name
	outcomes := make([]importOutcome, 0, len(req.Rows))
	var created, updated, skipped, failed int

	err := h.db.Transaction(func(tx *gorm.DB) error {
		if req.ManufacturerID != nil {
			if err := exists(tx, &models.Manufacturer{}, *req.ManufacturerID); err != nil {
				return invalid("Manufacturer does not exist")
			}
		}
		if req.CategoryID != nil {
			var category models.ManufacturerCategory
			if err := tx.First(&category, "id = ?", req.CategoryID).Error; err != nil {
				return invalid("Category does not exist")
			}
			if req.ManufacturerID == nil || category.ManufacturerID != *req.ManufacturerID {
				return invalid("Category must belong to the manufacturer")
			}
		}

		now := time.Now()
		seen := map[string]bool{}
		var firstImported uuid.UUID
		for _, row := range req.Rows {
			// File defaults remain compatible with older clients; explicit row fields win.
			req := req
			sku := strings.TrimSpace(row.SKU)
			name := strings.TrimSpace(row.Name)
			if sku == "" || name == "" {
				outcomes = append(outcomes, importOutcome{SKU: sku, Status: "failed", Reason: "A part number and a description are both required"})
				failed++
				continue
			}
			if seen[strings.ToUpper(sku)] {
				outcomes = append(outcomes, importOutcome{SKU: sku, Status: "skipped", Reason: "The file lists this part number more than once"})
				skipped++
				continue
			}
			seen[strings.ToUpper(sku)] = true

			var item models.Product
			// Serialize existing product updates. New SKUs are also protected by the DB unique constraint.
			err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("sku = ?", sku).First(&item).Error
			isNew := err == gorm.ErrRecordNotFound
			if err != nil && !isNew {
				return err
			}
			if !isNew && !req.UpdateExisting {
				outcomes = append(outcomes, importOutcome{SKU: sku, Status: "skipped", Reason: "Already in the catalog", Product: item.ID.String()})
				skipped++
				continue
			}

			if strings.TrimSpace(row.VendorName) != "" {
				req.VendorName, err = vendorName(row.VendorName)
				if err != nil {
					return invalid(sku + ": " + err.Error())
				}
			}
			if req.VendorName == "" {
				return invalid(sku + ": a vendor name is required (in the row or file defaults)")
			}
			if strings.TrimSpace(row.ManufacturerName) != "" {
				var matches []models.Manufacturer
				if err := tx.Where("lower(trim(name)) = lower(?) OR lower(trim(code)) = lower(?)", strings.TrimSpace(row.ManufacturerName), strings.TrimSpace(row.ManufacturerName)).Find(&matches).Error; err != nil {
					return err
				}
				if len(matches) != 1 {
					return invalid(sku + ": manufacturer must match one existing name or code; add it in Manufacturers first")
				}
				if req.ManufacturerID == nil || *req.ManufacturerID != matches[0].ID {
					req.CategoryID = nil
				}
				req.ManufacturerID = &matches[0].ID
			}
			if strings.TrimSpace(row.CategoryName) != "" {
				manufacturerID := req.ManufacturerID
				if manufacturerID == nil {
					manufacturerID = item.ManufacturerID
				}
				var matches []models.ManufacturerCategory
				if err := tx.Where("manufacturer_id = ? AND lower(trim(name)) = lower(?)", manufacturerID, strings.TrimSpace(row.CategoryName)).Find(&matches).Error; err != nil {
					return err
				}
				if len(matches) != 1 {
					return invalid(sku + ": category must match one existing category for this manufacturer")
				}
				req.ManufacturerID, req.CategoryID = manufacturerID, &matches[0].ID
			}
			if row.Currency != "" && row.Currency != req.OriginCurrency {
				req.OriginCurrency = row.Currency
				if row.Currency != models.CurrencySAR && row.FXRate == nil {
					return invalid(sku + ": an exchange rate to SAR is required for this row's currency")
				}
			}
			if row.FXRate != nil {
				req.FXRate = *row.FXRate
			}
			if req.OriginCurrency == models.CurrencySAR {
				req.ProductType, req.FXRate = models.ProductTypeLocal, 1
			} else {
				req.ProductType = models.ProductTypeImport
			}
			previousCost, previousCurrency, previousLanded := item.UnitCostOrigin, item.OriginCurrency, item.LandedCostSAR
			item.SKU, item.Name = sku, name
			if row.Description != "" {
				item.Description = strings.TrimSpace(row.Description)
			}
			if req.ManufacturerID != nil {
				// A blank category is not a request to erase a known classification.
				// Clear it only when the manufacturer changes, since categories belong to a maker.
				if req.CategoryID != nil || item.ManufacturerID == nil || *item.ManufacturerID != *req.ManufacturerID {
					item.CategoryID = req.CategoryID
				}
				item.ManufacturerID = req.ManufacturerID
			}
			item.ProductType, item.OriginCurrency, item.UnitCostOrigin = req.ProductType, req.OriginCurrency, row.UnitCost
			item.FXRate = req.FXRate
			if req.ProductType == models.ProductTypeLocal {
				item.FXRate = 1
			}
			item.FreightPercent, item.CustomsPercent, item.ClearancePercent = req.FreightPercent, req.CustomsPercent, req.ClearancePercent
			item.TargetMarginPercent, item.SupplierName, item.IsActive = req.TargetMarginPercent, req.VendorName, true
			if row.LeadTimeDays > 0 {
				item.LeadTimeDays = row.LeadTimeDays
			}
			item.RecalculateCosts()
			// Margin is a share of the selling price, matching the product editor.
			item.SellingPrice = models.RoundMoney(item.LandedCostSAR / (1 - req.TargetMarginPercent/100))
			item.RecalculateCosts()

			if err := tx.Omit(clause.Associations).Save(&item).Error; err != nil {
				return err
			}

			vendorSKU := row.VendorSKU
			if vendorSKU == "" {
				vendorSKU = sku
			}
			var vendor models.ProductVendorEntry
			vendorErr := tx.Where("product_id = ? AND lower(trim(vendor_name)) = lower(?)", item.ID, req.VendorName).First(&vendor).Error
			if vendorErr != nil && vendorErr != gorm.ErrRecordNotFound {
				return vendorErr
			}
			vendor.ProductID, vendor.VendorName, vendor.VendorSKU = item.ID, req.VendorName, vendorSKU
			vendor.UnitCost, vendor.Currency, vendor.LeadTimeDays = row.UnitCost, req.OriginCurrency, item.LeadTimeDays
			vendor.LastQuoteDate, vendor.CatalogSource = &now, req.SourceRef
			if vendor.MOQ < 1 {
				vendor.MOQ = 1
			}
			if err := tx.Save(&vendor).Error; err != nil {
				return err
			}

			if isNew || previousCost != item.UnitCostOrigin || previousCurrency != item.OriginCurrency || previousLanded != item.LandedCostSAR {
				qty := row.Qty
				if qty < 1 {
					qty = 1
				}
				record := models.ProductPriceRecord{ProductID: item.ID, Date: now, Source: req.Source, SourceRef: req.SourceRef, VendorName: req.VendorName, UnitCost: row.UnitCost, Currency: req.OriginCurrency, LandingCost: item.LandedCostSAR, Qty: qty}
				if err := tx.Create(&record).Error; err != nil {
					return err
				}
			}

			if firstImported == uuid.Nil {
				firstImported = item.ID
			}
			status := "updated"
			if isNew {
				status = "created"
				created++
			} else {
				updated++
			}
			outcomes = append(outcomes, importOutcome{SKU: sku, Status: status, Cost: item.LandedCostSAR, Price: item.SellingPrice, Product: item.ID.String()})
		}

		if created+updated == 0 {
			return nil // A repeated file with updates disabled is a valid all-skipped import.
		}
		// The log points at the first imported product so the entry links somewhere real.
		return recordActivity(tx, c, "product", firstImported, "imported "+req.VendorName+" price list")
	})
	if err != nil {
		apiError(c, err)
		return
	}

	response.Created(c, gin.H{"created": created, "updated": updated, "skipped": skipped, "failed": failed, "rows": outcomes})
}
