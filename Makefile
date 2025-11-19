.PHONY: build run clean docker-up docker-down create-topic produce-test seed-db test test-unit test-integration generate-mocks

build:
	go build -o bin/order-service ./cmd/server

run:
	go run ./cmd/server

clean:
	rm -rf bin/

docker-up:
	docker compose up -d

docker-down:
	docker compose down

create-topic:
	docker exec -it order-service-kafka-1 kafka-topics --create --topic orders --partitions 1 --replication-factor 1 --bootstrap-server localhost:9092

produce-test:
	go run ./cmd/producer

seed-db:
	go run ./cmd/seed

generate-mocks:
	go generate ./internal/ports/...

test-unit: generate-mocks
	go test ./internal/... -tags="!integration"

test-integration: generate-mocks
	go test ./internal/... -tags="integration"
