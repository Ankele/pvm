#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PROTO_DIR="${ROOT_DIR}/api/proto"
OUT_DIR="${ROOT_DIR}/api/gen"
PROTO_INCLUDE="$(cd "$(dirname "$(command -v protoc)")/.." && pwd)/include"

COMMON_FILES=(
  "${PROTO_DIR}/pvm/v1/common.proto"
  "${PROTO_DIR}/pvm/v1/system.proto"
  "${PROTO_DIR}/pvm/v1/vm.proto"
  "${PROTO_DIR}/pvm/v1/storage.proto"
  "${PROTO_DIR}/pvm/v1/network.proto"
  "${PROTO_DIR}/pvm/v1/interface.proto"
  "${PROTO_DIR}/pvm/v1/snapshot.proto"
)

protoc \
  -I "${PROTO_DIR}" \
  -I "${ROOT_DIR}/third_party" \
  -I "${PROTO_INCLUDE}" \
  --go_out="${OUT_DIR}" \
  --go_opt=paths=source_relative \
  --go-grpc_out="${OUT_DIR}" \
  --go-grpc_opt=paths=source_relative \
  "${COMMON_FILES[@]}"
