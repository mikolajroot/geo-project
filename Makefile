.PHONY: up down ent-generate

include .env
export


MIGRATE := migrate
MIGRATIONS_DIR := db/migrations
MIGRATION_WORKER := docker compose run --rm migration-worker

up:
	docker compose up -d --build

down:
	docker compose down

migrate-create:
	$(MIGRATE) create -ext sql -dir $(MIGRATIONS_DIR) $(NAME)

migrate-up:
	docker compose run --rm migration-worker up

migrate-down:
	docker compose run --rm migration-worker down

migrate-drop:
	docker compose run --rm migration-worker drop

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