package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func NewBadRequestError(message ...interface{}) *echo.HTTPError {
	return echo.NewHTTPError(http.StatusBadRequest, message...)
}
