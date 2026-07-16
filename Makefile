# Makefile — BackendDataPlatform
# ──────────────────────────────────────────────────────────────────────────
# Prerequisites: buf, protoc-gen-go, protoc-gen-go-grpc, grpc_tools (python)

.PHONY: proto proto-go proto-py build-ingestor build-streaming build-query build-api \
        build-all test lint docker-compose-up docker-compose-down topics

# ── Protobuf generation ───────────────────────────────────────────────────

## Generate Go stubs from all proto files using buf
proto-go:
	buf generate --template buf.gen.yml

## Generate Python stubs for batch service
proto-py:
	python -m grpc_tools.protoc \
		-I protobuf/batch-data \
		--python_out=batch \
		--grpc_python_out=batch \
		protobuf/batch-data/batch_data.proto

proto: proto-go proto-py

# ── Go builds ─────────────────────────────────────────────────────────────

build-ingestor:
	go build -o bin/ingestor ./cmd/ingestor

build-streaming:
	go build -o bin/streaming ./cmd/streaming

build-query:
	go build -o bin/query ./cmd/query

build-api:
	go build -o bin/api ./cmd/api

build-all: build-ingestor build-streaming build-query build-api

# ── Tests & lint ──────────────────────────────────────────────────────────

test:
	go test ./...

lint:
	golangci-lint run ./...

# ── Docker Compose ────────────────────────────────────────────────────────

## Start the full data platform stack (infra + services)
docker-compose-up:
	docker compose -f infra/docker/data-platform-compose.yml up -d

docker-compose-down:
	docker compose -f infra/docker/data-platform-compose.yml down

## Create Kafka topics (run after kafka is healthy)
topics:
	bash infra/kafka-topics.sh localhost:9092

# ── Go module tidy ────────────────────────────────────────────────────────

tidy:
	go mod tidy
