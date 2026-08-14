/** @type {import("lint-staged").Configuration} */
export default {
  "apps/web/**/*.{js,jsx,ts,tsx,mjs,cjs}": (filenames) => [
    `bunx eslint --fix --max-warnings=0 -c apps/web/eslint.config.mjs ${filenames.join(" ")}`,
    `bunx prettier --write ${filenames.join(" ")}`,
  ],
  "packages/ui/**/*.{js,jsx,ts,tsx,mjs,cjs}": (filenames) => [
    `bunx eslint --fix --max-warnings=0 -c packages/ui/eslint.config.mjs ${filenames.join(" ")}`,
    `bunx prettier --write ${filenames.join(" ")}`,
  ],
  // Keep OpenAPI + typed client in sync when contract sources change.
  "packages/{types,openapi}/**/*.{ts,tsx}": () => [
    "bun run generate",
    "git add apps/api/openapi.json packages/api-client/src/generated",
  ],
  // Keep sqlc Go output in sync when queries/schema change.
  "apps/api/{sqlc.yml,internal/database/queries/**/*.sql,migrations/**/*.sql}":
    () => ["bun run db:sqlc", "git add apps/api/internal/database/generated"],
  "*.{js,jsx,ts,tsx,mjs,cjs,json,yml,yaml,css}": ["bunx prettier --write"],
  "*.go": ["gofmt -w"],
};
