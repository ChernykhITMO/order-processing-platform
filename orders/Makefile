DB_URL = $(ORDERS_PG_DSN)
MIGRATIONS_DIR = ./migrations
DB=orders_db

run:
	go run cmd/main.go

migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down 1

db-create:
	docker compose -f ../docker-compose.yaml exec postgres createdb -U postgres $(DB)
