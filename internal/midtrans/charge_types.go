package midtrans

type chargeAction struct {
	Name   string `json:"name"`
	Method string `json:"method"`
	URL    string `json:"url"`
}

// ─── QRIS ─────────────────────────────────────────────────────────────────────

type createChargeQRISRequest struct {
	PaymentType        string             `json:"payment_type"`
	TransactionDetails transactionDetails `json:"transaction_details"`
	// QRIS field bersifat opsional: hanya dikirim bila acquirer diisi.
	// Mengirim acquirer="gopay" membutuhkan Merchant POP ID di dashboard Midtrans.
	// Kalau kosong, Midtrans menggunakan acquirer default merchant yang aktif.
	QRIS *createChargeQris `json:"qris,omitempty"`
	optionalChargeFields
}

type transactionDetails struct {
	OrderID     string  `json:"order_id"`
	GrossAmount float64 `json:"gross_amount"`
}

type createChargeQris struct {
	Acquirer string `json:"acquirer"`
}

type chargeQRISResponse struct {
	TransactionID     string         `json:"transaction_id"`
	OrderID           string         `json:"order_id"`
	TransactionStatus string         `json:"transaction_status"`
	PaymentType       string         `json:"payment_type"`
	Actions           []chargeAction `json:"actions"`
}

// ─── Bank Transfer (VA) ──────────────────────────────────────────────────────

type createChargeBankTransferRequest struct {
	PaymentType        string             `json:"payment_type"`
	TransactionDetails transactionDetails `json:"transaction_details"`
	BankTransfer       createBankTransfer `json:"bank_transfer"`
	optionalChargeFields
}

type createBankTransfer struct {
	Bank string `json:"bank"`
}

type chargeBankTransferResponse struct {
	TransactionID     string     `json:"transaction_id"`
	OrderID           string     `json:"order_id"`
	TransactionStatus string     `json:"transaction_status"`
	PaymentType       string     `json:"payment_type"`
	VaNumbers         []vaNumber `json:"va_numbers"`
}

type vaNumber struct {
	Bank     string `json:"bank"`
	VaNumber string `json:"va_number"`
}

// ─── E-Wallet (GoPay) ───────────────────────────────────────────────────────

type createChargeGoPayRequest struct {
	PaymentType        string             `json:"payment_type"`
	TransactionDetails transactionDetails `json:"transaction_details"`
	GoPay              createGoPay        `json:"gopay"`
	optionalChargeFields
}

type createGoPay struct {
	EnableCallback bool   `json:"enable_callback"`
	CallbackURL    string `json:"callback_url,omitempty"`
}

type chargeGoPayResponse struct {
	TransactionID     string         `json:"transaction_id"`
	OrderID           string         `json:"order_id"`
	TransactionStatus string         `json:"transaction_status"`
	PaymentType       string         `json:"payment_type"`
	Actions           []chargeAction `json:"actions"`
	RedirectURL       string         `json:"redirect_url,omitempty"`
}

// ─── Get Transaction Status ──────────────────────────────────────────────────

// TransactionStatusResponse berisi status transaksi dari Midtrans.
type TransactionStatusResponse struct {
	TransactionID     string `json:"transaction_id"`
	OrderID           string `json:"order_id"`
	TransactionStatus string `json:"transaction_status"`
	PaymentType       string `json:"payment_type"`
	GrossAmount       string `json:"gross_amount"`
	StatusCode        string `json:"status_code"`
	StatusMessage     string `json:"status_message"`
	FraudStatus       string `json:"fraud_status"`
}
