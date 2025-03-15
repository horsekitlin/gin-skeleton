package main

import (
	"context"
	"fmt"
	"log"
	"net"

	userRepo "github.com/yourusername/project/internal/repository/user"
	userService "github.com/yourusername/project/internal/service/user"
	"github.com/yourusername/project/pkg/bootstrap"
	"github.com/yourusername/project/pkg/config"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	app := fx.New(
		bootstrap.Module(),

		// 提供用戶倉庫和服務
		fx.Provide(
			userRepo.NewRepository,
			userService.NewService,
			newGrpcServer,
		),

		// 註冊生命週期鉤子
		fx.Invoke(registerHooks),
	)

	app.Run()
}

// newGrpcServer 創建gRPC服務器
func newGrpcServer() *grpc.Server {
	server := grpc.NewServer()
	reflection.Register(server)
	return server
}

// registerHooks 註冊應用生命週期鉤子
func registerHooks(
	lc fx.Lifecycle,
	grpcServer *grpc.Server,
	cfg *config.ServiceConfig,
	userService userService.Service,
) {
	// 在這裡註冊gRPC處理器
	// userProto.RegisterUserServiceServer(grpcServer, userGrpcHandler.NewUserHandler(userService))

	// 啟動gRPC服務器
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			port := cfg.Server.Port

			listener, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
			if err != nil {
				return err
			}

			log.Printf("User service starting on port %s...", port)
			go func() {
				if err := grpcServer.Serve(listener); err != nil {
					log.Fatalf("Failed to serve: %v", err)
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Println("Stopping gRPC server...")
			grpcServer.GracefulStop()
			return nil
		},
	})
}
