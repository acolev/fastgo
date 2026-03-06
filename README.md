# FastGo

FastGo is a lean Go API starter for teams that want to begin writing business logic immediately.

It gives you a clean feature-first structure, Fiber v3, GORM, Redis, i18n, JSON API errors, graceful shutdown, hot reload via `air`, and database read/write routing via GORM `dbresolver`.

The goal is not to be a framework.
The goal is to remove the boring setup work without dragging in unnecessary architecture.

## What You Get

- Fiber v3 HTTP server
- env-based config with local `.env` loading
- PostgreSQL via GORM
- optional read replicas via `dbresolver`
- Redis client
- typed Redis JSON cache helper for query/response caching
- graceful shutdown with provider cleanup
- JSON API error handling with stable `error.code`
- translations in `locales/en` and `locales/ru`
- feature-first HTTP structure
- health/readiness probes
- `air` config for local development
- Swagger UI and generated OpenAPI docs
- multi-stage Docker build

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
APP_PORT=3005
DB_DSN=postgres://postgres:pass@127.0.0.1:5432/app
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

`APP_NAME` is used for the Fiber app name and Swagger title.

`DB_DSN` is the primary write connection.

`DB_READ_DSNS` is optional and accepts a comma-separated list of replica DSNs.
If it is set, GORM `dbresolver` routes read queries to replicas and writes to the primary connection.

`REDIS_URL` accepts a full Redis URL like `redis://:password@127.0.0.1:6379/2`.
For backward compatibility, plain `host:port` is still accepted and uses Redis DB `0`.

## Local Development

Install `air`:

```bash
go install github.com/air-verse/air@latest
```

Start the app with hot reload:

```bash
air
```

Or run it directly:

```bash
go run cmd/api/main.go
```

Generate Swagger docs:

```bash
make docs
```

Swagger UI is available at:

```text
/docs
```

## Redis Cache Helper

For simple API/query caching, use the shared typed wrapper from `internal/infra/redis`.

Example:

```go
var numbersCache = redis.NewJSONCache[dto.ListResponse]("tests:numbers", 30*time.Second)

result, err := numbersCache.Remember(ctx, "list", func(ctx context.Context) (dto.ListResponse, error) {
	return loadNumbersFromDB(ctx)
})

if err := numbersCache.InvalidateAll(ctx); err != nil {
	return err
}
```

The helper handles:

- JSON encode/decode
- stable key namespacing
- TTL on write
- one-line cache invalidation by prefix

## API

Current endpoints:

```text
GET    /docs
GET    /docs/swagger.json

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

Success responses use:

```json
{
  "data": {}
}
```

Error responses use:

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

Swagger includes generated OpenAPI artifacts:

```text
docs/swagger.json
docs/swagger.yaml
```

If you change handlers, annotations, or request/response DTOs, regenerate docs with:

```bash
make docs
```

## Database and Redis Lifecycle

FastGo initializes database and Redis during bootstrap and closes both connections on shutdown.

The app includes:

- graceful Fiber shutdown
- database pool close
- Redis client close
- readiness checks for both providers

## Docker

Build the image:

```bash
docker build -t fastgo .
```

Run the container:

```bash
docker run --rm -p 3005:3005 --env-file .env fastgo
```

Make sure your `DB_DSN`, optional `DB_READ_DSNS`, and `REDIS_URL` point to services reachable from inside the container.
The Docker build generates Swagger docs automatically before compiling the binary.

## Testing

Run all tests:

```bash
go test ./...
```

## Notes for Extending the Starter

- Keep handlers thin.
- Put HTTP DTOs in `dto/`.
- Put database/domain entities in `internal/models/`.
- Prefer GORM models with relations and `Preload`.
- Use translations for user-facing API text.
- Keep shared layers minimal.
