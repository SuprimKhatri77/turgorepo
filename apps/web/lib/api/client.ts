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
  postApiV1AuthForgotPassword,
  postApiV1AuthLogin,
  postApiV1AuthLogout,
  postApiV1AuthLogoutAll,
  postApiV1AuthRefresh,
  postApiV1AuthRegister,
  postApiV1AuthResendVerification,
  postApiV1AuthResetPassword,
  postApiV1AuthVerifyEmail,
} from "@repo/api-client";

export type {
  ForgotPasswordBody,
  LoginBody,
  MeResponse,
  RegisterBody,
  ResendVerificationBody,
  ResetPasswordBody,
  User,
  VerifyEmailBody,
} from "@repo/api-client";
