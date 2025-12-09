---
status: pending
priority: p3
issue_id: "118"
tags: [code-review, documentation, api, developer-experience]
dependencies: ["100"]
---

# P3 MEDIUM: API Documentation Incomplete

## Problem Statement

Public APIs lack comprehensive documentation, making it difficult for external developers and wallets to integrate with the chain.

**Why it matters:** Poor documentation slows adoption, causes integration bugs, and increases support burden.

## Findings

### Missing Documentation

**1. Query Endpoints**
- No OpenAPI/Swagger specification
- Missing parameter descriptions
- No example requests/responses
- Error codes not documented

**2. Transaction Messages**
- No clear message schemas
- Missing field validation rules
- No example transactions
- Gas estimation not documented

**3. Module READMEs**
- Many modules lack README
- No architecture diagrams
- State transitions not documented
- No integration guides

### Current State

| Module | README | API Docs | Examples |
|--------|--------|----------|----------|
| identity | Partial | Missing | None |
| dex | Partial | Missing | None |
| bridge | Minimal | Missing | None |
| compliance | None | Missing | None |
| vcregistry | Partial | Missing | None |

## Proposed Solutions

### Solution A: Generate OpenAPI Docs + Module READMEs (Recommended)
**Effort:** 1 week | **Risk:** Low

**1. Enable OpenAPI generation:**

```bash
# In Makefile
docs-gen:
    @echo "Generating OpenAPI spec..."
    go run ./cmd/gendocs --output=./docs/api/openapi.yaml
```

**2. Module README template:**

```markdown
# Module Name

## Overview
Brief description of module purpose.

## Concepts
- **Term 1**: Definition
- **Term 2**: Definition

## State

### Entity Name
| Field | Type | Description |
|-------|------|-------------|
| id | string | Unique identifier |

## Messages

### MsgCreate
Creates a new entity.

**Parameters:**
| Field | Type | Required | Description |
|-------|------|----------|-------------|
| creator | string | Yes | Bech32 address |

**Example:**
```json
{
  "@type": "/aura.module.v1beta1.MsgCreate",
  "creator": "aura1...",
  "name": "example"
}
```

## Queries

### Entity
Get entity by ID.

**Request:**
```
GET /aura/module/v1beta1/entity/{id}
```

**Response:**
```json
{
  "entity": {
    "id": "1",
    "name": "example"
  }
}
```

## Events

| Event | Attributes | Description |
|-------|------------|-------------|
| entity_created | id, creator | Emitted when entity created |

## Errors

| Code | Name | Description |
|------|------|-------------|
| 1 | ErrNotFound | Entity not found |
```

## Recommended Action

**GO WITH SOLUTION A**: Generate API docs and create comprehensive READMEs.

## Technical Details

### Files to Create/Update

- `docs/api/openapi.yaml` - Generated API spec
- `chain/x/*/README.md` - Module documentation
- `docs/integration/README.md` - Integration guide
- `docs/examples/` - Example transactions

### Documentation Checklist (Per Module)

- [ ] README.md with full template
- [ ] All messages documented
- [ ] All queries documented
- [ ] All events documented
- [ ] All errors documented
- [ ] Example transactions
- [ ] State diagrams (where applicable)

## Acceptance Criteria

- [ ] OpenAPI spec generated and accurate
- [ ] All public modules have READMEs
- [ ] All messages documented with examples
- [ ] All queries documented with examples
- [ ] Error codes documented per module
- [ ] Integration guide for wallet developers

## Work Log

| Date | Action | Result |
|------|--------|--------|
| 2025-12-08 | Code review identified documentation gaps | P3 Medium |

## Resources

- [OpenAPI Specification](https://swagger.io/specification/)
- [Cosmos SDK Documentation Patterns](https://docs.cosmos.network/)
