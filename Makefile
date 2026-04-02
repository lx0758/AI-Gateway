.PHONY: build build-web build-server run clean test dev

VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo "dev")

build: build-server

build-web:
	cd web && VITE_APP_VERSION=$(VERSION) $(MAKE) build

build-server: build-web
	cd server && VERSION=$(VERSION) $(MAKE) build

dev:
	cd server && $(MAKE) dev &

clean:
	cd server && $(MAKE) clean
	cd web && $(MAKE) clean

.DEFAULT_GOAL := build
