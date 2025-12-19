package main

import (
	"live_stream/config"
	"live_stream/controllers"
	"live_stream/middleware"
	"live_stream/route"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	// -------------------------
	// Load .env file
	// -------------------------
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}
	// -------------------------
	// Initialize MongoDB & Redis
	// -------------------------
	mongoClient := config.InitMongo()
	redisClient := config.InitRedis()

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "streamapp"
	}

	// -------------------------
	// Initialize Controllers
	// -------------------------
	authCtrl := &controllers.AuthController{
		UserCol: mongoClient.Database(dbName).Collection("users"),
		Redis:   redisClient,
	}

	userCtrl := &controllers.UserController{
		UserCol:   mongoClient.Database(dbName).Collection("users"),
		StreamCol: mongoClient.Database(dbName).Collection("audiobooks"), // Changed from streams
		Redis:     redisClient,
	}

	audiobookCtrl := &controllers.AudiobookController{
		AudiobookCol:   mongoClient.Database(dbName).Collection("audiobooks"),
		InteractionCol: mongoClient.Database(dbName).Collection("audiobook_interactions"),
	}

	commentCtrl := &controllers.CommentController{
		CommentCol: mongoClient.Database(dbName).Collection("comments"),
		UserCol:    mongoClient.Database(dbName).Collection("users"),
	}

	adCtrl := &controllers.AdController{
		AdCol: mongoClient.Database(dbName).Collection("ads"),
	}

	siteCtrl := &controllers.SiteController{
		SiteChangesCol: mongoClient.Database(dbName).Collection("site_change"),
	}

	// -------------------------
	// Initialize Middleware
	// -------------------------
	middleware.InitRedis(redisClient)

	// -------------------------
	// Initialize Router
	// -------------------------
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5173",
			"http://raceraja.in",
			"https://raceraja.in",
			"https://www.raceraja.in", // added correctly
			"http://www.raceraja.in",  // optional if you serve over http
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Cache-Control", "no-store")
		c.Next()
	})

	// -------------------------
	// Setup All Routes
	// -------------------------
	route.SetupRoutes(router, redisClient, authCtrl, userCtrl, audiobookCtrl, commentCtrl, adCtrl, siteCtrl)

	// -------------------------
	// Start Server
	// -------------------------

	// router.Static("/static", "./dist/assets")
	// router.StaticFile("/favicon.ico", "./dist/favicon.ico")

	// server.Use(static.Serve("/data", static.LocalFile("./nz-data-build", true)))
	// server.Use(static.Serve("/static", static.LocalFile("./nz-data-build/static", true)))
	// server.Use(static.Serve("/static", static.LocalFile("./nsai/static", true)))
	// server.Use(static.Serve("/app", static.LocalFile("./fe", true)))
	// server.Use(static.Serve("/static", static.LocalFile("./fe/static", true)))
	// server.Use(static.Serve("/career", static.LocalFile("./fe-career", true)))
	// server.Use(static.Serve("/static", static.LocalFile("./fe-career/static", true)))

	// Catch-all: serve index.html for React Router
	// router.NoRoute(func(c *gin.Context) {
	// 	c.File("./dist/index.html")
	// })

	// router.Use(static.Serve("/", static.LocalFile("./dist", true)))
	// router.Use(static.Serve("/static", static.LocalFile("./dist/assests", true)))
	// router.Use(static.Serve("/vite.svg", static.LocalFile("./dist/vite.svg", true)))

	router.Static("/assets", "../live_stream-frontend/frotend/dist/assets")
	router.StaticFile("/favicon.ico", "../live_stream-frontend/frotend/dist/favicon.ico")

	// Catch-all for React Router
	router.NoRoute(func(c *gin.Context) {
		c.File("../live_stream-frontend/frotend/dist/index.html")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	router.Run(":" + port)
}
