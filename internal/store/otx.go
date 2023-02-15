package store

import "context"

func (s *PostgresStore) CreateOTX(ctx context.Context, otx OTX) (uint, error) {
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
		otx.Nonce,
	).Scan(&id); err != nil {
		return id, err
	}

	return id, nil
}
