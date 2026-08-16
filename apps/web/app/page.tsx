import Link from "next/link";
import { get_api_url } from "@/utils/get-api-url";

type HealthStatus = {
  online: boolean;
  message: string;
};

async function checkHealth(): Promise<HealthStatus> {
  const api = get_api_url();
  if (!api) return { online: false, message: "API URL is not configured" };

  try {
    const response = await fetch(`${api}/api/v1/health`, {
      cache: "no-store",
    });
    if (!response.ok) {
      return { online: false, message: `API responded ${response.status}` };
    }

    const body = (await response.json()) as { message?: string };
    return {
      online: true,
      message: body.message ?? "Server is up and running",
    };
  } catch {
    // API may be offline during local/CI builds
    return { online: false, message: "API unreachable" };
  }
}

const stack = [
  { label: "Monorepo", value: "Turborepo + Bun workspaces" },
  { label: "Frontend", value: "Next.js 16, React 19, TanStack Query" },
  { label: "Backend", value: "Go + Gin, sqlc, golang-migrate" },
  { label: "Database", value: "PostgreSQL 17" },
  { label: "Auth", value: "JWT access + refresh in HTTP-only cookies" },
  { label: "Contracts", value: "Zod → OpenAPI → typed API client" },
];

export default async function Home() {
  const health = await checkHealth();
  const publicApi = process.env.NEXT_PUBLIC_API_URL;

  return (
    <main className="mx-auto flex w-full max-w-4xl flex-1 flex-col gap-10 px-6 py-16">
      <header className="space-y-4">
        <div className="flex items-center gap-2 text-xs text-muted-foreground">
          <span
            aria-hidden
            className={`size-2 rounded-full ${
              health.online ? "bg-emerald-500" : "bg-destructive"
            }`}
          />
          <span>
            API {health.online ? "online" : "offline"} — {health.message}
          </span>
        </div>

        <h1 className="text-3xl font-semibold tracking-tight sm:text-4xl">
          Turgorepo
        </h1>
        <p className="max-w-2xl text-sm leading-relaxed text-muted-foreground">
          A full-stack starter that pairs a Go/Gin API with a Next.js frontend.
          Shared Zod schemas generate the OpenAPI spec, and the spec generates a
          fully typed API client — tRPC-like DX without a TypeScript backend.
        </p>

        <div className="flex flex-wrap gap-3 pt-1">
          <Link
            className="inline-flex h-9 items-center border border-transparent bg-primary px-4 text-xs font-medium text-primary-foreground transition hover:bg-primary/80"
            href="/auth/register"
          >
            Create an account
          </Link>
          <Link
            className="inline-flex h-9 items-center border border-border px-4 text-xs font-medium transition hover:bg-muted"
            href="/auth/login"
          >
            Sign in
          </Link>
          {publicApi ? (
            <a
              className="inline-flex h-9 items-center border border-border px-4 text-xs font-medium transition hover:bg-muted"
              href={`${publicApi}/api/v1/docs/`}
              target="_blank"
              rel="noreferrer"
            >
              API docs
            </a>
          ) : null}
        </div>
      </header>

      <section className="border border-border">
        <h2 className="border-b border-border px-4 py-3 text-xs uppercase tracking-widest text-muted-foreground">
          Stack
        </h2>
        <dl className="divide-y divide-border">
          {stack.map((item) => (
            <div
              key={item.label}
              className="flex flex-col gap-1 px-4 py-3 sm:flex-row sm:items-center sm:gap-6"
            >
              <dt className="w-40 shrink-0 text-xs uppercase tracking-widest text-muted-foreground">
                {item.label}
              </dt>
              <dd className="text-sm">{item.value}</dd>
            </div>
          ))}
        </dl>
      </section>

      <section className="space-y-3">
        <h2 className="text-xs uppercase tracking-widest text-muted-foreground">
          Get started
        </h2>
        <pre className="overflow-x-auto border border-border bg-muted/40 p-4 text-xs leading-relaxed">
          {`bun install
cp .env.example .env.local
bun run docker:dev:up
bun run db:migrate && bun run db:seed`}
        </pre>
      </section>
    </main>
  );
}
