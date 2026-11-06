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
  // The Content-Security-Policy needs a nonce per request, so proxy.ts sets it.
  // See lib/middleware/csp.ts.
  async headers() {
    return [
      {
        source: "/:path*",
        headers: [
          { key: "X-Frame-Options", value: "DENY" },
          { key: "X-Content-Type-Options", value: "nosniff" },
          { key: "Referrer-Policy", value: "strict-origin-when-cross-origin" },
          {
            key: "Permissions-Policy",
            value: "camera=(), microphone=(), geolocation=()",
          },
          // Browsers ignore this over HTTP, so local use isn't affected.
          {
            key: "Strict-Transport-Security",
            value: "max-age=31536000; includeSubDomains",
          },
        ],
      },
    ];
  },
};

export default nextConfig;
