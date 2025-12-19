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

type CommentController struct {
	CommentCol *mongo.Collection
	UserCol    *mongo.Collection
}

// AddComment - authenticated users
func (cc *CommentController) AddComment(c *gin.Context) {
	var req request.AddCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	uid, _ := primitive.ObjectIDFromHex(userID)

	var user models.User
	err := cc.UserCol.FindOne(context.TODO(), bson.M{"_id": uid}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user info"})
		return
	}

	now := time.Now()
	comment := models.Comment{
		AudiobookID: req.AudiobookID, // Changed from StreamID
		UserID:      uid,
		UserName:    user.FullName,
		Message:     req.Message,
		IsAdmin:     req.IsAdmin,
		Timestamp:   now,
		UpdatedAt:   now,
		IsDeleted:   false,
	}

	cc.CommentCol.InsertOne(context.TODO(), comment)
	c.JSON(http.StatusOK, gin.H{"message": "Comment added"})
}

// GetComments - public
func (cc *CommentController) GetComments(c *gin.Context) {
	audiobookID := c.Param("id") // Changed from streamID
	objID, _ := primitive.ObjectIDFromHex(audiobookID)

	cursor, _ := cc.CommentCol.Find(context.TODO(), bson.M{"audiobookId": objID, "isDeleted": false})
	var comments []models.Comment
	cursor.All(context.TODO(), &comments)
	c.JSON(http.StatusOK, comments)
}

func (cc *CommentController) DeleteComment(c *gin.Context) {
	audiobookID := c.Param("id") // Changed from streamID
	commentID := c.Param("commentId")
	objAudiobookID, err := primitive.ObjectIDFromHex(audiobookID) // Changed from objStreamID

	// Convert commentID into ObjectID
	objID, err := primitive.ObjectIDFromHex(commentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment ID"})
		return
	}

	filter := bson.M{"_id": objID, "audiobookId": objAudiobookID} // Changed from streamId
	result, err := cc.CommentCol.DeleteOne(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete comment"})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Comment not found or unauthorized"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Comment deleted successfully"})
}
