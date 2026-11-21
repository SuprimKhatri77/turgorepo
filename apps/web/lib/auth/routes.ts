import type { UserRole } from "@repo/types";

/**
 * The only place to protect a page: add its path here with the roles allowed.
 * proxy.ts runs on every page, so nothing else needs changing.
 */
export const ROLE_RULES: Record<string, UserRole[]> = {
  "/admin": ["admin", "superadmin"],
  "/dashboard": ["member", "admin", "superadmin"],
};

export const UNAUTHENTICATED_ONLY_ROUTES = ["/auth"];

/**
 * Exceptions to the above. Someone logged in can still follow a link from an
 * email, such as when they're logged in on this device as another account.
 */
export const OPEN_AUTH_ROUTES = ["/auth/reset-password", "/auth/verify-email"];

/** The login page. */
export const LOGIN_PATH = "/auth/login";

/** Sends the user to log in again, clearing a session the API has rejected. */
export const SESSION_EXPIRED_PATH = `${LOGIN_PATH}?expired=1`;

/** Where a user lands after logging in. */
export function homePathForRole(role: UserRole): string {
  return role === "admin" || role === "superadmin" ? "/admin" : "/dashboard";
}

/** Matches a route and its children, but not /admin against /administrator. */
export function matchesRoute(pathname: string, route: string): boolean {
  return pathname === route || pathname.startsWith(`${route}/`);
}

export function isUnauthenticatedOnlyRoute(pathname: string): boolean {
  if (OPEN_AUTH_ROUTES.some((route) => matchesRoute(pathname, route))) {
    return false;
  }
  return UNAUTHENTICATED_ONLY_ROUTES.some((route) =>
    matchesRoute(pathname, route),
  );
}

export function getRequiredRoles(pathname: string): UserRole[] | undefined {
  return Object.entries(ROLE_RULES).find(([route]) =>
    matchesRoute(pathname, route),
  )?.[1];
}
