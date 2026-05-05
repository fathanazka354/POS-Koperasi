package transaction

// RewardPointsFromPurchase menghitung poin loyalitas dari nilai transaksi (aturan bisnis: per Rp 10.000).
func RewardPointsFromPurchase(grandTotal float64) int {
	if grandTotal <= 0 {
		return 0
	}
	return int(grandTotal / 10000)
}

