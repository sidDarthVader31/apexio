#!/usr/bin/env bash
# Kubernetes: manifest validation (default) and optional cluster smoke (APEXIO_K8S_E2E=1).
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=test-lib.sh
source "${SCRIPT_DIR}/test-lib.sh"

K8S_KUSTOMIZE="${ROOT}/deploy"
CLUSTER_NAME="${KIND_CLUSTER_NAME:-apexio}"
SMOKE_MSG="k8s-smoke-$(date +%s)-$$"

test_manifests() {
  require_cmd kubectl
  kubectl kustomize "${K8S_KUSTOMIZE}" >/dev/null
  pass "kubectl kustomize build succeeds"

  [[ -f "${ROOT}/deploy/k8s/README.md" ]] || fail "missing deploy/k8s/README.md"
  grep -q 'kind create cluster' "${ROOT}/deploy/k8s/README.md" || fail "k8s README missing kind instructions"
  pass "deploy/k8s/README.md present"
}

image_exists() {
  docker image inspect "$1" >/dev/null 2>&1
}

# Host compile + thin image layers — avoids multi-minute in-docker go builds.
build_app_images() {
  require_cmd go
  require_cmd docker

  if [[ "${APEXIO_K8S_REBUILD:-0}" != "1" ]] && image_exists apexio-gateway:local && image_exists apexio-writer:local; then
    info "using existing apexio-gateway:local and apexio-writer:local (APEXIO_K8S_REBUILD=1 to force)"
    return 0
  fi

  info "building app binaries on host (fast path)"
  local bin_dir
  bin_dir="$(mktemp -d)"
  trap 'rm -rf "${bin_dir}"' RETURN

  CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o "${bin_dir}/gateway" ./cmd/gateway
  CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o "${bin_dir}/writer" ./cmd/writer

  docker build -t apexio-gateway:local -f- "${bin_dir}" <<'EOF'
FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget
COPY gateway /usr/local/bin/gateway
EXPOSE 8080 4317
ENTRYPOINT ["/usr/local/bin/gateway"]
EOF

  docker build -t apexio-writer:local -f- "${bin_dir}" <<'EOF'
FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget
COPY writer /usr/local/bin/writer
EXPOSE 8081
ENTRYPOINT ["/usr/local/bin/writer"]
EOF

  pass "app images built"
}

load_images_kind() {
  kind load docker-image apexio-gateway:local --name "${CLUSTER_NAME}"
  kind load docker-image apexio-writer:local --name "${CLUSTER_NAME}"
}

deploy_and_smoke() {
  info "deploying stack to namespace apexio"
  kubectl delete namespace apexio --ignore-not-found=true --wait=true >/dev/null 2>&1 || true
  kubectl apply -k "${K8S_KUSTOMIZE}"

  kubectl -n apexio rollout status statefulset/redpanda --timeout=300s
  kubectl -n apexio rollout status statefulset/clickhouse --timeout=300s
  kubectl -n apexio rollout status deployment/gateway --timeout=300s
  kubectl -n apexio rollout status deployment/writer --timeout=300s
  pass "all apexio workloads ready"

  kubectl -n apexio port-forward svc/gateway 18080:8080 >/tmp/apexio-k8s-pf.log 2>&1 &
  local pf_pid=$!
  trap 'kill "${pf_pid}" 2>/dev/null || true' EXIT

  wait_http_ok "http://127.0.0.1:18080/healthz" 60
  pass "gateway healthz via port-forward"

  local code
  code="$(curl -s -o /tmp/apexio-k8s-post.json -w '%{http_code}' \
    -X POST "http://127.0.0.1:18080/api/v1/log" \
    -H 'Content-Type: application/json' \
    -d "{
      \"id\": ${RANDOM}${RANDOM},
      \"timestamp\": $(date +%s000),
      \"logLevel\": \"INFO\",
      \"message\": \"${SMOKE_MSG}\",
      \"metadata\": {
        \"requestMethod\": \"POST\",
        \"requestPath\": \"/k8s-smoke\",
        \"responseStatus\": 201,
        \"responseDuration\": 10.5
      },
      \"source\": {
        \"service\": \"k8s-smoke\",
        \"host\": \"cluster\",
        \"environment\": \"dev\"
      }
    }")"
  [[ "${code}" == "201" ]] || fail "ingest failed (${code}): $(cat /tmp/apexio-k8s-post.json)"
  pass "REST ingest via k8s gateway (201)"

  wait_clickhouse_message "${SMOKE_MSG}"
  pass "log visible in ClickHouse"
}

run_kind_e2e() {
  if ! kind get clusters 2>/dev/null | grep -qx "${CLUSTER_NAME}"; then
    info "creating kind cluster ${CLUSTER_NAME}"
    kind create cluster --name "${CLUSTER_NAME}" --wait 120s
  fi
  kubectl config use-context "kind-${CLUSTER_NAME}" >/dev/null
  build_app_images
  load_images_kind
  deploy_and_smoke
}

run_minikube_e2e() {
  minikube start --driver=docker
  eval "$(minikube docker-env)"
  build_app_images
  deploy_and_smoke
}

test_cluster_e2e() {
  require_cmd docker

  if command -v kind >/dev/null 2>&1; then
    run_kind_e2e
    return
  fi
  if command -v minikube >/dev/null 2>&1; then
    run_minikube_e2e
    return
  fi
  fail "APEXIO_K8S_E2E=1 requires kind or minikube"
}

main() {
  cd "${ROOT}"
  test_manifests
  if [[ "${APEXIO_K8S_E2E:-0}" == "1" ]]; then
    test_cluster_e2e
  else
    info "skipping cluster smoke (set APEXIO_K8S_E2E=1 to enable)"
  fi
  echo
  pass "kubernetes tests passed"
}

main "$@"
