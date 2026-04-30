package seed

import "github.com/jmoiron/sqlx"

const (
	seedMemberCode  = "MBR001"
	seedMemberName  = "Budi Santoso"
	seedMemberPhone = "081234567890"
)

// MemberSeed memastikan member demo ada.
func MemberSeed(tx *sqlx.Tx) error {
	_, err := tx.Exec(
		`INSERT INTO members (member_code, full_name, phone) VALUES ($1, $2, $3)
		 ON CONFLICT (member_code) DO NOTHING`,
		seedMemberCode, seedMemberName, seedMemberPhone,
	)
	return err
}
