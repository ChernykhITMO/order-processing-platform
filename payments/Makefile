DB_URL = postgres://postgres:postgres@localhost:5436/payments_db?sslmode=disable
MIGRATIONS_DIR = ./migrations
DB=payments_db

run:
	go run cmd/main.go

migrate-up:
	go run ./cmd/migrator -db "$(DB_URL)" -dir "$(MIGRATIONS_DIR)" -command up

migrate-down:
	go run ./cmd/migrator -db "$(DB_URL)" -dir "$(MIGRATIONS_DIR)" -command down

db-create:
	docker compose -f ../docker-compose.yaml exec postgres createdb -U postgres $(DB)
