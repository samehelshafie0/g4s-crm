package handlers

import (
	"g4s-crm/api/internal/models"
	"g4s-crm/api/pkg/pagination"
	"g4s-crm/api/pkg/response"
	v "g4s-crm/api/pkg/validator"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserHandler struct {
	db *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{db: db}
}

func (h *UserHandler) List(c *gin.Context) {
	params := pagination.GetParams(c)
	role := c.Query("role")
	dept := c.Query("department")
	q := c.Query("q")

	var users []models.User
	var total int64

	query := h.db.Model(&models.User{}).Omit("password_hash")
	if role != "" {
		query = query.Where("role = ?", role)
	}
	if dept != "" {
		query = query.Where("department = ?", dept)
	}
	if q != "" {
		query = query.Where("first_name ILIKE ? OR last_name ILIKE ? OR email ILIKE ?", "%"+q+"%", "%"+q+"%", "%"+q+"%")
	}

	query.Count(&total)
	query.Order(params.Sort + " " + params.Order).Scopes(pagination.Paginate(params)).Find(&users)
	response.OKWithMeta(c, users, pagination.BuildMeta(params, total))
}

func (h *UserHandler) Get(c *gin.Context) {
	var user models.User
	if err := h.db.First(&user, "id = ?", c.Param("id")).Error; err != nil {
		response.NotFound(c, "User not found")
		return
	}
	response.OK(c, user)
}

func (h *UserHandler) Create(c *gin.Context) {
	var req struct {
		FirstName  string            `json:"firstName" validate:"required"`
		LastName   string            `json:"lastName" validate:"required"`
		Email      string            `json:"email" validate:"required,email"`
		Password   string            `json:"password" validate:"required,min=8"`
		Role       models.UserRole   `json:"role" validate:"required"`
		Department models.Department `json:"department"`
		Phone      string            `json:"phone"`
	}
	if !v.BindAndValidate(c, &req) {
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	user := &models.User{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		PasswordHash: string(hash),
		Role:         req.Role,
		Department:   req.Department,
		Phone:        req.Phone,
		IsActive:     true,
	}

	if err := h.db.Create(user).Error; err != nil {
		response.Conflict(c, "Email already exists")
		return
	}
	response.Created(c, user)
}

func (h *UserHandler) Update(c *gin.Context) {
	var user models.User
	if err := h.db.First(&user, "id = ?", c.Param("id")).Error; err != nil {
		response.NotFound(c, "User not found")
		return
	}
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	delete(req, "password_hash")
	delete(req, "password")
	h.db.Model(&user).Updates(req)
	response.OK(c, user)
}

func (h *UserHandler) Delete(c *gin.Context) {
	h.db.Model(&models.User{}).Where("id = ?", c.Param("id")).Update("is_active", false)
	response.NoContent(c)
}

func (h *UserHandler) Lookup(c *gin.Context) {
	role := c.Query("role")
	dept := c.Query("department")

	var users []struct {
		ID        string `json:"id"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
		Email     string `json:"email"`
		Role      string `json:"role"`
	}

	query := h.db.Model(&models.User{}).Where("is_active = true")
	if role != "" {
		query = query.Where("role = ?", role)
	}
	if dept != "" {
		query = query.Where("department = ?", dept)
	}
	query.Select("id, first_name, last_name, email, role").Find(&users)
	response.OK(c, users)
}
