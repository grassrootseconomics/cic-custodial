package api

import (
	"fmt"
	"math/big"
	"net/http"

	"github.com/grassrootseconomics/celoutils"
	"github.com/grassrootseconomics/cic-custodial/internal/custodial"
	"github.com/grassrootseconomics/w3-celo-patch"
	"github.com/grassrootseconomics/w3-celo-patch/module/eth"
	"github.com/labstack/echo/v4"
)

// HandleNetworkAccountStatus godoc
//	@Summary		Get an address's network balance and nonce.
//	@Description	Return network balance and nonce.
//	@Tags			network
//	@Accept			*/*
//	@Produce		json
//	@Param			address	path		string	true	"Account Public Key"
//	@Success		200		{object}	OkResp
//	@Failure		400		{object}	ErrResp
//	@Failure		500		{object}	ErrResp
//	@Router			/account/status/{address} [get]
func HandleNetworkAccountStatus(cu *custodial.Custodial) func(echo.Context) error {
	return func(c echo.Context) error {
		var (
			accountStatusRequest struct {
				Address string `param:"address" validate:"required,eth_addr_checksum"`
			}
			networkBalance big.Int
			networkNonce   uint64
		)

		if err := c.Bind(&accountStatusRequest); err != nil {
			return NewBadRequestError(ErrInvalidJSON)
		}

		if err := c.Validate(accountStatusRequest); err != nil {
			return err
		}

		if err := cu.CeloProvider.Client.CallCtx(
			c.Request().Context(),
			eth.Nonce(celoutils.HexToAddress(accountStatusRequest.Address), nil).Returns(&networkNonce),
			eth.Balance(celoutils.HexToAddress(accountStatusRequest.Address), nil).Returns(&networkBalance),
		); err != nil {
			return err
		}

		if networkNonce > 0 {
			networkNonce--
		}

		return c.JSON(http.StatusOK, OkResp{
			Ok: true,
			Result: H{
				"balance": fmt.Sprintf("%s CELO", w3.FromWei(&networkBalance, 18)),
				"nonce":   networkNonce,
			},
		})
	}
}
