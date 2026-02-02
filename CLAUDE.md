# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Communication Guidelines

- Write all code and code comments in English
- Communicate with the user in Russian
- Maintain professional style in both languages

### Agents for Testing

**Always use `order-test-coordinator`** to write tests — it distributes work between unit and isolation agents in parallel, avoiding duplication.

Available agents:
- **`order-test-coordinator`** — **USE THIS** for writing tests (delegates to agents below)
- **`order-unit-test-writer`** — unit tests (mocked dependencies, validation, error handling)
- **`order-isolation-test-writer`** — isolation tests (real app, API black-box testing)

## Build & Run Commands

```bash
# Запуск всего (инфраструктура + приложение)
task run-all

# Запуск только приложения (требует PostgreSQL)
task run

# Сборка
task build

# Запуск тестов
task test

# Форматирование кода
task fmt

# Линтинг
task lint

# Default (fmt + lint + build)
task
```

## Infrastructure

```bash
# Запуск PostgreSQL (Docker)
task infra-up

# Остановка PostgreSQL (Docker)
task infra-down

# Запуск PostgreSQL (Podman)
task podman-infra-up

# Остановка PostgreSQL (Podman)
task podman-infra-down
```

## Environment Variables

- `DATABASE_URL` - PostgreSQL connection string (default: `postgres://postgres:postgres@localhost:5432/orders?sslmode=disable`)

## Architecture

This is an **Order Service** - a Connect RPC microservice for managing orders with PostgreSQL persistence.

### Layer Structure

```
cmd/app/main.go                    → Entry point, server setup
internal/domain/orders/server.go   → Service orchestrator, exposes RPC methods
internal/domain/orders/grpc_*_handler.go → Individual RPC handlers
internal/store/order.go            → PostgreSQL data access layer
```

### Key Patterns

- **Handler Pattern**: Each RPC method (CreateOrder, GetOrder, ListOrders, CheckOrderOwner) has a dedicated handler file
- **Interface-based Storage**: `OrderStore` interface in `internal/store/order.go` with `PostgresStore` implementation
- **Connect RPC Error Codes**: Use `connect.CodeInvalidArgument`, `connect.CodeNotFound`, `connect.CodePermissionDenied`, `connect.CodeInternal`

### Dependencies

- Proto contracts imported from `github.com/demo/contracts` (local replace directive in go.mod)
- Connect RPC framework (`connectrpc.com/connect`)
- PostgreSQL driver (`github.com/lib/pq`)

### Database

PostgreSQL with auto-created schema. Table `orders` has: `id`, `user_id`, `item`, `amount`, `created_at`.

Docker-compose configuration:
- Image: PostgreSQL 14.5-alpine
- Port: 5432
- Credentials: postgres/postgres
- Database: orders
