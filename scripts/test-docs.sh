#!/usr/bin/env bash
# Documentation hygiene: no removed paths, README mentions current stack.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=test-lib.sh
source "${SCRIPT_DIR}/test-lib.sh"
cd "${ROOT}"

removed_paths=(
  log_ingestion_service
  log_processing_service
  visualization_service
  deployments
  deployments/k8-config
)

for p in "${removed_paths[@]}"; do
  [[ ! -e "${p}" ]] || fail "legacy path still exists: ${p}"
done
pass "no legacy directories on disk"

forbidden_patterns=(
  'log_ingestion_service'
  'log_processing_service'
  'visualization_service'
  'deployments/k8-config'
  'Elasticsearch'
  'elasticsearch'
)

md_files=(
  README.md
  deploy/README.md
  deploy/k8s/README.md
  pkg/README.md
  pkg/broker/README.md
  test.md
)

for f in "${md_files[@]}"; do
  [[ -f "${f}" ]] || fail "missing doc: ${f}"
done
pass "required markdown files present"

for f in "${md_files[@]}"; do
  for pat in "${forbidden_patterns[@]}"; do
    if grep -q "${pat}" "${f}"; then
      fail "${f} still references removed stack: ${pat}"
    fi
  done
done
pass "no forbidden legacy references in markdown"

# Root README should describe the current product surface.
readme_checks=(
  'make up'
  'Redpanda'
  'ClickHouse'
  'cmd/gateway'
  'OTLP'
  'deploy/k8s'
)

for needle in "${readme_checks[@]}"; do
  grep -q "${needle}" README.md || fail "README.md missing expected content: ${needle}"
done
pass "README.md smoke checks"

# Repo-wide grep for legacy paths (exclude this script and git history).
if git grep -l 'log_ingestion_service\|log_processing_service\|deployments/k8-config' -- . \
  ':(exclude)scripts/test-docs.sh' 2>/dev/null | grep -q .; then
  git grep -n 'log_ingestion_service\|log_processing_service\|deployments/k8-config' -- . \
    ':(exclude)scripts/test-docs.sh' 2>/dev/null || true
  fail "legacy path references remain in tracked files"
fi
pass "no legacy path references in source"

echo
pass "documentation tests passed"
