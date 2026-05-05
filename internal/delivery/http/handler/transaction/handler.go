package transaction

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/fathanazka354/pos-koperasi/internal/middleware"
	"github.com/fathanazka354/pos-koperasi/internal/midtrans"
	"github.com/fathanazka354/pos-koperasi/internal/delivery/http/handler/transaction/dto"
	"github.com/fathanazka354/pos-koperasi/internal/gateway/queue"
	"github.com/fathanazka354/pos-koperasi/internal/usecase/transaction"
	"github.com/fathanazka354/pos-koperasi/pkg/response"
)

type Handler struct {
	uc               transaction.Usecase
	midtransServerKey string
	webhookQueue      *queue.RedisMidtransQueue
}

func New(uc transaction.Usecase, midtransServerKey string, webhookQ *queue.RedisMidtransQueue) *Handler {
	return &Handler{uc: uc, midtransServerKey: midtransServerKey, webhookQueue: webhookQ}
}

// POST /api/v1/transactions
func (h *Handler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
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

	items := make([]transaction.TransactionItemInput, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, transaction.TransactionItemInput{ProductID: it.ProductID, Quantity: it.Quantity})
	}

	result, err := h.uc.CreateTransaction(transaction.CreateTransactionInput{
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
func (h *Handler) MidtransWebhook(w http.ResponseWriter, r *http.Request) {
	var payload midtrans.NotificationPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "invalid_payload"})
		return
	}

	if err := midtrans.VerifySignature(payload, h.midtransServerKey); err != nil {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "invalid_signature"})
		return
	}

	if h.webhookQueue != nil {
		n := midtrans.MapToTransactionNotification(payload)
		if err := h.webhookQueue.Enqueue(r.Context(), n); err != nil {
			log.Printf("midtrans webhook: enqueue gagal, fallback sinkron: %v", err)
			if err := h.uc.HandleMidtransNotification(n); err != nil {
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]string{"status": "error_logged"})
				return
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{"status": "received"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "queued"})
		return
	}

	if err := h.uc.HandleMidtransNotification(midtrans.MapToTransactionNotification(payload)); err != nil {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "error_logged"})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "received"})
}

