package main

import (
	"net/http"

	"github.com/arl/statsviz"
	"github.com/go-playground/validator"
	"github.com/grassrootseconomics/cic-custodial/internal/api"
	"github.com/labstack/echo/v4"
)

func initApiServer() *echo.Echo {
	server := echo.New()
	server.HideBanner = true
	server.HidePort = true

	if ko.Bool("service.statsviz_debug") {
		statsVizMux := http.NewServeMux()
		_ = statsviz.Register(statsVizMux)
		server.GET("/debug/statsviz/", echo.WrapHandler(statsVizMux))
		server.GET("/debug/statsviz/*", echo.WrapHandler(statsVizMux))
	}

	server.Validator = &api.CustomValidator{
		Validator: validator.New(),
	}

	apiRoute := server.Group("/api")

	apiRoute.POST("/register", api.RegistrationHandler(
		taskerClient,
		postgresKeystore,
	))

	apiRoute.POST("/transfer", api.TransferHandler(
		taskerClient,
	))

	return server
}