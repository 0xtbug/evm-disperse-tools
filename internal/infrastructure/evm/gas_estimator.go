package evm

import (
	"context"
	"fmt"
	"math/big"
	"strings"
)

// GasEstimator handles gas estimation for transactions
type GasEstimator struct {
	rpcClient *RPCClient
}

// NewGasEstimator creates a new gas estimator
func NewGasEstimator(rpcClient *RPCClient) *GasEstimator {
	return &GasEstimator{
		rpcClient: rpcClient,
	}
}

// normalizeAddress ensures an address has the 0x prefix.
func normalizeAddress(addr string) string {
	if !strings.HasPrefix(addr, "0x") {
		return "0x" + addr
	}
	return addr
}

// EstimateGas estimates the gas required for a transaction
func (ge *GasEstimator) EstimateGas(ctx context.Context, to, from, data string, value *big.Int) (uint64, error) {
	if to == "" {
		return 0, fmt.Errorf("to address cannot be empty")
	}

	if from == "" {
		return 0, fmt.Errorf("from address cannot be empty")
	}

	toAddr := normalizeAddress(to)
	fromAddr := normalizeAddress(from)

	callMsg := map[string]interface{}{
		"from": fromAddr,
		"to":   toAddr,
	}

	if data != "" {
		callMsg["data"] = data
	}

	if value != nil && value.Sign() > 0 {
		callMsg["value"] = "0x" + value.Text(16)
	}

	var result string
	err := ge.rpcClient.rpcClient.CallContext(ctx, &result, "eth_estimateGas", callMsg)
	if err != nil {
		return 0, fmt.Errorf("failed to estimate gas: %w", err)
	}

	// Parse the hex result
	gas := big.NewInt(0)
	gas.SetString(result, 0)

	return gas.Uint64(), nil
}
