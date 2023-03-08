package store

import (
	"context"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/grassrootseconomics/cic-custodial/pkg/enum"
)

type TxStatus struct {
	Type          string    `db:"type" json:"txType"`
	TxHash        string    `db:"tx_hash" json:"txHash"`
	TransferValue uint64    `db:"transfer_value" json:"transferValue"`
	CreatedAt     time.Time `db:"created_at" json:"createdAt"`
	Status        string    `db:"status" json:"status"`
}

func (s *PostgresStore) CreateOtx(ctx context.Context, otx OTX) (uint, error) {
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
	).Scan(&id); err != nil {
		return id, err
	}

	return id, nil
}

func (s *PostgresStore) GetTxStatusByTrackingId(ctx context.Context, trackingId string) ([]*TxStatus, error) {
	var (
		txs []*TxStatus
	)

	if err := pgxscan.Select(
		ctx,
		s.db,
		&txs,
		s.queries.GetTxStatusByTrackingId,
		trackingId,
	); err != nil {
		return nil, err
	}

	return txs, nil
}

func (s *PostgresStore) CreateDispatchStatus(ctx context.Context, dispatch DispatchStatus) error {
	if _, err := s.db.Exec(
		ctx,
		s.queries.CreateDispatchStatus,
		dispatch.OtxId,
		dispatch.Status,
	); err != nil {
		return err
	}

	return nil
}

func (s *PostgresStore) UpdateOtxStatusFromChainEvent(ctx context.Context, chainEvent MinimalTxInfo) error {
	var (
		status = enum.SUCCESS
	)

	if !chainEvent.Success {
		status = enum.REVERTED
	}

	if _, err := s.db.Exec(
		ctx,
		s.queries.UpdateChainStatus,
		chainEvent.TxHash,
		status,
		chainEvent.Block,
	); err != nil {
		return err
	}

	return nil
}
