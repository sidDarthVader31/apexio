COMPOSE_FILE := deploy/compose/docker-compose.yml
COMPOSE := docker compose -f $(COMPOSE_FILE)

.PHONY: up down logs ps test-phase1 test-phase2 test-phase3 restart clean-volumes

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

## Run Phase 1 infra smoke tests (starts stack if needed)
test-phase1:
	./scripts/test-phase1.sh

## Run Phase 2 shared-contract unit tests
test-phase2:
	./scripts/test-phase2.sh

## Run Phase 3 unit + E2E vertical-slice tests
test-phase3:
	./scripts/test-phase3.sh
