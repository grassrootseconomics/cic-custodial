package api

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/grassrootseconomics/cic-custodial/internal/custodial"
	"github.com/grassrootseconomics/cic-custodial/internal/tasker"
	"github.com/grassrootseconomics/cic-custodial/internal/tasker/task"
	"github.com/grassrootseconomics/cic-custodial/pkg/keypair"
	"github.com/labstack/echo/v4"
)

// HandleAccountCreate godoc
//	@Summary		Create a new custodial account.
//	@Description	Create a new custodial account.
//	@Tags			account
//	@Accept			*/*
//	@Produce		json
//	@Success		200	{object}	OkResp
//	@Failure		500	{object}	ErrResp
//	@Router			/account/create [post]
func HandleAccountCreate(cu *custodial.Custodial) func(echo.Context) error {
	return func(c echo.Context) error {
		generatedKeyPair, err := keypair.Generate()
		if err != nil {
			return err
		}

		id, err := cu.Store.WriteKeyPair(c.Request().Context(), generatedKeyPair)
		if err != nil {
			return err
		}

		trackingId := uuid.NewString()
		taskPayload, err := json.Marshal(task.AccountPayload{
			PublicKey:  generatedKeyPair.Public,
			TrackingId: trackingId,
		})
		if err != nil {
			return err
		}

		_, err = cu.TaskerClient.CreateTask(
			c.Request().Context(),
			tasker.AccountRegisterTask,
			tasker.DefaultPriority,
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
				"publicKey":   generatedKeyPair.Public,
				"custodialId": id,
				"trackingId":  trackingId,
			},
		})
	}
}
