#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
PROTO_DIR="$ROOT_DIR/proto"
OUT_DIR="$ROOT_DIR/pb"

if [[ ! -d "$PROTO_DIR" ]]; then
  echo "proto directory not found: $PROTO_DIR" >&2
  exit 1
fi

mkdir -p "$OUT_DIR"
find "$OUT_DIR" -maxdepth 1 -type f -name "*.pb.go" -delete

mapfile -t PROTO_FILES < <(find "$PROTO_DIR" -maxdepth 1 -type f -name "*.proto" | sort)
if [[ ${#PROTO_FILES[@]} -eq 0 ]]; then
  echo "no .proto files found in $PROTO_DIR" >&2
  exit 1
fi

protoc \
  --proto_path="$PROTO_DIR" \
  --go_out="$OUT_DIR" \
  --go_opt=paths=source_relative \
  --go-grpc_out="$OUT_DIR" \
  --go-grpc_opt=paths=source_relative \
  "${PROTO_FILES[@]}"

echo "proto generated into $OUT_DIR"
