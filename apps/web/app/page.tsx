import { getCurrentUser } from "@/lib/api/auth/auth";
import { createServerApiClient, getApiV1Health } from "@/lib/api/server-client";
import { LandingActions } from "@/modules/auth/components/landing-actions";

type HealthStatus = {
  online: boolean;
  message: string;
};

async function checkHealth(): Promise<HealthStatus> {
  try {
    const client = await createServerApiClient();

    const { data, error, response } = await getApiV1Health({
      client,
      cache: "no-store",
      signal: AbortSignal.timeout(5_000),
    });

    if (!response) {
      return { online: false, message: "API unreachable" };
    }

    if (error || !response.ok) {
      return {
        online: false,
        message: `API responded ${response.status}`,
      };
    }

    return {
      online: true,
      message: data?.message ?? "Server is up and running",
    };
  } catch (err) {
    if (
      err instanceof Error &&
      err.message.includes("API URL is not configured")
    ) {
      return { online: false, message: "API URL is not configured" };
    }
    // API may be offline during local/CI builds
    return { online: false, message: "API unreachable" };
  }
}

const stack = [
  { label: "Monorepo", value: "Turborepo + Bun workspaces" },
  { label: "Frontend", value: "Next.js 16, React 19, TanStack Query" },
  { label: "Backend", value: "Go + Gin, sqlc, golang-migrate" },
  { label: "Database", value: "PostgreSQL 17 + Redis 7" },
  { label: "Auth", value: "JWT access + refresh in HTTP-only cookies" },
  { label: "Contracts", value: "Zod → OpenAPI → axios + fetch typed clients" },
];

export default async function Home() {
  const health = await checkHealth();
  const user = await getCurrentUser();
  // The API turns docs off in production by default.
  const docsUrl =
    process.env.NODE_ENV !== "production" && process.env.NEXT_PUBLIC_API_URL
      ? `${process.env.NEXT_PUBLIC_API_URL}/api/v1/docs/`
      : null;

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
          <LandingActions user={user} />
          {docsUrl ? (
            <a
              className="inline-flex h-9 items-center border border-border px-4 text-xs font-medium transition hover:bg-muted"
              href={docsUrl}
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
bun run db:seed`}
        </pre>
      </section>
    </main>
  );
}
