# Founder Wallet Distribution

- **Founder Allocation:** 20% of total supply (200,000,000 AURA) reserved for core team, subject to broader vesting schedule.
- **Community Wallets:** Deduct 500,000 AURA from the founder allocation and distribute to five wallets to bootstrap early community incentives.

| Wallet | Address | Launch Unlock | 6-Month Unlock | Total |
| ------ | ------- | ------------- | --------------- | ----- |
| FW1 | `aura1zjdsnzqgqw22altpjvt5s3uqtx8da8qv5kfdr3` | 20,000 AURA | 80,000 AURA | 100,000 AURA |
| FW2 | `aura1prj0shc0043qaet66pltdqzhms6qfqt550scz3` | 20,000 AURA | 80,000 AURA | 100,000 AURA |
| FW3 | `aura1tv8m8aev62vx0azx84ct5mldf2g6kwkaa4q6y6` | 20,000 AURA | 80,000 AURA | 100,000 AURA |
| FW4 | `aura15a9x5x8uea7m48knm8ttkq0svz3l7dq90u85tx` | 20,000 AURA | 80,000 AURA | 100,000 AURA |
| FW5 | `aura1qrsx3awds7qv6cjc6tm7w5vhtc7xy4vkjmwtn5` | 20,000 AURA | 80,000 AURA | 100,000 AURA |

**Unlock Rules**
- Launch unlock occurs at genesis with no transfer restrictions.
- Six-month unlock occurs automatically via vesting module schedule (no additional governance gate).
- All amounts are deducted from the founder allocation; remaining founder tokens follow the standard vesting plan.

**Next Steps**
- Add these wallets to the genesis file once addresses are finalized. (Draft snippet below.)
- Encode vesting entries: `20k immediate`, `80k vesting with 6-month cliff` per wallet.
- Update tokenomics RFC to reference this distribution.

## Genesis Vesting Snippet (Draft)

```jsonc
{
  "app_state": {
    "vesting": {
      "vesting_accounts": [
        {
          "base_vesting_account": {
            "base_account": { "address": "aura1zjdsnzqgqw22altpjvt5s3uqtx8da8qv5kfdr3" },
            "original_vesting": [{ "denom": "uaura", "amount": "80000000000" }],
            "delegated_free": [],
            "delegated_vesting": []
          },
          "start_time": "0",
          "vesting_periods": [
            { "length": "15768000", "amount": [{ "denom": "uaura", "amount": "80000000000" }] }
          ],
          "immediate_release": [{ "denom": "uaura", "amount": "20000000000" }]
        },
        {
          "base_vesting_account": {
            "base_account": { "address": "aura1prj0shc0043qaet66pltdqzhms6qfqt550scz3" },
            "original_vesting": [{ "denom": "uaura", "amount": "80000000000" }],
            "delegated_free": [],
            "delegated_vesting": []
          },
          "start_time": "0",
          "vesting_periods": [
            { "length": "15768000", "amount": [{ "denom": "uaura", "amount": "80000000000" }] }
          ],
          "immediate_release": [{ "denom": "uaura", "amount": "20000000000" }]
        },
        {
          "base_vesting_account": {
            "base_account": { "address": "aura1tv8m8aev62vx0azx84ct5mldf2g6kwkaa4q6y6" },
            "original_vesting": [{ "denom": "uaura", "amount": "80000000000" }],
            "delegated_free": [],
            "delegated_vesting": []
          },
          "start_time": "0",
          "vesting_periods": [
            { "length": "15768000", "amount": [{ "denom": "uaura", "amount": "80000000000" }] }
          ],
          "immediate_release": [{ "denom": "uaura", "amount": "20000000000" }]
        },
        {
          "base_vesting_account": {
            "base_account": { "address": "aura15a9x5x8uea7m48knm8ttkq0svz3l7dq90u85tx" },
            "original_vesting": [{ "denom": "uaura", "amount": "80000000000" }],
            "delegated_free": [],
            "delegated_vesting": []
          },
          "start_time": "0",
          "vesting_periods": [
            { "length": "15768000", "amount": [{ "denom": "uaura", "amount": "80000000000" }] }
          ],
          "immediate_release": [{ "denom": "uaura", "amount": "20000000000" }]
        },
        {
          "base_vesting_account": {
            "base_account": { "address": "aura1qrsx3awds7qv6cjc6tm7w5vhtc7xy4vkjmwtn5" },
            "original_vesting": [{ "denom": "uaura", "amount": "80000000000" }],
            "delegated_free": [],
            "delegated_vesting": []
          },
          "start_time": "0",
          "vesting_periods": [
            { "length": "15768000", "amount": [{ "denom": "uaura", "amount": "80000000000" }] }
          ],
          "immediate_release": [{ "denom": "uaura", "amount": "20000000000" }]
        }
      ]
    }
  }
}
```

> `uaura` represents micro-AURA, so multiply AURA amounts by `1e6`. `length` uses seconds (≈6 months). `immediate_release` is a convenience field for the planned `tokenomics` module; if SDK tooling requires standard `PeriodicVestingAccount`, convert the immediate portion into the genesis balances/allocations list instead.

> **Hold Off (Jeff):** Do not run the vesting injection or script genesis edits without first checking with Jeff and ops. The wallet plan may change, so pause and get explicit approval before modifying `genesis.json`.
