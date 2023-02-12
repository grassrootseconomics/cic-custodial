package api

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/grassrootseconomics/cic-custodial/internal/keystore"
	"github.com/grassrootseconomics/cic-custodial/internal/tasker"
	"github.com/grassrootseconomics/cic-custodial/internal/tasker/task"
	"github.com/grassrootseconomics/cic-custodial/pkg/keypair"
	"github.com/labstack/echo/v4"
)

// CreateAccountHandler route.
// POST: /api/account/create
// Returns the public key.
func CreateAccountHandler(
	keystore keystore.Keystore,
	taskerClient *tasker.TaskerClient,
) func(echo.Context) error {
	return func(c echo.Context) error {
		trackingId := uuid.NewString()

		generatedKeyPair, err := keypair.Generate()
		if err != nil {
			return err
		}

		id, err := keystore.WriteKeyPair(c.Request().Context(), generatedKeyPair)
		if err != nil {
			return err
		}

		taskPayload, err := json.Marshal(task.AccountPayload{
			PublicKey:  generatedKeyPair.Public,
			TrackingId: trackingId,
		})
		if err != nil {
			return err
		}

		_, err = taskerClient.CreateTask(
			tasker.PrepareAccountTask,
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
