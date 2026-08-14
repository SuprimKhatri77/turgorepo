#!/usr/bin/env bun
/**
 * Runs sqlc generate, then fails if generated Go would change.
 * Keeps apps/api/internal/database/generated in sync with queries + migrations.
 */
import { spawnSync } from "node:child_process";

const generatedPath = "apps/api/internal/database/generated";

function run(command: string, args: string[], cwd?: string): number {
  const result = spawnSync(command, args, {
    cwd,
    stdio: "inherit",
    shell: false,
  });
  return result.status ?? 1;
}

const generateStatus = run("sqlc", ["generate"], "apps/api");
if (generateStatus !== 0) {
  console.error(`
sqlc generate failed. Install sqlc:
  go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0
`);
  process.exit(generateStatus);
}

const diffStatus = run("git", ["diff", "--exit-code", "--", generatedPath]);
if (diffStatus !== 0) {
  console.error(`
sqlc generated code is out of date.

Run:
  bun run db:sqlc

Then commit apps/api/internal/database/generated.
`);
  process.exit(1);
}

console.log("sqlc generated code is up to date.");
