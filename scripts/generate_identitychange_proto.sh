#!/bin/sh

# Helper script to regenerate the identity change protobuf bindings.
# Requires either `buf` (preferred) or `protoc` + `protoc-gen-go`.

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

generate_with_buf() {
  cd "$ROOT/proto"
  buf generate --template buf.gen.yaml
}

generate_with_protoc() {
  if ! command -v protoc >/dev/null 2>&1; then
    echo "protoc is not installed" >&2
    return 1
  fi
  if ! command -v protoc-gen-go >/dev/null 2>&1; then
    echo "protoc-gen-go is not installed" >&2
    return 1
  fi
  cd "$ROOT"
  protoc \
    --proto_path=proto \
    --proto_path=proto/cosmos \
    --go_out=paths=source_relative:proto \
    proto/aura/identitychange/v1beta1/identity_change.proto
}

if command -v buf >/dev/null 2>&1; then
  generate_with_buf
else
  generate_with_protoc
fi
