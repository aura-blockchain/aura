# Development

This repo contains multiple services. The chain code lives under `chain/`.

## Prerequisites

- Go 1.21+
- Make
- Git
- jq (for scripts)

Additional tooling is required per service (see each service README).

## Build the chain

```bash
cd chain
make build
```

## Tests

```bash
cd chain
make test
```

## Lint

```bash
cd chain
make lint
```

## Protobuf generation

```bash
cd chain
make proto-gen
```

## Pre-commit hooks (optional)

```bash
pre-commit install
```
