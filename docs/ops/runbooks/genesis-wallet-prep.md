# Genesis Wallet Prep Runbook

Purpose: ensure founder + community wallets are deterministic, documented, and encoded into the Cosmos SDK genesis file before devnet launch.

## Prerequisites
- Cosmos SDK `aurad` binary with `gentx` + `add-genesis-account` commands.
- Access to the encrypted seed phrases for FW1–FW5 custodians.
- `docs/economics/founder-wallets.md` kept current with labels, addresses, and vesting amounts.

## Steps
1. **Verify Addresses**
   - Ask each custodian to confirm the Bech32 address matches the table.
   - Store confirmations (PIV/Slack thread link) in the ops vault.
2. **Create Local Key Records**
   - `aurad keys add fw1 --recover` (repeat for fw2–fw5).
   - Export public keys for governance reference.
3. **Seed Genesis Balances**
   - `aurad add-genesis-account fw1 20000000000uaura` to load the immediate release (20k AURA) for each wallet.
4. **Attach Vesting Accounts**
   - Use `aurad add-genesis-account --vesting-amount 80000000000uaura --vesting-end-time 15768000 fw1` once CLI supports periodic vesting, or inject JSON per `docs/economics/founder-wallets.md`.
5. **Sanity Check Totals**
   - `jq '.app_state.bank.balances | map(.coins[0].amount|tonumber) | add' genesis.json` should equal founder allocation (500k AURA) in micro-units plus other allocations.
6. **Sign-Off**
   - Capture a diff of `genesis.json` and upload to the release checklist PR.

## Outputs
- Updated `genesis.json` with FW1–FW5 accounts.
- Confirmed custody acknowledgments from each wallet owner.
- Runbook checklist completed in the ops tracker.
