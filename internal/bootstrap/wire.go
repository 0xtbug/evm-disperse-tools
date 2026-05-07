package bootstrap

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"io/fs"
	"math/big"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"

	evmdisperse "github.com/0xtbug/evm-disperse-tools"
	"github.com/0xtbug/evm-disperse-tools/internal/application/usecase"
	"github.com/0xtbug/evm-disperse-tools/internal/domain/entity"
	"github.com/0xtbug/evm-disperse-tools/internal/infrastructure/config"
	"github.com/0xtbug/evm-disperse-tools/internal/infrastructure/evm"
	"github.com/0xtbug/evm-disperse-tools/internal/infrastructure/storage"
	"github.com/0xtbug/evm-disperse-tools/internal/presentation/tui"
)

// chainInfra holds per-chain infrastructure so that each chain gets its own
// RPC client, contract gateway, and use-case instances.
type chainInfra struct {
	rpcClient *evm.RPCClient
	disperse  *usecase.Disperse
}

// BuildApp builds the complete application with all dependencies wired
func BuildApp() (*tui.AppModel, error) {
	// Load chain configurations
	chains, err := loadChainConfigs()
	if err != nil {
		return nil, fmt.Errorf("failed to load chain configs: %w", err)
	}

	if len(chains) == 0 {
		return nil, fmt.Errorf("no chains configured")
	}

	// Load disperse contract ABI (embedded in binary)
	abiJSON := string(evmdisperse.DisperseABI)

	// Create shared repositories
	reportRepo := storage.NewReportsFileRepo("data/reports")

	// Create per-chain infrastructure
	chainInfras := make(map[string]*chainInfra)
	for _, chainCfg := range chains {
		rpcClient, err := evm.NewRPCClient(chainCfg.RPCURL)
		if err != nil {
			// Log but don't fail — other chains may still work
			fmt.Fprintf(os.Stderr, "warning: failed to connect to %s (%s): %v\n", chainCfg.Name, chainCfg.RPCURL, err)
			continue
		}

		chainGateway := NewRPCClientAdapter(rpcClient)

		disperseGateway, err := evm.NewDisperseContractGateway(
			rpcClient,
			chainCfg.DisperseContract,
			abiJSON,
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to create disperse gateway for %s: %v\n", chainCfg.Name, err)
			rpcClient.Close()
			continue
		}

		disperseGatewayAdapter := NewDisperseGatewayAdapter(disperseGateway)

		validatePlan := usecase.NewValidatePlan(chainGateway)
		disperse := usecase.NewDisperse(chainGateway, disperseGatewayAdapter, reportRepo, validatePlan)

		chainInfras[chainCfg.Key] = &chainInfra{
			rpcClient: rpcClient,
			disperse:  disperse,
		}
	}

	if len(chainInfras) == 0 {
		return nil, fmt.Errorf("failed to connect to any chain")
	}

	// Load app configuration
	appCfg, err := config.LoadAppConfig("configs/app.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to load app config: %w", err)
	}

	// Create disperse executor function with per-chain infrastructure
	disperseFunc := createDisperseFunc(chains, chainInfras, appCfg)

	// Derive sender address from private key for balance monitoring
	senderAddr := ""
	privKey := appCfg.GetPrivateKeyForChain(appCfg.App.DefaultChain)
	if privKey == "" && len(chains) > 0 {
		privKey = appCfg.GetPrivateKeyForChain(chains[0].Key)
	}
	if privKey != "" {
		pk, err := crypto.HexToECDSA(strings.TrimPrefix(privKey, "0x"))
		if err == nil {
			senderAddr = crypto.PubkeyToAddress(*pk.Public().(*ecdsa.PublicKey)).Hex()
		}
	}

	// Create TUI app
	app := tui.NewAppModel(chains, appCfg, disperseFunc, senderAddr, reportRepo)

	return app, nil
}

// loadChainConfigs loads all chain configurations from the embedded configs/chains directory
func loadChainConfigs() ([]*config.ChainConfig, error) {
	chainsDir := "configs/chains"

	entries, err := fs.ReadDir(evmdisperse.ChainFS, chainsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded chains directory: %w", err)
	}

	var chains []*config.ChainConfig

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		ext := filepath.Ext(entry.Name())
		if ext != ".yaml" && ext != ".yml" {
			continue
		}

		filePath := path.Join(chainsDir, entry.Name())
		data, err := fs.ReadFile(evmdisperse.ChainFS, filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read embedded chain config %s: %w", filePath, err)
		}

		chainConfig, err := config.ParseChainConfig(data)
		if err != nil {
			return nil, fmt.Errorf("failed to parse chain config %s: %w", filePath, err)
		}
		chains = append(chains, chainConfig)
	}

	if len(chains) == 0 {
		return nil, fmt.Errorf("no valid chain configurations found")
	}

	return chains, nil
}

// createDisperseFunc creates a disperse executor function that uses per-chain infrastructure.
// The correct RPC client and contract gateway are selected based on the chainKey parameter.
// When recipients exceed MaxBatchWalletPerTx, the operation is automatically split into
// multiple transactions to avoid gas limit issues and transaction hash not found errors.
func createDisperseFunc(
	chains []*config.ChainConfig,
	chainInfras map[string]*chainInfra,
	appCfg *config.AppConfig,
) tui.DisperseFunc {
	maxBatch := appCfg.App.MaxBatchWalletPerTx
	if maxBatch < 1 {
		maxBatch = 250
	}

	return func(ctx context.Context, mode string, chainKey string, recipients []string, amount string, token string) (*tui.DisperseFuncResult, error) {
		// Find chain config for metadata (name, chain ID, etc.)
		var chainCfg *config.ChainConfig
		for _, c := range chains {
			if c.Key == chainKey {
				chainCfg = c
				break
			}
		}
		if chainCfg == nil {
			return nil, fmt.Errorf("chain not found: %s", chainKey)
		}

		// Look up per-chain infrastructure
		infra, ok := chainInfras[chainKey]
		if !ok {
			return nil, fmt.Errorf("chain %s (%s) is not connected — check RPC configuration", chainCfg.Name, chainKey)
		}

		// Get private key
		privKey := appCfg.GetPrivateKeyForChain(chainKey)
		if privKey == "" {
			return nil, fmt.Errorf("no private key configured for chain %s — go to Settings (S key)", chainKey)
		}

		// Derive from address from private key
		privKeyHex := strings.TrimPrefix(privKey, "0x")
		privateKey, err := crypto.HexToECDSA(privKeyHex)
		if err != nil {
			return nil, fmt.Errorf("invalid private key: %w", err)
		}
		publicKey := privateKey.Public().(*ecdsa.PublicKey)
		fromAddress := crypto.PubkeyToAddress(*publicKey).Hex()

		// Create domain entities
		tokenMode := entity.TokenModeNative
		if mode == "ERC20" {
			tokenMode = entity.TokenModeERC20
		}

		// Convert amount from human-readable (e.g. "0.0000001") to wei/smallest-unit.
		// Native tokens always have 18 decimals; ERC20 uses the configured TokenDecimals.
		decimals := 18
		if mode == "ERC20" && appCfg.App.TokenDecimals > 0 {
			decimals = appCfg.App.TokenDecimals
		}
		amountWei, err := humanAmountToWei(amount, decimals)
		if err != nil {
			return nil, fmt.Errorf("invalid amount %q: %w", amount, err)
		}

		req := &entity.DisperseRequest{
			Mode:       tokenMode,
			Recipients: recipients,
			Amount:     amountWei,
			Token:      token,
		}

		chain := &entity.Chain{
			Name:             chainCfg.Name,
			ChainID:          chainCfg.ChainID,
			RPCURL:           chainCfg.RPCURL,
			DisperseContract: chainCfg.DisperseContract,
			Network:          chainCfg.Network,
		}

		// Determine if we need batching
		totalRecipients := len(recipients)
		if totalRecipients <= maxBatch {
			// Single batch — execute directly
			report, err := infra.disperse.Execute(ctx, req, chain, fromAddress, privKey)
			if err != nil {
				if report != nil {
					return &tui.DisperseFuncResult{
						TxHash:      report.TxHash,
						BlockNumber: report.BlockNumber,
						GasUsed:     report.GasUsed,
					}, err
				}
				return nil, err
			}

			return &tui.DisperseFuncResult{
				TxHash:      report.TxHash,
				BlockNumber: report.BlockNumber,
				GasUsed:     report.GasUsed,
			}, nil
		}

		// Multi-batch — use BatchDisperse
		batchCount := (totalRecipients + maxBatch - 1) / maxBatch
		fmt.Fprintf(os.Stderr, "[disperse] splitting %d recipients into %d batches (max %d per tx)\n", totalRecipients, batchCount, maxBatch)

		var reports []*entity.ExecutionReport
		reports, err = infra.disperse.BatchExecute(ctx, req, chain, fromAddress, privKey, maxBatch)

		// Collect results from all completed batches
		result := &tui.DisperseFuncResult{
			BatchCount:    batchCount,
			BatchTxHashes: make([]string, 0, len(reports)),
		}
		var totalGas uint64
		for _, r := range reports {
			result.BatchTxHashes = append(result.BatchTxHashes, r.TxHash)
			totalGas += r.GasUsed
			result.BlockNumber = r.BlockNumber // last confirmed block
		}
		result.GasUsed = totalGas
		if len(reports) > 0 {
			result.TxHash = reports[0].TxHash // first tx hash for backward compat
		}

		if err != nil {
			return result, err
		}
		return result, nil
	}
}

// humanAmountToWei converts a human-readable decimal amount (e.g. "0.0000001")
// to its wei/smallest-unit string representation (e.g. "100000000000" for 18 decimals).
func humanAmountToWei(amount string, decimals int) (string, error) {
	if amount == "" {
		return "", fmt.Errorf("amount cannot be empty")
	}
	if decimals < 0 {
		decimals = 0
	}

	// If the amount is already a plain integer (no decimal point),
	// assume it's already in the smallest unit.
	if !strings.Contains(amount, ".") {
		// Validate it's a valid integer
		if _, ok := new(big.Int).SetString(amount, 10); !ok {
			return "", fmt.Errorf("not a valid integer: %s", amount)
		}
		return amount, nil
	}

	// Parse as big.Float for precise decimal arithmetic
	f, _, err := big.ParseFloat(amount, 10, 256, big.ToNearestEven)
	if err != nil {
		return "", fmt.Errorf("invalid decimal amount: %w", err)
	}

	// Multiply by 10^decimals
	multiplier := new(big.Float).SetInt(
		new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil),
	)
	wei := new(big.Float).Mul(f, multiplier)

	// Convert to big.Int (truncates any sub-wei fractional part)
	weiInt, _ := wei.Int(nil)
	if weiInt.Sign() < 0 {
		return "", fmt.Errorf("amount cannot be negative")
	}

	return weiInt.String(), nil
}
