package seed

import (
	"gorm.io/gorm"
)

var seedSupplierNames = []string{"PT Unilever Indonesia", "PT Indofood"}

// SupplierSeed memastikan supplier ada; mengembalikan map nama → id.
func SupplierSeed(tx *gorm.DB) (map[string]int, error) {
	out := make(map[string]int, len(seedSupplierNames))
	for _, name := range seedSupplierNames {
		var id int
		if err := tx.Raw(`SELECT id FROM suppliers WHERE name = $1`, name).Scan(&id).Error; err != nil {
			return nil, err
		}
		if id != 0 {
			out[name] = id
			continue
		}
		if err := tx.Raw(`INSERT INTO suppliers (name) VALUES ($1) RETURNING id`, name).Scan(&id).Error; err != nil {
			return nil, err
		}
		out[name] = id
	}
	return out, nil
}
