package impl

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/yourname/pos-koperasi/internal/midtrans"
	"github.com/yourname/pos-koperasi/internal/modules/transaction/contract"
	"github.com/yourname/pos-koperasi/internal/modules/transaction/domain"
)

type Service struct {
	repo     contract.TransactionRepository
	midtrans *midtrans.Client
}

func New(repo contract.TransactionRepository, midtransClient *midtrans.Client) *Service {
	return &Service{repo: repo, midtrans: midtransClient}
}

func (s *Service) CreateTransaction(input contract.CreateTransactionInput, employeeID int) (*contract.CreateTransactionOutput, error) {
	var items []domain.TransactionItem
	var subtotal float64

	for _, in := range input.Items {
		product, err := s.repo.GetProductWithStock(in.ProductID, input.BranchID)
		if err != nil {
			return nil, fmt.Errorf("product %d: %w", in.ProductID, err)
		}
		if product.Stock < in.Quantity {
			return nil, fmt.Errorf("stok %s tidak cukup (tersedia: %d)", product.Name, product.Stock)
		}

		lineTotal := product.SellPrice * float64(in.Quantity)
		items = append(items, domain.TransactionItem{
			ProductID: in.ProductID,
			Quantity:  in.Quantity,
			UnitPrice: product.SellPrice,
			Discount:  0,
			Subtotal:  lineTotal,
		})
		subtotal += lineTotal
	}

	var memberID *int
	if input.MemberCode != "" {
		member, _ := s.repo.GetMemberByCode(input.MemberCode)
		if member != nil {
			memberID = &member.ID
		}
	}

	tax := math.Round(subtotal*0.11*100) / 100
	grandTotal := subtotal + tax

	invoiceNo := fmt.Sprintf("INV-%s-%d", time.Now().Format("20060102150405"), employeeID)

	txID, err := s.repo.CreatePendingTransaction(contract.ProcessTransactionInput{
		Transaction: domain.Transaction{
			BranchID:   input.BranchID,
			RegisterID: input.RegisterID,
			EmployeeID: employeeID,
			MemberID:   memberID,
			InvoiceNo:  invoiceNo,
			Subtotal:   subtotal,
			Discount:   0,
			Tax:        tax,
			GrandTotal: grandTotal,
		},
		Items:    items,
		BranchID: input.BranchID,
	})
	if err != nil {
		return nil, fmt.Errorf("gagal menyimpan transaksi: %w", err)
	}

	trxRow, err := s.repo.GetTransactionByID(txID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat transaksi dari DB: %w", err)
	}

	resp := &contract.CreateTransactionOutput{
		Transaction: *trxRow,
	}

	switch input.PayMethod {
	case "cash":
		err = s.repo.SettleTransaction(contract.SettleInput{
			TransactionID: txID,
			BranchID:      input.BranchID,
			MemberID:      memberID,
			GrandTotal:    grandTotal,
			Payment: domain.Payment{
				Method:       "cash",
				Amount:       grandTotal,
				ChangeAmount: 0,
				ReferenceNo:  invoiceNo,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("gagal settle cash: %w", err)
		}
		resp.Message = "Pembayaran cash berhasil"
		trxPaid, err := s.repo.GetTransactionByID(txID)
		if err != nil {
			return nil, fmt.Errorf("gagal memuat transaksi setelah settle cash: %w", err)
		}
		resp.Transaction = *trxPaid

	case "qris":
		qrisResp, qrURL, _, qrisErr := s.midtrans.CreateChargeQRIS(invoiceNo, grandTotal, "gopay")
		if qrisErr != nil {
			gopayResp, gopayQRURL, gopayRaw, gopayErr := s.midtrans.CreateChargeGoPay(invoiceNo, grandTotal)
			if gopayErr != nil {
				_ = s.repo.CancelTransaction(txID)
				return nil, fmt.Errorf("gagal membuat QRIS (Midtrans): %w (raw=%s)", gopayErr, gopayRaw)
			}
			resp.QRString = gopayQRURL
			resp.MidtransTransactionID = gopayResp.TransactionID
			resp.MidtransTransactionStatus = gopayResp.TransactionStatus
			resp.MidtransPaymentType = gopayResp.PaymentType
			for _, a := range gopayResp.Actions {
				resp.MidtransActions = append(resp.MidtransActions, contract.MidtransAction{
					Name:   a.Name,
					Method: a.Method,
					URL:    a.URL,
				})
			}
		} else {
			resp.QRString = qrURL
			resp.MidtransTransactionID = qrisResp.TransactionID
			resp.MidtransTransactionStatus = qrisResp.TransactionStatus
			resp.MidtransPaymentType = qrisResp.PaymentType
			for _, a := range qrisResp.Actions {
				resp.MidtransActions = append(resp.MidtransActions, contract.MidtransAction{
					Name:   a.Name,
					Method: a.Method,
					URL:    a.URL,
				})
			}
		}
		resp.Message = "Scan QR untuk menyelesaikan pembayaran"

	case "va":
		vaResp, vaNumber, vaRaw, err := s.midtrans.CreateChargeBankTransfer(invoiceNo, grandTotal, "bca")
		if err != nil {
			_ = s.repo.CancelTransaction(txID)
			return nil, fmt.Errorf("gagal membuat Virtual Account (Midtrans): %w (raw=%s)", err, vaRaw)
		}
		resp.VANumber = vaNumber
		resp.MidtransTransactionID = vaResp.TransactionID
		resp.MidtransTransactionStatus = vaResp.TransactionStatus
		resp.MidtransPaymentType = vaResp.PaymentType
		resp.MidtransRawResponse = vaRaw
		resp.Message = "Transfer ke VA untuk menyelesaikan pembayaran"

	case "ewallet":
		chargeResp, paymentURL, gopayRaw, err := s.midtrans.CreateChargeGoPay(invoiceNo, grandTotal)
		if err != nil {
			_ = s.repo.CancelTransaction(txID)
			return nil, fmt.Errorf("gagal membuat e-wallet (Midtrans): %w (raw=%s)", err, gopayRaw)
		}
		resp.PaymentURL = paymentURL
		resp.MidtransTransactionID = chargeResp.TransactionID
		resp.MidtransTransactionStatus = chargeResp.TransactionStatus
		resp.MidtransPaymentType = chargeResp.PaymentType
		resp.MidtransRawResponse = gopayRaw
		for _, a := range chargeResp.Actions {
			resp.MidtransActions = append(resp.MidtransActions, contract.MidtransAction{
				Name:   a.Name,
				Method: a.Method,
				URL:    a.URL,
			})
		}
		resp.Message = "Buka link untuk menyelesaikan pembayaran"

	default:
		return nil, fmt.Errorf("metode pembayaran tidak valid: %s", input.PayMethod)
	}

	return resp, nil
}

func (s *Service) HandleMidtransNotification(payload midtrans.NotificationPayload) error {
	if payload.TransactionStatus != "settlement" {
		return nil
	}

	exists, err := s.repo.PaymentExistsByRef(payload.TransactionID)
	if err != nil {
		return fmt.Errorf("idempotency check failed: %w", err)
	}
	if exists {
		return nil
	}

	trx, err := s.repo.GetTransactionByInvoice(payload.OrderID)
	if err != nil {
		return fmt.Errorf("transaction not found for order_id %s: %w", payload.OrderID, err)
	}

	if trx.Status != "pending" {
		return nil
	}

	method := mapMidtransMethod(payload.PaymentType)

	paidAmount, err := strconv.ParseFloat(string(payload.GrossAmount), 64)
	if err != nil {
		return fmt.Errorf("failed to parse midtrans gross_amount: %w", err)
	}

	return s.repo.SettleTransaction(contract.SettleInput{
		TransactionID: trx.ID,
		BranchID:      trx.BranchID,
		MemberID:      trx.MemberID,
		GrandTotal:    trx.GrandTotal,
		Payment: domain.Payment{
			Method:       method,
			Amount:       paidAmount,
			ChangeAmount: 0,
			ReferenceNo:  payload.TransactionID,
		},
	})
}

func mapMidtransMethod(paymentType string) string {
	switch paymentType {
	case "qris":
		return "qris"
	case "bank_transfer":
		return "va"
	case "gopay":
		return "ewallet"
	default:
		return "ewallet"
	}
}

