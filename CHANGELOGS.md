# Changelog

## v1.0.3

### Fixes

- **Update Checker** — Update notification now works for all builds including local/dev builds (previously skipped when version was `dev`)

## v1.0.2

### Fixes

- **Batch Disperse** — Fixed wallet list not refreshing addresses when saving to an existing list name (stale addresses in disperse form)
- **Batch Disperse** — Fixed gas estimate in balance validation to scale with recipient count (was hardcoded at 200k gas units, now uses 50k base + 25k per recipient)
- **Error Messages** — Removed triple-wrapped `"failed to send transaction"` error messages; actual error reason (e.g. `insufficient funds`) is now visible
- **Activity Log** — Fixed `truncateString` to show the end of long messages instead of the beginning, so the actual error is always readable
- **Activity Log** — Removed `fmt.Fprintf(os.Stderr)` debug print that corrupted TUI layout and caused `[disperse]` text to bleed into the footer

## v1.0.1

### Improvements

- **Wallet Manager** — New export features:
  - `P` Export private keys only (one per line to `<name>_privkeys.txt`)
  - `A` Export addresses only (one per line to `<name>_addresses.txt`)
- **Activity Log** — All wallet operations now logged in the Activity pane
- **Footer Shortcuts** — Updated to show all available wallet manager shortcuts

### Fixes

- Removed duplicate result display from wallet manager (now only in Activity log)
- Updated docs: keyboard shortcuts, configuration, supported chains

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
