package midtrans

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/yourname/pos-koperasi/internal/config"
)

type Client struct {
	cfg        *config.Config
	httpClient *http.Client
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

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

// CreateChargeQRIS membuat QRIS payment.
// Jika acquirer kosong (""), field qris tidak dikirim dan Midtrans menggunakan
// acquirer default merchant (menghindari error "Merchant pop id is not found").
func (c *Client) CreateChargeQRIS(orderID string, amount float64, acquirer string) (*chargeQRISResponse, string, string, error) {
	req := createChargeQRISRequest{
		PaymentType: "qris",
		TransactionDetails: transactionDetails{
			OrderID:     orderID,
			GrossAmount: amount,
		},
	}
	if acquirer != "" {
		req.QRIS = &createChargeQris{Acquirer: acquirer}
	}

	body, _ := json.Marshal(req)

	httpReq, err := http.NewRequest("POST", c.cfg.MidtransBaseURL+"/v2/charge", bytes.NewBuffer(body))
	if err != nil {
		return nil, "", "", err
	}
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.SetBasicAuth(c.cfg.MidtransServerKey, "")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, "", "", fmt.Errorf("midtrans qris charge failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, "", string(respBody), fmt.Errorf("midtrans qris error [%d]: %s", resp.StatusCode, string(respBody))
	}

	// Midtrans bisa mengembalikan error payload di dalam body walaupun HTTP-nya 200/201.
	// Contoh: {"status_code":"404","status_message":"Merchant pop id is not found", ...}
	var statusCheck struct {
		StatusCode    string `json:"status_code"`
		StatusMessage string `json:"status_message"`
	}
	if err := json.Unmarshal(respBody, &statusCheck); err == nil && statusCheck.StatusCode != "" && statusCheck.StatusCode != "200" {
		return nil, "", string(respBody), fmt.Errorf("midtrans qris error [%s]: %s", statusCheck.StatusCode, statusCheck.StatusMessage)
	}

	var result chargeQRISResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, "", string(respBody), fmt.Errorf("failed to parse midtrans qris response: %w", err)
	}

	qrURL := firstActionURL(result.Actions, "generate-qr-code")
	if qrURL == "" {
		// fallback bila hanya ada generate-qr-code-v2
		qrURL = firstActionURL(result.Actions, "generate-qr-code-v2")
	}

	return &result, qrURL, string(respBody), nil
}

// ─── Bank Transfer (VA) ──────────────────────────────────────────────────────

type createChargeBankTransferRequest struct {
	PaymentType        string             `json:"payment_type"`
	TransactionDetails transactionDetails `json:"transaction_details"`
	BankTransfer       createBankTransfer `json:"bank_transfer"`
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

func (c *Client) CreateChargeBankTransfer(orderID string, amount float64, bank string) (*chargeBankTransferResponse, string, string, error) {
	req := createChargeBankTransferRequest{
		PaymentType: "bank_transfer",
		TransactionDetails: transactionDetails{
			OrderID:     orderID,
			GrossAmount: amount,
		},
		BankTransfer: createBankTransfer{
			Bank: bank,
		},
	}

	body, _ := json.Marshal(req)

	httpReq, err := http.NewRequest("POST", c.cfg.MidtransBaseURL+"/v2/charge", bytes.NewBuffer(body))
	if err != nil {
		return nil, "", "", err
	}
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.SetBasicAuth(c.cfg.MidtransServerKey, "")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, "", "", fmt.Errorf("midtrans bank_transfer charge failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	raw := string(respBody)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, "", raw, fmt.Errorf("midtrans bank_transfer error [%d]: %s", resp.StatusCode, raw)
	}
	if err := checkMidtransBodyError(respBody); err != nil {
		return nil, "", raw, fmt.Errorf("midtrans bank_transfer error: %w", err)
	}

	var result chargeBankTransferResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, "", raw, fmt.Errorf("failed to parse midtrans bank_transfer response: %w", err)
	}

	vaNumber := ""
	if len(result.VaNumbers) > 0 {
		vaNumber = result.VaNumbers[0].VaNumber
	}

	return &result, vaNumber, raw, nil
}

// ─── E-Wallet (GoPay) ───────────────────────────────────────────────────────

type createChargeGoPayRequest struct {
	PaymentType        string             `json:"payment_type"`
	TransactionDetails transactionDetails `json:"transaction_details"`
	GoPay              createGoPay        `json:"gopay"`
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

// CreateChargeGoPay membuat charge e-wallet dan mengambil URL dari actions (deeplink) atau redirect_url.
func (c *Client) CreateChargeGoPay(orderID string, amount float64) (*chargeGoPayResponse, string, string, error) {
	req := createChargeGoPayRequest{
		PaymentType: "gopay",
		TransactionDetails: transactionDetails{
			OrderID:     orderID,
			GrossAmount: amount,
		},
		GoPay: createGoPay{
			EnableCallback: false,
		},
	}

	body, _ := json.Marshal(req)

	httpReq, err := http.NewRequest("POST", c.cfg.MidtransBaseURL+"/v2/charge", bytes.NewBuffer(body))
	if err != nil {
		return nil, "", "", err
	}
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.SetBasicAuth(c.cfg.MidtransServerKey, "")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, "", "", fmt.Errorf("midtrans gopay charge failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	raw := string(respBody)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, "", raw, fmt.Errorf("midtrans gopay error [%d]: %s", resp.StatusCode, raw)
	}
	if err := checkMidtransBodyError(respBody); err != nil {
		return nil, "", raw, fmt.Errorf("midtrans gopay error: %w", err)
	}

	var result chargeGoPayResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, "", raw, fmt.Errorf("failed to parse midtrans gopay response: %w", err)
	}

	// Midtrans menyediakan actions; prefer deeplink-redirect, fallback generate-qr-code.
	paymentURL := firstActionURL(result.Actions, "deeplink-redirect")
	if paymentURL == "" {
		paymentURL = firstActionURL(result.Actions, "generate-qr-code")
	}
	if paymentURL == "" {
		paymentURL = result.RedirectURL
	}

	return &result, paymentURL, raw, nil
}

// ─── Get Transaction Status ──────────────────────────────────────────────────

// TransactionStatusResponse berisi status transaksi dari Midtrans.
type TransactionStatusResponse struct {
	TransactionID     string            `json:"transaction_id"`
	OrderID           string            `json:"order_id"`
	TransactionStatus string            `json:"transaction_status"`
	PaymentType       string            `json:"payment_type"`
	GrossAmount       string            `json:"gross_amount"`
	StatusCode        string            `json:"status_code"`
	StatusMessage     string            `json:"status_message"`
	FraudStatus       string            `json:"fraud_status"`
}

// GetTransactionStatus query status transaksi ke Midtrans (untuk polling).
func (c *Client) GetTransactionStatus(orderID string) (*TransactionStatusResponse, error) {
	url := fmt.Sprintf("%s/v2/%s/status", c.cfg.MidtransBaseURL, orderID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(c.cfg.MidtransServerKey, "")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("midtrans get status failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result TransactionStatusResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse midtrans status response: %w", err)
	}
	// status_code 404 berarti order tidak ditemukan di Midtrans (belum dibuat charge atau sudah expire)
	if result.StatusCode == "404" {
		return nil, fmt.Errorf("order %s tidak ditemukan di Midtrans", orderID)
	}
	return &result, nil
}

func firstActionURL(actions []chargeAction, actionName string) string {
	for _, a := range actions {
		if a.Name == actionName {
			return a.URL
		}
	}
	return ""
}

// checkMidtransBodyError mendeteksi error payload Midtrans yang dikembalikan
// di dalam body JSON walaupun HTTP status-nya 200/201.
// Contoh: {"status_code":"401","status_message":"Access denied due to unauthorized transaction"}
func checkMidtransBodyError(respBody []byte) error {
	var s struct {
		StatusCode    string `json:"status_code"`
		StatusMessage string `json:"status_message"`
	}
	if err := json.Unmarshal(respBody, &s); err != nil {
		return nil
	}
	if s.StatusCode != "" && s.StatusCode != "200" && s.StatusCode != "201" {
		return fmt.Errorf("[%s] %s", s.StatusCode, s.StatusMessage)
	}
	return nil
}
