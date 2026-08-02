#!/usr/bin/env bash
# Shared contracts: package layout and schema/broker/store packages.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=test-lib.sh
source "${SCRIPT_DIR}/test-lib.sh"
cd "${ROOT}"

require_cmd go

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

go test ./pkg/... -count=1 -timeout 60s
pass "pkg tests"

echo
pass "contract tests passed"
