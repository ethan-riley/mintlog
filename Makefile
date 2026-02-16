.PHONY: build build-all test lint clean migrate-up migrate-down sqlc seed infra

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
BINARY_DIR=bin

SERVICES=ingestd pipelined apid alertd notifierd

# Database
PG_URL ?= postgres://mintlog:mintlog@localhost:6543/mintlog?sslmode=disable

build-all: $(addprefix build-,$(SERVICES))

build-%:
	$(GOBUILD) -o $(BINARY_DIR)/$* ./cmd/$*/

build: build-all

test:
	$(GOTEST) -race -count=1 ./...

test-integration:
	$(GOTEST) -race -tags=integration -count=1 ./tests/integration/...

lint:
	golangci-lint run ./...

clean:
	rm -rf $(BINARY_DIR)

migrate-up:
	migrate -path internal/storage/postgres/migrations -database "$(PG_URL)" up

migrate-down:
	migrate -path internal/storage/postgres/migrations -database "$(PG_URL)" down 1

migrate-create:
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir internal/storage/postgres/migrations -seq $$name

sqlc:
	sqlc generate -f sql/sqlc.yaml

infra:
	docker compose up -d

infra-down:
	docker compose down

seed:
	./scripts/seed.sh

run-ingestd: build-ingestd
	./$(BINARY_DIR)/ingestd

run-apid: build-apid
	./$(BINARY_DIR)/apid

run-pipelined: build-pipelined
	./$(BINARY_DIR)/pipelined

run-alertd: build-alertd
	./$(BINARY_DIR)/alertd

run-notifierd: build-notifierd
	./$(BINARY_DIR)/notifierd
