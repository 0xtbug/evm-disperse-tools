package entity

import "time"

// ExecutionReport represents a complete execution report of a disperse operation
type ExecutionReport struct {
	TxHash      string
	Status      string
	GasUsed     uint64
	BlockNumber uint64
	Recipients  int
	TotalAmount string
	Token       string
	ChainName   string
	Timestamp   time.Time
}
