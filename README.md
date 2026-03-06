# FastGo

Opinionated Go starter for developers who want to launch APIs quickly instead of spending the first days building project scaffolding.

FastGo is a minimal, fast, and readable backend project template for Go.
It provides the essential foundation needed to start building an API immediately without dealing with dependency injection frameworks, complex bootstrapping, or architectural overengineering.

The goal of FastGo is simple: start building business logic immediately.

---

# Philosophy

FastGo follows a few simple principles.

### Simplicity over abstraction
Avoid unnecessary layers, complex DI containers, and magic.

### Readable structure
The project structure should be obvious to any Go developer.

### Fast development
You should be able to start writing real endpoints in minutes.

### No framework lock-in
FastGo is not a framework. It is a starter structure.

### Infrastructure without DI chaos
Infrastructure like database or Redis is accessed via simple package accessors instead of large dependency containers.

---

# What FastGo provides

FastGo gives you the basic backend foundation out of the box:

- Fiber HTTP server
- env-based configuration
- PostgreSQL connection (via GORM)
- Redis client
- health / readiness probes
- simple HTTP feature structure
- clear bootstrap system
- minimal shared utilities
- migrations directory

No heavy frameworks. No code generators. No architectural ceremony.

---

# Project Structure

internal/
    bootstrap/
        app.go
        providers.go

    config/
        config.go

    infra/
        database/
            init.go
            db.go
            transaction.go

        redis/
            init.go
            redis.go

    http/
        <feature>/
            handlers/
            services/
            models/
            routes.go

    shared/
        response/
        errors/
        logger/

database/
    migrations/
    seeders/

cmd/
    api/
        main.go

---

# HTTP Feature Structure

Each HTTP feature lives inside its own directory:

internal/http/<feature>/

Example:

internal/http/probes/
    handlers/
        probes_handler.go

    services/
        probes_service.go

    models/
        probe_response.go

    routes.go

Responsibilities:

Handlers:
- parse request
- call service
- return response

Services:
- business logic

Models:
- request/response structures

Routes:
- register endpoints for the feature

---

# Infrastructure

Infrastructure is intentionally simple.

Examples:

database.DB()
redis.Client()

Infrastructure is initialized during application bootstrap inside:

bootstrap/providers.go

---

# Getting Started

1. Clone the repository

git clone https://github.com/yourname/fastgo

2. Create .env

DB_DSN=postgres://user:pass@localhost:5432/app
REDIS_URL=localhost:6379
APP_PORT=8080

3. Run the server

go run cmd/api/main.go

---

# Health Endpoints

GET /api/probes/ping
GET /api/probes/health
GET /api/probes/ready

---

# When to use FastGo

FastGo is ideal for:

- REST APIs
- SaaS backends
- microservices
- MVP products
- internal tools

It is designed for teams that prefer clarity and speed over architectural ceremony.

---

# What FastGo intentionally avoids

FastGo does NOT include:

- dependency injection frameworks
- heavy application containers
- auto-discovery systems
- complex module loaders
- excessive abstractions

You can add these later if needed.

FastGo simply does not force them on you.

---

# License

MIT
