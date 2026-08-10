import {
  client,
  createClient,
  createConfig,
  type Client,
} from "@repo/api-client";
import api from "@/lib/axios";

const apiConfig = {
  axios: api,
  // OpenAPI paths include `/api/v1/...`; baseURL is the API host only.
  baseURL: process.env.NEXT_PUBLIC_API_URL ?? "",
  withCredentials: true,
} as const;

/**
 * Typed API client backed by the shared axios instance (refresh interceptors).
 */
export const apiClient: Client = createClient(createConfig({ ...apiConfig }));

// Keep the package default client in sync for direct SDK imports.
client.setConfig({ ...apiConfig });

export {
  getApiV1AuthMe,
  getApiV1Health,
  postApiV1AuthLogin,
  postApiV1AuthLogout,
  postApiV1AuthRefresh,
  postApiV1AuthRegister,
} from "@repo/api-client";

export type {
  LoginBody,
  MeResponse,
  RegisterBody,
  User,
} from "@repo/api-client";
