ENV_FILE ?= ../../.env.local
# Optional: sqlc, build, lint and test need no environment, and CI has no env file.
-include $(ENV_FILE)
export

.PHONY: require-database-url migrate-up migrate-down sqlc seed run build lint test deps

require-database-url:
	@test -n "$${DATABASE_URL}" || { echo "DATABASE_URL is not set. Run: cp .env.example .env.local (repo root)"; exit 1; }

migrate-up: require-database-url
	migrate -path migrations -database "$${DATABASE_URL}" up
migrate-down: require-database-url
	migrate -path migrations -database "$${DATABASE_URL}" down ${N}

sqlc:
	sqlc generate

seed:
	go run ./cmd/seed

run:
	go run ./cmd/server
build:
	go build -o server ./cmd/server

lint:
	golangci-lint run ./...

test:
	go test ./...

deps:
	go mod tidy
	go mod download
