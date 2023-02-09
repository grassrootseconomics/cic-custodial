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

// CreateAccountHandler route.
// POST: /api/account/create
// JSON Body:
// trackingId -> Unique string
// Returns the public key.
func CreateAccountHandler(
	keystore keystore.Keystore,
	taskerClient *tasker.TaskerClient,
) func(echo.Context) error {
	return func(c echo.Context) error {
		var accountRequest struct {
			TrackingId string `json:"trackingId" validate:"required"`
		}

		if err := c.Bind(&accountRequest); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, errResp{
				Ok:   false,
				Code: INTERNAL_ERROR,
			})
		}

		if err := c.Validate(accountRequest); err != nil {
			return err
		}

		generatedKeyPair, err := keypair.Generate()
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, errResp{
				Ok:   false,
				Code: INTERNAL_ERROR,
			})
		}

		id, err := keystore.WriteKeyPair(c.Request().Context(), generatedKeyPair)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, errResp{
				Ok:   false,
				Code: INTERNAL_ERROR,
			})
		}

		taskPayload, err := json.Marshal(task.AccountPayload{
			PublicKey:  generatedKeyPair.Public,
			TrackingId: accountRequest.TrackingId,
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, errResp{
				Ok:   false,
				Code: INTERNAL_ERROR,
			})
		}

		_, err = taskerClient.CreateTask(
			tasker.PrepareAccountTask,
			tasker.DefaultPriority,
			&tasker.Task{
				Id:      accountRequest.TrackingId,
				Payload: taskPayload,
			},
		)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, errResp{
				Ok:   false,
				Code: INTERNAL_ERROR,
			})
		}

		return c.JSON(http.StatusOK, okResp{
			Ok: true,
			Result: H{
				"publicKey":   generatedKeyPair.Public,
				"custodialId": id,
			},
		})
	}
}
