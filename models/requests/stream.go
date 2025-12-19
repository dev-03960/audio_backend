package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type CreateStreamRequest struct {
	Name          string `json:"name" binding:"required"`
	Thumbnail     string `json:"thumbnail"`
	StreamUrl     string `json:"streamUrl" binding:"required"`
	IsLive        bool   `json:"isLive"`
	DisplayOnSite bool   `json:"displayOnSite"`
}

type UpdateStreamRequest struct {
	Name          string `json:"name"`
	Thumbnail     string `json:"thumbnail"`
	StreamUrl     string `json:"streamUrl"`
	IsLive        *bool  `json:"isLive"`
	DisplayOnSite *bool  `json:"displayOnSite"`
}

type AddCommentRequest struct {
	AudiobookID primitive.ObjectID `json:"audiobookId" binding:"required"` // Changed from StreamID
	Message     string             `json:"message" binding:"required"`
	IsAdmin     bool               `json:"isAdmin"`
}

type SiteChangesRequest struct {
	Site           string             `json:"site"`
	Logourl        string             `json:"logourl"`
	Calendarurl    string             `json:"calendarurl"`
	Notification   string             `json:"notification"`
	InviteOnlyMode bool               `json:"inviteonlymode"`
	ImageSlider    []ImgSliderRequest `json:"imageslider"`
	BuzzingText    string             `json:"buzzingtext"` // NEW
	NowText        string             `json:"nowtext"`     // NEW
	LiveTag        string             `json:"livetag"`     // NEW
}

type ImgSliderRequest struct {
	Title      string `json:"title"`
	Subtitle   string `json:"subtitle"`
	Image      string `json:"image"`
	Link       string `json:"link"`
	ButtonName string `json:"buttonname"`
}
