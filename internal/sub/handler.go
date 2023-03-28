package sub

import (
	"context"
	"encoding/json"

	"github.com/grassrootseconomics/cic-custodial/internal/store"
	"github.com/nats-io/nats.go"
)

func (s *Sub) handler(ctx context.Context, msg *nats.Msg) error {
	var (
		chainEvent store.MinimalTxInfo
	)

	if err := json.Unmarshal(msg.Data, &chainEvent); err != nil {
		return err
	}

	if err := s.cu.PgStore.UpdateOtxStatusFromChainEvent(ctx, chainEvent); err != nil {
		return err
	}

	switch msg.Subject {
	case "CHAIN.gas":
		if err := s.cu.PgStore.ResetGasQuota(ctx, chainEvent.To); err != nil {
			return err
		}
	}

	return nil
}
