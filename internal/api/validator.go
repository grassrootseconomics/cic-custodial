package api

import (
	"net/http"

	"github.com/go-playground/validator"
	"github.com/labstack/echo/v4"
)

type Validator struct {
	ValidatorProvider *validator.Validate
}

func (v *Validator) Validate(i interface{}) error {
	if err := v.ValidatorProvider.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, ErrResp{
			Ok:   false,
			Code: VALIDATION_ERROR,
		})
	}
	return nil
}
