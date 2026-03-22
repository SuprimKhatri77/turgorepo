# Run migrations (requires golang-migrate and DATABASE_URL)
# Install: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
.PHONY: migrate-up migrate-down
migrate-up:
	migrate -path migrations -database "$${DATABASE_URL}" up
migrate-down:
	migrate -path migrations -database "$${DATABASE_URL}" down

.PHONY: run build
run:
	go run ./cmd/server
build:
	go build -o server ./cmd/server

.PHONY: deps
deps:
	go mod tidy
	go mod download
