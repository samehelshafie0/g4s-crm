package validator

import (
	"encoding/json"
	"g4s-crm/api/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"io"
)

// BindStrict rejects unknown/immutable fields and trailing JSON before any write.
func BindStrict(c *gin.Context, value interface{}) bool {
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return false
	}
	var extra interface{}
	if err := decoder.Decode(&extra); err != io.EOF {
		response.BadRequest(c, "Request must contain one JSON object")
		return false
	}
	if err := binding.Validator.ValidateStruct(value); err != nil {
		response.BadRequest(c, "Validation failed: "+err.Error())
		return false
	}
	return true
}
