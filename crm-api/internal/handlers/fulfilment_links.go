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
