#!/usr/bin/env bash
set -euo pipefail

GO_BIN="${GO_BIN:-go}"
GOFMT_BIN="${GOFMT_BIN:-gofmt}"

"$GOFMT_BIN" -w src/*.go
test -z "$("$GOFMT_BIN" -l src/*.go)"
"$GO_BIN" vet ./...
"$GO_BIN" test ./...
node scripts/build.mjs --clean
node scripts/check-loc.mjs
node --test --experimental-strip-types "tests/node/*.test.ts"
