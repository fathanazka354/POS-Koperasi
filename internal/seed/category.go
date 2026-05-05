package seed

import (
	"gorm.io/gorm"
)

var seedCategoryNames = []string{"Makanan", "Minuman", "Kebersihan"}

// CategorySeed memastikan kategori ada; mengembalikan map nama → id.
func CategorySeed(tx *gorm.DB) (map[string]int, error) {
	out := make(map[string]int, len(seedCategoryNames))
	for _, name := range seedCategoryNames {
		var id int
		if err := tx.Raw(`SELECT id FROM categories WHERE name = $1`, name).Scan(&id).Error; err != nil {
			return nil, err
		}
		if id != 0 {
			out[name] = id
			continue
		}

		if err := tx.Raw(`INSERT INTO categories (name) VALUES ($1) RETURNING id`, name).Scan(&id).Error; err != nil {
			return nil, err
		}
		out[name] = id
	}
	return out, nil
}
