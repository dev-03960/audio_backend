package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Audiobook struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name          string             `bson:"name" json:"name"`
	Description   string             `bson:"description" json:"description"`
	AudioData     string             `bson:"audioData" json:"audioData"`         // Base64 encoded audio or file path
	Thumbnail     string             `bson:"thumbnail" json:"thumbnail"`         // Predefined thumbnail name or base64/URL
	Content       string             `bson:"content" json:"content"`             // Transcription/content of the audiobook
	ViewCount     int                `bson:"viewCount" json:"viewCount"`         // Total views
	Likes         int                `bson:"likes" json:"likes"`                 // Like count
	Dislikes      int                `bson:"dislikes" json:"dislikes"`           // Dislike count
	DisplayOnSite bool               `bson:"displayOnSite" json:"displayOnSite"` // Visibility flag
	CreatedAt     time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt     time.Time          `bson:"updatedAt" json:"updatedAt"`
}
