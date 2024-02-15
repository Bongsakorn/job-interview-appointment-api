package post

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Post struct
type Post struct {
	ID          primitive.ObjectID `bson:"_id" json:"_id"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description" json:"description"`
	Status      string             `bson:"status" json:"status"`
	Archived    bool               `bson:"archived" json:"archived"`
	CreatedBy   primitive.ObjectID `bson:"created_by" json:"created_by"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// RequestInput struct
type RequestInput struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description"`
}
