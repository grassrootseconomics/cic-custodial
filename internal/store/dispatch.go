package store

import (
	"context"

	"github.com/grassrootseconomics/cic-custodial/pkg/enum"
)

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

func (s *PostgresStore) UpdateChainStatus(ctx context.Context, txHash string, status enum.OtxStatus, block uint64) error {
	if _, err := s.db.Exec(
		ctx,
		s.queries.UpdateChainStatus,
		txHash,
		status,
		block,
	); err != nil {
		return err
	}

	return nil
}
