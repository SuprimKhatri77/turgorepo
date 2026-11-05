import type { User } from "@repo/types";
import { cookies } from "next/headers";
import { userFromAccessToken } from "./claims";

/**
 * Cheap current user from JWT claims (no API call).
 * Prefer for layout hydration / AuthProvider seed.
 *
 * For an authoritative DB-backed user, use fetchCurrentUser().
 */
export async function getCurrentUser(): Promise<User | null> {
  const cookieStore = await cookies();
  const accessToken = cookieStore.get("access_token")?.value;
  if (!accessToken) return null;

  return userFromAccessToken(accessToken);
}
