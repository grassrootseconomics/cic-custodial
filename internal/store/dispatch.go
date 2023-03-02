package store

import (
	"context"
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
