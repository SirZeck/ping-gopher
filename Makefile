.PHONY: build build-cli run-all run-api run-worker run-scheduler test docker-up docker-down docker-build clean

BINARY_NAME=pinggopher.exe
CLI_BINARY_NAME=pinggopher-cli.exe

build:
	go build -o bin/$(BINARY_NAME) ./cmd/pinggopher

build-cli:
	go build -o bin/$(CLI_BINARY_NAME) ./cmd/pinggopher-cli

run-all: build
	./bin/$(BINARY_NAME) --role=all

run-api: build
	./bin/$(BINARY_NAME) --role=api

run-worker: build
	./bin/$(BINARY_NAME) --role=worker

run-scheduler: build
	./bin/$(BINARY_NAME) --role=scheduler

test:
	go test -v ./...

docker-build:
	docker compose build

docker-up:
	docker compose up -d

docker-down:
	docker compose down

clean:
	rm -rf bin/ pinggopher.db
