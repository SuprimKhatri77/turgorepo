/** @type {import("lint-staged").Configuration} */
export default {
  "apps/web/**/*.{js,jsx,ts,tsx,mjs,cjs}": [
    "bunx eslint --fix --max-warnings=0",
    "bunx prettier --write",
  ],
  "packages/ui/**/*.{js,jsx,ts,tsx,mjs,cjs}": [
    "bunx eslint --fix --max-warnings=0",
    "bunx prettier --write",
  ],
  "*.{js,jsx,ts,tsx,mjs,cjs,json,yml,yaml,css}": ["bunx prettier --write"],
  "*.go": ["gofmt -w"],
};
