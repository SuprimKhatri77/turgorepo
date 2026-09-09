#!/usr/bin/env bun
/**
 * Regenerates OpenAPI + api-client, then fails if generated files would change.
 * Use in CI so docs/client stay in sync with @repo/types and @repo/openapi.
 */
import { spawnSync } from "node:child_process";

const generatedPaths = [
  "packages/api-client/src/generated",
  "packages/api-client/src/generated-server",
];

function run(command: string, args: string[]): number {
  const result = spawnSync(command, args, {
    stdio: "inherit",
    shell: false,
  });
  return result.status ?? 1;
}

function hasGeneratedDrift(paths: string[]): boolean {
  const diff = spawnSync("git", ["diff", "--exit-code", "--", ...paths], {
    stdio: "inherit",
    shell: false,
  });
  if ((diff.status ?? 1) !== 0) {
    return true;
  }

  // Catch brand-new generated files that git diff alone would miss.
  const status = spawnSync(
    "git",
    ["status", "--porcelain", "--untracked-files=all", "--", ...paths],
    { encoding: "utf8", shell: false },
  );
  return (status.stdout ?? "")
    .split("\n")
    .some((line) => line.startsWith("??"));
}

const generateStatus = run("bun", ["run", "generate"]);
if (generateStatus !== 0) {
  process.exit(generateStatus);
}

if (hasGeneratedDrift(generatedPaths)) {
  console.error(`
Generated OpenAPI client is out of date.

Run:
  bun run generate

Then commit packages/api-client/src/generated and packages/api-client/src/generated-server.
`);
  process.exit(1);
}

console.log("Generated api-client is up to date.");
