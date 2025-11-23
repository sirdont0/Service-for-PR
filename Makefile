.PHONY: all lint test docker-up 

start: all docker-up

all: lint test

lint:
	golangci-lint run ./...

test: lint
	go test ./... -v

docker-up: lint test
	docker-compose up --build

