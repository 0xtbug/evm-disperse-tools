package entity

// Chain represents a blockchain network configuration
type Chain struct {
	Name             string
	ChainID          uint64
	RPCURL           string
	DisperseContract string
	Network          string // "mainnet" or "testnet"
}
