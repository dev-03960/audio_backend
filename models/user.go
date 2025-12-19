package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user in the system
type User struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	FullName    string             `bson:"fullName" json:"fullName"`
	PhoneNumber string             `bson:"phoneNumber" json:"phoneNumber"`
	Password    string             `bson:"password" json:"password,omitempty"`
	IsAdmin     bool               `bson:"isAdmin" json:"isAdmin"`
	IsBlocked   bool               `bson:"isBlocked" json:"isBlocked"`
	IsVerified  bool               `bson:"isVerified" json:"isVerified"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`
}
