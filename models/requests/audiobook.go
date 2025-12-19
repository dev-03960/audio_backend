package models

// Audiobook requests
type CreateAudiobookRequest struct {
	Name          string `json:"name" binding:"required"`
	Description   string `json:"description" binding:"required"`
	AudioData     string `json:"audioData" binding:"required"` // Base64 encoded audio or file path
	Thumbnail     string `json:"thumbnail"`
	Content       string `json:"content"` // Transcription/content
	DisplayOnSite bool   `json:"displayOnSite"`
}

type UpdateAudiobookRequest struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	AudioData     string `json:"audioData"`
	Thumbnail     string `json:"thumbnail"`
	Content       string `json:"content"`
	DisplayOnSite *bool  `json:"displayOnSite"`
}

type LikeDislikeRequest struct {
	Action string `json:"action" binding:"required"` // "like" or "dislike"
}
