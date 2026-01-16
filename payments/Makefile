DB_URL = postgres://postgres:postgres@localhost:5436/payments_db?sslmode=disable
MIGRATIONS_DIR = ./migrations
DB=payments_db

run:
	go run cmd/app/main.go

migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down 1

db-create:
	docker compose -f ../docker-compose.yaml exec postgres createdb -U postgres $(DB)
