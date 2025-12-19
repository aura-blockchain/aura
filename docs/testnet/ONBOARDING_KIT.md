# Aura Testnet Onboarding Kit

Everything needed to join `aura-testnet-1`, test wallets, and verify endpoints.

## Essential Links
- **Chain registry JSON:** `docs/chain-registry/aura.json`
- **RPC:** `https://rpc.aura-testnet.com` (WS: `wss://rpc.aura-testnet.com/websocket`)
- **REST:** `https://api.aura-testnet.com`
- **gRPC:** `grpc.aura-testnet.com:443`
- **Local hardened proxy:** `http://localhost:8080/rpc` (from `docker-compose.proxy.yml`)
- **Explorer (local):** `http://localhost:8088`
- **Faucet:** `https://faucet.aura-testnet.com` (hosted) or `http://localhost:8081` (local compose)
- **Android installable:** `wallet/mobile/dist/aura-wallet-testnet-debug.apk` (web wallet packaged via Capacitor)
- **If DNS fails:** point RPC/REST to the local proxy (`http://localhost:8080/rpc` / `http://localhost:8080/api`) or a direct IP-based endpoint; update WalletConnect/Keplr configs accordingly.

## Join the Network (operators)
```bash
# Full or light node with state sync + peers + gas prices set
MODE=full ./scripts/join-aura-testnet.sh      # or MODE=light for pruned nodes
aurad start --home ~/.aura

# Light client trusted header helper (optional)
PRIMARY_RPC=http://localhost:8080/rpc ./scripts/generate-trust-header.sh
# Smoke test an already running light client
LC_RPC=http://localhost:8888 ./scripts/test-light-client.sh
# If you need a local RPC to generate trust headers:
#   docker compose -f docker-compose.observer.yml up -d
#   docker compose -f docker-compose.proxy.yml up -d
# (or start the full testnet stack)
# Local testnet chain-id: use `CHAIN_ID=aura-local-4` with host-published ports (`http://localhost:27657`, `http://localhost:27757`) when running the light client on the same machine as the validators.
```
- State sync & snapshot details: `docs/testnet/STATE_SYNC_SNAPSHOT.md`
- Light client validation: `docs/testnet/LIGHT_CLIENT_GUIDE.md`

## Wallet / dApp Configuration
Keplr/Leap chain info:
```javascript
const chainInfo = {
  chainId: 'aura-testnet-1',
  chainName: 'Aura Testnet',
  rpc: 'https://rpc.aura-testnet.com',
  rest: 'https://api.aura-testnet.com',
  bip44: { coinType: 118 },
  bech32Config: {
    bech32PrefixAccAddr: 'aura',
    bech32PrefixAccPub: 'aurapub',
    bech32PrefixValAddr: 'auravaloper',
    bech32PrefixValPub: 'auravaloperpub',
    bech32PrefixConsAddr: 'auravalcons',
    bech32PrefixConsPub: 'auravalconspub'
  },
  stakeCurrency: { coinDenom: 'AURA', coinMinimalDenom: 'uaura', coinDecimals: 6 },
  currencies: [{ coinDenom: 'AURA', coinMinimalDenom: 'uaura', coinDecimals: 6 }],
  feeCurrencies: [{ coinDenom: 'AURA', coinMinimalDenom: 'uaura', coinDecimals: 6, gasPriceStep: { low: 0.015, average: 0.025, high: 0.04 } }],
  features: ['stargate', 'ibc-transfer', 'cosmwasm']
};
```
WalletConnect (v2) URI example (replace `PROJECT_ID` with your relay project):
```
wc:?uri=wc%3Csession-id%3E%402%3Frelay-protocol%3Dirn%26symKey%3Cgenerated-key%3E&symKey=<generated-key>&relay-protocol=irn&relay-data={"projectId":"PROJECT_ID","relayUrl":"wss://relay.walletconnect.com"}
```
Recommended WalletConnect metadata for desktop/extension:
```json
{
  "name": "Aura Wallet",
  "description": "Aura testnet wallet",
  "url": "https://aura-testnet.com",
  "icons": ["https://aura-testnet.com/icon.png"],
  "chains": ["cosmos:aura-testnet-1"]
}
```

## First-Run Checklist (wallet users)
- Import or create mnemonic; verify address prefix `aura`
- Confirm chain ID shows `aura-testnet-1`
- Set gas price to `0.025uaura` (matches chain registry)
- Ping endpoints:
  - `curl https://rpc.aura-testnet.com/status | jq .result.sync_info`
  - `curl https://api.aura-testnet.com/cosmos/bank/v1beta1/supply | jq .supply[0]`
- Request funds from the faucet (1–5 AURA per request) and send a self-transfer to verify signing
- Enable hardware confirmation if using Ledger/Trezor/Keystone flows in the extension/desktop wallet

## Device & Hardware Matrix (test focus)
- Android: Pixel 6 (Android 13), Galaxy S22 (Android 12)
- iOS: iPhone 13 (iOS 16), iPhone 14 (iOS 17)
- Hardware wallets: Ledger Nano S+/Stax (WebHID), Trezor T (Bridge/WebUSB), Keystone QR
- Network paths: WalletConnect v2 (desktop QR + mobile deep link), direct Keplr/Leap provider
Record results in `WALLET_TESTING_REPORT.md` after each run.

## Support Scripts & Files
- Light client compose: `docker-compose.light-client.yml`
- State sync helper: `scripts/join-aura-testnet.sh`
- Chain registry source of truth: `docs/chain-registry/aura.json`
- Public endpoints validation: `docs/TESTNET_PUBLIC_ENDPOINTS.md`, `QUICK_START_RPC.md`
