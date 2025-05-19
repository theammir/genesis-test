#!/usr/bin/env bash
set -euo pipefail

DIR="debug/certs"
CERT="$DIR/cert.pem"
KEY="$DIR/key.pem"

if [ ! -d "$DIR" ]; then
  mkdir -p "$DIR"
fi

if [ ! -f "$CERT" ] || [ ! -f "$KEY" ]; then
  openssl req -x509 -newkey rsa:4096 -nodes \
    -keyout "$KEY" \
    -out    "$CERT" \
    -days   3650 \
    -subj   "/CN=localhost"
fi

docker compose -f compose.yaml -f compose.debug.yaml --profile debug "$@"
