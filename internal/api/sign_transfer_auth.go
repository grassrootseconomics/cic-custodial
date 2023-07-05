package api

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/grassrootseconomics/cic-custodial/internal/custodial"
	"github.com/grassrootseconomics/cic-custodial/internal/tasker"
	"github.com/grassrootseconomics/cic-custodial/internal/tasker/task"
	"github.com/labstack/echo/v4"
)

// Max 10k vouchers per approval session
const approvalSafetyLimit = 10000 * 1000000

// HandleSignTransferAuthorization godoc
//
//	@Summary		Sign and dispatch a transfer authorization (approve) request.
//	@Description	Sign and dispatch a transfer authorization (approve) request.
//	@Tags			network
//	@Accept			json
//	@Produce		json
//	@Param			signTransferAuthorzationRequest	body		object{amount=uint64,authorizer=string,authorizedAddress=string,voucherAddress=string}	true	"Sign Transfer Authorization (approve) Request"
//	@Success		200								{object}	OkResp
//	@Failure		400								{object}	ErrResp
//	@Failure		500								{object}	ErrResp
//	@Router			/sign/transferAuth [post]
func HandleSignTranserAuthorization(cu *custodial.Custodial) func(echo.Context) error {
	return func(c echo.Context) error {
		var (
			req struct {
				Amount            uint64 `json:"amount" validate:"gte=0"`
				Authorizer        string `json:"authorizer" validate:"required,eth_addr_checksum"`
				AuthorizedAddress string `json:"authorizedAddress" validate:"required,eth_addr_checksum"`
				VoucherAddress    string `json:"voucherAddress" validate:"required,eth_addr_checksum"`
			}
		)

		if err := c.Bind(&req); err != nil {
			return NewBadRequestError(ErrInvalidJSON)
		}

		if err := c.Validate(req); err != nil {
			return err
		}

		accountActive, gasLock, err := cu.Store.GetAccountStatus(c.Request().Context(), req.Authorizer)
		if err != nil {
			return err
		}

		if req.Amount > approvalSafetyLimit {
			return c.JSON(http.StatusForbidden, ErrResp{
				Ok:      false,
				Message: "Approval amount per session exceeds 10k.",
			})
		}

		if !accountActive {
			return c.JSON(http.StatusForbidden, ErrResp{
				Ok:      false,
				Message: "Account pending activation. Try again later.",
			})
		}

		if gasLock {
			return c.JSON(http.StatusForbidden, ErrResp{
				Ok:      false,
				Message: "Gas lock. Gas balance unavailable. Try again later.",
			})
		}

		trackingId := uuid.NewString()

		taskPayload, err := json.Marshal(task.TransferAuthPayload{
			TrackingId:        trackingId,
			Amount:            req.Amount,
			Authorizer:        req.Authorizer,
			AuthorizedAddress: req.AuthorizedAddress,
			VoucherAddress:    req.VoucherAddress,
		})
		if err != nil {
			return err
		}

		_, err = cu.TaskerClient.CreateTask(
			c.Request().Context(),
			tasker.SignTransferTaskAuth,
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
				"trackingId": trackingId,
			},
		})
	}
}
