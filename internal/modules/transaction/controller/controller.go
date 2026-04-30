package controller

import (
	"encoding/json"
	"net/http"

	"github.com/yourname/pos-koperasi/internal/middleware"
	"github.com/yourname/pos-koperasi/internal/midtrans"
	"github.com/yourname/pos-koperasi/internal/modules/transaction/contract"
	"github.com/yourname/pos-koperasi/internal/modules/transaction/controller/dto"
	"github.com/yourname/pos-koperasi/pkg/response"
)

type Controller struct {
	svc               contract.TransactionService
	midtransServerKey string
}

func New(svc contract.TransactionService, midtransServerKey string) *Controller {
	return &Controller{svc: svc, midtransServerKey: midtransServerKey}
}

// POST /api/v1/transactions
func (c *Controller) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if len(req.Items) == 0 {
		response.BadRequest(w, "Items tidak boleh kosong")
		return
	}
	if req.PayMethod == "" {
		response.BadRequest(w, "Metode pembayaran wajib diisi (cash/qris/va/ewallet)")
		return
	}

	claims := middleware.GetClaims(r)
	if claims == nil {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	if req.BranchID == 0 {
		req.BranchID = claims.BranchID
	}

	items := make([]contract.TransactionItemInput, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, contract.TransactionItemInput{ProductID: it.ProductID, Quantity: it.Quantity})
	}

	result, err := c.svc.CreateTransaction(contract.CreateTransactionInput{
		BranchID:   req.BranchID,
		RegisterID: req.RegisterID,
		MemberCode: req.MemberCode,
		Items:      items,
		PromoCode:  req.PromoCode,
		PayMethod:  req.PayMethod,
	}, claims.EmployeeID)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	response.Created(w, result.Message, result)
}

// POST /api/v1/midtrans/webhook
func (c *Controller) MidtransWebhook(w http.ResponseWriter, r *http.Request) {
	var payload midtrans.NotificationPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "invalid_payload"})
		return
	}

	if err := midtrans.VerifySignature(payload, c.midtransServerKey); err != nil {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "invalid_signature"})
		return
	}

	if err := c.svc.HandleMidtransNotification(payload); err != nil {
		_ = err
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "error_logged"})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "received"})
}

