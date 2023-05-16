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
) (bool, bool, error) {
	var (
		accountActive bool
		gasLock       bool
	)

	if err := s.db.QueryRow(
		ctx,
		s.queries.GetAccountStatus,
		publicAddress,
	).Scan(
		&accountActive,
		&gasLock,
	); err != nil {
		return false, false, err
	}

	return accountActive, gasLock, nil
}

func (s *PgStore) GasLock(
	ctx context.Context,
	publicAddress string,
) error {
	if _, err := s.db.Exec(
		ctx,
		s.queries.GasLock,
		publicAddress,
	); err != nil {
		return err
	}

	return nil
}

func (s *PgStore) GasUnlock(
	ctx context.Context,
	publicAddress string,
) error {
	if _, err := s.db.Exec(
		ctx,
		s.queries.GasUnlock,
		publicAddress,
	); err != nil {
		return err
	}

	return nil
}
