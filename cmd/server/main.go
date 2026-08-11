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
	var streakRepo repository.StreakRepository
	var itemRepo repository.ItemRepository
	var inventoryRepo repository.InventoryRepository
	var rewardRepo repository.RewardRepository

	if db != nil {
		userRepo = repository.NewUserRepository(db)
		planRepo = repository.NewPlanRepository(db)
		logRepo = repository.NewGymLogRepository(db)
		streakRepo = repository.NewStreakRepository(db)
		itemRepo = repository.NewItemRepository(db)
		inventoryRepo = repository.NewInventoryRepository(db)
		rewardRepo = repository.NewRewardRepository(db)
	}

	authService := service.NewAuthService(userRepo, planRepo)
	planService := service.NewPlanService(planRepo, userRepo)
	logService := service.NewGymLogService(logRepo, userRepo, planRepo)
	statsService := service.NewStatsService(userRepo, planRepo, logRepo)
	streakService := service.NewStreakService(streakRepo, userRepo, planRepo, logRepo)
	inventoryService := service.NewInventoryService(itemRepo, inventoryRepo, logRepo, streakRepo, userRepo)
	rewardService := service.NewRewardService(rewardRepo, itemRepo, inventoryRepo, streakRepo, logRepo)

	// 4. Initialize Handlers
	healthHandler := handler.NewHealthHandler()
	authHandler := handler.NewAuthHandler(authService)
	planHandler := handler.NewPlanHandler(planService)
	logHandler := handler.NewLogHandler(logService, authService)
	statsHandler := handler.NewStatsHandler(statsService, authService)
	streakHandler := handler.NewStreakHandler(streakService, inventoryService)
	inventoryHandler := handler.NewInventoryHandler(inventoryService)
	rewardHandler := handler.NewRewardHandler(rewardService)

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
		apiV1.GET("/items", inventoryHandler.GetCatalog)
		apiV1.GET("/rewards/plans", rewardHandler.GetAllPlans)

		// Inventory Group (Protected)
		inventoryGroup := apiV1.Group("/inventory")
		inventoryGroup.Use(middleware.AuthMiddleware(cfg.SupabaseJWTSecret))
		{
			inventoryGroup.GET("", inventoryHandler.GetInventory)
			inventoryGroup.POST("/use", inventoryHandler.UseItem)
		}

		// Rewards Group (Protected)
		rewardsGroup := apiV1.Group("/rewards")
		rewardsGroup.Use(middleware.AuthMiddleware(cfg.SupabaseJWTSecret))
		{
			rewardsGroup.GET("/roadmap", rewardHandler.GetRoadmap)
			rewardsGroup.POST("/claim", rewardHandler.ClaimReward)
		}

		// Admin Rewards Group (Protected)
		adminGroup := apiV1.Group("/admin")
		adminGroup.Use(middleware.AuthMiddleware(cfg.SupabaseJWTSecret))
		{
			adminRewardsGroup := adminGroup.Group("/rewards")
			{
				adminRewardsGroup.POST("/plans", rewardHandler.CreateRewardPlan)
				adminRewardsGroup.DELETE("/plans/:id", rewardHandler.DeleteRewardPlan)
				adminRewardsGroup.POST("/plans/:id/milestones", rewardHandler.UpsertMilestone)
				adminRewardsGroup.DELETE("/plans/:id/milestones/:milestone_id", rewardHandler.DeleteMilestone)
			}
		}

		// Plans Group (Protected queue route)
		plansGroup := apiV1.Group("/plans")
		plansGroup.Use(middleware.AuthMiddleware(cfg.SupabaseJWTSecret))
		{
			plansGroup.PUT("/queue", planHandler.QueuePlan)
		}

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

		// Streak & Cycle Group (Protected)
		streakGroup := apiV1.Group("/streak")
		streakGroup.Use(middleware.AuthMiddleware(cfg.SupabaseJWTSecret))
		{
			streakGroup.GET("", streakHandler.GetStreak)
			streakGroup.POST("/restore", streakHandler.RestoreStreak)
			streakGroup.POST("/freeze", streakHandler.FreezeStreak)
			streakGroup.POST("/unfreeze", streakHandler.UnfreezeStreak)
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
