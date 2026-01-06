package handler

import (
	"errors"
	"net/http"

	"github.com/g-villarinho/voxel-api/pkg/validation"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type ErrorResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Errors  interface{} `json:"errors,omitempty"`
}

func InternalServerError(c echo.Context) error {
	return c.JSON(http.StatusInternalServerError, ErrorResponse{
		Code:    "INTERNAL_SERVER_ERROR",
		Message: "An unexpected error occurred",
	})
}

func NotFound(c echo.Context, code, message string) error {
	return c.JSON(http.StatusNotFound, ErrorResponse{
		Code:    code,
		Message: message,
	})
}

func BadRequest(c echo.Context, code, message string) error {
	return c.JSON(http.StatusBadRequest, ErrorResponse{
		Code:    code,
		Message: message,
	})
}

func Unauthorized(c echo.Context, code, message string) error {
	return c.JSON(http.StatusUnauthorized, ErrorResponse{
		Code:    code,
		Message: message,
	})
}

func Conflict(c echo.Context, code, message string) error {
	return c.JSON(http.StatusConflict, ErrorResponse{
		Code:    code,
		Message: message,
	})
}

func ValidationError(c echo.Context, err error) error {
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		lang := c.Request().Header.Get("Accept-Language")
		validationErrors := validation.FormatValidationErrors(err, lang)

		return c.JSON(http.StatusUnprocessableEntity, ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "Your request is not valid",
			Errors:  validationErrors,
		})
	}

	return BadRequest(c, "VALIDATION_ERROR", err.Error())
}

func Forbidden(c echo.Context, code, message string) error {
	return c.JSON(http.StatusForbidden, ErrorResponse{
		Code:    code,
		Message: message,
	})
}

func NotImplemented(c echo.Context, message string) error {
	return c.JSON(http.StatusNotImplemented, ErrorResponse{
		Code:    "NOT_IMPLEMENTED",
		Message: message,
	})
}

func InvalidBind(c echo.Context) error {
	return c.JSON(http.StatusBadRequest, ErrorResponse{
		Code:    "INVALID_JSON_PAYLOAD",
		Message: "Invalid request payload",
	})
}
