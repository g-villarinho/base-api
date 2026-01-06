package api

import (
	"net/http"

	"github.com/g-villarinho/voxel-api/config"
	"github.com/g-villarinho/voxel-api/internal/api/handler"
	"github.com/g-villarinho/voxel-api/internal/api/middleware"
	"github.com/labstack/echo/v4"
)

func registerDevRoutes(e *echo.Echo, config *config.Config) {
	dev := e.Group("/dev")

	dev.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	dev.GET("/env", func(c echo.Context) error {
		return c.JSON(http.StatusOK, config)
	})
}

func registerAuthRoutes(
	e *echo.Echo,
	ih *handler.IdentityHandler,
	lh *handler.LoginHandler,
	sch *handler.SecurityHandler,
	m *middleware.AuthMiddleware,
	sh *handler.SessionHandler,
	cfg *config.Config) {

	authRateLimiter := middleware.AuthRateLimiter()

	auth := e.Group("/auth", middleware.NoCache())

	if cfg.IsPasswordAuthEnabled() {
		auth.POST("/register", ih.Register, authRateLimiter)
		auth.POST("/login", lh.Login, authRateLimiter)
		auth.POST("/forgot-password", sch.ForgotPassword, authRateLimiter)
		auth.POST("/reset-password", sch.ResetPassword, authRateLimiter)
		auth.PATCH("/password", sch.UpdatePassword, m.EnsuredAuthenticated)
	}

	// Magic link routes (only when magic link auth is enabled)
	if cfg.IsMagicLinkAuthEnabled() {
		auth.POST("/register/magic-link", ih.RegisterMagicLink, authRateLimiter)
		auth.POST("/magic-link/start", lh.RequestMagicLink, authRateLimiter)
		auth.GET("/magic-link/verify", lh.VerifyMagicLink, authRateLimiter)
	}

	// Always available routes
	auth.GET("/verify-email", ih.ConfirmEmail, authRateLimiter)
	auth.DELETE("/logout", sh.Logout, m.EnsuredAuthenticated)
	auth.POST("/change-email/start", sch.StartChangeEmail, m.EnsuredAuthenticated)
	auth.POST("/change-email/confirm", sch.ChangeEmail)
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
	swagger := e.Group("/swagger")
	swagger.GET("/doc.json", h.ServeSwaggerJSON)

	e.GET("/docs", h.ServeScalarUI)
}

func registerHealthRoutes(e *echo.Echo, h *handler.HealthHandler) {
	e.GET("/health", h.Check)
}
