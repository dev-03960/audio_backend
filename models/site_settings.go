package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ImgSlider struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Title      string             `bson:"title" json:"title"`
	Subtitle   string             `bson:"subtitle" json:"subtitle"`
	Image      string             `bson:"image" json:"image"`
	Link       string             `bson:"link" json:"link"`
	ButtonName string             `bson:"buttonname" json:"buttonname"`
}

type SiteChanges struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Site           string             `bson:"site" json:"site"`
	LogoUrl        string             `bson:"logourl" json:"logourl"`
	Calendarurl    string             `bson:"calendarurl" json:"calendarurl"`
	Notification   string             `bson:"notification" json:"notification"`
	InviteOnlyMode bool               `bson:"inviteonlymode" json:"inviteonlymode"`
	ImageSlider    []ImgSlider        `bson:"imageslider" json:"imageslider"`
	BuzzingText    string             `bson:"buzzingtext" json:"buzzingtext"`
	NowText        string             `bson:"nowtext" json:"nowtext"`
	LiveTag        string             `bson:"livetag" json:"livetag"`
	CreatedAt      time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt      time.Time          `bson:"updatedAt" json:"updatedAt"`
}
