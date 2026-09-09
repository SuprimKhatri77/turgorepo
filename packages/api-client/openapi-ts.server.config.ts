import { defineConfig } from "@hey-api/openapi-ts";

export default defineConfig({
  input: "../../apps/api/openapi.json",
  output: {
    path: "src/generated-server",
    // Avoid `.ts` import suffixes so Next.js / bundler resolution stays happy.
    importFileExtension: null,
  },
  plugins: [
    {
      name: "@hey-api/client-fetch",
      runtimeConfigPath: "../server-client-config",
    },
  ],
});
