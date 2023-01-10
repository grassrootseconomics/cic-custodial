package keystore

import (
	"context"
	"testing"

	"github.com/grassrootseconomics/cic-custodial/pkg/keypair"
	"github.com/grassrootseconomics/cic-custodial/pkg/logg"
	"github.com/grassrootseconomics/cic-custodial/pkg/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/suite"
	"github.com/zerodha/logf"
)

const (
	testDsn = "postgres://postgres:postgres@localhost:5432/cic_custodial"
)

type itKeystoreSuite struct {
	suite.Suite
	keystore Keystore
	pgPool   *pgxpool.Pool
	logg     logf.Logger
}

func TestItKeystoreSuite(t *testing.T) {
	suite.Run(t, new(itKeystoreSuite))
}

func (s *itKeystoreSuite) SetupSuite() {
	logg := logg.NewLogg(logg.LoggOpts{
		Debug:  true,
		Caller: true,
	})

	pgPool, err := postgres.NewPostgresPool(postgres.PostgresPoolOpts{
		DSN: testDsn,
	})
	s.Require().NoError(err)
	s.pgPool = pgPool
	s.logg = logg

	s.keystore, err = NewPostgresKeytore(Opts{
		PostgresPool: pgPool,
		Logg:         logg,
	})
	s.Require().NoError(err)
}

func (s *itKeystoreSuite) TearDownSuite() {
	_, err := s.pgPool.Exec(context.Background(), "DROP TABLE IF EXISTS keystore")
	s.Require().NoError(err)
}

func (s *itKeystoreSuite) Test_Write_And_Load_KeyPair() {
	ctx := context.Background()
	keypair, err := keypair.Generate()
	s.NoError(err)

	err = s.keystore.WriteKeyPair(ctx, keypair)
	s.NoError(err)

	_, err = s.keystore.LoadPrivateKey(ctx, keypair.Public)
	s.NoError(err)
}
