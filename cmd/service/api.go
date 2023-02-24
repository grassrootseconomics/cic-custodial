package main

import (
	"net/http"
	"time"

	"github.com/VictoriaMetrics/metrics"
	"github.com/go-playground/validator/v10"
	"github.com/grassrootseconomics/cic-custodial/internal/api"
	"github.com/grassrootseconomics/cic-custodial/internal/custodial"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const (
	contextTimeout = 5
)

// Bootstrap API server.
func initApiServer(custodialContainer *custodial.Custodial) *echo.Echo {
	customValidator := validator.New()
	customValidator.RegisterValidation("eth_checksum", api.EthChecksumValidator)

	server := echo.New()
	server.HideBanner = true
	server.HidePort = true

	server.Validator = &api.Validator{
		ValidatorProvider: customValidator,
	}

	server.HTTPErrorHandler = customHTTPErrorHandler

	server.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("cu", custodialContainer)
			return next(c)
		}
	})
	server.Use(middleware.Recover())
	server.Use(middleware.BodyLimit("1M"))
	server.Use(middleware.ContextTimeout(time.Duration(contextTimeout * time.Second)))

	if ko.Bool("service.metrics") {
		server.GET("/metrics", func(c echo.Context) error {
			metrics.WritePrometheus(c.Response(), true)
			return nil
		})
	}

	apiRoute := server.Group("/api")
	apiRoute.POST("/account/create", api.HandleAccountCreate)
	apiRoute.POST("/sign/transfer", api.HandleSignTransfer)
	apiRoute.GET("/track/:trackingId", api.HandleTrackTx)

	return server
}

func customHTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	he, ok := err.(*echo.HTTPError)
	if ok {
		var errorMsg string

		if m, ok := he.Message.(error); ok {
			errorMsg = m.Error()
		} else if m, ok := he.Message.(string); ok {
			errorMsg = m
		}

		c.JSON(he.Code, api.ErrResp{
			Ok:      false,
			Message: errorMsg,
		})
	} else {
		lo.Error("api: echo error", "path", c.Path(), "err", err)

		c.JSON(http.StatusInternalServerError, api.ErrResp{
			Ok:      false,
			Message: "Internal server error.",
		})
	}
}