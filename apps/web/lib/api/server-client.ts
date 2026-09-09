import {
  createClient,
  createConfig,
  type Client,
} from "@repo/api-client/server";
import { get_api_url } from "@/utils/get-api-url";
import { getAllCookies } from "@/utils/get-all-cookies";

/**
 * Per-request server client: forwards Next cookies and uses INTERNAL_API_URL
 * inside Docker via get_api_url(). No refresh interceptor — proxy handles that
 * on guarded routes; expired access simply returns 401.
 */
export async function createServerApiClient(): Promise<Client> {
  const cookie = await getAllCookies();
  return createClient(
    createConfig({
      baseUrl: get_api_url() ?? "",
      headers: cookie ? { Cookie: cookie } : {},
    }),
  );
}

export {
  getApiV1AuthMe,
  getApiV1Health,
  postApiV1AuthLogin,
  postApiV1AuthLogout,
  postApiV1AuthRefresh,
  postApiV1AuthRegister,
} from "@repo/api-client/server";

export type {
  LoginBody,
  MeResponse,
  RegisterBody,
  User,
} from "@repo/api-client/server";
