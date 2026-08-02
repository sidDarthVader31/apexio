COMPOSE_FILE := deploy/compose/docker-compose.yml
COMPOSE := docker compose -f $(COMPOSE_FILE)

.PHONY: up down logs ps test-phase1 restart clean-volumes

## Start Phase 1 infra (Redpanda, ClickHouse, Grafana)
up:
	$(COMPOSE) up -d

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

## Run Phase 1 infra smoke tests (starts stack if needed)
test-phase1:
	./scripts/test-phase1.sh
