.PHONY: init dev build clean

init:
	npm install

dev:
	npm run dev

build:
	npm run build -- --outDir ../server/res/web/

clean:
	rm -rf ../server/res/web/*
	touch ../server/res/web/.gitkeep

.DEFAULT_GOAL := build
