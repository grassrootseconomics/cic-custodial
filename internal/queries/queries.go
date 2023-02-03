package queries

import (
	"fmt"

	"github.com/knadh/goyesql/v2"
)

type Queries struct {
	// Keystore
	WriteKeyPair string `query:"write-key-pair"`
	LoadKeyPair  string `query:"load-key-pair"`
	// OTX
}

func LoadQueries(q goyesql.Queries) (*Queries, error) {
	loadedQueries := &Queries{}

	if err := goyesql.ScanToStruct(loadedQueries, q, nil); err != nil {
		return nil, fmt.Errorf("failed to scan queries %v", err)
	}

	return loadedQueries, nil
}
