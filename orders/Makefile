DB_URL = postgres://postgres:postgres@localhost:5436/orders_db?sslmode=disable
MIGRATIONS_DIR = ./migrations
run:
	go run cmd/app/main.go

migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down 1