package controllers

import (
	"context"
	"net/http"
	"time"

	models "live_stream/models"
	request "live_stream/models/requests"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AudiobookController struct {
	AudiobookCol   *mongo.Collection
	InteractionCol *mongo.Collection
}

// GetAudiobooks - public endpoint to list all visible audiobooks
func (ac *AudiobookController) GetAudiobooks(c *gin.Context) {
	cursor, err := ac.AudiobookCol.Find(context.TODO(), bson.M{"displayOnSite": true})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch audiobooks"})
		return
	}
	var audiobooks []models.Audiobook
	cursor.All(context.TODO(), &audiobooks)
	c.JSON(http.StatusOK, audiobooks)
}

// GetAudiobookByID - public endpoint to get audiobook details + increment view count
func (ac *AudiobookController) GetAudiobookByID(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid audiobook ID"})
		return
	}

	var audiobook models.Audiobook
	err = ac.AudiobookCol.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&audiobook)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Audiobook not found"})
		return
	}

	// Increment view count
	ac.AudiobookCol.UpdateOne(
		context.TODO(),
		bson.M{"_id": objID},
		bson.M{"$inc": bson.M{"viewCount": 1}},
	)

	// Update the viewCount in the response
	audiobook.ViewCount++

	c.JSON(http.StatusOK, audiobook)
}

// CreateAudiobook - admin endpoint to create audiobook
func (ac *AudiobookController) CreateAudiobook(c *gin.Context) {
	var req request.CreateAudiobookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	audiobook := models.Audiobook{
		Name:          req.Name,
		Description:   req.Description,
		AudioData:     req.AudioData,
		Thumbnail:     req.Thumbnail,
		Content:       req.Content,
		DisplayOnSite: req.DisplayOnSite,
		ViewCount:     0,
		Likes:         0,
		Dislikes:      0,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	result, err := ac.AudiobookCol.InsertOne(context.TODO(), audiobook)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create audiobook"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Audiobook created", "id": result.InsertedID})
}

// UpdateAudiobook - admin endpoint to update audiobook
func (ac *AudiobookController) UpdateAudiobook(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid audiobook ID"})
		return
	}

	var req request.UpdateAudiobookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	update := bson.M{}
	if req.Name != "" {
		update["name"] = req.Name
	}
	if req.Description != "" {
		update["description"] = req.Description
	}
	if req.AudioData != "" {
		update["audioData"] = req.AudioData
	}
	if req.Thumbnail != "" {
		update["thumbnail"] = req.Thumbnail
	}
	if req.Content != "" {
		update["content"] = req.Content
	}
	if req.DisplayOnSite != nil {
		update["displayOnSite"] = *req.DisplayOnSite
	}
	update["updatedAt"] = time.Now()

	result, err := ac.AudiobookCol.UpdateOne(
		context.TODO(),
		bson.M{"_id": objID},
		bson.M{"$set": update},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update audiobook"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Audiobook not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Audiobook updated"})
}

// DeleteAudiobook - admin endpoint to delete audiobook
func (ac *AudiobookController) DeleteAudiobook(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid audiobook ID"})
		return
	}

	result, err := ac.AudiobookCol.DeleteOne(context.TODO(), bson.M{"_id": objID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete audiobook"})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Audiobook not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Audiobook deleted"})
}

// LikeAudiobook - user endpoint to like audiobook
func (ac *AudiobookController) LikeAudiobook(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid audiobook ID"})
		return
	}

	userID := c.GetString("user_id")
	uid, _ := primitive.ObjectIDFromHex(userID)

	// Check if user has already liked
	var existing models.AudiobookInteraction
	err = ac.InteractionCol.FindOne(context.TODO(), bson.M{
		"audiobookId": objID,
		"userId":      uid,
		"action":      "like",
	}).Decode(&existing)

	if err == mongo.ErrNoDocuments {
		// User hasn't liked yet, add like
		interaction := models.AudiobookInteraction{
			AudiobookID: objID,
			UserID:      uid,
			Action:      "like",
			CreatedAt:   time.Now(),
		}
		ac.InteractionCol.InsertOne(context.TODO(), interaction)

		// Check and remove dislike if exists
		ac.InteractionCol.DeleteOne(context.TODO(), bson.M{
			"audiobookId": objID,
			"userId":      uid,
			"action":      "dislike",
		})

		// Update audiobook counts
		ac.AudiobookCol.UpdateOne(
			context.TODO(),
			bson.M{"_id": objID},
			bson.M{"$inc": bson.M{"likes": 1}},
		)

		// Decrement dislike if it existed
		var audiobook models.Audiobook
		ac.AudiobookCol.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&audiobook)
		if audiobook.Dislikes > 0 {
			ac.AudiobookCol.UpdateOne(
				context.TODO(),
				bson.M{"_id": objID},
				bson.M{"$inc": bson.M{"dislikes": -1}},
			)
		}

		c.JSON(http.StatusOK, gin.H{"message": "Audiobook liked"})
	} else {
		// User already liked, remove like
		ac.InteractionCol.DeleteOne(context.TODO(), bson.M{
			"audiobookId": objID,
			"userId":      uid,
			"action":      "like",
		})

		ac.AudiobookCol.UpdateOne(
			context.TODO(),
			bson.M{"_id": objID},
			bson.M{"$inc": bson.M{"likes": -1}},
		)

		c.JSON(http.StatusOK, gin.H{"message": "Like removed"})
	}
}

// DislikeAudiobook - user endpoint to dislike audiobook
func (ac *AudiobookController) DislikeAudiobook(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid audiobook ID"})
		return
	}

	userID := c.GetString("user_id")
	uid, _ := primitive.ObjectIDFromHex(userID)

	// Check if user has already disliked
	var existing models.AudiobookInteraction
	err = ac.InteractionCol.FindOne(context.TODO(), bson.M{
		"audiobookId": objID,
		"userId":      uid,
		"action":      "dislike",
	}).Decode(&existing)

	if err == mongo.ErrNoDocuments {
		// User hasn't disliked yet, add dislike
		interaction := models.AudiobookInteraction{
			AudiobookID: objID,
			UserID:      uid,
			Action:      "dislike",
			CreatedAt:   time.Now(),
		}
		ac.InteractionCol.InsertOne(context.TODO(), interaction)

		// Check and remove like if exists
		ac.InteractionCol.DeleteOne(context.TODO(), bson.M{
			"audiobookId": objID,
			"userId":      uid,
			"action":      "like",
		})

		// Update audiobook counts
		ac.AudiobookCol.UpdateOne(
			context.TODO(),
			bson.M{"_id": objID},
			bson.M{"$inc": bson.M{"dislikes": 1}},
		)

		// Decrement like if it existed
		var audiobook models.Audiobook
		ac.AudiobookCol.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&audiobook)
		if audiobook.Likes > 0 {
			ac.AudiobookCol.UpdateOne(
				context.TODO(),
				bson.M{"_id": objID},
				bson.M{"$inc": bson.M{"likes": -1}},
			)
		}

		c.JSON(http.StatusOK, gin.H{"message": "Audiobook disliked"})
	} else {
		// User already disliked, remove dislike
		ac.InteractionCol.DeleteOne(context.TODO(), bson.M{
			"audiobookId": objID,
			"userId":      uid,
			"action":      "dislike",
		})

		ac.AudiobookCol.UpdateOne(
			context.TODO(),
			bson.M{"_id": objID},
			bson.M{"$inc": bson.M{"dislikes": -1}},
		)

		c.JSON(http.StatusOK, gin.H{"message": "Dislike removed"})
	}
}

// GetAudiobookStats - public endpoint to get like/dislike/view counts
func (ac *AudiobookController) GetAudiobookStats(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid audiobook ID"})
		return
	}

	var audiobook models.Audiobook
	err = ac.AudiobookCol.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&audiobook)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Audiobook not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"viewCount": audiobook.ViewCount,
		"likes":     audiobook.Likes,
		"dislikes":  audiobook.Dislikes,
	})
}
