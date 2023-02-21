package main

import (
	"errors"
	"net/http"

	"github.com/VictoriaMetrics/metrics"
	"github.com/go-playground/validator/v10"
	"github.com/grassrootseconomics/cic-custodial/internal/api"
	"github.com/grassrootseconomics/cic-custodial/internal/custodial"
	"github.com/hibiken/asynq"
	"github.com/labstack/echo/v4"
)

// Bootstrap API server.
func initApiServer(custodialContainer *custodial.Custodial) *echo.Echo {
	lo.Debug("api: bootstrapping api server")
	server := echo.New()
	server.HideBanner = true
	server.HidePort = true

	server.HTTPErrorHandler = func(err error, c echo.Context) {
		// Handle asynq duplication errors across all api handlers.
		if errors.Is(err, asynq.ErrTaskIDConflict) {
			c.JSON(http.StatusForbidden, api.ErrResp{
				Ok:      false,
				Code:    api.DUPLICATE_ERROR,
				Message: "Request with duplicate tracking id submitted.",
			})
			return
		}

		if _, ok := err.(validator.ValidationErrors); ok {
			c.JSON(http.StatusForbidden, api.ErrResp{
				Ok:      false,
				Code:    api.VALIDATION_ERROR,
				Message: err.(validator.ValidationErrors).Error(),
			})
			return
		}

		// Log internal server error for further investigation.
		lo.Error("api:", "path", c.Path(), "err", err)

		c.JSON(http.StatusInternalServerError, api.ErrResp{
			Ok:      false,
			Code:    api.INTERNAL_ERROR,
			Message: "Internal server error.",
		})
	}

	if ko.Bool("service.metrics") {
		server.GET("/metrics", func(c echo.Context) error {
			metrics.WritePrometheus(c.Response(), true)
			return nil
		})
	}

	customValidator := validator.New()
	customValidator.RegisterValidation("eth_checksum", api.EthChecksumValidator)

	server.Validator = &api.Validator{
		ValidatorProvider: customValidator,
	}

	apiRoute := server.Group("/api")
	apiRoute.POST("/account/create", api.CreateAccountHandler(custodialContainer))
	apiRoute.POST("/sign/transfer", api.SignTransferHandler(custodialContainer))
	apiRoute.GET("/track/:trackingId", api.TxStatus(custodialContainer.PgStore))

	return server
}
