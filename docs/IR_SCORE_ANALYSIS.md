# Inclusion Routine Score Statistical Analysis

## Executive Summary

**CRITICAL FINDING**: The VerifiedHuman VC policy threshold of CS 50 is **200x lower** than the system's designed verification threshold of 10,000, creating a severe Sybil attack vulnerability.

## Key Statistics

### Overall IR Score Distribution
- **Total IRs**: 180 (excluding IR-000 anchor)
- **Score Range**: 50 - 2,500 points
- **Mean**: 581.67 points
- **Median**: 450 points
- **Total Possible**: 104,700 points

### Score Distribution
| Range | Count | Percentage |
|-------|-------|------------|
| 0-100 | 10 | 5.6% |
| 101-300 | 53 | 29.4% |
| 301-500 | 43 | 23.9% |
| 501-700 | 27 | 15.0% |
| 701-1000 | 24 | 13.3% |
| 1001-1500 | 18 | 10.0% |
| 1501-3000 | 5 | 2.8% |

## Arena Breakdown

| Arena | Count | Min | Max | Mean | Median | Total |
|-------|-------|-----|-----|------|--------|-------|
| Biometric | 20 | 50 | 700 | 317.50 | 300 | 6,350 |
| Possession | 25 | 50 | 800 | 390.00 | 400 | 9,750 |
| Knowledge | 30 | 50 | 1,500 | 430.00 | 350 | 12,900 |
| Geolocation | 25 | 100 | 1,000 | 438.00 | 400 | 10,950 |
| Social | 20 | 100 | 1,500 | 547.50 | 500 | 10,950 |
| Specialized | 25 | 100 | 1,500 | 682.00 | 600 | 17,050 |
| High-Assurance | 20 | 500 | 2,000 | 1,065.00 | 1,050 | 21,300 |
| Persistence | 15 | 250 | 2,500 | 1,030.00 | 900 | 15,450 |

## Threshold Analysis (Without Multipliers)

### CS Threshold Requirements

| Threshold | Min IRs | Max IRs | Avg IRs | Arena Diversity |
|-----------|---------|---------|---------|-----------------|
| 50 | 1 | 1 | 0.1 | 11.1% (1 arena) |
| 100 | 1 | 2 | 0.2 | 22.2% (2 arenas) |
| 500 | 1 | 7 | 1.1 | 44.4% (4 arenas) |
| 1,000 | 1 | 11 | 2.2 | 66.7% (6 arenas) |
| 5,000 | 3 | 32 | 11.1 | 66.7% (6 arenas) |
| **10,000** | **6** | **51** | **22.2** | **77.8% (7 arenas)** |

## Multiplier Impact

To reach CS 10,000 with bonuses:

| Scenario | IRs Needed | Avg Points/IR |
|----------|------------|---------------|
| Base (1.0x) | 22.2 | 450 |
| Velocity bonus (1.25x) | 17.8 | 562 |
| Arena bonus (1.5x) | 14.8 | 675 |
| Both bonuses (1.875x) | 11.9 | 843 |
| Jackpot 5x | 4.4 | 2,250 |
| Jackpot 25x | 0.9 | 11,250 |

## Sybil Resistance Analysis

### CS 50 (Current VerifiedHuman Threshold)
- ✗ **Only 1 IR required** (IR-101: Simple Liveness)
- ✗ **Only 1 arena needed** (11.1% diversity)
- ✗ **Zero prerequisite diversity**
- ✗ **Trivially automatable**
- ✗ **No cost to create fake identities**

**Attack Vector**: An attacker with basic AI/ML skills could automate IR-101 (Simple Liveness) and create thousands of "verified" identities per day.

### CS 10,000 (Recommended Threshold)
- ✓ **6-51 IRs required** (avg 22 without bonuses, 12-18 with bonuses)
- ✓ **7-8 arenas needed** (77.8% diversity)
- ✓ **Requires significant time investment** (Persistence IRs span weeks/months)
- ✓ **Requires real-world assets** (IDs, bills, physical items)
- ✓ **Requires social connections** (Social arena vouching)
- ✓ **Strong Sybil resistance**

**Example Realistic Path to 10,000 CS:**
1. IR-000: Government ID Anchor (0 pts, required)
2. IR-101: Simple Liveness (50 pts)
3. IR-303: SMS Verification (100 pts)
4. IR-502: Home Check-in (100 pts)
5. IR-213: House Key (50 pts)
6. IR-401: Peer Vouch Level 1 (100 pts)
7. IR-305: Bank Account Digital (700 pts)
8. IR-315: KYC Pass-Thru Exchange (800 pts)
9. IR-304: Utility Bill Digital (500 pts)
10. IR-203: Utility Bill Physical (400 pts)
11. IR-501: Mailbox Quest (500 pts)
12. IR-327: Paystub (600 pts)
13. IR-701: Daily Check-in 7 days (300 pts)
14. IR-703: Proof-of-Life 30d (500 pts)
15. IR-102: Randomized Pose (300 pts)
16. IR-104: Voiceprint Static (150 pts)
17. IR-201: Credit/Debit Card (300 pts)
18. IR-310: Photo History Quest (400 pts)
19. IR-311: Device History (300 pts)
20. IR-506: Public Landmark (400 pts)
21. IR-605: Voter Registration (900 pts)
22. IR-602: Bank Letter (1,200 pts)
23. Additional IRs to reach 10,000+

**With velocity/arena bonuses**: Achievable with 12-18 diverse IRs

## Critical Security Issue

### The Problem
The VerifiedHuman credential currently requires **CS 50**, which is:
- **200x lower** than the system's verification_threshold (10,000)
- Achievable with **1 single IR** (Simple Liveness check)
- Provides **no Sybil resistance**
- **Contradicts the entire IR system design**

### Impact
1. **Governance Attack**: Malicious actor creates 10,000 fake "verified" identities
2. **VC System Devaluation**: "Verified" status becomes meaningless
3. **Economic Exploit**: Fake identities claim airdrops, rewards, voting power
4. **Trust Erosion**: Users lose faith in the verification system

### Root Cause
The VC policy definitions were likely created with a misunderstanding of the CS scale:
- Assumed CS was on a 0-100 scale (like a percentage)
- Actually CS is on a 0-10,000+ scale (point accumulation)

## Recommendations

### 1. **CRITICAL: Fix VerifiedHuman Threshold**
```go
{
    VcTypeName:           "VerifiedHuman",
    VcTypeEnum:           VCTypeVerifiedHuman,
    CsThreshold:          10000,  // Changed from 50
    // ... rest of policy
}
```

### 2. **Adjust Other VC Policy Thresholds**
All VC policies should be reviewed and adjusted to the proper scale:

| Credential | Current CS | Recommended CS | Rationale |
|------------|-----------|----------------|-----------|
| VerifiedHuman | 50 | 10,000 | Base verification threshold |
| AgeOver18 | 60 | 10,000 | Same as VerifiedHuman |
| AgeOver21 | 60 | 10,000 | Same as VerifiedHuman |
| ResidentOf | 70 | 12,000 | Requires geo IRs |
| BiometricAuth | 80 | 13,000 | Requires biometric focus |
| KYCVerification | 90 | 15,000 | High-assurance threshold |
| NotaryPublic | 95 | 18,000 | Specialized + high-assurance |
| ProfessionalLicense | 90 | 15,000 | Specialized arena |
| BiometricFocus | 75 | 5,000 in Biometric | Arena focus |
| SocialFocus | 70 | 5,000 in Social | Arena focus |
| GeolocationFocus | 70 | 5,000 in Geolocation | Arena focus |
| HighAssuranceFocus | 95 | 5,000 in High-Assurance | Arena focus |
| PossessionFocus | 65 | 5,000 in Possession | Arena focus |
| KnowledgeFocus | 70 | 5,000 in Knowledge | Arena focus |
| PersistenceFocus | 60 | 5,000 in Persistence | Arena focus |
| SpecializedFocus | 85 | 5,000 in Specialized | Arena focus |

### 3. **Add Threshold Constants**
Define threshold constants to prevent future confusion:

```go
const (
    // Core thresholds from CS module
    CSVerificationThreshold     = 10000  // Basic verification
    CSHighAssuranceThreshold    = 15000  // High-trust credentials
    CSArenaFocusThreshold       = 5000   // Per-arena focus bonus

    // VC-specific thresholds
    CSVerifiedHuman             = CSVerificationThreshold
    CSSpecializedCredential     = 15000
    CSProfessionalCredential    = 18000
)
```

## Conclusion

The statistical analysis confirms that:
1. **CS 50 is dangerously low** and provides no Sybil resistance
2. **CS 10,000 is the appropriate threshold** based on:
   - System design (verification_threshold parameter)
   - IR score distribution (requires 12-22 diverse IRs)
   - Arena diversity requirements (7-8 arenas)
   - Time/effort investment (weeks to months)
3. **Immediate action required** to fix VC policy thresholds before mainnet launch

---
*Analysis performed: 2025-11-13*
*IR Definitions version: 1.0 (181 total IRs)*
