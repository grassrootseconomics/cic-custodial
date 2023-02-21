package store

import (
	"context"

	"github.com/grassrootseconomics/cic-custodial/pkg/enum"
)

type (
	OTX struct {
		TrackingId    string
		Type          enum.OtxType
		RawTx         string
		TxHash        string
		From          string
		Data          string
		GasLimit      uint64
		TransferValue uint64
		GasPrice      uint64
		Nonce         uint64
	}

	DispatchStatus struct {
		OtxId  uint
		Status enum.OtxStatus
	}

	Store interface {
		CreateOtx(ctx context.Context, otx OTX) (id uint, err error)
		CreateDispatchStatus(ctx context.Context, dispatch DispatchStatus) error
		GetTxStatusByTrackingId(ctx context.Context, trackingId string) ([]*TxStatus, error)
		UpdateChainStatus(ctx context.Context, txHash string, status enum.OtxStatus, block uint64) error
	}
)
