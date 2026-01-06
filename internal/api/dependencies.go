package api

import (
	"github.com/gbvillarinho/base-project/config"
	"github.com/gbvillarinho/base-project/internal/api/handler"
	"github.com/gbvillarinho/base-project/internal/api/middleware"
	"github.com/gbvillarinho/base-project/internal/infra/client"
	"github.com/gbvillarinho/base-project/internal/infra/database"
	"github.com/gbvillarinho/base-project/internal/infra/database/sqlc"
	"github.com/gbvillarinho/base-project/internal/infra/notification"
	"github.com/gbvillarinho/base-project/internal/service"
	"github.com/gbvillarinho/base-project/logger"
	"github.com/gbvillarinho/base-project/pkg/injector"
	"go.uber.org/dig"
)

func ProvideDependencies() *dig.Container {
	container := dig.New()

	// Infrastructure
	injector.Provide(container, config.NewConfig)
	injector.Provide(container, database.NewConnection)
	injector.Provide(container, sqlc.NewStore)
	injector.Provide(container, logger.NewLogger)
	injector.Provide(container, client.NewResendClient)
	injector.Provide(container, notification.NewEmailNotification)

	// Services
	injector.Provide(container, service.NewVerificationService)
	injector.Provide(container, service.NewURLService)
	injector.Provide(container, service.NewEmailService)
	injector.Provide(container, service.NewSessionService)
	injector.Provide(container, service.NewUserService)
	injector.Provide(container, service.NewCookieService)
	injector.Provide(container, service.NewLoginService)
	injector.Provide(container, service.NewIdentityService)
	injector.Provide(container, service.NewSecurityService)

	// Handlers
	injector.Provide(container, handler.NewHealthHandler)
	injector.Provide(container, handler.NewSwaggerHandler)
	injector.Provide(container, handler.NewCookieHandler)
	injector.Provide(container, handler.NewUserHandler)
	injector.Provide(container, handler.NewSessionHandler)
	injector.Provide(container, handler.NewIdentityHandler)
	injector.Provide(container, handler.NewLoginHandler)
	injector.Provide(container, handler.NewSecurityHandler)

	// Middleware
	injector.Provide(container, middleware.NewAuthMiddleware)

	// Server
	injector.Provide(container, NewAPI)

	return container
}
