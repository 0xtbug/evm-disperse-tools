# Changelog

## v1.0.0

Initial release of **evm-disperse-tools** — a free, terminal-based (TUI) tool for bulk sending native tokens and ERC20 tokens across multiple EVM chains.

### Features

- **Multi-Chain Support** — Polygon, Ethereum Sepolia, Polygon Amoy, Base Sepolia
- **Dual Mode** — Native tokens (`distributeEther`) and ERC20 tokens (`distribute`)
- **Smart Batching** — Auto-splits large recipient lists to avoid gas limit issues (configurable batch size)
- **Wallet Generator** — Bulk wallet generation with no limit (up to 9,999,999 wallets)
- **Fee Calculator** — Gas cost estimation with live RPC prices
- **Execution Reports** — JSON audit trails in `data/reports/`
- **RPC Monitor** — Live block height, balance, gas price, latency
- **Tokyo-Night TUI** — Built with BubbleTea + Lipgloss
- **Embedded Configs** — Chain configs and ABI bundled into the binary (no external files needed)
- **Settings** — Global or per-chain private key configuration

### Supported Chains

| Network  | Chain            | Chain ID | Token |
|----------|------------------|----------|-------|
| Mainnet  | Polygon          | 137      | POL   |
| Testnet  | Polygon Amoy     | 80002    | POL   |
| Testnet  | Ethereum Sepolia | 11155111 | ETH   |
| Testnet  | Base Sepolia     | 84532    | ETH   |

### CI/CD

- **CI Workflow** — Tests and builds on tag push (`v*`), runs across Linux, macOS, Windows
- **Release Workflow** — Sequential: waits for CI to pass, then builds binaries for all platforms and creates GitHub Release
- **Manual Release** — Supports `workflow_dispatch` for manual releases from GitHub Actions

### Platforms

Pre-built binaries available for:
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)
