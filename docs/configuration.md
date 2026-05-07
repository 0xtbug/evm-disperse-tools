# Configuration

All config files live in the `configs/` directory.

## App Config (`configs/app.yaml`)

```yaml
app:
  max_wallet_generate: 0  # 0 = unlimited
  max_batch_wallet_per_tx: 250
  default_amount: "0.0000001"
  token_decimals: 18
  key_mode: global
  default_chain: polygon
  sender_private_key: "0x..."
  chain_keys:
    polygon: "0x..."
    ethereum: "0x..."
```

| Field | Description | Default |
|---|---|---|
| `max_wallet_generate` | Max wallets to generate at once (0 = unlimited) | `0` |
| `max_batch_wallet_per_tx` | Recipients per transaction (auto-batches above this) | 250 |
| `default_amount` | Pre-filled amount in disperse form | `"0.01"` |
| `token_decimals` | Decimal places for ERC20 tokens | 18 |
| `key_mode` | `"global"` (one key) or `"per_chain"` (separate keys) | `"global"` |
| `default_chain` | Pre-selected chain in disperse form | `""` |
| `sender_private_key` | Private key used when `key_mode == global` | `""` |
| `chain_keys` | Map of chain key → private key when `key_mode == per_chain` | `{}` |

This file is git-ignored. Can also be edited in-app via the Settings screen.

## Chain Config (`configs/chains/*.yaml`) — Developer Only

Each chain has its own YAML file:

```yaml
key: polygon
name: Polygon
chain_id: 137
rpc_url: https://polygon-rpc.com
disperse_contract: "0x..."
native_token: POL
```

| Field | Description |
|---|---|
| `key` | Unique identifier (used in config references) |
| `name` | Display name |
| `chain_id` | EVM chain ID |
| `rpc_url` | RPC endpoint URL |
| `disperse_contract` | Disperse contract address |
| `native_token` | Native token symbol (POL, ETH, BNB) |

### Mainnet

| Key | Name | Chain ID | Token |
|---|---|---|---|
| `polygon` | Polygon | 137 | POL |

### Testnet

| Key | Name | Chain ID | Token |
|---|---|---|---|
| `polygon_amoy` | Polygon Amoy | 80002 | POL |
| `ethereum_sepolia` | Ethereum Sepolia | 11155111 | ETH |
| `base_sepolia` | Base Sepolia | 84532 | ETH |

> **Adding a new chain requires developer action.** The disperse contract must be deployed on the target chain first. If you need support for a new chain, [open a chain request](https://github.com/0xtbug/evm-disperse-tools/issues/new?template=chain_request.yml).

## Wallet Lists (`configs/wallets/*.json`)

JSON files containing recipient addresses:

```json
{
  "wallets": [
    { "address": "0x...", "private_key": "..." }
  ]
}
```

Generated via the Wallet Manager in-app, or manually created. This directory is git-ignored.