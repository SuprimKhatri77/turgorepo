import {
  client,
  createClient,
  createConfig,
  type Client,
} from "@repo/api-client";
import api from "@/lib/axios";

/**
 * Typed API client backed by the shared axios instance (refresh interceptors).
 * OpenAPI paths include `/api/v1/...`, so baseURL is the API host only.
 */
export const apiClient: Client = createClient(
  createConfig({
    axios: api,
    baseURL: process.env.NEXT_PUBLIC_API_URL ?? "",
    withCredentials: true,
  }),
);

// Keep the package default client in sync for direct SDK imports.
client.setConfig({
  axios: api,
  baseURL: process.env.NEXT_PUBLIC_API_URL ?? "",
  withCredentials: true,
});

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
