.PHONY: generate build migrate-up migrate-down test

generate:
	go generate ./...

build: generate
	go build -o ./main ./cmd/server

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down

test:
	go test ./...
