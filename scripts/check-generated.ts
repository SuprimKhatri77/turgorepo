#!/usr/bin/env bun
/**
 * Regenerates OpenAPI + api-client, then fails if generated files would change.
 * Use in CI so docs/client stay in sync with @repo/types and @repo/openapi.
 */
import { spawnSync } from "node:child_process";

const generatedPaths = [
  "apps/api/openapi.json",
  "packages/api-client/src/generated",
];

function run(command: string, args: string[]): number {
  const result = spawnSync(command, args, {
    stdio: "inherit",
    shell: false,
  });
  return result.status ?? 1;
}

const generateStatus = run("bun", ["run", "generate"]);
if (generateStatus !== 0) {
  process.exit(generateStatus);
}

const diffStatus = run("git", ["diff", "--exit-code", "--", ...generatedPaths]);
if (diffStatus !== 0) {
  console.error(`
Generated OpenAPI / API client is out of date.

Run:
  bun run generate

Then commit apps/api/openapi.json and packages/api-client/src/generated.
`);
  process.exit(1);
}

console.log("Generated OpenAPI and api-client are up to date.");
