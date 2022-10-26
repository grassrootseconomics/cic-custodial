package api

import (
	"net/http"

	"github.com/grassrootseconomics/cic-custodial/internal/ethereum"
	"github.com/grassrootseconomics/cic-custodial/internal/keystore"
	tasker_client "github.com/grassrootseconomics/cic-custodial/internal/tasker/client"
	"github.com/labstack/echo/v4"
)

type registrationResponse struct {
	PublicKey string `json:"publicKey"`
	JobRef    string `json:"jobRef"`
}

func handleRegistration(c echo.Context) error {
	var (
		tc = c.Get("tasker_client").(*tasker_client.TaskerClient)
		ks = c.Get("keystore").(keystore.Keystore)
	)

	generatedKeyPair, err := ethereum.GenerateKeyPair()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "ERR_GEN_KEYPAIR")
	}

	if err := ks.WriteKeyPair(c.Request().Context(), generatedKeyPair); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "ERR_SAVE_KEYPAIR")
	}

	job, err := tc.CreateRegistrationTask(tasker_client.RegistrationPayload{
		PublicKey: generatedKeyPair.Public,
	}, tasker_client.SetNewAccountNonceTask)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "ERR_START_TASK_CHAIN")
	}

	return c.JSON(http.StatusOK, okResp{
		registrationResponse{
			PublicKey: generatedKeyPair.Public,
			JobRef:    job.ID,
		},
	})
}
