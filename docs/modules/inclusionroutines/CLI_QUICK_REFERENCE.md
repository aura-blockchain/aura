# Inclusion Routines CLI Quick Reference

## Quick Access
```bash
# Full command
aurad tx inclusionroutines [command]
aurad query inclusionroutines [command]

# Short alias
aurad tx ir [command]
aurad query ir [command]
```

## Transaction Commands

### 1. Create IR
```bash
aurad tx ir create-ir [id] [name] [arena] [description] [score] [poi-reward] \
  --locale-tags "US,UK,EU" \
  --privacy-tier 3 \
  --version "1.0.0" \
  --metadata-hash "0xabc..." \
  --activation-height 1000 \
  --sunset-height 100000 \
  --from [key]
```

**Arenas:** 0=Unspecified, 1=Anchor, 2=Biometric, 3=Possession, 4=Knowledge, 5=Social, 6=Geolocation, 7=HighAssurance, 8=Persistence, 9=Specialized

**Privacy Tiers:** 0=Unspecified, 1=Low, 2=Medium, 3=High

### 2. Update IR
```bash
aurad tx ir update-ir [id] \
  --name "Updated Name" \
  --description "New description" \
  --score 150 \
  --poi-reward 75 \
  --from [key]
```

### 3. Delete IR
```bash
aurad tx ir delete-ir [id] --from [key]
```

### 4. Set Prerequisites
```bash
# Single prerequisite
aurad tx ir set-prerequisites [ir-id] [required-ir-id] --from [key]

# Multiple prerequisites
aurad tx ir set-prerequisites [ir-id] "req-id-1,req-id-2,req-id-3" --from [key]

# Clear prerequisites
aurad tx ir set-prerequisites [ir-id] "" --from [key]
```

### 5. Set Rate Limits
```bash
aurad tx ir set-rate-limit [ir-id] [per-hour] [per-day] [per-block] --from [key]

# Examples:
# Strict: 3/hour, 10/day, 100/block
aurad tx ir set-rate-limit "high-value-ir" 3 10 100 --from governance

# Relaxed: 100/hour, 1000/day, 10000/block
aurad tx ir set-rate-limit "low-value-ir" 100 1000 10000 --from governance

# Unlimited (use 0)
aurad tx ir set-rate-limit "unlimited-ir" 0 0 0 --from governance
```

### 6. Suspend IR
```bash
aurad tx ir suspend-ir [ir-id] "Reason for suspension" --from [key]
```

### 7. Activate IR
```bash
aurad tx ir activate-ir [ir-id] --from [key]
```

## Query Commands

### 1. Show IR Details
```bash
aurad query ir show [ir-id]

# Example
aurad query ir show "gov-id-verify"
```

### 2. List IRs
```bash
# List all
aurad query ir list

# Filter by status (1=Draft, 2=Reviewing, 3=Approved, 4=Active, 5=Suspended, 6=Deprecated, 7=Retired)
aurad query ir list --status 4

# Filter by arena
aurad query ir list --arena 2

# Filter by locale
aurad query ir list --locale "US"

# Combined filters
aurad query ir list --status 4 --arena 1 --locale "GLOBAL"

# With pagination
aurad query ir list --page 2 --limit 50
```

### 3. Query Dependency Graph
```bash
aurad query ir graph [ir-id]

# Example - see what must be completed before and what becomes available after
aurad query ir graph "advanced-biometric"
```

### 4. Query Rate Limits
```bash
aurad query ir rate-limit [ir-id]

# Example
aurad query ir rate-limit "gov-id-verify"
```

### 5. Query Module Parameters
```bash
aurad query ir params
```

## Common Workflows

### Workflow 1: Create Basic → Advanced IR Chain
```bash
# 1. Create basic IR
aurad tx ir create-ir "basic-verify" "Basic Verification" 1 \
  "Simple identity check" 50 25 \
  --privacy-tier 2 --from governance

# 2. Create advanced IR
aurad tx ir create-ir "advanced-verify" "Advanced Verification" 7 \
  "High assurance identity check" 200 100 \
  --privacy-tier 3 --from governance

# 3. Set basic as prerequisite for advanced
aurad tx ir set-prerequisites "advanced-verify" "basic-verify" --from governance

# 4. Verify the dependency
aurad query ir graph "advanced-verify"
```

### Workflow 2: Emergency IR Suspension
```bash
# 1. Suspend immediately
aurad tx ir suspend-ir "compromised-ir" "Security vulnerability CVE-2024-XXXX" \
  --from emergency-admin

# 2. Verify suspension
aurad query ir show "compromised-ir"

# 3. After fix, reactivate
aurad tx ir activate-ir "compromised-ir" --from governance
```

### Workflow 3: Tiered Verification System
```bash
# Tier 1: Basic (ANCHOR)
aurad tx ir create-ir "tier1-email" "Email Verification" 1 \
  "Verify email ownership" 10 5 --privacy-tier 1 --from governance

# Tier 2: Intermediate (BIOMETRIC)
aurad tx ir create-ir "tier2-face" "Face Verification" 2 \
  "Facial recognition" 50 25 --privacy-tier 2 --from governance
aurad tx ir set-prerequisites "tier2-face" "tier1-email" --from governance

# Tier 3: Advanced (HIGH_ASSURANCE)
aurad tx ir create-ir "tier3-gov-id" "Government ID" 7 \
  "Gov ID + liveness" 200 100 --privacy-tier 3 --from governance
aurad tx ir set-prerequisites "tier3-gov-id" "tier2-face,tier1-email" --from governance

# Query the complete graph
aurad query ir graph "tier3-gov-id"
```

### Workflow 4: Geographic IR Management
```bash
# Create US-specific IR
aurad tx ir create-ir "us-ssn-verify" "SSN Verification" 1 \
  "US Social Security Number verification" 100 50 \
  --locale-tags "US" --privacy-tier 3 --from governance

# Create EU-specific IR
aurad tx ir create-ir "eu-gdpr-verify" "GDPR Compliant ID" 1 \
  "EU GDPR-compliant ID verification" 100 50 \
  --locale-tags "EU,UK" --privacy-tier 3 --from governance

# List by region
aurad query ir list --locale "US"
aurad query ir list --locale "EU"
```

## Arena Reference

| Arena | Value | Description | Use Cases |
|-------|-------|-------------|-----------|
| ANCHOR | 1 | Core identity | Government ID, email, phone |
| BIOMETRIC | 2 | Biometric data | Face, fingerprint, voice, iris |
| POSSESSION | 3 | Asset ownership | Device, wallet, physical key |
| KNOWLEDGE | 4 | Knowledge-based | Security questions, passwords |
| SOCIAL | 5 | Social proof | Social media, reputation |
| GEOLOCATION | 6 | Location-based | GPS, IP address verification |
| HIGH_ASSURANCE | 7 | High security | Multi-factor, notarized documents |
| PERSISTENCE | 8 | Ongoing verification | Periodic re-verification |
| SPECIALIZED | 9 | Custom tasks | Industry-specific verification |

## Privacy Tier Guidelines

| Tier | Value | Data Sensitivity | Examples |
|------|-------|------------------|----------|
| LOW | 1 | Public data | Email, username |
| MEDIUM | 2 | Semi-private | Phone number, address |
| HIGH | 3 | Highly sensitive | Biometrics, government ID, SSN |

## Status Lifecycle

```
DRAFT (1) → REVIEWING (2) → APPROVED (3) → ACTIVE (4)
                                              ↓
                                         SUSPENDED (5) ← → ACTIVE (4)
                                              ↓
                                        DEPRECATED (6) → RETIRED (7)
```

## Rate Limit Best Practices

| IR Type | Per Hour | Per Day | Per Block | Rationale |
|---------|----------|---------|-----------|-----------|
| High-value (Gov ID) | 3 | 10 | 100 | Prevent abuse, expensive verification |
| Medium-value (Biometric) | 10 | 50 | 500 | Balance access and security |
| Low-value (Email) | 50 | 200 | 2000 | Allow retry, cheap verification |
| Test/Dev | 100 | 1000 | 10000 | Development flexibility |

## Common Flags

### Transaction Flags
- `--from [key]` - Key to sign transaction (required)
- `--chain-id [id]` - Chain identifier
- `--fees [amount]` - Transaction fees
- `--gas [amount]` - Gas limit
- `--gas-prices [prices]` - Gas prices
- `-y` - Skip confirmation prompt
- `-b [mode]` - Broadcast mode (sync|async|block)

### Query Flags
- `--height [height]` - Query at specific height
- `--output [format]` - Output format (text|json)
- `--page [num]` - Page number for pagination
- `--limit [num]` - Results per page
- `--count-total` - Count total results

## Tips

1. **Always check current state before updates**
   ```bash
   aurad query ir show [ir-id]
   ```

2. **Use dry-run for testing**
   ```bash
   aurad tx ir create-ir ... --dry-run
   ```

3. **Check prerequisites before creating dependencies**
   ```bash
   aurad query ir show "prerequisite-ir-id"
   ```

4. **Monitor rate limits**
   ```bash
   aurad query ir rate-limit [ir-id]
   ```

5. **Use meaningful IR IDs**
   - Good: `gov-id-verify-v2`, `face-recognition-liveness`
   - Bad: `ir1`, `test`, `abc123`

6. **Document suspension reasons**
   - Include ticket numbers, CVE IDs, or investigation links

7. **Test in development first**
   - Use `--chain-id aura-testnet` for testing
   - Verify on testnet before mainnet deployment

## Error Handling

Common errors and solutions:

**"IR not found"**
```bash
# Verify ID is correct
aurad query ir list | grep [partial-id]
```

**"Insufficient permissions"**
```bash
# Ensure using governance or authorized key
aurad tx ir create-ir ... --from governance
```

**"Prerequisite not found"**
```bash
# Check prerequisite exists
aurad query ir show [prerequisite-id]
```

**"Rate limit exceeded"**
```bash
# Check current limits
aurad query ir rate-limit [ir-id]
# Increase if appropriate
aurad tx ir set-rate-limit [ir-id] [higher-limits] --from governance
```

## Monitoring Commands

```bash
# List all active IRs
aurad query ir list --status 4

# List suspended IRs (need attention)
aurad query ir list --status 5

# Check specific IR health
aurad query ir show [ir-id]
aurad query ir rate-limit [ir-id]
aurad query ir graph [ir-id]

# Monitor module parameters
aurad query ir params
```

---

**Note:** Replace `[key]` with your actual key name or address. Most IR management commands require governance or admin privileges.
