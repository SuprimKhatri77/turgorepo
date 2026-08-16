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
| Typed API client | OpenAPI → `@repo/api-client` ([@hey-api/openapi-ts](https://heyapi.dev)) |

## Project structure

```text
turgorepo/
├── apps/
│   ├── api/          # Go/Gin REST API
│   └── web/          # Next.js frontend
├── packages/
│   ├── types/        # Shared Zod schemas + TypeScript types
│   ├── openapi/      # OpenAPI spec generation (outputs to apps/api/openapi.json)
│   ├── api-client/   # Typed TS client generated from OpenAPI (tRPC-like DX for Go)
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
- [golangci-lint](https://golangci-lint.run/welcome/install/) (for Go linting)

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
bun run docker:dev:up
# stop: bun run docker:dev:down
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

### `@repo/api-client`

Typed frontend SDK generated from `apps/api/openapi.json` with [@hey-api/openapi-ts](https://heyapi.dev). Same idea as tRPC (typed API calls) when the backend is Go, not TypeScript.

```sh
# Regenerates OpenAPI then the TS client (Turbo runs packages in dependency order)
bun run generate
```

Use from the web app via the wired client (reuses axios + refresh interceptors):

```ts
import {
  apiClient,
  postApiV1AuthLogin,
  getApiV1AuthMe,
} from "@/lib/api/client";

const { data } = await postApiV1AuthLogin({
  client: apiClient,
  body: { email, password },
  throwOnError: true,
});

const me = await getApiV1AuthMe({ client: apiClient, throwOnError: true });
```

After changing Zod schemas or OpenAPI path definitions, run `bun run generate` and commit `packages/api-client/src/generated`. `apps/api/openapi.json` is local/CI build output (gitignored) used for Scalar docs and client generation.

## Scripts

| Command | Description |
| --- | --- |
| `bun run dev` | Start all apps in dev mode |
| `bun run build` | Build all apps and packages |
| `bun run lint` | Lint JS/TS across the monorepo |
| `bun run lint:go` | Lint the Go API with golangci-lint |
| `bun run check-types` | TypeScript type checking (includes `web`) |
| `bun run generate` | Generate OpenAPI spec + typed `@repo/api-client` |
| `bun run generate:check` | Regenerate and fail if committed output is stale |
| `bun run db:migrate` | Run DB migrations up (`apps/api`) |
| `bun run db:migrate:down` | Roll back migrations (`N=1 bun run db:migrate:down`) |
| `bun run db:sqlc` | Regenerate sqlc Go code from SQL queries |
| `bun run db:sqlc:check` | Regenerate sqlc and fail if committed output is stale |
| `bun run db:seed` | Insert demo admin + member users (idempotent) |
| `bun run docker:dev:up` | Start Docker Compose stack with `.env.local` |
| `bun run docker:dev:down` | Stop Docker Compose stack |
| `bun run format` | Format with Prettier |
| `bun run prepush` | Typecheck + lint:go + build (same as the pre-push hook) |
| `bun run ci` | Full local CI: generate:check, sqlc:check, typecheck, lint, lint:go, build |

Filter to a single app:

```sh
bun run dev --filter=web
bun run dev --filter=api
```

## Developer tooling

### Git hooks (Husky + lint-staged)

`bun install` installs [Husky](https://typicode.github.io/husky/) via the `prepare` script.

**pre-commit** (lint-staged):

- **JS/TS in `apps/web` and `packages/ui`** — ESLint (`--fix`) + Prettier
- **`packages/types` or `packages/openapi`** — runs `bun run generate` and stages OpenAPI + api-client output
- **API SQL / sqlc config / migrations** — runs `bun run db:sqlc` and stages generated Go
- **Other JS/TS / JSON / YAML / CSS** — Prettier
- **Go** — `gofmt`

**commit-msg** — [Commitlint](https://commitlint.js.org/) enforces [Conventional Commits](https://www.conventionalcommits.org/):

```text
feat: add refresh token rotation
fix: validate user_id claim on logout
chore: bump golangci-lint config
```

**pre-push** — runs `bun run prepush` (`check-types` + `lint:go` + `build`) so the same quality gates as CI run before push. Skip hooks with `git push --no-verify` when you need to (use sparingly).

### GitHub Actions CI

PRs and pushes to `main` run `.github/workflows/ci.yml`:

- `generate:check` (OpenAPI + api-client must be committed and current)
- `db:sqlc:check` (sqlc generated Go must be committed and current)
- `check-types`
- `lint` (JS/TS)
- `lint:go` (golangci-lint)
- `build`

### Branch protection

`main` requires the CI check **Lint, typecheck, and build** before merge (configured via GitHub branch protection).

### Dependabot

Weekly update PRs are configured in `.github/dependabot.yml` for:

- Bun workspace deps (root)
- Go modules (`apps/api`)
- GitHub Actions

TypeScript **major** bumps are ignored until typescript-eslint supports them (e.g. TS 7).

### golangci-lint

Config lives at `apps/api/.golangci.yml` (sqlc generated code is excluded).

```sh
# install (v2 — note the /v2/ path; the Go extension's older suggestion is v1)
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
# or: brew install golangci-lint

# ensure the binary is on PATH (usually ~/go/bin)
export PATH="$(go env GOPATH)/bin:$PATH"

bun run lint:go
# or from apps/api:
make lint
```

If Cursor/VS Code still says the command is missing, reload the window after install. Workspace settings already point the Go extension at `${env:HOME}/go/bin/golangci-lint`.

### VS Code

Open the repo root in VS Code/Cursor. Recommended extensions are in `.vscode/extensions.json` (ESLint, Prettier, Go, Tailwind, Docker). Workspace settings enable format-on-save and golangci-lint on save for Go files.

## Backend conventions

### Request logging

Auth handlers use `internal/packages/rlog` for structured logging. It automatically attaches `request_id`, `path`, `method`, `ip`, and `actor_id` (when available) to every log line:

```go
rlog.Info(c, "login successful", "user_id", user.ID)
rlog.Warn(c, "invalid credentials (user not found)")
rlog.Error(c, "failed to fetch user", err)
```

Every response includes `X-Request-ID` (reuses the incoming header or mints a UUID). Pass the same header from the client when debugging.

### Database seed

After migrations, set seed passwords in `.env.local` (required — no defaults):

```sh
SEED_ADMIN_PASSWORD=your-local-admin-password
SEED_MEMBER_PASSWORD=your-local-member-password
bun run db:seed
```

Creates (if missing):

| Email | Role |
| --- | --- |
| `admin@example.com` | admin |
| `member@example.com` | member |

Optional overrides: `SEED_ADMIN_EMAIL`, `SEED_MEMBER_EMAIL`, `SEED_ADMIN_NAME`, `SEED_MEMBER_NAME` (see `.env.example`).

### Hot reload

The API uses [Air](https://github.com/air-verse/air) in Docker (`apps/api/.air.toml`). Local dev:

```sh
cd apps/api
go run github.com/air-verse/air@latest
```

## License

This project is licensed under the [MIT License](LICENSE).
