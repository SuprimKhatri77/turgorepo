import { matchesRoute, ROLE_RULES } from "@/lib/auth/routes";

export const AUTH_ENDPOINTS = [
  "/auth/refresh",
  "/auth/login",
  "/auth/logout",
  "/auth/register",
  "/auth/forgot-password",
  "/auth/reset-password",
];
export const PROTECTED_PATHS = Object.keys(ROLE_RULES);

export function isProtectedPath(pathname: string) {
  return PROTECTED_PATHS.some((path) => matchesRoute(pathname, path));
}
