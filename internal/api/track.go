package api

import (
	"net/http"

	"github.com/grassrootseconomics/cic-custodial/internal/custodial"
	"github.com/labstack/echo/v4"
)

// HandleTrackTx godoc
//	@Summary		Track an OTX (Origin transaction) status.
//	@Description	Track an OTX (Origin transaction) status.
//	@Tags			track
//	@Accept			*/*
//	@Produce		json
//	@Param			trackingId	path		string	true	"Tracking Id"
//	@Success		200			{object}	OkResp
//	@Failure		400			{object}	ErrResp
//	@Failure		500			{object}	ErrResp
//	@Router			/track/{trackingId} [get]
func HandleTrackTx(cu *custodial.Custodial) func(echo.Context) error {
	return func(c echo.Context) error {
		var (
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

		txs, err := cu.Store.GetTxStatus(c.Request().Context(), txStatusRequest.TrackingId)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, OkResp{
			Ok: true,
			Result: H{
				"transaction": txs,
			},
		})
	}
}
