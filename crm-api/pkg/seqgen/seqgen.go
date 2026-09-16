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
	return nextNumberForYear(db, name, time.Now().Year())
}

func nextNumberForYear(db *gorm.DB, name string, year int) (string, error) {
	var seq Sequence
	result := db.Raw(`UPDATE sequences
        SET current = CASE WHEN year = ? THEN current + 1 ELSE 1 END, year = ?
        WHERE name = ? AND year <= ? RETURNING prefix, year, current`, year, year, name, year).Scan(&seq)
	if result.Error != nil {
		return "", result.Error
	}
	if result.RowsAffected != 1 {
		return "", fmt.Errorf("sequence %q missing or ahead of current year", name)
	}
	return fmt.Sprintf("%s-%d-%04d", seq.Prefix, seq.Year, seq.Current), nil
}
