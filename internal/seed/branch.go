package seed

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

const (
	seedBranchName = "Koperasi Pusat"
	seedBranchCity = "Yogyakarta"
)

// BranchSeed memastikan cabang awal ada; mengembalikan id cabang.
func BranchSeed(tx *sqlx.Tx) (int, error) {
	var id int
	err := tx.Get(&id, `SELECT id FROM branches WHERE name = $1 LIMIT 1`, seedBranchName)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	err = tx.QueryRowx(
		`INSERT INTO branches (name, city) VALUES ($1, $2) RETURNING id`,
		seedBranchName, seedBranchCity,
	).Scan(&id)
	return id, err
}
