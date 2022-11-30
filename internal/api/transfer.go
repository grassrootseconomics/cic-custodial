package api

import (
	"encoding/json"
	"net/http"

	"github.com/grassrootseconomics/cic-custodial/internal/tasker"
	"github.com/labstack/echo/v4"
)

type (
	transferRequest struct {
		From           string `json:"from" validate:"required,eth_addr"`
		To             string `json:"to" validate:"required,eth_addr"`
		VoucherAddress string `json:"voucherAddress" validate:"required,eth_addr"`
		Amount         string `json:"amount" validate:"required,numeric"`
	}

	transferResponse struct {
		TaskRef string `json:"taskRef"`
	}
)

func TransferHandler(
	taskerClient *tasker.TaskerClient,
) func(echo.Context) error {
	return func(c echo.Context) error {
		transferPayload := new(transferRequest)

		if err := c.Bind(transferPayload); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, errResp{
				Ok:    false,
				Error: BIND_ERROR,
			})
		}
		if err := c.Validate(transferPayload); err != nil {
			return err
		}

		taskPayload, err := json.Marshal(transferPayload)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, errResp{
				Ok:    false,
				Error: JSON_MARSHAL_ERROR,
			})
		}

		task, err := taskerClient.CreateTask(
			tasker.TransferTokenTask,
			tasker.DefaultPriority,
			&tasker.Task{
				Payload: taskPayload,
			},
		)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, errResp{
				Ok:    false,
				Error: TASK_CHAIN_ERROR,
			})
		}

		return c.JSON(http.StatusOK, okResp{
			Ok: true,
			Data: transferResponse{
				TaskRef: task.ID,
			},
		})
	}
}
