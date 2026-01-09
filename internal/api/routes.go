package api

import (
	"net/http"

	"github.com/gbvillarinho/base-project/config"
	"github.com/gbvillarinho/base-project/internal/api/handler"
	"github.com/gbvillarinho/base-project/internal/api/middleware"
	"github.com/labstack/echo/v4"
)

func registerHealthRoutes(e *echo.Echo, h *handler.HealthHandler) {
	e.GET("/health", h.Check)
}

func registerDevRoutes(e *echo.Echo, cfg *config.Config) {
	dev := e.Group("/dev")

	dev.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	dev.GET("/env", func(c echo.Context) error {
		return c.JSON(http.StatusOK, cfg)
	})
}

func registerAuthRoutes(
	e *echo.Echo,
	identityHandler *handler.IdentityHandler,
	loginHandler *handler.LoginHandler,
	securityHandler *handler.SecurityHandler,
	sessionHandler *handler.SessionHandler,
	authMiddleware *middleware.AuthMiddleware,
	cfg *config.Config,
) {
	authRateLimiter := middleware.AuthRateLimiter()
	auth := e.Group("/auth", middleware.NoCache())

	if cfg.IsPasswordAuthEnabled() {
		auth.POST("/register", identityHandler.Register, authRateLimiter)
		auth.POST("/login", loginHandler.Login, authRateLimiter)
		auth.POST("/forgot-password", securityHandler.ForgotPassword, authRateLimiter)
		auth.POST("/reset-password", securityHandler.ResetPassword, authRateLimiter)
		auth.PATCH("/password", securityHandler.UpdatePassword, authMiddleware.EnsuredAuthenticated)
	}

	if cfg.IsMagicLinkAuthEnabled() {
		auth.POST("/register/magic-link", identityHandler.RegisterMagicLink, authRateLimiter)
		auth.POST("/magic-link/start", loginHandler.RequestMagicLink, authRateLimiter)
		auth.GET("/magic-link/verify", loginHandler.VerifyMagicLink, authRateLimiter)
	}

	auth.GET("/verify-email", identityHandler.ConfirmEmail, authRateLimiter)
	auth.DELETE("/logout", sessionHandler.Logout, authMiddleware.EnsuredAuthenticated)
	auth.POST("/change-email/start", securityHandler.StartChangeEmail, authMiddleware.EnsuredAuthenticated)
	auth.POST("/change-email/confirm", securityHandler.ChangeEmail)
}

func registerUserRoutes(e *echo.Echo, h *handler.UserHandler, m *middleware.AuthMiddleware) {
	user := e.Group("/user", m.EnsuredAuthenticated, middleware.NoCache())

	user.GET("/profile", h.GetProfile)
}

func registerSessionRoutes(e *echo.Echo, h *handler.SessionHandler, m *middleware.AuthMiddleware) {
	session := e.Group("/sessions", m.EnsuredAuthenticated)

	session.DELETE("/:session_id", h.RevokeSession)
	session.DELETE("", h.RevokeAllSessions)
}

func registerSwaggerRoutes(e *echo.Echo, h *handler.SwaggerHandler) {
	e.GET("/docs", h.ServeSwaggerUI)

	swagger := e.Group("/swagger")
	swagger.GET("/doc.json", h.ServeSwaggerJSON)
}
