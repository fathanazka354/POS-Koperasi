package outbox

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/fathanazka354/pos-koperasi/internal/gateway/queue"
	notifcontract "github.com/fathanazka354/pos-koperasi/internal/usecase/notification"
)

type memberPaymentNotifyPayload struct {
	MemberID int                    `json:"member_id"`
	Type     string                 `json:"type"`
	Title    string                 `json:"title"`
	Body     string                 `json:"body"`
	Data     map[string]interface{} `json:"data,omitempty"`
}

// RunPaymentNotifyWorker mengklaim dokumen outbox (Mongo) lalu publish notifikasi (Redis atau sinkron).
func RunPaymentNotifyWorker(
	ctx context.Context,
	store PaymentOutboxStore,
	nq *queue.RedisMemberNotifyQueue,
	notifSvc notifcontract.Usecase,
) {
	log.Printf("outbox: worker payment-notify started (Mongo / claim)")
	for {
		select {
		case <-ctx.Done():
			log.Printf("outbox: worker payment-notify stopped")
			return
		default:
		}

		rows, err := store.ClaimPending(ctx, 40)
		if err != nil {
			log.Printf("outbox: pull events: %v", err)
			time.Sleep(time.Second)
			continue
		}
		if len(rows) == 0 {
			time.Sleep(300 * time.Millisecond)
			continue
		}

		for _, row := range rows {
			if row.EventType != "member_payment_notify" {
				if err := store.MarkPublished(ctx, row.ID); err != nil {
					log.Printf("outbox: skip unknown event_type=%s id=%s mark: %v", row.EventType, row.ID.Hex(), err)
				}
				continue
			}

			var p memberPaymentNotifyPayload
			if err := json.Unmarshal(row.Payload, &p); err != nil {
				log.Printf("outbox: bad payload id=%s: %v", row.ID.Hex(), err)
				if err := store.MarkPublished(ctx, row.ID); err != nil {
					log.Printf("outbox: mark published id=%s: %v", row.ID.Hex(), err)
				}
				continue
			}

			queue.DispatchMemberNotify(ctx, nq, notifSvc.Notify, p.MemberID, p.Type, p.Title, p.Body, p.Data)

			if err := store.MarkPublished(ctx, row.ID); err != nil {
				log.Printf("outbox: mark published id=%s: %v", row.ID.Hex(), err)
			}
		}
	}
}

