package port

import "context"

// ChainGateway defines the interface for chain operations
type ChainGateway interface {
	GetBalance(ctx context.Context, address string) (string, error)
	GetNonce(ctx context.Context, address string) (uint64, error)
	SuggestFees(ctx context.Context) (map[string]string, error)
	ChainID(ctx context.Context) (uint64, error)
}
