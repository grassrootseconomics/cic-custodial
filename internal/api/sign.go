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

// SignTxHandler route.
// POST: /api/sign/transfer
// JSON Body:
// from -> ETH address
// to -> ETH address
// voucherAddress -> ETH address
// amount -> int (6 d.p. precision)
// e.g. 1000000 = 1 VOUCHER
// Returns the task id.
func SignTransferHandler(cu *custodial.Custodial) func(echo.Context) error {
	return func(c echo.Context) error {
		trackingId := uuid.NewString()

		var transferRequest struct {
			From           string `json:"from" validate:"required,eth_checksum"`
			To             string `json:"to" validate:"required,eth_checksum"`
			VoucherAddress string `json:"voucherAddress" validate:"required,eth_checksum"`
			Amount         int64  `json:"amount" validate:"required,numeric"`
		}

		if err := c.Bind(&transferRequest); err != nil {
			return err
		}

		if err := c.Validate(transferRequest); err != nil {
			return err
		}

		// TODO: Checksum addresses
		taskPayload, err := json.Marshal(task.TransferPayload{
			TrackingId:     trackingId,
			From:           transferRequest.From,
			To:             transferRequest.To,
			VoucherAddress: transferRequest.VoucherAddress,
			Amount:         transferRequest.Amount,
		})
		if err != nil {
			return err
		}

		_, err = cu.TaskerClient.CreateTask(
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
