package handler

import (
	"fmt"
	"net/http"

	"github.com/yourusername/project/internal/config"
	"github.com/yourusername/project/internal/interfaces"
	"github.com/yourusername/project/internal/middleware"
	"github.com/yourusername/project/internal/service"
	"github.com/yourusername/project/pkg/websocketManager"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type SuccessResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type UserResponse = interfaces.User

func NewRouter(
	cfg *config.Config,
	authHandler *AuthHandler,
	userHandler *UserHandler,
	authService service.AuthService,
	wsHandler *websocketManager.WebSocketHandler,
) *gin.Engine {
	r := gin.Default()
	r.Use(configureCORS())
	r.GET("/api-docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, SuccessResponse{Message: "Service is healthy"})
	})

	r.GET("/ws", wsHandler.HandleConnection)

	api := r.Group("/api/v1")
	{
		configurePublicRoutes(api, authHandler, userHandler)
		configureAuthenticatedRoutes(api, authHandler, userHandler, authService)
	}

	return r
}

func configureCORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func configurePublicRoutes(api *gin.RouterGroup, authHandler *AuthHandler, userHandler *UserHandler) {
	api.POST("/auth", authHandler.UserLogin)
	api.POST("/users", userHandler.CreateUser)
}

func configureAuthenticatedRoutes(api *gin.RouterGroup, authHandler *AuthHandler, userHandler *UserHandler, authService service.AuthService) {
	authorized := api.Group("/")
	authorized.Use(middleware.AuthMiddleware(authService))

	authorized.POST("/auth/logout", authHandler.UserLogout)
	authorized.POST("/auth/token", authHandler.ValidateToken)
	authorized.GET("/users", userHandler.GetUsers)
}

func StartServer(cfg *config.Config, router *gin.Engine) {
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	router.Run(addr)
}
