/**
 * The only place to protect a page: add its path here with the roles allowed.
 * proxy.ts runs on every page, so nothing else needs changing.
 */
export const ROLE_RULES: Record<string, string[]> = {
  "/admin": ["admin", "superadmin"],
  "/dashboard": ["member", "admin", "superadmin"],
};

export const UNAUTHENTICATED_ONLY_ROUTES = ["/auth"];

/**
 * Exceptions to the above. Someone logged in can still follow a link from an
 * email, such as when they're logged in on this device as another account.
 */
export const OPEN_AUTH_ROUTES = ["/auth/reset-password", "/auth/verify-email"];

/** Sends the user to log in again, clearing a session the API has rejected. */
export const SESSION_EXPIRED_PATH = "/auth/login?expired=1";

/** Where a user lands after logging in. */
export function homePathForRole(role: string): string {
  return role === "admin" || role === "superadmin" ? "/admin" : "/dashboard";
}
