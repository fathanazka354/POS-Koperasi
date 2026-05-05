package outbox

import (
	"context"

	"github.com/fathanazka354/pos-koperasi/internal/gateway/queue"
	notifcontract "github.com/fathanazka354/pos-koperasi/internal/usecase/notification"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DirectPaymentOutboxStore tanpa Mongo: enqueue = langsung DispatchMemberNotify (dev / fallback).
type DirectPaymentOutboxStore struct {
	nq    *queue.RedisMemberNotifyQueue
	notif notifcontract.Usecase
}

func NewDirectPaymentOutboxStore(nq *queue.RedisMemberNotifyQueue, notif notifcontract.Usecase) *DirectPaymentOutboxStore {
	return &DirectPaymentOutboxStore{nq: nq, notif: notif}
}

var _ PaymentOutboxStore = (*DirectPaymentOutboxStore)(nil)

func (s *DirectPaymentOutboxStore) EnqueueMemberPayment(ctx context.Context, _ int64, memberID int, m *MemberNotifyOutbox) error {
	if m == nil {
		return nil
	}
	queue.DispatchMemberNotify(ctx, s.nq, s.notif.Notify, memberID, m.Type, m.Title, m.Body, m.Data)
	return nil
}

func (s *DirectPaymentOutboxStore) ClaimPending(_ context.Context, _ int) ([]PaymentOutboxDoc, error) {
	return nil, nil
}

func (s *DirectPaymentOutboxStore) MarkPublished(_ context.Context, _ primitive.ObjectID) error {
	return nil
}

