.PHONY: run run-worker test test-race tidy lint sqlc migrate-up migrate-down migrate-create

run:
	go run ./apps/api

run-worker:
	go run ./apps/worker

test:
	go test ./...

test-race:
	go test -race ./...

tidy:
	go mod tidy

sqlc:
	sqlc generate -f database/sqlc/sqlc.yaml

migrate-up:
	migrate -path database/migrations -database "$$DATABASE_URL" up

migrate-down:
	migrate -path database/migrations -database "$$DATABASE_URL" down 1

migrate-create:
	@read -p "Name: " name; migrate create -ext sql -dir database/migrations -seq $$name
