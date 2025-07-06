package web

import (
	"heimdall/backend/internal/application/services"
	"heimdall/backend/internal/config"
	"heimdall/backend/internal/infra"
	"heimdall/backend/internal/infra/postgres/repositories"
	"heimdall/backend/internal/interfaces/web/controllers"
	"heimdall/backend/internal/interfaces/web/middlewares"

	"github.com/gin-gonic/gin"
)

type Router struct {
	engine    *gin.Engine
	DataStore *infra.DataStore
}

func NewRouter(DataStore *infra.DataStore) *Router {
	gin.SetMode(config.AppConfig.Server.GinMode)
	engine := gin.Default()
	engine.SetTrustedProxies([]string{"127.0.0.1"})

	return &Router{
		engine:    engine,
		DataStore: DataStore,
	}
}

func (r *Router) Setup() *gin.Engine {

	// Repositories
	userRepository := repositories.NewUserRepository(r.DataStore.Postgres)
	sessionRepository := repositories.NewSessionRepository(r.DataStore.Postgres)

	// Services
	accessTokenJWTService := services.NewAccessTokenJWTService()
	refreshTokenJWTService := services.NewRefreshTokenJWTService()
	sessionService := services.NewSessionService(sessionRepository, accessTokenJWTService, refreshTokenJWTService)
	authService := services.NewAuthService(sessionService)

	// Controllers
	healthController := controllers.NewHealthController()
	authController := controllers.NewAuthController(authService, accessTokenJWTService)

	// Middlewares
	authMiddleware := middlewares.NewAuthMiddleware(accessTokenJWTService, refreshTokenJWTService)

	// Apply Global Middlewares
	r.engine.Use(middlewares.HandleError)

	// Health Route
	r.engine.GET("/healthz", healthController.Health)

	// V1 Routes
	v1 := r.engine.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			// Local Auth
			if config.AppConfig.Auth.Providers.Local.Enabled {

				localAuthService := services.NewLocalAuthService(userRepository, sessionService)
				localAuthController := controllers.NewLocalAuthController(localAuthService)

				localAuth := auth.Group("/local")
				{
					localAuth.POST("/login", localAuthController.Login)
					localAuth.POST("/register", localAuthController.Register)
				}
			}

			// Protected Routes
			auth.POST("/logout", authMiddleware.VerifyBearerToken, authController.Logout)
			auth.POST("/refresh", authMiddleware.VerifyRefreshToken, authController.Refresh)
		}
	}

	return r.engine
}
