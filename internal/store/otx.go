package store

import (
	"context"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
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

func (s *PostgresStore) UpdateOtxStatus(ctx context.Context, status string) error {
	return nil
}
