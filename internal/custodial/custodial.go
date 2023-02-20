package custodial

import (
	"github.com/bsm/redislock"
	"github.com/grassrootseconomics/celoutils"
	"github.com/grassrootseconomics/cic-custodial/internal/events"
	"github.com/grassrootseconomics/cic-custodial/internal/keystore"
	"github.com/grassrootseconomics/cic-custodial/internal/nonce"
	"github.com/grassrootseconomics/cic-custodial/internal/store"
	"github.com/grassrootseconomics/cic-custodial/internal/tasker"
)

type Custodial struct {
	CeloProvider    *celoutils.Provider
	EventEmitter    events.EventEmitter
	Keystore        keystore.Keystore
	LockProvider    *redislock.Client
	Noncestore      nonce.Noncestore
	PgStore         store.Store
	SystemContainer *tasker.SystemContainer
	TaskerClient    *tasker.TaskerClient
}
