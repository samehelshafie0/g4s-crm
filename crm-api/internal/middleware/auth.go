package middleware

import (
	"crypto/rsa"
	"os"
	"strings"

	"g4s-crm/api/internal/models"
	"g4s-crm/api/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID uuid.UUID        `json:"userId"`
	Role   models.UserRole  `json:"role"`
	Email  string           `json:"email"`
	jwt.RegisteredClaims
}

var publicKey *rsa.PublicKey

func LoadPublicKey(path string) error {
	keyBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	key, err := jwt.ParseRSAPublicKeyFromPEM(keyBytes)
	if err != nil {
		return err
	}
	publicKey = key
	return nil
}

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			response.Unauthorized(c, "Authorization header is required")
			c.Abort()
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Unauthorized(c, "Authorization header format must be: Bearer <token>")
			c.Abort()
			return
		}

		tokenStr := parts[1]
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return publicKey, nil
		})

		if err != nil || !token.Valid {
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("userRole", claims.Role)
		c.Set("userEmail", claims.Email)
		c.Set("claims", claims)
		c.Next()
	}
}

func GetCurrentUserID(c *gin.Context) uuid.UUID {
	id, _ := c.Get("userID")
	if uid, ok := id.(uuid.UUID); ok {
		return uid
	}
	return uuid.Nil
}

func GetCurrentUserRole(c *gin.Context) models.UserRole {
	role, _ := c.Get("userRole")
	if r, ok := role.(models.UserRole); ok {
		return r
	}
	return ""
}
