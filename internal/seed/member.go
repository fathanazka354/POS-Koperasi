package seed

import "gorm.io/gorm"

const (
	seedMemberCode  = "MBR001"
	seedMemberName  = "Budi Santoso"
	seedMemberPhone = "081234567890"
)

// MemberSeed memastikan member demo ada.
func MemberSeed(tx *gorm.DB) error {
	return tx.Exec(
		`INSERT INTO members (member_code, full_name, phone) VALUES ($1, $2, $3)
		 ON CONFLICT (member_code) DO NOTHING`,
		seedMemberCode, seedMemberName, seedMemberPhone,
	).Error
}
