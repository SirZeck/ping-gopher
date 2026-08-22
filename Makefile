.PHONY: build run-all run-api run-worker run-scheduler test clean

BINARY_NAME=pinggopher.exe

build:
	go build -o bin/$(BINARY_NAME) ./cmd/pinggopher

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

clean:
	rm -rf bin/ pinggopher.db
