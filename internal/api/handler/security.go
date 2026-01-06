package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gbvillarinho/base-project/internal/api/echoctx"
	"github.com/gbvillarinho/base-project/internal/api/model"
	"github.com/gbvillarinho/base-project/internal/domain"
	"github.com/gbvillarinho/base-project/internal/service"
	"github.com/labstack/echo/v4"
)

type SecurityHandler struct {
	securityService service.SecurityService
	sessionService  service.SessionService
	cookieHandler   CookieHandler
	logger          *slog.Logger
}

func NewSecurityHandler(
	securityService service.SecurityService,
	sessionService service.SessionService,
	cookieHandler CookieHandler,
	logger *slog.Logger,
) *SecurityHandler {
	return &SecurityHandler{
		securityService: securityService,
		sessionService:  sessionService,
		cookieHandler:   cookieHandler,
		logger:          logger.With("handler", "security"),
	}
}

func (h *SecurityHandler) UpdatePassword(c echo.Context) error {
	userID := echoctx.GetUserID(c)
	log := h.logger.With(
		slog.String("func", "RevokeSession"),
		slog.String("user_id", userID.String()),
		slog.String("session_id", echoctx.GetSessionID(c).String()),
	)

	var payload model.UpdatePasswordPayload
	if err := c.Bind(&payload); err != nil {
		log.Error("error to bind payload", "error", err)
		return InvalidBind(c)
	}

	if err := c.Validate(&payload); err != nil {
		return ValidationError(c, err)
	}

	if err := h.securityService.UpdatePassword(c.Request().Context(), userID, payload.CurrentPassword, payload.NewPassword); err != nil {
		log.Error("error to update password", "error", err)
		return InternalServerError(c)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *SecurityHandler) ForgotPassword(c echo.Context) error {
	log := h.logger.With(slog.String("func", "ForgotPassword"))

	var payload model.ForgotPasswordPayload
	if err := c.Bind(&payload); err != nil {
		log.Error("error to bind payload", "error", err)
		return InvalidBind(c)
	}

	if err := c.Validate(&payload); err != nil {
		return ValidationError(c, err)
	}

	if err := h.securityService.StartPasswordReset(c.Request().Context(), payload.Email); err != nil {
		log.Error("error to start password reset", "error", err)
		return InternalServerError(c)
	}

	return c.NoContent(http.StatusOK)
}

func (h *SecurityHandler) ResetPassword(c echo.Context) error {
	log := h.logger.With(slog.String("func", "ResetPassword"))

	var payload model.ResetPasswordPayload
	if err := c.Bind(&payload); err != nil {
		log.Error("error to bind payload", "error", err)
		return InvalidBind(c)
	}

	if err := c.Validate(&payload); err != nil {
		return ValidationError(c, err)
	}

	result, err := h.securityService.ResetPassword(c.Request().Context(), payload.Token, payload.NewPassword)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidVerification) {
			log.Warn("error to reset password", "error", err)
			return BadRequest(c, "INVALID_TOKEN", "the provided token is invalid or has expired")
		}

		log.Error("error to reset password", "error", err)
		return InternalServerError(c)
	}

	clientInfo := echoctx.GetClientInfo(c)

	session, err := h.sessionService.CreateSession(c.Request().Context(), result.UserID, clientInfo.IPAddress, clientInfo.UserAgent, clientInfo.DeviceName)
	if err != nil {
		log.Error("error to create session after password reset", "error", err)
		return InternalServerError(c)
	}

	h.cookieHandler.Set(c, session.Token, session.ExpiresAt)

	return c.NoContent(http.StatusNoContent)
}

func (h *SecurityHandler) StartChangeEmail(c echo.Context) error {
	userID := echoctx.GetUserID(c)
	log := h.logger.With(
		slog.String("func", "StartChangeEmail"),
		slog.String("user_id", userID.String()),
		slog.String("session_id", echoctx.GetSessionID(c).String()),
	)

	var payload model.RequestEmailChangePayload
	if err := c.Bind(&payload); err != nil {
		log.Error("error to bind payload", "error", err)
		return InvalidBind(c)
	}

	if err := c.Validate(&payload); err != nil {
		return ValidationError(c, err)
	}

	if err := h.securityService.StartChangeEmail(c.Request().Context(), userID, payload.NewEmail); err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			log.Warn("the provided email is already in use", "error", err)
			return BadRequest(c, "EMAIL_ALREADY_EXISTS", "the provided email is already in use")
		}

		if errors.Is(err, domain.ErrEmailIsTheSame) {
			log.Warn("the provided email is the same as the current one", "error", err)
			return BadRequest(c, "EMAIL_IS_THE_SAME", "the provided email is the same as the current one")
		}

		log.Error("error to start email change", "error", err)
		return InternalServerError(c)
	}

	return c.NoContent(http.StatusOK)
}

func (h *SecurityHandler) ChangeEmail(c echo.Context) error {
	log := h.logger.With(slog.String("func", "ChangeEmail"))

	var payload model.ConfirmEmailChangePayload
	if err := c.Bind(&payload); err != nil {
		log.Error("error to bind payload", "error", err)
		return InvalidBind(c)
	}

	if err := c.Validate(&payload); err != nil {
		return ValidationError(c, err)
	}

	if err := h.securityService.ChangeEmail(c.Request().Context(), payload.Token); err != nil {
		if errors.Is(err, domain.ErrInvalidVerification) || errors.Is(err, domain.ErrInvalidVerificationPayload) {
			log.Warn("error to change email", "error", err)
			return BadRequest(c, "INVALID_TOKEN", "the provided token is invalid or has expired")
		}

		log.Error("error to change email", "error", err)
		return InternalServerError(c)
	}

	return c.NoContent(http.StatusNoContent)
}
