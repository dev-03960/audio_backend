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

type SiteController struct {
	SiteChangesCol *mongo.Collection
}

func (sc *SiteController) CreateSiteChanges(c *gin.Context) {

	var req request.SiteChangesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sliders := make([]models.ImgSlider, 0)
	for _, s := range req.ImageSlider {
		sliders = append(sliders, models.ImgSlider{
			ID:         primitive.NewObjectID(),
			Title:      s.Title,
			Subtitle:   s.Subtitle,
			Image:      s.Image,
			Link:       s.Link,
			ButtonName: s.ButtonName,
		})
	}

	now := time.Now()

	siteChanges := models.SiteChanges{
		Site:           req.Site,
		LogoUrl:        req.Logourl,
		Calendarurl:    req.Calendarurl,
		Notification:   req.Notification,
		ImageSlider:    sliders,
		InviteOnlyMode: req.InviteOnlyMode,
		BuzzingText:    req.BuzzingText,
		NowText:        req.NowText,
		LiveTag:        req.LiveTag,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	_, err := sc.SiteChangesCol.InsertOne(context.TODO(), siteChanges)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Site changes created"})
}

func (sc *SiteController) GetSiteChanges(c *gin.Context) {
	id := c.Param("id")
	objID, _ := primitive.ObjectIDFromHex(id)
	var sitechange models.SiteChanges
	err := sc.SiteChangesCol.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&sitechange)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Site settings not found"})
		return
	}
	c.JSON(http.StatusOK, sitechange)
}

func (sc *SiteController) UpdateSiteChanges(c *gin.Context) {
	id := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req request.SiteChangesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	update := bson.M{}
	if req.Site != "" {
		update["site"] = req.Site
	}
	if req.Calendarurl != "" {
		update["calendarurl"] = req.Calendarurl
	}

	if req.Logourl != "" {
		update["logourl"] = req.Logourl
	}
	if req.Notification != "" {
		update["notification"] = req.Notification
	}
	if req.BuzzingText != "" {
		update["buzzingtext"] = req.BuzzingText
	}
	if req.NowText != "" {
		update["nowtext"] = req.NowText
	}
	if req.LiveTag != "" {
		update["livetag"] = req.LiveTag
	}
	update["inviteonlymode"] = req.InviteOnlyMode
	if len(req.ImageSlider) > 0 {
		sliders := make([]models.ImgSlider, 0)
		for _, s := range req.ImageSlider {
			sliders = append(sliders, models.ImgSlider{
				ID:         primitive.NewObjectID(),
				Title:      s.Title,
				Subtitle:   s.Subtitle,
				Image:      s.Image,
				Link:       s.Link,
				ButtonName: s.ButtonName,
			})
		}
		update["imageslider"] = sliders
	}

	update["updatedAt"] = time.Now()

	res, err := sc.SiteChangesCol.UpdateOne(
		context.TODO(),
		bson.M{"_id": objID},
		bson.M{"$set": update},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "update failed"})
		return
	}

	if res.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no document found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Site changes updated"})
}
