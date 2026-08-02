#!/usr/bin/env bash
# Unit tests: pkg/, cmd/, and sample client.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=test-lib.sh
source "${SCRIPT_DIR}/test-lib.sh"
cd "${ROOT}"

require_cmd go

[[ -f go.mod ]] || fail "missing go.mod at repo root"

info "go test ./pkg/... ./cmd/... ./examples/..."
go test ./pkg/... ./cmd/... ./examples/... -count=1 -timeout 120s
pass "unit tests"

info "race detector (broker, store)"
go test ./pkg/broker/... ./pkg/store/... -race -count=1 -timeout 90s
pass "race tests"

echo
pass "unit tests passed"
