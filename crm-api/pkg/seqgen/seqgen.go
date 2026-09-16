package seqgen

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Sequence struct {
	Name    string `gorm:"primaryKey"`
	Prefix  string
	Year    int
	Current int
}

func NextNumber(db *gorm.DB, name string) (string, error) {
	var seq Sequence
	year := time.Now().Year()

	err := db.Transaction(func(tx *gorm.DB) error {
		result := tx.Raw(`
			UPDATE sequences SET current = current + 1
			WHERE name = ? AND year = ?
			RETURNING prefix, year, current
		`, name, year).Scan(&seq)

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			// Start a new year sequence by carrying over prefix from previous
			var prev Sequence
			if err := tx.Where("name = ?", name).Order("year DESC").First(&prev).Error; err != nil {
				return fmt.Errorf("sequence '%s' not found", name)
			}
			seq = Sequence{Name: name, Prefix: prev.Prefix, Year: year, Current: 1}
			if err := tx.Create(&seq).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s-%d-%04d", seq.Prefix, seq.Year, seq.Current), nil
}
