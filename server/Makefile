.PHONY: build run clean test dev

VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo "dev")

build:
	go build -ldflags "-X ai-model-proxy/res.Version=$(VERSION)" -o bin/ai-model-proxy ./cmd/server/main.go

dev:
	go run ./cmd/server/main.go

clean:
	rm -rf bin

test:
	go test ./...

.DEFAULT_GOAL := build
