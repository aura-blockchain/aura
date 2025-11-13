# Aequitas / AURA Identity Blockchain

## Overview
Aequitas is a zero-PII Layer-1 built with the Cosmos SDK to serve as a decentralized identity trust anchor for W3C verifiable credentials, AI assistants, and governance backed by proof-of-identity (PoI) rewards. This repository captures the technical narrative (see `Auquitas AURAcoin Blockchain.md`), RFCs, economics models, and operations playbook required to launch the chain, assistant network, and companion apps.

## Repository layout
- `Auquitas AURAcoin Blockchain.md`: the master technical specification for the protocol.
- `PROJECT_STATUS.md` / `progress.md`: living summaries of the project snapshot and short-term focus areas.
- `docs/`: architecture diagrams, RFCs, economics models, and runbooks referenced by implementations.
- `data/`: raw verifier fee event dumps consumed by the economics tooling.
- `tools/`: helper scripts such as `aggregate_verifier_fees.py` and `build_economics_notebook.py`.
- `chain/`, `ai-assistant/`, `wallet/`, `verifier-portal/`, `zkp/`, `infra/`: planned module areas for the chain, assistants, light wallet, verifier UX, zk-governance, and infrastructure/tooling. Each directory begins with a README that ties the modules back to the RFCs and next steps.
- `wallet/php/`: helper classes (`BalanceCalculator`, `RewardCalculator`) plus PHPUnit/PHPCS/PHPStan-ready coverage so the PHP tooling has concrete targets; extend this space for future wallet/assistant helpers.

## Getting started
1. Read `PROJECT_STATUS.md` and `progress.md` to understand current milestones.
2. Review `docs/rfcs/` (especially RFCs 0002, 0003, 0005, 0006, and 0007) to internalize the target flows, modules, and parameters.
3. Inspect `docs/economics/models/` and run `python tools/aggregate_verifier_fees.py` to regenerate the fee summaries before refreshing `docs/economics/models/economics-scenarios.ipynb` via `python tools/build_economics_notebook.py`.
4. Refer to the per-module README under `chain/`, `ai-assistant/`, etc., to see the immediate implementation targets.

## Tooling reminders
- `python tools/aggregate_verifier_fees.py --input-dir data/verifier-fee-events --output docs/economics/models/verifier-fee-data.csv` recalculates monthly totals with the tokenomics burn/distribution shares.
- `python tools/build_economics_notebook.py` rewrites the scenario notebook from the template cells tracked in the script. Run it after updating any of the CSV inputs.
- `python tools/sync_verifier_fee_events.py --source-url $VERIFIER_TELEMETRY_ZIP --run-aggregator` downloads the latest verifier fee archive to `data/verifier-fee-events`, extracts the CSVs, and triggers the aggregate refresh automatically.
- `php tools/bin/composer.phar install` installs PHPUnit, PHPStan, PHPCS, and WPCS so you can run `php tools/bin/composer.phar run phpstan|phpcs|phpunit` without a global Composer.
- The tracked PHPCS ruleset currently reuses PSR-12/Squiz helpers (see `phpcs.xml.dist`) because a few WordPress-specific sniffs still require PHP compatibility fixes, but the WPCS bundle is already available in case you want to add WP-focused checks later.
- Run `pre-commit install` after `pip install pre-commit` to have `.pre-commit-config.yaml` validate Go/PHP tooling locally, and rely on `.github/workflows/ci.yml` for the cross-language CI checks.
- `tests/`: contains PHPUnit suites such as `tests/Wallet/BalanceCalculatorTest.php` so the PHP toolchain has a meaningful workload. Add new PHP tests here for future helpers.
- **Local environment tools**
- `C:\Users\decri\GitClones\go-env\bin`: `gosec v2.dev`, `golangci-lint v1.64.8`, `goimports`, `govulncheck v1.1.4`.
- `C:\Users\decri\GitClones\python-env\Scripts`: `black 25.11.0`, `pylint 4.0.2`, `mypy 1.18.2`, `locust 2.42.2`, `pre-commit`.
- Node tooling is installed per-project; the PAW workspace already provides `ESLint v8.57.1`, `Prettier 3.6.2`, `commitlint v18.6.1`, and ~372 npm packages, which can be reused by other projects that install the same dependencies.
- Husky is configured here and manages the Git hooks for this repo instead of `pre-commit`.
- Husky is installed via `npm install` for this repo, and the tracked `package.json`/`.husky/pre-commit` scripts run `composer run test` followed by `composer run phpcs` before each commit (`npm run prepare` wires the hooks once husky's dependencies are available).
- `scripts/generate_identitychange_proto.sh` regenerates the identitychange proto bindings (via `buf generate` when available or `protoc --go_out=paths=source_relative`) and requires `buf`/`protoc-gen-go` plus `proto/cosmos/base/query/v1beta1/pagination.proto`, so run it before replacing the placeholder `.proto` output.

## Next milestones
- Lock down the RFC wording (`docs/rfcs/0002`, `0003`, `0005`–`0007`) and capture reviewer feedback ahead of coding.
- Begin implementing the chain identity modules (inclusion routines, manager) and the monitoring components that feed them from the assistant/verification stack.
- Wire the CI placeholder jobs to real build, lint, and test commands for Go, AI tooling, and the wallet stack.
- Coordinate genesis preparation and custodian approvals while refreshing the economics dashboard with Monte Carlo simulations tied to fee telemetry.

## References
- RFC overview: `docs/rfcs/README` (add if needed).
- Operations runbooks: `docs/ops/`.
- Keep the new `progress.md` updated as work ticks across modules.
