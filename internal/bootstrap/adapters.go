package bootstrap

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"

	"github.com/0xtbug/evm-disperse-tools/internal/domain/entity"
	"github.com/0xtbug/evm-disperse-tools/internal/infrastructure/evm"
)

// RPCClientAdapter adapts the EVM RPC client to the ChainGateway port
type RPCClientAdapter struct {
	rpcClient *evm.RPCClient
}

// NewRPCClientAdapter creates a new RPC client adapter
func NewRPCClientAdapter(rpcClient *evm.RPCClient) *RPCClientAdapter {
	return &RPCClientAdapter{
		rpcClient: rpcClient,
	}
}

// GetBalance implements the ChainGateway port
func (rca *RPCClientAdapter) GetBalance(ctx context.Context, address string) (string, error) {
	balance, err := rca.rpcClient.GetBalance(ctx, address)
	if err != nil {
		return "0", err
	}
	return balance.String(), nil
}

// GetNonce implements the ChainGateway port
func (rca *RPCClientAdapter) GetNonce(ctx context.Context, address string) (uint64, error) {
	return rca.rpcClient.GetNonce(ctx, address)
}

// SuggestFees implements the ChainGateway port
func (rca *RPCClientAdapter) SuggestFees(ctx context.Context) (map[string]string, error) {
	fees, err := rca.rpcClient.SuggestFees(ctx)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"gasPrice":  fees.GasPrice.String(),
		"gasTipCap": fees.GasTipCap.String(),
		"gasFeeCap": fees.GasFeeCap.String(),
	}, nil
}

// ChainID implements the ChainGateway port
func (rca *RPCClientAdapter) ChainID(ctx context.Context) (uint64, error) {
	chainID, err := rca.rpcClient.ChainID(ctx)
	if err != nil {
		return 0, err
	}
	return chainID.Uint64(), nil
}

// DisperseGatewayAdapter adapts the EVM disperse gateway to the DisperseGateway port
type DisperseGatewayAdapter struct {
	gateway *evm.DisperseContractGateway
}

// NewDisperseGatewayAdapter creates a new disperse gateway adapter
func NewDisperseGatewayAdapter(gateway *evm.DisperseContractGateway) *DisperseGatewayAdapter {
	return &DisperseGatewayAdapter{
		gateway: gateway,
	}
}

// BuildCallData implements the DisperseGateway port.
// req.Amount must be in the smallest unit (wei for native, smallest token unit for ERC20).
func (dga *DisperseGatewayAdapter) BuildCallData(ctx context.Context, req *entity.DisperseRequest) (string, error) {
	var callData []byte
	var err error

	// Parse per-recipient amount — must already be in wei/smallest-unit
	amountPerRecipient := new(big.Int)
	if _, ok := amountPerRecipient.SetString(req.Amount, 10); !ok {
		return "", fmt.Errorf("invalid amount: %s (must be an integer in smallest unit)", req.Amount)
	}

	amounts := make([]*big.Int, len(req.Recipients))
	for i := range amounts {
		amounts[i] = new(big.Int).Set(amountPerRecipient)
	}

	if req.Mode == entity.TokenModeNative {
		callData, err = dga.gateway.BuildNativeCallData(req.Recipients, amounts)
	} else {
		callData, err = dga.gateway.BuildERC20CallData(req.Token, req.Recipients, amounts)
	}

	if err != nil {
		return "", err
	}

	// Convert bytes to hex string
	hexStr := fmt.Sprintf("0x%x", callData)
	return hexStr, nil
}

// EstimateGas implements the DisperseGateway port
func (dga *DisperseGatewayAdapter) EstimateGas(ctx context.Context, callData string) (uint64, error) {
	// Return a reasonable fallback — real estimation requires from address
	// which isn't available through the current port signature
	return 300000, nil
}

// SendTx implements the DisperseGateway port.
// Converts the hex callData back to bytes and delegates to the real contract gateway.
func (dga *DisperseGatewayAdapter) SendTx(ctx context.Context, from string, privKey string, callData string, value *big.Int) (string, error) {
	if from == "" {
		return "", fmt.Errorf("from address is required")
	}
	if privKey == "" {
		return "", fmt.Errorf("private key is required")
	}
	if callData == "" {
		return "", fmt.Errorf("call data is required")
	}

	// Convert hex callData string to bytes
	callDataBytes := common.FromHex(callData)
	if len(callDataBytes) == 0 {
		return "", fmt.Errorf("invalid call data hex: %s", callData)
	}

	// Delegate to the real contract gateway which handles signing and sending
	txHash, err := dga.gateway.SendTransaction(ctx, from, privKey, callDataBytes, value)
	if err != nil {
		return "", err
	}

	return txHash, nil
}

// ConfirmTx implements the DisperseGateway port.
// Waits for the transaction receipt and checks if it succeeded on-chain.
func (dga *DisperseGatewayAdapter) ConfirmTx(ctx context.Context, txHash string) (uint64, uint64, error) {
	conf, err := dga.gateway.ConfirmTransaction(ctx, txHash)
	if err != nil && conf != nil {
		// Transaction was mined but reverted
		return conf.BlockNumber, conf.GasUsed, err
	}
	if err != nil {
		return 0, 0, err
	}
	return conf.BlockNumber, conf.GasUsed, nil
}
