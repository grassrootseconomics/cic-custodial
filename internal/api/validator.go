package api

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
)

type CustomValidator struct {
	Validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.Validator.Struct(i); err != nil {
		fmt.Println(err)
		return echo.NewHTTPError(http.StatusBadRequest, errResp{
			Ok:    false,
			Error: VALIDATION_ERROR,
		})
	}
	return nil
}
