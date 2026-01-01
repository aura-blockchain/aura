# Aura Blockchain Bug Bounty Program

## Program Overview

The Aura Bug Bounty Program invites security researchers to help protect the Aura blockchain ecosystem. As a community-driven project with limited resources, all rewards are paid exclusively in **AURA tokens** from our development fund.

### Community-First Approach

Aura is built by and for its community. Our bug bounty reflects this:

- **AURA Token Rewards**: All bounties paid in native AURA tokens
- **Vesting for Alignment**: Large rewards include vesting periods
- **Recognition Focus**: Hall of Fame, badges, and community acknowledgment
- **Collaborative Process**: We work with researchers as partners

## Scope

### In Scope

#### Core Blockchain
- Consensus mechanism
- Block production and validation
- State machine execution
- Transaction processing
- Network protocol (P2P, RPC, gRPC)

#### Custom Modules
- **Identity Module**: Identity management and updates
- **VC Registry Module**: Verifiable Credentials issuance/verification
- **Data Registry Module**: Data storage and IPFS integration
- **Inclusion Routines Module**: Data processing and scoring
- **Confidence Score Module**: Scoring algorithms
- **Prevalidation Module**: Data validation logic
- **DEX Module**: Decentralized exchange and liquidity pools
- **Bridge Module**: Cross-chain asset transfers
- **Validator Security Module**: Validator operations and slashing
- **Governance Module**: On-chain governance

#### Other Components
- Smart contracts deployed on Aura
- REST API endpoints
- gRPC services
- Cryptographic operations

### Out of Scope

- Third-party services and integrations
- Social engineering attacks
- Physical attacks
- Volumetric DoS attacks
- Third-party dependency bugs (report upstream)
- Theoretical vulnerabilities without PoC
- Already known issues

## Reward Structure

All rewards paid in **AURA tokens** from the development fund.

### Severity Tiers

| Severity | AURA Tokens | Vesting |
|----------|-------------|---------|
| Critical | 50,000 - 100,000 AURA | 6-month linear vest |
| High | 15,000 - 50,000 AURA | 3-month linear vest |
| Medium | 5,000 - 15,000 AURA | None |
| Low | 1,000 - 5,000 AURA | None |
| Informational | Recognition only | N/A |

### Critical

**Complete compromise or loss of user funds**

Examples:
- Consensus failure or chain halt
- Unauthorized token minting/burning
- Private key extraction
- Cross-chain bridge asset theft
- Smart contract fund drain
- Byzantine fault exploitation

### High

**Significant vulnerability affecting multiple users**

Examples:
- Authorization bypass in critical modules
- State corruption without consensus failure
- Transaction replay attacks
- Double-spend vulnerabilities
- Validator set manipulation
- Identity theft or VC forgery

### Medium

**Limited vulnerability affecting individual users**

Examples:
- Sensitive information disclosure
- Non-critical authentication bypass
- Gas manipulation exploits
- Module-specific DoS
- Race conditions in state transitions

### Low

**Minor issues with limited impact**

Examples:
- Non-sensitive information disclosure
- Timing attacks (low impact)
- Minor input validation issues
- Configuration issues

### Informational

**No immediate security risk**
- Code quality improvements
- Best practice suggestions
- Documentation errors

### Reward Modifiers

| Condition | Modifier |
|-----------|----------|
| High-quality report with PoC | +25% |
| Actionable fix suggestion | +15% |
| Fix PR included | +50% |
| First critical finding | +25% |
| Incomplete report | -25% to -50% |

## Vesting Terms

To protect AURA token economics and align long-term incentives:

- **Critical**: 6-month linear vesting (monthly releases)
- **High**: 3-month linear vesting
- **Medium/Low**: Immediate transfer

Vesting begins after fix deployment to mainnet. Researchers may opt for 50% immediate payment instead of vesting.

## Submission Process

### How to Submit

**Email**: security@aura-blockchain.io
**PGP Key**: https://aura-blockchain.io/pgp-key.txt

### Required Information

```markdown
# Vulnerability Report

## Summary
Brief description of the vulnerability

## Severity
Your assessment: Critical/High/Medium/Low

## Affected Components
- Component 1
- Component 2

## Vulnerability Details
Detailed technical description

## Reproduction Steps
1. Step 1
2. Step 2
3. ...

## Proof of Concept
Code, commands, or scripts

## Impact
Description of potential impact

## Suggested Fix (Optional)
Proposed remediation

## Contact
- Name/Handle
- AURA address for payment
```

### Response Timeline

| Stage | Timeframe |
|-------|-----------|
| Acknowledgment | 3 business days |
| Severity Assessment | 7 days |
| Status Updates | Weekly |
| Remediation | 30-90 days |
| Reward Payment | 30 days after fix |

## Responsible Disclosure

### Expectations

- Act in good faith
- Avoid privacy violations
- Do not exploit beyond demonstrating the issue
- Allow 90 days for remediation before public disclosure
- Keep information confidential until fixed
- Test only on testnets or local instances

### Our Commitments

- Acknowledge receipt within 3 business days
- Provide transparent communication on status
- Process valid submissions promptly
- Credit researchers in advisories (if desired)
- Not pursue legal action against good-faith researchers

## Legal Safe Harbor

Aura provides legal safe harbor for security research under this program. We will not pursue action against researchers who:

1. Follow the responsible disclosure policy
2. Act in good faith
3. Comply with program terms
4. Do not cause harm to users or network

## Recognition

### Hall of Fame

Contributors recognized by tier:

| Tier | Criteria |
|------|----------|
| Guardian | Critical vulnerability found |
| Champion | 3+ High severity findings |
| Contributor | Any valid finding |

### Recognition Options

- Hall of Fame listing
- Security advisory credit
- Community contributor badge
- Recommendation letter on request
- Anonymous recognition if preferred

## Program Rules

### Eligibility

- Open globally (subject to legal restrictions)
- 18+ or parental consent required
- Aura team members ineligible
- First valid report receives reward

### Payment

- AURA tokens only
- Transfer to provided AURA address
- Subject to vesting for larger rewards
- Researcher responsible for taxes

### Exclusions

- Bugs found during official audits
- Issues already reported
- Theoretical vulnerabilities without PoC
- Third-party code issues

### Modifications

- Program may be updated with 30 days notice
- Pending submissions use terms at submission time

## Contact

- **Security Email**: security@aura-blockchain.io
- **PGP Key**: https://aura-blockchain.io/pgp-key.txt
- **Response**: 3 business days

## FAQ

**Q: Why AURA tokens only?**
A: As a community-funded project, our development fund holds AURA. Token rewards align researcher incentives with project success.

**Q: What's the vesting period for?**
A: Large rewards vest over time to protect token economics and show long-term commitment.

**Q: Can I test on mainnet?**
A: No. Use testnets or local instances only. Mainnet exploitation disqualifies you.

**Q: Can I remain anonymous?**
A: Yes. We respect researcher privacy and will credit you as preferred.

**Q: What if I find a dependency vulnerability?**
A: Report to the upstream project. If it has Aura-specific impact, inform us for potential reduced reward.

---

**Last Updated**: January 1, 2026
**Program Version**: 2.0
**Program Status**: Active

*Thank you for helping keep Aura secure!*
