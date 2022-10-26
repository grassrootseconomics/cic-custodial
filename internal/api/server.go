package api

import (
	"net/http"

	"github.com/arl/statsviz"
	"github.com/grassrootseconomics/cic-custodial/internal/keystore"
	tasker_client "github.com/grassrootseconomics/cic-custodial/internal/tasker/client"
	"github.com/labstack/echo/v4"
)

type okResp struct {
	Data interface{} `json:"data"`
}

type Opts struct {
	Keystore     keystore.Keystore
	TaskerClient *tasker_client.TaskerClient
}

func BootstrapHTTPServer(o Opts) *echo.Echo {
	server := echo.New()
	server.HideBanner = true
	server.HidePort = true

	// Debug
	statsVizMux := http.NewServeMux()
	_ = statsviz.Register(statsVizMux)
	server.GET("/debug/statsviz/", echo.WrapHandler(statsVizMux))
	server.GET("/debug/statsviz/*", echo.WrapHandler(statsVizMux))

	server.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("keystore", o.Keystore)
			c.Set("tasker_client", o.TaskerClient)
			return next(c)
		}
	})

	api := server.Group("/api")
	api.GET("/register", handleRegistration)

	return server
}
