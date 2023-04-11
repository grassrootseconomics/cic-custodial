package store

import (
	"context"
)

func (s *PgStore) ActivateAccount(
	ctx context.Context,
	publicAddress string,
) error {
	if _, err := s.db.Exec(
		ctx,
		s.queries.ActivateAccount,
		publicAddress,
	); err != nil {
		return err
	}

	return nil
}

func (s *PgStore) GetAccountStatus(
	ctx context.Context,
	publicAddress string,
) (bool, int, error) {
	var (
		accountActive bool
		gasQuota      int
	)

	if err := s.db.QueryRow(
		ctx,
		s.queries.GetAccountStatus,
		publicAddress,
	).Scan(
		&accountActive,
		&gasQuota,
	); err != nil {
		return false, 0, err
	}

	return accountActive, gasQuota, nil
}

func (s *PgStore) DecrGasQuota(
	ctx context.Context,
	publicAddress string,
) error {
	if _, err := s.db.Exec(
		ctx,
		s.queries.DecrGasQuota,
		publicAddress,
	); err != nil {
		return err
	}

	return nil
}

func (s *PgStore) ResetGasQuota(
	ctx context.Context,
	publicAddress string,
) error {
	if _, err := s.db.Exec(
		ctx,
		s.queries.ResetGasQuota,
		publicAddress,
	); err != nil {
		return err
	}

	return nil
}
