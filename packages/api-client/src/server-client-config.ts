import type { CreateClientConfig } from "./generated-server/client.gen";

export const createClientConfig: CreateClientConfig = (config) => ({
  ...config,
  // OpenAPI paths are rooted at /api/v1/...; point at the API host only.
  // createServerApiClient always passes baseUrl (from getApiUrl()). Without
  // one, prefer the internal URL, as getApiUrl() does on the server.
  baseUrl:
    config?.baseUrl ||
    process.env.INTERNAL_API_URL ||
    process.env.NEXT_PUBLIC_API_URL ||
    "",
});
