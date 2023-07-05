package api

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/grassrootseconomics/cic-custodial/internal/custodial"

	"github.com/grassrootseconomics/cic-custodial/internal/tasker"
	"github.com/grassrootseconomics/cic-custodial/internal/tasker/task"
	"github.com/labstack/echo/v4"
)

// HandleSignTransfer godoc
//
//	@Summary		Sign and dispatch transfer request.
//	@Description	Sign and dispatch a transfer request.
//	@Tags			network
//	@Accept			json
//	@Produce		json
//	@Param			signTransferRequest	body		object{from=string,to=string,voucherAddress=string,amount=uint64}	true	"Sign Transfer Request"
//	@Success		200					{object}	OkResp
//	@Failure		400					{object}	ErrResp
//	@Failure		500					{object}	ErrResp
//	@Router			/sign/transfer [post]
func HandleSignTransfer(cu *custodial.Custodial) func(echo.Context) error {
	return func(c echo.Context) error {
		var (
			req struct {
				From           string `json:"from" validate:"required,eth_addr_checksum"`
				To             string `json:"to" validate:"required,eth_addr_checksum"`
				VoucherAddress string `json:"voucherAddress" validate:"required,eth_addr_checksum"`
				Amount         uint64 `json:"amount" validate:"gt=0"`
			}
		)

		if err := c.Bind(&req); err != nil {
			return NewBadRequestError(ErrInvalidJSON)
		}

		if err := c.Validate(req); err != nil {
			return err
		}

		accountActive, gasLock, err := cu.Store.GetAccountStatus(c.Request().Context(), req.From)
		if err != nil {
			return err
		}

		if !accountActive {
			return c.JSON(http.StatusForbidden, ErrResp{
				Ok:      false,
				Message: "Account pending activation. Try again later.",
			})
		}

		if gasLock {
			return c.JSON(http.StatusForbidden, ErrResp{
				Ok:      false,
				Message: "Gas lock. Gas balance unavailable. Try again later.",
			})
		}

		trackingId := uuid.NewString()

		taskPayload, err := json.Marshal(task.TransferPayload{
			TrackingId:     trackingId,
			From:           req.From,
			To:             req.To,
			VoucherAddress: req.VoucherAddress,
			Amount:         req.Amount,
		})
		if err != nil {
			return err
		}

		_, err = cu.TaskerClient.CreateTask(
			c.Request().Context(),
			tasker.SignTransferTask,
			tasker.HighPriority,
			&tasker.Task{
				Id:      trackingId,
				Payload: taskPayload,
			},
		)
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, OkResp{
			Ok: true,
			Result: H{
				"trackingId": trackingId,
			},
		})
	}
}
