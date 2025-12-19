package controllers

import (
	"context"
	"fmt"
	models "live_stream/models"
	request "live_stream/models/requests"
	"live_stream/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthController struct {
	UserCol *mongo.Collection
	Redis   *redis.Client
}

// Signup creates a new user account
func (ac *AuthController) Signup(c *gin.Context) {
	var req request.SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	var existingUser models.User
	err := ac.UserCol.FindOne(context.TODO(), bson.M{"phoneNumber": req.PhoneNumber}).Decode(&existingUser)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create user
	user := models.User{
		FullName:    req.FullName,
		PhoneNumber: req.PhoneNumber,
		Password:    hashedPassword,
		IsAdmin:     false,
		IsBlocked:   false,
		IsVerified:  false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	result, err := ac.UserCol.InsertOne(context.TODO(), user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully", "id": result.InsertedID})
}

// Login authenticates a user and creates a session
func (ac *AuthController) Login(c *gin.Context) {
	var req request.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user
	var user models.User
	err := ac.UserCol.FindOne(context.TODO(), bson.M{"phoneNumber": req.PhoneNumber}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check if user is blocked
	if user.IsBlocked {
		c.JSON(http.StatusForbidden, gin.H{"error": "User is blocked"})
		return
	}

	// Verify password
	if !utils.CheckPasswordHash(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate JWT
	token, err := utils.GenerateJWT(user.ID.Hex())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Store session in Redis
	sessionKey := "session:" + user.ID.Hex()
	ac.Redis.Set(context.TODO(), sessionKey, "active", 24*time.Hour)

	user.Password = "" // Hide password
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user,
	})
}

// Logout destroys a user session
func (ac *AuthController) Logout(c *gin.Context) {
	userID := c.GetString("user_id")
	sessionKey := "session:" + userID
	ac.Redis.Del(context.TODO(), sessionKey)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// SendOTP sends an OTP to the user's phone number
func (ac *AuthController) SendOTP(c *gin.Context) {
	var req request.SendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate OTP (6 digits)
	otp := fmt.Sprintf("%06d", time.Now().Unix()%1000000)

	// Store OTP in Redis with 5 minute expiration
	otpKey := "otp:" + req.PhoneNumber
	ac.Redis.Set(context.TODO(), otpKey, otp, 5*time.Minute)

	// TODO: Send OTP via SMS/WhatsApp
	// For now, return it in response (for development only)
	c.JSON(http.StatusOK, gin.H{"message": "OTP sent", "otp": otp})
}

// VerifyOTP verifies an OTP for a phone number
func (ac *AuthController) VerifyOTP(c *gin.Context) {
	var req request.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get OTP from Redis
	otpKey := "otp:" + req.PhoneNumber
	storedOTP, err := ac.Redis.Get(context.TODO(), otpKey).Result()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OTP expired or not found"})
		return
	}

	// Verify OTP
	if storedOTP != req.OTP {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid OTP"})
		return
	}

	// Delete OTP after successful verification
	ac.Redis.Del(context.TODO(), otpKey)

	c.JSON(http.StatusOK, gin.H{"message": "OTP verified successfully"})
}

// ForgotPassword initiates password reset process
func (ac *AuthController) ForgotPassword(c *gin.Context) {
	var req request.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user exists
	var user models.User
	err := ac.UserCol.FindOne(context.TODO(), bson.M{"phoneNumber": req.PhoneNumber}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Generate OTP
	otp := fmt.Sprintf("%06d", time.Now().Unix()%1000000)
	otpKey := "reset_otp:" + req.PhoneNumber
	ac.Redis.Set(context.TODO(), otpKey, otp, 10*time.Minute)

	// TODO: Send OTP via SMS/WhatsApp
	c.JSON(http.StatusOK, gin.H{"message": "Reset OTP sent", "otp": otp})
}

// ResetPassword resets user password with OTP verification
func (ac *AuthController) ResetPassword(c *gin.Context) {
	var req request.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify OTP
	otpKey := "reset_otp:" + req.PhoneNumber
	storedOTP, err := ac.Redis.Get(context.TODO(), otpKey).Result()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OTP expired or not found"})
		return
	}

	if storedOTP != req.OTP {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid OTP"})
		return
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Update password
	_, err = ac.UserCol.UpdateOne(
		context.TODO(),
		bson.M{"phoneNumber": req.PhoneNumber},
		bson.M{"$set": bson.M{"password": hashedPassword, "updatedAt": time.Now()}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset password"})
		return
	}

	// Delete OTP
	ac.Redis.Del(context.TODO(), otpKey)

	c.JSON(http.StatusOK, gin.H{"message": "Password reset successfully"})
}
