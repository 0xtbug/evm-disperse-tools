package usecase

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/0xtbug/evm-disperse-tools/internal/domain/entity"
	"github.com/0xtbug/evm-disperse-tools/internal/domain/port"
)

// EstimateResult contains gas estimation results
type EstimateResult struct {
	GasLimit      uint64
	GasPrice      string
	TotalGasCost  string
	EstimatedTime time.Duration
}

// EstimateDisperse estimates gas and costs for a disperse operation
type EstimateDisperse struct {
	chainGateway    port.ChainGateway
	disperseGateway port.DisperseGateway
}

// NewEstimateDisperse creates a new estimate disperse use-case
func NewEstimateDisperse(chainGateway port.ChainGateway, disperseGateway port.DisperseGateway) *EstimateDisperse {
	return &EstimateDisperse{
		chainGateway:    chainGateway,
		disperseGateway: disperseGateway,
	}
}

// ExecuteEstimate estimates the cost of a disperse operation
func (ed *EstimateDisperse) ExecuteEstimate(ctx context.Context, req *entity.DisperseRequest, chain *entity.Chain) (*EstimateResult, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	if chain == nil {
		return nil, fmt.Errorf("chain cannot be nil")
	}

	// Build call data
	callData, err := ed.disperseGateway.BuildCallData(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to build call data: %w", err)
	}

	// Estimate gas
	gasLimit, err := ed.disperseGateway.EstimateGas(ctx, callData)
	if err != nil {
		// Use fallback estimate
		gasLimit = uint64(len(req.Recipients)) * 40000
	}

	// Get gas price
	fees, err := ed.chainGateway.SuggestFees(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get fees: %w", err)
	}

	gasPrice := "0"
	if gp, ok := fees["gasPrice"]; ok {
		gasPrice = gp
	}

	// Calculate total gas cost
	gasPriceBig := new(big.Int)
	gasPriceBig.SetString(gasPrice, 10)

	totalCost := new(big.Int)
	totalCost.Mul(gasPriceBig, big.NewInt(int64(gasLimit)))

	// Estimate time based on block time (roughly 12-15 seconds per block)
	estimatedTime := time.Duration(15) * time.Second

	return &EstimateResult{
		GasLimit:      gasLimit,
		GasPrice:      gasPrice,
		TotalGasCost:  totalCost.String(),
		EstimatedTime: estimatedTime,
	}, nil
}

// EstimateMultiple estimates cost for multiple batches
func (ed *EstimateDisperse) EstimateMultiple(ctx context.Context, req *entity.DisperseRequest, chain *entity.Chain, batchSize int) (*EstimateResult, error) {
	if batchSize <= 0 {
		batchSize = 100
	}

	recipientCount := len(req.Recipients)
	batchCount := (recipientCount + batchSize - 1) / batchSize

	// Estimate for one batch and multiply
	estimate, err := ed.ExecuteEstimate(ctx, req, chain)
	if err != nil {
		return nil, err
	}

	totalCost := new(big.Int)
	totalCost.SetString(estimate.TotalGasCost, 10)
	totalCost.Mul(totalCost, big.NewInt(int64(batchCount)))

	return &EstimateResult{
		GasLimit:      estimate.GasLimit * uint64(batchCount),
		GasPrice:      estimate.GasPrice,
		TotalGasCost:  totalCost.String(),
		EstimatedTime: estimate.EstimatedTime * time.Duration(batchCount),
	}, nil
}
