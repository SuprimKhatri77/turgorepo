import { registerAuthPaths } from "./auth/index.ts";
import { registerHealthPath } from "./health.ts";

export function registerPaths() {
  registerHealthPath();
  registerAuthPaths();
}
