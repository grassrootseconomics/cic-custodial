package api

import (
	"github.com/grassrootseconomics/cic-custodial/internal/actions"
	tasker_client "github.com/grassrootseconomics/cic-custodial/internal/tasker/client"
	"github.com/labstack/echo/v4"
)

type okResp struct {
	Data interface{} `json:"data"`
}

type Opts struct {
	ActionsProvider *actions.ActionsProvider
	TaskerClient    *tasker_client.TaskerClient
}

func BootstrapHTTPServer(o Opts) *echo.Echo {
	server := echo.New()
	server.HideBanner = true
	server.HidePort = true

	server.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("actions", o.ActionsProvider)
			c.Set("tasker_client", o.TaskerClient)
			return next(c)
		}
	})

	api := server.Group("/api")
	api.GET("/register", handleRegistration)

	return server
}
