package handlers

import (
	"g4s-crm/api/internal/middleware"
	"g4s-crm/api/internal/models"
	"g4s-crm/api/pkg/pagination"
	"g4s-crm/api/pkg/response"
	v "g4s-crm/api/pkg/validator"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
)

type UserHandler struct {
	db *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{db: db}
}

func (h *UserHandler) List(c *gin.Context) {
	params := pagination.GetParams(c, "first_name", "last_name", "email", "role")
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

	if err := query.Count(&total).Error; err != nil {
		apiError(c, err)
		return
	}
	if err := query.Order(params.Sort + " " + params.Order).Scopes(pagination.Paginate(params)).Find(&users).Error; err != nil {
		apiError(c, err)
		return
	}
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
		Password   string            `json:"password" validate:"required,min=8,max=72"`
		Role       models.UserRole   `json:"role" validate:"required,oneof=admin sales_manager sales_executive pre_sales procurement_manager procurement_officer warehouse_manager project_manager viewer"`
		Department models.Department `json:"department" validate:"required,oneof=sales pre-sales technical support marketing management operations"`
		Phone      string            `json:"phone" validate:"max=100"`
		TeamID     *uuid.UUID        `json:"teamId"`
	}
	if !v.BindStrict(c, &req) {
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		apiError(c, err)
		return
	}
	user := &models.User{
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        strings.ToLower(strings.TrimSpace(req.Email)),
		PasswordHash: string(hash),
		Role:         req.Role,
		Department:   req.Department,
		Phone:        req.Phone,
		IsActive:     true,
		TeamID:       req.TeamID,
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := optionalExists(tx, &models.Team{}, req.TeamID); err != nil {
			return err
		}
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		return recordActivity(tx, c, "user", user.ID, "created")
	}); err != nil {
		apiError(c, err)
		return
	}
	response.Created(c, user)
}

const roleRule = "required,oneof=admin sales_manager sales_executive pre_sales procurement_manager procurement_officer warehouse_manager project_manager viewer"

func (h *UserHandler) Update(c *gin.Context)  { h.saveUser(c, false, false) }
func (h *UserHandler) Profile(c *gin.Context) { h.saveUser(c, true, false) }
func (h *UserHandler) Delete(c *gin.Context)  { h.saveUser(c, false, true) }
func (h *UserHandler) saveUser(c *gin.Context, profile, deactivate bool) {
	id := c.Param("id")
	if profile {
		id = middleware.GetCurrentUserID(c).String()
	}
	var user models.User
	err := h.db.Transaction(func(tx *gorm.DB) error {
		// Serialize administrative changes so two requests cannot remove the last admin.
		if err := tx.Exec("SELECT pg_advisory_xact_lock(71024001)").Error; err != nil {
			return err
		}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, "id = ?", id).Error; err != nil {
			return err
		}
		wasAdmin := user.Role == models.RoleAdmin && user.IsActive
		originalTeam := user.TeamID
		if deactivate {
			user.IsActive = false
		} else {
			fields := map[string]string{"firstName": "required,max=100", "lastName": "required,max=100", "email": "required,email,max=255", "phone": "max=100"}
			if !profile {
				fields["role"] = roleRule
				fields["department"] = departmentRule
				fields["teamId"] = ""
				fields["isActive"] = ""
			}
			if !bindFields(c, &user, fields) {
				return invalid("Invalid user")
			}
		}
		user.Email = strings.ToLower(strings.TrimSpace(user.Email))
		if err := optionalExists(tx, &models.Team{}, user.TeamID); err != nil {
			return err
		}
		if originalTeam != nil && (user.TeamID == nil || *originalTeam != *user.TeamID || !user.IsActive) {
			var leaders int64
			if err := tx.Model(&models.Team{}).Where("leader_id = ?", user.ID).Count(&leaders).Error; err != nil {
				return err
			}
			if leaders > 0 {
				return invalid("Choose another team leader before moving or deactivating this user")
			}
		}
		if wasAdmin && (!user.IsActive || user.Role != models.RoleAdmin) {
			var count int64
			if err := tx.Model(&models.User{}).Where("role = ? AND is_active = true AND id <> ?", models.RoleAdmin, user.ID).Count(&count).Error; err != nil {
				return err
			}
			if count == 0 {
				return invalid("At least one active administrator is required")
			}
		}
		if err := tx.Omit(clause.Associations).Save(&user).Error; err != nil {
			return err
		}
		return recordActivity(tx, c, "user", user.ID, "updated")
	})
	if err != nil {
		if !c.Writer.Written() {
			apiError(c, err)
		}
		return
	}
	if deactivate {
		response.NoContent(c)
	} else {
		user.Permissions = middleware.Permissions(user.Role)
		response.OK(c, user)
	}
}
func (h *UserHandler) ResetPassword(c *gin.Context) {
	var req struct {
		Password string `json:"password" validate:"required,min=8,max=72"`
	}
	if !v.BindStrict(c, &req) {
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		apiError(c, err)
		return
	}
	err = h.db.Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, "id = ?", c.Param("id")).Error; err != nil {
			return err
		}
		if err := tx.Model(&user).Update("password_hash", string(hash)).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", user.ID).Delete(&models.RefreshToken{}).Error; err != nil {
			return err
		}
		return recordActivity(tx, c, "user", user.ID, "password reset")
	})
	if err != nil {
		apiError(c, err)
		return
	}
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
