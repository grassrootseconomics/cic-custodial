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
	CreateOTX               string `query:"create-otx"`
	CreateDispatchStatus    string `query:"create-dispatch-status"`
	UpdateChainStatus       string `query:"update-chain-status"`
	GetTxStatusByTrackingId string `query:"get-tx-status-by-tracking-id"`
}

func LoadQueries(q goyesql.Queries) (*Queries, error) {
	loadedQueries := &Queries{}

	if err := goyesql.ScanToStruct(loadedQueries, q, nil); err != nil {
		return nil, fmt.Errorf("failed to scan queries %v", err)
	}

	return loadedQueries, nil
}
