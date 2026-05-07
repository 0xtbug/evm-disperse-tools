package evm

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

// TxConfirmation holds the result of waiting for a transaction receipt
type TxConfirmation struct {
	TxHash      string
	BlockNumber uint64
	GasUsed     uint64
	Status      uint64 // 1 = success, 0 = revert
}

// DisperseContractGateway manages interactions with the disperse contract
type DisperseContractGateway struct {
	rpcClient    *RPCClient
	contractAddr common.Address
	abi          abi.ABI
	gasEstimator *GasEstimator
}

// NewDisperseContractGateway creates a new contract gateway
func NewDisperseContractGateway(rpcClient *RPCClient, contractAddr string, abiJSON string) (*DisperseContractGateway, error) {
	if contractAddr == "" {
		return nil, fmt.Errorf("contract address cannot be empty")
	}

	contractAddr = normalizeAddress(contractAddr)

	contractAddrParsed := common.HexToAddress(contractAddr)

	// Parse ABI
	contractABI, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	gasEst := NewGasEstimator(rpcClient)

	return &DisperseContractGateway{
		rpcClient:    rpcClient,
		contractAddr: contractAddrParsed,
		abi:          contractABI,
		gasEstimator: gasEst,
	}, nil
}

// BuildNativeCallData builds call data for disperseNative function
func (dcg *DisperseContractGateway) BuildNativeCallData(recipients []string, amounts []*big.Int) ([]byte, error) {
	if len(recipients) == 0 {
		return nil, fmt.Errorf("recipients list cannot be empty")
	}

	if len(recipients) != len(amounts) {
		return nil, fmt.Errorf("recipients and amounts must have the same length")
	}

	// Convert recipients to common.Address
	recipientAddrs := make([]common.Address, len(recipients))
	for i, recipient := range recipients {
		recipientAddrs[i] = common.HexToAddress(normalizeAddress(recipient))
	}

	// Pack the function call
	callData, err := dcg.abi.Pack("distributeEther", recipientAddrs, amounts)
	if err != nil {
		return nil, fmt.Errorf("failed to pack call data: %w", err)
	}

	return callData, nil
}

// BuildERC20CallData builds call data for disperseERC20 function
func (dcg *DisperseContractGateway) BuildERC20CallData(tokenAddr string, recipients []string, amounts []*big.Int) ([]byte, error) {
	if tokenAddr == "" {
		return nil, fmt.Errorf("token address cannot be empty")
	}

	if len(recipients) == 0 {
		return nil, fmt.Errorf("recipients list cannot be empty")
	}

	if len(recipients) != len(amounts) {
		return nil, fmt.Errorf("recipients and amounts must have the same length")
	}

	tokenAddrParsed := common.HexToAddress(normalizeAddress(tokenAddr))

	// Convert recipients to common.Address
	recipientAddrs := make([]common.Address, len(recipients))
	for i, recipient := range recipients {
		recipientAddrs[i] = common.HexToAddress(normalizeAddress(recipient))
	}

	// Pack the function call
	callData, err := dcg.abi.Pack("distribute", tokenAddrParsed, recipientAddrs, amounts)
	if err != nil {
		return nil, fmt.Errorf("failed to pack call data: %w", err)
	}

	return callData, nil
}

// EstimateGasNative estimates gas for native disperse
func (dcg *DisperseContractGateway) EstimateGasNative(ctx context.Context, from string, recipients []string, amounts []*big.Int) (uint64, error) {
	callData, err := dcg.BuildNativeCallData(recipients, amounts)
	if err != nil {
		return 0, fmt.Errorf("failed to build call data: %w", err)
	}

	// Calculate total value to send
	totalValue := big.NewInt(0)
	for _, amount := range amounts {
		totalValue.Add(totalValue, amount)
	}

	gas, err := dcg.gasEstimator.EstimateGas(ctx, dcg.contractAddr.Hex(), from, "0x"+fmt.Sprintf("%x", callData), totalValue)
	if err != nil {
		return 0, fmt.Errorf("failed to estimate gas: %w", err)
	}

	return gas, nil
}

// SendTransaction sends a transaction with the given call data
func (dcg *DisperseContractGateway) SendTransaction(ctx context.Context, from, privKey string, callData []byte, value *big.Int) (string, error) {
	if from == "" {
		return "", fmt.Errorf("from address cannot be empty")
	}

	if privKey == "" {
		return "", fmt.Errorf("private key cannot be empty")
	}

	// Parse private key
	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(privKey, "0x"))
	if err != nil {
		return "", fmt.Errorf("failed to parse private key: %w", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("invalid public key")
	}

	fromAddr := crypto.PubkeyToAddress(*publicKeyECDSA)

	// Get nonce
	nonce, err := dcg.rpcClient.GetNonce(ctx, fromAddr.Hex())
	if err != nil {
		return "", fmt.Errorf("failed to get nonce: %w", err)
	}

	// Get gas price
	fees, err := dcg.rpcClient.SuggestFees(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get fees: %w", err)
	}

	// Get chain ID
	chainID, err := dcg.rpcClient.ChainID(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get chain ID: %w", err)
	}

	// Estimate gas
	gas, err := dcg.gasEstimator.EstimateGas(ctx, dcg.contractAddr.Hex(), fromAddr.Hex(), "0x"+fmt.Sprintf("%x", callData), value)
	if err != nil {
		gas = 500000 // fallback
	}

	// Add 20% buffer to gas estimate
	gas = uint64(float64(gas) * 1.2)

	// Build transaction
	if value == nil {
		value = big.NewInt(0)
	}

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: fees.GasTipCap,
		GasFeeCap: fees.GasFeeCap,
		Gas:       gas,
		To:        &dcg.contractAddr,
		Value:     value,
		Data:      callData,
	})

	// Sign transaction
	signer := types.LatestSignerForChainID(chainID)
	signedTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Send transaction
	err = dcg.rpcClient.ethClient.SendTransaction(ctx, signedTx)
	if err != nil {
		return "", fmt.Errorf("failed to send transaction: %w", err)
	}

	return signedTx.Hash().Hex(), nil
}

// ConfirmTransaction waits for the transaction receipt and returns confirmation details.
// Returns an error if the transaction reverted on-chain.
func (dcg *DisperseContractGateway) ConfirmTransaction(ctx context.Context, txHash string) (*TxConfirmation, error) {
	if txHash == "" {
		return nil, fmt.Errorf("tx hash cannot be empty")
	}

	receipt, err := dcg.rpcClient.WaitForReceipt(ctx, common.HexToHash(txHash))
	if err != nil {
		return nil, fmt.Errorf("failed waiting for receipt: %w", err)
	}

	conf := &TxConfirmation{
		TxHash:      txHash,
		BlockNumber: receipt.BlockNumber.Uint64(),
		GasUsed:     receipt.GasUsed,
		Status:      receipt.Status,
	}

	if receipt.Status == 0 {
		return conf, fmt.Errorf("transaction reverted on-chain (block %d, gas used %d)", conf.BlockNumber, conf.GasUsed)
	}

	return conf, nil
}
