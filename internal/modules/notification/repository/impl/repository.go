package impl

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/yourname/pos-koperasi/internal/modules/notification/contract"
	"github.com/yourname/pos-koperasi/internal/modules/notification/domain"
)

type Repository struct {
	col *mongo.Collection
}

func New(db *mongo.Database) *Repository {
	return &Repository{col: db.Collection("notifications")}
}

var _ contract.NotificationRepository = (*Repository)(nil)

func (r *Repository) Create(n domain.Notification) (*domain.Notification, error) {
	n.ID = primitive.NewObjectID()
	if n.CreatedAt.IsZero() {
		n.CreatedAt = time.Now()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := r.col.InsertOne(ctx, n)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *Repository) ListByMember(memberID int, limit int) ([]domain.Notification, error) {
	if limit <= 0 {
		limit = 50
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit))

	cur, err := r.col.Find(ctx, bson.M{"member_id": memberID}, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var result []domain.Notification
	if err := cur.All(ctx, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *Repository) MarkRead(id string, memberID int) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = r.col.UpdateOne(ctx,
		bson.M{"_id": oid, "member_id": memberID},
		bson.M{"$set": bson.M{"is_read": true}},
	)
	return err
}

func (r *Repository) MarkAllRead(memberID int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := r.col.UpdateMany(ctx,
		bson.M{"member_id": memberID, "is_read": false},
		bson.M{"$set": bson.M{"is_read": true}},
	)
	return err
}

func (r *Repository) CountUnread(memberID int) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	n, err := r.col.CountDocuments(ctx, bson.M{"member_id": memberID, "is_read": false})
	return int(n), err
}
