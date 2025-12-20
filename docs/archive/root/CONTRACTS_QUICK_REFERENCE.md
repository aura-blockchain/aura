# Smart Contracts Quick Reference

## Deployed Contracts (aura-local-4)

### VC Issuer
```
Code ID:  6
Address:  aura153r9tg33had5c5s54sqzn879xww2q2egektyqnpj6nwxt8wls70qgxvq7r
Admin:    aura16lhgey7k6fd563ysczdvs3pq86rgyu08gl8wac
```

### Schema
```
Code ID:  7
Address:  aura1f6jlx7d9y408tlzue7r2qcf79plp549n30yzqjajjud8vm7m4vdsmqktx7
Admin:    aura16lhgey7k6fd563ysczdvs3pq86rgyu08gl8wac
```

### Binding Tester
```
Code ID:  8
Address:  aura124tapgv8wsn5t3rv2cvywh4ckkmj6mc6fkya005qjmshnzewwm9q8k7mgq
Admin:    aura16lhgey7k6fd563ysczdvs3pq86rgyu08gl8wac
```

---

## Common Commands

### Query Contract Info
```bash
./aurad query aura_wasm_security contract <ADDRESS> \
  --node http://localhost:27657 \
  --chain-id aura-local-4
```

### Query Contract State
```bash
./aurad query aura_wasm_security contract-state-all <ADDRESS> \
  --node http://localhost:27657 \
  --chain-id aura-local-4
```

### Execute Contract
```bash
./aurad tx aura_wasm_security execute <ADDRESS> '<JSON_MSG>' \
  --from validator-1 \
  --home ./testnet-data/validator-1 \
  --chain-id aura-local-4 \
  --node http://localhost:27657 \
  --keyring-backend test \
  --yes \
  --broadcast-mode sync \
  --gas 1000000 \
  --gas-prices 0.025uaura
```

### Query Code Info
```bash
./aurad query aura_wasm_security code <CODE_ID> \
  --node http://localhost:27657 \
  --chain-id aura-local-4
```

---

## Environment Variables

```bash
export AURA_BINARY="./aurad"
export AURA_CHAIN_ID="aura-local-4"
export AURA_NODE="http://localhost:27657"
export AURA_HOME="./testnet-data/validator-1"
export AURA_KEYRING="test"
export AURA_GAS_PRICES="0.025uaura"
```

---

## Redeploy All Contracts

```bash
./scripts/deploy-contracts-simple.sh
```

---

## RPC Endpoints

- **Validator 1:** http://localhost:27657
- **Validator 2:** http://localhost:27757
- **Validator 3:** http://localhost:27857
- **Validator 4:** http://localhost:27957

---

## Monitoring

- **Prometheus:** http://localhost:9094
- **Grafana:** http://localhost:3002

---

## Deployment Log

JSON record of all deployments: `contract-deployments.json`

---

## Full Documentation

See `CONTRACT_DEPLOYMENT_REPORT.md` for complete deployment details.
