package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AdController struct {
	AdCol *mongo.Collection
}

type Ad struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title     string             `bson:"title" json:"title"`
	Content   string             `bson:"content" json:"content"`
	ImageURL  string             `bson:"imageUrl" json:"imageUrl"`
	Link      string             `bson:"link" json:"link"`
	Placement string             `bson:"placement" json:"placement"` // "banner", "sidebar", "footer"
	IsActive  bool               `bson:"isActive" json:"isActive"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type CreateAdRequest struct {
	Title     string `json:"title" binding:"required"`
	Content   string `json:"content"`
	ImageURL  string `json:"imageUrl"`
	Link      string `json:"link"`
	Placement string `json:"placement" binding:"required"`
	IsActive  bool   `json:"isActive"`
}

type UpdateAdRequest struct {
	Title     string `json:"title"`
	Content   string `json:"content"`
	ImageURL  string `json:"imageUrl"`
	Link      string `json:"link"`
	Placement string `json:"placement"`
	IsActive  *bool  `json:"isActive"`
}

// GetAds returns all active ads
func (ac *AdController) GetAds(c *gin.Context) {
	filter := bson.M{"isActive": true}
	cursor, err := ac.AdCol.Find(context.TODO(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch ads"})
		return
	}
	defer cursor.Close(context.TODO())

	var ads []Ad
	if err = cursor.All(context.TODO(), &ads); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse ads"})
		return
	}

	c.JSON(http.StatusOK, ads)
}

// CreateAd creates a new advertisement
func (ac *AdController) CreateAd(c *gin.Context) {
	var req CreateAdRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ad := Ad{
		Title:     req.Title,
		Content:   req.Content,
		ImageURL:  req.ImageURL,
		Link:      req.Link,
		Placement: req.Placement,
		IsActive:  req.IsActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result, err := ac.AdCol.InsertOne(context.TODO(), ad)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create ad"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Ad created", "id": result.InsertedID})
}

// UpdateAd updates an existing advertisement
func (ac *AdController) UpdateAd(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ad ID"})
		return
	}

	var req UpdateAdRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	update := bson.M{}
	if req.Title != "" {
		update["title"] = req.Title
	}
	if req.Content != "" {
		update["content"] = req.Content
	}
	if req.ImageURL != "" {
		update["imageUrl"] = req.ImageURL
	}
	if req.Link != "" {
		update["link"] = req.Link
	}
	if req.Placement != "" {
		update["placement"] = req.Placement
	}
	if req.IsActive != nil {
		update["isActive"] = *req.IsActive
	}
	update["updatedAt"] = time.Now()

	result, err := ac.AdCol.UpdateOne(
		context.TODO(),
		bson.M{"_id": objID},
		bson.M{"$set": update},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update ad"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ad not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ad updated"})
}

// DeleteAd deletes an advertisement
func (ac *AdController) DeleteAd(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ad ID"})
		return
	}

	result, err := ac.AdCol.DeleteOne(context.TODO(), bson.M{"_id": objID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete ad"})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ad not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ad deleted"})
}
