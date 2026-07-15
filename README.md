# Turgorepo

A full-stack monorepo template — **Tur**bo + **Go** + repo — with a Next.js frontend, Go/Gin API, PostgreSQL, and shared TypeScript packages for types and API docs.

**Author:** [Suprim Khatri](https://github.com/suprimkhatri77)

## Stack

| Layer | Tech |
| --- | --- |
| Monorepo | [Turborepo](https://turborepo.dev) + [Bun](https://bun.sh) workspaces |
| Frontend | [Next.js 16](https://nextjs.org), React 19, TanStack Query, Zustand, Tailwind CSS 4 |
| Backend | [Go](https://go.dev) + [Gin](https://gin-gonic.com), [sqlc](https://sqlc.dev), [golang-migrate](https://github.com/golang-migrate/migrate) |
| Database | PostgreSQL 17 |
| Auth | JWT in HTTP-only cookies (access + refresh), session tokens in DB |
| API docs | Zod schemas → OpenAPI 3 → [Scalar](https://scalar.com) UI |

## Project structure

```text
turgorepo/
├── apps/
│   ├── api/          # Go/Gin REST API
│   └── web/          # Next.js frontend
├── packages/
│   ├── types/        # Shared Zod schemas + TypeScript types
│   ├── openapi/      # OpenAPI spec generation (outputs to apps/api/openapi.json)
│   ├── ui/           # Shared React components
│   ├── eslint-config/
│   └── typescript-config/
├── docker-compose.dev.yml
└── .env.example
```

## Prerequisites

- [Bun](https://bun.sh) >= 1.3
- [Go](https://go.dev) >= 1.23
- [Docker](https://www.docker.com) & Docker Compose (for containerized dev)
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI (for local migrations)

## Getting started

### 1. Clone and install

```sh
git clone <your-repo-url> turgorepo
cd turgorepo
bun install
```

### 2. Environment

Copy the example env and adjust as needed:

```sh
cp .env.example .env.local
```

Key variables:

| Variable | Description |
| --- | --- |
| `DATABASE_URL` | PostgreSQL connection string |
| `JWT_ACCESS_SECRET` / `JWT_REFRESH_SECRET` | JWT signing secrets |
| `FRONTEND_URL` | Allowed CORS origin (e.g. `http://localhost:3000`) |
| `COOKIE_DOMAIN` | Cookie domain (e.g. `localhost`) |
| `NEXT_PUBLIC_API_URL` | API URL for the browser (e.g. `http://localhost:5000`) |
| `INTERNAL_API_URL` | API URL inside Docker network (e.g. `http://api:5000`) |

### 3. Run with Docker (recommended)

```sh
docker compose -f docker-compose.dev.yml up --build
```

| Service | URL |
| --- | --- |
| Web | <http://localhost:3000> |
| API | <http://localhost:5000> |
| API docs (Scalar) | <http://localhost:5000/api/v1/docs/> |
| Health check | <http://localhost:5000/api/v1/health> |
| PostgreSQL | localhost:5432 |

Run migrations inside the API container or locally:

```sh
cd apps/api
make migrate-up
```

### 4. Run locally (without Docker)

**Database** — start Postgres and set `DATABASE_URL` in `.env.local`.

**API:**

```sh
cd apps/api
make migrate-up   # first time
make run          # or: go run ./cmd/server
```

**Web:**

```sh
bun run dev --filter=web
```

**Everything via Turbo:**

```sh
bun run dev
```

## API

Base path: `/api/v1`

### Auth routes

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/health` | Health check |
| `POST` | `/auth/register` | Create account, set cookies |
| `POST` | `/auth/login` | Sign in, set cookies |
| `POST` | `/auth/logout` | Revoke session, clear cookies |
| `POST` | `/auth/refresh` | Refresh access token (rotates refresh token after 5 min) |
| `GET` | `/auth/me` | Current user (requires auth) |

Auth uses HTTP-only cookies:

- `access_token` — 15-minute JWT
- `refresh_token` — 30-day JWT, hashed and stored in DB
- `is_logged_in` — public flag for the frontend

User roles: `superadmin`, `admin`, `staff`, `member`

### API docs

Generate the OpenAPI spec from shared Zod schemas:

```sh
bun run generate
```

This writes `apps/api/openapi.json`. Scalar docs are served at `/api/v1/docs/`.

### Database workflow

```sh
cd apps/api

# Run migrations
make migrate-up
make migrate-down N=1   # rollback last N migrations

# Regenerate sqlc types after changing queries/schema
sqlc generate
```

Source layout:

- `migrations/` — SQL migrations
- `internal/database/schema/` — table definitions for sqlc
- `internal/database/queries/` — SQL queries
- `internal/database/generated/` — sqlc output (do not edit)

## Frontend (`apps/web`)

Next.js app with cookie-based auth, axios interceptors for token refresh, and route protection via Next.js proxy middleware.

### Proxy middleware (`_proxy.ts`)

The file is intentionally named `_proxy.ts` (underscore prefix disables it). When you need route middleware, rename it to `proxy.ts` and it works as-is.

It handles:

- `/auth/*` — redirect authenticated users away from login/register
- `/admin/*` — require `admin` or `superadmin` role

Route rules live in `apps/web/lib/middleware/config.ts`.

### Shared types

Import API types from the monorepo package:

```ts
import { LoginBodySchema, type User } from "@repo/types";
```

## Shared packages

### `@repo/types`

Zod schemas and inferred TypeScript types, organized by domain:

```text
src/
├── api/       # response wrappers, error codes
├── auth/      # login, register, auth responses
└── user/      # user model
```

### `@repo/openapi`

OpenAPI spec built from `@repo/types`. Each route lives in its own file:

```text
src/
├── schema.ts          # entry point
├── schemas/           # OpenAPI component schemas
└── paths/
    ├── health.ts
    └── auth/          # login.ts, register.ts, logout.ts, ...
```

## Scripts

| Command | Description |
| --- | --- |
| `bun run dev` | Start all apps in dev mode |
| `bun run build` | Build all apps and packages |
| `bun run lint` | Lint across the monorepo |
| `bun run check-types` | TypeScript type checking |
| `bun run generate` | Generate OpenAPI spec |
| `bun run format` | Format with Prettier |

Filter to a single app:

```sh
bun run dev --filter=web
bun run dev --filter=api
```

## Backend conventions

### Request logging

Auth handlers use `internal/packages/rlog` for structured logging. It automatically attaches `path`, `method`, `ip`, and `actor_id` (when available) to every log line:

```go
rlog.Info(c, "login successful", "user_id", user.ID)
rlog.Warn(c, "invalid credentials (user not found)")
rlog.Error(c, "failed to fetch user", err)
```

### Hot reload

The API uses [Air](https://github.com/air-verse/air) in Docker (`apps/api/.air.toml`). Local dev:

```sh
cd apps/api
go run github.com/air-verse/air@latest
```

## License

This project is licensed under the [MIT License](LICENSE).
