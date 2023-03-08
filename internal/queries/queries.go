package queries

import (
	"fmt"

	"github.com/knadh/goyesql/v2"
)

type Queries struct {
	// Keystore
	WriteKeyPair string `query:"write-key-pair"`
	LoadKeyPair  string `query:"load-key-pair"`
	// Store
	CreateOTX                  string `query:"create-otx"`
	CreateDispatchStatus       string `query:"create-dispatch-status"`
	ActivateAccount            string `query:"activate-account"`
	UpdateChainStatus          string `query:"update-chain-status"`
	GetTxStatusByTrackingId    string `query:"get-tx-status-by-tracking-id"`
	GetAccountActivationQuorum string `query:"get-account-activation-quorum"`
	GetAccountStatus           string `query:"get-account-status-by-address"`
	DecrGasQuota               string `query:"decr-gas-quota"`
	ResetGasQuota              string `query:"reset-gas-quota"`
}

func LoadQueries(q goyesql.Queries) (*Queries, error) {
	loadedQueries := &Queries{}

	if err := goyesql.ScanToStruct(loadedQueries, q, nil); err != nil {
		return nil, fmt.Errorf("failed to scan queries %v", err)
	}

	return loadedQueries, nil
}
