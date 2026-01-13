# Validator Setup Guide

This guide assumes you already run a synced full node.

## 1) Create a key

```bash
./build/aurad keys add validator-key
```

## 2) Fund the account

Use the faucet or a funded account to send tokens to the validator address:

```bash
./build/aurad keys show validator-key -a
```

## 3) Create the validator

```bash
./build/aurad tx staking create-validator \
  --amount 1000000uaura \
  --pubkey $(./build/aurad tendermint show-validator) \
  --moniker "<moniker>" \
  --chain-id aura-mvp-1 \
  --commission-rate 0.10 \
  --commission-max-rate 0.20 \
  --commission-max-change-rate 0.01 \
  --min-self-delegation 1 \
  --from validator-key \
  --gas auto \
  --gas-adjustment 1.2 \
  --gas-prices 0.025uaura
```

## 4) Verify

```bash
./build/aurad query staking validator <validator-operator-address>
```
