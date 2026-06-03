.PHONY: help up down api build test sqlc migrate migrate-down migrate-create token web-install web tidy

# --- defaults ---
DB_URL      ?= postgres://fintrack:fintrack@localhost:55432/fintrack?sslmode=disable
TEST_DB_URL ?= postgres://fintrack:fintrack@localhost:55432/fintrack_test?sslmode=disable
MIGRATE      = migrate -path database/migrations -database "$(DB_URL)"
MIGRATE_TEST = migrate -path database/migrations -database "$(TEST_DB_URL)"

help: ## list available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

# --- docker ---
up: ## start postgres + minio
	docker compose up -d
	@echo "postgres: localhost:5432  minio: localhost:9000 (console :9001)"

down: ## stop postgres + minio (keeps volumes)
	docker compose down

reset: ## NUKES docker volumes — postgres + minio data gone
	docker compose down -v

# --- backend ---
tidy: ## go mod tidy
	go mod tidy

api: ## run the API server with hot config from .env
	go run ./apps/api

build: ## build the API binary
	go build -o bin/api ./apps/api

test: test-db-ensure ## run all go tests (incl. integration against fintrack_test)
	TEST_DATABASE_URL="$(TEST_DB_URL)" go test ./...

test-unit: ## run only unit tests (no DB required)
	go test -short ./...

test-integration: test-db-ensure ## run only integration tests (requires postgres)
	TEST_DATABASE_URL="$(TEST_DB_URL)" go test -run Integration ./...

# --- database ---
migrate: ## apply all up migrations
	$(MIGRATE) up

migrate-down: ## roll back the most recent migration
	$(MIGRATE) down 1

migrate-create: ## create a new migration pair: make migrate-create name=add_transactions
	$(MIGRATE) create -ext sql -dir database/migrations -seq $(name)

migrate-test: ## apply migrations to fintrack_test
	$(MIGRATE_TEST) up

test-db-ensure: ## create fintrack_test DB if missing + apply migrations
	@docker exec fintrack-postgres psql -U fintrack -d fintrack -tAc "SELECT 1 FROM pg_database WHERE datname='fintrack_test'" | grep -q 1 \
		|| docker exec fintrack-postgres createdb -U fintrack fintrack_test
	@$(MIGRATE_TEST) up >/dev/null 2>&1 || true

sqlc: ## regenerate sqlc code from query/*.sql
	cd database/sqlc && sqlc generate

# --- auth helpers ---
token: ## mint a local JWT (override: make token SUB=<uuid>)
	@go run ./cmd/mint-jwt $(if $(SUB),-sub $(SUB))

# --- frontend ---
web-install: ## install web dependencies
	cd web && npm install

web: ## run the vite dev server
	cd web && npm run dev

test-e2e: test-db-ensure ## run playwright e2e tests (starts api + vite via webServer)
	cd web && npm run test:e2e
