import {
  createClient,
  createConfig,
  type Client,
} from "@repo/api-client/server";
import { cookies } from "next/headers";
import { getApiUrl } from "@/lib/api/url";

/** Thrown when neither NEXT_PUBLIC_API_URL nor INTERNAL_API_URL is set. */
export class ApiUrlNotConfiguredError extends Error {
  constructor() {
    super(
      "API URL is not configured. Set NEXT_PUBLIC_API_URL (and INTERNAL_API_URL in Docker).",
    );
    this.name = "ApiUrlNotConfiguredError";
  }
}

/**
 * Per-request server client: forwards Next cookies and uses INTERNAL_API_URL
 * inside Docker via getApiUrl(). No refresh interceptor — proxy handles that
 * on guarded routes; expired access simply returns 401.
 */
export async function createServerApiClient(): Promise<Client> {
  const baseUrl = getApiUrl();
  if (!baseUrl) {
    throw new ApiUrlNotConfiguredError();
  }

  const cookie = await getAllCookies();
  return createClient(
    createConfig({
      baseUrl,
      headers: cookie ? { Cookie: cookie } : {},
    }),
  );
}

async function getAllCookies() {
  const cookieStore = await cookies();
  const allCookies = cookieStore
    .getAll()
    .map((cookie) => `${cookie.name}=${cookie.value}`)
    .join("; ");

  return allCookies;
}
