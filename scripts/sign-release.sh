#!/usr/bin/env bash
set -euo pipefail

CHECKSUMS="${1:-}"
OUTPUT="${2:-}"

if [[ -z "$CHECKSUMS" ]]; then
  echo "usage: $0 <checksums-file> [signature-output]" >&2
  exit 1
fi

if [[ ! -f "$CHECKSUMS" ]]; then
  echo "checksums file not found: $CHECKSUMS" >&2
  exit 1
fi

if [[ -z "$OUTPUT" ]]; then
  OUTPUT="${CHECKSUMS}.sig"
fi

TMP_KEY="$(mktemp)"
cleanup() {
  rm -f "$TMP_KEY"
}
trap cleanup EXIT

if [[ -n "${MOOMOO_CLI_RELEASE_SIGNING_KEY_FILE:-}" ]]; then
  if [[ ! -f "$MOOMOO_CLI_RELEASE_SIGNING_KEY_FILE" ]]; then
    echo "MOOMOO_CLI_RELEASE_SIGNING_KEY_FILE does not point to a file: $MOOMOO_CLI_RELEASE_SIGNING_KEY_FILE" >&2
    exit 1
  fi
  cp "$MOOMOO_CLI_RELEASE_SIGNING_KEY_FILE" "$TMP_KEY"
elif [[ -n "${MOOMOO_CLI_RELEASE_SIGNING_KEY:-}" ]]; then
  printf '%s\n' "$MOOMOO_CLI_RELEASE_SIGNING_KEY" >"$TMP_KEY"
else
  echo "set MOOMOO_CLI_RELEASE_SIGNING_KEY to the PEM private key contents, or MOOMOO_CLI_RELEASE_SIGNING_KEY_FILE to a key path" >&2
  exit 1
fi
chmod 600 "$TMP_KEY"

openssl dgst -sha256 -sign "$TMP_KEY" -out "$OUTPUT" "$CHECKSUMS"
echo "wrote signature $OUTPUT"
