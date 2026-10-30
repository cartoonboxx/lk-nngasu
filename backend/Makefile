.PHONY: tidy run build test vet migrate admin up

tidy:    ## создать/обновить go.sum (нужен интернет)
	go mod tidy

run:     ## запуск API (переменные из окружения, см. .env.example)
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api
	go build -o bin/createadmin ./cmd/createadmin

vet:
	gofmt -l . && go vet ./...

test:
	go test ./...

migrate: ## применить миграцию к $DATABASE_URL (для БД вне docker-compose)
	psql "$$DATABASE_URL" -v ON_ERROR_STOP=1 -f migrations/001_init.sql

admin:   ## создать первого администратора: ADMIN_EMAIL, ADMIN_PASSWORD
	go run ./cmd/createadmin

up:      ## docker compose с БД и API
	docker compose up --build
