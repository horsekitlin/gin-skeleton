package main

import (
	"log"
	"github.com/yourusername/project/internal/config"
	"github.com/yourusername/project/internal/handler"
	"github.com/yourusername/project/internal/service"

	_ "github.com/yourusername/project/docs"
	"github.com/yourusername/project/pkg/core"
	"github.com/yourusername/project/pkg/utils"

	"go.uber.org/fx"
)

// @title           Passontw Auth Service API
// @description     Passontw Auth Service API.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @BasePath  /
func main() {
	if err := utils.InitSnowflake(2); err != nil {
		log.Fatalf("Failed to initialize Snowflake: %v", err)
	}

	app := fx.New(
		core.Module,

		config.Module,
		service.Module,
		handler.Module,
	)

	app.Run()
}
