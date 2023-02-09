package main

import (
	"github.com/VictoriaMetrics/metrics"
	"github.com/go-playground/validator"
	"github.com/grassrootseconomics/cic-custodial/internal/api"
	"github.com/labstack/echo/v4"
)

// Bootstrap API server.
func initApiServer(custodialContainer *custodial) *echo.Echo {
	lo.Debug("api: bootstrapping api server")
	server := echo.New()
	server.HideBanner = true
	server.HidePort = true

	if ko.Bool("service.metrics") {
		server.GET("/metrics", func(c echo.Context) error {
			metrics.WritePrometheus(c.Response(), true)
			return nil
		})
	}

	server.Validator = &api.Validator{
		ValidatorProvider: validator.New(),
	}

	apiRoute := server.Group("/api")
	apiRoute.POST("/account/create", api.CreateAccountHandler(
		custodialContainer.keystore,
		custodialContainer.taskerClient,
	))

	apiRoute.POST("/sign/transfer", api.SignTransferHandler(
		custodialContainer.taskerClient,
	))

	return server
}
