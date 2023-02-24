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

// CreateAccountHandler route.
// POST: /api/account/create
// Returns the public key.
func HandleAccountCreate(c echo.Context) error {
	var (
		cu = c.Get("cu").(*custodial.Custodial)
	)

	generatedKeyPair, err := keypair.Generate()
	if err != nil {
		return err
	}

	id, err := cu.Keystore.WriteKeyPair(c.Request().Context(), generatedKeyPair)
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
		tasker.AccountPrepareTask,
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
