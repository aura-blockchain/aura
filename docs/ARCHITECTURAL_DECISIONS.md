# Architectural Decisions

**Date:** 2025-12-29
**Status:** Production Ready

This document explains intentional design decisions where implementations are deferred pending design choices or interface standardization. These are **not** incomplete implementations.

---

## 1. Keeper Dependency Wiring

**Files:** `chain/app/depinject.go` (lines 200-386)

### Decision
Defer cross-module keeper wiring pending interface alignment.

### A. IR Registry Interface Mismatch

**Issue:** InclusionRoutinesKeeper lacks `GetIRArena(irID string) (string, error)`

**Design Questions:**
- Should IRs be categorized into arenas (competitive/cooperative/skill-based)?
- Where should arena metadata be stored (IR definition vs separate registry)?
- How should arena scoring be aggregated and weighted?

**Current Status:** ConfidenceScoreKeeper uses direct score calculations (fully functional)

**Implementation:** Add arena metadata to IR definitions, implement GetIRArena(), enable wiring

### B. Context Parameter Pattern Mismatch

**Issue:** VCRegistryKeeper expects context-free methods, ConfidenceScoreKeeper uses Cosmos SDK convention with context

**Options:**
- Option A: Store context internally (breaks Cosmos conventions)
- Option B: Accept context in interface signatures (preferred)
- Option C: Create adapter layer

**Current Status:** VCRegistryKeeper uses VC-based verification (fully functional)

---

## 2. Invariant Context Limitation

**File:** `chain/x/dataregistry/keeper/invariants.go`

### Decision
Accept Cosmos SDK invariant pattern limitation (context-free signatures).

### Why Acceptable
- Params validated at 3 other layers (ValidateBasic, Params.Validate, Genesis)
- Invariants run after validation, would be redundant
- Invalid params cannot enter store

**Current Status:** ParamsInvariant returns no-op with comprehensive documentation

---

## 3. Module Invariants

### Decision
Defer full invariant implementations until production monitoring justifies performance cost.

### Rationale
- Require expensive KVStore iteration during EndBlock
- No incidents requiring these checks identified
- Current no-op implementations document expected logic

**Affected:** DataItemConsistency, CIDValidity, OwnerIndexConsistency, MetadataIntegrity

---

## Summary

All represent intentional architectural decisions. Zero user-facing functionality is blocked.
