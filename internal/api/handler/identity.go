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

type IdentityHandler struct {
	identityService service.IdentityService
	sessionService  service.SessionService
	cookieHandler   CookieHandler
	logger          *slog.Logger
}

func NewIdentityHandler(
	identityService service.IdentityService,
	sessionService service.SessionService,
	cookieHandler CookieHandler,
	logger *slog.Logger,
) *IdentityHandler {
	return &IdentityHandler{
		identityService: identityService,
		sessionService:  sessionService,
		cookieHandler:   cookieHandler,
		logger:          logger.With("handler", "IdentityHandler"),
	}
}

// Register godoc
// @Summary Register a new user account
// @Description Creates a new user account with email and password. A verification email will be sent to confirm the email address.
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body model.RegisterAccountPayload true "Registration data"
// @Success 201 "Account created successfully"
// @Failure 400 {object} model.ProblemJSON "Validation error"
// @Failure 409 {object} model.ProblemJSON "Email already exists"
// @Failure 500 {object} model.ProblemJSON "Internal server error"
// @Failure 501 {object} model.ProblemJSON "Password authentication not enabled"
// @Router /auth/register [post]
func (i *IdentityHandler) Register(c echo.Context) error {
	log := i.logger.With("method", "Register")

	var payload model.RegisterAccountPayload
	if err := c.Bind(&payload); err != nil {
		log.Error("error to bind payload", slog.String("error", err.Error()))
		return InvalidBind(c)
	}

	if err := c.Validate(&payload); err != nil {
		return ValidationError(c, err)
	}

	if err := i.identityService.Register(c.Request().Context(), payload.Name, payload.Email, payload.Password); err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			log.Warn("cannot register account, email already exists")
			return Conflict(c, "EMAIL_ALREADY_EXISTS", "could not create account, email already in use")
		}

		if errors.Is(err, domain.ErrPasswordAuthNotEnabled) {
			log.Warn("password authentication is not enabled")
			return NotImplemented(c, "this authentication method is not available")
		}

		log.Error("error to register account", slog.String("error", err.Error()))
		return InternalServerError(c)
	}

	return c.NoContent(http.StatusCreated)
}

// RegisterMagicLink godoc
// @Summary Register a new user account with magic link
// @Description Creates a new user account without password. A magic link will be sent to the email for verification and login.
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body model.RegisterMagicLinkPayload true "Registration data"
// @Success 201 "Account created successfully, magic link sent"
// @Failure 400 {object} model.ProblemJSON "Validation error"
// @Failure 409 {object} model.ProblemJSON "Email already exists"
// @Failure 500 {object} model.ProblemJSON "Internal server error"
// @Failure 501 {object} model.ProblemJSON "Magic link authentication not enabled"
// @Router /auth/register/magic-link [post]
func (i *IdentityHandler) RegisterMagicLink(c echo.Context) error {
	log := i.logger.With("method", "RegisterMagicLink")

	var payload model.RegisterMagicLinkPayload
	if err := c.Bind(&payload); err != nil {
		log.Error("error to bind payload", slog.String("error", err.Error()))
		return InvalidBind(c)
	}

	if err := c.Validate(&payload); err != nil {
		return ValidationError(c, err)
	}

	if err := i.identityService.RegisterMagicLink(c.Request().Context(), payload.Name, payload.Email); err != nil {
		if errors.Is(err, domain.ErrMagicLinkNotEnabled) {
			log.Warn("magic link authentication is not enabled")
			return NotImplemented(c, "this authentication method is not available")
		}

		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			log.Warn("cannot register account, email already exists")
			return Conflict(c, "EMAIL_ALREADY_EXISTS", "could not create account, email already in use")
		}

		log.Error("error to register magic link account", slog.String("error", err.Error()))
		return InternalServerError(c)
	}

	return c.NoContent(http.StatusCreated)
}

// ConfirmEmail godoc
// @Summary Confirm email address
// @Description Verifies the user's email address using the token sent via email. Creates a session upon successful verification.
// @Tags Auth
// @Accept json
// @Produce json
// @Param token query string true "Email verification token"
// @Success 204 "Email confirmed successfully"
// @Failure 400 {object} model.ProblemJSON "Invalid or expired token"
// @Failure 500 {object} model.ProblemJSON "Internal server error"
// @Router /auth/verify-email [get]
func (i *IdentityHandler) ConfirmEmail(c echo.Context) error {
	log := i.logger.With("method", "ConfirmEmail")

	var payload model.VerifyEmailPayload
	if err := c.Bind(&payload); err != nil {
		log.Error("error to bind payload", slog.String("error", err.Error()))
		return InvalidBind(c)
	}

	if err := c.Validate(&payload); err != nil {
		return ValidationError(c, err)
	}

	result, err := i.identityService.ConfirmEmail(c.Request().Context(), payload.Token)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidVerification) {
			log.Warn("invalid verification token provided")
			return BadRequest(c, "INVALID_TOKEN", "the provided token is invalid or has expired")
		}

		log.Error("error to confirm email", slog.String("error", err.Error()))
		return InternalServerError(c)
	}

	clientInfo := echoctx.GetClientInfo(c)

	session, err := i.sessionService.CreateSession(c.Request().Context(), result.UserID, clientInfo.IPAddress, clientInfo.DeviceName, clientInfo.UserAgent)
	if err != nil {
		log.Error("error to create session", slog.String("error", err.Error()))
		return InternalServerError(c)
	}

	i.cookieHandler.Set(c, session.Token, session.ExpiresAt)

	return c.NoContent(http.StatusNoContent)
}
