.PHONY: init build build-web build-server clean

VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo "dev")

all: build

init:
	cd web && $(MAKE) init
	cd server && $(MAKE) init

build: build-server

build-web:
	cd web && VITE_APP_VERSION=$(VERSION) $(MAKE) build

build-server: build-web
	cd server && VERSION=$(VERSION) $(MAKE) build

clean:
	cd server && $(MAKE) clean
	cd web && $(MAKE) clean

.DEFAULT_GOAL := build
