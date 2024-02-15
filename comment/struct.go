package comment

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Comment struct
type Comment struct {
	ID        primitive.ObjectID `bson:"_id" json:"_id"`
	PostID    primitive.ObjectID `bson:"post_id" json:"post_id"`
	Message   string             `bson:"message" json:"message"`
	CreatedBy primitive.ObjectID `bson:"created_by" json:"created_by"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}
