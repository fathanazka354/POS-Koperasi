package midtrans

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
)

// rawStringPreserve keeps JSON numeric/string values as-is for signature calculation.
// Midtrans signature depends on concatenating order_id + status_code + gross_amount + server_key.
type rawStringPreserve string

func (r *rawStringPreserve) UnmarshalJSON(b []byte) error {
	s := strings.TrimSpace(string(b))
	if s == "null" {
		*r = ""
		return nil
	}
	// If quoted string, remove quotes.
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		var out string
		if err := json.Unmarshal([]byte(s), &out); err != nil {
			return err
		}
		*r = rawStringPreserve(out)
		return nil
	}
	// For numbers, keep exact textual representation.
	*r = rawStringPreserve(s)
	return nil
}

type NotificationPayload struct {
	OrderID           string             `json:"order_id"`
	TransactionID     string             `json:"transaction_id"`
	TransactionStatus string            `json:"transaction_status"`
	PaymentType       string             `json:"payment_type"`
	StatusCode        rawStringPreserve `json:"status_code"`
	GrossAmount       rawStringPreserve `json:"gross_amount"`
	SignatureKey      string             `json:"signature_key"`
}

func VerifySignature(p NotificationPayload, serverKey string) error {
	if serverKey == "" {
		return fmt.Errorf("midtrans server key is empty")
	}

	// SHA512(order_id + status_code + gross_amount + server_key)
	input := p.OrderID + string(p.StatusCode) + string(p.GrossAmount) + serverKey
	sum := sha512.Sum512([]byte(input))
	expected := hex.EncodeToString(sum[:])

	// Midtrans signature uses lowercase hex, but normalize just in case.
	got := strings.ToLower(strings.TrimSpace(p.SignatureKey))
	exp := strings.ToLower(expected)
	if got != exp {
		return fmt.Errorf("invalid midtrans notification signature")
	}
	return nil
}

