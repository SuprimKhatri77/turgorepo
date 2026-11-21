import {
  createClient,
  createConfig,
  type Client,
} from "@repo/api-client/server";
import { cookies } from "next/headers";
import { getApiUrl } from "@/lib/api/url";

/**
 * Per-request server client: forwards Next cookies and uses INTERNAL_API_URL
 * inside Docker via getApiUrl(). No refresh interceptor — proxy handles that
 * on guarded routes; expired access simply returns 401.
 */
export async function createServerApiClient(): Promise<Client> {
  const baseUrl = getApiUrl();
  if (!baseUrl) {
    throw new Error(
      "API URL is not configured. Set NEXT_PUBLIC_API_URL (and INTERNAL_API_URL in Docker).",
    );
  }

  const cookie = await getAllCookies();
  return createClient(
    createConfig({
      baseUrl,
      headers: cookie ? { Cookie: cookie } : {},
    }),
  );
}

export {
  getApiV1AuthMe,
  getApiV1Health,
  postApiV1AuthForgotPassword,
  postApiV1AuthLogin,
  postApiV1AuthLogout,
  postApiV1AuthRefresh,
  postApiV1AuthRegister,
  postApiV1AuthResetPassword,
} from "@repo/api-client/server";

export type {
  ForgotPasswordBody,
  LoginBody,
  MeResponse,
  RegisterBody,
  ResetPasswordBody,
  User,
} from "@repo/api-client/server";

async function getAllCookies() {
  const cookieStore = await cookies();
  const allCookies = cookieStore
    .getAll()
    .map((cookie) => `${cookie.name}=${cookie.value}`)
    .join("; ");

  return allCookies;
}
