# evm-disperse-tools
**Free EVM token disperse tool — pay only network gas fees, zero platform fees.**

> Alternative to [disperse.app](https://disperse.app) and [cryptosend.io](https://cryptosend.io)

A terminal-based (TUI) tool for bulk sending native tokens and ERC20 tokens across multiple EVM chains. No web interface, no middleman, no platform fees — just network gas.

<img width="2014" height="1114" alt="WindowsTerminal_B421QjfGHC" src="https://github.com/user-attachments/assets/add8fc31-2c19-425d-83db-7d889ae9601f" /><br />

## Difference
| | disperse.app / cryptosend.io | evm-disperse-tools |
|---|---|---|
| Platform Fee | Yes | **No** |
| Self-hosted | No | **Yes** |
| Private Key Control | Browser extension | **Your machine** |
| Open Source | No | **Yes** |

## Features

- **Multi-Chain** — 1 mainnet + 3 testnets ([request a chain](https://github.com/0xtbug/evm-disperse-tools/issues/new?template=chain_request.yml))
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
| Testnet | Polygon Amoy | 80002 | POL |
| Testnet | Ethereum Sepolia | 11155111 | ETH |
| Testnet | Base Sepolia | 84532 | ETH |

## Quick Start

### Download (Recommended)

Download the latest pre-built binary from [GitHub Releases](https://github.com/0xtbug/evm-disperse-tools/releases/latest):

| Platform | File |
|---|---|
| Linux (x64) | `evm-disperse-tools-linux-amd64` |
| Linux (ARM) | `evm-disperse-tools-linux-arm64` |
| macOS (Intel) | `evm-disperse-tools-darwin-amd64` |
| macOS (Apple Silicon) | `evm-disperse-tools-darwin-arm64` |
| Windows | `evm-disperse-tools-windows-amd64.exe` |

Verify checksum: `sha256sum -c checksums.txt`

### Build from Source

```bash
git clone https://github.com/0xtbug/evm-disperse-tools.git
cd evm-disperse-tools
go mod download
go build -o evm-disperse-tools ./cmd/evm-disperse-tools
./evm-disperse-tools
```

Requires Go 1.24.0+.

> **Note:** Building from source will show version as `dev`. To set a specific version, use:
> ```bash
> go build -ldflags="-X github.com/0xtbug/evm-disperse-tools/internal/version.Version=$(git describe --tags)" -o evm-disperse-tools ./cmd/evm-disperse-tools
> ```

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
| Wallet Manager (`W`) | Generate wallets, save JSON, export private keys or addresses |

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
go vet ./...                   # static analysis
```

## License

MIT
