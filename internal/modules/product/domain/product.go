package domain

import "time"

type Product struct {
	ID         int       `db:"id"          json:"id"`
	CategoryID int       `db:"category_id" json:"category_id"`
	SupplierID int       `db:"supplier_id" json:"supplier_id"`
	Barcode    string    `db:"barcode"     json:"barcode"`
	Name       string    `db:"name"        json:"name"`
	Unit       string    `db:"unit"        json:"unit"`
	BuyPrice   float64   `db:"buy_price"   json:"buy_price"`
	SellPrice  float64   `db:"sell_price"  json:"sell_price"`
	MinStock   int       `db:"min_stock"   json:"min_stock"`
	IsActive   bool      `db:"is_active"   json:"is_active"`
	CreatedAt  time.Time `db:"created_at"  json:"created_at"`
}

type ProductWithStock struct {
	Product
	Stock int `db:"stock" json:"stock"`
}

