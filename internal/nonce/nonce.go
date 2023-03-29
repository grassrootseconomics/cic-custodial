package nonce

import "context"

// Noncestore defines how a nonce store should be implemented for any storage backend.
type Noncestore interface {
	Peek(context.Context, string) (uint64, error)
	Acquire(context.Context, string) (uint64, error)
	Return(context.Context, string) error
	SetAccountNonce(context.Context, string, uint64) error
}
