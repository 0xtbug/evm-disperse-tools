package port

import (
	"context"
	"math/big"

	"github.com/0xtbug/evm-disperse-tools/internal/domain/entity"
)

// DisperseGateway defines the interface for disperse operations
type DisperseGateway interface {
	BuildCallData(ctx context.Context, req *entity.DisperseRequest) (string, error)
	EstimateGas(ctx context.Context, callData string) (uint64, error)
	SendTx(ctx context.Context, from string, privKey string, callData string, value *big.Int) (string, error)
	ConfirmTx(ctx context.Context, txHash string) (blockNumber uint64, gasUsed uint64, err error)
}
