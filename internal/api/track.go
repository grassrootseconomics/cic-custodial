package api

import (
	"net/http"

	"github.com/grassrootseconomics/cic-custodial/internal/custodial"
	"github.com/labstack/echo/v4"
)

// HandleTxStatus route.
// GET: /api/track/:trackingId
// Route param:
// trackingId -> tracking UUID
// Returns array of tx status.
func HandleTrackTx(c echo.Context) error {
	var (
		cu              = c.Get("cu").(*custodial.Custodial)
		txStatusRequest struct {
			TrackingId string `param:"trackingId" validate:"required,uuid"`
		}
	)

	if err := c.Bind(&txStatusRequest); err != nil {
		return NewBadRequestError(err)
	}

	if err := c.Validate(txStatusRequest); err != nil {
		return err
	}

	txs, err := cu.PgStore.GetTxStatusByTrackingId(c.Request().Context(), txStatusRequest.TrackingId)
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
