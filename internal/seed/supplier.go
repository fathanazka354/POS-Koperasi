package seed

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

var seedSupplierNames = []string{"PT Unilever Indonesia", "PT Indofood"}

// SupplierSeed memastikan supplier ada; mengembalikan map nama → id.
func SupplierSeed(tx *sqlx.Tx) (map[string]int, error) {
	out := make(map[string]int, len(seedSupplierNames))
	for _, name := range seedSupplierNames {
		var id int
		err := tx.Get(&id, `SELECT id FROM suppliers WHERE name = $1`, name)
		if err == nil {
			out[name] = id
			continue
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		err = tx.QueryRowx(`INSERT INTO suppliers (name) VALUES ($1) RETURNING id`, name).Scan(&id)
		if err != nil {
			return nil, err
		}
		out[name] = id
	}
	return out, nil
}
