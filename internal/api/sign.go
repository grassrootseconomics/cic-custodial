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

// HandleSignTransfer route.
// POST: /api/sign/transfer
// JSON Body:
// from -> ETH address
// to -> ETH address
// voucherAddress -> ETH address
// amount -> int (6 d.p. precision)
// e.g. 1000000 = 1 VOUCHER
// Returns the task id.
func HandleSignTransfer(cu *custodial.Custodial) func(echo.Context) error {
	return func(c echo.Context) error {
		var (
			req struct {
				From           string `json:"from" validate:"required,eth_addr_checksum"`
				To             string `json:"to" validate:"required,eth_addr_checksum"`
				VoucherAddress string `json:"voucherAddress" validate:"required,eth_addr_checksum"`
				Amount         uint64 `json:"amount" validate:"required"`
			}
		)

		if err := c.Bind(&req); err != nil {
			return NewBadRequestError(ErrInvalidJSON)
		}

		if err := c.Validate(req); err != nil {
			return err
		}

		accountActive, gasQuota, err := cu.PgStore.GetAccountStatusByAddress(c.Request().Context(), req.From)
		if err != nil {
			return err
		}

		if !accountActive {
			return c.JSON(http.StatusForbidden, ErrResp{
				Ok:      false,
				Message: "Account pending activation. Try again later.",
			})
		}

		trackingId := uuid.NewString()

		if gasQuota < 1 {
			gasRefillPayload, err := json.Marshal(task.AccountPayload{
				PublicKey:  req.From,
				TrackingId: trackingId,
			})
			if err != nil {
				return err
			}

			_, err = cu.TaskerClient.CreateTask(
				c.Request().Context(),
				tasker.AccountRefillGasTask,
				tasker.DefaultPriority,
				&tasker.Task{
					Id:      trackingId,
					Payload: gasRefillPayload,
				},
			)
			if err != nil {
				return err
			}

			return c.JSON(http.StatusForbidden, ErrResp{
				Ok:      false,
				Message: "Out of gas, refill pending. Try again later.",
			})
		}

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
