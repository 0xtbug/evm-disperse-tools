package smoke

import (
	"testing"

	"github.com/0xtbug/evm-disperse-tools/internal/infrastructure/config"
)

func TestChainConfigMatrixHasRequiredFields(t *testing.T) {
	chains := []struct {
		name       string
		path       string
		expectedID uint64
		network    string
	}{
		// Mainnet
		{"polygon", "../../configs/chains/polygon.yaml", 137, "mainnet"},
		{"ethereum", "../../configs/chains/ethereum.yaml", 1, "mainnet"},
		{"bnb", "../../configs/chains/bnb.yaml", 56, "mainnet"},
		{"base", "../../configs/chains/base.yaml", 8453, "mainnet"},
		// Testnet
		{"polygon_amoy", "../../configs/chains/polygon_amoy.yaml", 80002, "testnet"},
		{"ethereum_sepolia", "../../configs/chains/ethereum_sepolia.yaml", 11155111, "testnet"},
		{"bnb_testnet", "../../configs/chains/bnb_testnet.yaml", 97, "testnet"},
		{"base_sepolia", "../../configs/chains/base_sepolia.yaml", 84532, "testnet"},
	}

	for _, chain := range chains {
		t.Run(chain.name, func(t *testing.T) {
			cfg, err := config.LoadChainConfig(chain.path)
			if err != nil {
				t.Fatalf("failed to load %s config: %v", chain.name, err)
			}

			if cfg.Key != chain.name {
				t.Fatalf("expected key %s, got %s", chain.name, cfg.Key)
			}
			if cfg.ChainID != chain.expectedID {
				t.Fatalf("expected chain ID %d, got %d", chain.expectedID, cfg.ChainID)
			}
			if cfg.Name == "" {
				t.Fatalf("chain name cannot be empty")
			}
			if cfg.RPCURL == "" {
				t.Fatalf("RPC URL cannot be empty")
			}
			if cfg.DisperseContract == "" {
				t.Fatalf("disperse contract cannot be empty")
			}
			if cfg.Network != chain.network {
				t.Fatalf("expected network %s, got %s", chain.network, cfg.Network)
			}
			if cfg.IsTestnet() != (chain.network == "testnet") {
				t.Fatalf("IsTestnet() mismatch for %s", chain.name)
			}
		})
	}
}
