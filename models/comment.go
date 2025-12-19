package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Comment struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	AudiobookID primitive.ObjectID `bson:"audiobookId" json:"audiobookId"` // Changed from StreamID
	UserID      primitive.ObjectID `bson:"userId" json:"userId"`
	UserName    string             `bson:"userName" json:"userName"`
	Message     string             `bson:"message" json:"message"`
	IsAdmin     bool               `bson:"isAdmin" json:"isAdmin"`
	Timestamp   time.Time          `bson:"timestamp" json:"timestamp"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
	IsDeleted   bool               `bson:"isDeleted" json:"isDeleted"`
}
