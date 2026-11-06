import type { NextRequest, NextResponse } from "next/server";
import { userFromAccessToken } from "@/lib/api/auth/claims";
import {
  attemptRefresh,
  attemptRefreshForAuthRoute,
  clearAuthCookies,
} from "./auth";
import { continueRequest, redirectForRole, redirectTo } from "./redirects";

export async function handleAuthRoute(
  req: NextRequest,
  requestHeaders: Headers,
): Promise<NextResponse> {
  // The API rejected this session. The token can still look valid here, because
  // the proxy only decodes it, so redirecting back would loop forever.
  if (req.nextUrl.searchParams.has("expired")) {
    const response = continueRequest(requestHeaders);
    clearAuthCookies(response);
    return response;
  }

  const refreshToken = req.cookies.get("refresh_token")?.value;
  if (!refreshToken) {
    return continueRequest(requestHeaders);
  }

  const accessToken = req.cookies.get("access_token")?.value;
  if (accessToken) {
    const user = userFromAccessToken(accessToken);
    if (user) {
      return redirectForRole(req, user.role);
    }
  }

  return attemptRefreshForAuthRoute(refreshToken, req, requestHeaders);
}

export async function handleProtectedRoute(
  req: NextRequest,
  requestHeaders: Headers,
  requiredRoles: string[],
): Promise<NextResponse> {
  const accessToken = req.cookies.get("access_token")?.value;
  const refreshToken = req.cookies.get("refresh_token")?.value;

  if (!refreshToken) {
    return redirectTo(req, "/auth/login");
  }

  if (accessToken) {
    const user = userFromAccessToken(accessToken);
    if (user) {
      if (!requiredRoles.includes(user.role)) {
        return redirectTo(req, "/");
      }
      return continueRequest(requestHeaders);
    }
  }

  return attemptRefresh(refreshToken, req, requestHeaders, requiredRoles);
}
