package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"g4s-crm/api/internal/models"
	"g4s-crm/api/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Live rates come from the ExchangeRate-API open endpoint (daily, no key,
// attribution required) unless FX_PROVIDER_URL points elsewhere. The provider
// must answer with the open.er-api.com shape: {"result":"success","base_code":
// "SAR","time_last_update_unix":N,"rates":{"USD":0.2667,...}} where each rate is
// units of that currency per one SAR.
const defaultFXProvider = "https://open.er-api.com/v6/latest/SAR"

var liveCurrencies = []models.Currency{models.CurrencyUSD, models.CurrencyEUR, models.CurrencyGBP, models.CurrencyAED, models.CurrencyCNY}

type liveRates struct {
	Result         string             `json:"result"`
	BaseCode       string             `json:"base_code"`
	LastUpdateUnix int64              `json:"time_last_update_unix"`
	Provider       string             `json:"provider"`
	Rates          map[string]float64 `json:"rates"`
}

func fetchLiveRates(ctx context.Context) (*liveRates, error) {
	url := os.Getenv("FX_PROVIDER_URL")
	if url == "" {
		url = defaultFXProvider
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rate provider returned HTTP %d", res.StatusCode)
	}
	var payload liveRates
	if err := json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&payload); err != nil {
		return nil, err
	}
	if payload.Result != "success" || payload.BaseCode != "SAR" || len(payload.Rates) == 0 {
		return nil, errors.New("rate provider response was not a SAR-based rate table")
	}
	return &payload, nil
}

// Refresh pulls today's rates, updates each supported pair and appends a history
// row when the rate or effective date changed. Existing pairs are locked so a
// concurrent manual edit cannot interleave with the refresh.
func (h *ExchangeRateHandler) Refresh(c *gin.Context) {
	live, err := fetchLiveRates(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "error": gin.H{"message": "Live rates are unavailable right now: " + err.Error()}})
		return
	}
	effective := time.Unix(live.LastUpdateUnix, 0).UTC()
	if live.LastUpdateUnix == 0 {
		effective = time.Now().UTC()
	}
	effective = effective.Truncate(24 * time.Hour)
	updated := []models.ExchangeRate{}
	err = h.db.Transaction(func(tx *gorm.DB) error {
		for _, currency := range liveCurrencies {
			perSAR, ok := live.Rates[string(currency)]
			if !ok || perSAR <= 0 {
				continue
			}
			rate := models.RoundRate(1 / perSAR)
			var item models.ExchangeRate
			err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("from_currency = ? AND to_currency = 'SAR'", currency).First(&item).Error
			switch {
			case err == gorm.ErrRecordNotFound:
				item = models.ExchangeRate{FromCurrency: currency, ToCurrency: "SAR", CurrentRate: rate, EffectiveDate: effective}
				if err := tx.Create(&item).Error; err != nil {
					return err
				}
			case err != nil:
				return err
			case item.CurrentRate == rate && !item.EffectiveDate.Before(effective):
				continue // already current for this provider date
			default:
				item.CurrentRate, item.EffectiveDate = rate, effective
				if err := tx.Model(&item).Select("current_rate", "effective_date").Updates(&item).Error; err != nil {
					return err
				}
			}
			if err := tx.Create(&models.ExchangeRateHistory{ExchangeRateID: item.ID, Rate: rate, EffectiveDate: effective}).Error; err != nil {
				return err
			}
			updated = append(updated, item)
		}
		return nil
	})
	if err != nil {
		apiError(c, err)
		return
	}
	provider := live.Provider
	if provider == "" {
		provider = "open.er-api.com"
	}
	response.OK(c, gin.H{"provider": provider, "effectiveDate": effective.Format("2006-01-02"), "updated": updated})
}
