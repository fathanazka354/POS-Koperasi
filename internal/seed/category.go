package seed

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

var seedCategoryNames = []string{"Makanan", "Minuman", "Kebersihan"}

// CategorySeed memastikan kategori ada; mengembalikan map nama → id.
func CategorySeed(tx *sqlx.Tx) (map[string]int, error) {
	out := make(map[string]int, len(seedCategoryNames))
	for _, name := range seedCategoryNames {
		var id int
		err := tx.Get(&id, `SELECT id FROM categories WHERE name = $1`, name)
		if err == nil {
			out[name] = id
			continue
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}

		err = tx.QueryRowx(`INSERT INTO categories (name) VALUES ($1) RETURNING id`, name).Scan(&id)
		if err != nil {
			return nil, err
		}
		out[name] = id
	}
	return out, nil
}
