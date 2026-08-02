#!/usr/bin/env bash
# Phase 2 shared-contract tests.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

RED=$'\033[0;31m'
GREEN=$'\033[0;32m'
YELLOW=$'\033[0;33m'
NC=$'\033[0m'

pass() { echo -e "${GREEN}PASS${NC}: $*"; }
fail() { echo -e "${RED}FAIL${NC}: $*"; exit 1; }
info() { echo -e "${YELLOW}INFO${NC}: $*"; }

command -v go >/dev/null 2>&1 || fail "go not found"

[[ -f go.mod ]] || fail "missing go.mod at repo root"
grep -q 'module github.com/sidDarthVader31/apexio' go.mod || fail "unexpected module path in go.mod"

required=(
  pkg/schema/event.go
  pkg/broker/broker.go
  pkg/broker/memory.go
  pkg/broker/redpanda.go
  pkg/store/store.go
  pkg/store/memory.go
  pkg/store/clickhouse.go
  pkg/README.md
)
for f in "${required[@]}"; do
  [[ -f "${f}" ]] || fail "missing ${f}"
done
pass "package layout present"

info "running go test ./pkg/..."
go test ./pkg/... -count=1 -timeout 60s
pass "go test ./pkg/... "

info "race detector"
go test ./pkg/broker/... ./pkg/store/... -race -count=1 -timeout 90s
pass "race tests"

echo
pass "Phase 2 contract tests all passed"
