# Contributing to Aequitas / AURA

Thanks for helping build the AURA blockchain. To minimize confusion and keep security airtight, follow these guidelines.

## Workflow

1. **Open/Find an Issue:** Describe the change and link to relevant RFCs.
2. **Submit/Update an RFC:** Major protocol or architecture work must land in `docs/rfcs/` before code merges.
3. **Create a Feature Branch:** `git checkout -b feature/<scope>`.
4. **Add Tests + Docs:** Every change should include unit tests and documentation updates where applicable.
5. **Pull Request Checklist:**
   - Reference Issue/RFC IDs.
   - Describe testing performed.
   - Ensure CI passes (Go tests, linting, proto checks, wallet/assistant tests).
6. **Reviews:** CODEOWNERS must approve modules they own. Address all feedback before merging.

## Coding Standards

- Go 1.22+ for the chain, Rust/Python for assistants, React Native or Flutter for wallet.
- Follow existing formatting tools (`gofmt`, `eslint`, `rustfmt`).
- Keep commits scoped and descriptive.

## Security & Privacy

- Never commit user data, secrets, or private keys.
- Report vulnerabilities privately via security@ (placeholder) before filing public issues.

## Community Conduct

Be respectful, document decisions, and assume good intent. We are building a security-critical identity layer together.
