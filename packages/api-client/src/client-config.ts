import type { CreateClientConfig } from "./generated/client.gen";

export const createClientConfig: CreateClientConfig = (config) => ({
  ...config,
  // OpenAPI paths are rooted at /api/v1/...; point at the API host only.
  baseURL: process.env.NEXT_PUBLIC_API_URL ?? "",
  withCredentials: true,
});
