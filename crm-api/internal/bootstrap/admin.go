package bootstrap

import (
	"errors"
	"g4s-crm/api/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/mail"
	"strings"
)

// Admin creates the first administrator only; it never resets an existing user.
func Admin(db *gorm.DB, email, password, firstName, lastName string) error {
	email = strings.ToLower(strings.TrimSpace(email))
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email || len(password) < 12 || len(password) > 72 || strings.TrimSpace(firstName) == "" || strings.TrimSpace(lastName) == "" {
		return errors.New("provide ADMIN_EMAIL, ADMIN_PASSWORD (12–72 bytes), ADMIN_FIRST_NAME and ADMIN_LAST_NAME")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(473246532)").Error; err != nil {
			return err
		}
		var count int64
		if err := tx.Model(&models.User{}).Where("role = ? AND is_active = true", models.RoleAdmin).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New("an active administrator already exists; bootstrap refused")
		}
		return tx.Create(&models.User{Email: email, PasswordHash: string(hash), FirstName: strings.TrimSpace(firstName), LastName: strings.TrimSpace(lastName), Role: models.RoleAdmin, Department: models.DeptManagement, IsActive: true}).Error
	})
}
