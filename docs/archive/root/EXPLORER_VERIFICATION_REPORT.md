# Aura Block Explorer Verification Report
**Date:** 2025-12-14
**Chain:** aura-local-4
**Explorer URL:** http://localhost:8088
**Container:** aura-block-explorer (aura-explorer image)

## Executive Summary
The explorer is **ACCESSIBLE and FUNCTIONAL** for core features (blocks, transactions, validators), but has two notable issues: healthcheck failures and non-functional REST API endpoints.

## 1. Accessibility Status
- **HTTP Accessibility:** PASS (HTTP 200)
- **Frontend Load:** PASS (HTML + assets served)
- **Container Status:** RUNNING (21+ hours uptime)
- **Health Status:** UNHEALTHY (healthcheck failing)

## 2. Core Feature Verification

### Block Viewer - PASS
- Recent blocks visible via RPC
- Block height: 987+ (actively producing)
- Block metadata includes: hash, height, time, proposer, tx count
- Range queries working (tested blocks 1-5, 950-964)
- Genesis block accessible (height=1)

### Transaction Viewer - PARTIAL
- Transaction search by height: WORKING
- Recent tx query successful (tested height>900)
- Transaction detail lookup by hash: WORKING (tested tx 0x14D13FF3...)
- No transactions found via module search (chain has minimal tx activity)

### Address Lookup - FAIL
- REST API endpoints return HTTP 502 Bad Gateway
- Example failure: /api/cosmos/bank/v1beta1/balances/{address}
- Root cause: API server bound to localhost:1317 instead of 0.0.0.0:1317

### Search Functionality - PASS
- RPC search working (tx_search, block queries)
- Validators list accessible (4 validators visible)
- Network info available (3 connected peers)

### Validator List - PASS
- All 4 validators visible
- Addresses displayed correctly
- Validator metadata accessible

## 3. Aura-Specific Message Decoding
- **Status:** UNKNOWN (no Aura module transactions to test)
- Searched for identity module txs: 0 results
- Chain has empty blocks (no user transactions yet)
- **Recommendation:** Create test transactions to verify decoding

## 4. Container Health Status

### Health Check Configuration
```yaml
test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost/"]
interval: 30s
timeout: 10s
retries: 3
start_period: 40s
```

### Issue Identified
- Healthcheck returns: "Connection refused"
- Root cause: DNS resolution - `localhost` fails, `127.0.0.1` succeeds
- Container /etc/hosts may be missing localhost entry
- **Impact:** Container marked unhealthy despite being fully functional

## 5. Critical Errors Found

### Error 1: REST API Not Accessible (HIGH PRIORITY)
**Location:** validator-1:/home/aura/.aura/config/app.toml
**Issue:** API bound to localhost instead of 0.0.0.0
```toml
[api]
enable = true
address = "tcp://localhost:1317"  # WRONG - not accessible from other containers
```
**Fix Required:**
```toml
address = "tcp://0.0.0.0:1317"    # Correct - accessible from docker network
```
**Impact:** All /api/* endpoints return 502 Bad Gateway

### Error 2: Healthcheck DNS Resolution (LOW PRIORITY)
**Location:** docker-compose.explorer.yml
**Issue:** Healthcheck uses `localhost` which fails DNS resolution
**Fix Required:**
```yaml
test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://127.0.0.1/"]
```
**Impact:** Container incorrectly marked as unhealthy

## 6. Network Configuration

### Explorer Configuration
- Nginx proxy: WORKING
- RPC endpoint: http://aura-validator-1:26657 (accessible, working)
- API endpoint: http://aura-validator-1:1317 (not accessible - see Error 1)
- Network: aura-testnet (172.26.0.0/16)
- Explorer IP: Not assigned (dynamic)
- Validator-1 IP: 172.26.0.10

### Port Mappings
- Explorer: 8088:80 (host:container)
- Validator-1 RPC: 27657:26657 (working)
- Validator-1 API: 2317:1317 (port mapped but service unavailable)

## 7. Recommendations

### Immediate Fixes Required
1. **Fix API bind address** in validator app.toml (Error 1)
2. **Update healthcheck** to use 127.0.0.1 instead of localhost (Error 2)
3. **Restart validator-1** after config change to enable API
4. **Create test transactions** to verify Aura-specific message decoding

### Testing Plan
After fixes:
1. Verify /api/cosmos/bank/v1beta1/balances/{address} returns data
2. Test balance queries via explorer UI
3. Submit identity module transaction
4. Verify explorer decodes Aura-specific message types
5. Confirm healthcheck passes

## 8. Summary
**Explorer Status:** Functional but degraded
**Blockers:** REST API unavailable (blocks address lookups, balance queries)
**User Impact:** Basic exploration works (blocks, txs, validators) but address/balance features broken
**Estimated Fix Time:** 5 minutes (config change + container restart)
