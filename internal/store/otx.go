package store

import (
	"context"
	"math/big"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/grassrootseconomics/cic-custodial/pkg/enum"
)

type (
	Otx struct {
		TrackingId    string
		Type          enum.OtxType
		RawTx         string
		TxHash        string
		From          string
		Data          string
		GasLimit      uint64
		TransferValue uint64
		GasPrice      *big.Int
		Nonce         uint64
	}
	txStatus struct {
		CreatedAt     time.Time `db:"created_at" json:"createdAt"`
		Status        string    `db:"status" json:"status"`
		TransferValue uint64    `db:"transfer_value" json:"transferValue"`
		TxHash        string    `db:"tx_hash" json:"txHash"`
		Type          string    `db:"type" json:"txType"`
	}
)

func (s *PgStore) CreateOtx(
	ctx context.Context,
	otx Otx,
) (uint, error) {
	var (
		id uint
	)

	if err := s.db.QueryRow(
		ctx,
		s.queries.CreateOTX,
		otx.TrackingId,
		otx.Type,
		otx.RawTx,
		otx.TxHash,
		otx.From,
		otx.Data,
		otx.GasPrice,
		otx.GasLimit,
		otx.TransferValue,
		otx.Nonce,
	).Scan(
		&id,
	); err != nil {
		return id, err
	}

	return id, nil
}

func (s *PgStore) GetTxStatus(
	ctx context.Context,
	trackingId string,
) (txStatus, error) {
	var (
		tx txStatus
	)

	rows, err := s.db.Query(
		ctx,
		s.queries.GetTxStatusByTrackingId,
		trackingId,
	)
	if err != nil {
		return tx, err
	}

	if err := pgxscan.ScanOne(
		&tx,
		rows,
	); err != nil {
		return tx, err
	}

	return tx, nil
}

func (s *PgStore) CreateDispatchStatus(
	ctx context.Context,
	otxId uint,
	otxStatus enum.OtxStatus,
) error {
	if _, err := s.db.Exec(
		ctx,
		s.queries.CreateDispatchStatus,
		otxId,
		otxStatus,
	); err != nil {
		return err
	}

	return nil
}

func (s *PgStore) UpdateDispatchStatus(
	ctx context.Context,
	txSuccess bool,
	txHash string,
	txBlock uint64,
) error {
	var (
		status = enum.SUCCESS
	)

	if !txSuccess {
		status = enum.REVERTED
	}

	if _, err := s.db.Exec(
		ctx,
		s.queries.UpdateDispatchStatus,
		txHash,
		status,
		txBlock,
	); err != nil {
		return err
	}

	return nil
}
