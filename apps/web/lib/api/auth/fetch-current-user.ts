import type { User } from "@repo/types";
import { cookies } from "next/headers";
import { createServerApiClient, getApiV1AuthMe } from "@/lib/api/server-client";

/**
 * Authoritative current user from GET /auth/me (DB-backed).
 * Use when a page needs fresh user data.
 *
 * Returns null when there's no valid session, and throws when the API fails, so
 * an outage shows an error page instead of logging the user out.
 *
 * For cheap layout hydration from JWT claims, prefer getCurrentUser().
 */
export async function fetchCurrentUser(): Promise<User | null> {
  const cookieStore = await cookies();
  if (!cookieStore.get("access_token")?.value) {
    return null;
  }

  const client = await createServerApiClient();
  const { data, error, response } = await getApiV1AuthMe({
    client,
    cache: "no-store",
  });

  if (response?.status === 401 || response?.status === 404) {
    return null;
  }
  if (error || !data?.data) {
    throw new Error(`Failed to load the current user (${response?.status})`);
  }
  return data.data as User;
}
