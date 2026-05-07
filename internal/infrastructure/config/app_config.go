package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const appConfigPath = "configs/app.yaml"

// AppConfig holds application-level settings loaded from configs/app.yaml
type AppConfig struct {
	App struct {
		MaxWalletGenerate   int               `yaml:"max_wallet_generate"`
		MaxBatchWalletPerTx int               `yaml:"max_batch_wallet_per_tx"` // max wallets per single transaction (default 250)
		DefaultAmount       string            `yaml:"default_amount"`          // pre-fill amount in disperse form
		TokenDecimals       int               `yaml:"token_decimals"`          // decimals for ERC20 token (default 18)
		KeyMode             string            `yaml:"key_mode"`                // "global" or "per_chain"
		DefaultChain        string            `yaml:"default_chain"`           // pre-selected chain in disperse form
		SenderPrivateKey    string            `yaml:"sender_private_key"`      // used when key_mode == "global"
		ChainKeys           map[string]string `yaml:"chain_keys"`              // used when key_mode == "per_chain"
	} `yaml:"app"`
}

// DefaultAppConfig returns an AppConfig with sane defaults
func DefaultAppConfig() *AppConfig {
	cfg := &AppConfig{}
	cfg.App.MaxWalletGenerate = 1000
	cfg.App.MaxBatchWalletPerTx = 250
	cfg.App.DefaultAmount = "0.01"
	cfg.App.TokenDecimals = 18
	cfg.App.DefaultChain = ""
	cfg.App.KeyMode = "global"
	cfg.App.SenderPrivateKey = ""
	cfg.App.ChainKeys = map[string]string{}
	return cfg
}

// LoadAppConfig loads application configuration from the given YAML path.
// If the file does not exist, returns default config.
func LoadAppConfig(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultAppConfig(), nil
		}
		return nil, fmt.Errorf("failed to read app config: %w", err)
	}

	cfg := DefaultAppConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse app config: %w", err)
	}

	// Validate
	if cfg.App.MaxWalletGenerate < 1 {
		cfg.App.MaxWalletGenerate = 1000
	}
	if cfg.App.MaxBatchWalletPerTx < 1 {
		cfg.App.MaxBatchWalletPerTx = 250
	}
	if cfg.App.TokenDecimals < 0 {
		cfg.App.TokenDecimals = 18
	}
	if cfg.App.TokenDecimals == 0 {
		cfg.App.TokenDecimals = 18
	}
	if cfg.App.KeyMode != "global" && cfg.App.KeyMode != "per_chain" {
		cfg.App.KeyMode = "global"
	}
	if cfg.App.ChainKeys == nil {
		cfg.App.ChainKeys = map[string]string{}
	}

	return cfg, nil
}

// Save writes the current config to disk
func (cfg *AppConfig) Save() error {
	_ = os.MkdirAll("configs", 0755)
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal app config: %w", err)
	}
	header := "# EVM Disperse TUI Application Configuration\n"
	if err := os.WriteFile(appConfigPath, []byte(header+string(data)), 0644); err != nil {
		return fmt.Errorf("failed to write app config: %w", err)
	}
	return nil
}

// GetPrivateKeyForChain returns the private key to use for the given chain key.
// Returns empty string if not configured.
func (cfg *AppConfig) GetPrivateKeyForChain(chainKey string) string {
	if cfg.App.KeyMode == "per_chain" {
		return cfg.App.ChainKeys[chainKey]
	}
	return cfg.App.SenderPrivateKey
}
