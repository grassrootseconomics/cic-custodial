package noncestore

import "context"

// Noncestore represents a persistent distributed noncestore
type Noncestore interface {
	Peek(context.Context, string) (uint64, error)
	Acquire(context.Context, string) (uint64, error)
	Return(context.Context, string) (uint64, error)
	SyncNetworkNonce(context.Context, string) (uint64, error)
	SetNewAccountNonce(context.Context, string) error
}

// SystemNoncestore represents a standalone noncestore for a single system account
type SystemNoncestore interface {
	Peek() uint64
	Acquire() uint64
	Return()
}
