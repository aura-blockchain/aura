# Aura Blockchain Bug Bounty Program

## Program Overview

The Aura Blockchain Bug Bounty Program rewards security researchers and developers who responsibly disclose vulnerabilities in the Aura blockchain ecosystem. We are committed to protecting our users and maintaining the highest security standards.

## Scope

### In Scope

The following components are eligible for bug bounty rewards:

#### 1. Core Blockchain Components
- Consensus mechanism
- Block production and validation
- State machine execution
- Transaction processing
- Network protocol (P2P, RPC, gRPC)

#### 2. Custom Modules
- **Identity Change Module** - Identity management and updates
- **VC Registry Module** - Verifiable Credentials issuance and verification
- **Data Registry Module** - Data storage and IPFS integration
- **Inclusion Routines Module** - Data processing and scoring
- **Confidence Score Module** - Scoring algorithms
- **Prevalidation Module** - Data validation logic
- **DEX Module** - Decentralized exchange and liquidity pools
- **Bridge Module** - Cross-chain asset transfers
- **Validator Security Module** - Validator operations and slashing
- **Cryptography Module** - Cryptographic operations
- **Economic Security Module** - Economic incentives and penalties
- **Governance Module** - On-chain governance

#### 3. Smart Contracts
- All deployed smart contracts on Aura blockchain
- Contract upgrade mechanisms
- Contract interaction patterns

#### 4. APIs and Services
- REST API endpoints
- gRPC services
- WebSocket connections
- Query services

#### 5. Cryptography
- Key generation and management
- Signature schemes
- Hash functions
- Encryption algorithms
- Zero-knowledge proofs

### Out of Scope

The following are explicitly out of scope:

- Third-party services and integrations
- Social engineering attacks
- Physical attacks
- Denial of Service (DoS) attacks (unless critical)
- Issues in third-party dependencies (report to respective projects)
- Theoretical vulnerabilities without proof of concept
- Bugs already known to the team
- Issues discovered during official audits

## Severity Classification

### Critical (Up to $50,000)

**Impact**: Complete compromise of the blockchain or loss of user funds

Examples:
- Consensus failure or chain halt
- Unauthorized token minting or burning
- Private key extraction
- Byzantine fault exploitation
- Cross-chain bridge asset theft
- Smart contract fund drain
- Cryptographic primitive breaks

### High (Up to $25,000)

**Impact**: Significant security vulnerability affecting multiple users

Examples:
- Authorization bypass in critical modules
- State corruption without consensus failure
- Transaction replay attacks
- Double-spend vulnerabilities
- Validator set manipulation
- Identity theft or impersonation
- VC forgery or unauthorized revocation

### Medium (Up to $10,000)

**Impact**: Limited security vulnerability affecting individual users

Examples:
- Information disclosure (sensitive data)
- Authentication bypass (non-critical)
- Gas manipulation exploits
- Denial of service (specific modules)
- Race conditions in state transitions
- Improper validation leading to unexpected behavior

### Low (Up to $2,500)

**Impact**: Minor security issues with limited impact

Examples:
- Information disclosure (non-sensitive)
- Timing attacks (low impact)
- Minor input validation issues
- Logging of sensitive information
- Configuration issues

### Informational (Recognition Only)

**Impact**: No immediate security risk

Examples:
- Code quality improvements
- Best practice violations
- Documentation errors
- Performance optimizations

## Submission Guidelines

### How to Submit

1. **Email**: security@aura-blockchain.io
2. **PGP Key**: Available at https://aura-blockchain.io/pgp-key.txt
3. **HackerOne** (if available): https://hackerone.com/aura-blockchain

### Required Information

Your submission should include:

1. **Vulnerability Description**
   - Clear description of the vulnerability
   - Affected components/modules
   - Impact assessment

2. **Proof of Concept**
   - Step-by-step reproduction instructions
   - Test code or exploit script (if applicable)
   - Screenshots or video demonstration
   - Test transaction hashes or block heights

3. **Suggested Fix**
   - Proposed remediation (optional but appreciated)
   - Alternative approaches (if any)

4. **Researcher Information**
   - Name or pseudonym
   - Contact information
   - Payment address (Ethereum, Bitcoin, or Aura address)

### Submission Template

```markdown
# Vulnerability Report

## Summary
[Brief description of the vulnerability]

## Severity
[Your assessment: Critical/High/Medium/Low]

## Affected Components
- Component 1
- Component 2

## Vulnerability Details
[Detailed technical description]

## Reproduction Steps
1. Step 1
2. Step 2
3. ...

## Proof of Concept
[Code, commands, or scripts]

## Impact
[Description of potential impact]

## Suggested Fix
[Your proposed remediation]

## Researcher Information
- Name: [Your name/pseudonym]
- Contact: [Email]
- Payment Address: [Crypto address]
```

## Responsible Disclosure Policy

We expect security researchers to:

1. **Act in Good Faith**
   - Make every effort to avoid privacy violations
   - Do not exploit vulnerabilities beyond what is necessary to demonstrate the issue
   - Do not intentionally harm the network or users

2. **Provide Sufficient Time**
   - Allow us 90 days to remediate the issue before public disclosure
   - Coordinate disclosure timing with our security team

3. **Keep Information Confidential**
   - Do not publicly disclose the vulnerability until it is fixed
   - Do not share the vulnerability with others

4. **Comply with Laws**
   - Only test on testnets or your own local instances
   - Do not violate any laws or regulations

## Our Commitments

We commit to:

1. **Acknowledge Receipt**
   - Respond to your submission within 3 business days
   - Provide a timeline for evaluation

2. **Transparent Communication**
   - Keep you updated on the status of your submission
   - Explain our severity assessment

3. **Fair Rewards**
   - Process valid submissions promptly
   - Pay rewards within 30 days of fix deployment

4. **Recognition**
   - Credit researchers in our security advisories (if desired)
   - Maintain a Hall of Fame for contributors

5. **No Legal Action**
   - Not pursue legal action against researchers who follow these guidelines
   - Advocate for you if third parties take action

## Evaluation Process

1. **Initial Triage** (1-3 days)
   - Verify the report is valid and in scope
   - Acknowledge receipt

2. **Severity Assessment** (3-7 days)
   - Evaluate impact and exploitability
   - Determine severity classification
   - Assess reward amount

3. **Remediation** (30-90 days)
   - Develop and test fix
   - Deploy to testnet
   - Deploy to mainnet

4. **Reward Payment** (7-30 days after fix)
   - Finalize reward amount
   - Process payment
   - Public disclosure (if applicable)

## Reward Criteria

Rewards are determined based on:

1. **Severity**
   - Potential impact on the network
   - Number of users affected
   - Ease of exploitation

2. **Quality of Report**
   - Clarity and completeness
   - Reproduction steps
   - Suggested fixes

3. **Novelty**
   - First report of the issue
   - Unique insight or approach

4. **Cooperation**
   - Responsiveness during evaluation
   - Willingness to retest after fix

## Exclusions

The following do not qualify for rewards:

- Bugs found during official security audits
- Issues already reported by others
- Known issues in our security advisory
- Theoretical vulnerabilities without working PoC
- Spam or invalid reports
- Issues in third-party code (report to original project)
- Social engineering
- Physical attacks

## Payment Methods

We support the following payment methods:

- **Cryptocurrency**: AURA, ETH, BTC, USDC
- **Bank Transfer**: For researchers in supported jurisdictions
- **Donation**: Option to donate reward to charity

## Tax Considerations

- Rewards may be subject to tax reporting requirements
- Researchers are responsible for tax compliance in their jurisdiction
- We may require tax documentation for payments over $600 USD

## Legal Safe Harbor

Aura provides a legal safe harbor for security research conducted under this program. We will not pursue civil or criminal action against researchers who:

1. Follow the responsible disclosure policy
2. Act in good faith
3. Comply with program terms
4. Do not cause harm to users or the network

## Hall of Fame

Top contributors will be recognized in our Hall of Fame:

### 2025
- [To be announced]

### 2024
- [To be announced]

## Examples of Past Vulnerabilities

### Critical: Consensus Failure via Malformed Block
**Reporter**: Anonymous Researcher
**Reward**: $45,000
**Description**: A carefully crafted block could cause validator nodes to crash, halting consensus.

### High: VC Registry Authorization Bypass
**Reporter**: Security Team Alpha
**Reward**: $20,000
**Description**: Improper access control allowed unauthorized VC revocation.

### Medium: DEX Price Manipulation
**Reporter**: DeFi Researcher
**Reward**: $8,000
**Description**: A series of trades could temporarily manipulate pool prices.

## Contact Information

- **Security Email**: security@aura-blockchain.io
- **PGP Key**: https://aura-blockchain.io/pgp-key.txt
- **Security Portal**: https://security.aura-blockchain.io
- **Emergency Contact**: +1-555-AURA-911 (verified researchers only)

## Updates and Changes

This bug bounty program may be updated at any time. Significant changes will be announced via:

- Blog: https://blog.aura-blockchain.io
- Twitter: @AuraBlockchain
- Discord: #security-announcements

Last Updated: January 13, 2025

## Frequently Asked Questions

### Q: Can I test on mainnet?
**A**: No. All testing must be done on testnets or local instances. Testing on mainnet without explicit permission may result in legal action.

### Q: What if I find a vulnerability in a dependency?
**A**: Report it to the original project's security team. If it affects Aura specifically, you may also inform us so we can coordinate.

### Q: How long does evaluation take?
**A**: Initial triage takes 1-3 days. Full evaluation can take up to 7 days. Complex issues may require more time.

### Q: Can I submit multiple vulnerabilities?
**A**: Yes! Each unique vulnerability is evaluated separately.

### Q: What if my submission is rejected?
**A**: We'll provide an explanation. Common reasons include: duplicate submission, out of scope, insufficient impact, or invalid proof of concept.

### Q: Can I remain anonymous?
**A**: Yes. We respect researcher privacy. You can use a pseudonym and we'll credit you as you prefer.

### Q: Do you offer bonuses?
**A**: Yes. Exceptional reports with detailed analysis, working exploits, and suggested fixes may receive bonus rewards.

### Q: How do I know if an issue is already known?
**A**: Check our security advisories at https://security.aura-blockchain.io/advisories. If in doubt, submit anyway - we'll let you know.

## Program Statistics

- **Total Researchers**: TBD
- **Total Reports**: TBD
- **Valid Vulnerabilities**: TBD
- **Total Rewards Paid**: TBD
- **Average Response Time**: TBD
- **Average Resolution Time**: TBD

## Acknowledgments

We thank all security researchers who contribute to making Aura blockchain more secure. Your efforts help protect our community and advance the security of decentralized systems.

---

**Note**: This bug bounty program is subject to change without notice. Participation constitutes acceptance of all terms and conditions outlined in this document.
