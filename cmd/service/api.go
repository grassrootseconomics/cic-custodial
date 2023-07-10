package main

import (
	"net/http"

	"github.com/VictoriaMetrics/metrics"
	"github.com/go-playground/validator/v10"
	_ "github.com/grassrootseconomics/cic-custodial/docs"
	"github.com/grassrootseconomics/cic-custodial/internal/api"
	"github.com/grassrootseconomics/cic-custodial/internal/custodial"
	"github.com/grassrootseconomics/cic-custodial/pkg/util"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

const (
	systemGlobalLockKey = "system:global_lock"
)

// Bootstrap API server.
func initApiServer(custodialContainer *custodial.Custodial) *echo.Echo {
	customValidator := validator.New()

	server := echo.New()
	server.HideBanner = true
	server.HidePort = true
	server.Validator = &api.Validator{
		ValidatorProvider: customValidator,
	}
	server.HTTPErrorHandler = customHTTPErrorHandler

	server.Use(middleware.Recover())
	server.Use(middleware.BodyLimit("1M"))
	server.Use(middleware.ContextTimeout(util.SLATimeout))

	if ko.Bool("service.metrics") {
		server.GET("/metrics", func(c echo.Context) error {
			metrics.WritePrometheus(c.Response(), true)
			return nil
		})
	}

	if ko.Bool("service.docs") {
		server.GET("/docs/*", echoSwagger.WrapHandler)
	}

	apiRoute := server.Group("/api", systemGlobalLock(custodialContainer))

	apiRoute.POST("/account/create", api.HandleAccountCreate(custodialContainer))
	apiRoute.GET("/account/status/:address", api.HandleNetworkAccountStatus(custodialContainer))
	apiRoute.POST("/sign/transfer", api.HandleSignTransfer(custodialContainer))
	apiRoute.POST("/sign/transferAuth", api.HandleSignTranserAuthorization(custodialContainer))
	apiRoute.GET("/track/:trackingId", api.HandleTrackTx(custodialContainer))

	return server
}

func customHTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	if he, ok := err.(*echo.HTTPError); ok {
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
		return
	}

	lo.Error("api: echo error", "path", c.Path(), "err", err)
	c.JSON(http.StatusInternalServerError, api.ErrResp{
		Ok:      false,
		Message: "Internal server error.",
	})
}

func systemGlobalLock(cu *custodial.Custodial) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			locked, err := cu.RedisClient.Get(c.Request().Context(), systemGlobalLockKey).Bool()
			if err != nil {
				return err
			}

			if locked {
				return c.JSON(http.StatusServiceUnavailable, api.ErrResp{
					Ok:      false,
					Message: "System manually locked.",
				})
			}

			return next(c)
		}
	}
}
