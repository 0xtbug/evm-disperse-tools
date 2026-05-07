package evm

import (
	"context"
	"fmt"
	"math/big"
	"net/url"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

// GasFees represents current gas fee information
type GasFees struct {
	GasPrice  *big.Int
	GasTipCap *big.Int
	GasFeeCap *big.Int
}

// RPCClient wraps ethereum RPC client functionality
type RPCClient struct {
	ethClient *ethclient.Client
	rpcClient *rpc.Client
	url       string
}

// NewRPCClient creates a new RPC client with the given URL
func NewRPCClient(rpcURL string) (*RPCClient, error) {
	// Validate URL format
	if err := validateURL(rpcURL); err != nil {
		return nil, fmt.Errorf("invalid RPC URL: %w", err)
	}

	rpcC, err := rpc.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RPC: %w", err)
	}

	ethC := ethclient.NewClient(rpcC)

	return &RPCClient{
		ethClient: ethC,
		rpcClient: rpcC,
		url:       rpcURL,
	}, nil
}

// validateURL checks if the URL is in valid HTTP/HTTPS format
func validateURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		return fmt.Errorf("URL must start with http:// or https://")
	}

	_, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	return nil
}

// GetBalance retrieves the balance of an address
func (rc *RPCClient) GetBalance(ctx context.Context, address string) (*big.Int, error) {
	if address == "" {
		return nil, fmt.Errorf("address cannot be empty")
	}

	account := normalizeAddress(address)

	balance, err := rc.ethClient.BalanceAt(ctx, common.HexToAddress(account), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	return balance, nil
}

// GetNonce retrieves the nonce for an address
func (rc *RPCClient) GetNonce(ctx context.Context, address string) (uint64, error) {
	if address == "" {
		return 0, fmt.Errorf("address cannot be empty")
	}

	account := normalizeAddress(address)

	nonce, err := rc.ethClient.PendingNonceAt(ctx, common.HexToAddress(account))
	if err != nil {
		return 0, fmt.Errorf("failed to get nonce: %w", err)
	}

	return nonce, nil
}

// SuggestFees retrieves current gas fees
func (rc *RPCClient) SuggestFees(ctx context.Context) (*GasFees, error) {
	gasPrice, err := rc.ethClient.SuggestGasPrice(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get gas price: %w", err)
	}

	header, err := rc.ethClient.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get header: %w", err)
	}

	gasTipCap, err := rc.ethClient.SuggestGasTipCap(ctx)
	if err != nil {
		// Fallback for networks that don't support EIP-1559
		gasTipCap = big.NewInt(0)
	}

	gasFeeCap := new(big.Int)
	if header.BaseFee != nil {
		gasFeeCap.Add(header.BaseFee, gasTipCap)
		gasFeeCap.Mul(gasFeeCap, big.NewInt(2))
	} else {
		gasFeeCap = gasPrice
	}

	return &GasFees{
		GasPrice:  gasPrice,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
	}, nil
}

// ChainID retrieves the chain ID
func (rc *RPCClient) ChainID(ctx context.Context) (*big.Int, error) {
	chainID, err := rc.ethClient.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	return chainID, nil
}

// Close closes the RPC client connection
func (rc *RPCClient) Close() {
	if rc.rpcClient != nil {
		rc.rpcClient.Close()
	}
}

// WaitForReceipt polls for a transaction receipt until it is available or the context expires.
func (rc *RPCClient) WaitForReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timed out waiting for receipt: %w", ctx.Err())
		case <-ticker.C:
			receipt, err := rc.ethClient.TransactionReceipt(ctx, txHash)
			if err != nil {
				continue // not mined yet
			}
			return receipt, nil
		}
	}
}
