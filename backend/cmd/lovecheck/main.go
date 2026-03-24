package main

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"lovecheck/internal/bloom"
	"lovecheck/internal/db"
	"lovecheck/internal/handler"
	"lovecheck/internal/jobs"
	"lovecheck/internal/middleware"
	"lovecheck/internal/storage"
	"lovecheck/pkg/logger"
)

func main() {
	_ = godotenv.Load()

	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	db.InitDB()
	bloom.Init(db.DB, 10_000_000, 0.0001)
	middleware.InitRedis()
	storage.InitMinio()

	jobs.StartCleanseScheduler(db.DB, 15*24*time.Hour)

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     getAllowedOrigins(),
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "X-Admin-Secret"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.GET("/favicon.ico", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	writeLimiter := middleware.RateLimitMiddleware(10, 1*time.Minute)
	readLimiter := middleware.RateLimitMiddleware(60, 1*time.Minute)

	api := router.Group("/api/v1")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"status":  "UP",
				"message": "LoveCheck is running.",
			})
		})

		api.POST("/report", writeLimiter, handler.HandleReport)
		api.POST("/appeal", writeLimiter, handler.HandleAppeal)
		api.POST("/vote", writeLimiter, handler.HandleVote)
		api.POST("/activate", writeLimiter, handler.HandleActivate)

		api.GET("/query", readLimiter, handler.HandleQuery)
		api.GET("/check-access", readLimiter, handler.HandleCheckAccess)
		api.GET("/platforms", readLimiter, handler.HandlePublicPlatforms)
		api.GET("/stats/public", readLimiter, handler.HandlePublicStats)

		api.GET("/evidence/:filename", handler.GetEvidence)

		api.POST("/pay/create", writeLimiter, handler.HandlePayCreate)
		api.POST("/pay/notify", handler.HandlePayNotify)
		api.GET("/pay/status", readLimiter, handler.HandlePayStatus)
		api.POST("/pay/paypal-capture", writeLimiter, handler.HandlePayPalCapture)

		api.GET("/push/vapid-key", handler.HandleGetVAPIDKey)
		api.POST("/push/subscribe", writeLimiter, handler.HandlePushSubscribe)
		api.POST("/push/unsubscribe", writeLimiter, handler.HandlePushUnsubscribe)

		admin := api.Group("/admin")
		admin.Use(handler.AdminAuth())
		{
			admin.GET("/generate-codes", handler.HandleGenerateCodes)
			admin.GET("/unused-codes", handler.HandleListUnusedCodes)
			admin.GET("/stats", handler.HandleDashboardStats)
			admin.GET("/platforms", handler.HandleListPlatforms)
			admin.POST("/platforms", handler.HandleCreatePlatform)
			admin.PUT("/platforms/:id", handler.HandleUpdatePlatform)
			admin.DELETE("/platforms/:id", handler.HandleDeletePlatform)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	logger.Log.Info().Str("port", port).Msg("Server starting")
	if err := router.Run(":" + port); err != nil {
		logger.Log.Fatal().Err(err).Msg("Server failed to start")
	}
}

func getAllowedOrigins() []string {
	if origin := os.Getenv("CORS_ORIGIN"); origin != "" {
		return []string{origin}
	}
	return []string{"*"}
}
