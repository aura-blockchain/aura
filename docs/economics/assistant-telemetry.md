# AI Assistant ROI Telemetry

The assistant ROI spreadsheet (`docs/economics/models/assistant-roi-scenarios.csv`) now ties directly into on-chain/off-chain telemetry:

1. **Heartbeats** – The `chain/x/aiassistant` keeper emits `aiassistant_heartbeat_success_total` and `aiassistant_heartbeat_age_seconds` whenever assistants submit heartbeats. Import the `AI Assistant Ops Dashboard` (`grafana/dashboards/aiassistant-monitoring.json`) into Grafana to visualize uptime and SLA breaches.
2. **Voucher lifecycle** – The `aura-voucher` CLI pushes `assistant_voucher_issue_total` / `assistant_voucher_redeem_total` to Prometheus (via Pushgateway) whenever you issue or redeem sponsorship credits. These feed into ROI sheets so finance teams can reconcile how much subsidy remains for each locale or campaign.
3. **Misbehavior tracking** – `aiassistant_msg_misbehavior_total` increments whenever governance/layers report malicious assistants. Tie this into your risk scorecards to adjust revenue projections for bad actors.

## Automating the export to spreadsheets

1. Point the Prometheus datasource in Grafana (or directly via API) to query:

   ```
   sum(increase(aiassistant_heartbeat_success_total[30d])) by (assistant)
   sum(increase(assistant_voucher_issue_total[30d])) by (sponsor)
   sum(increase(assistant_voucher_redeem_total[30d])) by (assistant)
   ```

2. Export the CSV from Grafana or use `curl` against Prometheus' `/api/v1/query` endpoint. Feed those series into the ROI model columns (`IRs_per_day`, `Assistant_Share_Pct`, etc.) to ground-truth the assumptions.
3. Schedule the `aura-voucher` CLI with `cron`/Task Scheduler so sponsors issue vouchers on the cadence assumed in the model, and let the GUI handle redemptions for assistant operators.
