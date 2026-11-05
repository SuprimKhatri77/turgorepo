import { resolve } from "node:path";
import type { NextConfig } from "next";

// Next only reads env files from its own folder, but this repo keeps one at the
// root. Real environment variables win, and Docker and CI have no file to load.
try {
  process.loadEnvFile(resolve(process.cwd(), "../../.env.local"));
} catch (error) {
  if ((error as NodeJS.ErrnoException).code !== "ENOENT") throw error;
}

const nextConfig: NextConfig = {/* config options here */};

export default nextConfig;
