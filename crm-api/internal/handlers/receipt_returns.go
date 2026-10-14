package handlers

import (
	"strings"
	"time"

	"g4s-crm/api/internal/models"
	"g4s-crm/api/pkg/response"
	v "g4s-crm/api/pkg/validator"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ReturnDamaged records dispatch of quarantined units. They never entered
// usable inventory, so returning them must not subtract stock a second time.
func (h *ProcurementHandler) ReturnDamaged(c *gin.Context) {
	var req struct {
		Reference string `json:"reference" validate:"required,max=1000"`
	}
	if !v.BindStrict(c, &req) {
		return
	}
	if strings.TrimSpace(req.Reference) == "" {
		response.BadRequest(c, "A return dispatch reference is required")
		return
	}
	var item models.GoodsReceiptItem
	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&item, "id = ? AND goods_receipt_id = ?", c.Param("itemId"), c.Param("id")).Error; err != nil {
			return err
		}
		if item.ReturnStatus != "pending" || item.DamagedQty == 0 {
			return invalid("This receipt item has no pending damaged return")
		}
		now := time.Now()
		item.ReturnStatus, item.ReturnReference, item.ReturnedAt = "returned", strings.TrimSpace(req.Reference), &now
		if err := tx.Omit(clause.Associations).Save(&item).Error; err != nil {
			return err
		}
		return recordActivity(tx, c, "goods-receipt", item.GoodsReceiptID, "damaged goods returned: "+item.ReturnReference)
	})
	if err != nil {
		apiError(c, err)
		return
	}
	response.OK(c, item)
}
