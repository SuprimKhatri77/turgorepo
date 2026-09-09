import type { CreateClientConfig } from "./generated-server/client.gen";

export const createClientConfig: CreateClientConfig = (config) => ({
  ...config,
  // OpenAPI paths are rooted at /api/v1/...; point at the API host only.
  // Prefer an explicit override (e.g. get_api_url() via createServerApiClient).
  baseUrl:
    config?.baseUrl ||
    process.env.NEXT_PUBLIC_API_URL ||
    process.env.INTERNAL_API_URL ||
    "",
});
