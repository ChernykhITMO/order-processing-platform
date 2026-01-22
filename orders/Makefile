DB_URL = postgres://postgres:postgres@localhost:5436/orders_db?sslmode=disable
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
