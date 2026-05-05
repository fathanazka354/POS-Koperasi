package seed

import (
	"gorm.io/gorm"
)

const (
	seedBranchName = "Koperasi Pusat"
	seedBranchCity = "Yogyakarta"
)

// BranchSeed memastikan cabang awal ada; mengembalikan id cabang.
func BranchSeed(tx *gorm.DB) (int, error) {
	var id int
	if err := tx.Raw(`SELECT id FROM branches WHERE name = $1 LIMIT 1`, seedBranchName).Scan(&id).Error; err != nil {
		return 0, err
	}
	if id != 0 {
		return id, nil
	}
	if err := tx.Raw(
		`INSERT INTO branches (name, city) VALUES ($1, $2) RETURNING id`,
		seedBranchName, seedBranchCity,
	).Scan(&id).Error; err != nil {
		return 0, err
	}
	return id, nil
}
