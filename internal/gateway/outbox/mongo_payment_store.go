package outbox

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	collectionOutboxEvents = "outbox_events"
	staleClaimMinutes      = 15
)

type mongoPaymentDoc struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	EventType     string             `bson:"event_type"`
	TransactionID int64              `bson:"transaction_id"`
	MemberID      int                `bson:"member_id"`
	Payload       []byte             `bson:"payload"`
	CreatedAt     time.Time          `bson:"created_at"`
	PublishedAt   *time.Time         `bson:"published_at,omitempty"`
	ClaimedAt     *time.Time         `bson:"claimed_at,omitempty"`
}

// MongoPaymentOutboxStore menyimpan event outbox di MongoDB (log + referensi transaction_id / member_id).
type MongoPaymentOutboxStore struct {
	col *mongo.Collection
}

func NewMongoPaymentOutboxStore(db *mongo.Database) *MongoPaymentOutboxStore {
	s := &MongoPaymentOutboxStore{col: db.Collection(collectionOutboxEvents)}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, _ = s.col.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "published_at", Value: 1}, {Key: "_id", Value: 1}}},
		{Keys: bson.D{{Key: "transaction_id", Value: 1}}},
		{Keys: bson.D{{Key: "member_id", Value: 1}}},
	})
	return s
}

var _ PaymentOutboxStore = (*MongoPaymentOutboxStore)(nil)

func (s *MongoPaymentOutboxStore) EnqueueMemberPayment(ctx context.Context, transactionID int64, memberID int, m *MemberNotifyOutbox) error {
	if m == nil {
		return nil
	}
	inner := struct {
		MemberID int                    `json:"member_id"`
		Type     string                 `json:"type"`
		Title    string                 `json:"title"`
		Body     string                 `json:"body"`
		Data     map[string]interface{} `json:"data,omitempty"`
	}{
		MemberID: memberID,
		Type:     m.Type,
		Title:    m.Title,
		Body:     m.Body,
		Data:     m.Data,
	}
	b, err := json.Marshal(inner)
	if err != nil {
		return err
	}
	doc := mongoPaymentDoc{
		ID:            primitive.NewObjectID(),
		EventType:     "member_payment_notify",
		TransactionID: transactionID,
		MemberID:      memberID,
		Payload:       b,
		CreatedAt:     time.Now(),
	}
	_, err = s.col.InsertOne(ctx, doc)
	return err
}

func (s *MongoPaymentOutboxStore) ClaimPending(ctx context.Context, limit int) ([]PaymentOutboxDoc, error) {
	if limit <= 0 {
		limit = 40
	}
	stale := time.Now().Add(-staleClaimMinutes * time.Minute)
	filter := bson.M{
		"$and": []interface{}{
			bson.M{"$or": []interface{}{
				bson.M{"published_at": bson.M{"$exists": false}},
				bson.M{"published_at": nil},
			}},
			bson.M{"$or": []interface{}{
				bson.M{"claimed_at": bson.M{"$exists": false}},
				bson.M{"claimed_at": nil},
				bson.M{"claimed_at": bson.M{"$lt": stale}},
			}},
		},
	}
	opts := options.FindOneAndUpdate().
		SetReturnDocument(options.After).
		SetSort(bson.D{{Key: "_id", Value: 1}})

	var out []PaymentOutboxDoc
	for i := 0; i < limit; i++ {
		now := time.Now()
		update := bson.M{"$set": bson.M{"claimed_at": now}}
		var raw mongoPaymentDoc
		err := s.col.FindOneAndUpdate(ctx, filter, update, opts).Decode(&raw)
		if err == mongo.ErrNoDocuments {
			break
		}
		if err != nil {
			return out, fmt.Errorf("claim outbox: %w", err)
		}
		out = append(out, PaymentOutboxDoc{
			ID:        raw.ID,
			EventType: raw.EventType,
			Payload:   json.RawMessage(raw.Payload),
		})
	}
	return out, nil
}

func (s *MongoPaymentOutboxStore) MarkPublished(ctx context.Context, id primitive.ObjectID) error {
	now := time.Now()
	_, err := s.col.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"published_at": now, "claimed_at": nil}},
	)
	return err
}

