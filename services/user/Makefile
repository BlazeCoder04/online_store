tidy:
	go mod tidy

run:
	go run cmd/app/main.go

# Docker
dev:
	docker-compose up -d

prod:
	docker-compose up

down:
	docker-compose down

down-full:
	docker-compose down -v

# Migrations
include .env

migrate:
	migrate create -ext sql -dir migrations -seq init

migrate-up:
	migrate -path migrations -database \
	"postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=${POSTGRES_SSLMODE}" \
	-verbose up

migrate-down:
	migrate -path migrations -database \
	"postgresql://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=${POSTGRES_SSLMODE}" \
	-verbose down 
