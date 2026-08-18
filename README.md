# Forge

Mini internal deploy platform — **Go API + Postgres + Docker + CI**.

This is milestone 1: a production-shaped backend you can register apps against. Later: deploy worker, CLI, Kubernetes.

## Stack

| Layer | Choice |
|---|---|
| Language | Go 1.23 |
| HTTP | chi |
| Database | PostgreSQL 16 + pgx |
| Metrics | Prometheus `/metrics` |
| Logs | JSON `log/slog` |
| Ship | Docker, Compose, GitHub Actions |

## API

| Method | Path | Purpose |
|---|---|---|
| GET | `/health` | liveness |
| GET | `/ready` | Postgres ping |
| GET | `/metrics` | Prometheus |
| POST | `/v1/apps` | register an app |
| GET | `/v1/apps` | list apps |
| GET | `/v1/apps/{id}` | get one app |
| POST | `/v1/apps/{id}/deploy` | run the app image with Docker |

### Register an app

```bash
curl -s localhost:8080/v1/apps \
  -H 'content-type: application/json' \
  -d '{"name":"payments-api","repo_url":"https://github.com/salhe123/forge","image":"ghcr.io/salhe123/payments:latest"}'
```

## Run locally

```bash
cp .env.example .env
docker compose up postgres -d
make run
```

Or everything in Docker:

```bash
docker compose up --build
```

## Tests

```bash
make test
```

## Layout

```
cmd/api            HTTP process
internal/config    env config
internal/db        pool + schema
internal/apps      domain + postgres store
internal/httpserver  routes
.github/workflows  CI
```

### Deploy an app

Runs `docker run -d --name forge-<app>` using the stored `image`. Status goes `deploying` → `running` or `failed`.

Run the API with Compose (Docker socket is mounted so Forge can start containers on the host engine):

```bash
docker compose up --build
```

```bash
# after POST /v1/apps, use the returned id
curl -s -X POST localhost:8080/v1/apps/<id>/deploy
curl -s localhost:8080/v1/apps/<id>
docker ps --filter name=forge-
```

## Next (together)

1. `forge` CLI (`serve`, `apps list`, `deploy`)
2. Versioned SQL migrations
3. Helm chart + kind
