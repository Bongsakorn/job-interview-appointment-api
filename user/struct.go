package user

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DataWithPassword struct
type DataWithPassword struct {
	ID       primitive.ObjectID `bson:"_id" json:"_id"`
	Email    string             `bson:"email" json:"email"`
	Password string             `bson:"password" json:"password"`
}

// User struct
type User struct {
	ID        primitive.ObjectID `bson:"_id" json:"_id"`
	Email     string             `bson:"email" json:"email"`
	Name      string             `bson:"name" json:"name"`
	AvatarURL string             `bson:"avatar_url" json:"avatar_url"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
