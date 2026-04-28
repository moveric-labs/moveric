BINARY_DIR=bin

.PHONY: all build source-agent dest-agent api tidy migrate run-infra

all: build

build: source-agent dest-agent api

source-agent:
	go build -o $(BINARY_DIR)/source-agent ./cmd/source-agent

dest-agent:
	go build -o $(BINARY_DIR)/dest-agent ./cmd/dest-agent

api:
	go build -o $(BINARY_DIR)/api ./cmd/api

tidy:
	go mod tidy

migrate:
	psql "$(POSTGRES_DSN)" -f docs/schema.sql

run-infra:
	podman-compose up -d

stop-infra:
	podman-compose down
