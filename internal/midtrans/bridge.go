package midtrans

import txdto "github.com/fathanazka354/pos-koperasi/internal/usecase/transaction/dto"

// MapToTransactionNotification memetakan payload HTTP Midtrans ke DTO layanan transaksi.
func MapToTransactionNotification(p NotificationPayload) txdto.MidtransNotification {
	return txdto.MidtransNotification{
		OrderID:           p.OrderID,
		TransactionID:     p.TransactionID,
		TransactionStatus: p.TransactionStatus,
		PaymentType:       p.PaymentType,
		GrossAmount:       string(p.GrossAmount),
	}
}
