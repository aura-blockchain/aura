#!/bin/sh

# Helper script to regenerate the VC registry protobuf bindings.
# Requires either `buf` (preferred) or `protoc` + `protoc-gen-go` + `protoc-gen-go-grpc`.

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
  if ! command -v protoc-gen-go-grpc >/dev/null 2>&1; then
    echo "protoc-gen-go-grpc is not installed" >&2
    return 1
  fi
  cd "$ROOT"
  protoc \
    --proto_path=proto \
    --proto_path=proto/cosmos \
    --go_out=paths=source_relative:proto \
    --go-grpc_out=paths=source_relative:proto \
    proto/aura/vcregistry/v1beta1/vc_registry.proto
}

if command -v buf >/dev/null 2>&1; then
  generate_with_buf
else
  generate_with_protoc
fi
