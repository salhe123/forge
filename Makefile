APP=forge
GO=go

.PHONY: tidy test run docker-up docker-down

tidy:
	$(GO) mod tidy

test:
	$(GO) test ./...

run:
	DATABASE_URL=$${DATABASE_URL:-postgres://forge:forge@localhost:5432/forge?sslmode=disable} \
	HTTP_ADDR=$${HTTP_ADDR:-:8080} \
	$(GO) run ./cmd/api

docker-up:
	docker compose up --build

docker-down:
	docker compose down
