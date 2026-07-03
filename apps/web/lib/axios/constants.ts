import { ROLE_RULES } from "@/lib/middleware/config";

export const AUTH_ENDPOINTS = ["/auth/refresh", "/auth/login", "/auth/logout"];
export const PROTECTED_PATHS = Object.keys(ROLE_RULES);
export const COOKIE_DOMAIN = process.env.NEXT_PUBLIC_COOKIE_DOMAIN || "";

export function isProtectedPath(pathname: string) {
  return PROTECTED_PATHS.some((path) => pathname.startsWith(path));
}
