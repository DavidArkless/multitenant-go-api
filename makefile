MIGRATIONS_DIR := migrations
DATABASE_URL := sqlite3://app.db
GOFLAGS := -mod=vendor


UNAME_S := $(shell uname -s)


.PHONY: install-migrate
install-migrate:
	@echo "Installing golang-migrate/migrate with sqlite support..."
	go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest


.PHONY: migration-new
migration-new:
	@if [ -z "$(name)" ]; then \
		echo "Usage: make migration-new name=your_migration_name"; \
		exit 1; \
	fi
	@mkdir -p $(MIGRATIONS_DIR)
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

.PHONY: migrate-up
migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

.PHONY: migrate-down
migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1

.PHONY: migrate-reset
migrate-reset:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down -all
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up


.PHONY: test
test:
	@echo "Running all tests..."
	go test -v ./...

.PHONY: test-unit
test-unit:
	@echo "Running unit tests..."
	go test -v -short ./...

.PHONY: test-e2e
test-e2e:
	@echo "Running E2E tests..."
	go test -v ./internal/api/...

.PHONY: run
run:
	@echo "Starting API server..."
	go run cmd/api/main.go

.PHONY: build
build:
	@echo "Building binary..."
	go build -o bin/api cmd/api/main.go