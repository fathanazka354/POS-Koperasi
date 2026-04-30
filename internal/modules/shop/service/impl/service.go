package impl

import (
	"fmt"
	"math"
	"time"

	"github.com/yourname/pos-koperasi/internal/midtrans"
	"github.com/yourname/pos-koperasi/internal/modules/address/contract"
	shopcontract "github.com/yourname/pos-koperasi/internal/modules/shop/contract"
	txcontract "github.com/yourname/pos-koperasi/internal/modules/transaction/contract"
	txdomain "github.com/yourname/pos-koperasi/internal/modules/transaction/domain"
	vouchercontract "github.com/yourname/pos-koperasi/internal/modules/voucher/contract"
)

// NotifyFn ditetapkan dari inject.go setelah semua modul siap.
type NotifyFn func(memberID int, notifType, title, body string, data map[string]interface{}) error

type Service struct {
	txRepo     txcontract.TransactionRepository
	addrRepo   contract.AddressRepository
	voucherSvc vouchercontract.VoucherService
	midtrans   *midtrans.Client
	notifyFn   NotifyFn
}

func New(
	txRepo txcontract.TransactionRepository,
	addrRepo contract.AddressRepository,
	voucherSvc vouchercontract.VoucherService,
	mt *midtrans.Client,
) *Service {
	return &Service{
		txRepo:     txRepo,
		addrRepo:   addrRepo,
		voucherSvc: voucherSvc,
		midtrans:   mt,
	}
}

func (s *Service) WithNotify(fn NotifyFn) { s.notifyFn = fn }

var _ shopcontract.ShopService = (*Service)(nil)

func (s *Service) Checkout(input shopcontract.CheckoutInput) (*shopcontract.CheckoutOutput, error) {
	const branchID = 1

	// Validasi alamat milik member
	if input.AddressID > 0 {
		_, err := s.addrRepo.GetByID(input.AddressID, input.MemberID)
		if err != nil {
			return nil, fmt.Errorf("alamat tidak ditemukan atau bukan milik Anda")
		}
	}

	// Kalkulasi item
	var items []txdomain.TransactionItem
	var subtotal float64
	for _, in := range input.Items {
		p, err := s.txRepo.GetProductWithStock(in.ProductID, branchID)
		if err != nil {
			return nil, fmt.Errorf("produk %d tidak ditemukan", in.ProductID)
		}
		if p.Stock < in.Quantity {
			return nil, fmt.Errorf("stok %s tidak cukup (tersedia: %d)", p.Name, p.Stock)
		}
		line := p.SellPrice * float64(in.Quantity)
		items = append(items, txdomain.TransactionItem{
			ProductID: in.ProductID,
			Quantity:  in.Quantity,
			UnitPrice: p.SellPrice,
			Subtotal:  line,
		})
		subtotal += line
	}

	// Validasi & terapkan voucher
	var voucherDiscount float64
	var voucherID *int
	if input.VoucherCode != "" {
		res, err := s.voucherSvc.Validate(input.VoucherCode, subtotal)
		if err != nil {
			return nil, err
		}
		voucherDiscount = res.DiscountAmt
		voucherID = &res.Voucher.ID
	}

	tax := math.Round((subtotal-voucherDiscount)*0.11*100) / 100
	grandTotal := subtotal - voucherDiscount + tax

	invoiceNo := fmt.Sprintf("SHOP-%s-%d", time.Now().Format("20060102150405"), input.MemberID)

	txInput := txcontract.ProcessTransactionInput{
		Transaction: txdomain.Transaction{
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

	txID, err := s.txRepo.CreatePendingTransaction(txInput)
	if err != nil {
		return nil, fmt.Errorf("gagal membuat transaksi: %w", err)
	}
	trx, _ := s.txRepo.GetTransactionByID(txID)

	resp := &shopcontract.CheckoutOutput{
		CreateTransactionOutput: &txcontract.CreateTransactionOutput{
			Transaction: *trx,
		},
		VoucherDiscount: voucherDiscount,
	}

	switch input.PayMethod {
	case "cash":
		err = s.txRepo.SettleTransaction(txcontract.SettleInput{
			TransactionID: txID,
			BranchID:      branchID,
			MemberID:      &input.MemberID,
			GrandTotal:    grandTotal,
			Payment: txdomain.Payment{
				Method:      "cash",
				Amount:      grandTotal,
				ReferenceNo: invoiceNo,
			},
		})
		if err != nil {
			return nil, err
		}
		resp.Message = "Pembayaran cash berhasil"
		trxPaid, _ := s.txRepo.GetTransactionByID(txID)
		resp.Transaction = *trxPaid
		// Push notifikasi langsung untuk cash payment
		if s.notifyFn != nil {
			_ = s.notifyFn(input.MemberID, "payment_success",
				"Pembayaran Berhasil!",
				fmt.Sprintf("Pesanan %s telah dikonfirmasi. Terima kasih!", trxPaid.InvoiceNo),
				map[string]interface{}{
					"transaction_id": trxPaid.ID,
					"invoice_no":     trxPaid.InvoiceNo,
					"grand_total":    trxPaid.GrandTotal,
				})
		}

	case "qris":
		qrisResp, qrURL, _, qrisErr := s.midtrans.CreateChargeQRIS(invoiceNo, grandTotal, "gopay")
		if qrisErr != nil {
			gopayResp, gopayURL, _, gopayErr := s.midtrans.CreateChargeGoPay(invoiceNo, grandTotal)
			if gopayErr != nil {
				_ = s.txRepo.CancelTransaction(txID)
				return nil, fmt.Errorf("gagal membuat QRIS: %w", gopayErr)
			}
			resp.QRString = gopayURL
			resp.MidtransTransactionID = gopayResp.TransactionID
			for _, a := range gopayResp.Actions {
				resp.MidtransActions = append(resp.MidtransActions, txcontract.MidtransAction{
					Name: a.Name, Method: a.Method, URL: a.URL,
				})
			}
		} else {
			resp.QRString = qrURL
			resp.MidtransTransactionID = qrisResp.TransactionID
			for _, a := range qrisResp.Actions {
				resp.MidtransActions = append(resp.MidtransActions, txcontract.MidtransAction{
					Name: a.Name, Method: a.Method, URL: a.URL,
				})
			}
		}
		resp.Message = "Scan QR untuk menyelesaikan pembayaran"

	case "va":
		vaResp, vaNumber, _, err := s.midtrans.CreateChargeBankTransfer(invoiceNo, grandTotal, "bca")
		if err != nil {
			_ = s.txRepo.CancelTransaction(txID)
			return nil, fmt.Errorf("gagal membuat Virtual Account: %w", err)
		}
		resp.VANumber = vaNumber
		resp.MidtransTransactionID = vaResp.TransactionID
		resp.Message = "Transfer ke VA untuk menyelesaikan pembayaran"

	case "ewallet":
		chargeResp, paymentURL, _, err := s.midtrans.CreateChargeGoPay(invoiceNo, grandTotal)
		if err != nil {
			_ = s.txRepo.CancelTransaction(txID)
			return nil, fmt.Errorf("gagal membuat e-wallet: %w", err)
		}
		resp.PaymentURL = paymentURL
		resp.MidtransTransactionID = chargeResp.TransactionID
		for _, a := range chargeResp.Actions {
			resp.MidtransActions = append(resp.MidtransActions, txcontract.MidtransAction{
				Name: a.Name, Method: a.Method, URL: a.URL,
			})
		}
		resp.Message = "Buka link untuk menyelesaikan pembayaran"

	default:
		return nil, fmt.Errorf("metode pembayaran tidak valid: %s", input.PayMethod)
	}

	// Tandai voucher sudah dipakai (best-effort)
	if voucherID != nil {
		// increment di voucher repository dilakukan di inject level via event;
		// untuk demo kita langsung update di sini lewat txRepo (tidak ada akses langsung ke voucher).
		// Kita simpan info voucherID di transaksi saja.
		_ = voucherID
	}

	return resp, nil
}

func (s *Service) GetTransactionsByMember(memberID int) ([]txdomain.Transaction, error) {
	return s.txRepo.GetTransactionsByMember(memberID)
}

func (s *Service) GetTransactionDetail(invoiceNo string, memberID int) (*shopcontract.TransactionDetail, error) {
	trx, err := s.txRepo.GetTransactionByInvoice(invoiceNo)
	if err != nil {
		return nil, fmt.Errorf("transaksi tidak ditemukan")
	}
	if trx.MemberID == nil || *trx.MemberID != memberID {
		return nil, fmt.Errorf("akses ditolak")
	}
	items, err := s.txRepo.GetItemsByTransactionID(trx.ID)
	if err != nil {
		return nil, err
	}
	return &shopcontract.TransactionDetail{Transaction: *trx, Items: items}, nil
}

func (s *Service) CheckPaymentStatus(invoiceNo string, memberID int) (string, error) {
	trx, err := s.txRepo.GetTransactionByInvoice(invoiceNo)
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
	mtStatus, err := s.midtrans.GetTransactionStatus(trx.InvoiceNo)
	if err != nil {
		// Midtrans tidak kenal order ini (misal cash sudah settle sebelumnya), kembalikan status DB
		return trx.Status, nil
	}

	// Jika Midtrans sudah settlement dan DB masih pending → settle manual
	if (mtStatus.TransactionStatus == "settlement" || mtStatus.TransactionStatus == "capture") &&
		(mtStatus.FraudStatus == "" || mtStatus.FraudStatus == "accept") {

		const branchID = 1
		paidAmount := trx.GrandTotal
		_ = s.txRepo.SettleTransaction(txcontract.SettleInput{
			TransactionID: trx.ID,
			BranchID:      branchID,
			MemberID:      trx.MemberID,
			GrandTotal:    trx.GrandTotal,
			Payment: txdomain.Payment{
				Method:      mapMidtransMethod(mtStatus.PaymentType),
				Amount:      paidAmount,
				ReferenceNo: mtStatus.TransactionID,
			},
		})

		// Push notifikasi
		if s.notifyFn != nil {
			_ = s.notifyFn(memberID, "payment_success",
				"Pembayaran Berhasil!",
				fmt.Sprintf("Pesanan %s telah terkonfirmasi. Terima kasih!", trx.InvoiceNo),
				map[string]interface{}{
					"transaction_id": trx.ID,
					"invoice_no":     trx.InvoiceNo,
					"grand_total":    trx.GrandTotal,
				})
		}
		return "paid", nil
	}

	// Kalau Midtrans bilang cancel/expire
	if mtStatus.TransactionStatus == "cancel" || mtStatus.TransactionStatus == "expire" {
		_ = s.txRepo.CancelTransaction(trx.ID)
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

func ptrInt(v int) *int {
	if v == 0 {
		return nil
	}
	return &v
}
