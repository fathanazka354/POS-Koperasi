package outbox

import (
	"context"
	"encoding/json"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PaymentOutboxDoc satu dokumen event untuk worker (mirip baris outbox_events).
type PaymentOutboxDoc struct {
	ID        primitive.ObjectID
	EventType string
	Payload   json.RawMessage
}

// PaymentOutboxStore antrian log pembayaran → notifikasi (MongoDB atau jalur langsung jika tanpa Mongo).
type PaymentOutboxStore interface {
	EnqueueMemberPayment(ctx context.Context, transactionID int64, memberID int, m *MemberNotifyOutbox) error
	ClaimPending(ctx context.Context, limit int) ([]PaymentOutboxDoc, error)
	MarkPublished(ctx context.Context, id primitive.ObjectID) error
}

// MemberNotifyOutbox — payload notifikasi pembayaran; dipakai oleh store/outbox worker.
type MemberNotifyOutbox struct {
	Type  string
	Title string
	Body  string
	Data  map[string]interface{}
}

