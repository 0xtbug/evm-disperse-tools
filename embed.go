package evmdisperse

import "embed"

//go:embed configs/chains
var ChainFS embed.FS

//go:embed internal/infrastructure/evm/abi/disperse.json
var DisperseABI []byte
