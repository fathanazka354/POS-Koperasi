package seed

import "gorm.io/gorm"

var seedStockByBarcode = map[string]int{
	"8999999001": 100,
	"8999999002": 200,
	"8999999003": 50,
}

// StockSeed memastikan stok per produk di cabang (idempoten).
func StockSeed(tx *gorm.DB, branchID int) error {
	for barcode, qty := range seedStockByBarcode {
		var productID int
		if err := tx.Raw(`SELECT id FROM products WHERE barcode = $1`, barcode).Scan(&productID).Error; err != nil {
			return err
		}
		err := tx.Exec(
			`INSERT INTO stocks (product_id, branch_id, quantity) VALUES ($1, $2, $3)
			 ON CONFLICT (product_id, branch_id) DO NOTHING`,
			productID, branchID, qty,
		).Error
		if err != nil {
			return err
		}
	}
	return nil
}
