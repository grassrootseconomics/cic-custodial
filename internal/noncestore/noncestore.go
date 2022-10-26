package noncestore

import "context"

// Noncestore represents a persistent distributed noncestore
type Noncestore interface {
	Peek(context.Context, string) (uint64, error)
	Acquire(context.Context, string) (uint64, error)
	Return(context.Context, string) error
	SyncNetworkNonce(context.Context, string) (uint64, error)
	SetNewAccountNonce(context.Context, string) error
}
