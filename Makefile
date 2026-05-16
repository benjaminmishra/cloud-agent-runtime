.PHONY: dev build run up

dev:
	go run ./backend

build:
	go build -o bin/rune-backend ./backend
	go build -o bin/rune ./cli

run: build
	./bin/rune-backend

up:
	docker-compose up --build
