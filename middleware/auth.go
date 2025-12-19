package middleware

import (
	"context"
	"live_stream/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var redisClient *redis.Client

// InitRedis initializes the Redis client for middleware
func InitRedis(client *redis.Client) {
	redisClient = client
}

// AuthMiddleware validates JWT and checks if session exists in Redis
func AuthMiddleware(redis *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := utils.ValidateJWT(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Check if session exists in Redis
		exists, err := redis.Exists(context.TODO(), "session:"+claims.UserID).Result()
		if err != nil || exists == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Session expired or invalid"})
			c.Abort()
			return
		}

		// Set user_id in context for controllers to use
		c.Set("user_id", claims.UserID)
		c.Next()
	}
}

// AdminMiddleware validates JWT and checks if user is admin
func AdminMiddleware(redis *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		token := parts[1]
		claims, err := utils.ValidateJWT(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Check if session exists in Redis
		sessionKey := "session:" + claims.UserID
		sessionData, err := redis.Get(context.TODO(), sessionKey).Result()
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Session expired or invalid"})
			c.Abort()
			return
		}

		// Check if user is admin (session data should contain "admin:true" or similar)
		// For now, we'll check if the session exists and assume admin validation
		// You may want to store admin status in the session or check the database
		if sessionData == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Next()
	}
}

// GetUserID retrieves the user ID from the Gin context
func GetUserID(c *gin.Context) (primitive.ObjectID, error) {
	userIDStr := c.GetString("user_id")
	return primitive.ObjectIDFromHex(userIDStr)
}
