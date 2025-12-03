---
id: "065"
title: "Compliance Tax Report Path Traversal"
status: ready
priority: p2
category: security
module: compliance
severity: CRITICAL
source: compliance-audit
---

# Compliance Tax Report File Path Traversal

## Problem

The file_path field in tax reports is not validated, enabling path traversal attacks. Attacker can write to arbitrary paths.

## Affected Files

- `chain/x/compliance/keeper/msg_server.go:190-213`
- `proto/aura/compliance/v1beta1/compliance.proto:187`

## Impact

- Arbitrary file access
- System compromise potential
- Sensitive file exposure

## Required Fix

```go
func ValidateFilePath(path string) error {
    if filepath.IsAbs(path) {
        return fmt.Errorf("absolute paths not allowed")
    }
    if strings.Contains(path, "..") {
        return fmt.Errorf("path traversal not allowed")
    }
    return nil
}
```

## Acceptance Criteria

- [ ] Path validation implemented
- [ ] Absolute paths rejected
- [ ] Path traversal rejected
- [ ] Allowed directory enforcement
- [ ] Tests for path attacks
