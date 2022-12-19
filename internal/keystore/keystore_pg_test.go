package keystore

import (
	"context"
	"testing"

	"github.com/grassrootseconomics/cic-custodial/pkg/keypair"
	"github.com/grassrootseconomics/cic-custodial/pkg/logg"
	"github.com/grassrootseconomics/cic-custodial/pkg/postgres"
	"github.com/stretchr/testify/suite"
)

const (
	testDsn = "postgres://postgres:postgres@localhost:5432/cic_custodial"
)

type ItKeystoreSuite struct {
	suite.Suite
	Keystore Keystore
}

func TestItKeystoreSuite(t *testing.T) {
	suite.Run(t, new(ItKeystoreSuite))
}

func (s *ItKeystoreSuite) SetupSuite() {
	pgPool, err := postgres.NewPostgresPool(postgres.PostgresPoolOpts{
		DSN: testDsn,
	})
	s.Require().NoError(err)

	ks, err := NewPostgresKeytore(Opts{
		PostgresPool: pgPool,
		Logg: logg.NewLogg(logg.LoggOpts{
			Debug:  true,
			Caller: true,
		}),
	})
	s.Require().NoError(err)
	s.Keystore = ks
}

func (s *ItKeystoreSuite) Test_Write_And_Load_KeyPair() {
	ctx := context.Background()
	keypair, err := keypair.Generate()
	s.NoError(err)

	err = s.Keystore.WriteKeyPair(ctx, keypair)
	s.NoError(err)

	_, err = s.Keystore.LoadPrivateKey(ctx, keypair.Public)
	s.NoError(err)
}
