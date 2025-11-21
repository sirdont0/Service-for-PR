.PHONY: build run test docker-up migrate

build:
	go build ./...

run:
	go run ./cmd/server

test:
	go test ./... -v

docker-up:
	docker-compose up --build

migrate:
	docker-compose run --rm migrate
