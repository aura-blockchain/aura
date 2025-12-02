---
status: ready
priority: p1
issue_id: "005"
tags: [code-review, data-integrity, genesis]
dependencies: []
---

# Genesis Export Silently Skips Corrupted Records

## Problem Statement

In `ExportGenesis`, when iterating user records, unmarshal errors are silently ignored with `continue`. This means corrupted or legacy-format records are simply dropped from the export without any indication.

**Why it matters:** During a chain upgrade, if even ONE record is corrupted, it will be permanently lost. Users could lose their earned reputation/confidence scores with no audit trail.

## Findings

### Evidence
- **File:** `chain/x/confidencescore/keeper/genesis.go`
- **Lines:** 81-140

```go
iterator := prefixStore.Iterator(nil, nil)
defer iterator.Close()

for ; iterator.Valid(); iterator.Next() {
    var record types.UserConfidenceRecord
    if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
        // ERROR IS SILENTLY IGNORED!
        continue
    }
    // ...
}
```

### Data Loss Scenario
1. Chain has 1000 user confidence records
2. 1 record corrupted (disk error, bad migration, proto schema change)
3. Export runs, skips corrupted record silently
4. Export contains 999 records
5. Import to new chain = permanent loss of 1 user's data
6. User's confidence score reset to 0 = loss of earned reputation

### Impact
- Silent data loss during chain upgrades
- No audit trail of what was lost
- User reputation/scores permanently destroyed
- Potential regulatory issues (data integrity requirements)

## Proposed Solutions

### Option A: Log and Continue with Visibility (Recommended)
**Pros:** Doesn't block export, provides visibility
**Cons:** Still loses data, but operator is aware
**Effort:** Small (1 hour)
**Risk:** Low

```go
var exportErrors []string
for ; iterator.Valid(); iterator.Next() {
    var record types.UserConfidenceRecord
    if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
        addr := string(iterator.Key())
        exportErrors = append(exportErrors, fmt.Sprintf("record %s: %v", addr, err))
        ctx.Logger().Error("failed to unmarshal record during export",
            "key", addr, "error", err)
        continue
    }
    protoRecords = append(protoRecords, &pb.UserConfidenceRecord{...})
}

if len(exportErrors) > 0 {
    ctx.Logger().Error("genesis export completed with errors",
        "failed_records", len(exportErrors),
        "errors", exportErrors)
    // Emit event for monitoring
    ctx.EventManager().EmitEvent(
        sdk.NewEvent("genesis_export_warning",
            sdk.NewAttribute("module", types.ModuleName),
            sdk.NewAttribute("failed_count", strconv.Itoa(len(exportErrors))),
        ),
    )
}
```

### Option B: Fail Export on Any Error
**Pros:** No silent data loss
**Cons:** Blocks upgrade if any corruption exists
**Effort:** Small (30 min)
**Risk:** Medium (could block upgrades)

## Recommended Action
Implement Option A. Apply same pattern to all modules with genesis export.

## Technical Details

### Affected Files
- `chain/x/confidencescore/keeper/genesis.go`
- All other module genesis.go files (same pattern likely exists)

### Acceptance Criteria
- [ ] Export errors are logged with record identifiers
- [ ] Summary logged at end of export with count of failures
- [ ] Event emitted for monitoring systems
- [ ] Same pattern applied to all modules
- [ ] Test verifies error logging with corrupted record

## Work Log

| Date | Action | Learnings |
|------|--------|-----------|
| 2025-12-02 | Finding identified in data integrity review | Pattern exists in multiple modules |

## Resources
- Related: P1-006 (Governance genesis validation), P2-003 (Store key management)
