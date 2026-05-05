package notification

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Notification struct {
	ID        primitive.ObjectID     `bson:"_id,omitempty"  json:"id"`
	MemberID  int                    `bson:"member_id"      json:"member_id"`
	Type      string                 `bson:"type"           json:"type"`
	Title     string                 `bson:"title"          json:"title"`
	Body      string                 `bson:"body"           json:"body"`
	Data      map[string]interface{} `bson:"data,omitempty" json:"data,omitempty"`
	IsRead    bool                   `bson:"is_read"        json:"is_read"`
	CreatedAt time.Time              `bson:"created_at"     json:"created_at"`
}

