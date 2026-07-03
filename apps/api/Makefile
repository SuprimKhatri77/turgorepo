ENV_FILE ?= ../../.env.local
include $(ENV_FILE)
export

.PHONY: migrate-up migrate-down
migrate-up:
	migrate -path migrations -database "$${DATABASE_URL}" up
migrate-down:
	migrate -path migrations -database "$${DATABASE_URL}" down ${N}

# .PHONY: seed
# seed:
# 	go run ./seed

.PHONY: run build
run:
	go run ./cmd/server
build:
	go build -o server ./cmd/server

.PHONY: deps
deps:
	go mod tidy
	go mod download