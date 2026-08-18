APP=forge
GO=go

.PHONY: tidy test run cli install docker-up docker-down

tidy:
	$(GO) mod tidy

test:
	$(GO) test ./...

run:
	DATABASE_URL=$${DATABASE_URL:-postgres://forge:forge@localhost:5432/forge?sslmode=disable} \
	HTTP_ADDR=$${HTTP_ADDR:-:8080} \
	$(GO) run ./cmd/api

cli:
	$(GO) run ./cmd/forge $(ARGS)

install:
	$(GO) install ./cmd/forge


docker-down:
	docker compose down
