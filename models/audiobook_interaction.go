package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AudiobookInteraction struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	AudiobookID primitive.ObjectID `bson:"audiobookId" json:"audiobookId"` // Reference to audiobook
	UserID      primitive.ObjectID `bson:"userId" json:"userId"`           // Reference to user
	Action      string             `bson:"action" json:"action"`           // "like", "dislike", "view"
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
}
