# FastGo

FastGo is a lean Go API starter for teams that want to start shipping business logic immediately.

It keeps the stack practical: Fiber v3, GORM, Redis, Swagger, i18n, JSON API errors, feature-first HTTP structure, Docker, Docker Compose, `air`, `dbresolver`, and a small amount of tooling that actually helps.

The goal is not to be a framework.
The goal is to remove the boring setup without dragging in unnecessary architecture.

## What You Get

- Fiber v3 HTTP server
- env-based config with automatic `.env` loading
- PostgreSQL via GORM
- optional read replicas via GORM `dbresolver`
- Redis client and typed JSON cache helper
- JSON API error flow with stable `error.code`
- translations in `locales/en` and `locales/ru`
- graceful shutdown with DB and Redis cleanup
- grouped dev logs with Fiber-style request headers and nested SQL/cache lines
- JSON logs in production
- Swagger UI and generated OpenAPI docs
- optional Prometheus `/metrics`
- `air` for hot reload
- `docker-compose.yml` for app + Postgres primary/replica + Redis
- `Makefile` for daily commands
- `golangci-lint` config and GitHub Actions CI

## Philosophy

FastGo prefers:

- simple bootstrap
- readable structure
- feature-first HTTP organization
- shared infra through package accessors
- GORM models and `Preload` over raw SQL
- enough tooling for real work, without ceremony

FastGo avoids:

- DI containers
- heavy application frameworks
- speculative repository layers
- codegen-first architecture
- boilerplate for its own sake

## Structure

```text
cmd/
  api/
    main.go

internal/
  bootstrap/
  config/
  http/
    metrics/
    probes/
    tests/
  i18n/
  infra/
    database/
    redis/
  models/
  shared/

locales/
  en/
  ru/
```

HTTP features live in:

```text
internal/http/<feature>/
  dto/
  handlers/
  services/
  routes.go
```

Domain and database models live in:

```text
internal/models/
```

## Configuration

Local config is loaded from `.env` automatically.

Base variables:

```env
APP_NAME=FastGo
APP_ENV=development
APP_PORT=3005
ENABLE_SWAGGER=true
ENABLE_METRICS=true
DB_DSN=postgres://postgres:pass@127.0.0.1:5432/app
DB_LOG_LEVEL=info
REDIS_URL=redis://127.0.0.1:6379/0
```

Optional database resolver and pool tuning:

```env
DB_READ_DSNS=postgres://postgres:pass@127.0.0.1:5433/app,postgres://postgres:pass@127.0.0.1:5434/app
DB_MAX_IDLE_CONNS=10
DB_MAX_OPEN_CONNS=50
DB_CONN_MAX_LIFETIME=1h
DB_CONN_MAX_IDLE_TIME=15m
```

Notes:

- `APP_NAME` is used for the Fiber app name and Swagger title.
- `APP_ENV` controls runtime behavior. Use `development` locally and `production` in containers.
- `ENABLE_SWAGGER` controls `/docs` and `/api/docs/swagger.json`. By default it is enabled in development and disabled in production.
- `ENABLE_METRICS` controls the Prometheus `/metrics` endpoint.
- `DB_DSN` is the primary write connection.
- `DB_READ_DSNS` is optional. If set, reads go to replicas and writes stay on the primary.
- `DB_LOG_LEVEL` controls SQL logging: `info|warn|error|silent`.
- `REDIS_URL` accepts full Redis URLs like `redis://:password@127.0.0.1:6379/2`. Plain `host:port` still works and uses DB `0`.

## Local Development

Start infrastructure only:

```bash
make infra-up
```

Run the app with hot reload:

```bash
make dev
```

Or run it directly:

```bash
make run
```

Generate Swagger docs:

```bash
make docs
```

Useful day-to-day commands:

```bash
make test
make lint
make build
make down
```

## Docker Compose

The project ships with a local stack:

- app
- PostgreSQL primary
- PostgreSQL replica
- Redis

Start everything:

```bash
make up
```

Or with raw Docker Compose:

```bash
docker compose up -d --build
```

Services:

- app: `http://127.0.0.1:3005`
- postgres primary: `127.0.0.1:5432`
- postgres replica: `127.0.0.1:5433`
- redis: `127.0.0.1:6379`

## Redis Cache Helper

For simple API/query caching, use the shared typed wrapper from `internal/infra/redis`.

```go
var numbersCache = redis.NewJSONCache[dto.ListResponse]("tests:numbers", 30*time.Second)

result, err := numbersCache.Remember(ctx, "list", func(ctx context.Context) (dto.ListResponse, error) {
	return loadNumbersFromDB(ctx)
})

if err != nil {
	return err
}

numbersCache.InvalidateAllBestEffort(ctx)
```

The helper handles:

- JSON encode/decode
- stable key namespacing
- TTL on write
- invalidation by prefix
- fail-open behavior, so Redis issues do not break DB-backed API responses

## Logging

HTTP request logs use the standard Fiber logger middleware:

- `development`: real Fiber-style colored header, grouped into a single block per request
- `production`: compact JSON lines
- health, metrics and Swagger paths are skipped to reduce noise

Example development block:

```text
┌ 22:05:44 | 200 |   80.136083ms |       127.0.0.1 | POST    | /v2/auth/refresh -
│ [cache:miss] tests:numbers:list
│ [sql 1.531ms] [rows:100] SELECT * FROM "numbers" ORDER BY "number" ASC
└
```

Database logs are separate and come from GORM:

- `development`: colorized SQL with highlighted keywords
- `production`: default level is `warn`
- override with `DB_LOG_LEVEL`

Redis cache logs are emitted in development and grouped inside the same request block:

- `cache:hit`
- `cache:miss`
- `cache:set`
- `cache:invalidate`

## API

Current endpoints:

```text
GET    /docs
GET    /api/docs/swagger.json
GET    /metrics

GET    /ping
GET    /health
GET    /ready

POST   /api/v1/t/numbers/range
GET    /api/v1/t/numbers
GET    /api/v1/t/numbers/random
DELETE /api/v1/t/numbers?numbers=1,2,3
DELETE /api/v1/t/numbers/clear
```

Example request:

```bash
curl -X POST http://localhost:3005/api/v1/t/numbers/range \
  -H 'Content-Type: application/json' \
  -d '{"from":1,"to":10}'
```

Success responses:

```json
{
  "data": {}
}
```

Error responses:

```json
{
  "error": {
    "code": "invalid_request_body",
    "message": "Invalid JSON request body",
    "details": {}
  }
}
```

`error.message` is localized through `locales/en` and `locales/ru`.

## Swagger

Generated artifacts live in:

```text
docs/swagger.json
docs/swagger.yaml
```

If you change handlers, DTOs, or annotations:

```bash
make docs
```

To verify docs are committed and up to date:

```bash
make docs-check
```

## Docker

Build the image:

```bash
make docker-build
```

Or directly:

```bash
docker build -t fastgo:local .
```

Run it:

```bash
docker run --rm -p 3005:3005 \
  -e DB_DSN='postgres://postgres:pass@host.docker.internal:5432/app?sslmode=disable' \
  -e REDIS_URL='redis://host.docker.internal:6379/0' \
  fastgo:local
```

The Docker image:

- builds Swagger docs during image build
- defaults to `APP_ENV=production`
- disables Swagger by default
- exposes `/health` via `HEALTHCHECK`

## Tooling

This starter includes:

- `.air.toml` for hot reload
- `Makefile` for common workflows
- `.golangci.yml` for linting
- GitHub Actions CI in `.github/workflows/ci.yml`

Helpful commands:

```bash
make help
make fmt
make test
make lint
make docs
make build
make check
```

## Testing

Run all tests:

```bash
make test
```

## Notes for Extending the Starter

- Keep handlers thin.
- Put HTTP DTOs in `dto/`.
- Put database/domain entities in `internal/models/`.
- Prefer GORM models with relations and `Preload`.
- Use translations for user-facing API text.
- Keep shared layers minimal.
