import type { CreateClientConfig } from "./generated-server/client.gen";

export const createClientConfig: CreateClientConfig = (config) => ({
  ...config,
  // OpenAPI paths are rooted at /api/v1/...; point at the API host only.
  // RSC callers should override with get_api_url() (INTERNAL_API_URL in Docker).
  baseUrl: process.env.NEXT_PUBLIC_API_URL ?? "",
});
