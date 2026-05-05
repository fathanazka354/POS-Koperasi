package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	StreamMemberNotifications = "notifications:member"
	GroupMemberNotifyWorkers  = "member-notify-workers"
)

// MemberNotifyJob satu unit kerja untuk Notify(member).
type MemberNotifyJob struct {
	MemberID int                    `json:"member_id"`
	Type     string                 `json:"type"`
	Title    string                 `json:"title"`
	Body     string                 `json:"body"`
	Data     map[string]interface{} `json:"data,omitempty"`
}

// MemberNotifier mengirim notifikasi persist + WS (implementasi: NotificationService).
type MemberNotifier interface {
	Notify(memberID int, notifType, title, body string, data map[string]interface{}) error
}

// RedisMemberNotifyQueue antrian notifikasi member (Redis Stream).
type RedisMemberNotifyQueue struct {
	rdb *redis.Client
}

// NewRedisMemberNotifyQueue membuat antrian; rdb non-nil.
func NewRedisMemberNotifyQueue(rdb *redis.Client) *RedisMemberNotifyQueue {
	return &RedisMemberNotifyQueue{rdb: rdb}
}

// Enqueue menambahkan job ke stream.
func (q *RedisMemberNotifyQueue) Enqueue(ctx context.Context, j MemberNotifyJob) error {
	b, err := json.Marshal(j)
	if err != nil {
		return fmt.Errorf("marshal notify job: %w", err)
	}
	if err := q.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: StreamMemberNotifications,
		MaxLen: StreamApproxMaxLen,
		Approx: true,
		Values: map[string]interface{}{"payload": string(b)},
	}).Err(); err != nil {
		return fmt.Errorf("xadd notify: %w", err)
	}
	return nil
}

// RunConsumer memanggil Notify hingga ctx dibatalkan.
func (q *RedisMemberNotifyQueue) RunConsumer(ctx context.Context, svc MemberNotifier) {
	consumerName := consumerID()
	if err := q.ensureNotifyGroup(ctx); err != nil {
		log.Printf("queue: cannot create notify consumer group: %v", err)
		return
	}
	log.Printf("queue: member notify consumer %q started (stream=%s)", consumerName, StreamMemberNotifications)

	for {
		select {
		case <-ctx.Done():
			log.Printf("queue: member notify consumer stopped")
			return
		default:
		}

		res, err := q.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    GroupMemberNotifyWorkers,
			Consumer: consumerName,
			Streams:  []string{StreamMemberNotifications, ">"},
			Count:    50,
			Block:    3 * time.Second,
		}).Result()
		if err != nil {
			if err == context.Canceled || err == context.DeadlineExceeded {
				return
			}
			if isTimeoutErr(err) {
				continue
			}
			log.Printf("queue: member notify XReadGroup: %v", err)
			time.Sleep(time.Second)
			continue
		}
		for _, st := range res {
			for _, msg := range st.Messages {
				q.handleNotifyOne(ctx, svc, st.Stream, msg)
			}
		}
	}
}

func (q *RedisMemberNotifyQueue) ensureNotifyGroup(ctx context.Context) error {
	err := q.rdb.XGroupCreateMkStream(ctx, StreamMemberNotifications, GroupMemberNotifyWorkers, "0").Err()
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "BUSYGROUP") {
		return nil
	}
	return err
}

func (q *RedisMemberNotifyQueue) handleNotifyOne(ctx context.Context, svc MemberNotifier, stream string, msg redis.XMessage) {
	rawVal, ok := msg.Values["payload"]
	if !ok {
		log.Printf("queue: notify missing payload id=%s", msg.ID)
		_ = q.rdb.XAck(ctx, stream, GroupMemberNotifyWorkers, msg.ID)
		return
	}
	raw, ok := rawVal.(string)
	if !ok {
		log.Printf("queue: notify payload not string id=%s", msg.ID)
		_ = q.rdb.XAck(ctx, stream, GroupMemberNotifyWorkers, msg.ID)
		return
	}
	var j MemberNotifyJob
	if err := json.Unmarshal([]byte(raw), &j); err != nil {
		log.Printf("queue: bad notify job id=%s: %v", msg.ID, err)
		_ = q.rdb.XAck(ctx, stream, GroupMemberNotifyWorkers, msg.ID)
		return
	}
	if err := svc.Notify(j.MemberID, j.Type, j.Title, j.Body, j.Data); err != nil {
		log.Printf("queue: Notify member_id=%d failed: %v (akan di-ack; bisa kirim ulang dari sumber)", j.MemberID, err)
	}
	_ = q.rdb.XAck(ctx, stream, GroupMemberNotifyWorkers, msg.ID)
}

