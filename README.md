# Turgorepo

A full-stack monorepo with a Next.js frontend, Go/Gin API, PostgreSQL, Redis, and shared TypeScript packages for types and API docs.

<!-- template:start -->
**Author:** [Suprim Khatri](https://github.com/suprimkhatri77)

## Start a new project from this template

Turgorepo is **Tur**bo + **Go** + repo. Click **Use this template** on GitHub, clone your copy, then run:

```sh
bun run init --name acme-app --module github.com/acme/acme-app
git add -A && git commit -m "chore: initialize acme-app"
```

This renames the project, Go module, Docker project and database, creates `.env.local` with generated JWT secrets, and removes itself along with this section. Then continue from [Run with Docker](#2-run-with-docker-recommended).

<!-- template:end -->
## Stack

| Layer | Tech |
| --- | --- |
| Monorepo | [Turborepo](https://turborepo.dev) + [Bun](https://bun.sh) workspaces |
| Frontend | [Next.js 16](https://nextjs.org), React 19, TanStack Query, Zustand, Tailwind CSS 4 |
| Backend | [Go](https://go.dev) + [Gin](https://gin-gonic.com), [sqlc](https://sqlc.dev), [golang-migrate](https://github.com/golang-migrate/migrate) |
| Database | PostgreSQL 17 |
| Cache / rate-limit ready | Redis 7 |
| Auth | JWT in HTTP-only cookies (access + refresh), session tokens in DB |
| API docs | Zod schemas → OpenAPI 3 → [Scalar](https://scalar.com) UI |
| Typed API client | OpenAPI → `@repo/api-client` (axios) + `@repo/api-client/server` (fetch) |

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
│   └── typescript-config/
├── docker-compose.dev.yml
└── .env.example
```

## Prerequisites

- [Bun](https://bun.sh) 1.4.4
- [Go](https://go.dev) 1.27.3
- [sqlc](https://sqlc.dev) 1.31.1 (CI pins this version, and a different one fails `db:sqlc:check`)
- [Docker](https://www.docker.com) & Docker Compose (for containerized dev)
- [golang-migrate](https://github.com/golang-migrate/migrate) CLI (for local migrations)
- [golangci-lint](https://golangci-lint.run/welcome/install/) (for Go linting)

## Getting started

### 1. Clone and configure

```sh
git clone <your-repo-url> turgorepo
cd turgorepo
bun install
cp .env.example .env.local
```

The defaults work as they are. Key variables:

| Variable | Description |
| --- | --- |
| `DATABASE_URL` | PostgreSQL connection string |
| `REDIS_URL` | Redis connection URL |
| `JWT_ACCESS_SECRET` / `JWT_REFRESH_SECRET` | JWT signing secrets. Release mode requires at least 32 bytes, and they must differ |
| `FRONTEND_URL` | Allowed CORS origin |
| `COOKIE_DOMAIN` | Cookie domain, used in release mode |
| `TRUSTED_PROXIES` | Comma-separated CIDRs of proxies allowed to set `X-Forwarded-For`. Leave empty without a proxy |
| `NEXT_PUBLIC_API_URL` | API URL for the browser, built into the web app |
| `INTERNAL_API_URL` | API URL for server-side web requests |
| `SMTP_*` / `EMAIL_FROM` | How email is sent. Locally this is Mailpit, and production requires it |
| `CLOUDINARY_*` | Optional. Set all three to enable image uploads |

Keep one `.env.local` at the repo root. It works both on your machine and in Docker, because `docker-compose.dev.yml` swaps in Docker hostnames where needed.

### 2. Run with Docker (recommended)

```sh
bun run docker:dev:up      # builds, starts everything, and waits until it's healthy
bun run db:seed            # optional: creates the demo users
```

Migrations run automatically when the API container starts.

| Service | URL |
| --- | --- |
| Web | <http://localhost:3000> |
| API | <http://localhost:5000> |
| API docs (Scalar) | <http://localhost:5000/api/v1/docs/> |
| Health check | <http://localhost:5000/api/v1/health> |
| PostgreSQL | localhost:5432 |
| Redis | localhost:6379 |
| Mailpit (local email inbox) | <http://localhost:8025> |

Follow logs with `bun run docker:dev:logs`, and stop with `bun run docker:dev:down`.

**Port already in use?** Set `WEB_PORT`, `API_PORT`, `DB_PORT`, `REDIS_PORT`, `MAILPIT_PORT` or `MAILPIT_SMTP_PORT` in `.env.local`, then update the matching URLs in the same file. Without Docker, the API listens on `PORT` instead of `API_PORT`. On macOS, AirPlay Receiver often holds port 5000.

### 3. Run without Docker

Start Postgres, Redis and Mailpit yourself, or run only those in Docker:

```sh
bun run docker:dev:services
bun run db:migrate         # first time, and after adding migrations
```

Then start the apps, together or one at a time:

```sh
bun run dev                # API and web
bun run dev --filter=web
bun run dev --filter=api
```

### 4. Deploy

Production images, and a guide for AWS ECS with an external database, are in [docs/deploy-aws.md](docs/deploy-aws.md).

```sh
bun run docker:build:api
NEXT_PUBLIC_API_URL=https://api.example.com bun run docker:build:web
```

## API

Base path: `/api/v1`

### Auth routes

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/health` | Liveness: the process is up. Use this for load balancers |
| `GET` | `/ready` | Readiness: the database and Redis are reachable. Use this for monitoring |
| `POST` | `/auth/register` | Create account, set cookies |
| `POST` | `/auth/login` | Sign in, set cookies |
| `POST` | `/auth/forgot-password` | Email a reset link. Responds the same whether or not the account exists |
| `POST` | `/auth/reset-password` | Set a new password with an emailed token, and log out on all devices |
| `POST` | `/auth/logout` | End this device's session and its access token, clear cookies |
| `POST` | `/auth/refresh` | Refresh access token and rotate the refresh token. The previous token keeps working for 5 minutes for concurrent requests, but only gets a new access token |
| `GET` | `/auth/me` | Current user (requires auth) |

Auth uses HTTP-only cookies:

- `access_token` — 15-minute JWT
- `refresh_token` — 30-day JWT, hashed and stored in DB
- `is_logged_in` — public flag for the frontend

User roles: `superadmin`, `admin`, `member`

### Password reset

Reset links expire after 1 hour and work once, and requesting a new one cancels the old one. Only a hash of the token is stored. The link carries the token in its URL fragment (`#token=...`), which browsers don't send to servers, so it stays out of access logs.

A reset logs the user out everywhere straight away. Sessions are revoked in the database, and Redis holds a cutoff time that makes `RequireAuth` reject older access tokens. The cutoff removes itself after 15 minutes, when those tokens would have expired anyway. If Redis is down, requests are allowed and the error is logged, so old access tokens can keep working until they expire.

Logging out ends that device's session straight away, including its access token, and the user's other devices stay logged in. Each access token carries its session ID (`sid`), and logout sets a Redis marker for that session that `RequireAuth` checks alongside the cutoff, in the same round trip. The marker also removes itself after 15 minutes. Reusing an old refresh token outside the grace window, which suggests it was stolen, ends the session the same way.

To read reset emails locally, open Mailpit at <http://localhost:8025>.

### API docs

Generate the OpenAPI spec from shared Zod schemas:

```sh
bun run generate
```

This writes `apps/api/openapi.json`, which is committed. Scalar docs are served at `/api/v1/docs/`.

### Database workflow

```sh
cd apps/api

# Run migrations
make migrate-up
make migrate-down N=1   # rollback last N migrations

# Regenerate sqlc types after changing queries/schema
make sqlc
```

Source layout:

- `migrations/` — SQL migrations, which sqlc also reads as the schema
- `internal/database/queries/` — SQL queries
- `internal/database/generated/` — sqlc output (do not edit)

## Frontend (`apps/web`)

Next.js app with cookie-based auth, axios interceptors for token refresh, and route protection via Next.js proxy middleware.

### Proxy middleware (`proxy.ts`)

It handles:

It runs on every page, and handles:

- `/auth/*` — redirect authenticated users away from login/register
- `/admin/*` — require `admin` or `superadmin` role
- `/dashboard/*` — require any logged-in user

To protect a new page, add it to `ROLE_RULES` in `apps/web/lib/middleware/config.ts`. The proxy only decodes tokens, so call `requireUser()` in the page as well, which verifies the session with the API. `apps/web/app/admin/page.tsx` is an example.

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

Typed SDK generated from `apps/api/openapi.json` with [@hey-api/openapi-ts](https://heyapi.dev). Same idea as tRPC (typed API calls) when the backend is Go, not TypeScript.

`bun run generate` produces **two** clients:

| Import | Transport | Use |
| --- | --- | --- |
| `@repo/api-client` | Axios | Browser (mutations, TanStack Query) — refresh interceptors |
| `@repo/api-client/server` | Fetch | RSC / server helpers — forward cookies, `INTERNAL_API_URL` |

```sh
# Regenerates OpenAPI then both TS clients (Turbo runs packages in dependency order)
bun run generate
```

**Browser** (reuses axios + refresh interceptors):

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

**Server Components** (cookie forwarding via `createServerApiClient`):

```ts
import {
  createServerApiClient,
  getApiV1Health,
  getApiV1AuthMe,
} from "@/lib/api/server-client";

const client = await createServerApiClient();
const health = await getApiV1Health({ client });

// Authoritative /me (DB). For cheap layout JWT claims, use getCurrentUser().
import { fetchCurrentUser } from "@/lib/api/auth/fetch-current-user";
const user = await fetchCurrentUser();
```

After changing Zod schemas or OpenAPI path definitions, run `bun run generate` and commit `apps/api/openapi.json`, `packages/api-client/src/generated` **and** `packages/api-client/src/generated-server`. The pre-commit hook does this for you.

## Scripts

| Command | Description |
| --- | --- |
| `bun run dev` | Start all apps in dev mode |
| `bun run build` | Build all apps and packages |
| `bun run lint` | Lint JS/TS across the monorepo |
| `bun run lint:go` | Lint the Go API with golangci-lint |
| `bun run check-types` | TypeScript type checking (includes `web`) |
| `bun run generate` | Generate OpenAPI spec + axios + fetch `@repo/api-client` outputs |
| `bun run generate:check` | Regenerate and fail if committed output is stale |
| `bun run db:migrate` | Run DB migrations up (`apps/api`) |
| `bun run db:migrate:down` | Roll back migrations (`N=1 bun run db:migrate:down`) |
| `bun run db:sqlc` | Regenerate sqlc Go code from SQL queries |
| `bun run db:sqlc:check` | Regenerate sqlc and fail if committed output is stale |
| `bun run db:seed` | Insert demo admin + member users (idempotent) |
| `bun run docker:dev:up` | Build and start the Docker stack, and wait until it's healthy |
| `bun run docker:dev:services` | Start only Postgres, Redis and Mailpit in Docker |
| `bun run docker:dev:logs` | Follow Docker stack logs |
| `bun run docker:dev:down` | Stop the Docker stack |
| `bun run docker:build:api` | Build the production API image |
| `bun run docker:build:web` | Build the production web image (needs `NEXT_PUBLIC_API_URL` set) |
| `bun run start` | Build and run both apps without Docker |
| `bun run format` | Format with Prettier |
| `bun run prepush` | Typecheck + lint:go + test:go + build (same as the pre-push hook) |
| `bun run ci` | Full local CI: generate:check, sqlc:check, typecheck, lint, lint:go, test:go, build |

Filter to a single app:

```sh
bun run dev --filter=web
bun run dev --filter=api
```

## Developer tooling

### Git hooks (Husky + lint-staged)

`bun install` installs [Husky](https://typicode.github.io/husky/) via the `prepare` script.

**pre-commit** (lint-staged):

- **JS/TS in `apps/web`** — ESLint (`--fix`) + Prettier
- **`packages/types` or `packages/openapi`** — runs `bun run generate` and stages both api-client outputs
- **API SQL / sqlc config / migrations** — runs `bun run db:sqlc` and stages generated Go
- **Other JS/TS / JSON / YAML / CSS** — Prettier
- **Go** — `gofmt`

**commit-msg** — [Commitlint](https://commitlint.js.org/) enforces [Conventional Commits](https://www.conventionalcommits.org/):

```text
feat: add refresh token rotation
fix: validate user_id claim on logout
chore: bump golangci-lint config
```

**pre-push** — runs `bun run prepush` (`check-types` + `lint:go` + `test:go` + `build`) so the same quality gates as CI run before push. Skip hooks with `git push --no-verify` when you need to (use sparingly).

### GitHub Actions CI

PRs and pushes to `main` run `.github/workflows/ci.yml`:

- `generate:check` (OpenAPI + api-client must be committed and current)
- `db:sqlc:check` (sqlc generated Go must be committed and current)
- `check-types`
- `lint` (JS/TS)
- `lint:go` (golangci-lint)
- `go test`
- `govulncheck` (Go dependency and standard library vulnerabilities)
- `build`

### Branch protection

`main` requires the CI check **Lint, typecheck, and build** before merge (configured via GitHub branch protection).

### GitGuardian

`.gitguardian.yaml` stops secret scanners flagging the local defaults in `.env.example` and `.env.production.example`. Those two files aren't scanned, so never put real values in them.

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

`.env.example` sets placeholder passwords for local use, and release mode refuses them:

```sh
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
go run github.com/air-verse/air@v1.67.4
```

## License

This project is licensed under the [MIT License](LICENSE).
