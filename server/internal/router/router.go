package router

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"foodlog/server/internal/config"
	"foodlog/server/internal/handler"
	"foodlog/server/internal/middleware"
	"foodlog/server/internal/repository"
	"foodlog/server/internal/service"
)

// Setup 装配路由
func Setup(cfg *config.Config, pool *pgxpool.Pool) *gin.Engine {
	// 中间件
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS(cfg.ClientOrigin))

	// 静态文件（上传的图片）
	r.Static("/uploads", cfg.UploadDir)

	// 初始化依赖
	userRepo := repository.NewUserRepo(pool)
	recipeRepo := repository.NewRecipeRepo(pool)
	restaurantRepo := repository.NewRestaurantRepo(pool)
	dishRepo := repository.NewDishRepo(pool)

	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	recipeService := service.NewRecipeService(recipeRepo)
	restaurantService := service.NewRestaurantService(restaurantRepo, dishRepo)
	uploadService := service.NewUploadService(cfg.UploadDir)

	authHandler := handler.NewAuthHandler(authService)
	recipeHandler := handler.NewRecipeHandler(recipeService)
	restaurantHandler := handler.NewRestaurantHandler(restaurantService)
	uploadHandler := handler.NewUploadHandler(uploadService)

	// 认证中间件
	authMW := middleware.Auth(cfg.JWTSecret)

	api := r.Group("/api")
	{
		// 健康检查
		api.GET("/health", authHandler.Health)

		// 认证
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/me", authMW, authHandler.Me)
			auth.PUT("/me", authMW, authHandler.UpdateMe)
		}

		// 菜谱
		recipes := api.Group("/recipes", authMW)
		{
			recipes.GET("", recipeHandler.List)
			recipes.POST("", recipeHandler.Create)
			recipes.GET("/:id", recipeHandler.Get)
			recipes.PUT("/:id", recipeHandler.Update)
			recipes.DELETE("/:id", recipeHandler.Delete)
		}

		// 餐厅
		restaurants := api.Group("/restaurants", authMW)
		{
			restaurants.GET("", restaurantHandler.List)
			restaurants.POST("", restaurantHandler.Create)
			restaurants.GET("/:id", restaurantHandler.Get)
			restaurants.PUT("/:id", restaurantHandler.Update)
			restaurants.DELETE("/:id", restaurantHandler.Delete)
			restaurants.POST("/:id/dishes", restaurantHandler.AddDish)
		}

		// 菜品
		dishes := api.Group("/dishes", authMW)
		{
			dishes.PUT("/:id", restaurantHandler.UpdateDish)
			dishes.DELETE("/:id", restaurantHandler.DeleteDish)
		}

		// 上传
		api.POST("/upload", authMW, uploadHandler.Upload)
	}

	return r
}
