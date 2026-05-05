package shop

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/fathanazka354/pos-koperasi/internal/entity/transaction"
	"github.com/fathanazka354/pos-koperasi/internal/entity/voucher"
	"github.com/fathanazka354/pos-koperasi/internal/gateway/outbox"
	"github.com/fathanazka354/pos-koperasi/internal/midtrans"
	addruc "github.com/fathanazka354/pos-koperasi/internal/usecase/address"
	txuc "github.com/fathanazka354/pos-koperasi/internal/usecase/transaction"
	"github.com/fathanazka354/pos-koperasi/internal/usecase/transaction/notifyoutbox"
	voucheruc "github.com/fathanazka354/pos-koperasi/internal/usecase/voucher"
)

type usecase struct {
	txRepo     txuc.Repository
	addrRepo   addruc.Repository
	voucherSvc voucheruc.Usecase
	midtrans   *midtrans.Client
	outbox     outbox.PaymentOutboxStore
}

func New(
	txRepo txuc.Repository,
	addrRepo addruc.Repository,
	voucherSvc voucheruc.Usecase,
	mt *midtrans.Client,
	ob outbox.PaymentOutboxStore,
) Usecase {
	return &usecase{
		txRepo:     txRepo,
		addrRepo:   addrRepo,
		voucherSvc: voucherSvc,
		midtrans:   mt,
		outbox:     ob,
	}
}

func (u *usecase) Checkout(input CheckoutInput) (*CheckoutOutput, error) {
	const branchID = 1

	// Validasi alamat milik member
	if input.AddressID > 0 {
		_, err := u.addrRepo.GetByID(input.AddressID, input.MemberID)
		if err != nil {
			return nil, fmt.Errorf("alamat tidak ditemukan atau bukan milik Anda")
		}
	}

	// Kalkulasi item (+ baris untuk Midtrans item_details)
	var items []transaction.TransactionItem
	var subtotal float64
	var midLines []midtrans.ProductLineInput
	for _, in := range input.Items {
		p, err := u.txRepo.GetProductWithStock(in.ProductID, branchID)
		if err != nil {
			return nil, fmt.Errorf("produk %d tidak ditemukan", in.ProductID)
		}
		if p.Stock < in.Quantity {
			return nil, fmt.Errorf("stok %s tidak cukup (tersedia: %d)", p.Name, p.Stock)
		}
		line := p.SellPrice * float64(in.Quantity)
		items = append(items, transaction.TransactionItem{
			ProductID: in.ProductID,
			Quantity:  in.Quantity,
			UnitPrice: p.SellPrice,
			Subtotal:  line,
		})
		midLines = append(midLines, midtrans.ProductLineInput{
			ProductID: in.ProductID,
			Name:      p.Name,
			UnitPrice: p.SellPrice,
			Quantity:  in.Quantity,
		})
		subtotal += line
	}

	// Validasi & terapkan voucher
	var voucherDiscount float64
	var voucherID *int
	if input.VoucherCode != "" {
		res, err := u.voucherSvc.Validate(input.VoucherCode, subtotal)
		if err != nil {
			return nil, err
		}
		voucherDiscount = res.DiscountAmt
		voucherID = &res.Voucher.ID
	}

	tax := math.Round((subtotal-voucherDiscount)*0.11*100) / 100
	grandTotal := subtotal - voucherDiscount + tax

	invoiceNo := fmt.Sprintf("SHOP-%s-%d", time.Now().Format("20060102150405"), input.MemberID)

	txInput := txuc.ProcessTransactionInput{
		Transaction: transaction.Transaction{
			BranchID:        branchID,
			RegisterID:      1,
			MemberID:        &input.MemberID,
			InvoiceNo:       invoiceNo,
			Subtotal:        subtotal,
			Discount:        voucherDiscount,
			Tax:             tax,
			GrandTotal:      grandTotal,
			AddressID:       ptrInt(input.AddressID),
			VoucherID:       voucherID,
			VoucherDiscount: voucherDiscount,
		},
		Items:    items,
		BranchID: branchID,
	}

	txID, err := u.txRepo.CreatePendingTransaction(txInput)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat transaksi: %w", err)
	}
	trx, _ := u.txRepo.GetTransactionByID(txID)

	resp := &CheckoutOutput{
		CreateTransactionOutput: &txuc.CreateTransactionOutput{
			Transaction: *trx,
		},
		VoucherDiscount: voucherDiscount,
	}

	chargeAmt := float64(int64(math.Round(grandTotal)))
	mExtras := u.shopMidtransExtras(chargeAmt, midLines, voucherDiscount, tax, input.MemberID, input.AddressID)

	switch input.PayMethod {
	case "cash":
		err = u.txRepo.SettleTransaction(txuc.SettleInput{
			TransactionID: txID,
			BranchID:      branchID,
			MemberID:      &input.MemberID,
			GrandTotal:    grandTotal,
			RewardPoints:  transaction.RewardPointsFromPurchase(grandTotal),
			Payment: transaction.Payment{
				Method:      "cash",
				Amount:      grandTotal,
				ReferenceNo: invoiceNo,
			},
		})
		if err != nil {
			return nil, err
		}
		resp.Message = "Pembayaran cash berhasil"
		trxPaid, _ := u.txRepo.GetTransactionByID(txID)
		resp.Transaction = *trxPaid
		if m := notifyoutbox.PaymentSuccessPayload(trxPaid); m != nil && trxPaid.MemberID != nil {
			if err := u.outbox.EnqueueMemberPayment(context.Background(), trxPaid.ID, *trxPaid.MemberID, m); err != nil {
				log.Printf("outbox: enqueue payment notify tx=%d: %v", trxPaid.ID, err)
			}
		}

	case "qris":
		qrisResp, qrURL, _, qrisErr := u.midtrans.CreateChargeQRIS(invoiceNo, chargeAmt, "gopay", mExtras)
		if qrisErr != nil {
			gopayResp, gopayURL, _, gopayErr := u.midtrans.CreateChargeGoPay(invoiceNo, chargeAmt, mExtras)
			if gopayErr != nil {
				_ = u.txRepo.CancelTransaction(txID)
				return nil, fmt.Errorf("gagal membuat QRIS: %w", gopayErr)
			}
			resp.QRString = gopayURL
			resp.MidtransTransactionID = gopayResp.TransactionID
			for _, a := range gopayResp.Actions {
				resp.MidtransActions = append(resp.MidtransActions, txuc.MidtransAction{
					Name: a.Name, Method: a.Method, URL: a.URL,
				})
			}
		} else {
			resp.QRString = qrURL
			resp.MidtransTransactionID = qrisResp.TransactionID
			for _, a := range qrisResp.Actions {
				resp.MidtransActions = append(resp.MidtransActions, txuc.MidtransAction{
					Name: a.Name, Method: a.Method, URL: a.URL,
				})
			}
		}
		resp.Message = "Scan QR untuk menyelesaikan pembayaran"

	case "va":
		vaResp, vaNumber, _, err := u.midtrans.CreateChargeBankTransfer(invoiceNo, chargeAmt, "bca", mExtras)
		if err != nil {
			_ = u.txRepo.CancelTransaction(txID)
			return nil, fmt.Errorf("gagal membuat Virtual Account: %w", err)
		}
		resp.VANumber = vaNumber
		resp.MidtransTransactionID = vaResp.TransactionID
		resp.Message = "Transfer ke VA untuk menyelesaikan pembayaran"

	case "ewallet":
		chargeResp, paymentURL, _, err := u.midtrans.CreateChargeGoPay(invoiceNo, chargeAmt, mExtras)
		if err != nil {
			_ = u.txRepo.CancelTransaction(txID)
			return nil, fmt.Errorf("gagal membuat e-wallet: %w", err)
		}
		resp.PaymentURL = paymentURL
		resp.MidtransTransactionID = chargeResp.TransactionID
		for _, a := range chargeResp.Actions {
			resp.MidtransActions = append(resp.MidtransActions, txuc.MidtransAction{
				Name: a.Name, Method: a.Method, URL: a.URL,
			})
		}
		resp.Message = "Buka link untuk menyelesaikan pembayaran"

	default:
		return nil, fmt.Errorf("metode pembayaran tidak valid: %s", input.PayMethod)
	}

	// Tandai voucher sudah dipakai (best-effort) — saat ini cukup tersimpan di transaksi.
	if voucherID != nil {
		_ = voucherID
	}

	return resp, nil
}

func (u *usecase) ValidateVoucher(code string, subtotal float64) (*VoucherValidateResult, error) {
	res, err := u.voucherSvc.Validate(code, subtotal)
	if err != nil {
		return nil, err
	}
	return &VoucherValidateResult{
		Voucher:     res.Voucher,
		DiscountAmt: res.DiscountAmt,
		Label:       res.Label,
	}, nil
}

func (u *usecase) ListVouchers() ([]*voucher.Voucher, error) {
	return u.voucherSvc.ListActive()
}

func (u *usecase) GetTransactionsByMember(memberID int) ([]transaction.Transaction, error) {
	return u.txRepo.GetTransactionsByMember(memberID)
}

func (u *usecase) GetTransactionDetail(invoiceNo string, memberID int) (*TransactionDetail, error) {
	trx, err := u.txRepo.GetTransactionByInvoice(invoiceNo)
	if err != nil {
		return nil, fmt.Errorf("transaksi tidak ditemukan")
	}
	if trx.MemberID == nil || *trx.MemberID != memberID {
		return nil, fmt.Errorf("akses ditolak")
	}
	items, err := u.txRepo.GetItemsByTransactionID(trx.ID)
	if err != nil {
		return nil, err
	}
	return &TransactionDetail{Transaction: *trx, Items: items}, nil
}

func (u *usecase) CheckPaymentStatus(invoiceNo string, memberID int) (string, error) {
	trx, err := u.txRepo.GetTransactionByInvoice(invoiceNo)
	if err != nil {
		return "", fmt.Errorf("transaksi tidak ditemukan")
	}
	if trx.MemberID == nil || *trx.MemberID != memberID {
		return "", fmt.Errorf("akses ditolak")
	}

	// Kalau DB sudah final (bukan pending), langsung kembalikan
	if trx.Status != "pending" {
		return trx.Status, nil
	}

	// DB masih pending → tanya Midtrans langsung (polling fallback jika webhook belum datang)
	mtStatus, err := u.midtrans.GetTransactionStatus(trx.InvoiceNo)
	if err != nil {
		// Midtrans tidak kenal order ini (misal cash sudah settle sebelumnya), kembalikan status DB
		return trx.Status, nil
	}

	// Jika Midtrans sudah settlement dan DB masih pending → settle manual
	if (mtStatus.TransactionStatus == "settlement" || mtStatus.TransactionStatus == "capture") &&
		(mtStatus.FraudStatus == "" || mtStatus.FraudStatus == "accept") {

		const branchID = 1
		paidAmount := trx.GrandTotal
		if err := u.txRepo.SettleTransaction(txuc.SettleInput{
			TransactionID: trx.ID,
			BranchID:      branchID,
			MemberID:      trx.MemberID,
			GrandTotal:    trx.GrandTotal,
			RewardPoints:  transaction.RewardPointsFromPurchase(trx.GrandTotal),
			Payment: transaction.Payment{
				Method:      mapMidtransMethod(mtStatus.PaymentType),
				Amount:      paidAmount,
				ReferenceNo: mtStatus.TransactionID,
			},
		}); err == nil {
			trxPaid, _ := u.txRepo.GetTransactionByID(trx.ID)
			if trxPaid != nil {
				if m := notifyoutbox.PaymentSuccessPayload(trxPaid); m != nil && trxPaid.MemberID != nil {
					if err := u.outbox.EnqueueMemberPayment(context.Background(), trxPaid.ID, *trxPaid.MemberID, m); err != nil {
						log.Printf("outbox: enqueue payment notify tx=%d: %v", trxPaid.ID, err)
					}
				}
			}
		}

		return "paid", nil
	}

	// Kalau Midtrans bilang cancel/expire
	if mtStatus.TransactionStatus == "cancel" || mtStatus.TransactionStatus == "expire" {
		_ = u.txRepo.CancelTransaction(trx.ID)
		return "cancelled", nil
	}

	return trx.Status, nil
}

func mapMidtransMethod(paymentType string) string {
	switch paymentType {
	case "qris":
		return "qris"
	case "bank_transfer":
		return "va"
	case "gopay", "shopeepay", "dana", "linkaja":
		return "ewallet"
	default:
		return "ewallet"
	}
}

func (u *usecase) shopMidtransExtras(
	amount float64,
	lines []midtrans.ProductLineInput,
	voucherDiscount, tax float64,
	memberID int,
	addressID int,
) *midtrans.ChargeExtras {
	member, err := u.txRepo.GetMemberByID(memberID)
	cust := midtrans.CustomerInput{FullName: "Member", Phone: "", Email: ""}
	if err == nil && member != nil {
		cust = midtrans.CustomerInput{
			FullName: member.FullName,
			Phone:    member.Phone,
			Email:    "",
		}
	}
	var bill *midtrans.AddressInput
	if addressID > 0 {
		addr, err := u.addrRepo.GetByID(addressID, memberID)
		if err == nil && addr != nil {
			bill = &midtrans.AddressInput{
				RecipientName: addr.Recipient,
				Phone:         addr.Phone,
				AddressLine:   addr.AddressLine,
				City:          addr.City,
				PostalCode:    addr.PostalCode,
			}
		}
	}
	ex, err := midtrans.NewChargeExtrasFromCart(amount, lines, voucherDiscount, tax, cust, bill, bill)
	if err != nil {
		log.Printf("midtrans shop extras: %v", err)
		return nil
	}
	return ex
}

func ptrInt(v int) *int {
	if v == 0 {
		return nil
	}
	return &v
}

