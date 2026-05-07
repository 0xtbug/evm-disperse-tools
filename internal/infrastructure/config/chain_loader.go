package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ChainConfig represents the configuration for a blockchain chain
type ChainConfig struct {
	Key              string `yaml:"key"`
	Name             string `yaml:"name"`
	ChainID          uint64 `yaml:"chain_id"`
	RPCURL           string `yaml:"rpc_url"`
	DisperseContract string `yaml:"disperse_contract"`
	NativeToken      string `yaml:"native_token"` // native token symbol e.g. ETH, BNB, POL
	Network          string `yaml:"network"`      // "mainnet" or "testnet"
}

// GetNativeToken returns the native token symbol, falling back to a sensible default
// based on the chain key if not explicitly configured.
func (c *ChainConfig) GetNativeToken() string {
	if c.NativeToken != "" {
		return c.NativeToken
	}
	// Fallback based on well-known chain keys (including testnet variants)
	switch c.Key {
	case "ethereum", "base", "base_sepolia":
		return "ETH"
	case "bnb", "bnb_testnet":
		return "BNB"
	case "polygon", "polygon_amoy":
		return "POL"
	case "ethereum_sepolia":
		return "ETH"
	default:
		return "ETH"
	}
}

// IsTestnet returns true if the chain is a testnet.
func (c *ChainConfig) IsTestnet() bool {
	return c.Network == "testnet"
}

// LoadChainConfig loads a chain configuration from a YAML file
func LoadChainConfig(path string) (*ChainConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg ChainConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &cfg, nil
}
