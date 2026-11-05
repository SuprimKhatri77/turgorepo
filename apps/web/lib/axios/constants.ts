import { ROLE_RULES } from "@/lib/middleware/config";
import { matchesRoute } from "@/lib/middleware/routes";

export const AUTH_ENDPOINTS = [
  "/auth/refresh",
  "/auth/login",
  "/auth/logout",
  "/auth/register",
];
export const PROTECTED_PATHS = Object.keys(ROLE_RULES);

export function isProtectedPath(pathname: string) {
  return PROTECTED_PATHS.some((path) => matchesRoute(pathname, path));
}
