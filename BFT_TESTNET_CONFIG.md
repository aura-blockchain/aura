# AURA BFT Testnet Configuration

**Date:** 2025-12-14 | **Chain ID:** aura-local-4 | **Status:** ✅ Operational

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

| Validator | RPC | REST API | P2P | gRPC | Metrics |
|-----------|-----|----------|-----|------|---------|
| validator-1 | 27657 | 2317 | 27656 | 10090 | 27660 |
| validator-2 | 27757 | 2417 | 27756 | 10190 | 27760 |
| validator-3 | 27857 | 2517 | 27856 | 10290 | 27860 |
| validator-4 | 27957 | 2617 | 27956 | 10390 | 27960 |

## BFT Consensus Verification Results

### ✅ Test 1: All 4 Validators (100% power)
- **Result:** PASS - Normal consensus, 4/4 signatures per block, 3s block time

### ✅ Test 2: 3 Validators (75% power)
- **Test:** Stopped validator-4
- **Active Power:** 2,700,000 (75% > 67% threshold)
- **Result:** PASS - Consensus maintained (height 977→981 in 10s)

### ✅ Test 3: 2 Validators (50% power)
- **Test:** Stopped validator-3 and validator-4
- **Active Power:** 1,800,000 (50% < 67% threshold)
- **Result:** PASS - Consensus properly halted at height 987

### ✅ Test 4: Recovery
- **Test:** Restarted validator-3 and validator-4
- **Result:** PASS - Consensus resumed immediately (height 987→990)

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

✅ **BFT Consensus Verified:** 4-validator testnet correctly implements BFT consensus
with >2/3 voting power requirement. Network tolerates 1 validator failure (F=1, N=3F+1).
All validators participate equally with 25% voting power each.
