# AURA Inclusion Routine Management Guide

## Overview

This guide explains how to manage the 300 Inclusion Routines (IRs) in the AURA blockchain without touching core code. All IR operations are handled through simple interfaces and governance transactions.

---

## Quick Start

### Method 1: Command Line Interface (Recommended)

```bash
# Make the script executable
chmod +x scripts/ir_manager.sh

# Run the interactive manager
./scripts/ir_manager.sh
```

The interactive menu provides:
- List all IRs
- View IR details
- Create new IR
- Update existing IR
- Delete IR
- Bulk import from JSON
- Export to JSON
- View statistics
- Validate IRs

### Method 2: Direct CLI Commands

#### Create a New IR

```bash
aurad tx inclusionroutines create-ir \
  "IR-301" \
  "My New IR" \
  "ARENA_SPECIALIZED" \
  "Description of the new IR" \
  15 \
  8 \
  "global,us" \
  "PRIVACY_TIER_MEDIUM" \
  "1.0" \
  "" \
  0 \
  0 \
  --from <your-key> \
  --chain-id aura-1 \
  --yes
```

#### Update an Existing IR

```bash
aurad tx inclusionroutines update-ir \
  "IR-301" \
  --name="Updated IR Name" \
  --description="New description" \
  --score=20 \
  --from <your-key> \
  --chain-id aura-1 \
  --yes
```

#### Delete an IR

```bash
aurad tx inclusionroutines delete-ir \
  "IR-301" \
  --from <your-key> \
  --chain-id aura-1 \
  --yes
```

#### Query IRs

```bash
# List all IRs
aurad query inclusionroutines list-irs

# Get specific IR
aurad query inclusionroutines ir "IR-001"

# Filter by arena
aurad query inclusionroutines list-irs --arena-filter="ARENA_BIOMETRIC"

# Filter by status
aurad query inclusionroutines list-irs --status-filter="IR_STATUS_ACTIVE"
```

---

## Bulk Operations

### Import All 300 IRs from JSON

```bash
# Using the interactive script
./scripts/ir_manager.sh
# Select option 6 (Bulk import)
# Enter path: data/inclusion_routines/ir_genesis_300.json

# Or use direct command
cat data/inclusion_routines/ir_genesis_300.json | jq -r '.irs[] |
  @json' | while read ir; do
    # Import each IR
    echo "Importing $(echo $ir | jq -r .id)..."
    # ... create-ir command here
done
```

### Export Current IRs

```bash
# Using interactive script
./scripts/ir_manager.sh
# Select option 7 (Export)

# Or direct query
aurad query inclusionroutines list-irs --output json > my_irs.json
```

---

## IR Design Guidelines

### Creating New IRs

When creating a new IR, ensure:

1. **Point Value: 10-30**
   - Minimum: 10 points (ensures max 10 IRs for verification)
   - Maximum: 30 points (prevents single-method reliance)

2. **Trinity Category** (must specify one):
   - `official_document` - Government-issued documents
   - `biometric` - Biometric verification
   - `witnessed_activity` - AI-witnessed real-time activities

3. **Arena Assignment**:
   - `ARENA_HIGH_ASSURANCE` - Official documents (20-30 pts)
   - `ARENA_BIOMETRIC` - Biometric identity (10-25 pts)
   - `ARENA_POSSESSION` - Witnessed activities (15-25 pts)
   - `ARENA_SOCIAL` - Social graph (10-15 pts)
   - `ARENA_GEOLOCATION` - Location/device (10-15 pts)
   - `ARENA_PERSISTENCE` - Temporal consistency (10-15 pts)
   - `ARENA_KNOWLEDGE` - Knowledge-based (10-12 pts)
   - `ARENA_ANCHOR` - Multi-factor combos (20-30 pts)
   - `ARENA_SPECIALIZED` - Fun/creative (10-15 pts)

4. **Privacy Tier**:
   - `PRIVACY_TIER_HIGH` - Documents, biometrics
   - `PRIVACY_TIER_MEDIUM` - Social, location
   - `PRIVACY_TIER_LOW` - Public/fun challenges

### Example: Creating a Fun IR

```json
{
  "id": "IR-301",
  "name": "Coffee Shop Selfie Challenge",
  "arena": "ARENA_SPECIALIZED",
  "description": "Take selfie at your favorite coffee shop, AI validates face + location",
  "score": 10,
  "poi_reward": 5,
  "locale_tags": ["global"],
  "privacy_tier": "PRIVACY_TIER_LOW",
  "version": "1.0",
  "trinity_category": "",
  "fun_factor": 8
}
```

---

## Governance Process

### IR Lifecycle

1. **DRAFT** → Proposed by community
2. **REVIEWING** → Under governance review
3. **APPROVED** → Approved by governance vote
4. **ACTIVE** → Live and available to users
5. **SUSPENDED** → Temporarily disabled
6. **DEPRECATED** → Marked for removal
7. **RETIRED** → Removed from system

### Proposing New IRs

```bash
# Submit governance proposal to add new IR
aurad tx gov submit-proposal \
  --type="ir-addition" \
  --title="Add IR-301: Coffee Shop Selfie Challenge" \
  --description="Proposal to add fun location-based IR" \
  --deposit="10000000uaura" \
  --from <your-key> \
  --chain-id aura-1
```

### Suspending/Activating IRs

```bash
# Suspend an IR (if fraud detected)
aurad tx inclusionroutines suspend-ir \
  "IR-XXX" \
  "Reason for suspension" \
  --from <authority> \
  --chain-id aura-1

# Reactivate an IR
aurad tx inclusionroutines activate-ir \
  "IR-XXX" \
  --from <authority> \
  --chain-id aura-1
```

---

## Validation & Quality Control

### Automated Validation

The system automatically validates:
- ✅ Point values are 10-30
- ✅ No duplicate IR IDs
- ✅ Valid arena assignments
- ✅ Valid privacy tiers
- ✅ Trinity categories balanced

### Running Validation

```bash
# Using interactive script
./scripts/ir_manager.sh
# Select option 9 (Validate)

# Or check manually
aurad query inclusionroutines list-irs | jq '[.irs[] |
  select(.score < 10 or .score > 30)] | length'
```

### Trinity Balance Check

Every verification path MUST include:
1. At least 1 IR from `official_document` category
2. At least 1 IR from `biometric` category
3. At least 1 IR from `witnessed_activity` category

Check balance:
```bash
# Count IRs per trinity category
aurad query inclusionroutines list-irs | jq '
  [.irs[] | .trinity_category] |
  group_by(.) |
  map({category: .[0], count: length})'
```

---

## Prerequisites & Dependencies

### Setting IR Prerequisites

Some IRs require completing other IRs first:

```bash
# IR-032 requires IR-031 and IR-001
aurad tx inclusionroutines set-ir-prerequisites \
  "IR-032" \
  "IR-031,IR-001" \
  --from <authority> \
  --chain-id aura-1
```

### Query Prerequisite Graph

```bash
# See what IRs depend on a specific IR
aurad query inclusionroutines ir-graph "IR-031"
```

---

## Rate Limiting

### Setting Rate Limits

Prevent gaming by limiting how often users can attempt an IR:

```bash
aurad tx inclusionroutines set-ir-rate-limit \
  "IR-001" \
  --per-wallet-per-hour=1 \
  --per-wallet-per-day=1 \
  --per-block-global=100 \
  --from <authority> \
  --chain-id aura-1
```

### Rate Limit Guidelines

- **High-value IRs (25-30 pts)**: 1 per day
- **Medium IRs (15-24 pts)**: 3 per day
- **Fun IRs (10-14 pts)**: Unlimited or 10 per day
- **Spontaneous IRs**: No retry (one attempt only)

---

## Monitoring & Analytics

### View IR Statistics

```bash
# Using interactive script
./scripts/ir_manager.sh
# Select option 8 (Statistics)
```

Shows:
- Total IR count
- Active vs. suspended
- Distribution by arena
- Point distribution
- Average points per IR

### User Completion Analytics

```bash
# Query most popular IRs
aurad query inclusionroutines analytics --metric="completion-rate"

# Query average path to verification
aurad query inclusionroutines analytics --metric="average-path"
```

---

## Troubleshooting

### Common Issues

**Issue: "IR already exists"**
```bash
# Solution: Use update instead of create
aurad tx inclusionroutines update-ir "IR-XXX" ...
```

**Issue: "Invalid arena"**
```bash
# Solution: Use exact arena enum names
ARENA_HIGH_ASSURANCE
ARENA_BIOMETRIC
ARENA_POSSESSION
# etc.
```

**Issue: "Authority not authorized"**
```bash
# Solution: Only governance module can create/update IRs
# Submit governance proposal instead
```

### Reset IR System (Development Only)

```bash
# WARNING: This deletes all IRs!
# Only use in development/testing

# Delete all IRs
aurad query inclusionroutines list-irs | jq -r '.irs[].id' | while read id; do
  aurad tx inclusionroutines delete-ir "$id" --from <authority> --yes
done

# Re-import from genesis
./scripts/ir_manager.sh
# Select option 6, import from ir_genesis_300.json
```

---

## API Integration

### REST API Endpoints

```bash
# List IRs
GET /aura/inclusionroutines/v1beta1/irs

# Get specific IR
GET /aura/inclusionroutines/v1beta1/irs/{id}

# Get IR graph
GET /aura/inclusionroutines/v1beta1/irs/{id}/graph

# Get rate limits
GET /aura/inclusionroutines/v1beta1/irs/{id}/rate-limit
```

### Frontend Integration Example

```javascript
// Fetch all IRs
const response = await fetch('http://localhost:1317/aura/inclusionroutines/v1beta1/irs');
const { irs } = await response.json();

// Filter IRs by trinity category
const docIRs = irs.filter(ir => ir.trinity_category === 'official_document');
const bioIRs = irs.filter(ir => ir.trinity_category === 'biometric');
const activityIRs = irs.filter(ir => ir.trinity_category === 'witnessed_activity');

// Calculate user's progress
const userCompletedIRs = getUserCompletedIRs(); // Your function
const totalPoints = userCompletedIRs.reduce((sum, ir) => sum + ir.score, 0);
const hasDoc = userCompletedIRs.some(ir => ir.trinity_category === 'official_document');
const hasBio = userCompletedIRs.some(ir => ir.trinity_category === 'biometric');
const hasActivity = userCompletedIRs.some(ir => ir.trinity_category === 'witnessed_activity');

const isVerified = totalPoints >= 100 && hasDoc && hasBio && hasActivity;
```

---

## Best Practices

### ✅ Do's

- **Version IRs** - Increment version when making changes
- **Test thoroughly** - Validate IR definitions before activation
- **Document clearly** - Write detailed descriptions for users
- **Consider accessibility** - Ensure IRs are achievable globally
- **Balance fun and security** - Mix engaging and rigorous IRs
- **Monitor fraud** - Suspend IRs showing abuse patterns

### ❌ Don'ts

- **Don't skip trinity** - Every path must have doc+bio+activity
- **Don't exceed 30 points** - Prevents single-method verification
- **Don't go below 10 points** - Ensures max 10 IRs needed
- **Don't create impossible IRs** - Make sure they're achievable
- **Don't ignore privacy** - Set appropriate privacy tiers
- **Don't forget locale** - Consider international users

---

## Support & Community

### Getting Help

- **Documentation**: `/docs/modules/inclusionroutines/`
- **Discord**: #ir-development channel
- **Governance Forum**: forum.aura.network
- **GitHub Issues**: github.com/aequitas/aura/issues

### Contributing New IRs

1. Propose in governance forum
2. Gather community feedback
3. Submit formal governance proposal
4. Vote passes → IR added to system
5. Monitor usage and fraud patterns

---

## Quick Reference

### IR Point Values

| Points | Use Case | Example |
|--------|----------|---------|
| 30 | Highest-value official docs | Passport, National ID |
| 25 | High-value docs/biometrics | Driver's License, Iris Scan |
| 20 | Government portal logins | IRS.gov, SSA.gov |
| 15 | Supporting docs, biometrics | Birth cert, Face match |
| 12 | Social, spontaneous | Attestations, timed challenges |
| 10 | Fun, low-friction | Games, creative challenges |

### Arena Summary

| Arena | Point Range | Trinity Category | Count |
|-------|-------------|------------------|-------|
| HIGH_ASSURANCE | 15-30 | official_document | 30 |
| BIOMETRIC | 10-25 | biometric | 60 |
| POSSESSION | 15-25 | witnessed_activity | 50 |
| SOCIAL | 10-15 | - | 30 |
| GEOLOCATION | 10-15 | - | 25 |
| PERSISTENCE | 10-15 | - | 20 |
| KNOWLEDGE | 10-12 | - | 15 |
| ANCHOR | 20-30 | multiple | 20 |
| SPECIALIZED | 10-15 | - | 50 |

### Command Cheat Sheet

```bash
# Create
aurad tx inclusionroutines create-ir <id> <name> <arena> <desc> <score> <poi> <locales> <privacy> <version> "" 0 0

# Update
aurad tx inclusionroutines update-ir <id> --name=<name> --score=<score>

# Delete
aurad tx inclusionroutines delete-ir <id>

# Query
aurad query inclusionroutines ir <id>
aurad query inclusionroutines list-irs

# Manage
./scripts/ir_manager.sh
```

---

**Happy IR Management! 🎉**

*Remember: The IR system is designed to be flexible and community-driven. Don't be afraid to propose creative new IRs that make identity verification more accessible, secure, and fun!*
