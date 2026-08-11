package main

import (
	"database/sql"
	"log"

	"gymgit/backend/config"
	"gymgit/backend/internal/database"
	"gymgit/backend/internal/handler"
	"gymgit/backend/internal/middleware"
	"gymgit/backend/internal/repository"
	"gymgit/backend/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	gin.SetMode(cfg.GinMode)

	// 2. Initialize Database Connection
	var db *sql.DB
	if cfg.DatabaseURL != "" {
		var errDB error
		db, errDB = database.ConnectDB(cfg.DatabaseURL)
		if errDB != nil {
			log.Printf("Warning: Initial database connection error: %v", errDB)
		} else {
			defer db.Close()
		}
	} else {
		log.Println("DATABASE_URL not set; running in database-disconnected mode")
	}

	// 3. Initialize Repositories & Services
	var userRepo repository.UserRepository
	var planRepo repository.PlanRepository
	var logRepo repository.GymLogRepository

	if db != nil {
		userRepo = repository.NewUserRepository(db)
		planRepo = repository.NewPlanRepository(db)
		logRepo = repository.NewGymLogRepository(db)
	}

	authService := service.NewAuthService(userRepo, planRepo)
	planService := service.NewPlanService(planRepo)
	logService := service.NewGymLogService(logRepo, userRepo, planRepo)
	statsService := service.NewStatsService(userRepo, planRepo, logRepo)

	// 4. Initialize Handlers
	healthHandler := handler.NewHealthHandler()
	authHandler := handler.NewAuthHandler(authService)
	planHandler := handler.NewPlanHandler(planService)
	logHandler := handler.NewLogHandler(logService, authService)
	statsHandler := handler.NewStatsHandler(statsService, authService)

	// 5. Initialize Gin Router
	router := gin.Default()

	// CORS & Timezone Middlewares
	router.Use(middleware.CORSMiddleware(cfg.AllowedOrigins))
	router.Use(middleware.TimezoneMiddleware())

	// Public Health Routes
	router.GET("/health", healthHandler.HealthCheck)

	apiV1 := router.Group("/api/v1")
	{
		apiV1.GET("/health", healthHandler.HealthCheck)
		apiV1.GET("/plans", planHandler.GetPlans)

		// Auth Group (Protected)
		authGroup := apiV1.Group("/auth")
		authGroup.Use(middleware.AuthMiddleware(cfg.SupabaseJWTSecret))
		{
			authGroup.POST("/bootstrap", authHandler.Bootstrap)
			authGroup.GET("/me", authHandler.GetMe)
			authGroup.PUT("/plan", authHandler.UpdatePlan)
			authGroup.POST("/timezone", authHandler.UpdateTimezone)
			authGroup.PUT("/timezone", authHandler.UpdateTimezone)
			authGroup.POST("/logout", authHandler.Logout)
		}

		// Gym Logs Group (Protected)
		logsGroup := apiV1.Group("/logs")
		logsGroup.Use(middleware.AuthMiddleware(cfg.SupabaseJWTSecret))
		{
			logsGroup.GET("", logHandler.GetLogs)
			logsGroup.POST("", logHandler.UpsertLog)
			logsGroup.PUT("/:date", logHandler.UpsertLog)
			logsGroup.DELETE("/:date", logHandler.DeleteLog)
			logsGroup.POST("/reset", logHandler.ResetDemoLogs)
		}

		// Analytics Stats Group (Protected)
		statsGroup := apiV1.Group("/stats")
		statsGroup.Use(middleware.AuthMiddleware(cfg.SupabaseJWTSecret))
		{
			statsGroup.GET("", statsHandler.GetStats)
			statsGroup.GET("/power", statsHandler.GetPowerStats)
		}
	}

	// 6. Start Server
	log.Printf("Gym-Git server starting on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
