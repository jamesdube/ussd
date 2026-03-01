.PHONY: build test test-unit lint

build:
	go build ./...

test: test-unit

test-unit:
	gotestsum  --format-icons hivis -f testdox ./...

lint:
	golangci-lint run ./...
