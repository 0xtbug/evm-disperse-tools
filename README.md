# evm-disperse-tools

**Free EVM token disperse tool — pay only network gas fees, zero platform fees.**

> Alternative to [disperse.app](https://disperse.app) and [cryptosend.io](https://cryptosend.io)

A terminal-based (TUI) tool for bulk sending native tokens and ERC20 tokens across multiple EVM chains. No web interface, no middleman, no platform fees — just network gas.

| | disperse.app / cryptosend.io | evm-disperse-tools |
|---|---|---|
| Platform Fee | Yes | **No** |
| Self-hosted | No | **Yes** |
| Private Key Control | Browser extension | **Your machine** |
| Open Source | No | **Yes** |

## Features

- **Multi-Chain** — 4 mainnets + 4 testnets ([request a chain](https://github.com/0xtbug/evm-disperse-tools/issues/new?template=chain_request.yml))
- **Dual Mode** — Native tokens (`distributeEther`) or ERC20 tokens (`distribute`)
- **Smart Batching** — Auto-splits large recipient lists (configurable batch size)
- **Wallet Generator** — Bulk wallet generation from the TUI
- **Fee Calculator** — Gas cost estimation with live RPC prices
- **Execution Reports** — JSON audit trails in `data/reports/`
- **RPC Monitor** — Live block height, balance, gas price, latency
- **Tokyo-Night TUI** — Built with BubbleTea + Lipgloss

## Supported Chains

| Network | Chain | ID | Token |
|---|---|---|---|
| Mainnet | Polygon | 137 | POL |
| Mainnet | Ethereum | 1 | ETH |
| Mainnet | BNB Chain | 56 | BNB |
| Mainnet | Base | 8453 | ETH |
| Testnet | Polygon Amoy | 80002 | POL |
| Testnet | Ethereum Sepolia | 11155111 | ETH |
| Testnet | BNB Chain Testnet | 97 | BNB |
| Testnet | Base Sepolia | 84532 | ETH |

## Quick Start

```bash
git clone https://github.com/0xtbug/evm-disperse-tools.git
cd evm-disperse-tools
go mod download
go build -o evm-disperse-tools ./cmd/evm-disperse-tools
./evm-disperse-tools
```

Requires Go 1.24.0+.

### Development

```bash
go install github.com/air-verse/air@latest
air
```

## Screens

| Screen | Description |
|---|---|
| Disperse Native | Send native tokens to multiple recipients |
| Disperse ERC20 | Send ERC20 tokens to multiple recipients |
| View Reports | Browse execution history |
| Fee Calculator | Estimate gas costs for bulk transfers |
| Settings (`S`) | Configure keys, chains, batch size |
| Wallet Manager (`W`) | Generate and save wallet lists |

## Docs

- [Keyboard Shortcuts](docs/keyboard-shortcuts.md)
- [Configuration](docs/configuration.md)

## Request & Issues

- [Request a new chain](https://github.com/0xtbug/evm-disperse-tools/issues/new?template=chain_request.yml) — requires contract deployment
- [Request a feature](https://github.com/0xtbug/evm-disperse-tools/issues/new?template=feature_request.yml)
- [Report a bug](https://github.com/0xtbug/evm-disperse-tools/issues/new?template=bug_report.yml)

## Testing

```bash
go test ./...                  # all tests
go test ./tests/smoke -v       # chain config validation
go test ./tests/security -v    # hardcoded key scanner
go vet ./...                   # static analysis
```

## License

MIT

---

**Made with ❤️ by 0xtbug**
