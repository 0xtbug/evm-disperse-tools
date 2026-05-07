package entity

import "fmt"

// TokenMode represents the mode of token dispersal
type TokenMode string

const (
	TokenModeNative TokenMode = "native"
	TokenModeERC20  TokenMode = "erc20"
)

// DisperseRequest represents a request to disperse tokens
type DisperseRequest struct {
	Mode       TokenMode
	Recipients []string
	Amount     string
	Token      string
}

// Validate checks if the DisperseRequest is valid
func (d *DisperseRequest) Validate() error {
	if d.Mode != TokenModeNative && d.Mode != TokenModeERC20 {
		return fmt.Errorf("invalid mode: must be 'native' or 'erc20'")
	}

	if len(d.Recipients) == 0 {
		return fmt.Errorf("recipients list cannot be empty")
	}

	return nil
}
