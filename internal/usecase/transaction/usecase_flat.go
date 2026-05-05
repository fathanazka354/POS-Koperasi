package transaction

import (
	"context"
	"fmt"
	"log"
	"math"
	"strconv"
	"time"

	"github.com/fathanazka354/pos-koperasi/internal/entity/transaction"
	"github.com/fathanazka354/pos-koperasi/internal/gateway/outbox"
	"github.com/fathanazka354/pos-koperasi/internal/midtrans"
	"github.com/fathanazka354/pos-koperasi/internal/usecase/transaction/notifyoutbox"
)

type usecase struct {
	repo     Repository
	midtrans *midtrans.Client
	outbox   outbox.PaymentOutboxStore
}

func New(repo Repository, midtransClient *midtrans.Client, ob outbox.PaymentOutboxStore) Usecase {
	return &usecase{repo: repo, midtrans: midtransClient, outbox: ob}
}

func (u *usecase) CreateTransaction(input CreateTransactionInput, employeeID int) (*CreateTransactionOutput, error) {
	var items []transaction.TransactionItem
	var subtotal float64
	var midLines []midtrans.ProductLineInput

	for _, in := range input.Items {
		product, err := u.repo.GetProductWithStock(in.ProductID, input.BranchID)
		if err != nil {
			return nil, fmt.Errorf("product %d: %w", in.ProductID, err)
		}
		if product.Stock < in.Quantity {
			return nil, fmt.Errorf("stok %s tidak cukup (tersedia: %d)", product.Name, product.Stock)
		}

		lineTotal := product.SellPrice * float64(in.Quantity)
		items = append(items, transaction.TransactionItem{
			ProductID: in.ProductID,
			Quantity:  in.Quantity,
			UnitPrice: product.SellPrice,
			Discount:  0,
			Subtotal:  lineTotal,
		})
		midLines = append(midLines, midtrans.ProductLineInput{
			ProductID: in.ProductID,
			Name:      product.Name,
			UnitPrice: product.SellPrice,
			Quantity:  in.Quantity,
		})
		subtotal += lineTotal
	}

	var memberID *int
	if input.MemberCode != "" {
		member, _ := u.repo.GetMemberByCode(input.MemberCode)
		if member != nil {
			memberID = &member.ID
		}
	}

	tax := math.Round(subtotal*0.11*100) / 100
	grandTotal := subtotal + tax

	invoiceNo := fmt.Sprintf("INV-%s-%d", time.Now().Format("20060102150405"), employeeID)

	txID, err := u.repo.CreatePendingTransaction(ProcessTransactionInput{
		Transaction: transaction.Transaction{
			BranchID:   input.BranchID,
			RegisterID: input.RegisterID,
			EmployeeID: &employeeID,
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

	trxRow, err := u.repo.GetTransactionByID(txID)
	if err != nil {
		return nil, fmt.Errorf("gagal memuat transaksi dari DB: %w", err)
	}

	resp := &CreateTransactionOutput{
		Transaction: *trxRow,
	}

	chargeAmt := float64(int64(math.Round(grandTotal)))
	posEx := u.posMidtransExtras(chargeAmt, midLines, tax, memberID)

	switch input.PayMethod {
	case "cash":
		err = u.repo.SettleTransaction(SettleInput{
			TransactionID: txID,
			BranchID:      input.BranchID,
			MemberID:      memberID,
			GrandTotal:    grandTotal,
			RewardPoints:  transaction.RewardPointsFromPurchase(grandTotal),
			Payment: transaction.Payment{
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
		trxPaid, err := u.repo.GetTransactionByID(txID)
		if err != nil {
			return nil, fmt.Errorf("gagal memuat transaksi setelah settle cash: %w", err)
		}
		resp.Transaction = *trxPaid
		if m := notifyoutbox.PaymentSuccessPayload(trxPaid); m != nil && trxPaid.MemberID != nil {
			if err := u.outbox.EnqueueMemberPayment(context.Background(), trxPaid.ID, *trxPaid.MemberID, m); err != nil {
				log.Printf("outbox: enqueue payment notify tx=%d: %v", trxPaid.ID, err)
			}
		}

	case "qris":
		qrisResp, qrURL, _, qrisErr := u.midtrans.CreateChargeQRIS(invoiceNo, chargeAmt, "gopay", posEx)
		if qrisErr != nil {
			gopayResp, gopayQRURL, gopayRaw, gopayErr := u.midtrans.CreateChargeGoPay(invoiceNo, chargeAmt, posEx)
			if gopayErr != nil {
				_ = u.repo.CancelTransaction(txID)
				return nil, fmt.Errorf("gagal membuat QRIS (Midtrans): %w (raw=%s)", gopayErr, gopayRaw)
			}
			resp.QRString = gopayQRURL
			resp.MidtransTransactionID = gopayResp.TransactionID
			resp.MidtransTransactionStatus = gopayResp.TransactionStatus
			resp.MidtransPaymentType = gopayResp.PaymentType
			for _, a := range gopayResp.Actions {
				resp.MidtransActions = append(resp.MidtransActions, MidtransAction{
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
				resp.MidtransActions = append(resp.MidtransActions, MidtransAction{
					Name:   a.Name,
					Method: a.Method,
					URL:    a.URL,
				})
			}
		}
		resp.Message = "Scan QR untuk menyelesaikan pembayaran"

	case "va":
		vaResp, vaNumber, vaRaw, err := u.midtrans.CreateChargeBankTransfer(invoiceNo, chargeAmt, "bca", posEx)
		if err != nil {
			_ = u.repo.CancelTransaction(txID)
			return nil, fmt.Errorf("gagal membuat Virtual Account (Midtrans): %w (raw=%s)", err, vaRaw)
		}
		resp.VANumber = vaNumber
		resp.MidtransTransactionID = vaResp.TransactionID
		resp.MidtransTransactionStatus = vaResp.TransactionStatus
		resp.MidtransPaymentType = vaResp.PaymentType
		resp.MidtransRawResponse = vaRaw
		resp.Message = "Transfer ke VA untuk menyelesaikan pembayaran"

	case "ewallet":
		chargeResp, paymentURL, gopayRaw, err := u.midtrans.CreateChargeGoPay(invoiceNo, chargeAmt, posEx)
		if err != nil {
			_ = u.repo.CancelTransaction(txID)
			return nil, fmt.Errorf("gagal membuat e-wallet (Midtrans): %w (raw=%s)", err, gopayRaw)
		}
		resp.PaymentURL = paymentURL
		resp.MidtransTransactionID = chargeResp.TransactionID
		resp.MidtransTransactionStatus = chargeResp.TransactionStatus
		resp.MidtransPaymentType = chargeResp.PaymentType
		resp.MidtransRawResponse = gopayRaw
		for _, a := range chargeResp.Actions {
			resp.MidtransActions = append(resp.MidtransActions, MidtransAction{
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

func (u *usecase) HandleMidtransNotification(payload MidtransNotification) error {
	if payload.TransactionStatus != "settlement" {
		return nil
	}

	exists, err := u.repo.PaymentExistsByRef(payload.TransactionID)
	if err != nil {
		return fmt.Errorf("idempotency check failed: %w", err)
	}
	if exists {
		return nil
	}

	trx, err := u.repo.GetTransactionByInvoice(payload.OrderID)
	if err != nil {
		return fmt.Errorf("transaction not found for order_id %s: %w", payload.OrderID, err)
	}

	if trx.Status != "pending" {
		return nil
	}

	method := mapMidtransMethod(payload.PaymentType)

	paidAmount, err := strconv.ParseFloat(payload.GrossAmount, 64)
	if err != nil {
		return fmt.Errorf("failed to parse midtrans gross_amount: %w", err)
	}

	if err := u.repo.SettleTransaction(SettleInput{
		TransactionID: trx.ID,
		BranchID:      trx.BranchID,
		MemberID:      trx.MemberID,
		GrandTotal:    trx.GrandTotal,
		RewardPoints:  transaction.RewardPointsFromPurchase(trx.GrandTotal),
		Payment: transaction.Payment{
			Method:       method,
			Amount:       paidAmount,
			ChangeAmount: 0,
			ReferenceNo:  payload.TransactionID,
		},
	}); err != nil {
		return err
	}
	trxPaid, err := u.repo.GetTransactionByID(trx.ID)
	if err != nil {
		return fmt.Errorf("reload transaction after settle: %w", err)
	}
	if m := notifyoutbox.PaymentSuccessPayload(trxPaid); m != nil && trxPaid.MemberID != nil {
		if err := u.outbox.EnqueueMemberPayment(context.Background(), trxPaid.ID, *trxPaid.MemberID, m); err != nil {
			log.Printf("outbox: enqueue payment notify tx=%d: %v", trxPaid.ID, err)
		}
	}
	return nil
}

func (u *usecase) posMidtransExtras(amount float64, lines []midtrans.ProductLineInput, tax float64, memberID *int) *midtrans.ChargeExtras {
	cust := midtrans.CustomerInput{FullName: "Pelanggan", Phone: "", Email: ""}
	if memberID != nil {
		m, err := u.repo.GetMemberByID(*memberID)
		if err == nil && m != nil {
			cust = midtrans.CustomerInput{
				FullName: m.FullName,
				Phone:    m.Phone,
				Email:    "",
			}
		}
	}
	ex, err := midtrans.NewChargeExtrasFromCart(amount, lines, 0, tax, cust, nil, nil)
	if err != nil {
		log.Printf("midtrans pos extras: %v", err)
		return nil
	}
	return ex
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

