package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"g4s-crm/api/internal/middleware"
	"g4s-crm/api/internal/models"
	"g4s-crm/api/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"io"
	"reflect"
	"strings"
	"time"
)

type inputError string

func (e inputError) Error() string { return string(e) }
func invalid(message string) error { return inputError(message) }

func apiError(c *gin.Context, err error) {
	var input inputError
	var sqlState interface{ SQLState() string }
	switch {
	case errors.As(err, &input):
		response.BadRequest(c, string(input))
	case errors.Is(err, gorm.ErrRecordNotFound):
		response.NotFound(c, "Record not found")
	case errors.As(err, &sqlState):
		switch sqlState.SQLState() {
		case "23505":
			response.Conflict(c, "A record with this identifier already exists")
		case "23503":
			response.Conflict(c, "A related record is missing or still in use")
		case "23514", "22P02", "22003":
			response.BadRequest(c, "Invalid field value")
		default:
			response.InternalError(c, "Database operation failed")
		}
	default:
		response.InternalError(c, "Operation failed")
	}
}

// bindFields merges only explicit editable fields into a typed object. Existing
// values survive PATCH; protected fields, unknown fields and null scalars fail.
func bindFields(c *gin.Context, target any, rules map[string]string) bool {
	var fields map[string]json.RawMessage
	decoder := json.NewDecoder(c.Request.Body)
	if err := decoder.Decode(&fields); err != nil || fields == nil {
		response.BadRequest(c, "Expected a JSON object")
		return false
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		response.BadRequest(c, "Expected one JSON object")
		return false
	}
	rv := reflect.ValueOf(target).Elem()
	rt := rv.Type()
	indexes := map[string]int{}
	for i := 0; i < rt.NumField(); i++ {
		indexes[strings.Split(rt.Field(i).Tag.Get("json"), ",")[0]] = i
	}
	for name, raw := range fields {
		if _, ok := rules[name]; !ok {
			response.BadRequest(c, "Field is not editable: "+name)
			return false
		}
		index, ok := indexes[name]
		if !ok {
			response.InternalError(c, "Invalid field configuration")
			return false
		}
		field := rv.Field(index)
		if string(raw) == "null" && field.Kind() != reflect.Pointer && field.Kind() != reflect.Slice {
			response.BadRequest(c, name+" cannot be null")
			return false
		}
		if field.Type() == reflect.TypeOf(time.Time{}) || field.Type() == reflect.TypeOf((*time.Time)(nil)) {
			var value *string
			if err := json.Unmarshal(raw, &value); err != nil {
				response.BadRequest(c, "Invalid date for "+name)
				return false
			}
			date, err := parseDate(value)
			if err != nil {
				apiError(c, err)
				return false
			}
			if field.Kind() == reflect.Pointer {
				field.Set(reflect.ValueOf(date))
			} else if date != nil {
				field.Set(reflect.ValueOf(*date))
			} else {
				response.BadRequest(c, name+" is required")
				return false
			}
			continue
		}
		if err := json.Unmarshal(raw, field.Addr().Interface()); err != nil {
			response.BadRequest(c, "Invalid value for "+name)
			return false
		}
	}
	engine := binding.Validator.Engine().(*validator.Validate)
	for name, rule := range rules {
		if rule == "" {
			continue
		}
		index, ok := indexes[name]
		if !ok {
			continue
		}
		if err := engine.Var(rv.Field(index).Interface(), rule); err != nil {
			response.BadRequest(c, "Invalid "+name+": "+rule)
			return false
		}
	}
	return true
}

const currencyRule = "required,oneof=SAR USD EUR GBP AED CNY"
const moneyRule = "gte=0,lte=100000000000"
const percentRule = "gte=0,lte=100"
const departmentRule = "required,oneof=sales pre-sales technical support marketing management operations"

func parseDate(value *string) (*time.Time, error) {
	if value == nil || *value == "" {
		return nil, nil
	}
	date, err := time.Parse("2006-01-02", *value)
	if err != nil {
		date, err = time.Parse(time.RFC3339, *value)
	}
	if err != nil {
		return nil, invalid("Invalid date; use YYYY-MM-DD")
	}
	return &date, nil
}
func exists(tx *gorm.DB, model any, id uuid.UUID) error {
	if id == uuid.Nil {
		return invalid("A related record ID is required")
	}
	var count int64
	if err := tx.Model(model).Where("id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count != 1 {
		return invalid("Related record does not exist")
	}
	return nil
}
func optionalExists(tx *gorm.DB, model any, id *uuid.UUID) error {
	if id == nil {
		return nil
	}
	return exists(tx, model, *id)
}
func recordActivity(tx *gorm.DB, c *gin.Context, entity string, id uuid.UUID, action string) error {
	userID := middleware.GetCurrentUserID(c)
	return tx.Create(&models.ActivityLog{UserID: &userID, EntityType: entity, EntityID: id, Action: action, Description: fmt.Sprintf("%s %s", action, entity)}).Error
}
func deleteRecord(c *gin.Context, db *gorm.DB, model any) {
	result := db.Delete(model, "id = ?", c.Param("id"))
	if result.Error != nil {
		apiError(c, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		response.NotFound(c, "Record not found")
		return
	}
	response.NoContent(c)
}
