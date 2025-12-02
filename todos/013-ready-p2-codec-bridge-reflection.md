---
status: ready
priority: p2
issue_id: "013"
tags: [code-review, architecture, performance]
dependencies: []
---

# Codec Bridge Uses Runtime Reflection with Silent Panic Recovery

## Problem Statement

The `chain/app/codec_bridge.go` uses runtime reflection to bridge modern protobuf to legacy gogo registry, with silent panic recovery that masks registration failures.

**Why it matters:** Registration failures are silently ignored, which could cause runtime panics later. Runtime reflection adds overhead and complexity.

## Findings

### Evidence
- **File:** `chain/app/codec_bridge.go`
- **Lines:** 20-84

```go
func init() {
    protoregistry.GlobalFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
        // ... complex reflection-based registration ...
        func() {
            defer func() {
                // Ignore registration failures to avoid panics during CLI startup.
                _ = recover()  // SILENT FAILURE!
            }()
            gogoproto.RegisterFile(path, buf.Bytes())
        }()
        return true
    })
}
```

### Impact
- Registration failures silently ignored
- May cause msgservice panics later in execution
- Runtime overhead from reflection
- Difficult to debug registration issues

## Proposed Solutions

### Option A: Add Logging for Visibility (Short-term)
**Effort:** Small (1 hour)

### Option B: Fix Proto Generation (Long-term)
**Effort:** Large (investigate buf.build configuration)

## Technical Details

### Affected Files
- `chain/app/codec_bridge.go`

### Acceptance Criteria
- [ ] Registration failures logged with file names
- [ ] Summary count of registered/failed at startup
- [ ] Plan to remove bridge when SDK supports modern protobuf

## Work Log

| Date | Action | Learnings |
|------|--------|-----------|
| 2025-12-02 | Finding identified in architecture review | Technical debt from SDK migration |
