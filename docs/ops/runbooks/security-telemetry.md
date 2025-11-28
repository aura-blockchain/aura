# Security Telemetry Quick Reference

Aura now exposes explicit telemetry for the two most abuse-prone surfaces in the network: gossip deduplication and wallet-level enforcement.  Use this guide alongside the `security-monitoring` Grafana dashboard and the `security_alerts` Prometheus group.

## Metrics

| Metric | Description |
| --- | --- |
| `networksecurity_gossip_cache_hit_total` / `_miss_total` | Counters incremented every time a gossip payload is processed.  Use the ratio to detect duplicate floods. |
| `networksecurity_gossip_cache_size` | Current entry count in the dedup cache.  Evictions are exported via `networksecurity_gossip_cache_evictions`. |
| `walletsecurity_spending_limit_allowed_total` / `_blocked_total` | Number of transactions accepted or rejected by the spending-limit engine. |
| `walletsecurity_dust_filter_allowed_total` / `_blocked_total` | Dust filter results, useful for spotting targeted dust attacks. |
| `dex_swap_effective_price` | Gauge set on each swap showing the instantaneous price (AURA per counter-asset) tagged by `pool`. |
| `dex_market_price_price_aura` / `dex_market_price_sample_size` | Gauges driven by the persisted market price snapshots, tagged by `coin`.  Sample size lets you detect stale pools (no trades). |
| `aiassistant_heartbeat_success_total` | Counter incremented every time an assistant heartbeat succeeds; use derivative windows to detect offline assistants. |
| `aiassistant_heartbeat_age_seconds` | Histogram/gauge representing how long (in seconds) the previous heartbeat took to arrive.  Alerts fire when this creeps above the configured window + grace period. |
| `aiassistant_msg_misbehavior_total` | Counter for slashable events (fraud proofs, false attestations).  Tagging `assistant` lets responders spot repeat offenders. |
| `assistant_voucher_issue_total` / `assistant_voucher_redeem_total` | Pushgateway metrics emitted by the `aura-voucher` CLI so finance dashboards know how many sponsorship credits were minted vs consumed. |

All counters expose Prometheus-friendly names, so a key path of `walletsecurity -> spending_limit -> blocked` becomes `walletsecurity_spending_limit_blocked_total`.

## Grafana panels

1. **Gossip Cache Hit Ratio** – Values above ~0.85 for more than a few minutes usually indicate duplicate rebroadcast or DOS attempts.  Drill down into node logs and peer stats if the alert fires.
2. **Gossip Cache Size & Evictions** – Tracks how full the dedup window is and whether eviction churn spikes (which could highlight memory pressure).
3. **Wallet Spending Limit Decisions** – Stacked view showing accepted vs blocked operations per hour.  Sudden spikes in the blocked series typically correlate with compromised wallets.
4. **Dust Filter Blocks (5m)** – Rolling count of dust transactions rejected over five minutes, exposing low-and-slow dust campaigns.

## Alerts

Prometheus rules in `prometheus/rules/monitoring-alerts.yml` now include:

- `GossipDuplicateFlood` – Fires when >90% of gossip traffic is duplicated for five straight minutes.  Investigate relayers or validators that may be echoing stale packets.
- `WalletSpendingBlocksSpike` – Triggers when more than 10 operations are rejected by spending limits in 5 minutes.  This can indicate scripted draining attempts.
- `DustAttackDetected` – Fires after 20+ dust transactions are blocked within 5 minutes.  Blacklist the originating addresses and raise a comms bulletin.
- `DexMarketPriceStale` – Fires if `dex_market_price_sample_size` fails to change for 15 minutes, indicating a pool with no recent trades (possible stuck liquidity or halted relayer).
- `DexSpotPriceError` – Alerts when the spot-price gRPC endpoint returns validation errors more than 50 times in 5 minutes; correlate with the “DEX Query Validation Logs” section to identify offending clients.

Responders should acknowledge the Prometheus alert, capture relevant logs (`networksecurity` module for gossip, `walletsecurity` for user flows), and cite this runbook in the incident ticket for traceability.

## DEX Query Validation Logs

The DEX gRPC/CLI surfaces now log every rejected query (missing order IDs, malformed addresses, unsupported coins, etc.) as `dex query validation failed` entries. These are routed through Tendermint logging, tagged with the query name (`query` field) and reason. A spike of such logs often signals reconnaissance or malformed bot traffic; tie them to client IPs via your node’s upstream reverse proxy when investigating.

## Routing alerts to Slack / PagerDuty

Alertmanager is pre-configured via `docker/monitoring/alertmanager/config.yml` to send:

- `critical` alerts to PagerDuty using the `PAGERDUTY_ROUTING_KEY` environment variable.
- `high`/`warning` alerts to Slack using `SLACK_WEBHOOK_URL` and `SLACK_CHANNEL` (defaults to `#aura-alerts` if unset).

Before starting `docker-compose.monitoring.yml`, export the desired secrets, e.g.:

```bash
export SLACK_WEBHOOK_URL="https://hooks.slack.com/services/TXXXX/BXXXX/XXXXXXXX"
export SLACK_CHANNEL="#soc-alerts"
export PAGERDUTY_ROUTING_KEY="abc123..."
docker compose -f docker-compose.monitoring.yml up -d alertmanager prometheus grafana
```

If no environment variable is set, Alertmanager simply drops messages for that receiver, so make sure production deployments define them explicitly.
