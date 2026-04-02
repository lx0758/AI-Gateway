.PHONY: build run clean test dev

VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo "dev")

build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -ldflags '-linkmode external -extldflags "-static" -X ai-gateway/res.Version=$(VERSION)' -o bin/ai-gateway-server -v ./cmd/server/main.go

dev:
	go run ./cmd/server/main.go

clean:
	rm -rf bin

test:
	go test ./...

.DEFAULT_GOAL := build
