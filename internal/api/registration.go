package api

import (
	"encoding/json"
	"net/http"

	"github.com/grassrootseconomics/cic-custodial/internal/keystore"
	"github.com/grassrootseconomics/cic-custodial/internal/tasker"
	"github.com/grassrootseconomics/cic-custodial/internal/tasker/task"
	"github.com/grassrootseconomics/cic-custodial/pkg/keypair"
	"github.com/labstack/echo/v4"
)

type registrationResponse struct {
	PublicKey string `json:"publicKey"`
	TaskRef    string `json:"taskRef"`
}

func RegistrationHandler(
	taskerClient *tasker.TaskerClient,
	keystore keystore.Keystore,
) func(echo.Context) error {
	return func(c echo.Context) error {
		generatedKeyPair, err := keypair.Generate()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, errResp{
				Ok:    false,
				Error: KEYPAIR_ERROR,
			})
		}

		if err := keystore.WriteKeyPair(c.Request().Context(), generatedKeyPair); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, errResp{
				Ok:    false,
				Error: INTERNAL_ERROR,
			})
		}

		taskPayload, err := json.Marshal(task.SystemPayload{
			PublicKey: generatedKeyPair.Public,
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, errResp{
				Ok:    false,
				Error: JSON_MARSHAL_ERROR,
			})
		}

		task, err := taskerClient.CreateTask(
			tasker.PrepareAccountTask,
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
			Data: registrationResponse{
				PublicKey: generatedKeyPair.Public,
				TaskRef:    task.ID,
			},
		})
	}
}
