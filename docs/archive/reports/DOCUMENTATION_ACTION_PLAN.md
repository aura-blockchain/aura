# Documentation Action Plan - Testnet Launch Preparation

**Created**: 2025-12-24
**Target Completion**: 7 days before testnet launch
**Owner**: Documentation Team

## Overview

This action plan addresses documentation gaps identified in the comprehensive review. Tasks are prioritized by impact on testnet launch success.

---

## Phase 1: Critical (Days 1-3) - Launch Blockers

**Goal**: Complete minimum viable documentation for testnet launch.

### Task 1.1: AI Assistant Module Documentation
**Priority**: 🔴 CRITICAL
**Effort**: 8-12 hours
**Assignee**: Core team member familiar with AI assistant implementation

**Deliverable**: `chain/x/aiassistant/README.md` (400-500 lines)

**Required Sections**:
1. Overview
   - AI assistant network purpose
   - Role in identity verification
   - Reward mechanism
2. State Management
   - Assistant registration records
   - Verification task tracking
   - Reputation/performance metrics
3. Messages
   - `MsgRegisterAssistant` - Assistant onboarding
   - `MsgSubmitVerification` - Submit identity proof
   - `MsgUpdateAssistantMetadata` - Update locale/capabilities
   - `MsgDeactivateAssistant` - Exit network
4. Queries
   - Query assistant by address
   - List assistants by locale
   - Get verification tasks
   - Query performance metrics
5. Events (emit on registration, verification, rewards)
6. Errors (registration failures, invalid proofs, etc.)
7. CLI Examples
   - Register as assistant
   - Submit verification proof
   - Query assistant status
8. Integration Notes
   - For assistant operators
   - Reward calculation
   - Slashing conditions
9. Security Considerations
   - Proof validation
   - Anti-spam mechanisms
   - Reputation management
10. Best Practices
    - Recommended hardware specs
    - Uptime requirements
    - Multi-locale support

**Success Criteria**:
- [ ] All messages documented with JSON examples
- [ ] All queries have CLI examples
- [ ] Registration walkthrough complete
- [ ] Reviewed by 2+ team members

---

### Task 1.2: Auth Module Documentation
**Priority**: 🔴 CRITICAL
**Effort**: 6-8 hours
**Assignee**: Authentication/authorization expert

**Deliverable**: `chain/x/auth/README.md` (300-400 lines)

**Required Sections**:
1. Overview
   - Authentication mechanisms
   - Session management
   - Role-based access control (RBAC)
2. State Management
   - User sessions
   - Role assignments
   - Permission mappings
3. Messages
   - `MsgCreateSession` - Start user session
   - `MsgEndSession` - Logout
   - `MsgAssignRole` - Grant role
   - `MsgRevokeRole` - Remove role
4. Queries
   - Query active sessions
   - Check user permissions
   - List roles
5. Events (session lifecycle, role changes)
6. Errors (auth failures, expired sessions, etc.)
7. CLI Examples
   - Create session
   - Check permissions
   - Manage roles
8. Integration Notes
   - For wallet developers
   - Session token handling
   - Permission checks
9. Security Considerations
   - Session timeout defaults
   - Token rotation
   - Rate limiting

**Success Criteria**:
- [ ] RBAC system fully explained
- [ ] Session management workflow clear
- [ ] All CLI commands documented
- [ ] Security best practices included

---

### Task 1.3: Governance Module Documentation
**Priority**: 🔴 CRITICAL
**Effort**: 8-12 hours
**Assignee**: Governance/DAO expert

**Deliverable**: `chain/x/governance/README.md` (400-500 lines)

**Required Sections**:
1. Overview
   - DAO governance model
   - Proposal types
   - Voting mechanisms (including ZKP voting)
2. State Management
   - Proposals
   - Votes
   - Voting power calculation
3. Messages
   - `MsgSubmitProposal` - Create governance proposal
   - `MsgVote` - Cast vote (public)
   - `MsgVoteZKP` - Cast ZKP-protected vote
   - `MsgDeposit` - Deposit tokens for proposal
4. Queries
   - Query proposal by ID
   - List all proposals
   - Query votes for proposal
   - Check voting power
5. Events (proposal lifecycle, vote tallies)
6. Errors (insufficient deposit, invalid proposal, etc.)
7. CLI Examples
   - Submit text proposal
   - Submit parameter change proposal
   - Vote on proposal
   - Query proposal status
8. Integration Notes
   - For governance participants
   - Proposal format requirements
   - Voting period timelines
9. ZKP Voting Details
   - Privacy guarantees
   - Proof generation
   - Verification process
10. Best Practices
    - Proposal writing guidelines
    - Community engagement before submission
    - Voting delegation (if supported)

**Success Criteria**:
- [ ] All proposal types documented
- [ ] ZKP voting clearly explained
- [ ] Proposal lifecycle diagram included
- [ ] CLI walkthrough for common scenarios

---

### Task 1.4: Standardize GitHub URLs
**Priority**: 🔴 CRITICAL
**Effort**: 2-3 hours
**Assignee**: DevOps/Documentation lead

**Files to Update**:
1. `README.md` - Update all `github.com/decristofaroj/aura` → production org
2. `CONTRIBUTING.md` - Update fork/clone instructions
3. `docs/GETTING_STARTED.md` - Update clone URL
4. `chain/go.mod` - Verify module path
5. All proto files - Verify `go_package` directives

**Script to Find References**:
```bash
grep -r "decristofaroj\|github.com/aequitas" \
  README.md CONTRIBUTING.md docs/*.md \
  --exclude-dir=.git --exclude-dir=vendor
```

**Success Criteria**:
- [ ] All GitHub URLs point to production org
- [ ] All external links tested and working
- [ ] Go module paths verified
- [ ] Proto package paths consistent

---

### Task 1.5: Testnet Participation Guide
**Priority**: 🔴 CRITICAL
**Effort**: 6-8 hours
**Assignee**: Community/documentation lead

**Deliverable**: `docs/testnet/TESTNET_GUIDE.md` (250-300 lines)

**Required Sections**:
1. Introduction
   - Testnet purpose and goals
   - Expected timeline
   - Rewards/incentives (if any)
2. Prerequisites
   - System requirements
   - Initial AURA token acquisition
3. Quick Start
   - Join testnet in 10 minutes
   - Connect to public RPC/REST endpoints
   - Get tokens from faucet
4. Faucet Usage
   - Faucet URL
   - Request limits
   - Alternative methods
5. Explorer Access
   - Block explorer URL
   - Transaction lookup
   - Account monitoring
6. Common Scenarios
   - Send your first transaction
   - Stake tokens
   - Create a validator
   - Submit a governance proposal
   - Participate in identity verification
7. Testnet-Specific Configuration
   - Chain ID: `aura-testnet-1`
   - Genesis file location
   - Seed nodes
   - Persistent peers
8. Known Issues
   - Current limitations
   - Workarounds
9. Reporting Issues
   - GitHub issues
   - Discord/community channels
   - Bug bounty eligibility
10. Testnet Phases
    - Phase 1: Node operator testing
    - Phase 2: Public participation
    - Phase 3: Load testing
    - Mainnet preparation

**Additional Files**:
- `networks/testnet/genesis.json` - Testnet genesis
- `networks/testnet/seeds.txt` - Seed node list
- `networks/testnet/peers.txt` - Persistent peers

**Success Criteria**:
- [ ] Users can join testnet in <15 minutes
- [ ] Faucet accessible and documented
- [ ] Explorer link working
- [ ] All testnet endpoints listed
- [ ] Reviewed by 3+ team members

---

## Phase 2: High Priority (Days 4-5) - Enhanced UX

**Goal**: Improve contributor experience and standardize processes.

### Task 2.1: GitHub Issue Templates
**Priority**: 🟡 HIGH
**Effort**: 2-3 hours
**Assignee**: Developer relations lead

**Deliverables**:
1. `.github/ISSUE_TEMPLATE/bug_report.md`
2. `.github/ISSUE_TEMPLATE/feature_request.md`
3. `.github/ISSUE_TEMPLATE/question.md`
4. `.github/ISSUE_TEMPLATE/config.yml` (template selector)

**Bug Report Template** (150-200 lines):
```markdown
---
name: Bug Report
about: Report a bug in the Aura blockchain
title: "[BUG] "
labels: bug, needs-triage
assignees: ''
---

## Bug Description
A clear description of the bug.

## Steps to Reproduce
1. Step one
2. Step two
3. ...

## Expected Behavior
What should happen.

## Actual Behavior
What actually happens.

## Environment
- **OS**: [e.g., Ubuntu 22.04]
- **Aura Version**: [e.g., 0.1.0-testnet]
- **Go Version**: [e.g., 1.21.3]
- **Node Type**: [validator/full node/light client]

## Logs
```
Paste relevant logs here
```

## Additional Context
Screenshots, error messages, etc.

## Possible Solution
(Optional) Suggestions for fixing the bug.
```

**Success Criteria**:
- [ ] Templates cover all common issue types
- [ ] Labels auto-assigned
- [ ] Required fields enforced
- [ ] Tested by submitting sample issues

---

### Task 2.2: Pull Request Template
**Priority**: 🟡 HIGH
**Effort**: 1-2 hours
**Assignee**: Core contributor

**Deliverable**: `.github/PULL_REQUEST_TEMPLATE.md` (100-150 lines)

**Template Sections**:
1. **Description**
   - Summary of changes
   - Related issues
2. **Type of Change**
   - [ ] Bug fix
   - [ ] New feature
   - [ ] Breaking change
   - [ ] Documentation update
3. **Testing**
   - [ ] Unit tests added/updated
   - [ ] Integration tests added/updated
   - [ ] Manual testing performed
4. **Documentation**
   - [ ] README updated (if needed)
   - [ ] Module README updated (if needed)
   - [ ] CHANGELOG updated
   - [ ] API docs updated (if applicable)
5. **Checklist**
   - [ ] Code follows project style guidelines
   - [ ] Self-review completed
   - [ ] Comments added for complex logic
   - [ ] No new warnings generated
   - [ ] All tests passing locally
   - [ ] Conventional commit messages used
6. **Screenshots** (if UI changes)
7. **Reviewer Notes**
   - Special instructions for reviewers

**Success Criteria**:
- [ ] Template enforces documentation updates
- [ ] Testing requirements clear
- [ ] Used successfully in 3+ PRs

---

### Task 2.3: Remaining Module READMEs (Medium Priority)
**Priority**: 🟡 HIGH
**Effort**: 8-10 hours total
**Assignee**: Module owners

**Deliverables**:
1. `chain/x/dataregistry/README.md` (200-250 lines)
2. `chain/x/economicsecurity/README.md` (200-250 lines)
3. `chain/x/security/README.md` (200-250 lines)
4. `chain/x/walletsecurity/README.md` (200-250 lines)

**Template** (same structure as existing module READMEs):
- Overview and features
- State management
- Messages (with examples)
- Queries (with CLI examples)
- Events
- Errors
- CLI examples
- Integration notes
- Security considerations

**Success Criteria**:
- [ ] All 4 modules documented
- [ ] Consistent with existing module docs
- [ ] Reviewed by module owners

---

### Task 2.4: CONTRIBUTORS.md
**Priority**: 🟡 HIGH
**Effort**: 2-3 hours
**Assignee**: Community manager

**Deliverable**: `CONTRIBUTORS.md` (100-150 lines)

**Sections**:
1. **Core Team**
   - Project leads
   - Module owners
   - Role descriptions
2. **Contributors**
   - Alphabetical list with contribution areas
   - GitHub usernames
3. **Special Thanks**
   - Advisors
   - Early testers
   - Security researchers
4. **Contribution Statistics**
   - Total contributors: X
   - Total commits: Y
   - Languages represented: Z
5. **How to Become a Contributor**
   - Link to CONTRIBUTING.md
   - "Good First Issue" label
   - Contributor tiers (if applicable)

**Success Criteria**:
- [ ] All contributors recognized
- [ ] Updated monthly (automated?)
- [ ] Linked from README.md

---

### Task 2.5: Expanded FAQ
**Priority**: 🟡 HIGH
**Effort**: 3-4 hours
**Assignee**: Documentation team

**Deliverable**: `docs/FAQ.md` (200-250 lines)

**Sections**:
1. **General**
   - What is Aura?
   - How is it different from other blockchains?
   - What is the AURA token used for?
2. **Testnet**
   - How do I join the testnet?
   - Where do I get testnet tokens?
   - Can I run a validator on testnet?
   - Are there rewards for testnet participation?
3. **Technical**
   - What are the system requirements?
   - Which ports need to be open?
   - How do I upgrade my node?
   - What is the block time?
   - What consensus mechanism is used?
4. **Identity & Privacy**
   - How does identity verification work?
   - Is my personal data stored on-chain?
   - What are Inclusion Routines?
   - How do AI assistants verify identity?
5. **Governance**
   - How do I submit a proposal?
   - What is the voting period?
   - What is ZKP voting?
6. **Troubleshooting**
   - Node won't sync - common causes
   - Transaction failed - what to check
   - "Permission denied" errors
   - Validator getting slashed - why?
7. **Development**
   - How do I build a custom module?
   - Where is the API documentation?
   - Are there SDKs available?
   - Can I use CosmWasm contracts?
8. **Community**
   - How do I get help?
   - Where can I report bugs?
   - Is there a bug bounty program?
   - How do I contribute?

**Success Criteria**:
- [ ] Covers top 30+ questions
- [ ] Linked from README and docs index
- [ ] Regularly updated based on support tickets

---

## Phase 3: Medium Priority (Days 6-7) - Polish

**Goal**: Finalize professional touches and verify consistency.

### Task 3.1: Verify All Email Addresses
**Priority**: 🟢 MEDIUM
**Effort**: 2-3 hours
**Assignee**: Infrastructure lead

**Emails to Verify**:
1. `security@aura.network` (SECURITY.md)
2. `conduct@aequitas-labs.io` (CODE_OF_CONDUCT.md)
3. `support@aura.network` (docs/GETTING_STARTED.md)

**Steps**:
- [ ] Set up email accounts/forwarding
- [ ] Test sending/receiving
- [ ] Set up auto-responders (if needed)
- [ ] Generate PGP key for `security@aura.network`
- [ ] Publish PGP public key
- [ ] Update SECURITY.md with PGP fingerprint
- [ ] Document email response SLAs

**Success Criteria**:
- [ ] All emails active and tested
- [ ] PGP encryption available for security reports
- [ ] Monitored by appropriate team members

---

### Task 3.2: Quick Fixes Checklist
**Priority**: 🟢 MEDIUM
**Effort**: 1-2 hours
**Assignee**: Any team member

**Tasks**:
- [ ] Update version in README footer to `0.1.0-testnet`
- [ ] Add testnet genesis to `networks/testnet/genesis.json`
- [ ] Add testnet seeds to README configuration example
- [ ] Test all external links in README.md (use link checker tool)
- [ ] Activate or remove Discord badge
- [ ] Activate or remove Twitter badge
- [ ] Verify Codecov/SonarCloud badges working
- [ ] Spell-check all root documentation
- [ ] Validate all JSON examples in documentation
- [ ] Check for broken internal links

**Tools**:
```bash
# Link checker
npm install -g markdown-link-check
markdown-link-check README.md

# Spell checker
npm install -g markdown-spellcheck
mdspell README.md -n -a --en-us

# JSON validator
find docs -name "*.md" -exec grep -Pzo '```json\n\K.*?(?=\n```)' {} \; | jq empty
```

**Success Criteria**:
- [ ] All external links return 200 OK
- [ ] No spelling errors in root docs
- [ ] All JSON examples valid
- [ ] Badges reflect current status

---

### Task 3.3: Branding Consistency Review
**Priority**: 🟢 MEDIUM
**Effort**: 2-3 hours
**Assignee**: Marketing/brand lead

**Review Areas**:
1. **Project Naming**
   - Decide: "Aura" vs "Aequitas" vs "AURA"
   - Create brand guide: `docs/BRAND_GUIDE.md`
   - Standardize usage across all docs
2. **Logo/Visual Assets**
   - Ensure logo used consistently
   - README banner/header image
   - Favicon for docs site
3. **Tone/Voice**
   - Professional but accessible
   - Avoid jargon in user-facing docs
   - Consistent terminology

**Deliverable**: `docs/BRAND_GUIDE.md` (50-100 lines)

**Success Criteria**:
- [ ] Clear naming conventions documented
- [ ] Visual assets uploaded to `/docs/assets/`
- [ ] All docs use consistent terminology

---

## Phase 4: Future Enhancements (Post-Launch)

**Goal**: Continuous improvement based on community feedback.

### Future Task 4.1: Video Tutorials (1-2 weeks)
- Node setup walkthrough (10-15 min)
- First transaction tutorial (5-10 min)
- Validator creation guide (15-20 min)
- AI assistant registration (10-15 min)

**Platform**: YouTube, embedded in docs site

---

### Future Task 4.2: Architecture Decision Records (Ongoing)
- Document key design decisions
- Maintain ADR log in `docs/architecture/adr/`
- Format: ADR-0001-format.md

---

### Future Task 4.3: Interactive Tutorials (2-3 weeks)
- Katacoda/Killercoda scenarios
- In-browser node deployment
- Guided transaction creation

---

### Future Task 4.4: Localization (3-4 months)
- Translate core docs to:
  - Spanish
  - Chinese (Simplified)
  - Japanese
  - Korean
  - Russian
- Community translation program

---

## Progress Tracking

### Phase 1 (Critical) - Target: Days 1-3
| Task | Assignee | Status | Completion |
|------|----------|--------|------------|
| 1.1 AI Assistant README | TBD | Not Started | 0% |
| 1.2 Auth README | TBD | Not Started | 0% |
| 1.3 Governance README | TBD | Not Started | 0% |
| 1.4 GitHub URL Standardization | TBD | Not Started | 0% |
| 1.5 Testnet Guide | TBD | Not Started | 0% |

### Phase 2 (High Priority) - Target: Days 4-5
| Task | Assignee | Status | Completion |
|------|----------|--------|------------|
| 2.1 Issue Templates | TBD | Not Started | 0% |
| 2.2 PR Template | TBD | Not Started | 0% |
| 2.3 Module READMEs (4) | TBD | Not Started | 0% |
| 2.4 CONTRIBUTORS.md | TBD | Not Started | 0% |
| 2.5 Expanded FAQ | TBD | Not Started | 0% |

### Phase 3 (Medium Priority) - Target: Days 6-7
| Task | Assignee | Status | Completion |
|------|----------|--------|------------|
| 3.1 Email Verification | TBD | Not Started | 0% |
| 3.2 Quick Fixes | TBD | Not Started | 0% |
| 3.3 Branding Review | TBD | Not Started | 0% |

---

## Definition of Done

**For each documentation task:**
- [ ] Content written and formatted
- [ ] Technical accuracy verified by subject matter expert
- [ ] Reviewed by at least 2 team members
- [ ] Spell-checked and grammar-checked
- [ ] Code examples tested (CLI commands, JSON, etc.)
- [ ] Links verified
- [ ] Consistent with existing documentation style
- [ ] Merged to main branch
- [ ] Announced in team/community channels (if significant)

**For testnet launch readiness:**
- [ ] All Phase 1 tasks completed
- [ ] All Phase 2 tasks completed OR deferred with justification
- [ ] Phase 3 tasks at least 50% complete
- [ ] Final documentation review meeting conducted
- [ ] Launch announcement blog post/tweet drafted

---

## Risk Mitigation

### Risk 1: Key team members unavailable
**Mitigation**: Cross-train on module documentation, pair on critical docs

### Risk 2: Technical accuracy issues
**Mitigation**: Require SME review before merging, automated testing of code examples

### Risk 3: Timeline slippage
**Mitigation**: Phase 1 is non-negotiable, Phases 2-3 can be deferred if needed

### Risk 4: Community confusion post-launch
**Mitigation**: Monitor Discord/GitHub issues first 48 hours, rapid FAQ updates

---

## Success Metrics

**Week 1 Post-Launch**:
- [ ] <5 "how do I get started" questions per day
- [ ] >80% of GitHub issues use templates
- [ ] >90% of PRs use template
- [ ] Testnet guide used by 100+ participants

**Month 1 Post-Launch**:
- [ ] All modules have README documentation
- [ ] FAQ covers top 50 questions
- [ ] Documentation site has 1000+ unique visitors
- [ ] <10% of support questions are "where is this documented"

---

## Approval & Sign-Off

**Prepared by**: Documentation Team
**Reviewed by**: [Name, Role]
**Approved by**: [Project Lead]
**Date**: 2025-12-24

**Next Review**: After Phase 1 completion (Day 3)

---

## Notes

- This plan is living document - update as priorities shift
- Defer non-critical tasks if launch timeline is tight
- Community feedback will drive Phase 4 priorities
- Celebrate documentation milestones with team!
