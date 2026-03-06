
# AGENTS.md

Rules for AI agents working with the FastGo project.

The purpose of this file is to keep the architecture simple, predictable, and consistent.

Agents MUST follow these rules when generating or modifying code.

---

# Philosophy

FastGo is an opinionated Go API starter focused on:

- simplicity
- fast development
- minimal boilerplate
- clear structure
- avoiding unnecessary abstractions

FastGo intentionally avoids:

- dependency injection frameworks
- large dependency containers
- heavy application frameworks
- complex module systems

The goal is to start writing business logic immediately.

---

# Core Stack

Language: Go
HTTP: Fiber
ORM: GORM
Cache: Redis
Config: .env via godotenv

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
        redis/

    http/
        <feature>/
            handlers/
            services/
            dto/
            routes.go

    models/
        (domain / database models)

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

Agents must not introduce new architectural layers without instruction.

---

# Infrastructure Access

Infrastructure services are accessed through package singletons.

Examples:

database.DB()
redis.Client()

Agents must NOT introduce dependency injection containers.

For simple Redis-based query/response caching, prefer the shared typed wrapper in:

internal/infra/redis

Do not re-implement ad-hoc JSON marshal/get/set cache helpers in feature services if the shared wrapper can solve the task.

---

# HTTP Feature Structure

Each HTTP feature lives in:

internal/http/<feature>/

Structure:

handlers/
services/
dto/
routes.go

Example:

internal/http/probes/
handlers/
services/
dto/
routes.go

---

# DTO

DTO = Data Transfer Object.

DTOs are used for:

- HTTP request bodies
- HTTP responses
- transport structures

DTOs belong in:

internal/http/<feature>/dto/

Example:

type LoginRequest struct {
Email string `json:"email"`
Password string `json:"password"`
}

DTOs must NOT contain business logic.

---

# Models

Models represent domain entities or database records.

Models belong in:

internal/models/

Example:

type User struct {
ID uint
Email string
}

Models may include:

- GORM tags
- domain fields
- database relations
- nested relations across multiple levels

Models represent real business entities.

Agents should model relational data explicitly.

If the domain contains nested entities, the project prefers multi-level GORM models with proper relations over flattened structures or manual join mapping.

---

# Important Rule

If a struct represents:

HTTP transport → dto/

Domain or database entity → models/

Example:

LoginRequest → dto/
User → models/

---

# Handlers

Handlers contain ONLY HTTP logic.

Responsibilities:

- read request
- call service
- return response

Handlers must NOT contain business logic.

---

# Services

Services contain business logic.

Services may use:

- models
- infra packages
- repositories
- other services

---

# Query Rules

Relational data must be loaded through GORM models and `Preload` / nested `Preload` chains by default.

Preferred approach:

- `database.DB().Preload("Author").Preload("Comments.User")`
- explicit model relations
- readable ORM queries

Avoid:

- raw SQL
- manual join scanning into ad-hoc structs
- bypassing model relations when `Preload` can solve the task

Raw SQL is allowed only if it is explicitly requested or if GORM cannot express the query cleanly.

---

# Routes

Each feature exposes:

func RegisterRoutes(r fiber.Router)

Routes are registered in:

bootstrap/app.go

Example:

api := app.Group("/api")

probesGroup := api.Group("/probes")
probes.RegisterRoutes(probesGroup)

---

# Configuration

Configuration is loaded from `.env`.

Example:

DB_DSN=postgres://user:pass@localhost:5432/app
REDIS_URL=localhost:6379
APP_PORT=8080

Config is loaded via:

config.Load()

Agents must not read environment variables randomly across packages if config.Load() exists.

---

# Translations

The project has i18n dictionaries in:

locales/<lang>/*.json

Examples:

- locales/en/errors.json
- locales/ru/errors.json

All user-facing API text must go through translations.

This includes:

- error messages
- validation messages
- other response text visible to clients

When adding or changing user-facing text:

- update both locales/en and locales/ru
- use translation keys, not hardcoded response strings
- keep wording clean and production-ready

---

# API Errors

FastGo is an API project.

Errors must be returned as JSON, not default HTML or plain text responses.

Keep API errors consistent and machine-readable.

Preferred shape:

{
  "error": {
    "code": "invalid_request_body",
    "message": "Invalid JSON request body",
    "details": {}
  }
}

Rules:

- use a shared error format
- keep `error.code` stable for clients
- localize `error.message`
- use `error.details` only when it adds useful context
- never leak raw database, ORM, or parser errors to clients

---

# Code Style

Prefer:

- small files
- clear package names
- direct logic
- readability over abstraction

Avoid:

- unnecessary interfaces
- giant dependency containers
- hidden magic
- overengineering

---

# Core Principle

FastGo favors:

- simplicity
- speed of development
- structural clarity
