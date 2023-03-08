package store

import (
	"context"
)

func (s *PostgresStore) GetAccountStatusByAddress(ctx context.Context, publicAddress string) (bool, int, error) {
	var (
		accountActive bool
		gasQuota      int
	)

	if err := s.db.QueryRow(ctx, s.queries.GetAccountStatus, publicAddress).Scan(&accountActive, &gasQuota); err != nil {
		return false, 0, err
	}

	return accountActive, gasQuota, nil
}

func (s *PostgresStore) GetAccountActivationQuorum(ctx context.Context, trackingId string) (int, error) {
	var (
		quorum int
	)

	if err := s.db.QueryRow(ctx, s.queries.GetAccountActivationQuorum, trackingId).Scan(&quorum); err != nil {
		return 0, err
	}

	return quorum, nil
}

func (s *PostgresStore) DecrGasQuota(ctx context.Context, publicAddress string) error {
	if _, err := s.db.Exec(
		ctx,
		s.queries.DecrGasQuota,
		publicAddress,
	); err != nil {
		return err
	}

	return nil
}

func (s *PostgresStore) ResetGasQuota(ctx context.Context, publicAddress string) error {
	if _, err := s.db.Exec(
		ctx,
		s.queries.ResetGasQuota,
		publicAddress,
	); err != nil {
		return err
	}

	return nil
}

func (s *PostgresStore) ActivateAccount(ctx context.Context, publicAddress string) error {
	if _, err := s.db.Exec(
		ctx,
		s.queries.ActivateAccount,
		publicAddress,
	); err != nil {
		return err
	}

	return nil
}
