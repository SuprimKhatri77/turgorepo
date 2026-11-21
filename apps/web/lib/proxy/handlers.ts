import type { NextRequest, NextResponse } from "next/server";
import { userFromAccessToken } from "@/lib/auth/claims";
import { ACCESS_TOKEN_COOKIE, REFRESH_TOKEN_COOKIE } from "@/lib/auth/cookies";
import { LOGIN_PATH } from "@/lib/auth/routes";
import {
  clearAuthCookies,
  refreshForAuthRoute,
  refreshForProtectedRoute,
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

  const refreshToken = req.cookies.get(REFRESH_TOKEN_COOKIE)?.value;
  if (!refreshToken) {
    return continueRequest(requestHeaders);
  }

  const accessToken = req.cookies.get(ACCESS_TOKEN_COOKIE)?.value;
  if (accessToken) {
    const user = userFromAccessToken(accessToken);
    if (user) {
      return redirectForRole(req, user.role);
    }
  }

  return refreshForAuthRoute(refreshToken, req, requestHeaders);
}

export async function handleProtectedRoute(
  req: NextRequest,
  requestHeaders: Headers,
  requiredRoles: string[],
): Promise<NextResponse> {
  const accessToken = req.cookies.get(ACCESS_TOKEN_COOKIE)?.value;
  const refreshToken = req.cookies.get(REFRESH_TOKEN_COOKIE)?.value;

  if (!refreshToken) {
    return redirectTo(req, LOGIN_PATH);
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

  return refreshForProtectedRoute(
    refreshToken,
    req,
    requestHeaders,
    requiredRoles,
  );
}
