package activitylog

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ActivityLog structure
type ActivityLog struct {
	ID        primitive.ObjectID `bson:"_id" json:"_id"`
	PostID    primitive.ObjectID `bson:"post_id" json:"post_id"`
	Actor     primitive.ObjectID `bson:"actor" json:"actor"`
	Action    string             `bson:"action" json:"action"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
