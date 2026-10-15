package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS admits browser calls only from the listed origins. With none listed the
// API is same-origin only: nginx serves the app and the API on one address.
// Sessions travel in the Authorization header, so cookies are never shared.
func CORS(allowedOrigins []string) gin.HandlerFunc {
	if len(allowedOrigins) == 0 {
		return func(c *gin.Context) { c.Next() }
	}
	return cors.New(cors.Config{
		AllowOrigins:  allowedOrigins,
		AllowMethods:  []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:  []string{"Origin", "Content-Type", "Authorization", "Accept", "X-Requested-With"},
		ExposeHeaders: []string{"Content-Length", "Content-Disposition"},
		MaxAge:        12 * time.Hour,
	})
}
