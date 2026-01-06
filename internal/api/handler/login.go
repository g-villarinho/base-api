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

type LoginHandler struct {
	loginService   service.LoginService
	sessionService service.SessionService
	cookieHandler  CookieHandler
	logger         *slog.Logger
}

func NewLoginHandler(
	loginService service.LoginService,
	sessionService service.SessionService,
	cookieHandler CookieHandler,
	logger *slog.Logger,
) *LoginHandler {
	return &LoginHandler{
		loginService:   loginService,
		sessionService: sessionService,
		cookieHandler:  cookieHandler,
		logger:         logger,
	}
}

func (h *LoginHandler) Login(c echo.Context) error {
	log := h.logger.With("handler", "Login")

	var payload model.LoginPayload
	if err := c.Bind(&payload); err != nil {
		log.Error("failed to bind login payload", "error", err)
		return InvalidBind(c)
	}

	if err := c.Validate(&payload); err != nil {
		return ValidationError(c, err)
	}

	user, err := h.loginService.Authenticate(c.Request().Context(), payload.Email, payload.Password)
	if err != nil {
		if errors.Is(err, domain.ErrPasswordAuthNotEnabled) {
			log.Warn("password authentication is not enabled")
			return NotImplemented(c, "this authentication method is not available")
		}

		if errors.Is(err, domain.ErrInvalidCredentials) {
			log.Warn("invalid credentials provided")
			return Unauthorized(c, "INVALID_CREDENTIALS", "invalid email or password. Please try again.")
		}

		if errors.Is(err, domain.ErrUserBlocked) {
			log.Warn("blocked user attempted to login", "user_id", echoctx.GetUserID(c))
			return Forbidden(c, "USER_BLOCKED", "your account has been blocked. Please contact support.")
		}

		if errors.Is(err, domain.ErrEmailNotVerified) {
			log.Warn("unverified email attempted to login")
			return Forbidden(c, "EMAIL_NOT_VERIFIED", "your email address is not verified. Please verify your email before logging in.")
		}

		log.Error("error to authenticate user", "error", err)
		return InternalServerError(c)
	}

	clientInfo := echoctx.GetClientInfo(c)

	session, err := h.sessionService.CreateSession(c.Request().Context(), user.ID, clientInfo.IPAddress, clientInfo.UserAgent, clientInfo.DeviceName)
	if err != nil {
		log.Error("error to create session", "error", err)
		return InternalServerError(c)
	}

	h.cookieHandler.Set(c, session.Token, session.ExpiresAt)

	return c.NoContent(http.StatusOK)
}

func (h *LoginHandler) RequestMagicLink(c echo.Context) error {
	log := h.logger.With("handler", "RequestMagicLink")

	var payload model.RequestMagicLinkPayload
	if err := c.Bind(&payload); err != nil {
		log.Error("failed to bind request magic link payload", "error", err)
		return InvalidBind(c)
	}

	if err := c.Validate(&payload); err != nil {
		return ValidationError(c, err)
	}

	if err := h.loginService.RequestMagicLink(c.Request().Context(), payload.Email); err != nil {
		if errors.Is(err, domain.ErrMagicLinkNotEnabled) {
			log.Warn("magic link authentication is not enabled")
			return NotImplemented(c, "this authentication method is not available")
		}

		log.Error("error to request magic link", "error", err)
		return InternalServerError(c)
	}

	return c.NoContent(http.StatusOK)
}

func (h *LoginHandler) VerifyMagicLink(c echo.Context) error {
	log := h.logger.With("handler", "VerifyMagicLink")

	var payload model.VerifyMagicLinkPayload
	if err := c.Bind(&payload); err != nil {
		log.Error("failed to bind verify magic link payload", "error", err)
		return InvalidBind(c)
	}

	if err := c.Validate(&payload); err != nil {
		return ValidationError(c, err)
	}
	user, err := h.loginService.VerifyMagicLink(c.Request().Context(), payload.Token)
	if err != nil {
		if errors.Is(err, domain.ErrMagicLinkNotEnabled) {
			log.Warn("magic link authentication is not enabled")
			return NotImplemented(c, "this authentication method is not available")
		}

		if errors.Is(err, domain.ErrInvalidVerification) {
			log.Warn("invalid magic link token provided")
			return BadRequest(c, "INVALID_TOKEN", "the provided token is invalid or has expired")
		}

		if errors.Is(err, domain.ErrUserBlocked) {
			log.Warn("blocked user attempted to login", "user_id", echoctx.GetUserID(c))
			return Forbidden(c, "USER_BLOCKED", "your account has been blocked. Please contact support.")
		}

		log.Error("error to verify magic link", "error", err)
		return InternalServerError(c)
	}

	clientInfo := echoctx.GetClientInfo(c)

	session, err := h.sessionService.CreateSession(c.Request().Context(), user.ID, clientInfo.IPAddress, clientInfo.UserAgent, clientInfo.DeviceName)
	if err != nil {
		log.Error("error to create session", "error", err)
		return InternalServerError(c)
	}

	h.cookieHandler.Set(c, session.Token, session.ExpiresAt)
	return c.NoContent(http.StatusOK)
}
