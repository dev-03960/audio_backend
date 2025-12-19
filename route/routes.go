package route

import (
	"live_stream/controllers"
	"live_stream/middleware"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// SetupRoutes initializes all API routes
func SetupRoutes(
	r *gin.Engine,
	redisClient *redis.Client,
	authCtrl *controllers.AuthController,
	userCtrl *controllers.UserController,
	audiobookCtrl *controllers.AudiobookController,
	commentCtrl *controllers.CommentController,
	adCtrl *controllers.AdController,
	siteCtrl *controllers.SiteController,
) {
	api := r.Group("/api")

	// ===== Auth Routes =====
	auth := api.Group("/auth")
	auth.POST("/signup", authCtrl.Signup)
	auth.POST("/login", authCtrl.Login)
	auth.POST("/logout", middleware.AuthMiddleware(redisClient), authCtrl.Logout)
	auth.POST("/forgot-password", authCtrl.ForgotPassword)
	auth.POST("/reset-password", authCtrl.ResetPassword)
	auth.POST("/send-otp", authCtrl.SendOTP)
	auth.POST("/verify-otp", authCtrl.VerifyOTP)

	// ===== User Routes =====
	user := api.Group("/user")
	user.GET("/active-count", userCtrl.GetActiveUserCount)

	user.Use(middleware.AuthMiddleware(redisClient))
	user.GET("/profile", userCtrl.GetProfile)
	user.PUT("/change-password", userCtrl.ChangePassword)

	// ===== Audiobook Routes (replaced Stream routes) =====
	audiobook := api.Group("/audiobooks")
	audiobook.GET("", audiobookCtrl.GetAudiobooks)                                                                  // Public - list all audiobooks
	audiobook.GET("/:id", audiobookCtrl.GetAudiobookByID)                                                           // Public - get audiobook details
	audiobook.POST("/:id/like", middleware.AuthMiddleware(redisClient), audiobookCtrl.LikeAudiobook)                // Authenticated - like audiobook
	audiobook.POST("/:id/dislike", middleware.AuthMiddleware(redisClient), audiobookCtrl.DislikeAudiobook)          // Authenticated - dislike audiobook
	audiobook.GET("/:id/stats", audiobookCtrl.GetAudiobookStats)                                                    // Public - get stats
	audiobook.POST("/:id/comments", middleware.AuthMiddleware(redisClient), commentCtrl.AddComment)                 // Authenticated - add comment
	audiobook.GET("/:id/comments", commentCtrl.GetComments)                                                         // Public - get comments
	audiobook.DELETE("/:id/comments/:commentId", middleware.AuthMiddleware(redisClient), commentCtrl.DeleteComment) // Authenticated - delete comment

	// ===== Admin Routes =====
	admin := api.Group("/admin")
	admin.POST("/login", authCtrl.Login)
	admin.Use(middleware.AdminMiddleware(redisClient))
	admin.GET("/users", userCtrl.GetProfile) // placeholder
	admin.PUT("/users/:id/block", userCtrl.UpdateBlockStatus)
	admin.DELETE("/users/:id", userCtrl.DeleteUser)
	admin.PUT("/users/update/:id", userCtrl.UpdateUserAccess)
	admin.GET("/active-users", userCtrl.GetActiveUsers)
	admin.GET("/users/all", userCtrl.GetAllUsers)
	admin.POST("/sendcredential", userCtrl.SenduserCredential)

	// Admin Audiobook management (replaced Stream management)
	admin.POST("/audiobooks", audiobookCtrl.CreateAudiobook)
	admin.PUT("/audiobooks/:id", audiobookCtrl.UpdateAudiobook)
	admin.DELETE("/audiobooks/:id", audiobookCtrl.DeleteAudiobook)
	admin.GET("/audiobooks", audiobookCtrl.GetAudiobooks) // Admin can also list all

	// Admin Site_Changes management
	admin.POST("/site", siteCtrl.CreateSiteChanges)
	admin.GET("/site/:id", siteCtrl.GetSiteChanges) // admin-only
	admin.PUT("/site/:id", siteCtrl.UpdateSiteChanges)

	// ===== Public Site_Changes =====
	api.GET("/site/:id", siteCtrl.GetSiteChanges) // public access
	api.GET("/ads", adCtrl.GetAds)                // public ads access

	// Admin Ads
	admin.POST("/ads", adCtrl.CreateAd)
	admin.PUT("/ads/:id", adCtrl.UpdateAd)
	admin.DELETE("/ads/:id", adCtrl.DeleteAd)

	// Dashboard Stats
	admin.GET("/dashboard/stats", userCtrl.GetDashboardStats)

}
