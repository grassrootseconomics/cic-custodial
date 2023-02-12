package store

import (
	"context"
)

type Status string

func (s *PostgresStore) CreateDispatchStatus(ctx context.Context, dispatch DispatchStatus) (uint, error) {
	var (
		id uint
	)

	if err := s.db.QueryRow(
		ctx,
		s.queries.CreateDispatchStatus,
		dispatch.OtxId,
		dispatch.Status,
	).Scan(&id); err != nil {
		return id, err
	}

	return id, nil
}
