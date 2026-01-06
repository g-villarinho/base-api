package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/g-villarinho/voxel-api/internal/api/echoctx"
	"github.com/g-villarinho/voxel-api/internal/api/model"
	"github.com/g-villarinho/voxel-api/internal/domain"
	"github.com/g-villarinho/voxel-api/internal/service"
	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	userService service.UserService
	logger      *slog.Logger
}

func NewUserHandler(
	userService service.UserService,
	logger *slog.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger.With(slog.String("handler", "user")),
	}
}

// GetProfile godoc
// @Summary      Get user profile
// @Description  Retrieves the authenticated user's profile information
// @Tags         User
// @Accept       json
// @Produce      json
// @Security     CookieAuth
// @Success      200  {object}  model.ProfileResponse  "Profile retrieved successfully"
// @Failure      401  {object}  model.ProblemJSON  "Unauthorized - authentication required"
// @Failure      404  {object}  model.ProblemJSON  "User not found"
// @Failure      500  {object}  model.ProblemJSON  "Internal server error"
// @Router       /user/profile [get]
func (h *UserHandler) GetProfile(c echo.Context) error {
	log := h.logger.With(
		slog.String("method", "GetProfile"),
		slog.String("user_id", echoctx.GetUserID(c).String()),
	)

	user, err := h.userService.GetUser(c.Request().Context(), echoctx.GetUserID(c))
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			log.Error("user not found", "error", err)
			return InternalServerError(c)
		}

		log.Error("get user profile", "error", err)
		return InternalServerError(c)
	}

	response := model.ProfileResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}

	return c.JSON(http.StatusOK, response)
}
