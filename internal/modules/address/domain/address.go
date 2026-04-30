package domain

import "time"

type Address struct {
	ID          int       `db:"id"           json:"id"`
	MemberID    int       `db:"member_id"    json:"member_id"`
	Label       string    `db:"label"        json:"label"`
	Recipient   string    `db:"recipient"    json:"recipient"`
	Phone       string    `db:"phone"        json:"phone"`
	AddressLine string    `db:"address_line" json:"address_line"`
	City        string    `db:"city"         json:"city"`
	Province    string    `db:"province"     json:"province"`
	PostalCode  string    `db:"postal_code"  json:"postal_code"`
	IsDefault   bool      `db:"is_default"   json:"is_default"`
	CreatedAt   time.Time `db:"created_at"   json:"created_at"`
}
