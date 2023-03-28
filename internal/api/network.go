package api

import (
	"fmt"
	"math/big"
	"net/http"

	"github.com/grassrootseconomics/cic-custodial/internal/custodial"
	"github.com/grassrootseconomics/w3-celo-patch"
	"github.com/grassrootseconomics/w3-celo-patch/module/eth"
	"github.com/labstack/echo/v4"
)

func HandleNetworkAccountStatus(c echo.Context) error {
	var (
		cu                   = c.Get("cu").(*custodial.Custodial)
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
		eth.Nonce(w3.A(accountStatusRequest.Address), nil).Returns(&networkNonce),
		eth.Balance(w3.A(accountStatusRequest.Address), nil).Returns(&networkBalance),
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
