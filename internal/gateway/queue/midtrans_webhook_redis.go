package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	txcontract "github.com/fathanazka354/pos-koperasi/internal/usecase/transaction/dto"
	"github.com/redis/go-redis/v9"
)

const (
	StreamMidtransWebhooks = "midtrans:webhooks"
	GroupMidtransWorkers   = "midtrans-webhook-workers"
)

// MidtransWebhookDB memproses notifikasi Midtrans (sama dengan flow sinkron sebelumnya).
type MidtransWebhookDB interface {
	HandleMidtransNotification(payload txcontract.MidtransNotification) error
}

// RedisMidtransQueue menerbitkan payload webhook ke Redis Stream dan menjalankan consumer.
type RedisMidtransQueue struct {
	rdb *redis.Client
}

// NewRedisMidtransQueue membuat antrian. rdb wajib non-nil.
func NewRedisMidtransQueue(rdb *redis.Client) *RedisMidtransQueue {
	return &RedisMidtransQueue{rdb: rdb}
}

// Enqueue menambahkan notifikasi terverifikasi ke stream.
func (q *RedisMidtransQueue) Enqueue(ctx context.Context, p txcontract.MidtransNotification) error {
	b, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("marshal webhook: %w", err)
	}
	if err := q.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: StreamMidtransWebhooks,
		MaxLen: StreamApproxMaxLen,
		Approx: true,
		Values: map[string]interface{}{"payload": string(b)},
	}).Err(); err != nil {
		return fmt.Errorf("xadd: %w", err)
	}
	return nil
}

// RunConsumer memproses stream hingga ctx dibatalkan. Aman dipanggil dari goroutine.
func (q *RedisMidtransQueue) RunConsumer(ctx context.Context, svc MidtransWebhookDB) {
	consumerName := consumerID()
	if err := q.ensureGroup(ctx); err != nil {
		log.Printf("queue: cannot create consumer group: %v", err)
		return
	}
	log.Printf("queue: midtrans webhook consumer %q started (stream=%s)", consumerName, StreamMidtransWebhooks)

	for {
		select {
		case <-ctx.Done():
			log.Printf("queue: midtrans webhook consumer stopped")
			return
		default:
		}

		res, err := q.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    GroupMidtransWorkers,
			Consumer: consumerName,
			Streams:  []string{StreamMidtransWebhooks, ">"},
			Count:    20,
			Block:    3 * time.Second,
		}).Result()
		if err != nil {
			if err == context.Canceled || err == context.DeadlineExceeded {
				return
			}
			// timeout block → redis returns no error, empty
			if isTimeoutErr(err) {
				continue
			}
			log.Printf("queue: XReadGroup: %v", err)
			time.Sleep(time.Second)
			continue
		}
		for _, st := range res {
			for _, msg := range st.Messages {
				q.handleOne(ctx, svc, st.Stream, msg)
			}
		}
	}
}

func (q *RedisMidtransQueue) ensureGroup(ctx context.Context) error {
	err := q.rdb.XGroupCreateMkStream(ctx, StreamMidtransWebhooks, GroupMidtransWorkers, "0").Err()
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "BUSYGROUP") {
		return nil
	}
	return err
}

func (q *RedisMidtransQueue) handleOne(ctx context.Context, svc MidtransWebhookDB, stream string, msg redis.XMessage) {
	rawVal, ok := msg.Values["payload"]
	if !ok {
		log.Printf("queue: missing payload id=%s", msg.ID)
		_ = q.rdb.XAck(ctx, stream, GroupMidtransWorkers, msg.ID)
		return
	}
	raw, ok := rawVal.(string)
	if !ok {
		log.Printf("queue: payload not string id=%s", msg.ID)
		_ = q.rdb.XAck(ctx, stream, GroupMidtransWorkers, msg.ID)
		return
	}
	var p txcontract.MidtransNotification
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		log.Printf("queue: bad payload id=%s: %v", msg.ID, err)
		_ = q.rdb.XAck(ctx, stream, GroupMidtransWorkers, msg.ID)
		return
	}
	if err := svc.HandleMidtransNotification(p); err != nil {
		log.Printf("queue: handle order_id=%s: %v (midtrans may retry)", p.OrderID, err)
	}
	_ = q.rdb.XAck(ctx, stream, GroupMidtransWorkers, msg.ID)
}

func consumerID() string {
	h, _ := os.Hostname()
	if h == "" {
		h = "local"
	}
	return fmt.Sprintf("%s-%d", h, os.Getpid())
}

func isTimeoutErr(err error) bool {
	// go-redis: context deadline on Block
	if err == context.DeadlineExceeded {
		return true
	}
	// i/o timeout
	if s := err.Error(); strings.Contains(s, "i/o timeout") || strings.Contains(s, "timeout") {
		return true
	}
	return false
}

// NewRedisFromConfig membuat klien Redis; dipakai saat REDIS_ADDR diset.
func NewRedisFromConfig(addr, password string, db int) *redis.Client {
	if addr == "" {
		return nil
	}
	return redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
}

