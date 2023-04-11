package sub

import (
	"context"
	"encoding/json"

	"github.com/nats-io/nats.go"
)

type (
	ChainEvent struct {
		Block           uint64 `json:"block"`
		From            string `json:"from"`
		To              string `json:"to"`
		ContractAddress string `json:"contractAddress"`
		Success         bool   `json:"success"`
		TxHash          string `json:"transactionHash"`
		TxIndex         uint   `json:"transactionIndex"`
		Value           uint64 `json:"value"`
	}
)

func (s *Sub) processEventHandler(ctx context.Context, msg *nats.Msg) error {
	var (
		chainEvent ChainEvent
	)

	if err := json.Unmarshal(msg.Data, &chainEvent); err != nil {
		return err
	}

	if err := s.cu.Store.UpdateDispatchStatus(
		ctx,
		chainEvent.Success,
		chainEvent.TxHash,
		chainEvent.Block,
	); err != nil {
		return err
	}

	if chainEvent.Success {
		switch msg.Subject {
		case "CHAIN.register":
			if err := s.cu.Store.ActivateAccount(ctx, chainEvent.To); err != nil {
				return err
			}

			if err := s.cu.Store.ResetGasQuota(ctx, chainEvent.To); err != nil {
				return err
			}
		case "CHAIN.gas":
			if err := s.cu.Store.ResetGasQuota(ctx, chainEvent.To); err != nil {
				return err
			}
		}
	}

	return nil
}
