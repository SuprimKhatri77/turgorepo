import { NextRequest, NextResponse } from "next/server";
import { User } from "@repo/types";
import { jwtDecode } from "jwt-decode";
import { get_api_url } from "@/utils/get-api-url";

function clearAuthCookies(response: NextResponse) {
  const cookieOptions = {
    maxAge: 0,
    path: "/",
    domain:
      process.env.NODE_ENV === "production" ? process.env.COOKIE_DOMAIN : "",
    secure: process.env.NODE_ENV === "production",
  };

  response.cookies.set("access_token", "", {
    ...cookieOptions,
    httpOnly: true,
  });
  response.cookies.set("refresh_token", "", {
    ...cookieOptions,
    httpOnly: true,
  });
  response.cookies.set("is_logged_in", "", {
    ...cookieOptions,
    httpOnly: false,
  });
}

export function getSession(token: string): User | null {
  try {
    const decoded = jwtDecode<User & { exp: number }>(token);
    if (decoded.exp * 1000 < Date.now()) return null;
    return decoded;
  } catch {
    return null;
  }
}

function forwardSetCookie(response: NextResponse, cookies: string[]): void {
  cookies.forEach((cookie) => {
    response.headers.append("set-cookie", cookie);
  });
}

export async function attemptRefresh(
  refreshToken: string,
  req: NextRequest,
  requiredRoles: string[],
) {
  try {
    console.log("attemptRefresh called, path:", req.nextUrl.pathname);
    const res = await fetch(`${get_api_url()}/api/v1/auth/refresh`, {
      method: "POST",
      headers: { Cookie: `refresh_token=${refreshToken}` },
    });

    console.log("refresh response status:", res.status);

    if (!res.ok) {
      if (res.status === 401) {
        const redirect = NextResponse.redirect(new URL("/auth/login", req.url));
        clearAuthCookies(redirect);
        return redirect;
      }
      // Non-401 failure (e.g. 500): redirect without clearing cookies — session may still be valid
      return NextResponse.redirect(new URL("/auth/login", req.url));
    }

    const cookies = res.headers.getSetCookie();
    const newAccessToken = extractTokenFromCookie(cookies.join("; "));

    if (!newAccessToken) {
      return NextResponse.redirect(new URL("/auth/login", req.url));
    }

    const user = await getSession(newAccessToken);

    if (!user) {
      return NextResponse.redirect(new URL("/auth/login", req.url));
    }

    if (!requiredRoles.includes(user.role)) {
      const redirect = NextResponse.redirect(new URL("/", req.url));
      forwardSetCookie(redirect, cookies);
      return redirect;
    }

    // Grace-window response: API only sets access cookie, no new refresh cookie.
    // Retain the original refresh cookie instead of injecting null.
    const newRefreshToken = extractRefreshTokenFromCookie(cookies.join("; "));
    const refreshCookieValue = newRefreshToken ?? refreshToken;

    const response = NextResponse.next({
      request: {
        headers: new Headers({
          ...Object.fromEntries(req.headers),
          cookie: `access_token=${newAccessToken}; refresh_token=${refreshCookieValue}`,
        }),
      },
    });

    forwardSetCookie(response, cookies);
    return response;
  } catch (error) {
    // Network/transport error: don't clear cookies — the session may still be valid
    console.log("attemptRefresh threw an error:", error);
    return NextResponse.redirect(new URL("/auth/login", req.url));
  }
}

export async function attemptRefreshForAuthRoute(
  refreshToken: string,
  req: NextRequest,
) {
  try {
    const res = await fetch(`${get_api_url()}/api/v1/auth/refresh`, {
      method: "POST",
      headers: { Cookie: `refresh_token=${refreshToken}` },
    });

    if (!res.ok) {
      const response = NextResponse.next();
      if (res.status === 401) {
        clearAuthCookies(response);
      }
      return response;
    }

    const cookies = res.headers.getSetCookie();
    const newAccessToken = extractTokenFromCookie(cookies.join("; "));

    if (!newAccessToken) {
      const response = NextResponse.next();
      clearAuthCookies(response);
      return response;
    }

    const user = await getSession(newAccessToken);

    if (!user) {
      const response = NextResponse.next();
      clearAuthCookies(response);
      return response;
    }

    if (user.role === "admin" || user.role === "superadmin") {
      const redirect = NextResponse.redirect(new URL("/admin", req.url));
      forwardSetCookie(redirect, cookies);
      return redirect;
    }

    const response = NextResponse.next();
    forwardSetCookie(response, cookies);
    return response;
  } catch (error) {
    // Network/transport error: don't clear cookies
    console.log("attemptRefreshForAuthRoute threw an error:", error);
    return NextResponse.next();
  }
}

export function extractTokenFromCookie(
  setCookie: string | null,
): string | null {
  if (!setCookie) return null;
  const match = setCookie.match(/access_token=([^;]+)/);
  return match?.[1] ?? null;
}

export function extractRefreshTokenFromCookie(
  setCookie: string | null,
): string | null {
  if (!setCookie) return null;
  const match = setCookie.match(/refresh_token=([^;]+)/);
  return match?.[1] ?? null;
}
