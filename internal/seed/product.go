package seed

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type productRow struct {
	Barcode      string
	Name         string
	Unit         string
	BuyPrice     float64
	SellPrice    float64
	MinStock     int
	CategoryName string
	SupplierName string
}

var seedProducts = []productRow{
	{
		Barcode: "8999999001", Name: "Indomie Goreng", Unit: "pcs",
		BuyPrice: 2500, SellPrice: 3500, MinStock: 10,
		CategoryName: "Makanan", SupplierName: "PT Indofood",
	},
	{
		Barcode: "8999999002", Name: "Aqua 600ml", Unit: "pcs",
		BuyPrice: 2000, SellPrice: 3000, MinStock: 20,
		CategoryName: "Minuman", SupplierName: "PT Unilever Indonesia",
	},
	{
		Barcode: "8999999003", Name: "Sunlight 200ml", Unit: "pcs",
		BuyPrice: 5000, SellPrice: 7500, MinStock: 5,
		CategoryName: "Kebersihan", SupplierName: "PT Unilever Indonesia",
	},
}

// ProductSeed memastikan produk demo ada (berdasarkan barcode unik).
func ProductSeed(tx *sqlx.Tx, categories map[string]int, suppliers map[string]int) error {
	for _, p := range seedProducts {
		cid := categories[p.CategoryName]
		sid := suppliers[p.SupplierName]
		if cid == 0 || sid == 0 {
			return fmt.Errorf("category/supplier tidak lengkap untuk barcode %s", p.Barcode)
		}

		_, err := tx.Exec(
			`INSERT INTO products (category_id, supplier_id, barcode, name, unit, buy_price, sell_price, min_stock)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			 ON CONFLICT (barcode) DO NOTHING`,
			cid, sid, p.Barcode, p.Name, p.Unit, p.BuyPrice, p.SellPrice, p.MinStock,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
