SERVICES := orders payments
LINT_SERVICES := gateway notifications orders payments
GOLANGCI_LINT_IMAGE := golangci/golangci-lint:latest
COMPOSE ?= docker compose -f docker-compose.yaml
ORDERS_DSN_DOCKER := postgres://postgres:postgres@postgres:5432/orders_db?sslmode=disable
PAYMENTS_DSN_DOCKER := postgres://postgres:postgres@postgres:5432/payments_db?sslmode=disable

.PHONY: db-create migrate-up migrate-down docker-migrate-up docker-migrate-down docker-migrate-reset
.PHONY: lint

db-create:
	@for svc in $(SERVICES); do \
		$(MAKE) -C $$svc db-create; \
	done

migrate-up:
	@for svc in $(SERVICES); do \
		$(MAKE) -C $$svc migrate-up; \
	done

migrate-down:
	@for svc in $(SERVICES); do \
		$(MAKE) -C $$svc migrate-down; \
	done

docker-migrate-up:
	$(COMPOSE) run --rm migrate-orders -path /migrations -database "$(ORDERS_DSN_DOCKER)" up
	$(COMPOSE) run --rm migrate-payments -path /migrations -database "$(PAYMENTS_DSN_DOCKER)" up

docker-migrate-down:
	$(COMPOSE) run --rm migrate-orders -path /migrations -database "$(ORDERS_DSN_DOCKER)" down -all
	$(COMPOSE) run --rm migrate-payments -path /migrations -database "$(PAYMENTS_DSN_DOCKER)" down -all

docker-migrate-reset: docker-migrate-down docker-migrate-up

lint:
	@for svc in $(LINT_SERVICES); do \
		echo "==> lint $$svc"; \
		docker run --rm -t -v "$$(pwd)/$$svc":/app -w /app $(GOLANGCI_LINT_IMAGE) golangci-lint run; \
	done
