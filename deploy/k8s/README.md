# ☸️ Kubernetes

---

Kustomize manifests for the Apexio stack on **kind** or **minikube**.

## 📦 Stack

| Workload   | Kind        | Service        | Notes                          |
|------------|-------------|----------------|--------------------------------|
| Redpanda   | StatefulSet | `redpanda:9092` (headless) | PVC `redpanda-data`; RPC on `0.0.0.0`, advertises pod DNS |
| ClickHouse | StatefulSet | `clickhouse`   | Init SQL from ConfigMap        |
| Grafana    | Deployment  | NodePort 30030 | Provisioned dashboards         |
| Gateway    | Deployment  | NodePort 30080 / 30417 | Build image locally    |
| Writer     | Deployment  | NodePort 30081 | Build image locally            |

## ✅ Prerequisites

- `kubectl` 1.28+
- `docker` (build gateway/writer images)
- **kind** or **minikube** for cluster deployment

---

## 🚀 Quick start (kind)

From the repository root:

```bash
# 1. Create a cluster (once)
kind create cluster --name apexio

# 2. Build and load app images into kind
docker build -t apexio-gateway:local -f cmd/gateway/Dockerfile .
docker build -t apexio-writer:local -f cmd/writer/Dockerfile .
kind load docker-image apexio-gateway:local --name apexio
kind load docker-image apexio-writer:local --name apexio

# 3. Deploy
kubectl apply -k deploy

# 4. Wait for pods
kubectl -n apexio wait --for=condition=ready pod --all --timeout=300s
```

### 🔍 Smoke test

```bash
kubectl -n apexio port-forward svc/gateway 18080:8080 &
curl -sf http://127.0.0.1:18080/healthz

curl -sS -X POST http://127.0.0.1:18080/api/v1/log \
  -H 'Content-Type: application/json' \
  -d '{
    "id": 1,
    "timestamp": 1732974309000,
    "logLevel": "INFO",
    "message": "hello from k8s",
    "source": {"service": "demo", "host": "kind", "environment": "dev"}
  }'

kubectl -n apexio exec clickhouse-0 -- clickhouse-client --query \
  "SELECT message FROM apexio.logs ORDER BY timestamp DESC LIMIT 3"
```

Grafana: `kubectl -n apexio port-forward svc/grafana 3000:3000` → [http://127.0.0.1:3000](http://127.0.0.1:3000) (`admin` / `admin`).

Or use NodePorts on kind: gateway `http://localhost:30080`, Grafana `http://localhost:30030`.

---

## 🖥️ Minikube

```bash
minikube start
eval $(minikube docker-env)
docker build -t apexio-gateway:local -f cmd/gateway/Dockerfile .
docker build -t apexio-writer:local -f cmd/writer/Dockerfile .
kubectl apply -k deploy
minikube service -n apexio gateway --url
```

With `minikube docker-env`, images are built into the minikube daemon (no separate load step).

---

## ⚙️ Configuration

Shared non-secret settings: ConfigMap `apexio-config` (`GATEWAY_API_KEY`, `GATEWAY_API_KEY_HEADER`, broker/ClickHouse FQDNs). Patch before apply or use a Kustomize overlay:

```bash
kubectl -n apexio patch configmap apexio-config --type merge \
  -p '{"data":{"GATEWAY_API_KEY":"your-secret"}}'
kubectl -n apexio rollout restart deployment/gateway
```

ClickHouse schema and Grafana provisioning are generated from the same files as Docker Compose (`deploy/clickhouse/init`, `deploy/grafana/provisioning`).

---

## 🧹 Teardown

```bash
kubectl delete -k deploy
# kind only:
kind delete cluster --name apexio
```

### Upgrading Redpanda networking

Kubernetes cannot change a Service from ClusterIP → headless or change StatefulSet `serviceName` in place. One-time fix:

```bash
kubectl -n apexio delete statefulset redpanda --cascade=orphan
kubectl -n apexio delete svc redpanda --ignore-not-found
kubectl apply -k deploy
kubectl -n apexio delete pod redpanda-0 --ignore-not-found
```

`--cascade=orphan` keeps the `redpanda-data` PVC. The pod is recreated with the new spec.

If Redpanda still crash-loops, reset data:

```bash
kubectl -n apexio delete pvc redpanda-data
kubectl -n apexio delete pod redpanda-0
```

---

## 🧪 Verify

```bash
make test-k8s          # manifest validation only (fast)
make test-k8s-e2e      # cluster smoke; tears down namespace and cluster on exit
```

`test-k8s-e2e` compiles binaries on the host when E2E runs. Set `APEXIO_K8S_REBUILD=1` to force image rebuild; existing `apexio-*:local` images are reused by default.
