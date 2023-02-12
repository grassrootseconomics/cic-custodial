package store

import (
	"context"

	"github.com/grassrootseconomics/cic-custodial/pkg/status"
)

type (
	OTX struct {
		TrackingId string
		Type       string
		RawTx      string
		TxHash     string
		From       string
		Data       string
		GasPrice   uint64
		Nonce      uint64
	}

	DispatchStatus struct {
		OtxId  uint
		Status status.Status
	}

	Store interface {
		// OTX (Custodial originating transactions).
		CreateOTX(ctx context.Context, otx OTX) (id uint, err error)
		// Dispatch status.
		CreateDispatchStatus(ctx context.Context, dispatch DispatchStatus) (id uint, err error)
	}
)
