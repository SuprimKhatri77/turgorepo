import type { User } from "@repo/types";
import { redirect } from "next/navigation";
import { SESSION_EXPIRED_PATH } from "@/lib/middleware/config";
import { fetchCurrentUser } from "./fetch-current-user";

/**
 * Loads the user for a protected page. proxy.ts only decodes tokens, so this is
 * the check the API has verified, including the role stored in the database.
 */
export async function requireUser(allowedRoles?: string[]): Promise<User> {
  const user = await fetchCurrentUser();
  if (!user) {
    redirect(SESSION_EXPIRED_PATH);
  }
  if (allowedRoles && !allowedRoles.includes(user.role)) {
    redirect("/");
  }
  return user;
}
