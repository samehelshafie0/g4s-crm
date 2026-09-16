package middleware

import (
	"net/http"

	"g4s-crm/api/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Error().Interface("panic", err).
					Str("path", c.Request.URL.Path).
					Msg("panic recovered")
				c.JSON(http.StatusInternalServerError, response.Error("INTERNAL_ERROR", "An unexpected error occurred"))
				c.Abort()
			}
		}()
		c.Next()
	}
}
