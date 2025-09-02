.PHONY: build run test clean

build:
	go build -o order-service ./cmd/server

run:
	go run ./cmd/server

test:
	go test ./...

clean:
	rm -f order-service

docker-up:
	docker compose up -d

docker-down:
	docker compose down

create-topic:
	docker exec -it order-service-kafka-1 kafka-topics --create --topic orders --partitions 1 --replication-factor 1 --bootstrap-server localhost:9092

produce-test:
	go run ./cmd/producer
