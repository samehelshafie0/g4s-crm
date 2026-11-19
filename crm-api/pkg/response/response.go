package response

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
	Error   *ErrorBody  `json:"error,omitempty"`
}

type ErrorBody struct {
	Code    string       `json:"code"`
	Message string       `json:"message"`
	Details []FieldError `json:"details,omitempty"`
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Success: true, Data: visibleData(c, data)})
}

func OKWithMeta(c *gin.Context, data interface{}, meta interface{}) {
	c.JSON(http.StatusOK, Response{Success: true, Data: visibleData(c, data), Meta: meta})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{Success: true, Data: visibleData(c, data)})
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

func BadRequest(c *gin.Context, message string, details ...FieldError) {
	c.JSON(http.StatusBadRequest, Response{
		Success: false,
		Error: &ErrorBody{
			Code:    "VALIDATION_ERROR",
			Message: message,
			Details: details,
		},
	})
}

func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{
		Success: false,
		Error:   &ErrorBody{Code: "UNAUTHORIZED", Message: message},
	})
}

func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Response{
		Success: false,
		Error:   &ErrorBody{Code: "FORBIDDEN", Message: message},
	})
}

func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, Response{
		Success: false,
		Error:   &ErrorBody{Code: "NOT_FOUND", Message: message},
	})
}

func Conflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, Response{
		Success: false,
		Error:   &ErrorBody{Code: "CONFLICT", Message: message},
	})
}

func UnprocessableEntity(c *gin.Context, message string) {
	c.JSON(http.StatusUnprocessableEntity, Response{
		Success: false,
		Error:   &ErrorBody{Code: "BUSINESS_RULE_VIOLATION", Message: message},
	})
}

func TooManyRequests(c *gin.Context) {
	c.JSON(http.StatusTooManyRequests, Response{
		Success: false,
		Error:   &ErrorBody{Code: "RATE_LIMITED", Message: "Too many requests. Please slow down."},
	})
}

func InternalError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, Response{
		Success: false,
		Error:   &ErrorBody{Code: "INTERNAL_ERROR", Message: message},
	})
}

func Error(code string, message string) Response {
	return Response{
		Success: false,
		Error:   &ErrorBody{Code: code, Message: message},
	}
}

func nonNilSlice(data any) any {
	if data == nil {
		return nil
	}
	value := reflect.ValueOf(data)
	if value.Kind() == reflect.Slice && value.IsNil() {
		return reflect.MakeSlice(value.Type(), 0, 0).Interface()
	}
	return data
}

func visibleData(c *gin.Context, data any) any {
	data = nonNilSlice(data)
	_, scopeTeams := c.Get("quoteTeams")
	if !c.GetBool("hideQuoteMargins") && !c.GetBool("hideQuotes") && !scopeTeams {
		return data
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return nil
	}
	var out any
	if json.Unmarshal(raw, &out) != nil {
		return nil
	}
	if c.GetBool("hideQuoteMargins") {
		// Catalog is the quotation builder's price source. Other modules retain
		// their own permissions; nested quotation objects are still redacted.
		path := c.Request.URL.Path
		quoteResponse := strings.Contains(path, "/quotes") || strings.Contains(path, "/dashboard/") || strings.HasSuffix(path, "/price-history") && strings.Contains(path, "/customers/") || strings.HasSuffix(path, "/catalog")
		if quoteResponse {
			redactMargins(out)
		} else {
			redactNestedQuoteMargins(out)
		}
	}
	return redactQuotes(c, out)
}
func redactMargins(value any) {
	switch item := value.(type) {
	case []any:
		for _, child := range item {
			redactMargins(child)
		}
	case map[string]any:
		for _, key := range []string{"unitCost", "totalCost", "costTotal", "marginAmount", "marginPercent", "estimatedCost", "estimatedMargin", "unitCostOrigin", "costInSAR", "landedCostSAR", "targetMarginPercent", "landingCost", "monthlyCost", "cost", "grossProfit", "averageMargin", "margin"} {
			delete(item, key)
		}
		for _, child := range item {
			redactMargins(child)
		}
	}
}

func redactNestedQuoteMargins(value any) {
	switch item := value.(type) {
	case []any:
		for _, child := range item {
			redactNestedQuoteMargins(child)
		}
	case map[string]any:
		if _, quote := item["quoteNumber"]; quote {
			redactMargins(item)
			return
		}
		for _, child := range item {
			redactNestedQuoteMargins(child)
		}
	}
}

func redactQuotes(c *gin.Context, value any) any {
	switch item := value.(type) {
	case []any:
		out := []any{}
		for _, child := range item {
			if visible := redactQuotes(c, child); visible != nil {
				out = append(out, visible)
			}
		}
		return out
	case map[string]any:
		if _, quote := item["quoteNumber"]; quote && !c.GetBool("operationalQuoteResponse") {
			if c.GetBool("hideQuotes") {
				return nil
			}
			if team, ok := item["teamId"].(string); ok && team != "" {
				if raw, scoped := c.Get("quoteTeams"); scoped {
					if allowed, ok := raw.(map[string]bool); !ok || !allowed[team] {
						return nil
					}
				}
			}
		}
		if c.GetBool("hideQuotes") {
			delete(item, "quoteIds")
			delete(item, "currentQuote")
		}
		for key, child := range item {
			item[key] = redactQuotes(c, child)
		}
	}
	return value
}
