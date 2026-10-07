package router

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"foodlog/server/internal/ai"
	"foodlog/server/internal/config"
	"foodlog/server/internal/handler"
	"foodlog/server/internal/httpx"
	"foodlog/server/internal/middleware"
	"foodlog/server/internal/repository"
	"foodlog/server/internal/service"
	"foodlog/server/internal/web"
)

// Setup 装配路由
func Setup(cfg *config.Config, pool *pgxpool.Pool) *gin.Engine {
	// 中间件
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS(cfg.ClientOrigin))

	// 大图上传：超过该阈值的内存占用转存磁盘（默认 32MB）
	r.MaxMultipartMemory = 8 << 20

	// 静态文件（上传的图片：/images/{recipe,restaurant}/YYYY/MM/xxx.jpg）
	r.Static("/images", cfg.ImagesRoot())

	// 初始化依赖
	userRepo := repository.NewUserRepo(pool)
	recipeRepo := repository.NewRecipeRepo(pool)
	restaurantRepo := repository.NewRestaurantRepo(pool)
	dishRepo := repository.NewDishRepo(pool)
	favoriteRepo := repository.NewFavoriteRepo(pool)
	tagRepo := repository.NewTagRepo(pool)

	// AI Provider（未配置则 nil，自动标签/识别静默降级或明确报错）
	var aiProvider ai.Provider
	if cfg.AIProvider == "ollama" {
		aiProvider = ai.NewOllama(cfg.AIBaseURL, cfg.AIModel, cfg.AITimeout)
	} else if cfg.AIEnabled() {
		aiProvider = ai.NewOpenAICompat(cfg.AIBaseURL, cfg.AIAPIKey, cfg.AIModel, cfg.AITimeout)
	}

	authService := service.NewAuthService(userRepo, cfg.JWTSecret)
	tagService := service.NewTagService(tagRepo, aiProvider)
	recipeService := service.NewRecipeService(pool, recipeRepo, favoriteRepo, tagRepo, tagService)
	restaurantService := service.NewRestaurantService(restaurantRepo, dishRepo)
	uploadService := service.NewUploadService(cfg)
	searchService := service.NewSearchService(recipeRepo, restaurantRepo, dishRepo)
	exportService := service.NewExportService(pool, recipeRepo, restaurantRepo, dishRepo, favoriteRepo, tagRepo, tagService)

	authHandler := handler.NewAuthHandler(authService)
	recipeHandler := handler.NewRecipeHandler(recipeService, tagService, aiProvider)
	tagHandler := handler.NewTagHandler(tagService)
	restaurantHandler := handler.NewRestaurantHandler(restaurantService)
	uploadHandler := handler.NewUploadHandler(uploadService)
	searchHandler := handler.NewSearchHandler(searchService)
	exportHandler := handler.NewExportHandler(exportService)

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
			// 静态路由需在 /:id 之前注册，避免被参数路由捕获
			recipes.GET("/tags", tagHandler.List)
			recipes.GET("/random", recipeHandler.Random)
			recipes.POST("/ai/recognize", recipeHandler.Recognize)
			recipes.GET("/:id", recipeHandler.Get)
			recipes.PUT("/:id", recipeHandler.Update)
			recipes.DELETE("/:id", recipeHandler.Delete)
			recipes.POST("/:id/favorite", recipeHandler.Favorite)
			recipes.DELETE("/:id/favorite", recipeHandler.Unfavorite)
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

		// 全局搜索
		api.GET("/search", authMW, searchHandler.Search)

		// 标签字典（自定义分类/标签维护）
		tagsManage := api.Group("/tags", authMW)
		{
			tagsManage.POST("", tagHandler.CreateTag)
			tagsManage.PUT("/:id", tagHandler.UpdateTag)
			tagsManage.DELETE("/:id", tagHandler.DeleteTag)
		}
		tagCategories := api.Group("/tag-categories", authMW)
		{
			tagCategories.POST("", tagHandler.CreateCategory)
			tagCategories.PUT("/:id", tagHandler.UpdateCategory)
			tagCategories.DELETE("/:id", tagHandler.DeleteCategory)
		}

		// 数据导出/导入
		api.GET("/export", authMW, exportHandler.Export)
		api.POST("/export/import", authMW, exportHandler.Import)

		// 上传
		api.POST("/upload", authMW, uploadHandler.Upload)
	}

	// 未匹配路由：/api 走 JSON 404，/images 缺失文件返回 404，其余交给前端静态资源 + SPA 回退
	spa := web.Handler()
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		switch {
		case strings.HasPrefix(path, "/api/"):
			httpx.RespondError(c, http.StatusNotFound, httpx.CodeNotFound, "接口不存在")
		case strings.HasPrefix(path, "/images/"), path == "/images":
			// gin 静态目录未命中时会转交 NoRoute，这里必须返回 404 而不是 SPA 页面
			http.NotFound(c.Writer, c.Request)
		default:
			spa.ServeHTTP(c.Writer, c.Request)
		}
	})

	return r
}
