package handlers

import (
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
	return int(line.Quantity*multiplier + 0.999)
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
		return nil, nil
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
// Products quoted on more than one line are left unlinked rather than guessed.
func jobLines(tx *gorm.DB, opportunityID, sourceQuoteID *uuid.UUID) (*uuid.UUID, *models.Quote, map[uuid.UUID]uuid.UUID, error) {
	if opportunityID == nil && sourceQuoteID != nil {
		var source models.Quote
		if err := tx.First(&source, "id = ?", sourceQuoteID).Error; err != nil {
			return nil, nil, nil, err
		}
		opportunityID = source.OpportunityID
	}
	if opportunityID == nil {
		return nil, nil, nil, nil
	}
	var accepted models.Quote
	err := tx.Where("opportunity_id = ? AND status = ?", opportunityID, models.QuoteStatusAccepted).Order("version DESC").First(&accepted).Error
	if err == gorm.ErrRecordNotFound {
		return opportunityID, nil, nil, nil
	}
	if err != nil {
		return nil, nil, nil, err
	}
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
