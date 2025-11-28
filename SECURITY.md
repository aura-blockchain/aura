# Security Policy

## Reporting a Vulnerability

The AURA team takes security vulnerabilities seriously. We appreciate your efforts to responsibly disclose your findings.

### How to Report

**Please DO NOT report security vulnerabilities through public GitHub issues.**

Instead, please report them via email to: **security@[project-domain].com**

Include the following information:
- Type of vulnerability (e.g., reentrancy, overflow, access control)
- Full path to the affected source file(s)
- Step-by-step instructions to reproduce the issue
- Proof of concept or exploit code (if available)
- Potential impact of the vulnerability

### What to Expect

- **Acknowledgment**: Within 48 hours of your report
- **Initial Assessment**: Within 7 days
- **Resolution Timeline**: Depends on severity, typically 30-90 days
- **Disclosure**: Coordinated with reporter after fix is deployed

## Bug Bounty Program

We offer bug bounties paid in **AURA tokens** for responsibly disclosed vulnerabilities.

### Severity Levels and Rewards

| Severity | Description | Reward |
|----------|-------------|--------|
| **Critical** | Direct loss of funds, consensus failure, chain halt | Up to 50,000 AURA |
| **High** | Significant impact on functionality or security | Up to 10,000 AURA |
| **Medium** | Limited impact, requires specific conditions | Up to 2,500 AURA |
| **Low** | Minor issues, best practice violations | Up to 500 AURA |

### Scope

**In Scope:**
- Smart contracts and CosmWasm modules
- Consensus mechanism
- Cryptographic implementations
- Node software (aurad)
- Identity and privacy modules
- Bridge and DEX modules
- WASM execution environment

**Out of Scope:**
- Third-party dependencies (report to upstream)
- Issues already known or reported
- Theoretical vulnerabilities without proof of concept
- Social engineering attacks
- Denial of service attacks

### Rules

- Do not exploit vulnerabilities beyond proof of concept
- Do not access or modify other users' data
- Do not disrupt network operations
- Act in good faith

## Supported Versions

| Version | Supported |
|---------|-----------|
| Latest mainnet | Yes |
| Latest testnet | Yes |
| Previous major versions | Limited |

## Security Best Practices

When running an AURA node:
- Keep your node software updated
- Use hardware security modules (HSM) for validator keys
- Enable firewall and restrict RPC access
- Monitor for unusual activity
- Back up your keys securely offline
- Review WASM contracts before interaction
