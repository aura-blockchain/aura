# Infrastructure & Tooling

## Purpose
Stand up CI, deployment, monitoring, and operational guardrails that keep the chain, assistants, wallet, and portals reliable and auditable.

## Anchors
- `.github/workflows/ci.yml` is the current scaffold for language-agnostic builds/tests.
- `docs/ops/` contains runbooks, AML/compliance notes, and incident templates for the operations team.
- `tools/aggregate_verifier_fees.py` and `tools/build_economics_notebook.py` demonstrate the automation approach for economics modeling.

## Next steps
1. Populate the GitHub Actions jobs with actual build, lint, and test commands for each stack before enabling merge gating.
2. Formalize monitoring/SLA requirements in the runbooks and add Terraform/Ansible snippets if needed for validators/wallet infra.
3. Document secrets management, key rotation, and PAT usage so contributors keep private data out of the repo (see `PROJECT_STATUS.md`).
4. Automate the economics notebook refresh as part of nightly or pre-release pipelines so fee/metrics stay current.

