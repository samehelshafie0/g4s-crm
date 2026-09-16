package validator

import (
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"g4s-crm/api/pkg/response"
)

func Setup() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// Use JSON field names in error messages
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})
	}
}

// BindAndValidate binds JSON body and returns validation field errors
func BindAndValidate(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		var details []response.FieldError

		if ve, ok := err.(validator.ValidationErrors); ok {
			for _, e := range ve {
				details = append(details, response.FieldError{
					Field:   e.Field(),
					Message: buildMessage(e),
					Code:    e.Tag(),
				})
			}
			response.BadRequest(c, "Validation failed", details...)
		} else {
			response.BadRequest(c, err.Error())
		}
		return false
	}
	return true
}

func buildMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return e.Field() + " is required"
	case "email":
		return e.Field() + " must be a valid email address"
	case "min":
		return e.Field() + " must be at least " + e.Param() + " characters"
	case "max":
		return e.Field() + " must be at most " + e.Param() + " characters"
	case "oneof":
		return e.Field() + " must be one of: " + e.Param()
	case "uuid4":
		return e.Field() + " must be a valid UUID"
	case "gt":
		return e.Field() + " must be greater than " + e.Param()
	case "gte":
		return e.Field() + " must be greater than or equal to " + e.Param()
	case "lte":
		return e.Field() + " must be less than or equal to " + e.Param()
	default:
		return e.Field() + " failed " + e.Tag() + " validation"
	}
}
