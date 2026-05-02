.PHONY: up down ent-generate

include .env
export

DB_USER := $(shell cat secrets/db_user.txt)
DB_PASSWORD := $(shell cat secrets/db_password.txt)

DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@localhost:$(DB_PORT)/$(POSTGRES_DB)?sslmode=disable

MIGRATE := migrate
MIGRATIONS_DIR := db/migrations

up:
	docker compose up -d --build

down:
	docker compose down

migrate-create:
	$(MIGRATE) create -ext sql -dir $(MIGRATIONS_DIR) $(NAME)

migrate-up:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

migrate-down:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down 1

migrate-force:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" force $(VERSION)

migrate-version:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" version

migrate-drop:
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DB_URL)" drop -f

seed:
	docker compose up -d postgres
	sleep 2
	make migrate-up
	docker compose build svc-layers
	docker compose run -e SEED_DB=true --rm svc-layers

ent-generate:
	go generate ./internal/auth/ent

ent-diff:
	go run ./internal/auth/scripts/diff.go $(NAME)