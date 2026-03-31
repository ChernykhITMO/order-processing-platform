DB_URL = $(ORDERS_PG_DSN)
MIGRATIONS_DIR = ./migrations
DB=orders_db

run:
	go run cmd/main.go

migrate-up:
	go run ./cmd/migrator -db "$(DB_URL)" -dir "$(MIGRATIONS_DIR)" -command up

migrate-down:
	go run ./cmd/migrator -db "$(DB_URL)" -dir "$(MIGRATIONS_DIR)" -command down

db-create:
	docker compose -f ../docker-compose.yaml exec postgres createdb -U postgres $(DB)
