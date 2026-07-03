import { ROLE_RULES, UNAUTHENTICATED_ONLY_ROUTES } from "./config";

export function isUnauthenticatedOnlyRoute(pathname: string): boolean {
  return UNAUTHENTICATED_ONLY_ROUTES.some((route) =>
    pathname.startsWith(route),
  );
}

export function getRequiredRoles(pathname: string): string[] | undefined {
  return Object.entries(ROLE_RULES).find(([path]) =>
    pathname.startsWith(path),
  )?.[1];
}
