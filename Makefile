COMPOSE_FILE := deploy/compose/docker-compose.yml
COMPOSE := docker compose -f $(COMPOSE_FILE)

.PHONY: up down logs ps restart clean-volumes \
	test test-unit test-contracts test-infra test-pipeline test-otlp test-grafana test-auth test-k8s test-k8s-e2e test-e2e \
	k8s-apply k8s-delete

## Start stack (infra + gateway + writer); rebuild app images
up:
	$(COMPOSE) up -d --build

## Stop infra (keeps volumes)
down:
	$(COMPOSE) down

## Follow container logs
logs:
	$(COMPOSE) logs -f

## Show compose service status
ps:
	$(COMPOSE) ps

## Restart all services
restart:
	$(COMPOSE) restart

## Remove containers and named volumes (destructive)
clean-volumes:
	$(COMPOSE) down -v

## Fast default: unit + contract layout + k8s manifest validation
test: test-unit test-contracts test-k8s

## Go unit and race tests
test-unit:
	./scripts/test-unit.sh

## Shared package layout + unit tests
test-contracts:
	./scripts/test-contracts.sh

## Compose infra: Redpanda, ClickHouse schema, Grafana datasource
test-infra:
	./scripts/test-infra.sh

## E2E REST ingest through full compose pipeline
test-pipeline:
	./scripts/test-pipeline.sh

## OTLP HTTP ingest + sample client
test-otlp:
	./scripts/test-otlp.sh

## Grafana dashboard provisioning
test-grafana:
	./scripts/test-grafana.sh

## API-key auth, writer metrics, broker docs
test-auth:
	./scripts/test-auth.sh

## Kubernetes manifest validation (no cluster required)
test-k8s:
	./scripts/test-k8s.sh

## Kubernetes cluster smoke (requires kind or minikube; set APEXIO_K8S_E2E=1)
test-k8s-e2e:
	APEXIO_K8S_E2E=1 ./scripts/test-k8s.sh

## All Docker Compose E2E component tests
test-e2e: test-infra test-pipeline test-otlp test-grafana test-auth

## Apply Kubernetes stack (requires images loaded into cluster)
k8s-apply:
	kubectl apply -k deploy

## Remove Kubernetes stack (keeps kind/minikube cluster)
k8s-delete:
	kubectl delete -k deploy --ignore-not-found
