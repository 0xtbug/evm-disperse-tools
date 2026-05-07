package usecase

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/0xtbug/evm-disperse-tools/internal/domain/entity"
	"github.com/0xtbug/evm-disperse-tools/internal/domain/port"
)

// Disperse handles both native and ERC20 token disperse operations.
type Disperse struct {
	chainGateway     port.ChainGateway
	disperseGateway  port.DisperseGateway
	reportRepository port.ReportRepository
	validatePlan     *ValidatePlan
}

// NewDisperse creates a new unified disperse use-case.
func NewDisperse(
	chainGateway port.ChainGateway,
	disperseGateway port.DisperseGateway,
	reportRepository port.ReportRepository,
	validatePlan *ValidatePlan,
) *Disperse {
	return &Disperse{
		chainGateway:     chainGateway,
		disperseGateway:  disperseGateway,
		reportRepository: reportRepository,
		validatePlan:     validatePlan,
	}
}

// Execute disperses tokens to multiple recipients in a single transaction.
func (d *Disperse) Execute(ctx context.Context, req *entity.DisperseRequest, chain *entity.Chain, fromAddress, privKey string) (*entity.ExecutionReport, error) {
	if err := d.validatePlan.ExecuteValidate(ctx, req, chain, fromAddress); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	if req.Mode == entity.TokenModeERC20 && req.Token == "" {
		return nil, fmt.Errorf("token address is required for ERC20 mode")
	}

	callData, err := d.disperseGateway.BuildCallData(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to build call data: %w", err)
	}

	value := d.calculateValue(req)

	txHash, err := d.disperseGateway.SendTx(ctx, fromAddress, privKey, callData, value)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	blockNumber, gasUsed, err := d.disperseGateway.ConfirmTx(ctx, txHash)
	if err != nil {
		report := d.newReport(txHash, "reverted", blockNumber, gasUsed, req, chain)
		_ = d.reportRepository.Save(report)
		return report, fmt.Errorf("transaction reverted on-chain: %w", err)
	}

	report := d.newReport(txHash, "confirmed", blockNumber, gasUsed, req, chain)
	if err := d.reportRepository.Save(report); err != nil {
		return nil, fmt.Errorf("failed to save report: %w", err)
	}

	return report, nil
}

// BatchExecute splits large recipient lists into batches and executes each sequentially.
func (d *Disperse) BatchExecute(ctx context.Context, req *entity.DisperseRequest, chain *entity.Chain, fromAddress, privKey string, batchSize int) ([]*entity.ExecutionReport, error) {
	if batchSize <= 0 {
		batchSize = 100
	}

	var reports []*entity.ExecutionReport

	for i := 0; i < len(req.Recipients); i += batchSize {
		end := i + batchSize
		if end > len(req.Recipients) {
			end = len(req.Recipients)
		}

		batchReq := &entity.DisperseRequest{
			Mode:       req.Mode,
			Recipients: req.Recipients[i:end],
			Amount:     req.Amount,
			Token:      req.Token,
		}

		report, err := d.Execute(ctx, batchReq, chain, fromAddress, privKey)
		if err != nil {
			return reports, fmt.Errorf("batch %d failed: %w", i/batchSize, err)
		}

		reports = append(reports, report)

		time.Sleep(2 * time.Second)
	}

	return reports, nil
}

// calculateValue returns the native token value to send (zero for ERC20).
func (d *Disperse) calculateValue(req *entity.DisperseRequest) *big.Int {
	if req.Mode == entity.TokenModeERC20 {
		return big.NewInt(0)
	}
	amount := new(big.Int)
	amount.SetString(req.Amount, 10)
	return new(big.Int).Mul(amount, big.NewInt(int64(len(req.Recipients))))
}

// newReport creates an ExecutionReport for the given transaction result.
func (d *Disperse) newReport(txHash, status string, blockNumber, gasUsed uint64, req *entity.DisperseRequest, chain *entity.Chain) *entity.ExecutionReport {
	token := "native"
	if req.Mode == entity.TokenModeERC20 {
		token = req.Token
	}
	return &entity.ExecutionReport{
		TxHash:      txHash,
		Status:      status,
		BlockNumber: blockNumber,
		GasUsed:     gasUsed,
		Recipients:  len(req.Recipients),
		TotalAmount: req.Amount,
		Token:       token,
		ChainName:   chain.Name,
		Timestamp:   time.Now(),
	}
}
