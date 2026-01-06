# Release Artifacts Policy

This project follows standard blockchain OSS expectations:
- Source repos contain source code, configs, docs, and tests.
- Built binaries and large artifacts are published separately.
- Releases include checksums and signatures.

## What belongs in the repo
- Source code, scripts, configs, docs
- Build tooling and CI workflows
- Tests and fixtures

## What must NOT be committed
- Compiled binaries and test executables (example: `aurad`, `*.test`)
- Large build outputs (example: `build/`, `dist/`, `target/`)
- Docker image exports or tarballs
- Secrets, keys, mnemonics, or credentials

## Release artifacts: required deliverables
Each tagged release should publish:
- Platform binaries (Linux/macOS/Windows as needed)
- `SHA256SUMS` (or equivalent)
- Signature for checksums (GPG or cosign)
- Optional SBOM (recommended)

## Suggested release flow
1) Update `CHANGELOG.md` and tag the release.
2) Build artifacts from a clean environment.
3) Generate checksums and sign them.
4) Upload artifacts and signatures to your release host.

Example checksum + signature steps:
```bash
# Generate checksums
sha256sum aurad-* > SHA256SUMS

# Sign checksums with GPG
gpg --armor --detach-sign SHA256SUMS

# Or sign with cosign (keyless or key-based)
cosign sign-blob --output-signature SHA256SUMS.sig SHA256SUMS
```

## Verification for users
```bash
# Verify checksum
sha256sum -c SHA256SUMS

# Verify GPG signature
gpg --verify SHA256SUMS.asc SHA256SUMS
```

## Reproducibility guidance
- Build from a clean checkout
- Pin toolchains and dependencies
- Document build steps and environment
- Prefer deterministic builds where possible

## Enforcement
- `.gitignore` files are configured to exclude binaries and build outputs.
- If a binary is needed for distribution, publish it via the release channel.
