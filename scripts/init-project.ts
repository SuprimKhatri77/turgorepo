#!/usr/bin/env bun
/**
 * Turns a fresh copy of this template into your own project, then removes itself.
 *
 *   bun run init --name acme-app --module github.com/acme/acme-app
 */
import { spawnSync } from "node:child_process";
import { randomBytes } from "node:crypto";
import { existsSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { parseArgs } from "node:util";

const TEMPLATE_NAME = "turgorepo";
const TEMPLATE_TITLE = "Turgorepo";
const TEMPLATE_MODULE = "github.com/suprimkhatri77/turgorepo";
const SCRIPT_PATH = "scripts/init-project.ts";

const USAGE = `Usage: bun run init --name <name> --module <go-module-path>

  --name    Lowercase project name, such as acme-app
  --module  Go module path for the repo, such as github.com/acme/acme-app`;

function fail(message: string): never {
  console.error(`\n${message}\n`);
  process.exit(1);
}

function run(command: string, args: string[], cwd?: string) {
  console.log(`> ${command} ${args.join(" ")}`);
  const result = spawnSync(command, args, { cwd, stdio: "inherit" });
  if (result.status !== 0) {
    fail(
      `"${command} ${args.join(" ")}" failed. Fix the cause and run it again.`,
    );
  }
}

function output(command: string, args: string[]): string {
  const result = spawnSync(command, args, { encoding: "utf8" });
  if (result.status !== 0) {
    fail(`"${command} ${args.join(" ")}" failed: ${result.stderr}`);
  }
  return result.stdout;
}

function parseOptions() {
  const { values } = parseArgs({
    options: { name: { type: "string" }, module: { type: "string" } },
  });
  const { name, module: goModule } = values;

  if (!name || !goModule) fail(USAGE);
  if (!/^[a-z][a-z0-9]*(-[a-z0-9]+)*$/.test(name)) {
    fail(
      `Invalid --name "${name}". Use lowercase letters, numbers and hyphens.`,
    );
  }
  if (!/^[a-z0-9.-]+(\/[A-Za-z0-9._~-]+)+$/.test(goModule)) {
    fail(`Invalid --module "${goModule}". Example: github.com/acme/acme-app`);
  }
  return { name, goModule };
}

function titleFrom(name: string): string {
  return name
    .split("-")
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}

/** Removes sections that only make sense in the template itself. */
function removeTemplateSections(content: string): string {
  return content.replace(
    /^[ \t]*(?:<!-- template:start -->|# template:start)\n[\s\S]*?(?:<!-- template:end -->|# template:end)\n+/gm,
    "",
  );
}

function renameFiles(name: string, goModule: string) {
  const title = titleFrom(name);
  // Hyphens would need quoting in SQL, so database names use underscores.
  const database = name.replaceAll("-", "_");

  const files = output("git", ["ls-files"])
    .split("\n")
    .filter((file) => file && file !== SCRIPT_PATH && file !== "LICENSE");

  let changed = 0;
  for (const file of files) {
    if (!existsSync(file)) continue;
    const bytes = readFileSync(file);
    if (bytes.includes(0)) continue; // binary

    const before = bytes.toString("utf8");
    const after = removeTemplateSections(before)
      .replaceAll(TEMPLATE_MODULE, goModule)
      .replaceAll(`/${TEMPLATE_NAME}?`, `/${database}?`)
      .replaceAll(`POSTGRES_DB=${TEMPLATE_NAME}`, `POSTGRES_DB=${database}`)
      .replaceAll(`POSTGRES_DB: ${TEMPLATE_NAME}`, `POSTGRES_DB: ${database}`)
      .replaceAll(TEMPLATE_NAME, name)
      .replaceAll(TEMPLATE_TITLE, title);

    if (after !== before) {
      writeFileSync(file, after);
      changed++;
    }
  }
  console.log(`Updated ${changed} files.`);
}

function removeInitScript() {
  const packageJson = JSON.parse(readFileSync("package.json", "utf8"));
  delete packageJson.scripts.init;
  writeFileSync("package.json", `${JSON.stringify(packageJson, null, 2)}\n`);
  rmSync(SCRIPT_PATH);
}

/** Creates a local env file with real secrets, keeping one that already exists. */
function createEnvFile() {
  if (existsSync(".env.local")) {
    console.log("Keeping the existing .env.local.");
    return;
  }
  const secret = () => randomBytes(48).toString("base64");
  const env = readFileSync(".env.example", "utf8")
    .replace(/^JWT_ACCESS_SECRET=.*$/m, `JWT_ACCESS_SECRET=${secret()}`)
    .replace(/^JWT_REFRESH_SECRET=.*$/m, `JWT_REFRESH_SECRET=${secret()}`);
  writeFileSync(".env.local", env);
  console.log("Created .env.local with generated JWT secrets.");
}

const { name, goModule } = parseOptions();

if (output("git", ["status", "--porcelain"]).trim()) {
  fail(
    "Commit or stash your changes first, so this is easy to review and undo.",
  );
}

renameFiles(name, goModule);
removeInitScript();
createEnvFile();

run("bun", ["install"]);
run("bun", ["run", "generate"]);
run("gofmt", ["-w", "."], "apps/api");
run("go", ["mod", "tidy"], "apps/api");

console.log(`
${titleFrom(name)} is ready. Review the changes, commit them, and start it:

  git add -A && git commit -m "chore: initialize ${name}"
  bun run docker:dev:up
  bun run db:seed
`);
