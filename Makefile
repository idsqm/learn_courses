.PHONY: run build test docker-up docker-down migrate-up migrate-down migrate-create

DB_URL ?= postgres://courses:courses@localhost:5433/courses?sslmode=disable

run:
	docker compose up --build

build:
	docker compose build

test:
	go test ./...

docker-up:
	docker compose up -d

docker-down:
	docker compose down

migrate-up:
	goose -dir migrations postgres "$(DB_URL)" up

migrate-down:
	goose -dir migrations postgres "$(DB_URL)" down

migrate-create:
	goose -dir migrations create $(name) sql
