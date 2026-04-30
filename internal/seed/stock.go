package seed

import "github.com/jmoiron/sqlx"

var seedStockByBarcode = map[string]int{
	"8999999001": 100,
	"8999999002": 200,
	"8999999003": 50,
}

// StockSeed memastikan stok per produk di cabang (idempoten).
func StockSeed(tx *sqlx.Tx, branchID int) error {
	for barcode, qty := range seedStockByBarcode {
		var productID int
		if err := tx.Get(&productID, `SELECT id FROM products WHERE barcode = $1`, barcode); err != nil {
			return err
		}
		_, err := tx.Exec(
			`INSERT INTO stocks (product_id, branch_id, quantity) VALUES ($1, $2, $3)
			 ON CONFLICT (product_id, branch_id) DO NOTHING`,
			productID, branchID, qty,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
