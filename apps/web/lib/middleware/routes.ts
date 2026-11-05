import {
  OPEN_AUTH_ROUTES,
  ROLE_RULES,
  UNAUTHENTICATED_ONLY_ROUTES,
} from "./config";

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

export function getRequiredRoles(pathname: string): string[] | undefined {
  return Object.entries(ROLE_RULES).find(([route]) =>
    matchesRoute(pathname, route),
  )?.[1];
}
