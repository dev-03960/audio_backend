package controllers

import (
	"context"
	"net/http"
	"strings"

	// "time"

	models "live_stream/models"
	request "live_stream/models/requests"
	"live_stream/utils"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserController struct {
	UserCol   *mongo.Collection
	StreamCol *mongo.Collection // Now refers to audiobooks collection
	Redis     *redis.Client
}

func (uc *UserController) GetProfile(c *gin.Context) {
	userID := c.GetString("user_id")

	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var user models.User
	err = uc.UserCol.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.Password = "" // Hide password
	c.JSON(http.StatusOK, user)
}

func (uc *UserController) ChangePassword(c *gin.Context) {
	var req request.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	objID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var user models.User
	err = uc.UserCol.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if !utils.CheckPasswordHash(req.OldPassword, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Old password is incorrect"})
		return
	}

	newHash, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password"})
		return
	}

	_, err = uc.UserCol.UpdateOne(context.TODO(), bson.M{"_id": objID}, bson.M{"$set": bson.M{"password": newHash}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

func (uc *UserController) GetActiveUsers(c *gin.Context) {
	var activeUsers []models.User

	var cursor uint64
	var keys []string
	for {
		var err error
		var result []string
		result, cursor, err = uc.Redis.Scan(c, cursor, "session:*", 100).Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan Redis"})
			return
		}
		keys = append(keys, result...)
		if cursor == 0 {
			break
		}
	}

	// 2. Extract user IDs from session keys
	var userIDs []primitive.ObjectID
	for _, key := range keys {
		parts := strings.Split(key, "session:")
		if len(parts) != 2 {
			continue
		}
		idHex := parts[1]
		objID, err := primitive.ObjectIDFromHex(idHex)
		if err != nil {
			continue
		}
		userIDs = append(userIDs, objID)
	}

	// 3. Fetch users from MongoDB
	cursorMongo, err := uc.UserCol.Find(context.TODO(), bson.M{
		"_id": bson.M{"$in": userIDs},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	defer cursorMongo.Close(context.TODO())

	for cursorMongo.Next(context.TODO()) {
		var user models.User
		if err := cursorMongo.Decode(&user); err == nil {
			user.Password = "" // hide password
			activeUsers = append(activeUsers, user)
		}
	}

	c.JSON(http.StatusOK, activeUsers)
}

func (uc *UserController) UpdateBlockStatus(c *gin.Context) {
	userIDParam := c.Param("id")

	userID, err := primitive.ObjectIDFromHex(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		IsBlocked bool `json:"isBlocked"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	_, err = uc.UserCol.UpdateOne(
		context.TODO(),
		bson.M{"_id": userID},
		bson.M{"$set": bson.M{"isBlocked": req.IsBlocked}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	status := "unblocked"
	if req.IsBlocked {
		status = "blocked"
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User successfully " + status,
	})
}

func (uc *UserController) DeleteUser(c *gin.Context) {
	userIDParam := c.Param("id")

	// Convert to ObjectID
	userID, err := primitive.ObjectIDFromHex(userIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Delete user from MongoDB
	res, err := uc.UserCol.DeleteOne(context.TODO(), bson.M{"_id": userID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	if res.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Optionally: delete session from Redis
	if uc.Redis != nil {
		uc.Redis.Del(context.TODO(), "session:"+userID.Hex())
	}

	c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
}

func (uc *UserController) GetActiveUserCount(c *gin.Context) {
	var keys []string
	var cursor uint64

	for {
		result, nextCursor, err := uc.Redis.Scan(c, cursor, "session:*", 100).Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan Redis"})
			return
		}
		keys = append(keys, result...)
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"count": len(keys),
	})
}

func (uc *UserController) GetDashboardStats(c *gin.Context) {
	ctx := context.TODO()

	// 1️⃣ Total Users
	totalUsers, err := uc.UserCol.CountDocuments(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count users"})
		return
	}

	// 2️⃣ Active Users (based on Redis sessions)
	var keys []string
	var cursor uint64
	for {
		result, nextCursor, err := uc.Redis.Scan(c, cursor, "session:*", 100).Result()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan Redis"})
			return
		}
		keys = append(keys, result...)
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	activeUsers := len(keys)

	// 3️⃣ Total Audiobooks (changed from streams)
	totalAudiobooks, err := uc.StreamCol.CountDocuments(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count audiobooks"})
		return
	}

	// Return JSON
	c.JSON(http.StatusOK, gin.H{
		"totalUsers":      totalUsers,
		"activeUsers":     activeUsers,
		"totalAudiobooks": totalAudiobooks,
	})
}

func (uc *UserController) GetAllUsers(c *gin.Context) {
	ctx := context.TODO()

	cursor, err := uc.UserCol.Find(ctx, bson.M{}) // Fetch all users
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	defer cursor.Close(ctx)

	var users []models.User
	for cursor.Next(ctx) {
		var user models.User
		if err := cursor.Decode(&user); err == nil {
			user.Password = "" // Hide password
			users = append(users, user)
		}
	}

	c.JSON(http.StatusOK, users)
}

func (ac *UserController) SenduserCredential(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		PhoneNumber string `json:"phoneNumber" binding:"required"`
		Password    string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Phone number is required"})
		return
	}

	// Send OTP via Innosate API
	err := utils.SendUserCredentialWhatsApp(req.Name, req.PhoneNumber, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP sent successfully"})
}

func (uc *UserController) UpdateUserAccess(c *gin.Context) {
	userId := c.Param("id")
	oid, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid user ID"})
		return
	}

	var body struct {
		FullName    *string `json:"fullName"`
		PhoneNumber *string `json:"phoneNumber"`
		IsAdmin     *bool   `json:"isAdmin"`
		IsBlocked   *bool   `json:"isBlocked"`
		IsVerified  *bool   `json:"isVerified"`
		Password    *string `json:"password"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	update := bson.M{}
	if body.FullName != nil {
		update["fullName"] = *body.FullName
	}
	if body.PhoneNumber != nil {
		update["phoneNumber"] = *body.PhoneNumber
	}
	if body.IsAdmin != nil {
		update["isAdmin"] = *body.IsAdmin
	}
	if body.IsBlocked != nil {
		update["isBlocked"] = *body.IsBlocked
	}
	if body.IsVerified != nil {
		update["isVerified"] = *body.IsVerified
	}
	if body.Password != nil {
		hashed, _ := utils.HashPassword(*body.Password)
		update["password"] = hashed
	}

	if len(update) == 0 {
		c.JSON(400, gin.H{"error": "No fields to update"})
		return
	}

	_, err = uc.UserCol.UpdateOne(
		context.Background(),
		bson.M{"_id": oid},
		bson.M{"$set": update},
	)

	if err != nil {
		c.JSON(500, gin.H{"error": "Database error"})
		return
	}

	c.JSON(200, gin.H{
		"message": "User updated successfully",
		"updated": update,
	})
}
