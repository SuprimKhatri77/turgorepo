import { NextRequest, NextResponse } from "next/server";
import { attemptRefresh, attemptRefreshForAuthRoute, getSession } from "./auth";
import { redirectForRole, redirectTo } from "./redirects";

export async function handleAuthRoute(req: NextRequest): Promise<NextResponse> {
  const { pathname } = req.nextUrl;
  const refreshToken = req.cookies.get("refresh_token")?.value;

  if (!refreshToken) {
    console.log("auth route: no refresh token, allowing through:", pathname);
    return NextResponse.next();
  }

  console.log("auth route: has refresh token:", pathname);

  const accessToken = req.cookies.get("access_token")?.value;
  if (accessToken) {
    const user = getSession(accessToken);
    console.log("auth route: session user:", user?.role ?? "null");

    if (user) {
      const roleRedirect = redirectForRole(req, user.role);
      if (roleRedirect) return roleRedirect;
    }
  }

  console.log("auth route: access missing or expired, attempting refresh");
  return attemptRefreshForAuthRoute(refreshToken, req);
}

export async function handleProtectedRoute(
  req: NextRequest,
  requiredRoles: string[],
): Promise<NextResponse> {
  const { pathname } = req.nextUrl;
  const accessToken = req.cookies.get("access_token")?.value;
  const refreshToken = req.cookies.get("refresh_token")?.value;

  console.log("protected route:", pathname, {
    hasAccessToken: !!accessToken,
    hasRefreshToken: !!refreshToken,
  });

  if (!refreshToken) {
    console.log("protected route: no refresh token, redirecting to login");
    return redirectTo(req, "/auth/login");
  }

  if (accessToken) {
    const user = getSession(accessToken);
    console.log("protected route: session user:", user?.role ?? "null");

    if (user) {
      if (!requiredRoles.includes(user.role)) {
        console.log("protected route: wrong role, redirecting home");
        return redirectTo(req, "/");
      }
      return NextResponse.next();
    }
  }

  console.log("protected route: attempting refresh");
  return attemptRefresh(refreshToken, req, requiredRoles);
}
