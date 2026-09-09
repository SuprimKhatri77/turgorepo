import type { User } from "@repo/types";
import { cookies } from "next/headers";
import { createServerApiClient, getApiV1AuthMe } from "@/lib/api/server-client";

/**
 * Authoritative current user from GET /auth/me (DB-backed).
 * Use when a page needs fresh user data.
 *
 * For cheap layout hydration from JWT claims, prefer getCurrentUser().
 */
export async function fetchCurrentUser(): Promise<User | null> {
  const cookieStore = await cookies();
  if (!cookieStore.get("access_token")?.value) {
    return null;
  }

  try {
    const client = await createServerApiClient();
    const { data, error } = await getApiV1AuthMe({ client });
    if (error || !data?.data) {
      return null;
    }
    return data.data as User;
  } catch {
    return null;
  }
}
