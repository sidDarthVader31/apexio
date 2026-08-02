# 🧪 Testing guide

---

How Apexio validates the pipeline locally and in CI.

## ⚡ Quick commands

| Command | What it runs |
|---------|----------------|
| `make test` | Unit + contracts + k8s manifest validation (fast) |
| `make test-e2e` | Full Docker Compose E2E suite |
| `make test-k8s-e2e` | Kubernetes cluster smoke (opt-in) |
| `make test-docs` | No stale paths; README smoke checks |

---

## 📂 Component scripts

All scripts live in [`scripts/`](scripts/) and source [`scripts/test-lib.sh`](scripts/test-lib.sh) for `pass` / `fail` and cleanup traps.

| Target | Script | Scope |
|--------|--------|-------|
| `test-unit` | `test-unit.sh` | Go tests + race detector |
| `test-contracts` | `test-contracts.sh` | `pkg/` layout and tests |
| `test-infra` | `test-infra.sh` | Redpanda, ClickHouse, Grafana up |
| `test-pipeline` | `test-pipeline.sh` | REST → broker → ClickHouse |
| `test-otlp` | `test-otlp.sh` | OTLP HTTP + sample client |
| `test-grafana` | `test-grafana.sh` | Provisioned dashboards |
| `test-auth` | `test-auth.sh` | API-key middleware, writer metrics |
| `test-k8s` | `test-k8s.sh` | `kubectl kustomize` validation |
| `test-docs` | `test-docs.sh` | Documentation hygiene |

---

## 🐳 Compose E2E

`make test-e2e` runs infra → pipeline → OTLP → Grafana → auth in sequence.

- Starts or reuses the compose stack
- Registers `EXIT` cleanup: `docker compose down -v`
- Requires Docker and ~2–5 minutes on first run (image builds)

---

## ☸️ Kubernetes E2E

`make test-k8s-e2e` sets `APEXIO_K8S_E2E=1` and requires **kind** or **minikube**.

- Builds gateway/writer images on the host (faster than in-cluster compile)
- Applies `kubectl apply -k deploy`, smoke-tests ingest + ClickHouse
- Tears down `apexio` namespace; deletes kind/minikube cluster if the test created it

Set `APEXIO_K8S_REBUILD=1` to force image rebuild.

---

## ✅ Before opening a PR

```bash
make test
make test-e2e        # if you touched gateway, writer, deploy, or pkg
make test-docs
```

Pipeline or K8s changes: also run `make test-k8s-e2e` when you have a local cluster.
