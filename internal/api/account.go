package api

import (
	"net/http"

	"github.com/grassrootseconomics/cic-custodial/internal/keystore"
	"github.com/grassrootseconomics/cic-custodial/internal/tasker"
	"github.com/grassrootseconomics/cic-custodial/pkg/keypair"
	"github.com/labstack/echo/v4"
)

// CreateAccountHandler route.
// POST: /api/account/create.
// Returns the public key and tasker account prep receipt.
func CreateAccountHandler(
	taskerClient *tasker.TaskerClient,
	keystore keystore.Keystore,
) func(echo.Context) error {
	return func(c echo.Context) error {
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

		return c.JSON(http.StatusOK, okResp{
			Ok: true,
			Result: H{
				"publicKey": generatedKeyPair.Public,
				"keyId":     id,
			},
		})
	}
}

// AccountStatusHandler route.
// GET: /api/account/status.
// Check if an account is ready to be used.
// Returns the status as a bool.
func AccountStatusHandler() func(echo.Context) error {
	return func(c echo.Context) error {
		return c.JSON(http.StatusOK, okResp{
			Ok: true,
		})
	}
}
