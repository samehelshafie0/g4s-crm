package handlers

import (
	"fmt"
	"math"
	"sort"
	"strconv"

	"g4s-crm/api/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// committedLine loads the customer quote line that a reservation or order line
// claims to fulfil. Only an accepted version is a commitment, and the product
// must be the one the line quotes, so stock cannot be held against a draft or
// against the wrong item.
func committedLine(tx *gorm.DB, lineID uuid.UUID, productID uuid.UUID) (models.QuoteLineItem, models.Quote, error) {
	var line models.QuoteLineItem
	var quote models.Quote
	if err := tx.First(&line, "id = ?", lineID).Error; err != nil {
		return line, quote, invalid("The quote line does not exist")
	}
	if err := tx.Preload("Customer").First(&quote, "id = ?", line.QuoteID).Error; err != nil {
		return line, quote, err
	}
	if quote.Status != models.QuoteStatusAccepted {
		return line, quote, invalid("Stock and orders can only be committed to an accepted quote")
	}
	if line.ProductID == nil || *line.ProductID != productID {
		return line, quote, invalid("The product does not match the quote line it fulfils")
	}
	if (line.IsOptional && !line.IsSelected) || (line.RowType != "" && line.RowType != "item") || (line.FulfilmentSource != nil && *line.FulfilmentSource == "service") {
		return line, quote, invalid("Only selected material lines can be committed")
	}
	return line, quote, nil
}

func quoteRef(quote models.Quote) string {
	return quote.QuoteNumber + " v" + strconv.Itoa(quote.Version)
}

func lineNeed(line models.QuoteLineItem) int {
	multiplier := line.Multiplier
	if multiplier == 0 {
		multiplier = 1
	}
	return int(math.Ceil(line.Quantity * multiplier))
}

// Callers hold the product lock before reading commitments. Draft orders count
// too, so two buyers cannot reserve the same demand before approval.
func lineRemaining(tx *gorm.DB, line models.QuoteLineItem, excludePO uuid.UUID) (int, error) {
	var held, outstanding int
	if err := tx.Model(&models.StockReservation{}).Where("quote_line_item_id = ? AND status IN ('active', 'fulfilled')", line.ID).Select("COALESCE(SUM(qty), 0)").Scan(&held).Error; err != nil {
		return 0, err
	}
	if err := tx.Table("purchase_order_items AS i").Joins("JOIN purchase_orders p ON p.id = i.purchase_order_id").
		Where("i.quote_line_item_id = ? AND i.deleted_at IS NULL AND p.deleted_at IS NULL AND p.status <> 'cancelled' AND p.id <> ?", line.ID, excludePO).
		Select("COALESCE(SUM(GREATEST(i.quantity - i.received_qty, 0)), 0)").Scan(&outstanding).Error; err != nil {
		return 0, err
	}
	return max(0, lineNeed(line)-held-outstanding), nil
}

func validatePOCommitments(tx *gorm.DB, po *models.PurchaseOrder) error {
	ids := map[string]uuid.UUID{}
	for _, item := range po.Items {
		ids[item.ProductID.String()] = item.ProductID
	}
	keys := []string{}
	for key := range ids {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if err := lockProduct(tx, ids[key]); err != nil {
			return err
		}
	}
	used := map[uuid.UUID]int{}
	for _, item := range po.Items {
		if item.QuoteLineItemID == nil {
			if po.SourceQuoteID != nil || po.OpportunityID != nil {
				return invalid("Every job order item must link to an accepted quote line; correct the draft or cancel and recreate the legacy order")
			}
			continue
		}
		line, quote, err := committedLine(tx, *item.QuoteLineItemID, item.ProductID)
		if err != nil {
			return err
		}
		if po.SourceQuoteID != nil && *po.SourceQuoteID != quote.ID {
			return invalid("Order lines must belong to the source quote")
		}
		if po.OpportunityID != nil && (quote.OpportunityID == nil || *po.OpportunityID != *quote.OpportunityID) {
			return invalid("Order lines must belong to the selected opportunity")
		}
		remaining, err := lineRemaining(tx, line, po.ID)
		if err != nil {
			return err
		}
		used[line.ID] += item.Quantity
		if used[line.ID] > remaining {
			return invalid(fmt.Sprintf("Quote line %s has only %d units remaining to order or hold", line.Description, remaining))
		}
	}
	return nil
}

// jobHold works out the hold a good receipt should create for the job its order
// line was bought for: what the quote line still needs and no more. It returns
// nil when the line is not a job commitment, so the goods stay free stock.
func jobHold(tx *gorm.DB, item models.PurchaseOrderItem, received int, location models.WarehouseLocation) (*models.StockReservation, error) {
	if item.QuoteLineItemID == nil {
		return nil, nil
	}
	line, quote, err := committedLine(tx, *item.QuoteLineItemID, item.ProductID)
	if err != nil {
		return nil, err
	}
	var held int
	if err := tx.Model(&models.StockReservation{}).Where("quote_line_item_id = ? AND status IN ('active', 'fulfilled')", line.ID).Select("COALESCE(SUM(qty), 0)").Scan(&held).Error; err != nil {
		return nil, err
	}
	qty := min(received, lineNeed(line)-held)
	if qty <= 0 {
		return nil, nil
	}
	hold := models.StockReservation{ProductID: item.ProductID, WarehouseLocation: location, Qty: qty, Source: models.ReservationSourceQuote, SourceRef: quoteRef(quote), SourceLabel: line.Description, QuoteLineItemID: &line.ID, Status: models.ReservationStatusActive}
	if quote.Customer != nil {
		hold.CustomerName = quote.Customer.CompanyName
	}
	return &hold, nil
}

// jobLines maps each product on a job's accepted quote to the one line that
// quotes it, so a supplier quote converted to an order stays tied to the job.
// Products quoted on more than one line require an explicit line selection.
func jobLines(tx *gorm.DB, opportunityID, sourceQuoteID *uuid.UUID) (*uuid.UUID, *models.Quote, map[uuid.UUID]uuid.UUID, error) {
	if opportunityID == nil && sourceQuoteID == nil {
		return nil, nil, nil, nil
	}
	var accepted models.Quote
	query := tx.Where("status = ?", models.QuoteStatusAccepted)
	if sourceQuoteID != nil {
		query = query.Where("id = ?", sourceQuoteID)
	} else {
		query = query.Where("opportunity_id = ?", opportunityID)
	}
	err := query.Order("version DESC").First(&accepted).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil, nil, invalid("Choose an accepted customer quote before ordering for a job")
	}
	if err != nil {
		return nil, nil, nil, err
	}
	if opportunityID != nil && (accepted.OpportunityID == nil || *opportunityID != *accepted.OpportunityID) {
		return nil, nil, nil, invalid("Source quote belongs to a different opportunity")
	}
	opportunityID = accepted.OpportunityID
	var items []models.QuoteLineItem
	if err := tx.Where("quote_id = ? AND product_id IS NOT NULL AND (row_type = 'item' OR row_type = '')", accepted.ID).Find(&items).Error; err != nil {
		return nil, nil, nil, err
	}
	lines, count := map[uuid.UUID]uuid.UUID{}, map[uuid.UUID]int{}
	for _, item := range items {
		if item.IsOptional && !item.IsSelected {
			continue
		}
		lines[*item.ProductID] = item.ID
		count[*item.ProductID]++
	}
	for product, n := range count {
		if n > 1 {
			delete(lines, product)
		}
	}
	return opportunityID, &accepted, lines, nil
}
