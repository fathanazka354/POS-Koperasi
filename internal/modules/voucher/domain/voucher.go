package domain

import "time"

type Voucher struct {
	ID           int       `db:"id"            json:"id"`
	Code         string    `db:"code"          json:"code"`
	Name         string    `db:"name"          json:"name"`
	DiscountType string    `db:"discount_type" json:"discount_type"` // percent | fixed
	Value        float64   `db:"value"         json:"value"`
	MinPurchase  float64   `db:"min_purchase"  json:"min_purchase"`
	MaxDiscount  *float64  `db:"max_discount"  json:"max_discount,omitempty"`
	Quota        int       `db:"quota"         json:"quota"`
	UsedCount    int       `db:"used_count"    json:"used_count"`
	StartDate    time.Time `db:"start_date"    json:"start_date"`
	EndDate      time.Time `db:"end_date"      json:"end_date"`
	IsActive     bool      `db:"is_active"     json:"is_active"`
}
