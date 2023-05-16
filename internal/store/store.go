package store

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"os"

	"github.com/grassrootseconomics/cic-custodial/pkg/enum"
	"github.com/grassrootseconomics/cic-custodial/pkg/keypair"
	"github.com/grassrootseconomics/cic-custodial/pkg/util"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/tern/v2/migrate"
	"github.com/knadh/goyesql/v2"
)

type (
	Store interface {
		// Keypair related actions.
		LoadPrivateKey(context.Context, string) (*ecdsa.PrivateKey, error)
		WriteKeyPair(context.Context, keypair.Key) (uint, error)
		// Otx related actions.
		CreateOtx(context.Context, Otx) (uint, error)
		GetNextNonce(context.Context, string) (uint64, error)
		GetTxStatus(context.Context, string) (TxStatus, error)
		CreateDispatchStatus(context.Context, uint, enum.OtxStatus) error
		UpdateDispatchStatus(context.Context, bool, string, uint64) error
		// Account related actions.
		ActivateAccount(context.Context, string) error
		GetAccountStatus(context.Context, string) (bool, bool, error)
		// Gas quota related actions.
		GasLock(context.Context, string) error
		GasUnlock(context.Context, string) error
	}

	Opts struct {
		DSN                  string
		MigrationsFolderPath string
		QueriesFolderPath    string
	}

	PgStore struct {
		db      *pgxpool.Pool
		queries *queries
	}

	queries struct {
		// Keystore related queries.
		WriteKeyPair string `query:"write-key-pair"`
		LoadKeyPair  string `query:"load-key-pair"`
		// Otx related queries.
		CreateOTX               string `query:"create-otx"`
		GetNextNonce            string `query:"get-next-nonce"`
		GetTxStatusByTrackingId string `query:"get-tx-status-by-tracking-id"`
		CreateDispatchStatus    string `query:"create-dispatch-status"`
		UpdateDispatchStatus    string `query:"update-dispatch-status"`
		// Account related queries.
		ActivateAccount  string `query:"activate-account"`
		GetAccountStatus string `query:"get-account-status-by-address"`
		GasLock          string `query:"acc-gas-lock"`
		GasUnlock        string `query:"acc-gas-unlock"`
	}
)

func NewPgStore(o Opts) (Store, error) {
	parsedConfig, err := pgxpool.ParseConfig(o.DSN)
	if err != nil {
		return nil, err
	}

	dbPool, err := pgxpool.NewWithConfig(context.Background(), parsedConfig)
	if err != nil {
		return nil, err
	}

	queries, err := loadQueries(o.QueriesFolderPath)
	if err != nil {
		return nil, err
	}

	if err := runMigrations(context.Background(), dbPool, o.MigrationsFolderPath); err != nil {
		return nil, err
	}

	return &PgStore{
		db:      dbPool,
		queries: queries,
	}, nil
}

func loadQueries(queriesPath string) (*queries, error) {
	parsedQueries, err := goyesql.ParseFile(queriesPath)
	if err != nil {
		return nil, err
	}

	loadedQueries := &queries{}

	if err := goyesql.ScanToStruct(loadedQueries, parsedQueries, nil); err != nil {
		return nil, fmt.Errorf("failed to scan queries %v", err)
	}

	return loadedQueries, nil
}

func runMigrations(ctx context.Context, dbPool *pgxpool.Pool, migrationsPath string) error {
	ctx, cancel := context.WithTimeout(ctx, util.SLATimeout)
	defer cancel()

	conn, err := dbPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	migrator, err := migrate.NewMigrator(ctx, conn.Conn(), "schema_version")
	if err != nil {
		return err
	}

	if err := migrator.LoadMigrations(os.DirFS(migrationsPath)); err != nil {
		return err
	}

	if err := migrator.Migrate(ctx); err != nil {
		return err
	}

	return nil
}
