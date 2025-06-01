package app

import (
	"context"
	"electronic-muyu-backend/internal/config"
	"electronic-muyu-backend/internal/database"
	"electronic-muyu-backend/internal/handlers"
	"electronic-muyu-backend/internal/middleware"
	"electronic-muyu-backend/internal/services"
	"electronic-muyu-backend/internal/websocket"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

// Module 定义应用程序的主要模块
var Module = fx.Options(
	// 配置模块
	fx.Provide(config.LoadConfig),

	// 数据库模块
	fx.Provide(database.InitDB),

	// 服务模块
	fx.Provide(
		services.NewLogger,
		services.NewRedisService,
		services.NewUserService,
		services.NewAuthService,
		services.NewRefreshTokenService,
		// 提供 RefreshTokenRepository 接口绑定
		fx.Annotate(
			services.NewRefreshTokenService,
			fx.As(new(services.RefreshTokenRepository)),
		),
		services.NewScoreService,
		services.NewLeaderboardService,
		services.NewRateLimiterService,
		services.NewLeaderboardBroadcastService,
	),

	// WebSocket 模块
	fx.Provide(websocket.NewHub),

	// 处理器模块
	fx.Provide(
		handlers.NewAuthHandler,
		handlers.NewUserHandler,
		handlers.NewScoreHandler,
		handlers.NewGameHandler,
		handlers.NewLeaderboardWebSocketHandler,
	),

	// 中间件模块
	fx.Provide(
		middleware.NewRateLimiterMiddleware,
	),

	// 路由器模块
	fx.Provide(NewRouter),

	// HTTP 服务器模块
	fx.Provide(NewHTTPServer),

	// 启动应用程序
	fx.Invoke(StartApplication),
)

// RouterParams 定义路由器的依赖参数
type RouterParams struct {
	fx.In

	AuthHandler           *handlers.AuthHandler
	UserHandler           *handlers.UserHandler
	ScoreHandler          *handlers.ScoreHandler
	GameHandler           *handlers.GameHandler
	LeaderboardWSHandler  *handlers.LeaderboardWebSocketHandler
	RateLimiterMiddleware *middleware.RateLimiterMiddleware
	Config                *config.Config
}

// NewRouter 创建并配置 Gin 路由器
func NewRouter(params RouterParams) *gin.Engine {
	// 设置 Gin 模式
	if params.Config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// 基本中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// CORS 中间件
	r.Use(middleware.CORS())

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"service":   "electronic-muyu-backend",
			"timestamp": getCurrentTimestamp(),
		})
	})

	// API 路由组
	api := r.Group("/api/v1")

	// 应用速率限制中间件到需要的路由
	authGroup := api.Group("/auth")
	authGroup.Use(params.RateLimiterMiddleware.AuthRateLimit())
	{
		authGroup.POST("/register", params.UserHandler.Register)
		authGroup.POST("/login", params.UserHandler.Login)
		authGroup.POST("/google", params.AuthHandler.GoogleSSO)
		authGroup.POST("/apple", params.AuthHandler.AppleSSO)
	}

	// 需要认证的路由
	authenticated := api.Group("")
	authenticated.Use(middleware.JWTAuth(params.Config.JWTSecret))
	{
		// 用户相关
		authenticated.GET("/user/profile", params.UserHandler.GetProfile)
		authenticated.PUT("/user/profile", params.UserHandler.UpdateProfile)

		// 分数相关 - 应用分数提交速率限制
		scoreGroup := authenticated.Group("/scores")
		scoreGroup.Use(params.RateLimiterMiddleware.ScoreSubmitRateLimit())
		{
			scoreGroup.POST("", params.ScoreHandler.SubmitScore)
		}

		// 其他分数相关路由（不需要特殊限制）
		authenticated.GET("/scores/best", params.ScoreHandler.GetUserScore)
		authenticated.GET("/scores/history", params.ScoreHandler.GetScoreHistory)

		// 排行榜相关
		authenticated.GET("/leaderboard", params.ScoreHandler.GetLeaderboard)
		authenticated.GET("/leaderboard/daily", params.ScoreHandler.GetDailyLeaderboard)
		authenticated.GET("/leaderboard/weekly", params.ScoreHandler.GetWeeklyLeaderboard)

		// WebSocket 连接
		authenticated.GET("/ws", params.GameHandler.HandleWebSocket)
		authenticated.GET("/ws/leaderboard", params.LeaderboardWSHandler.HandleLeaderboardWebSocket)

		// Token 刷新
		authenticated.POST("/auth/refresh", params.AuthHandler.RefreshToken)
	}

	// 应用通用 API 速率限制
	api.Use(params.RateLimiterMiddleware.APIRateLimit())

	return r
}

// ServerParams 定义 HTTP 服务器的依赖参数
type ServerParams struct {
	fx.In

	Router *gin.Engine
	Config *config.Config
}

// NewHTTPServer 创建 HTTP 服务器
func NewHTTPServer(params ServerParams) *http.Server {
	return &http.Server{
		Addr:    fmt.Sprintf(":%s", params.Config.Port),
		Handler: params.Router,
	}
}

// AppParams 定义应用程序启动的依赖参数
type AppParams struct {
	fx.In

	Server               *http.Server
	Hub                  *websocket.Hub
	LeaderboardBroadcast *services.LeaderboardBroadcastService
	Config               *config.Config
	Logger               *services.Logger
	Lifecycle            fx.Lifecycle
}

// StartApplication 启动应用程序
func StartApplication(params AppParams) {
	params.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// 启动 WebSocket Hub
			go params.Hub.Run()

			// 启动 HTTP 服务器
			go func() {
				params.Logger.Info("Starting server on port %s", params.Config.Port)
				if err := params.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					params.Logger.Error("Failed to start server: %v", err)
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			params.Logger.Info("Stopping server...")

			// 停止 HTTP 服务器
			return params.Server.Shutdown(ctx)
		},
	})
}

// getCurrentTimestamp 获取当前时间戳
func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}
