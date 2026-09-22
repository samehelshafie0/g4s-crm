package services

import (
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"strings"
	"time"

	"g4s-crm/api/internal/config"
	"g4s-crm/api/internal/middleware"
	"g4s-crm/api/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AuthService struct {
	db         *gorm.DB
	cfg        *config.JWTConfig
	privateKey *rsa.PrivateKey
}

func NewAuthService(db *gorm.DB, cfg *config.JWTConfig) (*AuthService, error) {
	keyBytes, err := os.ReadFile(cfg.PrivateKeyPath)
	if err != nil {
		return nil, err
	}
	key, err := jwt.ParseRSAPrivateKeyFromPEM(keyBytes)
	if err != nil {
		return nil, err
	}
	return &AuthService{db: db, cfg: cfg, privateKey: key}, nil
}

type LoginResponse struct {
	User         *models.User `json:"user"`
	AccessToken  string       `json:"accessToken"`
	RefreshToken string       `json:"refreshToken"`
}

func (s *AuthService) Login(email, password string) (*LoginResponse, error) {
	var user models.User
	if err := s.db.Where("email = ? AND is_active = true", strings.ToLower(strings.TrimSpace(email))).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	accessToken, err := s.generateAccessToken(&user)
	if err != nil {
		return nil, err
	}

	refreshToken, rawToken, err := s.generateRefreshToken(&user)
	if err != nil {
		return nil, err
	}

	if err := s.db.Create(refreshToken).Error; err != nil {
		return nil, err
	}

	now := time.Now()
	s.db.Model(&user).Update("last_login_at", now)

	user.Permissions = middleware.Permissions(user.Role)
	return &LoginResponse{
		User:         &user,
		AccessToken:  accessToken,
		RefreshToken: rawToken,
	}, nil
}

func (s *AuthService) Refresh(rawRefreshToken string) (*LoginResponse, error) {
	var result *LoginResponse
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var stored models.RefreshToken
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("token_hash = ? AND revoked_at IS NULL AND expires_at > ?", hashToken(rawRefreshToken), time.Now()).First(&stored).Error; err != nil {
			return errors.New("invalid or expired refresh token")
		}
		var user models.User
		if err := tx.Where("id = ? AND is_active = true", stored.UserID).First(&user).Error; err != nil {
			return errors.New("user not found or inactive")
		}
		access, err := s.generateAccessToken(&user)
		if err != nil {
			return err
		}
		refresh, raw, err := s.generateRefreshToken(&user)
		if err != nil {
			return err
		}
		if err := tx.Model(&stored).Update("revoked_at", time.Now()).Error; err != nil {
			return err
		}
		if err := tx.Create(refresh).Error; err != nil {
			return err
		}
		user.Permissions = middleware.Permissions(user.Role)
		result = &LoginResponse{User: &user, AccessToken: access, RefreshToken: raw}
		return nil
	})
	return result, err
}

func (s *AuthService) Logout(rawRefreshToken string) error {
	hash := hashToken(rawRefreshToken)
	now := time.Now()
	return s.db.Model(&models.RefreshToken{}).
		Where("token_hash = ? AND revoked_at IS NULL", hash).
		Update("revoked_at", now).Error
}

func (s *AuthService) Register(firstName, lastName, email, password string, role models.UserRole, dept models.Department) (*models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		FirstName:    firstName,
		LastName:     lastName,
		Email:        email,
		PasswordHash: string(hash),
		Role:         role,
		Department:   dept,
		IsActive:     true,
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthService) ChangePassword(userID uuid.UUID, currentPassword, newPassword string) error {
	var user models.User
	if err := s.db.First(&user, "id = ?", userID).Error; err != nil {
		return errors.New("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return errors.New("current password is incorrect")
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return err
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&user).Update("password_hash", string(newHash)).Error; err != nil {
			return err
		}
		return tx.Model(&models.RefreshToken{}).Where("user_id = ? AND revoked_at IS NULL", userID).Update("revoked_at", time.Now()).Error
	})
}

func (s *AuthService) generateAccessToken(user *models.User) (string, error) {
	claims := middleware.Claims{
		UserID: user.ID,
		Role:   user.Role,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.AccessExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID.String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(s.privateKey)
}

func (s *AuthService) generateRefreshToken(user *models.User) (*models.RefreshToken, string, error) {
	rawToken := uuid.New().String() + uuid.New().String()
	hash := hashToken(rawToken)

	rt := &models.RefreshToken{
		UserID:    user.ID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(s.cfg.RefreshExpiry),
	}
	return rt, rawToken, nil
}

func hashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
