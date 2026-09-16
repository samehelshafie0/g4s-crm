package handlers

import (
	"g4s-crm/api/internal/middleware"
	"g4s-crm/api/internal/models"
	"g4s-crm/api/internal/services"
	"g4s-crm/api/pkg/response"
	v "g4s-crm/api/pkg/validator"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	svc *services.AuthService
}

func NewAuthHandler(svc *services.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type registerRequest struct {
	FirstName  string              `json:"firstName" validate:"required"`
	LastName   string              `json:"lastName" validate:"required"`
	Email      string              `json:"email" validate:"required,email"`
	Password   string              `json:"password" validate:"required,min=8"`
	Role       models.UserRole     `json:"role" validate:"required,oneof=admin sales_manager sales_executive pre_sales procurement_manager procurement_officer warehouse_manager project_manager viewer"`
	Department models.Department   `json:"department"`
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" validate:"required"`
	NewPassword     string `json:"newPassword" validate:"required,min=8"`
}

// Login godoc
// @Summary      Authenticate user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body loginRequest true "Credentials"
// @Success      200  {object}  response.Response
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if !v.BindAndValidate(c, &req) {
		return
	}

	result, err := h.svc.Login(req.Email, req.Password)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.OK(c, result)
}

// Register godoc
// @Summary      Register a new user (admin only)
// @Tags         auth
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body registerRequest true "New user details"
// @Success      201  {object}  response.Response
// @Router       /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if !v.BindAndValidate(c, &req) {
		return
	}

	user, err := h.svc.Register(req.FirstName, req.LastName, req.Email, req.Password, req.Role, req.Department)
	if err != nil {
		response.Conflict(c, "Email already exists or registration failed: "+err.Error())
		return
	}

	response.Created(c, user)
}

// Refresh godoc
// @Summary      Refresh access token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body body refreshRequest true "Refresh token"
// @Success      200  {object}  response.Response
// @Router       /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if !v.BindAndValidate(c, &req) {
		return
	}

	result, err := h.svc.Refresh(req.RefreshToken)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.OK(c, result)
}

// Logout godoc
// @Summary      Logout and revoke refresh token
// @Tags         auth
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body refreshRequest true "Refresh token to revoke"
// @Success      200  {object}  response.Response
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req refreshRequest
	if !v.BindAndValidate(c, &req) {
		return
	}

	if err := h.svc.Logout(req.RefreshToken); err != nil {
		response.InternalError(c, "Failed to logout")
		return
	}

	response.OK(c, gin.H{"message": "Logged out successfully"})
}

// Me godoc
// @Summary      Get current user profile
// @Tags         auth
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  response.Response
// @Router       /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	userID := middleware.GetCurrentUserID(c)
	if userID.String() == "00000000-0000-0000-0000-000000000000" {
		response.Unauthorized(c, "Not authenticated")
		return
	}
	response.OK(c, gin.H{
		"id":    userID,
		"role":  middleware.GetCurrentUserRole(c),
		"email": c.GetString("userEmail"),
	})
}

// ChangePassword godoc
// @Summary      Change current user password
// @Tags         auth
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body changePasswordRequest true "Passwords"
// @Success      200  {object}  response.Response
// @Router       /auth/change-password [patch]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req changePasswordRequest
	if !v.BindAndValidate(c, &req) {
		return
	}

	userID := middleware.GetCurrentUserID(c)
	if err := h.svc.ChangePassword(userID, req.CurrentPassword, req.NewPassword); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.OK(c, gin.H{"message": "Password changed successfully"})
}
