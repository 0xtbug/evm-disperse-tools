package usecase

import (
	"context"
	"fmt"
	"math/big"

	"github.com/0xtbug/evm-disperse-tools/internal/domain/entity"
	"github.com/0xtbug/evm-disperse-tools/internal/domain/port"
)

// ValidatePlan validates a disperse request before execution
type ValidatePlan struct {
	chainGateway port.ChainGateway
}

// NewValidatePlan creates a new validate plan use-case
func NewValidatePlan(chainGateway port.ChainGateway) *ValidatePlan {
	return &ValidatePlan{
		chainGateway: chainGateway,
	}
}

// ExecuteValidate validates a disperse request
func (vp *ValidatePlan) ExecuteValidate(ctx context.Context, req *entity.DisperseRequest, chain *entity.Chain, fromAddress string) error {
	// Validate request
	if err := req.Validate(); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	// Validate chain
	if chain == nil {
		return fmt.Errorf("chain cannot be nil")
	}

	if chain.Name == "" || chain.RPCURL == "" || chain.DisperseContract == "" {
		return fmt.Errorf("chain configuration incomplete")
	}

	// Validate from address
	if fromAddress == "" {
		return fmt.Errorf("from address cannot be empty")
	}

	// Validate recipients
	if len(req.Recipients) == 0 {
		return fmt.Errorf("recipients list cannot be empty")
	}

	// Check recipient validity (all must be valid addresses)
	for _, recipient := range req.Recipients {
		if recipient == "" || len(recipient) < 40 {
			return fmt.Errorf("invalid recipient address: %s", recipient)
		}
	}

	// Validate amount
	if req.Amount == "" {
		return fmt.Errorf("amount cannot be empty")
	}

	// For ERC20, validate token address
	if req.Mode == entity.TokenModeERC20 && req.Token == "" {
		return fmt.Errorf("token address required for ERC20 mode")
	}

	// Check chain connectivity
	chainID, err := vp.chainGateway.ChainID(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect to chain: %w", err)
	}

	if chainID == 0 {
		return fmt.Errorf("invalid chain ID: %d", chainID)
	}

	// Check balance
	balance, err := vp.chainGateway.GetBalance(ctx, fromAddress)
	if err != nil {
		return fmt.Errorf("failed to get balance: %w", err)
	}

	// Parse per-recipient amount (must be in wei/smallest unit)
	amountPerRecipient := new(big.Int)
	if _, ok := amountPerRecipient.SetString(req.Amount, 10); !ok {
		return fmt.Errorf("invalid amount: %s", req.Amount)
	}

	// Calculate total disperse amount: per-recipient × number of recipients
	totalDisperseAmount := new(big.Int).Mul(amountPerRecipient, big.NewInt(int64(len(req.Recipients))))

	balanceBig := new(big.Int)
	balanceBig.SetString(balance, 10)

	// For native disperse, ensure we have enough for the disperse amount + gas
	if req.Mode == entity.TokenModeNative {
		// Base gas + per-recipient gas estimate (each transfer ~21k-30k gas for storage writes)
		const baseGas int64 = 50000
		const gasPerRecipient int64 = 25000
		gasUnits := baseGas + gasPerRecipient*int64(len(req.Recipients))
		gasEstimate := big.NewInt(gasUnits)
		fees, err := vp.chainGateway.SuggestFees(ctx)
		if err == nil {
			if gasPriceStr, ok := fees["gasPrice"]; ok {
				gasPriceBig := new(big.Int)
				gasPriceBig.SetString(gasPriceStr, 10)
				gasEstimate.Mul(gasEstimate, gasPriceBig)
			}
		}

		totalRequired := new(big.Int).Add(totalDisperseAmount, gasEstimate)

		if balanceBig.Cmp(totalRequired) < 0 {
			return fmt.Errorf("insufficient balance: have %s, need %s (total disperse %s + gas %s)", balance, totalRequired.String(), totalDisperseAmount.String(), gasEstimate.String())
		}
	}

	return nil
}
