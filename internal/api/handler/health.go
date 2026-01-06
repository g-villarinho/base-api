package handler

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"github.com/gbvillarinho/base-project/internal/api/model"
	"github.com/labstack/echo/v4"
)

type HealthHandler struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewHealthHandler(db *sql.DB, logger *slog.Logger) *HealthHandler {
	return &HealthHandler{
		db:     db,
		logger: logger.With(slog.String("handler", "health")),
	}
}

// Check godoc
// @Summary health check endpoint
// @Description  Returns the health status of the API including database connectivity and version information
// @Tags         Health
// @Produce      JSON
// @Success      200  {object}  model.HealthResponse  "Service is healthy"
// @Failure      503  {object}  model.HealthResponse  "Service is unhealthy"
// @Router       /health [get]
func (h *HealthHandler) Check(c echo.Context) error {
	logger := h.logger.With(slog.String("method", "Check"))

	response := model.HealthResponse{
		Status: "ok",
	}

	if err := h.checkDatabase(c.Request().Context()); err != nil {
		response.Status = "unhealthy"
		logger.Error("health check failed",
			slog.String("status", "unhealthy"),
			slog.String("error", err.Error()),
		)

		return c.JSON(http.StatusServiceUnavailable, response)
	}

	return c.JSON(http.StatusOK, response)
}

func (h *HealthHandler) checkDatabase(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	return h.db.PingContext(pingCtx)
}
