package api

import (
	"net/http"

	"github.com/grassrootseconomics/cic-custodial/internal/store"
	"github.com/labstack/echo/v4"
)

func TxStatus(store store.Store) func(echo.Context) error {
	return func(c echo.Context) error {
		var txStatusRequest struct {
			TrackingId string `param:"trackingId" validate:"required,uuid"`
		}

		if err := c.Bind(&txStatusRequest); err != nil {
			return err
		}

		if err := c.Validate(txStatusRequest); err != nil {
			return err
		}

		// TODO: handle potential timeouts
		txs, err := store.GetTxStatusByTrackingId(c.Request().Context(), txStatusRequest.TrackingId)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, OkResp{
			Ok: true,
			Result: H{
				"transactions": txs,
			},
		})
	}
}
