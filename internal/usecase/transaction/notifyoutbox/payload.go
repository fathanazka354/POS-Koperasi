package notifyoutbox

import (
	"fmt"

	"github.com/fathanazka354/pos-koperasi/internal/entity/transaction"
	"github.com/fathanazka354/pos-koperasi/internal/gateway/outbox"
)

// PaymentSuccessPayload menghasilkan payload notifikasi pembayaran untuk outbox MongoDB (nil jika bukan member).
func PaymentSuccessPayload(trx *transaction.Transaction) *outbox.MemberNotifyOutbox {
	if trx == nil || trx.MemberID == nil {
		return nil
	}
	return &outbox.MemberNotifyOutbox{
		Type:  "payment_success",
		Title: "Pembayaran Berhasil!",
		Body:  fmt.Sprintf("Pesanan %s telah terkonfirmasi. Terima kasih!", trx.InvoiceNo),
		Data: map[string]interface{}{
			"transaction_id": trx.ID,
			"invoice_no":     trx.InvoiceNo,
			"grand_total":    trx.GrandTotal,
		},
	}
}

