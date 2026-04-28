package errors

import (
	"errors"
	"log"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
)

var httpStatusByCode = map[string]int{
	"NOT_FOUND":        http.StatusNotFound,
	"CONFLICT":         http.StatusConflict,
	"VALIDATION_ERROR": http.StatusUnprocessableEntity,
	"FORBIDDEN":        http.StatusForbidden,
	"BAD_REQUEST":		http.StatusBadRequest,
}

type AppError struct {
	ErrCode    string `json:"code"`
	ErrMessage string `json:"error"`
}

func (e *AppError) Error() string {
	return e.ErrMessage
}

func (e *AppError) Code() string {
	return e.ErrCode
}

func NewAppError(code string, message string) *AppError {
	return &AppError{
		ErrCode:    code,
		ErrMessage: message,
	}
}

type ApiErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Details any    `json:"details,omitempty"`
}

type invalidParam struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

func validationInvalidParams(err error) []invalidParam {
	var validationErrs validator.ValidationErrors
	if !errors.As(err, &validationErrs) {
		return nil
	}

	result := make([]invalidParam, 0, len(validationErrs))
	for _, fieldErr := range validationErrs {
		reason := fieldErr.Tag()
		if reason == "required" {
			reason = "is required"
		}
		result = append(result, invalidParam{
			Name:   fieldErr.Field(),
			Reason: reason,
		})
	}
	return result
}

func CustomHTTPErrorHandler(c *echo.Context, err error) {
	if err == nil {
		return
	}

	if resp, uErr := echo.UnwrapResponse(c.Response()); uErr == nil {
		if resp.Committed {
			return
		}
	}

	// Default value
	status := http.StatusInternalServerError
	response := ApiErrorResponse{
		Error: "Internal Server Error",
		Code:  "INTERNAL_ERROR",
	}

	if echoErr, ok := errors.AsType[*echo.HTTPError](err); ok {
		status = echoErr.Code
		response.Error = echoErr.Message
		response.Code = "HTTP_ERROR"
	}

	if appErr, ok := errors.AsType[*AppError](err); ok {
		response.Error = appErr.ErrMessage
		response.Code = appErr.ErrCode

		if mappedStatus, ok := httpStatusByCode[appErr.Code()]; ok {
			status = mappedStatus
		} else {
			status = http.StatusBadRequest
		}
	}

	// Validation handler
	if _, ok := errors.AsType[validator.ValidationErrors](err); ok {
		status = http.StatusUnprocessableEntity
		response.Error = "Validation error"
		response.Code = "VALIDATION_ERROR"
		response.Details = validationInvalidParams(err)
	}

	// Server Security (500)
	if status >= http.StatusInternalServerError {
		log.Printf("[CRITICAL ERROR]: %v", err)

	}

	c.JSON(status, response)
}
