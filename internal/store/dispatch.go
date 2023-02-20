package store

import (
	"context"
)

type Status string

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
