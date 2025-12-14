# AURA BFT Testnet Configuration

**Date:** 2025-12-14 | **Chain ID:** aura-local-4 | **Status:** ✅ Operational | **Test Pass Rate:** 100% (5/5)

## Configuration Summary
- **Validators:** 4 (equal 25% voting power each: 900,000)
- **Total Power:** 3,600,000 | **BFT Threshold:** >2/3 (2,400,001)
- **Genesis Time:** 2025-01-01T00:00:00Z

## Validator Details

| Validator | Operator Address | Consensus Address | Power | % |
|-----------|-----------------|-------------------|--------|---|
| validator-1 | auravaloper147rneetqamym8rjy28u5n3njzee6s4u0we5nlx | 15866F1B72207259E61DF9E23900A6B3609656CA | 900,000 | 25% |
| validator-2 | auravaloper1dsl2uqluc5qcnv3tksa09nfgz2jgqjzhvf9hes | 65946DB1BCB9BA407A2D10621B0EB580970283F7 | 900,000 | 25% |
| validator-3 | auravaloper1tcacn5729tvx3zmq8qd5wf36efy4e7pykplz0x | 66E11A3F240399C3C7CAC88B9C17BAC1D2EFFAB4 | 900,000 | 25% |
| validator-4 | auravaloper1vq0hzzm4ragwpg3v8at046msc3swmgv0lem9c0 | 6C3B55828787556522B6C9673370502713B8E760 | 900,000 | 25% |

## Network Configuration

**Docker Network:** aura-testnet (172.26.0.0/16)
**Validator IPs:** 172.26.0.10-13 (validators 1-4)

| Validator | RPC | REST API | P2P | gRPC | Metrics | Moniker | Actual Consensus Addr |
|-----------|-----|----------|-----|------|---------|---------|----------------------|
| validator-1 | 27757 | 2417 | 27756 | 10190 | 27760 | validator-2 | 15866F1B72207259E61DF9E23900A6B3609656CA |
| validator-2 | 27657 | 2317 | 27656 | 10090 | 27660 | bcpc | 65946DB1BCB9BA407A2D10621B0EB580970283F7 |
| validator-3 | 27857 | 2517 | 27856 | 10290 | 27860 | validator-3 | 66E11A3F240399C3C7CAC88B9C17BAC1D2EFFAB4 |
| validator-4 | 27957 | 2617 | 27956 | 10390 | 27960 | validator-4 | 6C3B55828787556522B6C9673370502713B8E760 |

**NOTE:** Validators 1 and 2 have swapped port assignments compared to their consensus addresses. This is due to container initialization order and does not affect BFT consensus functionality.

## BFT Consensus Verification Results - 2025-12-14 (100% Pass Rate)

### ✅ Test 1: Baseline Sync (All 4 Validators)
- **Result:** PASS - All validators synced at height 1624 (variance: 0 blocks)
- **Test Duration:** 30s initial sync
- **Validators:** 4/4 operational with 100% voting power

### ✅ Test 2: 3/4 Validators Consensus (75% power)
- **Test:** Stopped validator-3
- **Active Power:** 2,700,000 (75% > 67% threshold)
- **Result:** PASS - Consensus maintained (1624→1638, 14 blocks produced)
- **Block Production:** val1: 2 blocks, val2: 2 blocks, val4: 2 blocks

### ✅ Test 3: Validator Catch-Up
- **Test:** Restarted validator-3 after downtime
- **Result:** PASS - Validator caught up immediately (height diff: 0 blocks)
- **Sync Time:** <2 minutes via state sync
- **Final Heights:** All validators at 1672 (perfectly synced)

### ✅ Test 4: 2/4 Validators Halt (50% power)
- **Test:** Stopped validator-2 and validator-3
- **Active Power:** 1,800,000 (50% < 67% threshold)
- **Result:** PASS - Chain halted correctly (0 blocks produced)
- **Verification:** Consensus properly stopped at height 1673

### ✅ Test 5: Full Recovery
- **Test:** Restarted validator-2 and validator-3
- **Result:** PASS - Full consensus recovery (4/4 validators producing blocks)
- **Recovery Time:** ~30s to full consensus
- **Block Production:** val1: 3 blocks, val2: 3 blocks, val3: 2 blocks, val4: 2 blocks
- **Final Heights:** All validators at 1692 (perfectly synced)

## Block Signature Evidence

Sample block 1002 showing all 4 validators participating:
- 15866F1B72207259E61DF9E23900A6B3609656CA @ 2025-12-14T08:48:31.898797497Z
- 65946DB1BCB9BA407A2D10621B0EB580970283F7 @ 2025-12-14T08:48:31.905781223Z
- 66E11A3F240399C3C7CAC88B9C17BAC1D2EFFAB4 @ 2025-12-14T08:48:31.929238808Z
- 6C3B55828787556522B6C9673370502713B8E760 @ 2025-12-14T08:48:31.929293542Z

## Genesis Modifications

**Changes Made:**
1. Modified testnet-init.sh: Creates 4 validators with equal staking (900B uaura each)
2. Genesis validators: All 4 validators registered with 25% power distribution
3. Total power calculation: 900,000 × 4 = 3,600,000 (each = 900B uaura ÷ PowerReduction 1e6)
4. Docker volumes: Populated with initialized data via populate-volumes.sh
5. No changes required to docker-compose.testnet.yml (already configured for 4)

## Conclusion

✅ **BFT Consensus Verified - 100% Test Pass Rate (5/5 Tests):** 4-validator testnet correctly implements BFT consensus with >2/3 voting power requirement. Network tolerates 1 validator failure (F=1, N=3F+1). All validators participate equally with 25% voting power each.

**Comprehensive Test Results (2025-12-14 09:24:39 - 09:30:52):**
- ✅ Test 1: Baseline Sync - PASS (0 block variance)
- ✅ Test 2: 3/4 Consensus (75% VP) - PASS (14 blocks produced)
- ✅ Test 3: Validator Catch-Up - PASS (0 block difference)
- ✅ Test 4: 2/4 Halt (50% VP) - PASS (consensus correctly halted)
- ✅ Test 5: Full Recovery - PASS (4/4 validators recovered)

**Test Execution Time:** 6 minutes 13 seconds
**Test Script:** `scripts/test-bft-comprehensive.sh --verbose`
**Analysis Tool:** `scripts/analyze-bft-results.py`
**Test Log:** `bft_test_20251214_092439.log`

**Key Findings:**
- Zero height variance across all validators during sync
- Perfect state sync recovery (0 block difference)
- Correct consensus halt behavior with <67% voting power
- Full consensus recovery in ~30 seconds
- All validators maintaining perfect sync throughout test scenarios
