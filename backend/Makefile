.PHONY: tidy run build test vet migrate admin up

tidy:    ## обновить go.mod/go.sum (нужен интернет)
	go mod tidy

run:     ## запуск API; новые миграции применяются при старте (переменные — см. README.md)
	go run ./cmd/api

build:
	go build -o bin/api ./cmd/api
	go build -o bin/migrate ./cmd/migrate
	go build -o bin/createadmin ./cmd/createadmin

vet:
	gofmt -l . && go vet ./...

test:
	go test ./...

migrate: ## применить новые миграции к $$DATABASE_URL, не запуская сервер
	go run ./cmd/migrate

admin:   ## создать первого администратора: ADMIN_EMAIL, ADMIN_PASSWORD
	go run ./cmd/createadmin

up:      ## docker compose с БД и API
	docker compose up --build
