import { resolve } from "node:path";
import type { NextConfig } from "next";

// Next only reads env files from its own folder, but this repo keeps one at the
// root. Real environment variables win, and Docker and CI have no file to load.
try {
  process.loadEnvFile(resolve(process.cwd(), "../../.env.local"));
} catch (error) {
  if ((error as NodeJS.ErrnoException).code !== "ENOENT") throw error;
}

const nextConfig: NextConfig = {
  // Set by apps/web/Dockerfile. Off otherwise, because `next start` can't serve
  // standalone output.
  output: process.env.NEXT_OUTPUT === "standalone" ? "standalone" : undefined,
  // Workspace packages live above this folder, so tracing has to start at the root.
  outputFileTracingRoot: resolve(process.cwd(), "../.."),
};

export default nextConfig;
