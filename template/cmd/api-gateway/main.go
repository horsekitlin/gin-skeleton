package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/yourusername/project/internal/gateway"
	"github.com/yourusername/project/pkg/bootstrap"
	"github.com/yourusername/project/pkg/config"
	"github.com/yourusername/project/pkg/middleware"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		bootstrap.Module(),

		// 提供API網關依賴
		fx.Provide(
			newGinRouter,
			newWebSocketUpgrader,
			gateway.NewServer,
		),

		// 註冊生命週期鉤子
		fx.Invoke(registerHooks),
	)

	app.Run()
}

// newGinRouter 創建Gin路由器
func newGinRouter(cfg *config.ServiceConfig) *gin.Engine {
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.Cors())

	return router
}

// newWebSocketUpgrader 創建WebSocket升級器
func newWebSocketUpgrader() websocket.Upgrader {
	return websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // 生產環境中應當更嚴格
		},
	}
}

// registerHooks 註冊應用生命週期鉤子
func registerHooks(
	lc fx.Lifecycle,
	router *gin.Engine,
	cfg *config.ServiceConfig,
	gatewayServer *gateway.Server,
) {
	// 設置API路由
	gatewayServer.SetupRoutes()

	// 創建HTTP服務器
	server := &http.Server{
		Addr:              ":" + cfg.Server.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 MB
	}

	// 生命週期鉤子
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Printf("API Gateway starting on port %s...", cfg.Server.Port)
			go func() {
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Fatalf("Failed to start server: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println("Shutting down API Gateway...")
			shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()

			if err := server.Shutdown(shutdownCtx); err != nil {
				log.Printf("Server forced to shutdown: %v", err)
				return err
			}
			return nil
		},
	})
}
